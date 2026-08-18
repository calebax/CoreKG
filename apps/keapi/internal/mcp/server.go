package mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/apps/keapi/conf"
	"github.com/insmtx/corekg/apps/keapi/internal/mcp/tools"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/logs"
)

type APIKeyValidator interface {
	GetAPIKeyInfo(ctx context.Context, key string) (*accounttype.APIKey, error)
}

type apiKeyValidatorImpl struct{}

func (apiKeyValidatorImpl) GetAPIKeyInfo(ctx context.Context, key string) (*accounttype.APIKey, error) {
	return apikey.GetAPIKeyInfo(ctx, key)
}

func RegistryRouter(eng *server.Router) {
	client := mcpcommon.NewInternalClient(conf.MCPCfg.Addr)
	mcpServer := newMCPServer(client, apiKeyValidatorImpl{})
	registerStreamableHTTPServer(eng, mcpServer)
}

func newMCPServer(client *mcpcommon.InternalClient, validator APIKeyValidator) *mcpgoserver.MCPServer {
	mcpServer := mcpgoserver.NewMCPServer(
		"keapi-mcp-server",
		"v1.0.0",
		mcpgoserver.WithToolCapabilities(true),
		mcpgoserver.WithInstructions("KEAPI Knowledge Base MCP Server - provides knowledge base management, document management, chat, and search tools. All tools require API key authentication."),
		mcpgoserver.WithToolHandlerMiddleware(authMiddleware(validator)),
		mcpgoserver.WithRecovery(),
	)

	tools.RegisterAllTools(mcpServer, client)

	return mcpServer
}

func registerStreamableHTTPServer(eng *server.Router, mcpServer *mcpgoserver.MCPServer) {
	streamableServer := mcpgoserver.NewStreamableHTTPServer(mcpServer,
		mcpgoserver.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			rawKey := extractBearerToken(r)
			if rawKey != "" {
				ctx = mcpcommon.ContextWithRawAPIKey(ctx, rawKey)
			}
			return ctx
		}),
	)

	eng.Any("keapi/mcp", gin.WrapH(streamableServer))
}

func authMiddleware(validator APIKeyValidator) mcpgoserver.ToolHandlerMiddleware {
	return func(next mcpgoserver.ToolHandlerFunc) mcpgoserver.ToolHandlerFunc {
		return func(ctx context.Context, req mcpgomcp.CallToolRequest) (*mcpgomcp.CallToolResult, error) {
			rawKey := mcpcommon.RawAPIKeyFromContext(ctx)
			if rawKey == "" {
				return mcpgomcp.NewToolResultError("unauthorized: missing API key"), nil
			}
			keyInfo, err := validator.GetAPIKeyInfo(ctx, rawKey)
			if err != nil {
				logs.ErrorContextf(ctx, "[mcp_auth] validate API key failed: %v", err)
				return mcpgomcp.NewToolResultError("unauthorized: invalid API key"), nil
			}
			if keyInfo.Status != accounttype.AccessKeyStatusNormal {
				return mcpgomcp.NewToolResultError("unauthorized: API key not active"), nil
			}
			if keyInfo.IsExpired() {
				return mcpgomcp.NewToolResultError("unauthorized: API key expired"), nil
			}
			return next(ctx, req)
		}
	}
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
