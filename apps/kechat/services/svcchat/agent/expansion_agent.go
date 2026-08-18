package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/ygpkg/yg-go/logs"
)

const (
	expansionPrompt = `你是一个专业的用户问题语义扩充专家。你的任务是基于上下文信息，将用户的简短问题扩充为完整、明确的问题描述。

## 核心原则
1. 保持用户原始意图不变
2. 补充必要的上下文信息
3. 消除指代不明的代词（如"这个"、"它"、"那个"等）
4. 使问题可以独立理解，无需依赖历史对话

## 输入信息

### 相关文章问答记录
{desc}

### 历史问答记录
{history}

### 用户当前问题
{question}

## 输出要求
- 只输出扩充后的问题，不要任何额外说明或解释
- 确保问题完整、具体、无歧义
- 保留用户问题的原有语气和风格
- 如果原问题已经足够明确，可以直接返回原问题

示例：
- 原问题："这个怎么用？" → 扩充后："[根据上文具体内容]怎么使用？"
- 原问题："还有其他方法吗？" → 扩充后："除了[上文提到的方法]，还有其他[具体目标]的方法吗？"

`
)

// NewExpansionAgent 语义扩写
func NewExpansionAgent(ctx context.Context, chatModel model.ToolCallingChatModel, history, desc string) adk.Agent {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "GenerateExpansionQuestion",
		Description: "扩写用户问题，补充上下文信息",
		Model:       chatModel,
		Instruction: `你是一个专业的语义扩写助手，能够基于上下文信息将用户的简短问题扩充为完整明确的问题描述`,
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			logs.InfoContextf(ctx, "NewExpansionAgent GenModelInput : %s", logs.JSON(input))
			ct := prompt.FromMessages(schema.FString,
				schema.SystemMessage(instruction),
				schema.UserMessage(expansionPrompt),
			)
			msgs, err := ct.Format(ctx, map[string]any{
				"question": input.Messages[0].Content,
				"history":  history,
				"desc":     desc,
			})
			if err != nil {
				logs.ErrorContextf(ctx, "NewExpansionAgent GenModelInput ct.Format error: %v", err)
				return nil, err
			}

			return msgs, nil
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewExpansionAgent NewChatModelAgent error: %v", err)
		return nil
	}
	return a
}

// ExecuteExpansionAgent 生成扩写问题
func ExecuteExpansionAgent(ctx context.Context, chatModel model.ToolCallingChatModel, question, history, desc string) (string, error) {
	ag := NewExpansionAgent(ctx, chatModel, history, desc)

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: ag,
	})
	iter := runner.Query(ctx, question)

	var results string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			logs.ErrorContextf(ctx, "ExecuteExpansionAgent iter.Next error: %v", event.Err)
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if m := event.Output.MessageOutput.Message; m != nil {
				if len(m.Content) > 0 {
					logs.InfoContextf(ctx, "ExecuteExpansionAgent answer: %s", m.Content)
					results = m.Content
				}
			}
		}
	}

	return results, nil
}
