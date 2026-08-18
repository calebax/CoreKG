package agentclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
)

// TestChatClientIntegration 测试流式请求
func TestChatClientIntegration(t *testing.T) {
	client := NewChatClient(
		nil,
		"https://api.example.com/v3/chat.Agent/chat/completions",
		"yg-xxxxxxxxxxxxxxx",
	)

	// 2. 构造测试请求（明确设置Stream: false）
	req := &ChatRequestBody{
		Model: "AEQBaig",
		ChatOptions: ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: "五年级"},
				{Name: "input2", Value: "数学"},
				{Name: "input3", Value: "乘除法"},
			},
		},
		Stream: false, // 明确设置为非流式
	}
	ctx := context.Background()
	resp, err := client.SendChat(ctx, req)
	if err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}
	t.Logf("%+v", resp)
	// 4. 验证响应数据
	if len(resp.Choices) == 0 {
		t.Fatal("错误: 响应中未包含任何Choices")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		t.Error("警告: 响应内容为空")
	} else {
		t.Logf("AI回复 : %v", content)
	}

}
func TestChatStreamWithCallback(t *testing.T) {
	client := NewChatClient(nil, "https://api.example.com/v3/chat.Agent/chat/completions", "xxx")

	req := &ChatRequestBody{
		Model: "AEQBaig",
		ChatOptions: ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: "五年级"},
				{Name: "input2", Value: "数学"},
				{Name: "input3", Value: "乘除法"},
			},
		},
	}

	err := client.SendChatStreamWithCallback(context.Background(), req, func(chunk *ChatStreamResponseBody) error {
		fmt.Println("Received:", chunk.Choices[0].Delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("流式请求失败: %v", err)
	}
}
