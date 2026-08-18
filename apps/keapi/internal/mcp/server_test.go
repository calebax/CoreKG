package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

// mockAPIKeyValidator implements APIKeyValidator for testing.
type mockAPIKeyValidator struct {
	keys map[string]*accounttype.APIKey
}

func (m *mockAPIKeyValidator) GetAPIKeyInfo(_ context.Context, key string) (*accounttype.APIKey, error) {
	k, ok := m.keys[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return k, nil
}

func newMockValidator() *mockAPIKeyValidator {
	now := time.Now()
	return &mockAPIKeyValidator{
		keys: map[string]*accounttype.APIKey{
			"valid-key": {
				APIKey: "valid-key",
				Status: accounttype.AccessKeyStatusNormal,
			},
			"disabled-key": {
				APIKey: "disabled-key",
				Status: accounttype.AccessKeyStatusDisabled,
			},
			"expired-key": {
				APIKey:  "expired-key",
				Status:  accounttype.AccessKeyStatusNormal,
				ExpiredAt: &now,
			},
			"never-expire-key": {
				APIKey: "never-expire-key",
				Status: accounttype.AccessKeyStatusNormal,
			},
		},
	}
}

// mockInternalAPIHandler returns an http.Handler that simulates the internal API backend.
// It responds to specific action paths with mock data.
func mockInternalAPIHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v3/keapi.ListForest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"response": map[string]any{
				"total": 2,
				"items": []map[string]any{
					{"id": 1, "name": "知识库A"},
					{"id": 2, "name": "知识库B"},
				},
			},
		})
	})
	mux.HandleFunc("/v3/keapi.BatchGetForest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"response": []map[string]any{
				{"id": 1, "name": "知识库A"},
			},
		})
	})
	mux.HandleFunc("/v3/keapi.UploadFile", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"response": map[string]any{
				"file_id": 100,
				"name":    "uploaded.txt",
			},
		})
	})
	mux.HandleFunc("/v3/keapi.chat/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"response": map[string]any{
				"reply":   "这是一条回复",
				"stream":  false,
				"session_id": body["session_id"],
			},
		})
	})
	return mux
}

// setupMCPServer creates a full MCP server with mock dependencies for testing.
func setupMCPServer(t *testing.T) (*httptest.Server, *httptest.Server) {
	t.Helper()

	mockBackend := httptest.NewServer(mockInternalAPIHandler())
	t.Cleanup(mockBackend.Close)

	client := mcpcommon.NewInternalClient(mockBackend.URL)

	mcpServer := newMCPServer(client, newMockValidator())

	testMCPServer := mcpgoserver.NewTestStreamableHTTPServer(mcpServer,
		mcpgoserver.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			rawKey := extractBearerToken(r)
			if rawKey != "" {
				ctx = mcpcommon.ContextWithRawAPIKey(ctx, rawKey)
			}
			return ctx
		}),
	)
	t.Cleanup(testMCPServer.Close)

	return testMCPServer, mockBackend
}

// setupMCPClient creates an MCP client connected to the test server.
func setupMCPClient(t *testing.T, serverURL string, apiKey string) *mcpgoclient.Client {
	t.Helper()

	var opts []transport.StreamableHTTPCOption
	if apiKey != "" {
		opts = append(opts, transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + apiKey,
		}))
	}

	cli, err := mcpgoclient.NewStreamableHttpClient(serverURL, opts...)
	if err != nil {
		t.Fatalf("Failed to create MCP client: %v", err)
	}
	t.Cleanup(func() { cli.Close() })

	ctx := context.Background()
	if err := cli.Start(ctx); err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	initReq := mcpgomcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgomcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgomcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}
	_, err = cli.Initialize(ctx, initReq)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	return cli
}

func TestAuthMissingAPIKey(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name:      "list_forest",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result, got success")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "unauthorized") {
		t.Fatalf("expected unauthorized error, got: %s", textContent.Text)
	}
}

func TestAuthInvalidAPIKey(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "non-existent-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name:      "list_forest",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result, got success")
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "invalid API key") {
		t.Fatalf("expected invalid API key error, got: %s", textContent.Text)
	}
}

func TestAuthDisabledAPIKey(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "disabled-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name:      "list_forest",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result, got success")
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "not active") {
		t.Fatalf("expected not active error, got: %s", textContent.Text)
	}
}

func TestAuthExpiredAPIKey(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "expired-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name:      "list_forest",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result, got success")
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "expired") {
		t.Fatalf("expected expired error, got: %s", textContent.Text)
	}
}

func TestToolListForest(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "valid-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name:      "list_forest",
			Arguments: map[string]any{"offset": 0, "limit": 10},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		textContent, _ := result.Content[0].(mcpgomcp.TextContent)
		t.Fatalf("expected success, got error: %s", textContent.Text)
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "知识库A") {
		t.Fatalf("expected 知识库A in response, got: %s", textContent.Text)
	}
}

func TestToolUploadFile(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "valid-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name: "upload_file",
			Arguments: map[string]any{
				"forest_id":    1,
				"file_name":    "test.txt",
				"file_base64":  "aGVsbG8gd29ybGQ=",
				"parent_id":    0,
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		textContent, _ := result.Content[0].(mcpgomcp.TextContent)
		t.Fatalf("expected success, got error: %s", textContent.Text)
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "file_id") {
		t.Fatalf("expected file_id in response, got: %s", textContent.Text)
	}
}

func TestToolUploadFileInvalidBase64(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "valid-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name: "upload_file",
			Arguments: map[string]any{
				"forest_id":    1,
				"file_name":    "test.txt",
				"file_base64":  "not-valid-base64!!!",
				"parent_id":    0,
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for invalid base64, got success")
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "base64") {
		t.Fatalf("expected base64 decode error, got: %s", textContent.Text)
	}
}

func TestToolChatCompletions(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "valid-key")

	ctx := context.Background()
	result, err := cli.CallTool(ctx, mcpgomcp.CallToolRequest{
		Params: mcpgomcp.CallToolParams{
			Name: "chat_completions",
			Arguments: map[string]any{
				"forest_file_ids": []any{1, 2},
				"session_id":      float64(123),
				"messages": []any{
					map[string]any{"role": "user", "content": "你好"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		textContent, _ := result.Content[0].(mcpgomcp.TextContent)
		t.Fatalf("expected success, got error: %s", textContent.Text)
	}
	textContent, ok := result.Content[0].(mcpgomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "回复") {
		t.Fatalf("expected reply in response, got: %s", textContent.Text)
	}
}

func TestToolRegistrationComplete(t *testing.T) {
	testMCPServer, _ := setupMCPServer(t)
	cli := setupMCPClient(t, testMCPServer.URL, "valid-key")

	ctx := context.Background()
	result, err := cli.ListTools(ctx, mcpgomcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	expectedTools := map[string]bool{
		"list_forest":        false,
		"batch_get_forest":   false,
		"create_forest":      false,
		"update_forest":      false,
		"delete_forest":      false,
		"list_file":          false,
		"batch_get_file":     false,
		"get_file_chunks":    false,
		"upload_file":        false,
		"preview_file_url":   false,
		"create_dir":         false,
		"rename_path":        false,
		"delete_path":        false,
		"create_chat":        false,
		"batch_get_chat_info": false,
		"update_chat_name":   false,
		"delete_chat":        false,
		"create_chat_message": false,
		"list_chat_messages":  false,
		"chat_completions":   false,
		"search":             false,
	}

	for _, tool := range result.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		} else {
			t.Errorf("unexpected tool registered: %s", tool.Name)
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool %q not registered", name)
		}
	}
}