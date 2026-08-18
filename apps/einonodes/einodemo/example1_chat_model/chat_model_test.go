package example1chatmodel

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/einonodes/einodemo/utils"
	"github.com/stretchr/testify/assert"
)

func TestChat(t *testing.T) {

	ctx := context.Background()
	messages := []*schema.Message{
		schema.SystemMessage("你是一个助手"),
		schema.UserMessage("你好"),
	}

	// 初始化模型
	// chatModel, err := utils.NewOpenAiChatModel(ctx, utils.OpenAiDefaultConfig)
	// chatModel, err := utils.NewOpenAiChatModel(ctx, utils.QWenDefaultConfig)
	chatModel, err := utils.NewOpenAiChatModel(ctx, utils.DsDefaultConfigByYYGU)
	// chatModel, err := utils.NewOpenAiChatModel(ctx, &utils.Config{
	// 	APIKey:  "",
	// 	Model:   "deepseek-chat",
	// 	Timeout: 30 * time.Second,
	// 	BaseURL: "https://api.deepseek.com",
	// })
	assert.NoError(t, err)

	t.Run("ChatByGenerate", func(t *testing.T) {
		reasoningContent, content, err := ChatByGenerate(ctx, chatModel, messages)
		assert.NoError(t, err)
		t.Log(reasoningContent)
		t.Log(content)
	})

	t.Run("ChatByStream", func(t *testing.T) {
		err := ChatByStream(ctx, chatModel, messages, func(content string) {
			t.Log(content)
		})
		assert.NoError(t, err)
	})

}
