package example3agentchain

import (
	"context"
	"fmt"
	"log"

	duckduckgo "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	callbackHelper "github.com/cloudwego/eino/utils/callbacks"
	localtool "github.com/insmtx/corekg/apps/einonodes/einodemo/example3_agent_chain/tool"
	"github.com/insmtx/corekg/apps/einonodes/einodemo/utils"
)

/*
eino Agent 简单示例。用 Chain 构建。
参考:
https://www.cloudwego.io/zh/docs/eino/quick_start/agent_llm_with_tools/

核心部分：
ChatModel 和 Tool
*/

type Config struct {
	ChatModel model.ToolCallingChatModel
}

func SimpleAgent(ctx context.Context, cfg *Config, message string) {
	//绑定工具
	todoTools, err := GetTools(ctx)
	if err != nil {
		panic(err)
	}
	// 获取工具信息并绑定到 ChatModel
	toolInfos, err := genToolInfos(ctx, todoTools)
	if err != nil {
		panic(err)
	}
	toolChatModel, err := cfg.ChatModel.WithTools(toolInfos)
	if err != nil {
		panic(err)
	}

	// 创建 callback handler
	handler := &callbackHelper.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			fmt.Printf("开始执行工具，参数: %s\n", input.ArgumentsInJSON)
			fmt.Println("------------------------------------")
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			fmt.Printf("工具执行完成，结果: %s\n", output.Response)
			fmt.Println("------------------------------------")
			return ctx
		},
	}
	// 使用 callback handler
	helper := callbackHelper.NewHandlerHelper().
		Tool(handler).
		Handler()

	// 创建 tools 节点
	todoToolsNode, err := compose.NewToolNode(context.Background(), &compose.ToolsNodeConfig{
		Tools: todoTools,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 构建完整的处理链
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.
		AppendChatModel(toolChatModel, compose.WithNodeName("chat_model")).
		AppendToolsNode(todoToolsNode, compose.WithNodeName("tools"))

	// 编译并运行 chain
	agent, err := chain.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 运行agent
	resp, err := agent.Invoke(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: message,
		},
	}, compose.WithCallbacks(helper))
	if err != nil {
		log.Fatal(err)
	}

	// 输出结果
	for _, msg := range resp {
		fmt.Println(msg.Content)
	}
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

	// 创建 summary tool
	summaryChatModel, err := utils.NewOpenAiChatModel(ctx, utils.DsDefaultConfigByYYGU)
	if err != nil {
		return nil, err
	}
	summaryTool, err := localtool.NewTool(ctx, &localtool.Config{
		ChatModel: summaryChatModel,
	})
	if err != nil {
		return nil, err
	}
	tools = append(tools, summaryTool)

	return tools, nil
}

// 把可执行的 Tool 转化为大模型可用的 Tool 信息
func genToolInfos(ctx context.Context, tools []tool.BaseTool) ([]*schema.ToolInfo, error) {
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
