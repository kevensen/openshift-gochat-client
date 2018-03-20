package openshift

import (
	"net/http"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/gomniauth/oauth2"
	"github.com/stretchr/objx"
)

const (
	openshiftName        string = "openshift"
	openshiftDisplayName string = "OpenShift"
)

type OpenShiftProvider struct {
	Config         *common.Config
	tripperFactory common.TripperFactory
	Scope          string
}

func New(clientId string, clientSecret string, namespace string, metadata *OAuthServerMetadata) *OpenShiftProvider {

	p := new(OpenShiftProvider)
	p.Config = &common.Config{Map: objx.MSI(
		oauth2.OAuth2KeyAuthURL, metadata.AuthorizationEndpoint,
		oauth2.OAuth2KeyTokenURL, metadata.TokenEndpoint,
		oauth2.OAuth2KeyClientID, clientId,
		oauth2.OAuth2KeySecret, clientSecret,
		oauth2.OAuth2KeyScope, "user:info user:check-access role:edit:"+namespace,
		oauth2.OAuth2KeyAccessType, oauth2.OAuth2AccessTypeOnline,
		oauth2.OAuth2KeyGrantType, "implicit",
		oauth2.OAuth2KeyApprovalPrompt, oauth2.OAuth2ApprovalPromptAuto,
		oauth2.OAuth2KeyResponseType, oauth2.OAuth2KeyCode)}
	return p
}

// TripperFactory gets an OAuth2TripperFactory
func (provider *OpenShiftProvider) TripperFactory() common.TripperFactory {

	if provider.tripperFactory == nil {
		provider.tripperFactory = new(oauth2.OAuth2TripperFactory)
	}

	return provider.tripperFactory
}

// PublicData gets a public readable view of this provider.
func (provider *OpenShiftProvider) PublicData(options map[string]interface{}) (interface{}, error) {
	return gomniauth.ProviderPublicData(provider, options)
}

// Name is the unique name for this provider.
func (provider *OpenShiftProvider) Name() string {
	return openshiftName
}

// DisplayName is the human readable name for this provider.
func (provider *OpenShiftProvider) DisplayName() string {
	return openshiftDisplayName
}

// GetBeginAuthURL gets the URL that the client must visit in order
// to begin the authentication process.
//
// The state argument contains anything you wish to have sent back to your
// callback endpoint.
// The options argument takes any options used to configure the auth request
// sent to the provider. In the case of OAuth2, the options map can contain:
//   1. A "scope" key providing the desired scope(s). It will be merged with the default scope.
func (provider *OpenShiftProvider) GetBeginAuthURL(state *common.State, options objx.Map) (string, error) {
	if options != nil {
		scope := options.Get(oauth2.OAuth2KeyScope).Str()
		provider.Config.Set(oauth2.OAuth2KeyScope, scope)
	}
	return oauth2.GetBeginAuthURLWithBase(provider.Config.Get(oauth2.OAuth2KeyAuthURL).Str(), state, provider.Config)
}

// Get makes an authenticated request and returns the data in the
// response as a data map.
func (provider *OpenShiftProvider) Get(creds *common.Credentials, endpoint string) (objx.Map, error) {
	return oauth2.Get(provider, creds, endpoint)
}

// GetUser uses the specified common.Credentials to access the users profile
// from the remote provider, and builds the appropriate User object.
func (provider *OpenShiftProvider) GetUser(creds *common.Credentials) (common.User, error) {

	/* 	profileData, err := provider.Get(creds, openshiftEndpointProfile)

	   	if err != nil {
	   		return nil, err
	   	}

	   	// build user
	   	user := NewUser(profileData, creds, provider) */

	return nil, nil
}

// CompleteAuth takes a map of arguments that are used to
// complete the authorisation process, completes it, and returns
// the appropriate Credentials.
func (provider *OpenShiftProvider) CompleteAuth(data objx.Map) (*common.Credentials, error) {
	return oauth2.CompleteAuth(provider.TripperFactory(), data, provider.Config, provider)
}

// GetClient returns an authenticated http.Client that can be used to make requests to
// protected Github resources
func (provider *OpenShiftProvider) GetClient(creds *common.Credentials) (*http.Client, error) {
	return oauth2.GetClient(provider.TripperFactory(), creds, provider)
}
