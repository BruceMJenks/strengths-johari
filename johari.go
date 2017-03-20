package main

import (
  "github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
  "net/http"
  "html/template"
  "encoding/json"
  "encoding/base64"
	"encoding/gob"
  "fmt"
  "strings"
  "os"
  "io/ioutil"
  "crypto/md5"
  "math/rand"
  "time"
  "errors"
  "net/url"
)

var (
  store = sessions.NewCookieStore([]byte("lksdjfwijeoijflsdknlsndfpwiejfwnsldkfsdifjflsjkdflsoei"))
  sessionName = "gss-johari"
  CLIENTID string
	CLIENTSECRET string
  OauthURLParams string
  OauthDomain string
  DBURL string
  BASEURL = "http://localhost"
  LOCALBASEURL string // used for testing purposes
  LoginCfg *oauth2.Config

  startPageTemplate = template.Must(template.ParseFiles("tmpl/start.tmpl")) // root page
  windowPageTemplate = template.Must(template.ParseFiles("tmpl/window.tmpl")) 
  feedbackPageTemplate = template.Must(template.ParseFiles("tmpl/feedback.tmpl"))
  thanksPageTemplate = template.Must(template.ParseFiles("tmpl/thanks.tmpl"))
  notAuthenticatedTemplate = template.Must(template.ParseFiles("tmpl/noPermission.tmpl")) // login failure
)



/*
	this is the standard error struct sent back to the frontend in case of internal errors
*/
type statusResponse struct {
	Status string `json:"errmessage"`
}

/*
	Build the statusResponse struct and return marshalled version of struct
*/
func (sr *statusResponse) getJson(s string) []byte {
	sr.Status = s
	jsr, _ := json.Marshal(sr)
	return jsr
}

/*
        conversts struct into json object and writes it back to responseWriter
*/
func WriteJsonResponse(w http.ResponseWriter, data interface{}) {
        sr := new(statusResponse)
        jdata, err := json.Marshal(data)
        if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write(sr.getJson("Failed to marshal data: " + err.Error()))
                return
        }

        w.Write(jdata)
        return
}


/*
	Root handler first checks if user is logged in.  If not logged in then authenticate
	useing oauth2.  The return of the AuthCodeURL will be https://accounts.google.com/o/oauth2/auth

	If user is logged in then we just load the start page

	Authentication Logic
	- access_type=offline means request access token with offline access
	  access token expires after 1 hour so if user is logging in for the first time Google will prompt user
	  to grant permission for offline access.  Upon accepting offline access google returns access token with refresh token

	- Once refresh toekn is aquired for the first time we insert it into the database for future retrevial

	- We use customers access token to send a gmail and store this access token a browser cookie along with user name
      if the access token expires then email will be sent using the refresh token
    - After 12 hours the users cookie will expire and they will need to travel through the roothandler again for login

*/
func rootHandler(w http.ResponseWriter, r *http.Request) {

	if !verifyLogin(r) {
		url := LoginCfg.AuthCodeURL("")
		url = url + OauthURLParams
		// this will preseve the casenumber in the URI path during Oauth2 redirect
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

	// if user is not using https then redirect them
	if ( r.Header.Get("x-forwarded-proto") != "https" && BASEURL != LOCALBASEURL) {
		fmt.Printf("TLS handshake is https=false x-forwarded-proto=%s\n", r.Header.Get("x-forwarded-proto"))
		http.Redirect(w, r, BASEURL, http.StatusFound)
		return
	}

  startPageTemplate.Execute(w, "")
}

/*
	This is the oauth2 callback which will authenticate the user and get the tokent
	A token will last for 3600 seconds and can be used to access the users gmail services.
	We drop the token in the LoginCfg.Exchange return because we don't need for intial login

	Once user is authenticated then create a new session and set maxage to 24 hours. This means
	user will be logged in for 24 hours
*/
func logincallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	tok, err := LoginCfg.Exchange(oauth2.NoContext, code)
  if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		notAuthenticatedTemplate.Execute(w, err.Error())
		return
	}

  // get the users name from google
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
  notAuthenticatedTemplate.Execute(w, template.HTML(err.Error()))
  return
}

if p.Emails == nil {
  w.WriteHeader(http.StatusUnauthorized)
  notAuthenticatedTemplate.Execute(w, template.HTML(fmt.Sprintf("Could not get user profile info: %s", body)))
  return
}

/*
  There is a case where user has an expired ticket
  rootHandler Calls verifyLogin which returns false because of expired cookie error.
  So the user now has to relogin.  So trash the cookie if it exists and start fresh
*/
session, _ := store.Get(r, sessionName)
session.Values["LoggedIn"] = "no"
session.Save(r, w)

session, err = store.Get(r, sessionName)
if err != nil {
  w.WriteHeader(http.StatusUnauthorized)
  notAuthenticatedTemplate.Execute(w, template.HTML(err.Error()))
  return
}

for i := range p.Emails {
		if strings.Contains(p.Emails[i].Value, OauthDomain) {
			session.Values["Email"] = p.Emails[i].Value
			break
		}
	}

	/*
		Check if refresh token was return by google
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
			http.Redirect(w, r, LoginCfg.AuthCodeURL("") + OauthURLParams + "&approval_prompt=force", http.StatusFound)
			return
		}
	} else {
		// we have a refresh token so update databse
		tp = &TokenTuple{0, session.Values["Email"].(string), tok.RefreshToken}
		err = tp.UpdateToken()
		if err != nil {
			fmt.Printf("Token Database Update Error: %s\n", err) // can't get to specific as it could lean to security issue
			w.WriteHeader(http.StatusUnauthorized)
			notAuthenticatedTemplate.Execute(w, template.HTML(err.Error()))
			return
		}
	}

  /*
		Some notes about the access token.
		http://stackoverflow.com/questions/10827920/google-oauth-refresh-token-is-not-being-received

		Basically we need to request offline access to the users google applications so we get a refresh token
		When we attempt to make a google api call we need a access token.  the access token only lasts for 1 hour
		but we allow a users session to last for 24 hours.  So we need google to give us the refresh token.  to force
		the refresh token we need to add &access_type=offline&approval_prompt=force to the auth URL as per above
	*/
  session.Values["LoggedIn"] = "yes"
	session.Values["username"] = p.DisplayName
	session.Values["GCODE"] = code
	session.Values["AuthToken"] = tok
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 43200, // 12 hours even though user can refresh token up to 24 hours.
	}
	session.Save(r, w)

	url := ""
	if r.FormValue("state") != "" {
		url = BASEURL + "/index.html?state=" + r.FormValue("state")
	} else {
		url = BASEURL + "/index.html"
	}

	/*
		We redirect to index.html so we can clean up the users URL info and drop logincallback.
		This will prevent errors should user attemmpt to refresh the page
		We do not want to send the user back to parent because because it could cause an infinite Oauth2 loop
	*/
	http.Redirect(w, r, url, http.StatusFound)

}

/*
	After a successful new login this function will serve the main page
	If user logs in with cookie then roothangler will take of this.
	The only flow i see here is if the user decides to book mark index.html.
	Bookmarking index.html will never allow the user to loging.
	So in error we offer hints on how to login again

*/
func LoginStart(w http.ResponseWriter, r *http.Request) {
	if !verifyLogin(r) {
		http.Redirect(w, r, BASEURL, http.StatusFound)
		return
	}

	// if user is not using https then redirect them
	if ( r.Header.Get("x-forwarded-proto") != "https" && BASEURL != LOCALBASEURL) {
		fmt.Printf("TLS handshake is https=false x-forwarded-proto=%s\n", r.Header.Get("x-forwarded-proto"))
		http.Redirect(w, r, BASEURL, http.StatusFound)
		return
	}

	startPageTemplate.Execute(w, "")
}

/*
	Set the users session cookie key "LoggedIn" to no and redirect user back to
	root page for re-authentication
*/
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	session.Values["LoggedIn"] = "no"
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

/*
	Grabs the users session cookie and verifies login
	return true if logged in and false if not
*/
func verifyLogin(r *http.Request) bool {
	session, err := store.Get(r, sessionName)
	if err != nil {
		fmt.Printf("Failed to get session: %s", err)
		return false
	}
	if session.Values["LoggedIn"] != "yes" {
		return false
	}
	return true
}

/*
  If not logged in write error to ResponseWritter and return false 
  otherwise return true
*/
func checkLogin(w http.ResponseWriter, r *http.Request, fail bool) bool {
  sr := new(statusResponse)
  if !verifyLogin(r) {
    if fail {
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusUnauthorized)
  		w.Write(sr.getJson("Not Authorized"))
    }
    return false
	}
  return true
}

/*
  Parse Get requests and pass them through to handlers
*/
func getHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  sr := new(statusResponse)
  if !checkLogin(w, r, true) {return}
  
  uid, err := getSessionUid(r)
  if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write(sr.getJson("Could not fetch your user id: " + err.Error()))
      return
  }
  
  vals := r.URL.Query()
  if vals.Get("words") == "t" {writeWords(w)}
  if vals.Get("windows") == "t" {writeWindows(w, uid)}
  if vals.Get("submissions") == "t" {writeSubmissionStats(w,vals,uid)}
  if vals.Get("panedata") == "t" {writeJCWindowPanes(w,vals,uid)}
  if vals.Get("user") == "t" {writeUserInfo(w,vals)}
  if vals.Get("history") == "t" {writeHistoryData(w,vals)}
  if vals.Get("previouswindows") == "t" {writePreviousWindows(w,uid)}
}

func writePreviousWindows(w http.ResponseWriter, uid int) {
  sr := new(statusResponse)
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db connection failed: " + err.Error()))
    return
  }
  defer dbi.Close()
  
  type Res struct {
    Pane string `json:"pane"`
    Nickname string `json:"nickname"`
  }
  res := make([]Res,0)
  rows, dberr := dbi.GetRowSet(fmt.Sprintf("SELECT DISTINCT s.session, s.nickname FROM sessions s JOIN subjects sj ON sj.session = s.session WHERE uid = %d ORDER by s.timecreated", uid))
  if dberr != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Query Failed: " + err.Error()))
    return
  }
  for rows.Next() {
    var sess string 
    var nickname string 
    err := rows.Scan(&sess, &nickname)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write(sr.getJson("Scan Failed: " + err.Error()))
      return
    }
    res = append(res, Res{sess, nickname})
  }
  WriteJsonResponse(w, res)
}

/*
  Fetch words from database and write them to ResponseWriter
*/
func writeWords(w http.ResponseWriter) {
  sr := new(statusResponse)
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db connection failed: " + err.Error()))
    return
  }
  defer dbi.Close()
  
  type APIResponse struct {
    Words []string `json:"words"`
  }
  res := APIResponse{}
  
  res.Words, err = dbi.GetStringList("SELECT word FROM words order by 1")
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Fetching words failed: " + err.Error()))
    return
  }
  WriteJsonResponse(w, res)
}

/*
  Fetch and return all the user windows 
*/
func writeWindows(w http.ResponseWriter, uid int) {
  sr := new(statusResponse)
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db connection failed: " + err.Error()))
    return
  }
  defer dbi.Close()
  
  type Window struct {
    CreatedAt time.Time `json:"createdat"`
    Session string `json:"session"`
    Nickname string `json:"nickname"`
  }
  res := make([]Window,0)
  rows, dberr := dbi.GetRowSet(fmt.Sprintf("SELECT s.timecreated, s.session, s.nickname FROM sessions s JOIN subjects sj ON sj.session = s.session WHERE sj.uid = %d order by 1", uid)) 
  if dberr != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Getting sessions fialed: " + dberr.Error()))
    return
  }
  
  for rows.Next() {
    var t time.Time
    var s string 
    var n string 
    err := rows.Scan(&t,&s,&n)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write(sr.getJson("Parsing tuple fialed: " + err.Error()))
      return
    } 
    res = append(res, Window{t, s, n})
  }
  WriteJsonResponse(w, res)
}

func writeSubmissionStats(w http.ResponseWriter, vals url.Values, uid int) {
  sr := new(statusResponse)
  
  sess := vals.Get("pane")
  if sess == "" {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("no session id found url"))
    return
  }
  
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db connection failed: " + err.Error()))
    return
  }
  defer dbi.Close()
  
  type Res struct {
    Submissions int `json:"submissions"`
  }
  res := new(Res)
  res.Submissions, err = dbi.GetIntValue(fmt.Sprintf("select count(*) from (SELECT DISTINCT uid FROM peers p WHERE session = '%s') sub", sess))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db error: " + err.Error()))
    return
  }
  WriteJsonResponse(w,res)
}


/*
  
*/
type WindowPanes struct {
  Arena []string `json:"arena"`
  Blind []string `json:"blind"`
  Facade []string `json:"facade"`
  Unknown []string `json:"unknown"`
}
type JCWindows struct {
  Johari WindowPanes `json:"johari"`
  Clifton WindowPanes `json:"clifton"`
}
func writeJCWindowPanes(w http.ResponseWriter, vals url.Values, uid int) {
  sr := new(statusResponse)
  
  sess := vals.Get("pane")
  if sess == "" {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("no session id found url"))
    return
  }
  
  res := new(JCWindows)
  derr := GetWindowPanesFromDB(res, uid, sess)
  if derr != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Getting panes resturn error: " + derr.Error()))
    return
  }
  WriteJsonResponse(w,res)
}

func GetWindowPanesFromDB(res *JCWindows, uid int, sess string) (error) {
  dbi, err := NewDBI()
  if err != nil {
    return err 
  }
  defer dbi.Close()
  
  jaq := fmt.Sprintf(JOHARI_ARENA_QUERY, sess)
  jbq := fmt.Sprintf(JOHARI_BLIND_QUERY, sess, sess)
  jfq := fmt.Sprintf(JOHARI_FACADE_QUERY, sess, sess)
  juq := fmt.Sprintf(JOHARI_UNKOWN_QUERY, sess, sess)
  caq := fmt.Sprintf(CLIFTON_ARENA_QUERY, sess)
  cbq := fmt.Sprintf(CLIFTON_BLIND_QUERY, sess, sess)
  cfq := fmt.Sprintf(CLIFTON_FACADE_QUERY, sess, sess)
  cuq := fmt.Sprintf(CLIFTON_UNKOWN_QUERY, sess, sess)
  
  res.Johari.Arena, err = dbi.GetStringList(jaq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", jaq, err)
    return err
  }
  res.Johari.Blind, err = dbi.GetStringList(jbq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", jbq, err)
    return err
  }
  res.Johari.Facade, err = dbi.GetStringList(jfq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", jfq, err)
    return err
  }
  res.Johari.Unknown, err = dbi.GetStringList(juq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", juq, err)
    return err
  }
  
  res.Clifton.Arena, err = dbi.GetStringList(caq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", caq, err)
    return err
  }
  res.Clifton.Blind, err = dbi.GetStringList(cbq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", cbq, err)
    return err
  }
  res.Clifton.Facade, err = dbi.GetStringList(cfq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", cfq, err)
    return err
  }
  res.Clifton.Unknown, err = dbi.GetStringList(cuq)
  if err != nil {
    fmt.Printf("Query:%s\nError:%s\n", cuq, err)
    return err
  }
  return nil
}

func writeUserInfo(w http.ResponseWriter, vals url.Values) {
  sr := new(statusResponse)
  
  sess := vals.Get("pane")
  if sess == "" {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("no session id found url"))
    return
  }
  
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db conn error: " + err.Error()))
    return 
  }
  defer dbi.Close()
  
  type Res struct {
    Email string `json:"email"`
  }
  res := new(Res)
  res.Email, err = dbi.GetStringValue(fmt.Sprintf("SELECT DISTINCT u.username FROM users u JOIN subjects s ON s.uid = u.id WHERE s.session = '%s'", sess))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Getting user info Fialed: " + err.Error()))
    return 
  }
  WriteJsonResponse(w,res)
}

func writeHistoryData(w http.ResponseWriter, vals url.Values) {
  sr := new(statusResponse)
  
  sess := vals.Get("pane")
  if sess == "" {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("no session id found url"))
    return
  }
  
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db conn error: " + err.Error()))
    return 
  }
  defer dbi.Close()
  
  type User struct  {
    Themes []string `json:"themes"`
    Words []string `json:"words"`
  }
  type Users struct {
    Users map[string]User `json:"users"`
  }
  
  res := new(Users)
  res.Users = make(map[string]User)
  rows, dberr := dbi.GetRowSet(fmt.Sprintf("SELECT u.username, w.theme, w.word FROM peers p JOIN users u ON u.id = p.uid JOIN words w ON w.wid = p.word WHERE p.session = '%s'", sess))
  if dberr != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Geting submission history failed: " + err.Error()))
    return 
  }
  for rows.Next() {
    var email string 
    var theme string 
    var word string 
    err := rows.Scan(&email,&theme,&word)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write(sr.getJson("bad tuple found in history data: " + err.Error()))
      return 
    }
    _,ok := res.Users[email]
    if ok {
      // workaround bug in golan https://github.com/golang/go/issues/3117 where you can not assign directly map[string]User.Theme = x
      u := res.Users[email]
      u.Themes = append(u.Themes, theme)
      u.Words = append(u.Words, word)
      res.Users[email] = u
    } else {
      res.Users[email] = User{[]string{theme}, []string{word}}
    }
  }
  WriteJsonResponse(w,res)
}


/*
  Parse Get requests and pass them through to handlers
*/
func postHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  sr := new(statusResponse)
  if !checkLogin(w, r, true) {return}
  
  uid, err := getSessionUid(r)
  if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write(sr.getJson("Could not fetch your user id: " + err.Error()))
      return
  }
  
  vals := r.URL.Query()
  if vals.Get("new") == "t" {CreateNewWindow(w,r, uid)}
  if vals.Get("feedback") == "t" {SubmitFeedback(w,r,uid)}
}

func CreateNewWindow(w http.ResponseWriter, r *http.Request, uid int) {
  sr := new(statusResponse)
  type Req struct {
    Nickname string `json:"nickname"`
    Words []string `json:"words"`
  }
  type Res struct {
    Pane string `json:"pane"`
  }
  
  req := new(Req)
  res := new(Res)
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&req)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("decoder error " + err.Error()))
    return
  }
  sessionID := generateSessionID(uid)
  
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db connection failed: " + err.Error()))
    return
  }
  defer dbi.Close()
  
  wids := make([]int,0)
  wids, err = dbi.GetIntList(fmt.Sprintf("SELECT wid FROM words WHERE word in ('%s')", strings.Join(req.Words, "','")))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("resolving word id's failed: " + err.Error()))
    return
  }
  
  stmt, err  := dbi.SQLSession.Prepare("INSERT INTO sessions VALUES(?,?,?)")
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Preparing nickname insert failed: " + err.Error()))
    return
  }
  _, err = stmt.Exec(time.Now().Format("2006-01-02 15:04:05"), sessionID, req.Nickname)
  //_, err = dbi.SQLSession.Exec(fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%s', '%s')", time.Now().Format(time.RFC3339), sessionID, req.Nickname))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Creating nickname failed: " + err.Error()))
    return
  }
  
  query := "INSERT INTO subjects VALUES "
  for i := range wids {
    query += fmt.Sprintf("(%d, '%s', %d),", uid, sessionID, wids[i])
  }
  query = query[:len(query)-1]
  
  _, err = dbi.SQLSession.Exec(query)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Could not create window: " + err.Error()))
    return
  }
  res.Pane = sessionID
  WriteJsonResponse(w, res)
}

/*
  - subjects should not submit feedback to themselves
  - users can not submit multiple feedbacks for the same subject
*/
func SubmitFeedback(w http.ResponseWriter, r *http.Request, uid int) {
  sr := new(statusResponse)
  
  dbi, err := NewDBI()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("db conn error: " + err.Error()))
    return 
  }
  defer dbi.Close()
  
  type Req struct {
    Pane string `json:"pane"`
    Words []string `json:"words"`
  }
  req := new(Req)
  decoder := json.NewDecoder(r.Body)
  err = decoder.Decode(&req)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("decoder error " + err.Error()))
    return
  }
  
  if len(req.Words) <= 0 || req.Pane == "" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write(sr.getJson(fmt.Sprintf("Missing words or session id in submission: %v", req)))
    return
  }
  
  var count int
  count, err = dbi.GetIntValue(fmt.Sprintf("SELECT count(*) FROM ( SELECT * FROM peers WHERE uid = %d AND session = '%s' UNION SELECT * from subjects WHERE uid = %d and session = '%s') sub", uid, req.Pane, uid,req.Pane))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Query error: " + err.Error()))
    return
  }
  if count > 0 {
    w.WriteHeader(http.StatusBadRequest)
    w.Write(sr.getJson("You can not submit feedback to yourself and multiple feedback submission to others are not supported"))
    return
  }
  
  words := make(map[string]int)
  rows, dberr := dbi.GetRowSet(fmt.Sprintf("SELECT wid, word from words where word in ('%s')", strings.Join(req.Words, "','") ))
  if dberr != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(sr.getJson("Query error: " + dberr.Error()))
    return
  }
  for rows.Next() {
    var id int 
    var word string 
    err := rows.Scan(&id,&word)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write(sr.getJson("Scan error: " + err.Error()))
      return
    }
    words[word] = id
  }
  
  q := "INSERT INTO peers VALUES "
  for i := range req.Words {
    q += fmt.Sprintf("(%d, '%s', %d),", uid, req.Pane, words[req.Words[i]])
  }
  q = q[:len(q)-1] // strip the last comma
  
  _,err = dbi.SQLSession.Exec(q)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write(sr.getJson("You can not submit feedback to yourself and multiple feedback submission to others are not supported"))
    return
  }
  
  w.Write(sr.getJson("success"))
}


func windowHandler(w http.ResponseWriter, r *http.Request) {
  if !checkLogin(w, r, false) {
    http.Redirect(w, r, BASEURL, http.StatusFound)
    return
  }
  windowPageTemplate.Execute(w, BASEURL)
}

func feedbackHandler(w http.ResponseWriter, r *http.Request) {
  if !checkLogin(w, r, false) {
    vals := r.URL.Query()
    pane := vals.Get("pane")
    http.Redirect(w, r, BASEURL + "?feedbackpane=" + pane, http.StatusFound)
    return
  }
  feedbackPageTemplate.Execute(w, BASEURL)
}
func thanksHandler(w http.ResponseWriter, r *http.Request) {
  if !checkLogin(w, r, false) {
    http.Redirect(w, r, BASEURL, http.StatusFound)
    return
  }
  thanksPageTemplate.Execute(w, BASEURL)
}



/*
  Get the user name from session and query for uid
*/
func getSessionUid(r *http.Request) (int, error) {
  session, err := store.Get(r, sessionName)
  if err != nil {
    return 0, errors.New(fmt.Sprintf("Failed to get session: %s", err.Error()))
  }
  email := session.Values["Email"]
  if email == nil || email == "" {
    return 0, errors.New(fmt.Sprintf("Could not get email from session cookie"))
  }
  
  dbi, err := NewDBI()
  if err != nil {
    return 0, errors.New("db connection failed: " + err.Error())
  }
  defer dbi.Close()
  return dbi.GetIntValue(fmt.Sprintf("SELECT id FROM users WHERE username = '%s' limit 1", email))
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

/*
	unmarshal ENV variable VCAP_APPLICATION and set the BASEURL string
	default baseurl to http://localhost:8080 for testing
*/
func ParseApllicationCred() {
	VCAP_ENV := os.Getenv("VCAP_APPLICATION")
	LOCALBASEURL = BASEURL + ":" + os.Getenv("PORT")
	if VCAP_ENV == "" {
		fmt.Println("VCAP_APPLICATION ENV variable not found")
		BASEURL += ":" + os.Getenv("PORT")
		fmt.Printf("Using url %s for callback\n", BASEURL)
		return
	}
	fmt.Printf("%v\n", VCAP_ENV)

	type VCAP_APP struct {
		URIs []string `json:"uris"`
	}
	MyApp := new(VCAP_APP)

	err := json.Unmarshal([]byte(VCAP_ENV), &MyApp)
	if err != nil {
		fmt.Printf("Failed to decode VCAP_APP: %s\n", err)
		return
	}

	for i := range MyApp.URIs {
		BASEURL = "https://" + MyApp.URIs[i]
		break
	}
	fmt.Printf("Using url %s for callback\n", BASEURL)
}

func ParseServiceCred() {
  VCAP_ENV := os.Getenv("VCAP_SERVICES")
  if VCAP_ENV == "" {
		fmt.Println("VCAP_SERVICES ENV variable not found")
		fmt.Printf("Using DBURL %s\n", os.Getenv("DBURL"))
    DBURL = os.Getenv("DBURL")
		return
	}
  type Cred struct {
    Uri string `json:"uri"`
    Hostname string `json:"hostname"`
    Port int `json:"port"`
    Database string `json:"name"`
    User string `json:"username"`
    Pass string `json:"password"`
  }
  type Obj struct {
    Credentials Cred `json:"credentials"`
  }
  type DBService struct {
    SQLService []Obj `json:"p-mysql"`
  }
  MyService := new(DBService)
  
  err := json.Unmarshal([]byte(VCAP_ENV), &MyService)
	if err != nil {
		fmt.Printf("Failed to decode VCAP_SERVICES: %s\n", err)
		return
	}
  // only care about the first one found because we hard coded it with p-mysql in DBService struct
  for i := range  MyService.SQLService {
    DBURL = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", 
      MyService.SQLService[i].Credentials.User, 
      MyService.SQLService[i].Credentials.Pass, 
      MyService.SQLService[i].Credentials.Hostname, 
      MyService.SQLService[i].Credentials.Port,
      MyService.SQLService[i].Credentials.Database)
    break
  }
  if DBURL == "" {
    fmt.Println("ERROR DBURL is not set!!")
  }
}

func main() {
  gob.Register(&oauth2.Token{})
  ParseApllicationCred()
  ParseServiceCred()


  CLIENTID = os.Getenv("CLIENTID")
	CLIENTSECRET = os.Getenv("CLIENTSECRET")
  OauthURLParams = os.Getenv("OAUTHURLPARAMS")
  OauthDomain = os.Getenv("OAUTHDOMAIN")
  //DBURL = os.Getenv("DBURL")

  LoginCfg = &oauth2.Config{
		ClientID:     CLIENTID,
		ClientSecret: CLIENTSECRET,
		RedirectURL:  BASEURL + "/logincallback",
		Scopes:       []string{"profile", "email", "https://www.googleapis.com/auth/gmail.compose"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

  fmt.Printf("Going to use port %s\n", os.Getenv("PORT"))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index.html", LoginStart)
	http.HandleFunc("/logincallback", logincallbackHandler)
	http.HandleFunc("/logout/", logoutHandler)

  // api Calls
  http.HandleFunc("/get", getHandler)
  http.HandleFunc("/post", postHandler)
  
  http.HandleFunc("/window", windowHandler)
  http.HandleFunc("/feedback", feedbackHandler)
  http.HandleFunc("/thanks", thanksHandler)
  
  // File serving handlers
	http.Handle("/img/", http.FileServer(http.Dir("")))
  http.Handle("/fonts/", http.FileServer(http.Dir("")))
	http.Handle("/js/", http.FileServer(http.Dir("")))
	http.Handle("/css/", http.FileServer(http.Dir("")))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		fmt.Printf("Failed to start http server: %s\n", err)
	}
	return
}
