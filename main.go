//The entry point for the gochat program
package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"github.com/kevensen/openshift-gochat-client/gomniauth/providers/openshift"
	"github.com/koding/websocketproxy"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"

	openshiftV1 "github.com/openshift/api/user/v1"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

//Primary handler
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join(*templatePath, t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)

}

var Users map[string]openshiftV1.User
var UserTokens map[string]string

func readToken() string {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		if err != nil {
			glog.Errorln(err)
			return ""
		}
	} else {
		return ""
	}
	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		glog.Errorln(err)
		return ""
	}
	return string(token)

}

var openshiftApiHost *string
var openshiftNamespace *string
var templatePath *string
var openshiftRegistry *string
var allowInsecure *bool

/*
* Main entry point.  Flags, Handlers, and authentication providers configured here.
*
 */
func main() {
	var host = flag.String("host", "localhost:8080", "The host address of the application.")
	templatePath = flag.String("templatePath", "templates/", "The path to the HTML templates.  This is relative to the location from which \"gochat\" is executed.  Can be absolute.")
	openshiftApiHost = flag.String("openshiftApiHost", "172.30.0.1", "The location of the OpenShift API.")
	openshiftNamespace = flag.String("project", os.Getenv("OPENSHIFT_BUILD_NAMESPACE"), "The current working project.")
	openshiftRegistry = flag.String("registry", "docker-registry.default.svc:5000", "The location of the container registry.")
	var chatServer = flag.String("chatServer", "localhost:8081", "The location of the OpenShift Gochat Server")
	var serviceAccount = flag.String("serviceAccount", "default", "The service account to talk to the OpenShift API")
	var serviceAccountToken = flag.String("serviceAccountToken", readToken(), "The service account token.")
	allowInsecure = flag.Bool("insecure", false, "Allow insecure TLS connections")
	flag.Parse()
	Users = make(map[string]openshiftV1.User)
	UserTokens = make(map[string]string)

	authServerMetadata := openshift.NewOAuthServerMetadata()
	gomniauth.SetSecurityKey("Aua1nuYA1C0ANLdrJalRDSPc0hl8MhfO903hC9cJRb4E2pA76PRcT6bQSveW2kYH")
	gomniauth.WithProviders(
		openshift.New("system:serviceaccount:"+*openshiftNamespace+":"+*serviceAccount,
			*serviceAccountToken, *openshiftNamespace, authServerMetadata))

	myAuthHandler := NewAuthHandler(
		*serviceAccount,
		*serviceAccountToken,
		authServerMetadata.AuthorizationEndpoint,
		authServerMetadata.TokenEndpoint,
		&templateHandler{filename: "chat.html"})

	glog.Infoln("Registry", openshiftRegistry)

	http.Handle("/", myAuthHandler)
	http.Handle("/chat", myAuthHandler)
	http.Handle("/denied", &templateHandler{filename: "denied.html"})
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", myAuthHandler.loginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/logoutpage"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	http.Handle("/logoutpage", &templateHandler{filename: "logoutpage.html"})

	chatServerURL, err := url.Parse("ws://" + *chatServer)
	glog.Infoln("The backend is", *chatServer)
	if err != nil {
		glog.Errorln(err)
	}
	http.Handle("/room", websocketproxy.ProxyHandler(chatServerURL))

	http.HandleFunc("/roll", RollDiceHandler)

	glog.Infoln("Starting the web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		glog.Fatalln("ListenAndServe: ", err)
	}

}
