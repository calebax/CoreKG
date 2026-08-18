// through Confluence.
// Package confluence implements the OAuth2 protocol for authenticating users
package confluence

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

const (
	endpointProfile string = "https://api.atlassian.com/me"
	authURL         string = "https://auth.atlassian.com/authorize"
	tokenURL        string = "https://auth.atlassian.com/oauth/token"
)

// New creates a new Google provider, and sets up important connection details.
// You should always call `google.New` to get a new Provider. Never try to create
// one manually.
func New(clientKey, secret, callbackURL string, scopes ...string) *Provider {
	p := &Provider{
		ClientKey:       clientKey,
		Secret:          secret,
		CallbackURL:     callbackURL,
		providerName:    "confluence",
		HTTPClient:      &http.Client{},
		authCodeOptions: []oauth2.AuthCodeOption{},
	}
	p.config = newConfig(p, scopes)
	return p
}

// Provider is the implementation of `goth.Provider` for accessing Google.
type Provider struct {
	ClientKey       string
	Secret          string
	CallbackURL     string
	HTTPClient      *http.Client
	config          *oauth2.Config
	providerName    string
	authCodeOptions []oauth2.AuthCodeOption
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return p.providerName
}

// SetName is to update the name of the provider (needed in case of multiple providers of 1 type)
func (p *Provider) SetName(name string) {
	p.providerName = name
}

// Client returns an HTTP client to be used in all fetch operations.
func (p *Provider) Client() *http.Client {
	return goth.HTTPClientWithFallBack(p.HTTPClient)
}

// Debug is a no-op for the google package.
func (p *Provider) Debug(debug bool) {}

// BeginAuth asks Google for an authentication endpoint.
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	url := p.config.AuthCodeURL(state, p.authCodeOptions...)
	session := &Session{
		AuthURL: url,
	}
	return session, nil
}

type confluenceUser struct {
	AccountID       string `json:"account_id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Picture         string `json:"picture"`
	AccountStatus   string `json:"account_status"`
	Characteristics struct {
		NotMentionable bool `json:"not_mentionable"`
	} `json:"characteristics"`
	LastUpdated     time.Time `json:"last_updated"`
	Nickname        string    `json:"nickname"`
	Locale          string    `json:"locale"`
	ExtendedProfile struct {
		PhoneNumbers []string `json:"phone_numbers"`
	} `json:"extended_profile"`
	AccountType   string `json:"account_type"`
	EmailVerified bool   `json:"email_verified"`
}

// FetchUser will go to Google and access basic information about the user.
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	sess := session.(*Session)
	user := goth.User{
		AccessToken:  sess.AccessToken,
		Provider:     p.Name(),
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.ExpiresAt,
	}

	if user.AccessToken == "" {
		// Data is not yet retrieved, since accessToken is still empty.
		return user, fmt.Errorf("%s cannot get user information without accessToken", p.providerName)
	}

	// Get the userID, Slack needs userID in order to get user profile info
	req, _ := http.NewRequest("GET", endpointProfile, nil)
	req.Header.Add("Authorization", "Bearer "+sess.AccessToken)
	response, err := p.Client().Do(req)
	if err != nil {
		return user, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return user, fmt.Errorf("%s responded with a %d trying to fetch user information", p.providerName, response.StatusCode)
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	var cuser confluenceUser
	if err := json.Unmarshal(responseBytes, &cuser); err != nil {
		return user, err
	}

	user.Name = cuser.Name
	user.FirstName = cuser.Name
	user.LastName = cuser.Name
	user.NickName = cuser.Nickname
	user.Email = cuser.Email
	user.AvatarURL = cuser.Picture
	user.UserID = cuser.AccountID

	return user, nil
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"offline_access",
		},
	}

	if len(scopes) > 0 {
		c.Scopes = append(c.Scopes, scopes...)
	}
	return c
}

// RefreshTokenAvailable refresh token is provided by auth provider or not
func (p *Provider) RefreshTokenAvailable() bool {
	return true
}

// RefreshToken get new access token based on the refresh token
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	ts := p.config.TokenSource(goth.ContextForClient(p.Client()), token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, err
}

// SetPrompt sets the prompt values for the google OAuth call. Use this to
// force users to choose and account every time by passing "select_account",
// for example.
// See https://confluence.atlassian.com/enterprise/
func (p *Provider) SetPrompt(prompt ...string) {
	if len(prompt) == 0 {
		return
	}
	p.authCodeOptions = append(p.authCodeOptions, oauth2.SetAuthURLParam("prompt", strings.Join(prompt, " ")))
}
