package main

import (
	"crypto/md5"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/golang/glog"
	userv1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type authHandler struct {
	next         http.Handler
	omniAuthConf oauth2.Config
}

func NewAuthHandler(saName string,
	saToken string,
	authUrl string,
	tokenUrl string,
	next http.Handler) *authHandler {

	newAuthHandler := new(authHandler)
	conf := oauth2.Config{
		ClientID:     "system:serviceaccount:" + *openshiftNamespace + ":" + saName,
		ClientSecret: saToken,
		Scopes:       []string{"user:info", "user:check-access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl,
			TokenURL: tokenUrl,
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

	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]

	switch action {
	case "login":

		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("Error when trying to get provider", provider, "-", err)
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("Error when trying to GetBeginAuthUrl for", provider, "-", err)
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			glog.Errorln("Error when trying to get provider", provider, "-", err)
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
		if *allowInsecure {
			glog.Warningln("Skipping TLS Verify")
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
		if err != nil {
			glog.Errorln("Error when trying to complete auth for", provider, "-", err)
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

		config := &rest.Config{
			Host:            *openshiftApiHost,
			BearerToken:     creds.Get("access_token").String(),
			TLSClientConfig: rest.TLSClientConfig{Insecure: *allowInsecure},
		}
		userV1Client, err := userv1.NewForConfig(config)
		if err != nil {
			glog.Errorln("could not connect to OpenShift API:", err)
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

		user, err := userV1Client.Users().Get("~", metav1.GetOptions{})
		if err != nil {
			glog.Errorln("could not get User:", err)
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			glog.Infof("User: ", user.ObjectMeta.Name)
		}

		kubeClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			glog.Errorln(err)
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

		authorizationClent := kubeClient.AuthorizationV1().SelfSubjectAccessReviews()

		selfSubjectAccessReview := &authv1.SelfSubjectAccessReview{
			Spec: authv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Resource:  "pods",
					Verb:      "get",
					Namespace: *openshiftNamespace,
				},
			},
		}
		result, err := authorizationClent.Create(selfSubjectAccessReview)
		if err != nil {
			glog.Errorln(err)
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

		if !result.Status.Allowed {
			w.Header().Set("Location", "/denied")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

		m := md5.New()
		io.WriteString(m, strings.ToLower(user.ObjectMeta.Name))

		authCookieValue := objx.New(map[string]interface{}{
			"name": user.ObjectMeta.Name,
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})
		Users[user.ObjectMeta.Name] = *user
		UserTokens[user.ObjectMeta.Name] = creds.Get("access_token").String()

		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
