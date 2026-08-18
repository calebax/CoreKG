package conf

import (
	"fmt"
	"strings"
	"time"

	runtimeauth "github.com/ygpkg/yg-go/apis/runtime/auth"
)

// Config contains the validated runtime settings consumed by WebFetch.
type Config struct {
	APIKey            string
	RequestTimeout    time.Duration
	MaxRequestTimeout time.Duration
	HTTPTimeout       time.Duration
	BrowserEnabled    bool
	BrowserTimeout    time.Duration
	BrowserWait       time.Duration
	BrowserSlots      int
	ChromePath        string
	ChromeProfileDir  string
	ChromeHeadless    bool
	ChromeNoSandbox   bool
	FreshTTL          time.Duration
	StaleTTL          time.Duration
	CacheMaxItems     int
	MaxBodyBytes      int64
	MaxRedirects      int
	HostAllowlist     []string
	CacheBypass       bool
	Diagnostics       bool
	LogStoreURLQuery  bool
	RobotsPolicy      string
	UserAgent         string
}

// ResolveAPIKey applies the command-line override and validates the token.
func ResolveAPIKey(configValue, commandLineValue string) (string, error) {
	value := strings.TrimSpace(commandLineValue)
	if value == "" {
		value = strings.TrimSpace(configValue)
	}
	if value == "" {
		return "", fmt.Errorf("api key is required in auth.api_key or --api-key")
	}
	if !strings.HasPrefix(value, runtimeauth.AuthAPIKeyPrefix) {
		return "", fmt.Errorf("api key must use the %q prefix", runtimeauth.AuthAPIKeyPrefix)
	}
	return value, nil
}

// Load reads one roc configuration file and returns the validated WebFetch settings.
func Load(path string) (Config, error) {
	config := Config{
		RequestTimeout: 20 * time.Second, MaxRequestTimeout: 60 * time.Second, HTTPTimeout: 6 * time.Second,
		BrowserEnabled: true, BrowserTimeout: 12 * time.Second, BrowserWait: 250 * time.Millisecond,
		BrowserSlots: 4, ChromeProfileDir: "./var/chrome-profile", ChromeHeadless: true,
		FreshTTL: 30 * time.Minute, StaleTTL: 24 * time.Hour, CacheMaxItems: 500,
		MaxBodyBytes: 5 << 20, MaxRedirects: 5, RobotsPolicy: "ignore",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/150 Safari/537.36",
	}
	if path != "" {
		if err := applyYAML(&config, path); err != nil {
			return Config{}, err
		}
	}
	if config.RobotsPolicy != "ignore" {
		return Config{}, fmt.Errorf("http.robots_policy=%s is not implemented", config.RobotsPolicy)
	}
	if config.StaleTTL <= config.FreshTTL {
		return Config{}, fmt.Errorf("cache.stale_ttl must be greater than cache.fresh_ttl")
	}
	if config.RequestTimeout < 100*time.Millisecond || config.RequestTimeout > config.MaxRequestTimeout || config.MaxRequestTimeout > 60*time.Second {
		return Config{}, fmt.Errorf("request timeout configuration must satisfy 100ms <= default <= max <= 60s")
	}
	return config, nil
}
