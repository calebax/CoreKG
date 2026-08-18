package example4agentgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	callbackHelper "github.com/cloudwego/eino/utils/callbacks"
	"github.com/insmtx/corekg/apps/einonodes/einodemo/utils"
)

func TestChatTemplate(t *testing.T) {
	ctx := context.Background()

	tools, err := GetTools(ctx)
	if err != nil {
		t.Fatal(err)
	}

	plannerModel, err := utils.NewOpenAiChatModel(ctx, utils.DsDefaultConfigByYYGU)
	if err != nil {
		t.Fatal(err)
	}
	executorModel, err := utils.NewOpenAiChatModel(ctx, utils.OpenAiDefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	reviserModel, err := utils.NewOpenAiChatModel(ctx, utils.OpenAiDefaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	config := &Config{
		PlannerModel:  plannerModel,
		ExecutorModel: executorModel,
		ReviserModel:  reviserModel,
		ToolsConfig:   compose.ToolsNodeConfig{Tools: tools},
	}

	t.Run("ChatMultiAgent", func(t *testing.T) {
		useAgent, err := NewMultiAgent(ctx, config)
		if err != nil {
			t.Fatal(err)
		}

		// 创建 callback handler
		handler := &callbackHelper.ModelCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
				jsonStr, err := json.Marshal(input.Messages)
				if err != nil {
					fmt.Printf("marshal input messages failed: %v\n", err)
					return ctx
				}
				fmt.Printf("step %s 开始生成，输入消息:\n %s\n\n", info.Name, string(jsonStr))
				return ctx
			},
			OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
				jsonStr, err := json.Marshal(output.Message)
				if err != nil {
					fmt.Printf("marshal output failed: %v\n", err)
					return ctx
				}
				fmt.Printf("step %s 生成完成，输出消息:\n %s\n\n", info.Name, string(jsonStr))
				return ctx
			},
			OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
				fmt.Println("开始接收流式输出")
				defer output.Close()
				return ctx
			},
		}
		// 使用 callback handler
		helper := callbackHelper.NewHandlerHelper().
			ChatModel(handler).
			Handler()

		resp, err := useAgent.Generate(ctx, []*schema.Message{
			schema.UserMessage("帮我计算在今天，一万元能买多少克黄金。"),
		}, agent.WithComposeOptions(compose.WithCallbacks(helper)))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(resp.Content)
	})

}

func GetTools(ctx context.Context) (tools []tool.BaseTool, err error) {
	// 创建 duckduckgo Search 工具.  官方封装的工具
	searchTool, err := duckduckgo.NewTextSearchTool(ctx, &duckduckgo.Config{
		ToolName: "web_search",
	})
	if err != nil {
		return nil, err
	}
	tools = append(tools, searchTool)

	return tools, nil
}
