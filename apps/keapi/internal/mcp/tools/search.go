package tools

import (
	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func RegisterSearchTools(s *mcpgoserver.MCPServer, client *mcpcommon.InternalClient) {
	s.AddTool(mcpgomcp.NewTool("search",
		mcpgomcp.WithDescription("在知识库中检索相关内容"),
		mcpgomcp.WithArray("forest_ids", mcpgomcp.Required(), mcpgomcp.Description("知识库ID列表"), mcpgomcp.WithNumberItems()),
		mcpgomcp.WithString("query", mcpgomcp.Required(), mcpgomcp.Description("检索关键词")),
	), makeHandler(client, "keapi.Search"))
}