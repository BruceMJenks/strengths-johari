package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

var (
	cookieStore    *sessions.CookieStore
	sessionName    string
	clientID       string
	clientSecret   string
	oauthURLParams string
	oauthDomain    string
	dbURL          string
	authURL        string
	tokenURL       string
	encryptionKey  string
	baseURL        = "http://localhost"
	localBaseURL   string // used for testing purposes
	loginCfg       *oauth2.Config

	mainPage                 = template.Must(template.ParseFiles("tmpl/mainPage.tmpl"))
	mainTemplates            = template.Must(template.ParseFiles("tmpl/mainTemplates.tmpl"))
	windowPage               = template.Must(template.ParseFiles("tmpl/windowPage.tmpl"))
	feedbackPage             = template.Must(template.ParseFiles("tmpl/feedbackPage.tmpl"))
	thanksPage               = template.Must(template.ParseFiles("tmpl/thanksPage.tmpl"))
	notAuthenticatedTemplate = template.Must(template.ParseFiles("tmpl/noPermissionPage.tmpl"))
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
	Grabs the users session cookie and verifies login
	return true if logged in and false if not
*/
func verifyLogin(r *http.Request) bool {

	session, err := cookieStore.Get(r, sessionName)
	if err != nil {
		fmt.Printf("Failed to get session: %s", err)
		return false
	}
	// if our session has expired then re-login
	tok := session.Values["AuthToken"].(*oauth2.Token)
	if (tok.Expiry.Unix() - time.Now().Unix()) < 0 {
		fmt.Printf("%s:%s: session expired\n", r.Method, r.URL.Path)
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
	unmarshal ENV variable VCAP_APPLICATION and set the BASEURL string
	default baseurl to http://localhost:8080 for testing
*/
func ParseApllicationCred() {
	vcapENV := os.Getenv("VCAP_APPLICATION")
	localBaseURL = baseURL + ":" + os.Getenv("PORT")
	if vcapENV == "" {
		fmt.Println("VCAP_APPLICATION ENV variable not found")
		baseURL += ":" + os.Getenv("PORT")
		fmt.Printf("Using url %s for callback\n", baseURL)
		return
	}
	fmt.Printf("%v\n", vcapENV)

	type VCAP_APP struct {
		URIs []string `json:"uris"`
	}
	MyApp := new(VCAP_APP)

	err := json.Unmarshal([]byte(vcapENV), &MyApp)
	if err != nil {
		fmt.Printf("Failed to decode VCAP_APP: %s\n", err)
		return
	}

	for i := range MyApp.URIs {
		baseURL = "https://" + MyApp.URIs[i]
		break
	}
	fmt.Printf("Using url %s for callback\n", baseURL)
}

func ParseServiceCred() {
	vcapENV := os.Getenv("VCAP_SERVICES")
	if vcapENV == "" {
		fmt.Println("VCAP_SERVICES ENV variable not found")
		fmt.Printf("Using DBURL %s\n", os.Getenv("DBURL"))
		dbURL = os.Getenv("DBURL")
		return
	}
	type Cred struct {
		URI      string `json:"uri"`
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Database string `json:"name"`
		User     string `json:"username"`
		Pass     string `json:"password"`
	}
	type Obj struct {
		Credentials Cred `json:"credentials"`
	}
	type DBService struct {
		SQLService []Obj `json:"p-mysql"`
	}
	MyService := new(DBService)

	err := json.Unmarshal([]byte(vcapENV), &MyService)
	if err != nil {
		fmt.Printf("Failed to decode VCAP_SERVICES: %s\n", err)
		return
	}
	// only care about the first one found because we hard coded it with p-mysql in DBService struct
	for i := range MyService.SQLService {
		dbURL = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			MyService.SQLService[i].Credentials.User,
			MyService.SQLService[i].Credentials.Pass,
			MyService.SQLService[i].Credentials.Hostname,
			MyService.SQLService[i].Credentials.Port,
			MyService.SQLService[i].Credentials.Database)
		break
	}
	if dbURL == "" {
		fmt.Println("ERROR DBURL is not set!!")
	}
}

func newRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/index.html", LoginStart).Methods("GET")
	r.HandleFunc("/logincallback", logincallbackHandler).Methods("GET")
	r.HandleFunc("/logout/", logoutHandler).Methods("GET")

	// api Calls
	r.HandleFunc("/get", getHandler).Methods("GET")
	r.HandleFunc("/post", postHandler).Methods("GET")

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
	ParseApllicationCred()
	ParseServiceCred()

	clientID = os.Getenv("CLIENTID")
	clientSecret = os.Getenv("CLIENTSECRET")
	oauthURLParams = os.Getenv("OAUTHURLPARAMS")
	oauthDomain = os.Getenv("OAUTHDOMAIN")
	sessionName = os.Getenv("SESSION_NAME")
	cookieStore = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_STORE_KEY")))
	encryptionKey = os.Getenv("PRIVATE_ENCRYPTION_KEY")
	//DBURL = os.Getenv("DBURL")

	loginCfg = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/logincallback",
		Scopes:       []string{"profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	fmt.Printf("Going to use port %s\n", os.Getenv("PORT"))

	err := http.ListenAndServe(":"+os.Getenv("PORT"), context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		fmt.Printf("Failed to start http server: %s\n", err)
	}

	r := newRouter()
	err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
	if err != nil {
		panic(err)
	}

	return
}
