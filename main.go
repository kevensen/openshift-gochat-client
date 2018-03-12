//The entry point for the gochat program
package main

import (
	"flag"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"github.com/koding/websocketproxy"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

var templatePath *string

//Primary handler
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join(*templatePath, t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	glog.Infoln("Server Host:", data["Host"])
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)

}

var OpenshiftApiHost *string
var OpenshiftNamespace *string
var OpenshiftRegistry *string
var Users map[string]User

/*
* Main entry point.  Flags, Handlers, and authentication providers configured here.
*
 */
func main() {
	var host = flag.String("host", "localhost:8080", "The host address of the application.")
	templatePath = flag.String("templatePath", "templates/", "The path to the HTML templates.  This is relative to the location from which \"gochat\" is executed.  Can be absolute.")
	OpenshiftApiHost = flag.String("openshiftApiHost", "172.30.0.1", "The location of the OpenShift API.")
	OpenshiftNamespace = flag.String("project", os.Getenv("OPENSHIFT_BUILD_NAMESPACE"), "The current working project.")
	OpenshiftRegistry = flag.String("registry", "docker-registry.default.svc:5000", "The location of the container registry.")
	var chatServer = flag.String("chatServer", "localhost:8081", "The location of the OpenShift Gochat Server")
	flag.Parse()
	Users = make(map[string]User)

	myAuthHandler := new(authHandler)
	myAuthHandler.next = &templateHandler{filename: "chat.html"}

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
	http.HandleFunc("/roll", RollDice)

	glog.Infoln("Starting the web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		glog.Fatalln("ListenAndServe: ", err)
	}

}
