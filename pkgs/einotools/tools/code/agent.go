package tools

import (
	"context"
	"regexp"
	"strings"

	sysPrompt "github.com/insmtx/corekg/pkgs/einotools/prompt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/pkgs/einotools/sandbox"
)

const (
	nodeKeyPrepareInput      = "prepare_input"
	nodeKeyCodeGenerateModel = "code_generate_model"
	nodeKeyExtractCode       = "extract_code" // 提取代码
)

type codeGenState struct {
	Requirement    string // 需求描述
	LastCode       string // 最后一次生成的代码
	LastError      string // 最后一次生成的错误信息
	IterationCount int    // 重试次数
	Lang           string // 生成代码的语言
	Sandbox        sandbox.Sandbox
}

type CodeGenConfig struct {
	CodeModel model.BaseChatModel
	Sandbox   sandbox.Sandbox
	Lang      string // 生成代码的语言，如 "go"、"python" 等
	Prompt    string // 自定义提示词，可选，默认使用 DefaultCodeGenPrompt
	MaxStep   int    // 最大重试次数
}

type CodeGenerateAgent struct {
	// 根据需求，生成最终代码
	runnable compose.Runnable[*CodeGenerateTaskRequest, string]
	graph    *compose.Graph[*CodeGenerateTaskRequest, string]
}

type CodeGenerateTaskRequest struct {
	Task      string `json:"code,omitempty" jsonschema:"description=生成的代码内容"`
	LastError string `json:"last_error,omitempty" jsonschema:"description=上传执行的错误信息"`
	LastCode  string `json:"last_code,omitempty" jsonschema:"description=上次执行的完整代码"`
}

// NewGenerateCodeAgent 创建一个新的代码生成智能体
func NewGenerateCodeAgent(ctx context.Context, cfg *CodeGenConfig) (*CodeGenerateAgent, error) {
	var (
		genCodePrompt = cfg.Prompt
		maxStep       = cfg.MaxStep
	)
	if len(genCodePrompt) == 0 {
		genCodePrompt = sysPrompt.CodeGenerateSystemPrompt
	}
	if maxStep <= 0 {
		maxStep = 5
	}

	chatTemplate := prompt.FromMessages(schema.GoTemplate, schema.SystemMessage(genCodePrompt),
		schema.UserMessage(sysPrompt.CodeGenMessageTemplate),
	)

	genStage := func(ctx context.Context) *codeGenState {
		return &codeGenState{
			Lang:    cfg.Lang,
			Sandbox: cfg.Sandbox,
		}
	}

	prepareInput := func(ctx context.Context, input any) ([]*schema.Message, error) {
		_ = compose.ProcessState(ctx, func(_ context.Context, state *codeGenState) error {
			switch v := input.(type) {
			case *CodeGenerateTaskRequest:
				state.Requirement = v.Task
				state.LastCode = v.LastCode
				state.LastError = v.LastError
			default:
			}
			return nil
		})
		return []*schema.Message{schema.UserMessage("消息占位")}, nil
	}

	modelPreHandler := func(ctx context.Context, input []*schema.Message, state *codeGenState) ([]*schema.Message, error) {
		params := map[string]any{
			"lang":        state.Lang,
			"requirement": state.Requirement,
			"lastCode":    state.LastCode,
			"lastError":   state.LastError,
		}
		messages, err := chatTemplate.Format(ctx, params)
		if err != nil {
			return nil, err
		}
		return messages, nil
	}

	g := compose.NewGraph[*CodeGenerateTaskRequest, string](compose.WithGenLocalState(genStage))

	_ = g.AddLambdaNode(nodeKeyPrepareInput, compose.InvokableLambda(prepareInput), compose.WithNodeName(nodeKeyPrepareInput))
	_ = g.AddChatModelNode(nodeKeyCodeGenerateModel,
		cfg.CodeModel,
		compose.WithStatePreHandler(modelPreHandler),
		compose.WithNodeName(nodeKeyCodeGenerateModel))
	_ = g.AddLambdaNode(nodeKeyExtractCode, compose.InvokableLambda(extractCode), compose.WithNodeName(nodeKeyExtractCode))

	_ = g.AddEdge(compose.START, nodeKeyPrepareInput)
	_ = g.AddEdge(nodeKeyPrepareInput, nodeKeyCodeGenerateModel)
	_ = g.AddBranch(nodeKeyCodeGenerateModel, compose.NewGraphBranch(validateCode, map[string]bool{
		nodeKeyPrepareInput: true,
		nodeKeyExtractCode:  true,
	}))
	_ = g.AddEdge(nodeKeyExtractCode, compose.END)

	runnable, err := g.Compile(ctx, compose.WithNodeTriggerMode(compose.AnyPredecessor), compose.WithMaxRunSteps(maxStep))
	if err != nil {
		return nil, err
	}

	return &CodeGenerateAgent{
		runnable: runnable,
	}, nil
}

func validateCode(ctx context.Context, msg *schema.Message) (endNode string, err error) {
	code := RemoveThinkBlocks(msg.Content)
	if code == "" {
		return nodeKeyPrepareInput, nil
	}

	msg.Content = code

	var sandbox sandbox.Sandbox
	var lang string
	_ = compose.ProcessState(ctx, func(_ context.Context, state *codeGenState) error {
		sandbox = state.Sandbox
		lang = state.Lang
		return nil
	})

	chkResult, err := sandbox.CheckSyntax(ctx, lang, code)
	if err == nil && chkResult.Valid {
		return nodeKeyExtractCode, nil
	}

	_ = compose.ProcessState(ctx, func(_ context.Context, state *codeGenState) error {
		state.LastError = chkResult.Stderr
		state.LastCode = code
		state.IterationCount++
		return nil
	})
	return nodeKeyPrepareInput, nil
}

func extractCode(ctx context.Context, msg *schema.Message) (string, error) {
	return RemoveThinkBlocks(msg.Content), nil
}

func (r *CodeGenerateAgent) Generate(ctx context.Context, request *CodeGenerateTaskRequest, opts ...agent.AgentOption) (output string, err error) {
	return r.runnable.Invoke(ctx, request, agent.GetComposeOptions(opts...)...)
}

func (r *CodeGenerateAgent) Stream(ctx context.Context, request *CodeGenerateTaskRequest, opts ...agent.AgentOption) (
	output *schema.StreamReader[string], err error) {
	return r.runnable.Stream(ctx, request, agent.GetComposeOptions(opts...)...)
}

// RemoveThinkBlocks 移除所有 <think>...</think> 标签及其内部内容。
// - 支持跨多行内容
// - 使用非贪婪匹配，避免误删多个 think 块之间的正常内容
// - 会顺带清理因移除标签产生的多余空行
func RemoveThinkBlocks(code string) string {
	if code == "" {
		return code
	}

	// (?s) 让 '.' 匹配包括换行在内的任意字符
	// 非贪婪匹配，确保只移除单个 <think>...</think> 区块
	re := regexp.MustCompile(`(?s)<think\b[^>]*>.*?</think\s*>`)
	out := re.ReplaceAllString(code, "")

	// 处理自闭合形式的标签，如 <think/> 或 <think ... />
	reSelfClosing := regexp.MustCompile(`(?i)<think\b[^>]*/\s*>`)
	out = reSelfClosing.ReplaceAllString(out, "")

	// 去掉常见的 Markdown 代码块包裹（按需可扩展更多语言标记）
	code = strings.TrimPrefix(code, "```python")
	code = strings.TrimPrefix(code, "```go")
	code = strings.TrimPrefix(code, "```")
	code = strings.TrimSuffix(code, "```")

	// 清理因删除 think 块导致的多余空行
	out = normalizeBlankLines(out)

	return strings.TrimSpace(out)
}

// normalizeBlankLines 将连续 3 行及以上的空行压缩为最多 1 个空行
// 以保持文本结构清晰、紧凑
func normalizeBlankLines(s string) string {
	re := regexp.MustCompile(`\n{3,}`)
	return re.ReplaceAllString(s, "\n\n")
}
