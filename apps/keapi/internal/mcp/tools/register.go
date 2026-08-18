package tools

import (
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func RegisterAllTools(s *mcpgoserver.MCPServer, client *mcpcommon.InternalClient) {
	RegisterForestTools(s, client)
	RegisterFileTools(s, client)
	RegisterNodeTools(s, client)
	RegisterChatTools(s, client)
	RegisterSearchTools(s, client)
}