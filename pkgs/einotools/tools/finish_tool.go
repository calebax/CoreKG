package tools

import (
	"context"
	"fmt"
	"sync/atomic"

	tool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
)

// 用于配置 Final Answer 标记工具
type FinalAnswerMarkerConfig struct {
	ToolName string `json:"tool_name"`
	ToolDesc string `json:"tool_desc"`

	FinalSignal *FinalAnswerSignal `json:"final_signal"`
}

// 内部工具，用于在进入最终回复阶段时标记状态
type FinalAnswerMarkerTool struct {
	signal *FinalAnswerSignal
}

// FinalAnswerSignal 表示“Agent 是否已进入最终回复阶段”
type FinalAnswerSignal struct {
	final atomic.Bool
}

func (s *FinalAnswerSignal) MarkFinal() {
	s.final.Store(true)
}

func (s *FinalAnswerSignal) IsFinal() bool {
	return s.final.Load()
}

// FinalAnswerRequest is the input for FinalAnswerMarkerTool
type FinalAnswerRequest struct {
	// TODO 不必要的参数，是因为llm gateway参数400 "required": null，待修复
	Marker string `json:"marker" jsonschema:"required,enum=final_answer"`
}

// FinalAnswerResponse is the output for FinalAnswerMarkerTool
type FinalAnswerResponse struct {
	Message string `json:"message"`
}

// NewFinalAnswerMarkerTool creates a new FinalAnswerMarkerTool
func NewFinalAnswerMarkerTool(ctx context.Context, conf *FinalAnswerMarkerConfig) (tool.InvokableTool, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if conf.FinalSignal == nil {
		return nil, fmt.Errorf("FinalSignal is required")
	}

	toolName := conf.ToolName
	if toolName == "" {
		toolName = "final_answer_marker"
	}
	toolDesc := conf.ToolDesc
	if toolDesc == "" {
		toolDesc = "Mark the conversation as entering the final answer phase. Use this tool when you have gathered enough information and are ready to provide the final answer to the user. No parameters needed."
	}

	ft := &FinalAnswerMarkerTool{
		signal: conf.FinalSignal,
	}

	return toolutils.InferTool(toolName, toolDesc, ft.invoke)
}

func (t *FinalAnswerMarkerTool) invoke(ctx context.Context, req *FinalAnswerRequest) (*FinalAnswerResponse, error) {
	t.signal.MarkFinal()
	return &FinalAnswerResponse{Message: "Marked as final. Generate the final answer without using any further tools."}, nil
}
