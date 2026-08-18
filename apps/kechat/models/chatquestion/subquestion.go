package chatquestion

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kechat/chat/modelhelper"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/ygpkg/yg-go/logs"
)

// GetLLmSubQuestion 获取模型生成的子问题
func GetLLmSubQuestion(ctx context.Context, question, answer string) ([]string, error) {
	template := prompt.FromMessages(schema.FString,
		&schema.Message{
			Role:    schema.User,
			Content: sessionNamePrompt,
		},
	)
	params := map[string]any{
		"question": question,
		"answer":   answer,
	}
	messages, err := template.Format(ctx, params)
	if err != nil {
		logs.ErrorContextf(ctx, "GetLLmSessionName format message err:%v", err)
		return nil, err
	}
	// TODO: 如何配置模型 core_setting？
	cmodel, err := chatmodel.GetModelByID(ctx, 1)
	if err != nil {
		logs.ErrorContextf(ctx, "GetLLmSessionName GetModelByID err:%v", err)
		return nil, err
	}
	model, err := modelhelper.NewToolCallingChatModel(ctx, cmodel, modelhelper.ToolCallingChatModelOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "GetLLmSessionName NewOpenAiChatModel err:%v", err)
		return nil, err
	}
	response, err := model.Generate(ctx, messages)
	if err != nil {
		logs.ErrorContextf(ctx, "request llm err: %v", err)
		return nil, err
	}
	subquestion := strings.Split(response.Content, ",")
	return subquestion, nil
}

const (
	sessionNamePrompt = `    你是一个专业的子问题生成助手，你的任务是根据用户本轮问答创建用户下一轮可能会问答的问题。
	##默认工作语言： **中文**
	- 如果明确提供，则使用用户指定的语言作为工作语言
	- 所有思维和响应必须使用工作语言

    ## 格式要求
    - 生成的问题不能超过三个
    - 问题之间仅使用英文逗号,分隔
    - 直接输出标题文本，不要有任何前缀、解释或标点符号
	
    ## Few-shot示例

    用户问题: 知识库讲了什么？？
    标题: 如何查找特定产品的规格文档,故障处理的具体步骤是什么,如何获取产品的认证证书

    ## 用户的问题是：
	{question}
	## 模型回答是：
	{answer}
`
)
