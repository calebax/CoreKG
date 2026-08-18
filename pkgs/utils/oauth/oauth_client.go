package utils

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"golang.org/x/oauth2"
)

// CreateOAuth2Client 创建OAuth2 HTTP客户端
func CreateOAuth2Client(accessToken string) *http.Client {
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	config := &oauth2.Config{}

	return config.Client(context.Background(), token)
}

// CreateOAuth2ClientWithProxy 创建OAuth2 HTTP客户端
func CreateOAuth2ClientWithProxy(accessToken string) *http.Client {
	proxy := GetProxyUrlConfig()
	if proxy == "" {
		return CreateOAuth2Client(accessToken)
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return CreateOAuth2Client(accessToken)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// 创建基础客户端
	baseClient := &http.Client{
		Transport: transport,
	}

	// 使用自定义的基础客户端创建OAuth2客户端
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, baseClient)

	config := &oauth2.Config{}
	return config.Client(ctx, token)
}

func CreateHttpClient() *http.Client {
	return &http.Client{}
}

func CreateHttpClientWithProxy() *http.Client {
	proxy := GetProxyUrlConfig()
	if proxy == "" {
		return CreateHttpClient()
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		// 如果代理URL解析失败，使用默认客户端
		return CreateHttpClient()
	}
	// 创建 http.Transport，支持代理
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
	}
}

func GetProxyUrlConfig() string {
	ctx := context.TODO()
	proxy, err := settings.GetText("core", "proxy_url")
	if err != nil {
		logs.WarnContextf(ctx, "GetProxyUrlConfig failed: %s", err)
		return ""
	}
	return proxy
}
