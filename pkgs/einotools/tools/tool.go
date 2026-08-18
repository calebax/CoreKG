package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"

	einomcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/insmtx/corekg/pkgs/einotools/sandbox"
	baidusearch "github.com/insmtx/corekg/pkgs/einotools/tools/baidu"
	codeTool "github.com/insmtx/corekg/pkgs/einotools/tools/code"
	fileTool "github.com/insmtx/corekg/pkgs/einotools/tools/file"
)

type ToolOption string

const (
	ToolOptionFile         ToolOption = "file"
	ToolOptionCode         ToolOption = "code"
	ToolOptionChart        ToolOption = "chart"
	ToolOptionSearch       ToolOption = "search"
	ToolOptionAnalysisFile ToolOption = "analysis_file"
)

type ToolEnv struct {
	Sandbox sandbox.Config     `json:"sandbox"`
	Mcp     map[string]McpInfo `json:"mcp"`
}

type McpInfo struct {
	Mode string `json:"mode"`
	URL  string `json:"url"`
}

func GetTools(ctx context.Context, model model.ToolCallingChatModel, toolArr []ToolOption, saveChartFunc func(string) (uint, error)) (tools []tool.BaseTool, err error) {
	if len(toolArr) == 0 {
		return []tool.BaseTool{}, nil
	}

	toolEnv := &ToolEnv{}
	if err = settings.GetYaml("corekg", "agentenv", toolEnv); err != nil {
		return nil, err
	}

	for _, tool := range toolArr {
		switch tool {
		case ToolOptionCode:
			defaultSandbox, err := sandbox.NewSandbox(&toolEnv.Sandbox)
			if err != nil {
				return nil, err
			}

			codeGenTool, err := codeTool.NewCodeGeneratorTool(ctx, &codeTool.CodeGeneratorConfig{
				CodeModel: model,
				Sandbox:   defaultSandbox,
			})
			if err != nil {
				return nil, err
			}
			tools = append(tools, WrapToolWithErrorHandling(ctx, codeGenTool))

		case ToolOptionFile:
			fileInfoTool, err := fileTool.NewFileInfoTool(ctx, &fileTool.FileInfoConfig{})
			if err != nil {
				return nil, err
			}
			tools = append(tools, WrapToolWithErrorHandling(ctx, fileInfoTool))

			sheetInfoTool, err := fileTool.NewFileTool(ctx, &fileTool.FileToolConfig{})
			if err != nil {
				return nil, err
			}
			tools = append(tools, WrapToolWithErrorHandling(ctx, sheetInfoTool))
		case ToolOptionChart:
			if chartInfo, ok := toolEnv.Mcp["chart"]; ok {
				chartTools, err := GetMcpTools(ctx, chartInfo.Mode, chartInfo.URL)
				if err != nil {
					return nil, err
				}
				for _, t := range chartTools {
					t = WrapToolWithOutputTransform(t, func(ctx context.Context, output string) (string, error) {
						var contentMap map[string]interface{}
						if err = json.Unmarshal([]byte(output), &contentMap); err != nil {
							return "", fmt.Errorf("error unmarshaling chart output: %v", err)
						}
						structuredContent, ok := contentMap["structuredContent"]
						if !ok {
							return "", fmt.Errorf("structuredContent is not a chart")
						}
						// TODO 图表数据持久化，增加返回id，需要 question 和 session
						var id uint
						if saveChartFunc != nil {
							jsonContent, err := json.Marshal(structuredContent)
							if err != nil {
								return "", fmt.Errorf("error marshaling structuredContent: %v", err)
							}
							id, err = saveChartFunc(string(jsonContent))
							if err != nil {
								logs.ErrorContextf(ctx, "saveFunc error: %v", err)
								return "", err
							}
						}
						var chartData = make(map[string]any)
						chartData["chart_id"] = id
						chartData["chart_content"] = structuredContent

						jsonData, err := json.Marshal(chartData)
						if err != nil {
							return "", err
						}
						return string(jsonData), nil
					})
					tools = append(tools, WrapToolWithErrorHandling(ctx, t))
				}
			}
		case ToolOptionSearch:
			apiKey, _ := settings.GetText("corekg", "baidu_bce_api_key")
			searchTool, err := baidusearch.NewBaiduWebSearch(ctx, &baidusearch.Config{
				ApiKey: apiKey,
			})
			if err != nil {
				return nil, err
			}
			tools = append(tools, WrapToolWithErrorHandling(ctx, searchTool))
		case ToolOptionAnalysisFile:
			analysisFileTool, err := fileTool.NewAnalysisFileTool(ctx, &fileTool.AnalysisConfig{
				Model: model,
			})
			if err != nil {
				return nil, err
			}
			tools = append(tools, WrapToolWithErrorHandling(ctx, analysisFileTool))
		default:
			fmt.Printf("⚠️ unknown tool: %s\n", tool)
		}
	}

	return tools, nil
}

func GetMcpTools(ctx context.Context, linkMode string, mcpServerUrl string) (tools []tool.BaseTool, err error) {
	if linkMode != "sse" {
		linkMode = "streamable"
	}
	var cli *client.Client
	if linkMode == "streamable" {
		cli, err = client.NewStreamableHttpClient(mcpServerUrl)
	} else {
		cli, err = client.NewSSEMCPClient(mcpServerUrl)
	}
	if err != nil {
		return nil, err
	}
	err = cli.Start(ctx)
	if err != nil {
		return nil, err
	}

	// 创建初始化请求
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-tools-client",
		Version: "1.0.0",
	}

	// 执行初始化
	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}

	return einomcp.GetTools(ctx, &einomcp.Config{Cli: cli})
}

func GenToolInfos(ctx context.Context, tools []tool.BaseTool) ([]*schema.ToolInfo, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		tl, err := t.Info(ctx)
		if err != nil {
			return nil, err
		}

		toolInfos = append(toolInfos, tl)
	}

	return toolInfos, nil
}

func WrapToolWithErrorHandling(ctx context.Context, innerTool tool.BaseTool) tool.BaseTool {
	toolName := "unknown_tool"
	if info, _ := innerTool.Info(ctx); info != nil && info.Name != "" {
		toolName = info.Name
	}

	errorHandlingTool := utils.WrapToolWithErrorHandler(innerTool, func(ctx context.Context, err error) string {
		errPayload := map[string]any{
			"tool_failed": true,
			"error": map[string]string{
				"tool":    toolName,
				"type":    "execution_error",
				"message": err.Error(),
			},
		}

		b, _ := json.Marshal(errPayload)
		return string(b)
	})
	return errorHandlingTool
}
