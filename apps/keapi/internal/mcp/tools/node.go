package tools

import (
	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func RegisterNodeTools(s *mcpgoserver.MCPServer, client *mcpcommon.InternalClient) {
	s.AddTool(mcpgomcp.NewTool("create_dir",
		mcpgomcp.WithDescription("在知识库中创建文件夹"),
		mcpgomcp.WithNumber("forest_id", mcpgomcp.Required(), mcpgomcp.Description("知识库ID")),
		mcpgomcp.WithNumber("parent_id", mcpgomcp.Description("父目录ID，0为根目录")),
		mcpgomcp.WithString("name", mcpgomcp.Required(), mcpgomcp.Description("文件夹名称")),
	), makeHandler(client, "keapi.CreateDir"))

	s.AddTool(mcpgomcp.NewTool("rename_path",
		mcpgomcp.WithDescription("重命名知识库中的文件或文件夹"),
		mcpgomcp.WithNumber("forest_file_id", mcpgomcp.Required(), mcpgomcp.Description("文件或文件夹ID")),
		mcpgomcp.WithString("name", mcpgomcp.Required(), mcpgomcp.Description("新名称")),
	), makeHandler(client, "keapi.RenamePath"))

	s.AddTool(mcpgomcp.NewTool("delete_path",
		mcpgomcp.WithDescription("删除知识库中的文件或文件夹"),
		mcpgomcp.WithArray("forest_file_ids", mcpgomcp.Required(), mcpgomcp.Description("要删除的文件或文件夹ID列表"), mcpgomcp.WithNumberItems()),
	), makeHandler(client, "keapi.DeletePath"))
}