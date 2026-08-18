package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// 使用内置的 prompt 组件示例
func useBuiltinPromptComponent() error {
	ctx := context.Background()

	// 使用内置的 ChatTemplate 组件
	msgTpl := []schema.MessagesTemplate{
		schema.SystemMessage(
			"你是一个有用的AI助手。请回答用户的问题。"),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("问题: {question}"),
	}
	chatTemplate := prompt.FromMessages(schema.FString, msgTpl...)

	// 准备输入数据
	input := map[string]any{
		"question": "Go语言的特点是什么？",
	}

	// 调用组件
	msgs, err := chatTemplate.Format(ctx, input)
	if err != nil {
		return fmt.Errorf("调用 ChatTemplate 失败: %w", err)
	}

	// 输出结果
	fmt.Println("=== 内置 ChatTemplate 组件结果 ===")
	for i, msg := range msgs {
		fmt.Printf("Message %d [%s]: %s\n", i+1, msg.Role, msg.Content)
	}

	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://api.example.com/v3/llm.chat/",
		APIKey:  "xxx",
		Model:   "deepseek-v3",
	})
	if err != nil {
		return fmt.Errorf("创建 ChatModel 失败: %w", err)
	}
	output, err := cm.Generate(ctx, msgs)
	if err != nil {
		return fmt.Errorf("调用 ChatModel 失败: %w", err)
	}

	fmt.Println("=== 内置 ChatModel 组件结果 ===")
	fmt.Printf("Message %d [%s]: %s\n", 1, output.Role, output.Content)
	//callbacks.ModelCallbackHandler()
	callbacks.NewHandlerBuilder()
	return nil
}

func main() {
	fmt.Println("Eino Prompt 组件示例")
	fmt.Println("===================")

	// 1. 使用内置的 prompt 组件
	if err := useBuiltinPromptComponent(); err != nil {
		log.Printf("内置组件示例失败: %v", err)
	}

}
