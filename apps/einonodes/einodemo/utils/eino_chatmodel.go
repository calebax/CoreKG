package utils

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

// Config 可选配置
type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

var OpenAiDefaultConfig = &Config{
	APIKey:  "",
	Model:   "gpt-4o-mini",
	Timeout: 30 * time.Second,
}

var QWenDefaultConfig = &Config{
	APIKey:  "",
	Model:   "qwen-flash",
	Timeout: 30 * time.Second,
	BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
}

var DsDefaultConfigByYYGU = &Config{
	APIKey:  "xxx",
	Model:   "deepseek-v3",
	Timeout: 30 * time.Second,
	BaseURL: "https://api.example.com/v3/llm.chat",
}

func NewOpenAiChatModel(ctx context.Context, cfg *Config) (*openai.ChatModel, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	// 默认值
	if cfg.APIKey == "" {
		cfg.APIKey = "your-api-key"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}

	// 配置 ChatModel
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		Timeout: cfg.Timeout,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	return model, nil
}
