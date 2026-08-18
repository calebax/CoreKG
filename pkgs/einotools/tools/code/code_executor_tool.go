package tools

import (
	"context"
	"fmt"

	tool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/insmtx/corekg/pkgs/einotools/sandbox"
)

// CodeExecutorConfig defines configuration for the code executor tool.
type CodeExecutorConfig struct {
	ToolName string `json:"tool_name"` // default: code_executor_tool
	ToolDesc string `json:"tool_desc"` // default: Execute Python code in sandbox
	Sandbox  sandbox.Sandbox
}

// CodeExecuteRequest contains Python code to run in sandbox.
type CodeExecuteRequest struct {
	Lang string `json:"lang" jsonschema:"required,description=Programming language used to run the code"`
	Code string `json:"code" jsonschema:"required,description=The code to run"`
}

// CodeExecuteResponse returns syntax check and execution results.
type CodeExecuteResponse struct {
	ExecCode   string `json:"exec_code"`
	ExecStdout string `json:"exec_stdout"`
	ExecStderr string `json:"exec_stderr"`
	ExecExit   int    `json:"exec_exit"`
}

type codeExecutorTool struct {
	conf    *CodeExecutorConfig
	sandbox sandbox.Sandbox
}

// NewCodeExecutorTool creates an InvokableTool that validates and executes Python code inside py-sandbox.
func NewCodeExecutorTool(ctx context.Context, conf *CodeExecutorConfig) (tool.InvokableTool, error) {
	if conf == nil {
		conf = &CodeExecutorConfig{}
	}
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	if toolName == "" {
		toolName = "code_executor_tool"
	}
	if toolDesc == "" {
		toolDesc = `Execute code in an isolated sandbox environment.
This tool should be invoked whenever a task requires running code to complete.
`
	}

	cet := &codeExecutorTool{conf: conf, sandbox: conf.Sandbox}
	tl, err := toolutils.InferTool(toolName, toolDesc, cet.invoke)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (t *codeExecutorTool) invoke(ctx context.Context, req *CodeExecuteRequest) (*CodeExecuteResponse, error) {
	if req == nil || req.Code == "" {
		return nil, fmt.Errorf("input code cannot be empty")
	}

	resp := &CodeExecuteResponse{}

	execRes, _ := t.sandbox.Exec(ctx, req.Lang, req.Code)

	resp.ExecStdout = execRes.Stdout
	resp.ExecStderr = execRes.Stderr
	resp.ExecExit = execRes.ExitCode

	return resp, nil
}
