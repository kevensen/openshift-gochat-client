//The entry point for the gochat program
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

var templatePath *string
var HtpasswdPath *string

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

/*
* Main entry point.  Flags, Handlers, and authentication providers configured here.
*
 */
func main() {
	var host = flag.String("host", ":8080", "The host address of the application.")
	templatePath = flag.String("templatePath", "templates/", "The path to the HTML templates.  This is relative to the location from which \"gochat\" is executed.  Can be absolute.")
	var openshiftApiHost = flag.String("openshiftApiHost", "172.30.0.1", "The location of the OpenShift API.")
	flag.Parse()

	r := newRoom()
	http.Handle("/", MustAuth(&templateHandler{filename: "chat.html"}, *openshiftApiHost))
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}, *openshiftApiHost))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)

	go r.run()
	log.Println("Starting the web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
