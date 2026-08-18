package renderer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

const (
	defaultSystemMessage   = "你是一个具有函数调用能力的AI助手。"
	noToolsSystemMessage   = "你是一个有帮助的AI助手。"
	withToolsSystemMessage = "你可以使用以下工具来协助回答用户的问题。当你需要调用工具时，请按照指定的JSON格式返回。"
)

type GenericRenderer struct{}

// Render 将消息列表和工具列表转换为格式化的提示词字符串
func (r *GenericRenderer) Render(messages []kellmtype.Message, tools []kellmtype.Tool) (string, error) {
	var sb strings.Builder

	// 提取系统消息
	var systemMessage *kellmtype.Message
	filteredMessages := make([]kellmtype.Message, 0, len(messages))
	for i, message := range messages {
		if message.Role == "system" {
			if systemMessage == nil {
				systemMessage = &messages[i]
			}
			continue
		}
		filteredMessages = append(filteredMessages, message)
	}

	// 渲染系统消息和工具定义
	if err := r.renderSystemMessage(&sb, systemMessage, tools); err != nil {
		return "", err
	}

	// 渲染对话消息
	for i, message := range filteredMessages {
		lastMessage := i == len(filteredMessages)-1
		if err := r.renderMessage(&sb, message, lastMessage); err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

// renderSystemMessage 渲染系统消息和工具定义
func (r *GenericRenderer) renderSystemMessage(sb *strings.Builder, systemMessage *kellmtype.Message, tools []kellmtype.Tool) error {
	sb.WriteString("系统: ")

	// 使用自定义系统消息或默认消息
	if systemMessage != nil && systemMessage.Content.Text != "" {
		sb.WriteString(systemMessage.Content.Text)
	} else {
		if len(tools) > 0 {
			sb.WriteString(defaultSystemMessage)
			sb.WriteString(" ")
			sb.WriteString(withToolsSystemMessage)
		} else {
			sb.WriteString(noToolsSystemMessage)
		}
	}

	// 如果有工具则添加工具定义
	if len(tools) > 0 {
		if err := r.renderToolDefinitions(sb, tools); err != nil {
			return err
		}
	}

	sb.WriteString("\n\n")
	return nil
}

// renderToolDefinitions 渲染工具定义的结构化格式
func (r *GenericRenderer) renderToolDefinitions(sb *strings.Builder, tools []kellmtype.Tool) error {
	sb.WriteString("\n\n## 可用工具\n\n")

	// 渲染工具列表
	for i, tool := range tools {
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, tool.Function.Name))

		if tool.Function.Description != "" {
			sb.WriteString(fmt.Sprintf("**功能描述**: %s\n", tool.Function.Description))
		}

		// 渲染参数定义
		if len(tool.Function.Parameters) > 0 || tool.Function.Parameters.Type() != "" {
			sb.WriteString("\n**参数定义**:\n```json\n")
			paramsJSON, err := json.MarshalIndent(tool.Function.Parameters, "", "  ")
			if err != nil {
				return fmt.Errorf("工具 %s 的参数序列化失败: %w", tool.Function.Name, err)
			}
			sb.Write(paramsJSON)
			sb.WriteString("\n```\n\n")
		}
	}

	// 添加使用说明
	r.renderToolUsageInstructions(sb)

	return nil
}

// renderToolUsageInstructions 提供清晰的工具调用说明
func (r *GenericRenderer) renderToolUsageInstructions(sb *strings.Builder) {
	sb.WriteString("\n```\n\n")
	sb.WriteString("## 调用格式\n\n")
	sb.WriteString("**重要**：当你需要调用工具时，必须按照以下两步骤格式返回，缺一不可：\n\n")
	sb.WriteString("**步骤1**：在第一行写一句话，描述你要调用什么工具以及传入什么参数\n\n")
	sb.WriteString("**步骤2**：空一行后，写 JSON 代码块（以 ```json 开头，以 ``` 结尾）\n\n")
	sb.WriteString("### 标准示例（必须遵循）\n\n")
	sb.WriteString("调用天气查询工具获取西安的天气信息\n\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{"name": "get_weather", "arguments": "{\"location\": \"西安\"}"}`)
	sb.WriteString("\n```\n\n")
	sb.WriteString("### 错误示例（禁止这样做）\n\n")
	sb.WriteString("禁止直接返回 JSON 代码块而没有前面的描述：\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{"name": "get_weather", "arguments": "{\"location\": \"西安\"}"}`)
	sb.WriteString("\n```\n\n")
	sb.WriteString("### JSON 格式说明\n\n")
	sb.WriteString("- `name`: 工具名称（必须是上述工具列表中的某个）\n")
	sb.WriteString("- `arguments`: **必须是转义后的 JSON 字符串**（注意：是字符串，不是对象）\n")
	sb.WriteString("- 在 `arguments` 字符串中，所有双引号必须使用 `\\\"` 转义\n\n")
	sb.WriteString("### 完整格式模板\n\n")
	sb.WriteString("<一句话描述调用的工具和参数>\n\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{"name": "<工具名称>", "arguments": "<转义后的JSON字符串>"}`)
	sb.WriteString("\n```\n\n")
	// sb.WriteString("**记住**：每次调用工具时，必须同时包含描述文字和 JSON 代码块！\n")
}

// renderMessage 渲染单条对话消息
func (r *GenericRenderer) renderMessage(sb *strings.Builder, message kellmtype.Message, lastMessage bool) error {
	switch message.Role {
	case "user":
		sb.WriteString("用户: ")
		if message.Content.Text != "" {
			sb.WriteString(message.Content.Text)
		}
		sb.WriteString("\n\n")

	case "assistant":
		sb.WriteString("助手: ")

		// 渲染文本内容
		if message.Content.Text != "" {
			sb.WriteString(message.Content.Text)
		}

		// 渲染工具调用
		if len(message.ToolCalls) > 0 {
			if message.Content.Text != "" {
				sb.WriteString("\n\n")
			}
			if err := r.renderToolCalls(sb, message.ToolCalls); err != nil {
				return err
			}
		}

		// 如果是最后一条消息且是 prefill（仅有内容，无工具调用），则不添加换行
		if !lastMessage || len(message.ToolCalls) > 0 {
			sb.WriteString("\n\n")
		}

	case "tool":
		sb.WriteString("工具返回: ")
		if message.Content.Text != "" {
			sb.WriteString(message.Content.Text)
		}
		sb.WriteString("\n\n")
	}

	return nil
}

// renderToolCalls 以 JSON 格式渲染工具调用
func (r *GenericRenderer) renderToolCalls(sb *strings.Builder, toolCalls []kellmtype.ToolCall) error {
	for i, tc := range toolCalls {
		if i > 0 {
			sb.WriteString("\n")
		}

		// 创建具有正确结构的工具调用对象
		toolCallObj := map[string]interface{}{
			"name":      tc.Function.Name,
			"arguments": tc.Function.Arguments,
		}

		// 使用排序键进行序列化，保证输出的确定性
		toolCallJSON, err := r.marshalSorted(toolCallObj)
		if err != nil {
			return fmt.Errorf("工具调用 %s 序列化失败: %w", tc.Function.Name, err)
		}

		sb.WriteString("```json\n")
		sb.Write(toolCallJSON)
		sb.WriteString("\n```")
	}

	return nil
}

// marshalSorted 使用排序后的键序列化 map，保证输出的确定性
func (r *GenericRenderer) marshalSorted(v interface{}) ([]byte, error) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return json.MarshalIndent(v, "", "  ")
	}

	// 获取排序后的键
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 手动构建 JSON，保证键的顺序
	var sb strings.Builder
	sb.WriteString("{\n")
	for i, key := range keys {
		if i > 0 {
			sb.WriteString(",\n")
		}

		keyJSON, _ := json.Marshal(key)
		sb.WriteString("  ")
		sb.Write(keyJSON)
		sb.WriteString(": ")

		valueJSON, err := json.MarshalIndent(m[key], "  ", "  ")
		if err != nil {
			return nil, err
		}
		sb.Write(valueJSON)
	}
	sb.WriteString("\n}")

	return []byte(sb.String()), nil
}
