package tools

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

const (
	getHeaderRow = "你是一名表格解析的程序，请解析一下下面的表格内容，返回一个JSON格式的结果，包括唯一的表头`header_row`和第一行数据的位置`first_data_row`，按如下格式返回，不要带其他任何上下文：" +
		`{"header_row":1,"first_data_row":2}` +

		"\n表内容如下"
)

// SimplePromptComponent 最简单的 prompt 组件
type SimplePromptComponent struct {
	template string
}

// Invoke 实现 Component 接口
func (p *SimplePromptComponent) Invoke(ctx context.Context, input map[string]any) (map[string]any, error) {
	// 创建 ChatMessage
	message := schema.SystemMessage(p.template)

	// 如果输入中有用户消息，添加到消息列表中
	messages := []*schema.Message{message}
	// compose.NewChain[1, 2]( )
	if userInput, ok := input["user_input"].(string); ok && userInput != "" {
		userMessage := schema.UserMessage(userInput)
		messages = append(messages, userMessage)
	}
	return map[string]any{
		"messages": messages,
		"prompt":   p.template,
	}, nil
}
