package main

import (
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

var (
	DBI            *DBInstance
	CookieStore    *sessions.CookieStore
	SessionName    string
	ClientID       string
	ClientSecret   string
	OauthURLParams string
	OauthDomain    string
	AuthURL        string
	TokenURL       string
	EncryptionKey  string
	baseURL        = "http://localhost"
	localBaseURL   string // used for testing purposes
	LoginCfg       *oauth2.Config

	MainPage                 = template.Must(template.ParseFiles("tmpl/mainPage.tmpl", "tmpl/mainTemplates.tmpl"))
	MainTemplates            = template.Must(template.ParseFiles("tmpl/mainTemplates.tmpl"))
	WindowPage               = template.Must(template.ParseFiles("tmpl/windowPage.tmpl", "tmpl/mainTemplates.tmpl"))
	FeedbackPage             = template.Must(template.ParseFiles("tmpl/feedbackPage.tmpl", "tmpl/mainTemplates.tmpl"))
	ThanksPage               = template.Must(template.ParseFiles("tmpl/thanksPage.tmpl", "tmpl/mainTemplates.tmpl"))
	LoginPage                = template.Must(template.ParseFiles("tmpl/loginPage.tmpl", "tmpl/mainTemplates.tmpl"))
	NotAuthenticatedTemplate = template.Must(template.ParseFiles("tmpl/noPermissionPage.tmpl", "tmpl/mainTemplates.tmpl"))

	EnableOauth = flag.Bool("oauth", false, "Flag to disable authentication")
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
func (sr *statusResponse) getJSON(s string) []byte {
	sr.Status = s
	jsr, _ := json.Marshal(sr)
	return jsr
}

// WriteJSONResponse conversts struct into json object and writes it back to responseWriter
func WriteJSONResponse(w http.ResponseWriter, data interface{}) {
	sr := new(statusResponse)
	jdata, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(sr.getJSON("Failed to marshal data: " + err.Error()))
		return
	}

	w.Write(jdata)
	return
}

/*
	Grabs the users session cookie and verifies login
	return true if logged in and false if not
*/
func verifyLogin(r *http.Request) bool {

	session, err := CookieStore.Get(r, SessionName)
	if err != nil {
		fmt.Printf("Failed to get session: %s", err)
		return false
	}
	// if our session has expired then re-login

	if *EnableOauth {
		if session.Values["AuthToken"] == nil {
			fmt.Printf("No Token Found in Session")
			return false
		}

		tok := session.Values["AuthToken"].(*oauth2.Token)
		if (tok.Expiry.Unix() - time.Now().Unix()) < 0 {
			fmt.Printf("%s:%s: session expired\n", r.Method, r.URL.Path)
			return false
		}
	} else {
		if email, ok := session.Values["Email"].(string); ok {
			_, err := DBI.GetStringValue(fmt.Sprintf(SELECT_USERNAME_QUERY, email))
			if err != nil {
				fmt.Printf("User not found in database: %s\n", err)
				return false
			}
		} else {
			return false
		}
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
			w.Write(sr.getJSON("Not Authorized"))
		}
		return false
	}
	return true
}

// checkAuthorization verify the user request information has permission to access the session
func checkAuthorization(w http.ResponseWriter, vals url.Values, uid int) bool {
	sr := new(statusResponse)
	_, err := DBI.GetStringValue(fmt.Sprintf(SELECT_SUBJECT_SESSION_QUERY, uid, vals.Get("pane")))
	if err != nil {
		fmt.Printf("Failed to find and authorize session: %s\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(sr.getJSON("You are not authorized to access this window pane"))
		return false
	}
	return true
}

/*
	unmarshal ENV variable VCAP_APPLICATION and set the BASEURL string
	default baseurl to http://localhost:8080 for testing
*/
// ParseApllicationCred unmarshal ENV variable VCAP_APPLICATION and set the baseURL string
func ParseApllicationCred() (myBaseURL string) {
	vcapENV := os.Getenv("VCAP_APPLICATION")
	localBaseURL = "http://localhost:" + os.Getenv("PORT")
	if vcapENV == "" {
		fmt.Println("VCAP_APPLICATION ENV variable not found")
		myBaseURL = localBaseURL
		fmt.Printf("Using url %s for callback\n", myBaseURL)
		return myBaseURL
	}

	type VCAP_APP struct {
		URIs []string `json:"uris"`
	}
	MyApp := new(VCAP_APP)

	err := json.Unmarshal([]byte(vcapENV), &MyApp)
	if err != nil {
		fmt.Printf("Failed to decode VCAP_APP: %s\n", err)
		return myBaseURL
	}

	for i := range MyApp.URIs {
		myBaseURL = "https://" + MyApp.URIs[i]
		break
	}
	fmt.Printf("Using url %s for callback\n", myBaseURL)
	return myBaseURL
}

// ParseServiceCred returns the database connection string
func ParseServiceCred() (dbURL string) {
	vcapENV := os.Getenv("VCAP_SERVICES")
	if vcapENV == "" {
		fmt.Println("VCAP_SERVICES ENV variable not found")
		fmt.Printf("Using DBURL %s\n", os.Getenv("DBURL"))
		dbURL = os.Getenv("DBURL")
		return
	}
	type Cred struct {
		URI      string      `json:"uri"`
		Hostname string      `json:"hostname"`
		Port     interface{} `json:"port"`
		Database string      `json:"name"`
		User     string      `json:"username"`
		Pass     string      `json:"password"`
	}
	type Obj struct {
		Credentials Cred `json:"credentials"`
	}
	type DBService struct {
		SQLService []Obj `json:"p-mysql"`
		CLearDB    []Obj `json:"cleardb"`
	}
	MyService := new(DBService)

	err := json.Unmarshal([]byte(vcapENV), &MyService)
	if err != nil {
		fmt.Printf("Failed to decode VCAP_SERVICES: %s\n", err)
		return
	}
	// only care about the first one found because we hard coded it with p-mysql in DBService struct
	if len(MyService.SQLService) > 0 {
		for i := range MyService.SQLService {
			dbURL = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
				MyService.SQLService[i].Credentials.User,
				MyService.SQLService[i].Credentials.Pass,
				MyService.SQLService[i].Credentials.Hostname,
				MyService.SQLService[i].Credentials.Port.(int),
				MyService.SQLService[i].Credentials.Database)
			break
		}
	} else {
		for i := range MyService.CLearDB {
			dbURL = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
				MyService.CLearDB[i].Credentials.User,
				MyService.CLearDB[i].Credentials.Pass,
				MyService.CLearDB[i].Credentials.Hostname,
				MyService.CLearDB[i].Credentials.Port.(string),
				MyService.CLearDB[i].Credentials.Database)
			break
		}
	}

	if dbURL == "" {
		fmt.Println("ERROR DBURL is not set!!")
	}
	return dbURL
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/index.html", LoginStart).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/login/submit", submitLoginHandler).Methods("POST")
	r.HandleFunc("/login/register", registerUserHandler).Methods("POST")
	r.HandleFunc("/logincallback", logincallbackHandler).Methods("GET")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")

	// api Calls
	r.HandleFunc("/get", getHandler).Methods("GET")
	r.HandleFunc("/post", postHandler).Methods("POST")

	r.HandleFunc("/window", windowHandler).Methods("GET")
	r.HandleFunc("/feedback", feedbackHandler).Methods("GET")
	r.HandleFunc("/thanks", thanksHandler).Methods("GET")

	staticFileDirectory := http.Dir("./assets/")
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")
	return r
}

func main() {
	gob.Register(&oauth2.Token{})
	baseURL = ParseApllicationCred()

	var err error
	DBI, err = NewDBI(ParseServiceCred())
	if err != nil {
		fmt.Println("Failed to connect to database")
		panic(err)
	}
	defer DBI.Close()
	err = DBI.CreateSchema()
	if err != nil {
		panic(err)
	}

	ClientID = os.Getenv("CLIENTID")
	ClientSecret = os.Getenv("CLIENTSECRET")
	OauthURLParams = os.Getenv("OAUTHURLPARAMS")
	OauthDomain = os.Getenv("OAUTHDOMAIN")
	SessionName = os.Getenv("SESSION_NAME")
	CookieStore = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_STORE_KEY")))
	EncryptionKey = os.Getenv("PRIVATE_ENCRYPTION_KEY")
	AuthURL = os.Getenv("AUTH_URL")
	TokenURL = os.Getenv("TOKEN_URL")
	LoginCfg = &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		RedirectURL:  baseURL + "/logincallback",
		Scopes:       []string{"profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}

	fmt.Printf("Going to use port %s\n", os.Getenv("PORT"))

	/*err = http.ListenAndServe(":"+os.Getenv("PORT"), context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		fmt.Printf("Failed to start http server: %s\n", err)
	}*/

	r := NewRouter()
	err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
	if err != nil {
		panic(err)
	}

	return
}
