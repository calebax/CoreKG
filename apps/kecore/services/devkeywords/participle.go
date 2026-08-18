package devkeywords

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/ygpkg/yg-go/logs"
)

const (
	participlePrompt = `你是一个**专业的关键词提取助手**，任务是从用户输入的一句话中**提取所有具有明确语义的名词关键词**。

#### 提取规则

1. **仅提取名词或名词性短语**
2. **不包含动词、形容词、副词、介词、连词、语气词**
3. **不对原文本做任何改写、缩写或补全**
4. **严格保持原文中的字符、顺序和大小写**
5. **不得捏造、合并或拆分任何名词**
6. **不遗漏任何一个具有实际语义的名词**

#### 输出格式

* 仅输出结果，不要任何解释或说明
* 使用英文逗号 , 分隔
* 不要添加空格或其他字符

#### 示例
输入：后倒车雷达维修步骤
输出：后倒车雷达,维修步骤


# 用户问题如下：
{query}

`
)

// NewParticipleAgent 提取关键词
func NewParticipleAgent(ctx context.Context, chatModel model.ToolCallingChatModel) adk.Agent {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TreeNodeSelector",
		Description: "在用户输入中提取关键词",
		Model:       chatModel,
		Instruction: `你是一个**专业的关键词提取助手**，任务是从用户输入的一句话中**提取所有具有明确语义的名词关键词**。`,
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			logs.InfoContextf(ctx, "NewParticipleAgent GenModelInput : %s", logs.JSON(input))
			ct := prompt.FromMessages(schema.FString,
				schema.SystemMessage(instruction),
				schema.UserMessage(participlePrompt),
			)
			msgs, err := ct.Format(ctx, map[string]any{
				"query": input.Messages[0].Content,
			})
			if err != nil {
				logs.ErrorContextf(ctx, "NewParticipleAgent GenModelInput ct.Format error: %v", err)
				return nil, err
			}

			return msgs, nil
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueAgent NewChatModelAgent error: %v", err)
		return nil
	}
	return a
}
