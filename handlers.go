package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// HTMLTemplateVars used to pass variables to various html templates
type HTMLTemplateVars struct {
	Username    string
	EnableOauth bool
	BaseURL     string
}

/*
	Root handler first checks if user is logged in.  If not logged in then authenticate
	using oauth2.

	If user is logged in then we just load the start page

	Authentication Logic
	- access_type=offline means request access token with offline access if user is logging in
	  for the first time Google will prompt user to grant permission for offline access.
	  Upon accepting offline access google returns access token with refresh token

	- Once refresh tokenn is aquired for the first time we insert it into the database for future retrieval
    - After 12 hours the users cookie will expire and they will need to travel through the roothandler again for login

*/
func rootHandler(w http.ResponseWriter, r *http.Request) {

	if !checkLogin(w, r, false) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// if user is not using https then redirect them
	if r.Header.Get("x-forwarded-proto") != "https" && baseURL != localBaseURL {
		fmt.Printf("TLS handshake is https=false x-forwarded-proto=%s\n", r.Header.Get("x-forwarded-proto"))
		http.Redirect(w, r, baseURL, http.StatusFound)
		return
	}

	err := MainPage.Execute(w, HTMLTemplateVars{"peopi", *EnableOauth, baseURL})
	if err != nil {
		fmt.Println(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// if user is not using https then redirect them
	if r.Header.Get("x-forwarded-proto") != "https" && baseURL != localBaseURL {
		fmt.Printf("TLS handshake is https=false x-forwarded-proto=%s\n", r.Header.Get("x-forwarded-proto"))
		http.Redirect(w, r, baseURL, http.StatusFound)
		return
	}

	err := LoginPage.Execute(w, HTMLTemplateVars{"", *EnableOauth, baseURL})
	if err != nil {
		fmt.Println(err)
	}
}

// LoginRequest is json request used to authenticate users.  password should be base64 encoded
type LoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// authenticates users via oauth or internal auth
func submitLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sr := new(statusResponse)
	if *EnableOauth {
		url := LoginCfg.AuthCodeURL("")
		url = url + OauthURLParams
		params := r.URL.Query()
		paramkeys := make([]string, 0)
		for k := range params {
			for i := range params[k] {
				paramkeys = append(paramkeys, k+"="+params[k][i])
			}
		}
		if len(paramkeys) > 0 {
			url = url + "&state=" + base64.StdEncoding.EncodeToString([]byte(strings.Join(paramkeys, "?")))
		}

		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	var lr LoginRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&lr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(sr.getJSON(err.Error()))
		return
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(lr.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON(err.Error()))
		return
	}

	u, err := GetExistingUser(lr.User)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(sr.getJSON(err.Error()))
		return
	}

	err = u.MatchPassword(string(decodedPassword))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(sr.getJSON("Username or password incorrect"))
		return
	}

	err = createNewSession(w, r, lr.User, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON(err.Error()))
		return
	}
	w.Write(sr.getJSON("success"))
}

// creates and stores a new user in the database for internal auth
func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sr := new(statusResponse)

	if *EnableOauth {
		w.WriteHeader(http.StatusBadRequest)
		NotAuthenticatedTemplate.Execute(w, template.HTML("Unable to register new user as Oauth is enabled"))
		return
	}

	var lr LoginRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&lr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(sr.getJSON(err.Error()))
		return
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(lr.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON(err.Error()))
		return
	}
	_, err = NewUserMustNotExist(lr.User, string(decodedPassword))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON(err.Error()))
		return
	}

	w.Write(sr.getJSON("success"))
}

/*
	This is the oauth2 callback which will authenticate the user and get the token
	A token will last for 3600 seconds and can be used to access the users gmail profile.

	Once user is authenticated then create a new session and set maxage to 12 hours. This means
	user will be logged in for 12 hours
*/
func logincallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	tok, err := LoginCfg.Exchange(oauth2.NoContext, code)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		NotAuthenticatedTemplate.Execute(w, err.Error())
		return
	}

	// get the users profile from google
	CLIENT := LoginCfg.Client(oauth2.NoContext, tok)
	resp, ee := CLIENT.Get("https://www.googleapis.com/plus/v1/people/me")
	if ee != nil {
		fmt.Fprintf(w, "Fetching profile err: %s", ee)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var p ProfileBlob
	err = json.Unmarshal(body, &p)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		NotAuthenticatedTemplate.Execute(w, template.HTML(err.Error()))
		return
	}

	if p.Emails == nil {
		w.WriteHeader(http.StatusUnauthorized)
		NotAuthenticatedTemplate.Execute(w, template.HTML(fmt.Sprintf("Could not get user profile info: %s", body)))
		return
	}

	/*
	   There is a case where user has an expired cookie
	   rootHandler Calls verifyLogin which returns false because of expired cookie error.
	   So the user now has to relogin.  So trash the cookie if it exists and start fresh
	*/
	session, _ := CookieStore.Get(r, SessionName)
	session.Save(r, w)

	session, err = CookieStore.Get(r, SessionName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		NotAuthenticatedTemplate.Execute(w, template.HTML(err.Error()))
		return
	}

	for i := range p.Emails {
		if strings.Contains(p.Emails[i].Value, OauthDomain) {
			session.Values["Email"] = p.Emails[i].Value
			break
		}
	}

	/*
		Check if refresh token was returned by google
		- if no refresh token then check to see if we have one in the database.  If its not in the databse then send the user
		  back to google with approval_prompt=force to ensure we get a new refresh token
		- If refresh token is supplied then update the database with the new token
	*/
	tp := new(TokenTuple)
	if tok.RefreshToken == "" {
		tp, err = GetToken(session.Values["Email"].(string))
		if err != nil {
			// so in this case user did get a access token it came with no refresh token. Since the database does not have the
			// refresh token we must force the user to login again
			http.Redirect(w, r, LoginCfg.AuthCodeURL("")+OauthURLParams+"&approval_prompt=force", http.StatusFound)
			return
		}
	} else {
		// we have a refresh token so update databse
		tp = &TokenTuple{0, session.Values["Email"].(string), EncryptedText{tok.RefreshToken}}
		err = tp.UpdateToken()
		if err != nil {
			fmt.Printf("Token Database Update Error: %s\n", err) // can not be too verbose here as it could leak sensitive info
			w.WriteHeader(http.StatusUnauthorized)

			NotAuthenticatedTemplate.Execute(w, template.HTML(err.Error()))
			return
		}
	}

	for i := range p.Names {
		session.Values["username"] = p.Names[i].displayName // grab the first display name found and end loop
		break
	}
	session.Values["AuthToken"] = tok
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 43200, // 12 hours even though user can refresh token up to 24 hours.
	}
	session.Save(r, w)

	url := ""
	if r.FormValue("state") != "" {
		url = baseURL + "/index.html?state=" + r.FormValue("state")
	} else {
		url = baseURL + "/index.html"
	}

	/*
		We redirect back too root so we can clean up the users URL info and drop logincallback.
		This will prevent errors should user attemmpt to refresh the page
		We do not want to send the user back to parent because because it could cause an infinite Oauth2 loop
	*/
	http.Redirect(w, r, url, http.StatusFound)

}

// LoginStart After a successful new login this function will serve the main page
// If user logs in with cookie then roothandler will take care of this.
// The only flow i see here is if the user decides to book mark index.html.
// Bookmarking index.html will never allow the user to login.
// So in error we offer hints on how to login again.
func LoginStart(w http.ResponseWriter, r *http.Request) {
	if !verifyLogin(r) {
		http.Redirect(w, r, baseURL, http.StatusFound)
		return
	}

	// if user is not using https then redirect through a secure endpoint.  But if basURL is localhost then assume this is a sandbox and let pass
	if r.Header.Get("x-forwarded-proto") != "https" && baseURL != localBaseURL {
		fmt.Printf("TLS handshake is https=false x-forwarded-proto=%s\n", r.Header.Get("x-forwarded-proto"))
		http.Redirect(w, r, baseURL, http.StatusFound)
		return
	}

	MainPage.Execute(w, HTMLTemplateVars{getUserName(r), *EnableOauth, baseURL})
}

/*
	Set the users session cookie key "LoggedIn" to no and redirect user back to
	root page for re-authentication
*/
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := CookieStore.Get(r, SessionName)
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

/*
  Parse Get requests and pass them through to handlers
*/
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sr := new(statusResponse)
	if !checkLogin(w, r, true) {
		return
	}

	uid, err := getSessionUID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Could not fetch your user id: " + err.Error()))
		return
	}

	vals := r.URL.Query()
	if vals.Get("words") == "t" {
		writeWords(w)
	}
	if vals.Get("windows") == "t" {
		writeWindows(w, uid)
	}
	if vals.Get("submissions") == "t" {
		if checkAuthorization(w, vals, uid) {
			writeSubmissionStats(w, vals, uid)
		}
	}
	if vals.Get("panedata") == "t" {
		if checkAuthorization(w, vals, uid) {
			writeJCWindowPanes(w, vals, uid)
		}
	}
	if vals.Get("user") == "t" {
		if checkAuthorization(w, vals, uid) {
			writeUserInfo(w, vals)
		}
	}
	if vals.Get("history") == "t" {
		if checkAuthorization(w, vals, uid) {
			writeHistoryData(w, vals)
		}
	}
	if vals.Get("previouswindows") == "t" {
		writePreviousWindows(w, uid)
	}
}

type WritePreiviouWindowRes struct {
	Pane     string `json:"pane"`
	Nickname string `json:"nickname"`
}

func writePreviousWindows(w http.ResponseWriter, uid int) {
	sr := new(statusResponse)

	res := make([]WritePreiviouWindowRes, 0)
	rows, err := DBI.GetRowSet(fmt.Sprintf(PREVIOUS_WINDOWS_QUERY, uid))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Query Failed: " + err.Error()))
		return
	}
	for rows.Next() {
		var sess string
		var nickname string
		err := rows.Scan(&sess, &nickname)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sr.getJSON("Scan Failed: " + err.Error()))
			return
		}
		res = append(res, WritePreiviouWindowRes{sess, nickname})
	}
	WriteJSONResponse(w, res)
}

/*
  Fetch words from database and write them to ResponseWriter
*/
func writeWords(w http.ResponseWriter) {
	sr := new(statusResponse)
	type APIResponse struct {
		Words []string `json:"words"`
	}
	res := APIResponse{}
	var err error
	res.Words, err = DBI.GetStringList("SELECT word FROM words order by 1")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Fetching words failed: " + err.Error()))
		return
	}
	WriteJSONResponse(w, res)
}

/*
  Fetch and return all the user windows
*/
func writeWindows(w http.ResponseWriter, uid int) {
	sr := new(statusResponse)
	type Window struct {
		CreatedAt time.Time `json:"createdat"`
		Session   string    `json:"session"`
		Nickname  string    `json:"nickname"`
	}
	res := make([]Window, 0)
	rows, dberr := DBI.GetRowSet(fmt.Sprintf("SELECT s.timecreated, s.session, s.nickname FROM sessions s JOIN subjects sj ON sj.session = s.session WHERE sj.uid = %d order by 1", uid))
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Getting sessions fialed: " + dberr.Error()))
		return
	}

	for rows.Next() {
		var t time.Time
		var s string
		var n string
		err := rows.Scan(&t, &s, &n)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sr.getJSON("Parsing tuple fialed: " + err.Error()))
			return
		}
		res = append(res, Window{t, s, n})
	}
	WriteJSONResponse(w, res)
}

// WriteSubmissionsResp returns number of submissions for given pane
type WriteSubmissionsResp struct {
	Submissions int `json:"submissions"`
}

func writeSubmissionStats(w http.ResponseWriter, vals url.Values, uid int) {
	sr := new(statusResponse)
	var err error
	sess := vals.Get("pane")
	if sess == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("no session id found url"))
		return
	}

	res := new(WriteSubmissionsResp)
	res.Submissions, err = DBI.GetIntValue(fmt.Sprintf("select count(*) from (SELECT DISTINCT uid FROM peers p WHERE session = '%s') sub", sess))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("db error: " + err.Error()))
		return
	}
	WriteJSONResponse(w, res)
}

func writeJCWindowPanes(w http.ResponseWriter, vals url.Values, uid int) {
	sr := new(statusResponse)

	sess := vals.Get("pane")
	if sess == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("no session id found url"))
		return
	}

	res := new(JCWindows)
	derr := GetWindowPanesFromDB(res, uid, sess)
	if derr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Getting panes resturn error: " + derr.Error()))
		return
	}
	WriteJSONResponse(w, res)
}

func writeUserInfo(w http.ResponseWriter, vals url.Values) {
	sr := new(statusResponse)
	var err error
	sess := vals.Get("pane")
	if sess == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("no session id found url"))
		return
	}

	type Res struct {
		Email string `json:"email"`
	}
	res := new(Res)
	res.Email, err = DBI.GetStringValue(fmt.Sprintf("SELECT DISTINCT u.username FROM users u JOIN subjects s ON s.uid = u.id WHERE s.session = '%s'", sess))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Getting user info Fialed: " + err.Error()))
		return
	}
	WriteJSONResponse(w, res)
}

// UserHistory themes and words from a pane submission
type UserHistory struct {
	Themes []string `json:"themes"`
	Words  []string `json:"words"`
}

// UsersHistory array of user submissions
type UsersHistory struct {
	Users map[string]UserHistory `json:"users"`
}

func writeHistoryData(w http.ResponseWriter, vals url.Values) {
	sr := new(statusResponse)
	var err error
	sess := vals.Get("pane")
	if sess == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("no session id found url"))
		return
	}

	res := new(UsersHistory)
	res.Users = make(map[string]UserHistory)
	rows, dberr := DBI.GetRowSet(fmt.Sprintf("SELECT u.username, w.theme, w.word FROM peers p JOIN users u ON u.id = p.uid JOIN words w ON w.wid = p.word WHERE p.session = '%s'", sess))
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Geting submission history failed: " + err.Error()))
		return
	}
	for rows.Next() {
		var email string
		var theme string
		var word string
		err := rows.Scan(&email, &theme, &word)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sr.getJSON("bad tuple found in history data: " + err.Error()))
			return
		}
		_, ok := res.Users[email]
		if ok {
			// workaround bug in golan https://github.com/golang/go/issues/3117 where you can not assign directly map[string]User.Theme = x
			u := res.Users[email]
			u.Themes = append(u.Themes, theme)
			u.Words = append(u.Words, word)
			res.Users[email] = u
		} else {
			res.Users[email] = UserHistory{[]string{theme}, []string{word}}
		}
	}
	WriteJSONResponse(w, res)
}

/*
  Parse Get requests and pass them through to handlers
*/
func postHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sr := new(statusResponse)
	if !checkLogin(w, r, true) {
		return
	}

	uid, err := getSessionUID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Could not fetch your user id: " + err.Error()))
		return
	}

	vals := r.URL.Query()
	if vals.Get("new") == "t" {
		CreateNewWindow(w, r, uid)
	}
	if vals.Get("feedback") == "t" {
		SubmitFeedback(w, r, uid)
	}
}

type CreateWindowReq struct {
	Nickname string   `json:"nickname"`
	Words    []string `json:"words"`
}
type CreateWindowRes struct {
	Pane string `json:"pane"`
}

// CreateNewWindow creates a new johari/clifton window
func CreateNewWindow(w http.ResponseWriter, r *http.Request, uid int) {
	sr := new(statusResponse)

	req := new(CreateWindowReq)
	res := new(CreateWindowRes)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write(sr.getJSON("decoder error " + err.Error()))
		return
	}
	sessionID := generateSessionID(uid)

	wids := make([]int, 0)
	wids, err = DBI.GetIntList(fmt.Sprintf("SELECT wid FROM words WHERE word in ('%s')", strings.Join(req.Words, "','")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write(sr.getJSON("resolving word id's failed: " + err.Error()))
		return
	}

	stmt, err := DBI.SQLSession.Prepare("INSERT INTO sessions VALUES(?,?,?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write(sr.getJSON("Preparing nickname insert failed: " + err.Error()))
		return
	}
	_, err = stmt.Exec(time.Now().Format("2006-01-02 15:04:05"), sessionID, req.Nickname)
	//_, err = dbi.SQLSession.Exec(fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%s', '%s')", time.Now().Format(time.RFC3339), sessionID, req.Nickname))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write(sr.getJSON("Creating nickname failed: " + err.Error()))
		return
	}

	query := "INSERT INTO subjects VALUES "
	for i := range wids {
		query += fmt.Sprintf("(%d, '%s', %d),", uid, sessionID, wids[i])
	}
	query = query[:len(query)-1]

	_, err = DBI.SQLSession.Exec(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write(sr.getJSON("Could not create window: " + err.Error()))
		return
	}
	res.Pane = sessionID
	w.WriteHeader(http.StatusCreated)
	WriteJSONResponse(w, res)
}

// SubmitFeedbackReq request for feedback submission
type SubmitFeedbackReq struct {
	Pane  string   `json:"pane"`
	Words []string `json:"words"`
}

// SubmitFeedback - subjects should not submit feedback to themselves , users can not submit multiple feedbacks for the same subject
func SubmitFeedback(w http.ResponseWriter, r *http.Request, uid int) {
	sr := new(statusResponse)

	req := new(SubmitFeedbackReq)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("decoder error " + err.Error()))
		return
	}

	if len(req.Words) <= 0 || req.Pane == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(sr.getJSON(fmt.Sprintf("Missing words or session id in submission: %v", req)))
		return
	}

	var count int
	count, err = DBI.GetIntValue(fmt.Sprintf("SELECT count(*) FROM ( SELECT * FROM peers WHERE uid = %d AND session = '%s' UNION SELECT * from subjects WHERE uid = %d and session = '%s') sub", uid, req.Pane, uid, req.Pane))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Query error: " + err.Error()))
		return
	}
	if count > 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(sr.getJSON("You can not submit feedback to yourself and multiple feedback submission to others are not supported"))
		return
	}

	words := make(map[string]int)
	rows, dberr := DBI.GetRowSet(fmt.Sprintf("SELECT wid, word from words where word in ('%s')", strings.Join(req.Words, "','")))
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Query error: " + dberr.Error()))
		return
	}
	for rows.Next() {
		var id int
		var word string
		err := rows.Scan(&id, &word)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sr.getJSON("Scan error: " + err.Error()))
			return
		}
		words[word] = id
	}

	for i := range req.Words {
		err := DBI.ExecTXQuery("INSERT INTO peers VALUES (?, ?, ?)", uid, req.Pane, words[req.Words[i]])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(sr.getJSON("You can not submit feedback to yourself and multiple feedback submission to others are not supported"))
			return
		}
	}
	w.Write(sr.getJSON("success"))
}

func windowHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(w, r, false) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	err := WindowPage.Execute(w, HTMLTemplateVars{getUserName(r), *EnableOauth, baseURL})
	if err != nil {
		fmt.Println(err)
	}
}

func feedbackHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(w, r, false) {
		vals := r.URL.Query()
		pane := vals.Get("pane")
		http.Redirect(w, r, baseURL+"?feedbackpane="+pane, http.StatusFound)
		return
	}
	FeedbackPage.Execute(w, HTMLTemplateVars{getUserName(r), *EnableOauth, baseURL})
}
func thanksHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(w, r, false) {
		http.Redirect(w, r, baseURL, http.StatusFound)
		return
	}
	ThanksPage.Execute(w, HTMLTemplateVars{getUserName(r), *EnableOauth, baseURL})
}

/*
  Get the user name from session and query for uid
*/
func getSessionUID(r *http.Request) (int, error) {
	session, err := CookieStore.Get(r, SessionName)
	if err != nil {
		return 0, fmt.Errorf(fmt.Sprintf("Failed to get session: %s", err.Error()))
	}
	var email string
	var ok bool
	if email, ok = session.Values["Email"].(string); !ok {
		return 0, fmt.Errorf(fmt.Sprintf("Could not get email from session cookie"))
	}
	if email == "" {
		return 0, fmt.Errorf(fmt.Sprintf("Could not get email from session cookie"))
	}
	return DBI.GetIntValue(fmt.Sprintf("SELECT id FROM users WHERE username = '%s' limit 1", email))
}

// try to get username from session
func getUserName(r *http.Request) string {
	u := ""
	session, err := CookieStore.Get(r, SessionName)
	if err != nil {
		fmt.Printf("Failed to get session: %s\n", err.Error())
		return u
	}
	var ok bool
	if u, ok = session.Values["Email"].(string); !ok {
		fmt.Println("Could not get email from session cookie")
		return u
	}
	if u == "" {
		fmt.Println("Could not get email from session cookie")
		return u
	}
	return u
}

/*
  Given uid generate md5 hash session id
*/
func generateSessionID(uid int) string {
	tNano := time.Now().UnixNano()
	randSource := rand.NewSource(tNano)
	randGen := rand.New(randSource)
	data := []byte(fmt.Sprintf("%d:%d:%d", uid, tNano, randGen.Intn(1000000)))
	return fmt.Sprintf("%x", md5.Sum(data))
}

func createNewSession(w http.ResponseWriter, r *http.Request, user, token string) error {
	session, err := CookieStore.Get(r, SessionName)
	if err != nil {
		return fmt.Errorf("Failed to get session from store: %s", err)
	}
	session.Values["Email"] = user

	if *EnableOauth {
		session.Values["Auth_Token"] = token
	} else {
		session.Values["Auth_Token"] = ""
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 43200, // 12 hours even though user can refresh token up to 24 hours.
	}
	session.Save(r, w)
	return nil
}
