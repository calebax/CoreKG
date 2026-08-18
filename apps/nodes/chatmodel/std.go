package chatmodel

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/settings"
)

// BaseModel .
func BaseModel(ctx context.Context) (*openai.ChatModel, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://api.example.com/v3/llm.chat/",
		APIKey:  "yg-xxx",
		Model:   "deepseek-v3",
	})
	if err != nil {
		return nil, err
	}
	return cm, err
}

// GetBaseModelFromSetting 创建基础模型，从core.settings配置中获取。
func GetBaseModelFromSetting(ctx context.Context, group, key string) (*openai.ChatModel, error) {
	cfg := config.LLMModelConfig{}
	err := settings.GetYaml(group, key, &cfg)
	if err != nil {
		return nil, err
	}

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKEY,
		Model:   cfg.ModelName,
	})
}
