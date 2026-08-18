package chattype

import (
	"encoding/json"
)

const (
	DefultChatMaxToken = 4096
	MaxMaxToken        = 8192
	DefultFIMMaxToken  = 1024
)

// Role 定义消息角色的枚举类型
type Role string

// 定义具体的角色枚举值
const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// ContentType 定义消息内容类型的枚举
type ContentType string

// 定义具体的内容类型枚举值
const (
	// 文本类型
	ContentTypeText ContentType = "text"
	// 图片类型
	ContentTypeImage ContentType = "image"
	// 图片URL类型
	ContentTypeImageURL ContentType = "image_url"
	// 视频类型
	ContentTypeVideo ContentType = "video"
)

// Message 消息结构体
type Message struct {
	Content          MessageContent `json:"content"`
	ReasoningContent string         `json:"reasoning_content,omitempty"`
	Role             Role           `json:"role"`
	Prefix           bool           `json:"prefix,omitempty"`
}

// ChatMessage 消息结构体 用于非流式返回
type ChatMessage struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
	Role             Role   `json:"role,omitempty"`
	Prefix           bool   `json:"prefix,omitempty"`
}

// MessageContent 消息内容
type MessageContent struct {
	Text  string
	Items []MessageContentItem
}

// UnmarshalJSON 自定义 JSON 反序列化
func (m *MessageContent) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		// 处理数组格式
		var items []MessageContentItem
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		m.Items = items
	} else {
		// 处理对象格式
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		m.Text = str
	}
	return nil
}

// MarshalJSON 自定义 JSON 序列化
func (m *MessageContent) MarshalJSON() ([]byte, error) {
	if len(m.Items) > 0 {
		// 如果有内容项，序列化为数组
		return json.Marshal(m.Items)
	}
	// 否则序列化为字符串
	return json.Marshal(m.Text)
}

// 检查空内容
func (m MessageContent) IsEmpty() bool {
	return m.Text == "" && len(m.Items) == 0
}

// MessageContentItem 消息内容项
type MessageContentItem struct {
	// text/image_url/video
	Type ContentType `json:"type"`
	// 文本内容
	Text string `json:"text,omitempty"`
	// 图片URL
	ImageURL ImageURL `json:"image_url,omitempty"`
	// 视频URL列表
	Video []string `json:"video,omitempty"`
}

// ImageURL 定义图片URL结构
type ImageURL struct {
	URL string `json:"url"` // 图片公开访问URL
}

// Function 函数结构体
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  struct {
		Type       string `json:"type"`
		Properties map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"properties"`
		Required []string `json:"required"`
	} `json:"parameters"`
}

// Tool Tool结构体
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// ChatOptions chat选项
type ChatOptions struct {
	RetryTimes int `json:"retry_times,omitempty"`
	// 备选模型
	BackupModels []string  `json:"backup_models,omitempty"`
	Input        InputList `json:"input,omitempty"`
}

// ChatRequestBody chat请求结构体
type ChatRequestBody struct {
	ChatOptions   ChatOptions `json:"chat_options"`
	Messages      []Message   `json:"messages"`
	Model         string      `json:"model"`
	LLMModelID    uint        `json:"llm_model_id"`
	MaxTokens     int         `json:"max_tokens"`
	Stream        bool        `json:"stream"`
	StreamOptions struct {
		IncludeUsage bool `json:"include_usage"`
	} `json:"stream_options"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	ResponseFormat   struct {
		Type string `json:"type"`
	} `json:"response_format"`
	Stop        []string `json:"stop"`
	Temperature float64  `json:"temperature"`
	TopP        float64  `json:"top_p"`
	Tools       []Tool   `json:"tools"`
	ToolChoice  string   `json:"tool_choice"`
	LogProbs    bool     `json:"logprobs"`
	TopLogProbs *int     `json:"top_logprobs"`
}

// Pretreat 预处理
func (req *ChatRequestBody) Pretreat() *ChatRequestBody {
	if req.MaxTokens == 0 {
		req.MaxTokens = DefultChatMaxToken
	} else if req.MaxTokens > MaxMaxToken {
		req.MaxTokens = MaxMaxToken
	}
	if req.ResponseFormat.Type == "" {
		req.ResponseFormat.Type = "text"
	}
	if req.Temperature == 0 {
		req.Temperature = 1
	}
	if req.TopP == 0 {
		req.TopP = 1
	}
	if req.ToolChoice == "" {
		req.ToolChoice = "none"
	}
	if req.TopLogProbs != nil && *req.TopLogProbs == 0 {
		req.TopLogProbs = nil
	}
	if req.ChatOptions.RetryTimes <= 0 {
		req.ChatOptions.RetryTimes = 1
	} else if req.ChatOptions.RetryTimes >= 15 {
		req.ChatOptions.RetryTimes = 15
	}
	return req
}

// Usage 使用情况
type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

// ResponseError 错误
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ChatResponseBody 返回体
type ChatResponseBody struct {
	// Error 错误
	Error ResponseError `json:"-"`
	// 唯一标识符，用于标识此次请求。
	ID string `json:"id"`
	// 包含多个选择（通常是模型生成的回复）。
	Choices []Choice `json:"choices"`
	// 时间戳，表示响应创建的时间。
	Created int64 `json:"created"`
	// 使用的模型名称。
	Model string `json:"model"`
	// 系统指纹，可能用于调试或追踪。
	SystemFingerprint string `json:"system_fingerprint"`
	// 对象类型，例如 "chat.completion"。
	Object string `json:"object"`
	// 资源使用情况统计。`
	Usage Usage `json:"usage"`
}

// ImageGenerationRequest 图像生成请求体
type ImageGenerationRequest struct {
	// TODO
}

// ChatStreamResponseBody chat流返回体
type ChatStreamResponseBody struct {
	// Error 错误
	Error ResponseError `json:"-"`
	// 唯一标识符，用于标识此次请求。
	ID string `json:"id"`
	// 包含多个选择（通常是模型生成的回复）。
	Choices []ChoiceStream `json:"choices"`
	// 时间戳，表示响应创建的时间。
	Created int64 `json:"created"`
	// 使用的模型名称。
	Model string `json:"model"`
	// 系统指纹，可能用于调试或追踪。
	SystemFingerprint string `json:"system_fingerprint"`
	// 对象类型，例如 "chat.completion"。
	Object string `json:"object"`
	// 资源使用情况统计。
	Usage *Usage `json:"usage"`
}

func (w ChatStreamResponseBody) String() string {
	jsonData, err := json.Marshal(w)
	if err != nil {
		return "\n"
	}
	return string(jsonData) + "\n"
}

// ChoiceStream 表示一个回复。
type ChoiceStream struct {
	// 回复结束的原因，例如 "stop" 表示正常结束。
	FinishReason string `json:"finish_reason"`
	// 当前选择的索引（从 0 开始）。
	Index int `json:"index"`
	// 模型生成的消息内容。
	Delta Delta `json:"delta"`
	// 日志概率信息，用于分析生成的 token。
	Logprobs Logprobs `json:"logprobs"`
}

// Choice 表示一个回复。
type Choice struct {
	// 回复结束的原因，例如 "stop" 表示正常结束。
	FinishReason string `json:"finish_reason"`
	// 当前选择的索引（从 0 开始）。
	Index int `json:"index"`
	// 模型生成的消息内容。
	Message ChatMessage `json:"message"`
	// 日志概率信息，用于分析生成的 token。
	Logprobs Logprobs `json:"logprobs,omitempty"`
}

// Delta 表示模型生成的消息内容。
type Delta struct {
	// 消息的主要内容。
	Content string `json:"content"`
	// 推理过程的内容（可选）。
	ReasoningContent string `json:"reasoning_content,omitempty"`
	// 调用的工具列表（如函数调用）。
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// 消息的角色，例如 "assistant"。
	Role string `json:"role,omitempty"`
}

// ToolCall 表示一个工具调用（如函数调用）。
type ToolCall struct {
	// 工具调用的唯一标识符。
	ID string `json:"id"`
	// 工具类型，例如 "function"。
	Type string `json:"type"`
	// 具体的函数调用信息。
	Function Function `json:"function"`
}

// Logprobs 表示日志概率信息。
type Logprobs struct {
	// 最高概率的候选 token 列表。
	Content []LogprobContent `json:"content"` // 日志概率内容列表。
}

// LogprobContent 表示单个 token 的日志概率信息。
type LogprobContent struct {
	// 当前 token。
	Token string `json:"token"`
	// 当前 token 的对数概率。
	Logprob float64 `json:"logprob"`
	// 当前 token 的字节表示。
	Bytes []byte `json:"bytes"`
	// 最高概率的候选 token 列表。
	TopLogprobs []TopLogprob `json:"top_logprobs"`
}

// TopLogprob 表示最高概率的候选 token。
type TopLogprob struct {
	// 候选 token。
	Token string `json:"token"`
	// 候选 token 的对数概率。
	Logprob float64 `json:"logprob"`
	// 候选 token 的字节表示。
	Bytes []byte `json:"bytes"`
}

// BalanceInfo 余额信息
type BalanceInfo struct {
	Currency     string `json:"currency"`
	TotalBalance string `json:"total_balance"`
}

// BalanceResponse API 返回的余额信息
type BalanceResponse struct {
	IsAvailable  bool          `json:"is_available"`
	BalanceInfos []BalanceInfo `json:"balance_infos"`
}
