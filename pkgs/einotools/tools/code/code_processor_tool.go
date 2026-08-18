package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	tool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/insmtx/corekg/pkgs/einotools/sandbox"
)

// CodeGeneratorConfig defines configuration for the code generator tool.
type CodeGeneratorConfig struct {
	ToolName  string              `json:"tool_name"`  // default: code_generator_tool
	ToolDesc  string              `json:"tool_desc"`  // default: This is a code writing tool used to generate Python code
	Lang      string              `json:"lang"`       // default language, e.g., "python"
	CodeModel model.BaseChatModel `json:"code_model"` // default: deepseek
	Sandbox   sandbox.Sandbox     `json:"sandbox"`    // default: py-sandbox
	Prompt    string              `json:"prompt"`     // optional custom system prompt
	MaxStep   int                 `json:"max_step"`   // maximum retries for generation
}

// CodeGenerateRequest is the input to generate code.
type CodeGenerateRequest struct {
	Task string `json:"task" jsonschema:"required,description=任务需求说明（结构化文本）。必须严格按照指定格式完整填写，作为代码生成的唯一事实来源。该描述需在不依赖任何其他字段的情况下，即可独立生成可执行代码。不得省略任何关键条件，不得假设或推断未提供的信息。\\n\\n【填写规范（必须包含以下全部区块，顺序固定）】\\n1.【任务目标】\\n- 明确说明要完成的具体目标与最终产出形式（例如生成什么结果、输出方式、输出格式）。\\n\\n2.【业务背景与规则】\\n- 描述业务语境、计算或处理规则、特殊约束条件、边界情况处理方式。\\n\\n3.【输入数据来源】\\n- 明确列出所有输入数据来源。每一项需说明数据类型（Excel/CSV/JSON/数据库/API 等）、访问方式（本地路径或 URL）、是否需要鉴权，以及在本任务中的用途。\\n\\n4.【表与字段使用约定】\\n- 说明将使用 TableSchemaJSON 中声明的哪些表、字段或子结构，以及它们在本任务中的语义含义。不得引用任何未在 TableSchemaJSON 中明确声明的表或字段。\\n\\n5.【处理步骤要求】\\n- 用有序列表描述整体处理流程或算法步骤，步骤需明确、可复现、可验证，不得省略关键处理环节。\\n\\n6.【输出要求】\\n- 明确输出内容、结构、字段含义、排序规则，以及输出或展示方式。\\n\\n【强约束】\\n- 所有可用信息必须显式写在本 Task 中。\\n- 不得依赖隐含上下文、历史对话或常识进行补全。\\n- 若 Task 中信息不足以完成任务，必须在代码中显式抛出错误，而非自行补充假设。"`

	TableSchemaJSON string `json:"table_schema_json" jsonschema:"required,description=需要解析的表结构声明(JSON字符串)。用于描述表格中包含的表、子表或信息区块及其字段结构,并带有适量的示例数据。代码执行会严格基于该JSON内容进行解析，不得假设、补充或推断未声明的表结构或字段。"`

	LastGeneratedCode string `json:"last_generated_code,omitempty" jsonschema:"description=仅在此需求上次生成代码执行异常时提供。用于参考或增量优化的上一次完整生成代码内容。"`
	LastErrorMessage  string `json:"last_error_message,omitempty" jsonschema:"description=仅在此需求上次生成代码执行异常时提供。包含上一次执行时捕获的错误输出或异常堆栈，用于辅助分析与修复。"`
}

type codeGeneratorTool struct {
	conf  *CodeGeneratorConfig
	agent *CodeGenerateAgent
}

// NewCodeGeneratorTool creates an InvokableTool that returns generated code using the GenerateCodeAgent.
func NewCodeGeneratorTool(ctx context.Context, conf *CodeGeneratorConfig) (tool.InvokableTool, error) {
	if conf == nil {
		conf = &CodeGeneratorConfig{}
	}
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	if toolName == "" {
		toolName = "code_processor_tool"
	}
	if toolDesc == "" {
		toolDesc = `这是一个专门处理Excel数据分析计算的智能体。
它接收数据明确的处理任务指令，自动生成并执行相应的Python代码来完成任务（主要使用pandas进行数据分析计算）。
功能包括：
读取和分析Excel文件内容
执行数据清洗、筛选和转换操作
进行数据计算、统计和分析
生成返回运行结果。

调用说明：
1. 任务描述必须清晰、完整且结构化，应包含任务目标、操作步骤、输入来源、文件URL、业务背景、相关数据、依赖条件等全部必要信息。
2. 对于需要读取文件的数据，必须明确提供文件URL。
3. 需要提供当前表格结构，包括列名、数据类型、示例值等。
4. 描述应确保系统能够仅基于这些信息生成可执行的代码。
`
	}
	if conf.CodeModel == nil {
		return nil, fmt.Errorf("code model cannot be nil")
	}
	if conf.Lang == "" {
		conf.Lang = "python"
	}
	lang := conf.Lang
	maxStep := conf.MaxStep
	if maxStep <= 0 {
		maxStep = 3
	}

	agentCfg := &CodeGenConfig{
		CodeModel: conf.CodeModel,
		Lang:      lang,
		Prompt:    conf.Prompt,
		MaxStep:   maxStep,
		Sandbox:   conf.Sandbox,
	}

	codeAgent, err := NewGenerateCodeAgent(ctx, agentCfg)
	if err != nil {
		return nil, fmt.Errorf("create code agent failed: %w", err)
	}

	cgt := &codeGeneratorTool{conf: conf, agent: codeAgent}
	tl, err := toolutils.InferTool(toolName, toolDesc, cgt.invoke)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (t *codeGeneratorTool) invoke(ctx context.Context, req *CodeGenerateRequest) (*CodeExecuteResponse, error) {
	if req == nil || req.Task == "" {
		return nil, fmt.Errorf("input task cannot be empty")
	}
	// Add table schema to task
	if req.TableSchemaJSON != "" {
		req.Task = "Task: " + req.Task + fmt.Sprintf("\n当前表格结构:\n%s", req.TableSchemaJSON)
	}

	code, err := t.agent.Generate(ctx, &CodeGenerateTaskRequest{
		Task:      req.Task,
		LastCode:  req.LastGeneratedCode,
		LastError: req.LastErrorMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	resp := &CodeExecuteResponse{}

	execRes, _ := t.conf.Sandbox.Exec(ctx, t.conf.Lang, code)

	resp.ExecStdout = execRes.Stdout
	resp.ExecStderr = execRes.Stderr
	resp.ExecExit = execRes.ExitCode

	if resp.ExecExit != 0 || resp.ExecStderr != "" {
		resp.ExecCode = code
	}

	return resp, nil
}
