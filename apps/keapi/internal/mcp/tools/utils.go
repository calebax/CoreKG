package tools

import (
	"context"

	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func makeHandler(client *mcpcommon.InternalClient, action string) mcpgoserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgomcp.CallToolRequest) (*mcpgomcp.CallToolResult, error) {
		apiKey := mcpcommon.RawAPIKeyFromContext(ctx)
		data, err := client.CallAPI(ctx, apiKey, action, req.Params.Arguments)
		if err != nil {
			return mcpgomcp.NewToolResultError(err.Error()), nil
		}
		return mcpgomcp.NewToolResultText(string(data)), nil
	}
}