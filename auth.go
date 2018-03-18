package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"
)

type authHandler struct {
	next         http.Handler
	omniAuthConf oauth2.Config
}

func NewAuthHandler(saName string, namespace string, saToken string, apiUrl string, next http.Handler) *authHandler {
	newAuthHandler := new(authHandler)
	conf := oauth2.Config{
		ClientID:     "system:serviceaccount:" + namespace + ":" + saName,
		ClientSecret: saToken,
		Scopes:       []string{"user:info", "user:check-access", "role:edit:" + namespace},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + apiUrl + "/oauth/authorize",
			TokenURL: "https://" + apiUrl + "/oauth/token",
		},
	}
	newAuthHandler.omniAuthConf = conf
	newAuthHandler.next = next
	return newAuthHandler

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
	//token := r.PostFormValue("token")
	//user := new(User)
	//user.token = token

	//err, status := user.login()
	provider := ""
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	if len(segs) > 3 {
		provider = segs[3]
	}
	glog.Infoln("URL", r.URL)
	glog.Infoln("Action", action, "Provider", provider)
	switch action {
	case "login":
		ctx := context.Background()
		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		url := h.omniAuthConf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		glog.Infoln("Visit the URL for the auth dialog: %v", url)
		// Use the authorization code that is pushed to the redirect
		// URL. Exchange will do the handshake to retrieve the
		// initial access token. The HTTP Client returned by
		// conf.Client will refresh the token as necessary.
		var code string
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatal(err)
		}
		tok, err := h.omniAuthConf.Exchange(ctx, code)
		if err != nil {
			log.Fatal(err)
		} else {
			glog.Infoln("TOKEN - ", tok)
		}

	case "callback":
		glog.Infoln("callback")
	}

	/*if err != nil {
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
	w.WriteHeader(http.StatusTemporaryRedirect)*/
}
