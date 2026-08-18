package tools

import (
	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func RegisterForestTools(s *mcpgoserver.MCPServer, client *mcpcommon.InternalClient) {
	s.AddTool(mcpgomcp.NewTool("list_forest",
		mcpgomcp.WithDescription("列出知识库列表，支持分页查询"),
		mcpgomcp.WithNumber("offset", mcpgomcp.Description("分页偏移量")),
		mcpgomcp.WithNumber("limit", mcpgomcp.Description("每页数量")),
		mcpgomcp.WithArray("order_by", mcpgomcp.Description("排序规则，如 created_at desc"), mcpgomcp.WithStringItems()),
	), makeHandler(client, "keapi.ListForest"))

	s.AddTool(mcpgomcp.NewTool("batch_get_forest",
		mcpgomcp.WithDescription("批量查询知识库信息"),
		mcpgomcp.WithArray("forest_ids", mcpgomcp.Required(), mcpgomcp.Description("知识库ID列表"), mcpgomcp.WithNumberItems()),
	), makeHandler(client, "keapi.BatchGetForest"))

	s.AddTool(mcpgomcp.NewTool("create_forest",
		mcpgomcp.WithDescription("创建知识库"),
		mcpgomcp.WithString("name", mcpgomcp.Required(), mcpgomcp.Description("知识库名称")),
		mcpgomcp.WithString("avatar_url", mcpgomcp.Description("知识库头像URL")),
		mcpgomcp.WithString("description", mcpgomcp.Description("知识库描述")),
		mcpgomcp.WithString("forest_type", mcpgomcp.Description("知识库类型，file 或 data，默认 file")),
	), makeHandler(client, "keapi.CreateForest"))

	s.AddTool(mcpgomcp.NewTool("update_forest",
		mcpgomcp.WithDescription("更新知识库信息"),
		mcpgomcp.WithNumber("forest_id", mcpgomcp.Required(), mcpgomcp.Description("知识库ID")),
		mcpgomcp.WithString("name", mcpgomcp.Description("知识库新名称")),
		mcpgomcp.WithString("description", mcpgomcp.Description("知识库新描述")),
	), makeHandler(client, "keapi.UpdateForest"))

	s.AddTool(mcpgomcp.NewTool("delete_forest",
		mcpgomcp.WithDescription("删除知识库"),
		mcpgomcp.WithNumber("forest_id", mcpgomcp.Required(), mcpgomcp.Description("知识库ID")),
	), makeHandler(client, "keapi.DeleteForest"))
}