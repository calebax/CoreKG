package example5graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/einonodes/einodemo/utils"
	"github.com/stretchr/testify/assert"
)

func TestQAAgent(t *testing.T) {
	ctx := context.Background()

	chatModel, err := utils.NewOpenAiChatModel(ctx, utils.OpenAiDefaultConfig)
	assert.NoError(t, err)
	// 创建QA代理
	tAgent, err := NewQAAgent(ctx, &Config{
		ChatModel: chatModel,
	})
	assert.NoError(t, err)

	LoggerCallbackOption := agent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{}))

	// 测试场景1：一般性问题，应该直接由LLM回答
	t.Run("DirectLLM", func(t *testing.T) {
		msg := []*schema.Message{
			schema.UserMessage("什么是人工智能?"),
		}
		resp, err := tAgent.Generate(ctx, msg, LoggerCallbackOption)
		assert.NoError(t, err)

		t.Log(resp.Content)
	})

	t.Run("SubRagGraph", func(t *testing.T) {
		msg := []*schema.Message{
			schema.UserMessage("Eino如何实现RAG模式?"),
		}

		resp, err := tAgent.Generate(ctx, msg, LoggerCallbackOption)
		assert.NoError(t, err)

		t.Log(resp.Content)
	})

}

type LoggerCallback struct {
	callbacks.HandlerBuilder // 可以用 callbacks.HandlerBuilder 来辅助实现 callback
}

func (cb *LoggerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	fmt.Println("==================")
	inputStr, _ := json.Marshal(input)
	fmt.Printf("[%s OnStart]\n %s\n", info.Name, string(inputStr))
	return ctx
}

func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	fmt.Printf("=========[%s OnEnd]=========\n", info.Name)
	outputStr, _ := json.Marshal(output)
	fmt.Println(string(outputStr))
	return ctx
}

func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	fmt.Println("=========[OnError]=========")
	fmt.Println(err)
	return ctx
}

func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	var graphInfoName = react.GraphName

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[OnEndStream] panic err:", err)
			}
		}()

		defer output.Close() // remember to close the stream in defer

		fmt.Println("=========[OnEndStream]=========")
		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				// finish
				break
			}
			if err != nil {
				fmt.Printf("internal error: %s\n", err)
				return
			}

			s, err := json.Marshal(frame)
			if err != nil {
				fmt.Printf("internal error: %s\n", err)
				return
			}

			if info.Name == graphInfoName { // 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
				fmt.Printf("%s: %s\n", info.Name, string(s))
			}
		}

	}()
	return ctx
}

func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}
