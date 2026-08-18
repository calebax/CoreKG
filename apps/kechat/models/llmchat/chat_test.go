package llmchat

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
)

func TestWithDisableThinkingIfNeeded(t *testing.T) {
	version.SetDeployMode(global.DeployModeOnPremise)
	input := `{"model":"Qwen3.6-35B-A3B","messages":[{"role":"user","content":"hi"}]}`

	result, err := withDisableThinkingIfNeeded(input, "Qwen3.6-35B-A3B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal([]byte(result), &reqBody); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	kwargs, ok := reqBody["chat_template_kwargs"]
	if !ok {
		t.Fatal("expected chat_template_kwargs to be present")
	}
	kwargsMap, ok := kwargs.(map[string]any)
	if !ok {
		t.Fatalf("expected chat_template_kwargs to be map, got %T", kwargs)
	}
	enableThinking, ok := kwargsMap["enable_thinking"]
	if !ok {
		t.Fatal("expected enable_thinking to be present")
	}
	if enableThinking != false {
		t.Fatalf("expected enable_thinking=false, got %v", enableThinking)
	}
}

func TestChatRequest(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}]}`))
	}))
	defer server.Close()

	version.SetDeployMode(global.DeployModeOnPremise)

	w := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(w, gin.New())

	model := &chattype.ChatModel{
		ModelName: "Qwen3.6-35B-A3B",
		ModelUrl:  server.URL,
		APIKey:    "test-api-key",
	}
	req := &ChatReqBody{
		Messages: []*Message{
			{Role: MessageRoleUser, Content: "hi"},
		},
		Stream: false,
	}
	wrapper := NewLLmChatWrapper(ctx, req, model)

	resp, err := wrapper.ChatRequest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	kwargs, ok := payload["chat_template_kwargs"]
	if !ok {
		t.Fatal("expected chat_template_kwargs in payload")
	}
	kwargsMap, ok := kwargs.(map[string]any)
	if !ok {
		t.Fatalf("expected chat_template_kwargs to be map, got %T", kwargs)
	}
	if kwargsMap["enable_thinking"] != false {
		t.Fatalf("expected enable_thinking=false in payload, got %v", kwargsMap["enable_thinking"])
	}
}
