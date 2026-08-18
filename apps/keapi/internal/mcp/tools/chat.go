package tools

import (
	"context"
	"fmt"

	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func RegisterChatTools(s *mcpgoserver.MCPServer, client *mcpcommon.InternalClient) {
	s.AddTool(mcpgomcp.NewTool("create_chat",
		mcpgomcp.WithDescription("创建对话会话，关联指定文档"),
		mcpgomcp.WithArray("forest_file_ids", mcpgomcp.Required(), mcpgomcp.Description("关联的文档ID列表"), mcpgomcp.WithNumberItems()),
		mcpgomcp.WithString("name", mcpgomcp.Description("会话名称")),
	), makeHandler(client, "keapi.CreateChat"))

	s.AddTool(mcpgomcp.NewTool("batch_get_chat_info",
		mcpgomcp.WithDescription("批量查询对话会话信息"),
		mcpgomcp.WithArray("session_ids", mcpgomcp.Required(), mcpgomcp.Description("会话ID列表"), mcpgomcp.WithNumberItems()),
	), makeHandler(client, "keapi.BatchGetChatInfo"))

	s.AddTool(mcpgomcp.NewTool("update_chat_name",
		mcpgomcp.WithDescription("更新对话会话名称"),
		mcpgomcp.WithNumber("session_id", mcpgomcp.Required(), mcpgomcp.Description("会话ID")),
		mcpgomcp.WithString("name", mcpgomcp.Required(), mcpgomcp.Description("新会话名称")),
	), makeHandler(client, "keapi.UpdateChatName"))

	s.AddTool(mcpgomcp.NewTool("delete_chat",
		mcpgomcp.WithDescription("删除对话会话"),
		mcpgomcp.WithNumber("session_id", mcpgomcp.Required(), mcpgomcp.Description("会话ID")),
	), makeHandler(client, "keapi.DeleteChat"))

	s.AddTool(mcpgomcp.NewTool("create_chat_message",
		mcpgomcp.WithDescription("在对话会话中创建用户消息"),
		mcpgomcp.WithNumber("session_id", mcpgomcp.Required(), mcpgomcp.Description("会话ID")),
		mcpgomcp.WithString("content", mcpgomcp.Required(), mcpgomcp.Description("消息内容")),
	), makeHandler(client, "keapi.CreateChatMessage"))

	s.AddTool(mcpgomcp.NewTool("list_chat_messages",
		mcpgomcp.WithDescription("查询对话会话的消息列表"),
		mcpgomcp.WithNumber("session_id", mcpgomcp.Required(), mcpgomcp.Description("会话ID")),
	), makeHandler(client, "keapi.ListChatMessages"))

	s.AddTool(mcpgomcp.NewTool("chat_completions",
		mcpgomcp.WithDescription("基于知识库文档进行对话补全，返回完整答案（非流式）"),
		mcpgomcp.WithArray("forest_file_ids", mcpgomcp.Description("关联的文档ID列表"), mcpgomcp.WithNumberItems()),
		mcpgomcp.WithNumber("session_id", mcpgomcp.Description("已有会话ID，复用会话时传入")),
		mcpgomcp.WithArray("messages", mcpgomcp.Description("对话消息列表，每条消息包含 role 和 content 字段")),
	), chatCompletionsHandler(client))
}

func chatCompletionsHandler(client *mcpcommon.InternalClient) mcpgoserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgomcp.CallToolRequest) (*mcpgomcp.CallToolResult, error) {
		apiKey := mcpcommon.RawAPIKeyFromContext(ctx)
		args, ok := req.Params.Arguments.(map[string]any)
		if !ok {
			return mcpgomcp.NewToolResultError("invalid arguments format"), nil
		}

		reqBody := map[string]any{
			"forest_file_id": args["forest_file_ids"],
			"session_id":     args["session_id"],
			"messages":       args["messages"],
			"stream":         false,
		}

		data, err := client.CallAPI(ctx, apiKey, fmt.Sprintf("keapi.chat/chat/completions"), reqBody)
		if err != nil {
			return mcpgomcp.NewToolResultError(err.Error()), nil
		}
		return mcpgomcp.NewToolResultText(string(data)), nil
	}
}