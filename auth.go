package main

import (
	"crypto/md5"
	"io"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/stretchr/objx"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == "" {
		//not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		userName := objx.MustFromBase64(cookie.Value)["name"].(string)
		if _, ok := Users[userName]; !ok {
			w.Header().Set("Location", "/logoutpage")
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			h.next.ServeHTTP(w, r)
		}
	}
}

func (h *authHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.PostFormValue("token")
	user := new(User)
	user.token = token

	err, status := user.login()

	if err != nil {
		glog.Fatalln("Error Logging into OpenShift:", err)
	}

	if status == 403 || status == 401 {
		w.Header()["Location"] = []string{"/denied"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if status != 200 {
		glog.Fatalln("Error Logging into OpenShift resulted in status:", status)
	}

	m := md5.New()
	io.WriteString(m, strings.ToLower(user.Metadata.Name))

	authCookieValue := objx.New(map[string]interface{}{
		"name": user.Metadata.Name,
	}).MustBase64()

	http.SetCookie(w, &http.Cookie{
		Name:  "auth",
		Value: authCookieValue,
		Path:  "/",
	})
	Users[user.Metadata.Name] = *user

	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}
