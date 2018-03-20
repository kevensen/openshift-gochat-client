package openshift

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

type OAuthServerMetadata struct {
	Issuer                                     string   `json:"issuer,omitempty"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                              string   `json:"token_endpoint,omitempty"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`
	UserInfoEndpoint                           string   `json:"userinfo_endpoint,omitempty"`
	JwksURI                                    string   `json:"jwks_uri,omitempty"`
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported,omitempty"`
	ResponseTypesSupported                     []string `json:"response_types_supported,omitempty"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"`
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty"`
}

var OCPWellKnownURL = "https://openshift.default.svc/.well-known/oauth-authorization-server"

func NewOAuthServerMetadata() *OAuthServerMetadata {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	metadata := new(OAuthServerMetadata)
	resp, err := http.Get(OCPWellKnownURL)
	if err != nil {
		glog.Errorln("Unable to obtain server metadata.")
		return metadata
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(metadata)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	return metadata
}
