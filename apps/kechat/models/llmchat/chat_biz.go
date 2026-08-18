package llmchat

import (
	"encoding/json"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
)

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// NewStreamOptions 创建带有默认值的 StreamOptions
func NewStreamOptions() *StreamOptions {
	return &StreamOptions{
		IncludeUsage: true,
	}
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type Message struct {
	Role    MessageRole `json:"role"`    // 角色
	Content string      `json:"content"` // 用于保存文本
}

// ChatReqBody 发往llm的请求体
type ChatReqBody struct {
	Messages         []*Message     `json:"messages,omitempty"`
	Model            string         `json:"model,omitempty"`             // 模型
	FrequencyPenalty *float32       `json:"frequency_penalty,omitempty"` // 禁止重复率 -2--2
	MaxTokens        *int           `json:"max_tokens,omitempty"`        // 最大生成token
	PresencePenalty  *float32       `json:"presence_penalty,omitempty"`  // 是否包含禁用词 -2--2
	Temperature      *float32       `json:"temperature,omitempty"`       // 温度 0--2
	Stream           bool           `json:"stream,omitempty"`            // 是否是流式接口
	StreamOptions    *StreamOptions `json:"stream_options,omitempty"`
}

func (m *ChatReqBody) ToString() (string, error) {
	jsonPayload, err := json.Marshal(m)
	if err != nil {
		// logs.Errorf("[ChatReqBody] Failed to convert request body to JSON: %s", err.Error())
		return "", err
	}
	return string(jsonPayload), nil
}

type QaRes struct {
	Usage         Usage  `json:"usage,omitempty"`
	Reasoning     string `json:"reasoning,omitempty"`
	ReasoningTime int    `json:"reasoning_time,omitempty"`
	Content       string `json:"content,omitempty"`
	CostSeconds   int    `json:"cost_seconds,omitempty"`
}

// Usage deepseek的usage使用情况
type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

// WriteResult 流式返回结构体
type WriteResult struct {
	ReasoningContent string                   `json:"reasoning_content"`
	Content          string                   `json:"content"`
	ReasoningSeconds int                      `json:"reasoning_seconds"`
	Reference        *chattype.QueryReference `json:"reference"`
	Flag             FlagAnswer               `json:"flag"`
}
type FlagAnswer string

const (
	FlagThinking       FlagAnswer = "thinking"
	FlagSearching      FlagAnswer = "searching"
	FlagFound          FlagAnswer = "found"
	FlagAnswering      FlagAnswer = "answering"
	StreamsFlagECharts FlagAnswer = "echarts"
	StreamsFlagSQL     FlagAnswer = "sql"
)

func (w WriteResult) String() string {
	jsonData, err := json.Marshal(w)
	if err != nil {
		return "\n"
	}
	return string(jsonData) + "\n"
}

type Reference struct {
	ForestID uint   `json:"forest_id"`
	FileID   uint   `json:"file_id"`
	Name     string `json:"name"`
}
