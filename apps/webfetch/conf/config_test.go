package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsUnimplementedRobotsPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("http:\n  robots_policy: respect\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unimplemented robots policy error")
	}
}

func TestLoadAcceptsRocMainAndLoggerSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "main:\n  app: webfetch\n  env: test\n  http_addr: ':8080'\nlogger:\n  default:\n    - writer: console\n      level: debug\nauth:\n  api_key: yg-config-token\nserver: {}\n"
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
	if err := os.WriteFile(path, []byte("browser:\n  unknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown YAML field error")
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
