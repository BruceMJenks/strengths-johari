package main_test

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"

	. "github.com/pivotal-gss/johari"
)

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
	time.Sleep(1 * time.Second) // give time for http to star tup

	// setup databsae connection
	myDBURL := os.Getenv("DBURL")
	if myDBURL == "" {
		Fail("Must have DBURL environment variable set in order to connect to the database")
	}
})

var _ = AfterSuite(func() {
})

var _ = Describe("API Handlers", func() {
	defer GinkgoRecover()

	var baseURL string
	var hc *http.Client
	var sc *securecookie.SecureCookie
	var cookie *http.Cookie
	BeforeEach(func() {
		baseURL = "http://localhost:" + os.Getenv("PORT")
		hc = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		sc = securecookie.New([]byte(EncryptionKey), []byte(EncryptionKey))
		values := map[string]string{
			"Email":     "testuser@user.email",
			"AuthToken": "xxx",
		}
		encoded, err := sc.Encode(SessionName, values)
		if err != nil {
			Fail(err.Error())
		}
		cookie = &http.Cookie{
			Name:  SessionName,
			Value: encoded,
			Path:  "/",
		}
	})
	JustBeforeEach(func() {})
	Describe("testing the johari handlers", func() {
		Context("when auth is disabled", func() {
			*DisableAuth = true
			It("Root handler returns 200 response", func() {
				req, err := http.NewRequest("GET", baseURL, nil)
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				resp, err := hc.Do(req)
				if err != nil {
					Fail(err.Error())
				}
				Expect(resp.StatusCode).To(Equal(200))
			})

			It("/logout sets a cookie and returns 302", func() {
				req, err := http.NewRequest("GET", baseURL+"/logout", nil)
				req.Header.Add("X-Forwarded-Proto", "https")
				req.AddCookie(cookie)
				resp, err := hc.Do(req)
				if err != nil {
					Fail(err.Error())
				}
				Expect(resp.StatusCode).To(Equal(302))
				Expect(resp.Header.Get("Set-Cookie")).ShouldNot(Equal(""))
			})
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
