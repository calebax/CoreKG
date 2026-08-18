package modelhelper

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/logs"
)

const DefaultChatModelTimeout = 600 * time.Second

type ToolCallingChatModelOptions struct {
	Timeout         time.Duration
	Temperature     *float32
	TopP            *float32
	PresencePenalty *float32
	ResponseFormat  *openai.ChatCompletionResponseFormat
}

func NewToolCallingChatModel(
	ctx context.Context,
	chatModel *chattype.ChatModel,
	options ToolCallingChatModelOptions,
) (model.ToolCallingChatModel, error) {
	if options.Timeout <= 0 {
		options.Timeout = DefaultChatModelTimeout
	}
	modelConfig := &openai.ChatModelConfig{
		APIKey:  chatModel.APIKey,
		Model:   chatModel.ModelName,
		Timeout: options.Timeout,
		BaseURL: strings.TrimSuffix(chatModel.ModelUrl, "/chat/completions"),
	}

	// 私有化部署下，针对 DisableThinkingModelKeywords 中的模型关闭思考模式，避免返回额外思考内容。
	disableThinking := false
	if strings.EqualFold(version.DeployMode(), global.DeployModeOnPremise) {
		lowerModelName := strings.ToLower(modelConfig.Model)
		for _, keyword := range global.DisableThinkingModelKeywords {
			if strings.Contains(lowerModelName, strings.ToLower(keyword)) {
				disableThinking = true
				break
			}
		}
	}
	if disableThinking {
		modelConfig.ExtraFields = map[string]any{
			"chat_template_kwargs": map[string]any{
				"enable_thinking": false,
			},
		}
	}

	modelConfig.Temperature = options.Temperature
	modelConfig.TopP = options.TopP
	modelConfig.PresencePenalty = options.PresencePenalty
	modelConfig.ResponseFormat = options.ResponseFormat

	toolCallingChatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		logs.ErrorContextf(ctx, "[NewToolCallingChatModel] failed to create OpenAiChatModel: %v", err)
		return nil, err
	}
	return toolCallingChatModel, nil
}
