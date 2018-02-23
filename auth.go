package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type authHandler struct {
	next    http.Handler
	apiHost string
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == "" {
		//not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		h.next.ServeHTTP(w, r)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]

	log.Println("Action", action, "Provider", provider)

}

func MustAuth(handler http.Handler, openshiftApiHost string) http.Handler {
	fmt.Println("OpenShift API Host being set in MustAuth is", openshiftApiHost)
	return &authHandler{next: handler, apiHost: openshiftApiHost}
}
