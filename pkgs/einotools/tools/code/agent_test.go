package tools

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/insmtx/corekg/pkgs/einotools/sandbox"
	"github.com/ygpkg/yg-go/logs"
)

func TestGenCode(t *testing.T) {
	ctx := t.Context()

	aiKey := ""

	if len(aiKey) == 0 {
		t.Skip("DS_KEY not set, skipping TestGenCode")
	}

	logs.InfoContextf(ctx, "start status: %s", "ok")

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  aiKey,
		Model:   "deepseek-v3",
		Timeout: 300 * time.Second,
		BaseURL: "https://api.example.com/v3/llm.chat",
	})
	if err != nil {
		logs.ErrorContextf(ctx, "创建OpenAiChatModel失败: %v", err)
	}

	defaultSandbox, err := sandbox.DefaultSandbox()
	if err != nil {
		t.Fatalf("创建DefaultSandbox失败: %v", err)
	}

	codeGenConfig := &CodeGenConfig{
		CodeModel: chatModel,
		Lang:      "python",
		Sandbox:   defaultSandbox,
	}

	codeAgent, err := NewGenerateCodeAgent(ctx, codeGenConfig)
	if err != nil {
		logs.ErrorContextf(ctx, "创建CodeGenerateAgent失败: %v", err)
	}

	t.Run("TestGenerateCode_CalcOddSumUnder100", func(t *testing.T) {
		runCodeGenTest(t, ctx, defaultSandbox, codeAgent, "计算100以内的奇数和")
	})

	t.Run("TestGenerateCode_CountNonEmptyCellsRowsCols", func(t *testing.T) {
		requirement := `计算表格中，总行数，总列数，以及所有非空单元格的个数。
		文件位置为: data/3466-split-sub.xlsx`
		runCodeGenTest(t, ctx, defaultSandbox, codeAgent, requirement)
	})

}

func runCodeGenTest(t *testing.T, ctx context.Context, sandbox sandbox.Sandbox, codeAgent *CodeGenerateAgent, requirement string) {
	t.Helper()

	handler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			fmt.Printf("=== 节点开始: %s ===\n", info.Name)
			fmt.Printf("输入: %+v\n", input)
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			fmt.Printf("=== 节点结束: %s ===\n", info.Name)
			fmt.Printf("输出: %+v\n", output)
			return ctx
		}).
		Build()

	code, err := codeAgent.Generate(ctx, &CodeGenerateTaskRequest{
		Task: requirement,
	}, agent.WithComposeOptions(compose.WithCallbacks(handler)))
	if err != nil {
		t.Fatalf("代码生成失败: %v", err)
	}

	if code == "" {
		t.Fatalf("生成的代码为空")
	}
	t.Logf("生成的代码:\n%s", code)

	res, err := sandbox.Exec(ctx, "python", code)
	if err != nil || res.ExitCode != 0 {
		t.Fatalf("代码执行失败, exit=%d, err=%v, stderr=%q",
			res.ExitCode, err, res.Stderr)
	}

	logs.InfoContextf(ctx, "执行结果: stdout=%q, stderr=%q, exit=%d",
		res.Stdout, res.Stderr, res.ExitCode)
}
