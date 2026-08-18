package connectors

import (
	"context"
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/slack"
	"github.com/insmtx/corekg/pkgs/connectors/providers/confluence"
	oauthUtils "github.com/insmtx/corekg/pkgs/utils/oauth"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

const (
	PlatformGoogle     = "google"
	PlatformSlack      = "slack"
	PlatformConfluence = "confluence"
)

var supportedProviders = make(map[string]ProviderInfo)

func InitProviders(ctx context.Context, group, key string) error {

	connectorsConfig := &ConnectorsConfig{}
	err := settings.GetYaml(group, key, connectorsConfig)

	if err != nil {
		logs.ErrorContextf(ctx, "[connectors] load providers config failed, %s", err)
		return err
	}

	dir := ".connect_sessions"

	// 判断目录是否存在，如果不存在则创建
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Failed to create session directory: %v", err)
		}
	}

	// 初始化 FilesystemStore
	store := sessions.NewFilesystemStore(dir, []byte(key))
	store.MaxLength(32 * 1024)
	gothic.Store = store

	client := oauthUtils.CreateHttpClientWithProxy()

	for _, p := range connectorsConfig.Providers {
		if !p.Enable {
			continue
		}
		var provider goth.Provider
		switch p.Platform {
		case PlatformGoogle:
			googleProvider := google.New(p.ClientID, p.ClientSecret, p.RedirectURL, p.Scopes...)
			googleProvider.SetAccessType("offline")
			googleProvider.SetPrompt("consent")
			googleProvider.HTTPClient = client
			provider = googleProvider
		case PlatformSlack:
			slackProvider := slack.New(p.ClientID, p.ClientSecret, p.RedirectURL, p.Scopes...)
			slackProvider.HTTPClient = client
			provider = slackProvider
		case PlatformConfluence:
			confluenceProvider := confluence.New(p.ClientID, p.ClientSecret, p.RedirectURL, p.Scopes...)
			confluenceProvider.HTTPClient = client
			confluenceProvider.SetPrompt("consent")
			provider = confluenceProvider
		}

		if provider == nil {
			logs.ErrorContextf(ctx, "[connectors] init provider failed, platform = %s", p.Platform)
			continue
		}
		provider.SetName(p.Name)
		goth.UseProviders(provider)

		supportedProviders[p.Name] = ProviderInfo{
			Provider: p.Name,
			Logo:     p.Logo,
		}
	}

	return nil
}

func IsProviderSupported(name string) bool {
	_, ok := supportedProviders[name]
	return ok
}

func ListSupportedProviders() ([]ProviderInfo, error) {
	var providers []ProviderInfo
	for _, info := range supportedProviders {
		providers = append(providers, info)
	}
	return providers, nil
}
