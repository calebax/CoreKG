package gh

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-github/github"
	"github.com/insmtx/corekg/pkgs/utils/httptools"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"golang.org/x/oauth2"
)

var cli = &http.Client{
	Timeout: time.Second * 10,
}

// GetOauthApp 获取github授权地址
func GetOauthApp(ctx context.Context, group, key string) (*OAuthApp, error) {
	ghApp := &OAuthApp{}
	err := settings.GetYaml(group, key, ghApp)
	if err != nil {
		logs.ErrorContextf(ctx, "GHAuthCallback: get github oauth config failed, %s", err)
		return nil, err
	}
	return ghApp, nil
}

// OAuthApp: github App
type OAuthApp struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	CallbackURL  string `yaml:"callback_url"`
}

// // AuthURL
// func (a *OAuthApp) AuthURL() string {
// 	return strings.TrimSuffix(a.CallbackURL, "/") + "/auth"
// }

// // BindURL
// func (a *OAuthApp) BindURL() string {
// 	return strings.TrimSuffix(a.CallbackURL, "/") + "/bind"
// }

// // bind or auth
// func (a *OAuthApp) AccessURL(use string) *url.URL {
// 	u, _ := url.Parse("https://github.com/login/oauth/authorize")
// 	callback := a.AuthURL()
// 	if use == "bind" {
// 		callback = a.BindURL()
// 	}
// 	query := u.Query()
// 	query.Set("client_id", a.ClientID)
// 	query.Set("redirect_uri", callback)
// 	scopes := []string{
// 		"user:email",
// 		"public_repo",
// 		"repo",
// 		"repo_deployment",
// 		"repo:status",
// 		"admin:repo_hook",
// 		"admin:org_hook",
// 		"admin:org",
// 		// "admin:public_key",
// 	}
// 	query.Add("scope", strings.Join(scopes, ","))

// 	u.RawQuery = query.Encode()

// 	return u
// }

// GetToken .
func (a *OAuthApp) GetToken(ctx context.Context, code string) (string, error) {
	if code == "" {
		logs.ErrorContextf(ctx, "get github-app token failed, invalid code")
		return "", nil
	}
	u, _ := url.Parse("https://github.com/login/oauth/access_token")

	query := u.Query()
	query.Set("client_id", a.ClientID)
	query.Set("client_secret", a.ClientSecret)
	query.Set("code", code)
	u.RawQuery = query.Encode()

	req, _ := http.NewRequest("POST", u.String(), nil)
	req.Header.Set("Accept", "application/json")

	if cli.Transport == nil {
		httptools.ProxySocks5FromSetting(cli, "core", "overseas_proxy")
	}
	resp, err := cli.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "get github-app token failed, %s", err.Error())
		return "", err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "read github-app token failed, %s", err.Error())
		return "", err
	}

	logs.DebugContextf(ctx, "get github-app token: %s", string(respBody))

	var ghToken struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}
	err = json.Unmarshal(respBody, &ghToken)
	if err != nil {
		logs.ErrorContextf(ctx, "read github-app token failed, %s", err.Error())
		return "", err
	}
	if ghToken.Error != "" {
		logs.ErrorContextf(ctx, "get github-app token failed, %s", ghToken.Error)
		return "", err
	}
	logs.DebugContextf(ctx, "get github-app token: %+v", ghToken)
	return ghToken.AccessToken, nil
}

// GetUserAccount
func (a *OAuthApp) GetUserAccount(ctx context.Context, token string) (*github.User, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	tc.Timeout = time.Second * 10

	client := github.NewClient(tc)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		logs.ErrorContextf(ctx, "GetUserAccount: %s", err)
		return nil, err
	}
	return user, nil
}
