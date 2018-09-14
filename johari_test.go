package main_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"

	. "github.com/pivotal-gss/johari"
)

func createUser(u string) {
	baseURL := "http://localhost:" + os.Getenv("PORT")
	hc := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", baseURL+"/login/register", bytes.NewBuffer([]byte(u)))
	req.Header.Add("X-Forwarded-Proto", "https")
	resp, err := hc.Do(req)
	if err != nil {
		Fail(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		b, _ := ioutil.ReadAll(resp.Body)
		Fail(fmt.Sprintf("Create new user failed with status code: %d: %s: %s", resp.StatusCode, b, u))
	}
}

var mockHTTPServer *http.Server
var _ = BeforeSuite(func() {

	gob.Register(&oauth2.Token{})
	r := NewRouter()
	mockHTTPServer = &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
	}

	ClientID = os.Getenv("CLIENTID")
	ClientSecret = os.Getenv("CLIENTSECRET")
	CookieStore = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_STORE_KEY")))
	SessionName = os.Getenv("SESSION_NAME")
	EncryptionKey = os.Getenv("PRIVATE_ENCRYPTION_KEY")
	AuthURL = os.Getenv("AUTH_URL")
	TokenURL = os.Getenv("TOKEN_URL")
	LoginCfg = &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		RedirectURL:  "http://localhost/logincallback",
		Scopes:       []string{"profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	go func() {
		err := mockHTTPServer.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	myDBURL := os.Getenv("DBURL")
	if myDBURL == "" {
		Fail("Must have DBURL environment variable set in order to connect to the database")
	}

	var err error
	DBI, err = NewDBI(myDBURL)
	if err != nil {
		Fail(err.Error())
	}
	_, err = DBI.SQLSession.Exec("CREATE DATABASE jhandlerdb")
	if err != nil {
		Fail(err.Error())
	}
	_, err = DBI.SQLSession.Exec("use jhandlerdb")
	if err != nil {
		Fail(err.Error())
	}
	err = DBI.CreateSchema()
	if err != nil {
		Fail(err.Error())
	}
	time.Sleep(1 * time.Second) // give time for http to start up

	// Create test users
	createUser("{ \"user\": \"testuser1@user.email\", \"password\": \"Y2hhbmdlbWU=\"}")
	createUser("{ \"user\": \"%s\", \"testuser2@user.email\": \"cGFzc3dvcmQ=\"}")

})

var _ = AfterSuite(func() {
	_, err := DBI.SQLSession.Exec("DROP DATABASE jhandlerdb")
	if err != nil {
		Fail(err.Error())
	}
	DBI.Close()
})

var _ = Describe("API Handlers", func() {
	defer GinkgoRecover()

	var baseURL string
	var hc *http.Client
	var cookie *http.Cookie

	BeforeEach(func() {

		mockUser := "testuser1@user.email"
		mockUserPass := "Y2hhbmdlbWU="
		userBody := fmt.Sprintf("{ \"user\": \"%s\", \"password\": \"%s\"}", mockUser, mockUserPass)
		baseURL = "http://localhost:" + os.Getenv("PORT")
		hc = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		// Authenticate and save cookie
		req, err := http.NewRequest("POST", baseURL+"/login/submit", bytes.NewBuffer([]byte(userBody)))
		req.Header.Add("X-Forwarded-Proto", "https")
		resp, err := hc.Do(req)
		if err != nil {
			Fail(err.Error())
		}
		if resp.StatusCode != 302 {
			b, _ := ioutil.ReadAll(resp.Body)
			Fail(fmt.Sprintf("Response status is not 302: %d : %s", resp.StatusCode, b))
		}

		for _, c := range resp.Cookies() {
			if c.Name == SessionName {
				cookie = c // save the cookie for future use
			}
		}

		if cookie == nil {
			Fail("Could not acquire auth cookie")
		}
	})

	Describe("testing the johari handlers", func() {
		Context("when internal auth is enabled", func() {

			It("Root handler returns 200 response", func() {
				req, err := http.NewRequest("GET", baseURL, nil)
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				resp, err := hc.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))
			})

			It("/logout sets a cookie and returns 302", func() {
				req, err := http.NewRequest("GET", baseURL+"/logout", nil)
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				resp, err := hc.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(302))
				Expect(resp.Header.Get("Set-Cookie")).ShouldNot(Equal(""))
			})

			It("creating and sharing window works", func() {

				var user1Pane string
				//var user2Cookie *http.Cookie

				By("/post?new=t creates a new window pane")
				req, err := http.NewRequest("POST", baseURL+"/post?new=t", bytes.NewBuffer([]byte(`{ "nickname": "test-sample12345", "words": ["Aware","Inquisitive","Self-motivated","Driven","Meticulous","Vivid","Artistic","Serious","Questioning","Impatient","Collaborative"]}`)))
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				Expect(err).ShouldNot(HaveOccurred())
				resp, err := hc.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(201))
				data := new(CreateWindowRes)
				decoder := json.NewDecoder(resp.Body)
				Expect(decoder.Decode(&data)).ShouldNot(HaveOccurred()) // no err expected
				Expect(data.Pane).ShouldNot(BeEmpty())
				//user1Pane = data.Pane

				By("/window?pane=ID returns the users window pane")
				req, err = http.NewRequest("GET", baseURL+"/window?pane="+data.Pane, nil)
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				resp, err = hc.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				By("/get?previouswindows=t should populate previous windows")
				req, err = http.NewRequest("GET", baseURL+"/get?previouswindows=t", nil)
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				resp, err = hc.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))
				wpdata := make([]WritePreiviouWindowRes, 0)
				decoder = json.NewDecoder(resp.Body)
				Expect(decoder.Decode(&wpdata)).ShouldNot(HaveOccurred())
				Expect(len(wpdata)).NotTo(BeZero())

				By("/get/submissions=t&pane=xxx")

				By("/get?panedata=t&pane=xxx")
				By("/get?history=t&pane=xxx")

				By("a new user opens the shared pane")

				By("the new user can not access the other users pane")
			})
		})
	})

})

var _ = Describe("Database", func() {

	myDBURL := os.Getenv("DBURL")
	if myDBURL == "" {
		Fail("Must have DBURL environment variable set in order to connect to the database")
	}
	var dbi *DBInstance
	BeforeEach(func() {
		var err error
		dbi, err = NewDBI(myDBURL)
		if err != nil {
			Fail(err.Error())
		}
	})
	Context("When database is empty", func() {
		BeforeEach(func() {
			dbi.SQLSession.Exec(DROP_PEERS_TABLE)
			dbi.SQLSession.Exec(DROP_USERS_TABLE)
			dbi.SQLSession.Exec(DROP_WORDS_TABLE)
			dbi.SQLSession.Exec(DROP_SUBJECTS_TABLE)
			dbi.SQLSession.Exec(DROP_SESSIONS_TABLE)
		})
		It("Creates the database tables", func() {
			Expect(dbi.CreateSchema()).To(BeNil())
			v, err := dbi.GetIntValue(SELECT_WORDS_TABLE)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).ShouldNot(BeZero())
		})
	})

	Context("When Database is not empty", func() {
		It("Creating schema tables does not error", func() {
			Expect(dbi.CreateSchema()).To(BeNil())
			Expect(dbi.CreateSchema()).To(BeNil())
		})
	})
})

var _ = Describe("VCAP variables", func() {
	defer GinkgoRecover()

	var mockVCAPServices = ""
	var mockVCAPApplication = ""
	BeforeEach(func() {

		mockVCAPApplication = `{
			"cf_api": "https://donotuseapi.run.io",
			"limits": {
			  "fds": 16384
			},
			"application_name": "johari",
			"application_uris": [
			  "johari.io"
			],
			"name": "johari",
			"space_name": "SPACE",
			"space_id": "MY-SPACE-GUID",
			"uris": [
			  "johari.io"
			],
			"users": null,
			"application_id": "MY-APP-GUID"
		  }`
		mockVCAPServices = `{
			"p-mysql": [
			  {
				"name": "johari-mysql",
				"instance_name": "johari-mysql",
				"binding_name": null,
				"credentials": {
				  "hostname": "p-mysql-proxy.run.io",
				  "port": 3306,
				  "name": "cf_database_guid",
				  "username": "user",
				  "password": "password",
				  "uri": "mysql://user:password@p-mysql-proxy.run.io:3306/cf_database_guid?reconnect=true",
				  "jdbcUrl": "jdbc:mysql://p-mysql-proxy.run.io:3306/cf_cf_database_guid?user=user&password=password"
				},
				"syslog_drain_url": null,
				"volume_mounts": [],
				"label": "p-mysql",
				"provider": null,
				"plan": "100mb",
				"tags": [
				  "mysql"
				]
			  }
			]
		  }`
	})

	JustBeforeEach(func() {
		os.Setenv("DBURL", "")
		os.Setenv("VCAP_SERVICES", mockVCAPServices)
		os.Setenv("VCAP_APPLICATION", mockVCAPApplication)
	})

	Describe("Parsing VCAP_SERVICES environment", func() {
		Context("when VCAP_SERVICES has json string", func() {
			It("Should build the dbURL connection string", func() {
				dbURL := ParseServiceCred()
				Expect(dbURL).To(Equal("user:password@tcp(p-mysql-proxy.run.io:3306)/cf_database_guid?parseTime=true"))
			})
		})

		Context("when VCAP_APPLICATION has json string", func() {
			It("Should correctly return my baseURL", func() {
				baseURL := ParseApllicationCred()
				Expect(baseURL).To(Equal("https://johari.io"))
			})
		})
	})
})
