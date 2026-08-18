package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLProviderControls(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("api:\n  enabled_providers: [bing, brave]\n  allow_request_providers: true\n  provider_visibility: hidden\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !config.AllowRequestProviders || config.ProviderVisibility != "hidden" || len(config.EnabledProviders) != 2 || config.EnabledProviders[0] != "bing" {
		t.Fatalf("config=%#v", config)
	}
}

func TestLoadAcceptsRocMainAndLoggerSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "main:\n  app: websearch\n  env: test\n  http_addr: ':8080'\nlogger:\n  default:\n    - writer: console\n      level: debug\nauth:\n  api_key: yg-config-token\nserver: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatalf("load roc config sections: %v", err)
	}
	if config.APIKey != "yg-config-token" {
		t.Fatalf("api key = %q", config.APIKey)
	}
}

func TestLoadYAMLRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("server:\n  unknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown YAML field error")
	}
}

func TestLoadYAMLCoversAdvancedRuntimeSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := `browser:
  chrome_timeout: 9s
routing:
  jitter_min: 100ms
  jitter_max: 300ms
  profile_manifest_root: ./manifests
providers:
  baidu:
    mobile_timeout: 3s
    session_min_interval: 2s
    session_jitter_max: 0s
    captcha_cooldown: 10m
    rate_limit_cooldown: 2m
    fallback_reserve: 0s
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.ChromeTimeout.String() != "9s" || config.MobileTimeout.String() != "3s" || config.JitterMin.String() != "100ms" || config.BaiduSessionMaxJitter != 0 || config.BaiduFallbackReserve != 0 || config.ProfileManifestRoot != "./manifests" {
		t.Fatalf("config=%#v", config)
	}
}

func TestLoadYAMLRejectsMultipleDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("server: {}\n---\nserver: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected multiple YAML document error")
	}
}

func TestResolveAPIKeyUsesCommandLineOverride(t *testing.T) {
	value, err := ResolveAPIKey("yg-config-token", " yg-command-token ")
	if err != nil || value != "yg-command-token" {
		t.Fatalf("ResolveAPIKey() = %q, %v", value, err)
	}
	if _, err := ResolveAPIKey("plain-token", ""); err == nil {
		t.Fatal("expected invalid API key error")
	}
}
