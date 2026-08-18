package example3agentchain

import (
	"context"
	"testing"

	"github.com/insmtx/corekg/apps/einonodes/einodemo/utils"
)

func TestChatTemplate(t *testing.T) {
	ctx := context.Background()
	// 初始化模型(需支持Function Calling)
	chatModel, err := utils.NewOpenAiChatModel(ctx, utils.OpenAiDefaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ChatAgent", func(t *testing.T) {
		SimpleAgent(ctx, &Config{ChatModel: chatModel},
			"介绍了使用 Eino 框架构建 Agent 的基本方法。通过 Chain、Tool Calling 和 ReAct 等不同方式，我们可以根据实际需求灵活地构建 AI Agent。Agent 是 AI 技术发展的重要方向。它不仅能够理解用户意图，还能主动采取行动，通过调用各种工具来完成复杂任务。随着大语言模型能力的不断提升，Agent 将在未来扮演越来越重要的角色，成为连接 AI 与现实世界的重要桥梁。我们期待 Eino 能为用户带来更强大、易用的 agent 构建方案，推动更多基于 Agent 的应用创新。并帮我总结一下")
	})

	t.Run("ChatAgent2", func(t *testing.T) {
		SimpleAgent(ctx, &Config{ChatModel: chatModel},
			"查询告诉我今日金价是多少？")
	})

}
