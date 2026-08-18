package conf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type yamlConfig struct {
	Main   yaml.Node `yaml:"main"`
	Logger yaml.Node `yaml:"logger"`
	Auth   struct {
		APIKey *string `yaml:"api_key"`
	} `yaml:"auth"`
	Server struct {
		DefaultTimeout *string `yaml:"default_timeout"`
		MaxTimeout     *string `yaml:"max_timeout"`
	} `yaml:"server"`
	HTTP struct {
		Timeout       *string  `yaml:"timeout"`
		MaxBodyBytes  *int64   `yaml:"max_body_bytes"`
		MaxRedirects  *int     `yaml:"max_redirects"`
		HostAllowlist []string `yaml:"host_allowlist"`
		UserAgent     *string  `yaml:"user_agent"`
		RobotsPolicy  *string  `yaml:"robots_policy"`
	} `yaml:"http"`
	Browser struct {
		Enabled      *bool   `yaml:"enabled"`
		Timeout      *string `yaml:"timeout"`
		PostLoadWait *string `yaml:"post_load_wait"`
		Slots        *int    `yaml:"slots"`
		ChromePath   *string `yaml:"chrome_path"`
		ProfileDir   *string `yaml:"profile_dir"`
		Headless     *bool   `yaml:"headless"`
		NoSandbox    *bool   `yaml:"no_sandbox"`
	} `yaml:"browser"`
	Cache struct {
		FreshTTL *string `yaml:"fresh_ttl"`
		StaleTTL *string `yaml:"stale_ttl"`
		MaxItems *int    `yaml:"max_items"`
		Bypass   *bool   `yaml:"bypass"`
	} `yaml:"cache"`
	Observability struct {
		DiagnosticsEnabled *bool `yaml:"diagnostics_enabled"`
		StoreURLQuery      *bool `yaml:"store_url_query"`
	} `yaml:"observability"`
}

func applyYAML(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var file yamlConfig
	if err := decoder.Decode(&file); err != nil {
		return fmt.Errorf("decode config %s: %w", path, err)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode config %s: multiple YAML documents are not allowed", path)
	}
	setString := func(target *string, value *string) {
		if value != nil {
			*target = *value
		}
	}
	setBool := func(target *bool, value *bool) {
		if value != nil {
			*target = *value
		}
	}
	setInt := func(target *int, value *int) {
		if value != nil {
			*target = *value
		}
	}
	setInt64 := func(target *int64, value *int64) {
		if value != nil {
			*target = *value
		}
	}
	setDuration := func(name string, target *time.Duration, value *string) error {
		if value == nil {
			return nil
		}
		parsed, err := time.ParseDuration(*value)
		if err != nil || parsed <= 0 {
			return fmt.Errorf("%s must be a positive duration", name)
		}
		*target = parsed
		return nil
	}
	setString(&config.UserAgent, file.HTTP.UserAgent)
	setString(&config.APIKey, file.Auth.APIKey)
	setString(&config.RobotsPolicy, file.HTTP.RobotsPolicy)
	if file.HTTP.HostAllowlist != nil {
		config.HostAllowlist = append([]string(nil), file.HTTP.HostAllowlist...)
	}
	for _, item := range []struct {
		name   string
		target *time.Duration
		value  *string
	}{
		{"server.default_timeout", &config.RequestTimeout, file.Server.DefaultTimeout}, {"server.max_timeout", &config.MaxRequestTimeout, file.Server.MaxTimeout}, {"http.timeout", &config.HTTPTimeout, file.HTTP.Timeout}, {"browser.timeout", &config.BrowserTimeout, file.Browser.Timeout}, {"browser.post_load_wait", &config.BrowserWait, file.Browser.PostLoadWait}, {"cache.fresh_ttl", &config.FreshTTL, file.Cache.FreshTTL}, {"cache.stale_ttl", &config.StaleTTL, file.Cache.StaleTTL},
	} {
		if err := setDuration(item.name, item.target, item.value); err != nil {
			return err
		}
	}
	setInt64(&config.MaxBodyBytes, file.HTTP.MaxBodyBytes)
	setInt(&config.MaxRedirects, file.HTTP.MaxRedirects)
	setBool(&config.BrowserEnabled, file.Browser.Enabled)
	setInt(&config.BrowserSlots, file.Browser.Slots)
	setString(&config.ChromePath, file.Browser.ChromePath)
	setString(&config.ChromeProfileDir, file.Browser.ProfileDir)
	setBool(&config.ChromeHeadless, file.Browser.Headless)
	setBool(&config.ChromeNoSandbox, file.Browser.NoSandbox)
	setInt(&config.CacheMaxItems, file.Cache.MaxItems)
	setBool(&config.CacheBypass, file.Cache.Bypass)
	setBool(&config.Diagnostics, file.Observability.DiagnosticsEnabled)
	setBool(&config.LogStoreURLQuery, file.Observability.StoreURLQuery)
	return nil
}
