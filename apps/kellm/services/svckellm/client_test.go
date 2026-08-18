package svckellm

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/kellm/drivers/openai"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
	svcmiddleware "github.com/insmtx/corekg/apps/kellm/services/svckellm/middleware"
)

func TestExtractBearerToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		header := http.Header{}
		header.Set("Authorization", "Bearer test-token")

		token, err := ExtractBearerToken(header)
		if err != nil {
			t.Fatalf("ExtractBearerToken() error = %v", err)
		}
		if token != "test-token" {
			t.Fatalf("ExtractBearerToken() token = %q, want %q", token, "test-token")
		}
	})

	t.Run("missing header", func(t *testing.T) {
		_, err := ExtractBearerToken(http.Header{})
		if err != ErrMissingAuthorizationHeader {
			t.Fatalf("ExtractBearerToken() error = %v, want %v", err, ErrMissingAuthorizationHeader)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		header := http.Header{}
		header.Set("Authorization", "Bearer ")

		_, err := ExtractBearerToken(header)
		if err != ErrEmptyAuthorizationToken {
			t.Fatalf("ExtractBearerToken() error = %v, want %v", err, ErrEmptyAuthorizationToken)
		}
	})
}

func TestChatCompletionsURL(t *testing.T) {
	t.Run("append completions path", func(t *testing.T) {
		got := openai.ChatCompletionsURL("https://api.example.com/v1")
		want := "https://api.example.com/v1/chat/completions"
		if got != want {
			t.Fatalf("ChatCompletionsURL() = %q, want %q", got, want)
		}
	})

	t.Run("keep completions path", func(t *testing.T) {
		got := openai.ChatCompletionsURL("https://api.example.com/v1/chat/completions")
		want := "https://api.example.com/v1/chat/completions"
		if got != want {
			t.Fatalf("ChatCompletionsURL() = %q, want %q", got, want)
		}
	})
}

func TestAdaptCustomFCResponse(t *testing.T) {
	payload := kellmtype.ChatResponseBody{
		ID:     "chatcmpl-1",
		Model:  "demo",
		Object: "chat.completion",
		Choices: []kellmtype.Choice{{
			Index:        0,
			FinishReason: "stop",
			Message: kellmtype.ChatMessage{
				Role:    "assistant",
				Content: "调用天气工具\n\n```json\n{\"name\":\"get_weather\",\"arguments\":\"{\\\"location\\\":\\\"西安\\\"}\"}\n```",
			},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	result := &ProxyResult{
		StatusCode: 200,
		Header:     map[string][]string{},
		Body:       body,
	}

	result, err = svcmiddleware.AdaptToolCallFallbackResponse(result)
	if err != nil {
		t.Fatalf("AdaptToolCallFallbackResponse() error = %v", err)
	}

	var adapted kellmtype.ChatResponseBody
	if err := json.Unmarshal(result.Body, &adapted); err != nil {
		t.Fatalf("unmarshal adapted body: %v", err)
	}
	if adapted.Choices[0].FinishReason != "tool_calls" {
		t.Fatalf("finish_reason = %q, want tool_calls", adapted.Choices[0].FinishReason)
	}
	if len(adapted.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool_calls len = %d, want 1", len(adapted.Choices[0].Message.ToolCalls))
	}
}

func TestParseChatRequest(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := ParseChatRequest([]byte("{"))
		if err == nil || !strings.Contains(err.Error(), ErrInvalidRequestBody.Error()) {
			t.Fatalf("ParseChatRequest() error = %v, want contains %q", err, ErrInvalidRequestBody.Error())
		}
	})

	t.Run("missing model", func(t *testing.T) {
		_, err := ParseChatRequest([]byte(`{"messages":[]}`))
		if err != ErrModelRequired {
			t.Fatalf("ParseChatRequest() error = %v, want %v", err, ErrModelRequired)
		}
	})

	t.Run("missing messages", func(t *testing.T) {
		_, err := ParseChatRequest([]byte(`{"model":"demo"}`))
		if err == nil || !strings.Contains(err.Error(), ErrInvalidRequestBody.Error()) {
			t.Fatalf("ParseChatRequest() error = %v, want invalid request body", err)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		_, err := ParseChatRequest([]byte(`{"model":"demo","messages":[{"role":"user","content":"hi"}],"chat_options":{"retry_times":1}}`))
		if err == nil || !strings.Contains(err.Error(), ErrInvalidRequestBody.Error()) {
			t.Fatalf("ParseChatRequest() error = %v, want invalid request body", err)
		}
	})

	t.Run("preserve optional fields without defaults", func(t *testing.T) {
		req, err := ParseChatRequest([]byte(`{"model":"demo","messages":[{"role":"user","content":"hi"}]}`))
		if err != nil {
			t.Fatalf("ParseChatRequest() error = %v", err)
		}
		if req.MaxTokens != nil {
			t.Fatalf("MaxTokens = %v, want nil", *req.MaxTokens)
		}
	})

	t.Run("accept full tool schema", func(t *testing.T) {
		payload := []byte(`{
			"model":"deepseek/deepseek-r1",
			"stream":false,
			"messages":[
				{
					"role":"system",
					"content":"你是一个超级智能体。"
				},
				{
					"role":"user",
					"content":"计算80000以内，所有偶数减去奇数的差是多少？"
				}
			],
			"tools":[
				{
					"type":"function",
					"function":{
						"name":"create_area_chart_option",
						"description":"Create a chart option for area chart.",
						"parameters":{
							"type":"object",
							"properties":{
								"axisYTitle":{
									"description":"Set the y-axis title of chart.",
									"default":"",
									"type":"string"
								},
								"data":{
									"type":"array",
									"minItems":1,
									"items":{
										"type":"object",
										"properties":{
											"category":{"type":"string"},
											"value":{"type":"number"},
											"group":{"type":"string"}
										},
										"required":["category","value"]
									}
								}
							},
							"required":["data"]
						}
					}
				}
			],
			"tool_choice":"auto"
		}`)

		req, err := ParseChatRequest(payload)
		if err != nil {
			t.Fatalf("ParseChatRequest() error = %v", err)
		}
		if len(req.Tools) != 1 {
			t.Fatalf("tools len = %d, want 1", len(req.Tools))
		}
		if got := req.Tools[0].Function.Parameters.Type(); got != "object" {
			t.Fatalf("parameters type = %q, want object", got)
		}
		properties, ok := req.Tools[0].Function.Parameters["properties"].(map[string]any)
		if !ok {
			t.Fatalf("properties missing or wrong type: %#v", req.Tools[0].Function.Parameters["properties"])
		}
		dataSchema, ok := properties["data"].(map[string]any)
		if !ok {
			t.Fatalf("data schema missing or wrong type: %#v", properties["data"])
		}
		if got := dataSchema["minItems"]; got == nil {
			t.Fatalf("minItems missing from data schema: %#v", dataSchema)
		}
	})
}
