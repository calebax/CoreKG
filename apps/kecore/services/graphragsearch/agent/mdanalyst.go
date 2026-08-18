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
	// ConciseEfficientPrompt 简洁高效提示词
	ConciseEfficientPrompt = `你是一个专业的智能信息检索助手。  
你如同一位高级秘书，专门根据检索到的文档信息回答用户问题，兼具专业性、准确性与表达丰富性。

## 角色定位
- 你**只能**依据提供的检索信息回答用户问题。  
- 不得使用或推理任何外部知识、行业经验或常识。  
- 所有回答必须基于事实，并可追溯到原始文件。

## 数据说明
以下是系统检索到的文档信息（每一项代表一个文件）：
------BEGIN------
informations: {informations}
------END------

每个文件对象包含：
- content：文档正文内容（可能包含 HTML、图片描述、链接等）。
- analysis：文档摘要内容。


## 回答要求

1. **信息来源限制**
   - 只能引用 informations 中提供的内容；
   - 禁止编造、猜测、推理或使用任何外部知识。

2. **图片引用**
   - 若 informations 中存在图片地址（以 .jpg, .png, .jpeg 等结尾），你可以在回答中插入 Markdown 图片：
     ![图片说明](图片URL)
   - 不得虚构、生成或外链任何不存在于 informations 中的图片。

3. **输出格式**
   - 使用 **Markdown** 结构化输出；
   - 严格遵循信息来源引用格式。

4. **无法回答时**
   - 若检索信息不足，请直接回复：
     > 现有资料不足，无法回答。

5. **风格控制：极简主义**
   - **直击要点**：**禁止**包含“你好”、“根据检索结果”、“综上所述”等任何客套话或过渡语。直接给出结论。
   - **列表优先**：所有可以列举的内容，必须使用无序列表（Bullet points）呈现。
   - **文字精炼**：能用短语不用句子，能用一句话说清楚的绝不写两句。

## 用户问题
{query}
---
请严格依据以上规则与提供的信息生成你的最终回答。
`
)

var (
	fileIDS = []uint{977, 275, 265, 888, 940, 941, 943, 890, 893, 885, 889, 968}
)

type Content struct {
	Content  string `json:"content"`
	Analysis string `json:"analysis"`
}

func NewAnalystAgent(ctx context.Context, chatModel model.ToolCallingChatModel, files []*FileData) (adk.Agent, error) {
	con := []*Content{}
	// fMap := map[uint]search.DirectoryInfo{}
	for _, v := range files {

		con = append(con, &Content{
			Content:  v.Content,
			Analysis: v.Analysis,
		})
	}

	constr := truncateStringByRune(logs.JSON(con), 230000)
	ag, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SearchAnalyst",
		Description: "搜索到的内容进行分析",
		Model:       chatModel,
		Instruction: `你是一个专业的内容分析助手，能够根据用户提供的内容，进行深入分析并提供有价值的见解。`,
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			logs.InfoContextf(ctx, "NewCatalpgueAgent GenModelInput : %s", logs.JSON(input))
			ct := prompt.FromMessages(schema.FString,
				schema.SystemMessage(instruction),
				schema.UserMessage(ConciseEfficientPrompt),
			)
			msgs, err := ct.Format(ctx, map[string]any{
				"informations": constr,
				"query":        input.Messages[0].Content,
			})
			if err != nil {
				logs.ErrorContextf(ctx, "NewCatalpgueAgent GenModelInput ct.Format error: %v", err)
				return nil, err
			}

			// logs.InfoContextf(ctx, "NewAnalystAgent GenModelInput msgs : %s", logs.JSON(msgs))

			return msgs, nil
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueAgent NewChatModelAgent error: %v", err)
		return nil, err
	}
	return ag, nil
}

func truncateStringByRune(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
