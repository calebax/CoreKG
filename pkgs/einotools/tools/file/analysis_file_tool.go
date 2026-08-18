package file

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	tool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/pkgs/einotools/filecontent"
	"github.com/ygpkg/yg-go/logs"
)

const (
	defaultSystemPrompt = `你是一名资深的文件分析助手。

你的任务是：基于提供的文件名称、关注点和文件内容，围绕“关注点”进行定向信息抽取与分析。

执行要求：
1. 始终以“关注点”为核心，仅提取与其直接相关的内容。
2. 优先抽取要点、结论、关键数据、核心论据、定义或操作步骤。
3. 若文件内容存在截断、缺失或不完整情况，必须在回答开头明确标注“内容可能不完整”，不得进行臆测或补全。
4. 输出为简洁的中文要点列表；仅在必要时对要点作简短说明。
5. 避免复述原文，避免无关背景，避免泛化总结。

输出目标：
提供围绕关注点的高密度信息摘要与分析结果。`
	defaultToolName        = "analysis_file_tool"
	defaultToolDescription = "读取远程文件内容，结合指定关注点进行分析并提取关键信息。"
	defaultMaxContentBytes = int64(200_000)
	defaultMaxIterations   = 6
)

// AnalysisConfig defines configuration for the analysis tool.
type AnalysisConfig struct {
	ToolName      string                     `json:"tool_name"`
	ToolDesc      string                     `json:"tool_desc"`
	Model         model.ToolCallingChatModel `json:"model"`
	SystemPrompt  string                     `json:"system_prompt"`
	MaxIterations int                        `json:"max_iterations"`
	MaxReadBytes  int64                      `json:"max_read_bytes"`
}

type AnalysisRequest struct {
	// 必填：文件名称
	FileName string `json:"file_name" jsonschema:"required,description=File name of the remote file, e.g. report.pdf or article.md."`
	// 必填：远程 URL
	FileURL string `json:"file_url" jsonschema:"required,pattern=^https?://,description=Remote URL of the file to analyze.Must start with http:// or https://."`
	// 必填：描述“需要提取文章内容的重点”
	Focus string `json:"focus" jsonschema:"required,description=Describe what to focus on when extracting key content. Example: '提取文章的核心观点、关键数据、结论与建议'."`
}

type AnalysisResponse struct {
	FileName string `json:"file_name"`
	FileUrl  string `json:"file_url"`
	Analysis string `json:"analysis"`
}

type analysisFileTool struct {
	runner       *adk.Runner
	maxReadBytes int64
}

// NewAnalysisFileTool creates an InvokableTool that reads a remote file and asks the model to extract focused insights.
func NewAnalysisFileTool(ctx context.Context, conf *AnalysisConfig) (tool.InvokableTool, error) {
	if conf == nil {
		conf = &AnalysisConfig{}
	}
	if conf.ToolName == "" {
		conf.ToolName = defaultToolName
	}
	if conf.ToolDesc == "" {
		conf.ToolDesc = defaultToolDescription
	}
	if conf.Model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}
	if conf.SystemPrompt == "" {
		conf.SystemPrompt = defaultSystemPrompt
	}
	if conf.MaxIterations <= 0 {
		conf.MaxIterations = defaultMaxIterations
	}
	if conf.MaxReadBytes <= 0 {
		conf.MaxReadBytes = defaultMaxContentBytes
	}

	runner, err := buildRunner(ctx, conf)
	if err != nil {
		logs.ErrorContextf(ctx, "build runner failed: %v", err)
		return nil, err
	}

	aft := &analysisFileTool{runner: runner, maxReadBytes: conf.MaxReadBytes}
	tl, err := toolutils.InferTool(conf.ToolName, conf.ToolDesc, aft.invoke)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (t *analysisFileTool) invoke(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.FileURL) == "" {
		return nil, fmt.Errorf("file url is required")
	}
	if strings.TrimSpace(req.Focus) == "" {
		return nil, fmt.Errorf("focus is required")
	}

	content, filename, truncated, err := filecontent.Read(ctx, req.FileURL, t.maxReadBytes)
	if err != nil {
		return nil, err
	}

	fileNameStr := req.FileName
	if fileNameStr == "" {
		fileNameStr = filename
	}
	userPrompt := buildUserPrompt(fileNameStr, req.FileURL, req.Focus, content, truncated)

	iter := t.runner.Run(ctx, []adk.Message{schema.UserMessage(userPrompt)})

	var analysis strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return nil, event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		output := event.Output.MessageOutput
		if output.Message != nil && output.Message.Role == schema.Assistant {
			analysis.WriteString(output.Message.Content)
		}
	}

	result := strings.TrimSpace(analysis.String())
	if result == "" {
		return nil, fmt.Errorf("analysis result is empty")
	}

	return &AnalysisResponse{
		FileName: req.FileName,
		FileUrl:  req.FileURL,
		Analysis: result,
	}, nil
}

func buildRunner(ctx context.Context, conf *AnalysisConfig) (*adk.Runner, error) {

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "file_analysis_agent",
		Description:   "Extracts information relevant to specified focus areas from files and provides structured analysis results.",
		Instruction:   conf.SystemPrompt,
		Model:         conf.Model,
		MaxIterations: conf.MaxIterations,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model agent failed: %w", err)
	}

	return adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	}), nil
}

func buildUserPrompt(fileName, fileURL, focus, content string, truncated bool) string {
	var b strings.Builder
	if fileName != "" {
		fmt.Fprintf(&b, "文件名称: %s\n", fileName)
	}
	fmt.Fprintf(&b, "文件地址: %s\n", fileURL)
	fmt.Fprintf(&b, "关注点: %s\n", focus)
	if truncated {
		b.WriteString("提示: 文件内容已截断，仅包含前段内容，请谨慎推断。\n")
	}
	b.WriteString("\n文件内容:\n")
	b.WriteString(content)
	return b.String()
}
