package middleware

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/kellm/drivers"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

func TestParseStreamContentSupportsPureJSONLines(t *testing.T) {
	body := []byte("{\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"created\":1,\"model\":\"demo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hello\"},\"finish_reason\":\"\"}]}\n" +
		"{\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"created\":1,\"model\":\"demo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}]}\n")

	state, err := parseStreamContent(body)
	if err != nil {
		t.Fatalf("parseStreamContent() error = %v", err)
	}
	if got := state.content.String(); got != "hello world" {
		t.Fatalf("parseStreamContent() content = %q, want %q", got, "hello world")
	}
	if state.id != "1" {
		t.Fatalf("parseStreamContent() id = %q, want %q", state.id, "1")
	}
}

func TestAdaptToolCallFallbackStreamResponseReplaysOriginalBodyWhenNoToolCall(t *testing.T) {
	original := "data: {\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"created\":1,\"model\":\"demo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hello\"},\"finish_reason\":\"\"}]}\n\n" +
		"data: {\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"created\":1,\"model\":\"demo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: [DONE]\n\n"

	result, err := AdaptToolCallFallbackStreamResponse(&drivers.ProxyResult{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       []byte(original),
		Stream:     true,
	})
	if err != nil {
		t.Fatalf("AdaptToolCallFallbackStreamResponse() error = %v", err)
	}
	if result.BodyReader == nil {
		t.Fatalf("BodyReader = nil, want replay stream body")
	}
	defer result.BodyReader.Close()

	replayed, err := io.ReadAll(result.BodyReader)
	if err != nil {
		t.Fatalf("ReadAll(BodyReader) error = %v", err)
	}
	if string(replayed) != original {
		t.Fatalf("replayed body = %q, want %q", string(replayed), original)
	}
	if strings.TrimSpace(string(result.Body)) != "" {
		t.Fatalf("Body = %q, want empty when BodyReader is used", string(result.Body))
	}
}

func TestPrepareToolCallFallbackRequestClearsToolChoiceWhenToolsRemoved(t *testing.T) {
	req := &kellmtype.ChatRequestBody{
		Model: "demo",
		Messages: []kellmtype.Message{
			{
				Role: "system",
				Content: kellmtype.MessageContent{
					Text: "你是一个超级智能体。",
				},
			},
			{
				Role: "user",
				Content: kellmtype.MessageContent{
					Text: "计算80000以内，所有偶数减去奇数的差是多少？",
				},
			},
		},
		Tools: []kellmtype.Tool{
			{
				Type: "function",
				Function: kellmtype.Function{
					Name:        "code_executor_tool",
					Description: "Execute code in sandbox",
					Parameters: kellmtype.JSONSchema{
						"type": "object",
						"properties": map[string]any{
							"code": map[string]any{"type": "string"},
						},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	modifiedReq, err := prepareToolCallFallbackRequest(req)
	if err != nil {
		t.Fatalf("prepareToolCallFallbackRequest() error = %v", err)
	}
	if modifiedReq.ToolChoice != nil {
		t.Fatalf("ToolChoice = %#v, want nil", modifiedReq.ToolChoice)
	}
	if len(modifiedReq.Tools) != 0 {
		t.Fatalf("tools len = %d, want 0", len(modifiedReq.Tools))
	}
	if len(modifiedReq.Messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(modifiedReq.Messages))
	}
	if strings.TrimSpace(modifiedReq.Messages[0].Content.Text) == "" {
		t.Fatalf("fallback prompt is empty")
	}
}
