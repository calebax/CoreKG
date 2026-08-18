package agent

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/ygpkg/yg-go/logs"
)

const (
	hotWordPrompt = `你是一个专业的【用户问题热词分析与总结助手】。

## 任务目标
基于用户的历史问题文本，从中抽取能够代表用户关注重点的【热词】。

## 输入内容
=========
{query}
=========

## 输出要求
1. 从用户历史问题中总结并抽取【若干热词】，**不超过 15 个**  
2. 热词需体现用户的**高频关注主题、技术方向或业务场景**
3. 按【热度从高到低】排序（出现频率 + 关注重要性综合判断）
4. 热词之间使用【英文逗号 ","】分隔
5. **仅输出热词本身，不要输出任何解释、序号、换行或多余文字**

## 质量规范
- 热词应为 **名词或名词短语**
- 避免过于宽泛或无信息量的词（如：问题、方法、系统）
- 优先选择：技术名词、产品名、领域关键词、核心概念
- 允许中英文混合（如技术名词为英文）

## 输出示例（仅示例格式）
网络建设流程,网络架构规划,防火墙,CPU,内存,机房灰尘标准,H3C SecPath F100,POE交换机,网络设备,知识库
`
)

// NewHotWordAgent 生成热词
func NewHotWordAgent(ctx context.Context, chatModel model.ToolCallingChatModel) adk.Agent {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "GenerateHotWords",
		Description: "根据用户历史问题生成热词",
		Model:       chatModel,
		Instruction: `你是一个专业的热词总结生成助手，在用户的历史问题中总结抽取热词`,
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			logs.InfoContextf(ctx, "NewHotWordAgent GenModelInput : %s", logs.JSON(input))
			ct := prompt.FromMessages(schema.FString,
				schema.SystemMessage(instruction),
				schema.UserMessage(hotWordPrompt),
			)
			msgs, err := ct.Format(ctx, map[string]any{
				"query": input.Messages[0].Content,
			})
			if err != nil {
				logs.ErrorContextf(ctx, "NewHotWordAgent GenModelInput ct.Format error: %v", err)
				return nil, err
			}

			return msgs, nil
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewHotWordAgent NewChatModelAgent error: %v", err)
		return nil
	}
	return a
}

// ExecuteHotWordAgent 生成热词
func ExecuteHotWordAgent(ctx context.Context, chatModel model.ToolCallingChatModel, questions []string) ([]string, error) {
	ag := NewHotWordAgent(ctx, chatModel)

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: ag,
	})
	iter := runner.Query(ctx, "["+strings.Join(questions, ",")+"]")

	var results []string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			logs.ErrorContextf(ctx, "ExecuteCatalogueAgent iter.Next error: %v", event.Err)
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if m := event.Output.MessageOutput.Message; m != nil {
				if len(m.Content) > 0 {
					logs.InfoContextf(ctx, "ExecuteCatalogueAgent answer: %s", m.Content)
					results = append(results, strings.Split(m.Content, ",")...)
				}
			}
		}
	}

	return results, nil
}
