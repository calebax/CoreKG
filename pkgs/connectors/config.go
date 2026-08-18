package connectors

type ConnectorsConfig struct {
	Providers []Config `json:"providers" yaml:"providers"`
}

type ProviderInfo struct {
	Provider string `json:"provider"` // 平台标识，如 "gmail", "slack"
	Logo     string `json:"logo"`     // 图标 URL
}

type Config struct {
	Name         string   `json:"name" yaml:"name"`
	Enable       bool     `json:"enable" yaml:"enable"`
	Platform     string   `json:"platform" yaml:"platform"`
	Logo         string   `json:"logo" yaml:"logo"`
	Method       string   `json:"method" yaml:"method"`
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
}
