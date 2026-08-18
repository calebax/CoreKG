package example3agent

/*
自定义总结tool。tool调用llm生成摘要。
*/

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/ygpkg/yg-go/logs"
)

type Config struct {
	ToolName  string              `json:"tool_name"` // default: summary_tool
	ToolDesc  string              `json:"tool_desc"` // default: Summarize long text or multiple documents into a concise summary
	ChatModel model.BaseChatModel `json:"-" jsonschema:"description=Injected chat model instance used to perform summarization"`
}

// SummaryRequest 定义输入
type SummaryRequest struct {
	Text      string `json:"text" jsonschema:"required,description=The text content to summarize"`
	MaxTokens int    `json:"max_tokens,omitempty" jsonschema:"description=Maximum tokens for the summary"`
	// Language 指定摘要输出的语言，使用 ISO 639-1 语言代码。
	Language string `json:"language,omitempty" jsonschema:"description=Language code (ISO 639-1) for the summary output, e.g., zh-CN, en-US, ja-JP, fr-FR; default is zh-CN"`
	Style    string `json:"style,omitempty" jsonschema:"description=Summary style, e.g., concise, detailed, bullet-points"`
}

type summaryTool struct {
	conf *Config
}

// ✅ 创建 Tool
func NewTool(ctx context.Context, conf *Config) (tool.InvokableTool, error) {
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	if toolName == "" {
		toolName = "summary_tool"
	}
	if toolDesc == "" {
		toolDesc = "Summarize long text or multiple documents into a concise summary"
	}

	if conf.ChatModel == nil {
		return nil, fmt.Errorf("chat model is nil")
	}

	st := &summaryTool{conf: conf}

	// 通过 InferTool 构造标准的 InvokableTool
	tl, err := utils.InferTool(toolName, toolDesc, st.summarize)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

// ✅ 核心逻辑：生成摘要
func (st *summaryTool) summarize(ctx context.Context, req *SummaryRequest) (*string, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("input text cannot be empty")
	}

	lang := req.Language
	if lang == "" {
		lang = "zh-CN"
	}
	style := req.Style
	if style == "" {
		style = "concise"
	}

	prompt := fmt.Sprintf(
		`请以「%s」的方式，总结以下内容的摘要：
------------------------
%s
------------------------
要求：
1. 摘要语言：%s
2. 保留核心要点，避免冗余。
3. 语气自然、连贯。`,
		req.Style,    // 摘要风格：简洁、详细、要点式等
		req.Text,     // 要总结的内容
		req.Language, // 指定语言
	)

	msgs := []*schema.Message{
		schema.SystemMessage("你是一名专业的文本总结助手。"),
		schema.UserMessage(prompt),
	}

	resp, err := st.conf.ChatModel.Generate(ctx, msgs)
	if err != nil {
		logs.WarnContextf(ctx, "summary tool invoke failed: %v", err)
		return nil, err
	}

	return &resp.Content, nil
}
