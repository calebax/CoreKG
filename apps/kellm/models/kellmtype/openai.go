package kellmtype

import "encoding/json"

type Message struct {
	Content      MessageContent      `json:"content,omitempty"`
	Role         string              `json:"role"`
	Name         string              `json:"name,omitempty"`
	Refusal      string              `json:"refusal,omitempty"`
	Audio        *AudioRef           `json:"audio,omitempty"`
	ToolCallID   string              `json:"tool_call_id,omitempty"`
	ToolCalls    []ToolCall          `json:"tool_calls,omitempty"`
	FunctionCall *ToolCallFunction   `json:"function_call,omitempty"`
	Annotations  []MessageAnnotation `json:"annotations,omitempty"`
}

type ChatMessage struct {
	Content      string              `json:"content,omitempty"`
	Role         string              `json:"role,omitempty"`
	Refusal      string              `json:"refusal,omitempty"`
	Audio        *AudioRef           `json:"audio,omitempty"`
	ToolCalls    []ToolCall          `json:"tool_calls,omitempty"`
	FunctionCall *ToolCallFunction   `json:"function_call,omitempty"`
	Annotations  []MessageAnnotation `json:"annotations,omitempty"`
}

type MessageContent struct {
	Text  string
	Items []MessageContentItem
}

func (m *MessageContent) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		var items []MessageContentItem
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		m.Items = items
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	m.Text = str
	return nil
}

func (m MessageContent) MarshalJSON() ([]byte, error) {
	if len(m.Items) > 0 {
		return json.Marshal(m.Items)
	}
	return json.Marshal(m.Text)
}

type MessageContentItem struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	Refusal    string      `json:"refusal,omitempty"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
	File       *InputFile  `json:"file,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type InputFile struct {
	FileData string `json:"file_data,omitempty"`
	FileID   string `json:"file_id,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type AudioRef struct {
	ID string `json:"id"`
}

type MessageAnnotation struct {
	Type string `json:"type,omitempty"`
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Parameters  JSONSchema `json:"parameters"`
	Strict      bool       `json:"strict,omitempty"`
}

type JSONSchema map[string]any

func (s JSONSchema) Type() string {
	if s == nil {
		return ""
	}

	value, ok := s["type"]
	if !ok {
		return ""
	}

	typ, ok := value.(string)
	if !ok {
		return ""
	}

	return typ
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type ToolChoiceFunction struct {
	Name string `json:"name"`
}

type ToolChoiceObject struct {
	Type     string             `json:"type"`
	Function ToolChoiceFunction `json:"function"`
}

type ChatOptions struct {
	RetryTimes   int      `json:"retry_times,omitempty"`
	BackupModels []string `json:"backup_models,omitempty"`
}

type ChatRequestBody struct {
	Messages            []Message `json:"messages"`
	Model               string    `json:"model"`
	MaxTokens           *int      `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int      `json:"max_completion_tokens,omitempty"`
	N                   *int      `json:"n,omitempty"`
	ParallelToolCalls   *bool     `json:"parallel_tool_calls,omitempty"`
	Store               *bool     `json:"store,omitempty"`
	Stream              bool      `json:"stream,omitempty"`
	StreamOptions       struct {
		IncludeUsage bool `json:"include_usage,omitempty"`
	} `json:"stream_options,omitempty"`
	FrequencyPenalty *float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64        `json:"presence_penalty,omitempty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	Stop             []string        `json:"stop,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	Tools            []Tool          `json:"tools,omitempty"`
	ToolChoice       any             `json:"tool_choice,omitempty"`
	LogitBias        map[string]int  `json:"logit_bias,omitempty"`
	LogProbs         *bool           `json:"logprobs,omitempty"`
	TopLogProbs      *int            `json:"top_logprobs,omitempty"`
	User             string          `json:"user,omitempty"`
	Seed             *int            `json:"seed,omitempty"`
	ServiceTier      string          `json:"service_tier,omitempty"`
	ReasoningEffort  string          `json:"reasoning_effort,omitempty"`
}

type ResponseFormat struct {
	Type       string                    `json:"type,omitempty"`
	JSONSchema *ResponseFormatJSONSchema `json:"json_schema,omitempty"`
}

type ResponseFormatJSONSchema struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
	Strict      bool        `json:"strict,omitempty"`
}

type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ChatResponseBody struct {
	Error             ResponseError `json:"-"`
	ID                string        `json:"id"`
	Choices           []Choice      `json:"choices"`
	Created           int64         `json:"created"`
	Model             string        `json:"model"`
	ServiceTier       string        `json:"service_tier,omitempty"`
	SystemFingerprint string        `json:"system_fingerprint"`
	Object            string        `json:"object"`
	Usage             Usage         `json:"usage"`
}

type ChatStreamResponseBody struct {
	Error             ResponseError  `json:"-"`
	ID                string         `json:"id"`
	Choices           []ChoiceStream `json:"choices"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	ServiceTier       string         `json:"service_tier,omitempty"`
	SystemFingerprint string         `json:"system_fingerprint"`
	Object            string         `json:"object"`
	Usage             *Usage         `json:"usage,omitempty"`
}

type ChoiceStream struct {
	FinishReason string    `json:"finish_reason"`
	Index        int       `json:"index"`
	Delta        Delta     `json:"delta"`
	Logprobs     *Logprobs `json:"logprobs,omitempty"`
}

type Choice struct {
	FinishReason string      `json:"finish_reason"`
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	Logprobs     *Logprobs   `json:"logprobs,omitempty"`
}

type Delta struct {
	Content      string            `json:"content,omitempty"`
	Refusal      string            `json:"refusal,omitempty"`
	ToolCalls    []ToolCall        `json:"tool_calls,omitempty"`
	Role         string            `json:"role,omitempty"`
	FunctionCall *ToolCallFunction `json:"function_call,omitempty"`
}

type ToolCall struct {
	Index    int              `json:"index,omitempty"`
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Logprobs struct {
	Content []LogprobContent `json:"content,omitempty"`
	Refusal []LogprobContent `json:"refusal,omitempty"`
}

type LogprobContent struct {
	Token       string       `json:"token"`
	Logprob     float64      `json:"logprob"`
	Bytes       []byte       `json:"bytes"`
	TopLogprobs []TopLogprob `json:"top_logprobs"`
}

type TopLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes"`
}

type OpenAIErrorEnvelope struct {
	Error OpenAIError `json:"error"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

type ModelType string

const (
	ModelTypeOpenAI ModelType = "openai"
)

type ModelCapabilities struct {
	// 是否支持原生 tool call
	ToolCall bool `json:"tool_call" yaml:"tool_call"`
	// 是否支持流式
	Stream bool `json:"stream" yaml:"stream"`
}

type ModelConfig struct {
	ModelType    ModelType         `json:"model_type" yaml:"model_type"`
	BaseURL      string            `json:"base_url" yaml:"base_url"`
	ModelName    string            `json:"model_name" yaml:"model_name"`
	Capabilities ModelCapabilities `json:"capabilities" yaml:"capabilities"`
	Headers      map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}
