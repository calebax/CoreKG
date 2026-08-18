package tools

import (
	"context"
	"encoding/base64"
	"strconv"

	mcpgomcp "github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
	"github.com/insmtx/corekg/apps/keapi/internal/mcpcommon"
)

func RegisterFileTools(s *mcpgoserver.MCPServer, client *mcpcommon.InternalClient) {
	s.AddTool(mcpgomcp.NewTool("list_file",
		mcpgomcp.WithDescription("列出知识库下的文档列表"),
		mcpgomcp.WithNumber("forest_id", mcpgomcp.Required(), mcpgomcp.Description("知识库ID")),
		mcpgomcp.WithNumber("offset", mcpgomcp.Description("分页偏移量")),
		mcpgomcp.WithNumber("limit", mcpgomcp.Description("每页数量")),
		mcpgomcp.WithArray("order_by", mcpgomcp.Description("排序规则"), mcpgomcp.WithStringItems()),
	), makeHandler(client, "keapi.ListFile"))

	s.AddTool(mcpgomcp.NewTool("batch_get_file",
		mcpgomcp.WithDescription("批量查询文档信息"),
		mcpgomcp.WithArray("forest_file_ids", mcpgomcp.Required(), mcpgomcp.Description("文档ID列表"), mcpgomcp.WithNumberItems()),
	), makeHandler(client, "keapi.BatchGetFile"))

	s.AddTool(mcpgomcp.NewTool("get_file_chunks",
		mcpgomcp.WithDescription("查询文档的 Chunk 分段内容"),
		mcpgomcp.WithNumber("forest_file_id", mcpgomcp.Required(), mcpgomcp.Description("文档ID")),
		mcpgomcp.WithArray("chunk_sequences", mcpgomcp.Required(), mcpgomcp.Description("Chunk序号列表"), mcpgomcp.WithNumberItems()),
	), makeHandler(client, "keapi.GetFileChunks"))

	s.AddTool(mcpgomcp.NewTool("upload_file",
		mcpgomcp.WithDescription("上传文档到知识库，文件内容以base64编码传入"),
		mcpgomcp.WithNumber("forest_id", mcpgomcp.Required(), mcpgomcp.Description("知识库ID")),
		mcpgomcp.WithString("file_name", mcpgomcp.Required(), mcpgomcp.Description("文件名")),
		mcpgomcp.WithString("file_base64", mcpgomcp.Required(), mcpgomcp.Description("文件内容的base64编码")),
		mcpgomcp.WithNumber("parent_id", mcpgomcp.Description("父目录ID，0为根目录")),
	), uploadFileHandler(client))

	s.AddTool(mcpgomcp.NewTool("preview_file_url",
		mcpgomcp.WithDescription("获取文档的预览或下载URL"),
		mcpgomcp.WithNumber("forest_file_id", mcpgomcp.Required(), mcpgomcp.Description("文档ID")),
	), makeHandler(client, "keapi.PreviewFileByURL"))
}

func uploadFileHandler(client *mcpcommon.InternalClient) mcpgoserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgomcp.CallToolRequest) (*mcpgomcp.CallToolResult, error) {
		apiKey := mcpcommon.RawAPIKeyFromContext(ctx)
		fileBase64 := req.GetString("file_base64", "")
		fileName := req.GetString("file_name", "")
		forestID := strconv.FormatInt(int64(req.GetInt("forest_id", 0)), 10)
		parentID := ""
		pid := req.GetInt("parent_id", 0)
		if pid > 0 {
			parentID = strconv.FormatInt(int64(pid), 10)
		}

		fileBytes, err := base64.StdEncoding.DecodeString(fileBase64)
		if err != nil {
			return mcpgomcp.NewToolResultError("decode base64 failed: " + err.Error()), nil
		}

		data, err := client.UploadFile(ctx, apiKey, forestID, parentID, fileName, fileBytes)
		if err != nil {
			return mcpgomcp.NewToolResultError(err.Error()), nil
		}
		return mcpgomcp.NewToolResultText(string(data)), nil
	}
}