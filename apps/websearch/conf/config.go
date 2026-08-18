package conf

import (
	"fmt"
	"slices"
	"strings"
	"time"

	runtimeauth "github.com/ygpkg/yg-go/apis/runtime/auth"
)

// Config contains the validated runtime settings consumed by WebSearch.
type Config struct {
	APIKey                    string
	TotalTimeout              time.Duration
	MaxRequestTimeout         time.Duration
	CursorSecret              string
	CursorTTL                 time.Duration
	CacheBypass               bool
	AllowRequestProviders     bool
	EnabledProviders          []string
	ProviderVisibility        string
	Debug                     bool
	DebugDir                  string
	DebugPreviewBytes         int
	LogStoreQuery             bool
	LogQueryPreviewChars      int
	ChromePath                string
	ChromeHeadless            bool
	ChromeNoSandbox           bool
	ChromeProfileDir          string
	BingProfileDir            string
	BraveProfileDir           string
	ProviderBrowserSlots      int
	DesktopURL                string
	MobileURL                 string
	DuckDuckGoURL             string
	BingURL                   string
	BraveURL                  string
	UserAgent                 string
	DesktopTimeout            time.Duration
	MobileTimeout             time.Duration
	ChromeTimeout             time.Duration
	DuckDuckGoTimeout         time.Duration
	BingTimeout               time.Duration
	BraveTimeout              time.Duration
	BaiduProfileCount         int
	BaiduProfileCapacity      int
	BingProfileCount          int
	BingProfileCapacity       int
	BraveProfileCount         int
	BraveProfileCapacity      int
	DuckDuckGoProfileCount    int
	DuckDuckGoProfileCapacity int
	AutoRouteWait             time.Duration
	ExplicitRouteWait         time.Duration
	MaxProviderAttempts       int
	MinimumAttemptBudget      time.Duration
	ProfileManifestRoot       string
	GlobalInflightMax         int
	AutoQueueMax              int
	CapacityRouterEnabled     bool
	ProfilePoolEnabled        bool
	FreshTTL                  time.Duration
	StaleTTL                  time.Duration
	ProviderRate              float64
	ProviderBurst             int
	JitterMin                 time.Duration
	JitterMax                 time.Duration
	BaiduSessionMinInterval   time.Duration
	BaiduSessionMaxJitter     time.Duration
	BaiduCaptchaCooldown      time.Duration
	BaiduRateLimitCooldown    time.Duration
	BaiduFallbackReserve      time.Duration
	CacheMaxItems             int
	MaxBodyBytes              int64
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

// Load reads one roc configuration file and returns the validated WebSearch settings.
func Load(path string) (Config, error) {
	config := Config{
		TotalTimeout: 20 * time.Second, MaxRequestTimeout: 60 * time.Second, CursorSecret: "local-development-only", CursorTTL: 15 * time.Minute,
		AllowRequestProviders: true, EnabledProviders: []string{"baidu", "bing", "brave", "duckduckgo"}, ProviderVisibility: "public",
		DebugDir: "./var/debug", DebugPreviewBytes: 32 << 10, LogQueryPreviewChars: 32,
		ChromeHeadless: true, ChromeProfileDir: "./var/chrome-profile", BingProfileDir: "./var/chrome-profile-bing", BraveProfileDir: "./var/chrome-profile-brave", ProviderBrowserSlots: 2,
		DesktopURL: "https://www.baidu.com/s", MobileURL: "https://m.baidu.com/s", DuckDuckGoURL: "https://html.duckduckgo.com/html/", BingURL: "https://www.bing.com/search", BraveURL: "https://search.brave.com/search",
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/150 Safari/537.36",
		DesktopTimeout: 4 * time.Second, MobileTimeout: 4 * time.Second, ChromeTimeout: 10 * time.Second, DuckDuckGoTimeout: 5 * time.Second, BingTimeout: 10 * time.Second, BraveTimeout: 10 * time.Second,
		BaiduProfileCount: 6, BaiduProfileCapacity: 1, BingProfileCount: 5, BingProfileCapacity: 2, BraveProfileCount: 3, BraveProfileCapacity: 1, DuckDuckGoProfileCount: 4, DuckDuckGoProfileCapacity: 1,
		AutoRouteWait: 2 * time.Second, ExplicitRouteWait: 5 * time.Second, MaxProviderAttempts: 3, MinimumAttemptBudget: 3 * time.Second,
		ProfileManifestRoot: "./var/provider-manifests", GlobalInflightMax: 100, AutoQueueMax: 100, CapacityRouterEnabled: true, ProfilePoolEnabled: true,
		FreshTTL: 15 * time.Minute, StaleTTL: 24 * time.Hour, ProviderRate: 1, ProviderBurst: 3, JitterMin: 200 * time.Millisecond, JitterMax: 800 * time.Millisecond,
		BaiduSessionMinInterval: 3 * time.Second, BaiduSessionMaxJitter: 2 * time.Second, BaiduCaptchaCooldown: 30 * time.Minute, BaiduRateLimitCooldown: 5 * time.Minute, BaiduFallbackReserve: 5 * time.Second,
		CacheMaxItems: 1000, MaxBodyBytes: 4 << 20,
	}
	if path != "" {
		if err := applyYAML(&config, path); err != nil {
			return Config{}, err
		}
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

// Validate verifies cross-field constraints that YAML decoding cannot express.
func (config Config) Validate() error {
	if config.ProviderVisibility != "hidden" && config.ProviderVisibility != "public" {
		return fmt.Errorf("api.provider_visibility must be hidden or public")
	}
	if config.TotalTimeout < 100*time.Millisecond || config.TotalTimeout > config.MaxRequestTimeout || config.MaxRequestTimeout > 60*time.Second {
		return fmt.Errorf("request timeout configuration must satisfy 100ms <= default <= max <= 60s")
	}
	if config.StaleTTL <= config.FreshTTL {
		return fmt.Errorf("cache.stale_ttl must be greater than cache.fresh_ttl")
	}
	if config.JitterMax < config.JitterMin {
		return fmt.Errorf("routing.jitter_max must be greater than or equal to routing.jitter_min")
	}
	valid := []string{"baidu", "bing", "brave", "duckduckgo"}
	if len(config.EnabledProviders) == 0 {
		return fmt.Errorf("api.enabled_providers must not be empty")
	}
	seen := make(map[string]struct{}, len(config.EnabledProviders))
	for _, provider := range config.EnabledProviders {
		if !slices.Contains(valid, provider) {
			return fmt.Errorf("unsupported enabled provider %q", provider)
		}
		if _, exists := seen[provider]; exists {
			return fmt.Errorf("enabled provider %q is duplicated", provider)
		}
		seen[provider] = struct{}{}
	}
	return nil
}
