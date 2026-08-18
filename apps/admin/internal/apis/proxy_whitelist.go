package apis

import (
	"fmt"
	"net/url"
	"strings"
)

// settings: admin / proxy_whitelist
type proxyWhitelist struct {
	BaseURLs []string `yaml:"base_urls"`
	Routers  []string `yaml:"routers"`
}

// normalizeProxyBaseURL returns scheme://host:port without path, no trailing slash.
func normalizeProxyBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty base url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("only http/https allowed")
	}
	if u.Host == "" {
		return "", fmt.Errorf("missing host")
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

func (w *proxyWhitelist) baseAllowed(normalizedBase string) bool {
	for _, b := range w.BaseURLs {
		nb, err := normalizeProxyBaseURL(b)
		if err != nil {
			continue
		}
		if nb == normalizedBase {
			return true
		}
	}
	return false
}

func (w *proxyWhitelist) pathAllowed(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}
	for _, rule := range w.Routers {
		if matchProxyRouterPath(path, rule) {
			return true
		}
	}
	return false
}

func matchProxyRouterPath(path, rule string) bool {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false
	}
	if strings.HasSuffix(rule, "/") {
		return strings.HasPrefix(path, rule)
	}
	return path == rule || strings.HasPrefix(path, rule+"/")
}

func joinProxyBaseAndPath(base, path string) string {
	base = strings.TrimSuffix(strings.TrimSpace(base), "/")
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}
