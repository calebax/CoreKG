package chatsession

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kechat/chat/modelhelper"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/ygpkg/yg-go/logs"
)

// GetLLmSessionName 获取模型总结的会话名称
func GetLLmSessionName(ctx context.Context, question, answer string) (string, error) {
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
		return "", err
	}
	// TODO: 如何配置模型 core_setting？
	cmodel, err := chatmodel.GetModelByID(ctx, 1)
	if err != nil {
		logs.ErrorContextf(ctx, "GetLLmSessionName GetModelByID err:%v", err)
		return "", err
	}
	model, err := modelhelper.NewToolCallingChatModel(ctx, cmodel, modelhelper.ToolCallingChatModelOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "GetLLmSessionName NewOpenAiChatModel err:%v", err)
		return "", err
	}
	response, err := model.Generate(ctx, messages)
	if err != nil {
		logs.ErrorContextf(ctx, "request llm err: %v", err)
		return "", err
	}
	return response.Content, nil
}

const (
	sessionNamePrompt = `    你是一个专业的会话标题生成助手，你的任务是为用户提问创建简洁、精准且具描述性的标题。
    ## 格式要求
    - 标题长度必须在15个字以内
    - 标题应准确反映用户问题的核心主题
    - 使用名词短语结构，避免使用问句
    - 保持简洁明了，删除非必要词语
    - 不要使用"关于"、"如何"等冗余词语开头
    - 直接输出标题文本，不要有任何前缀、解释或标点符号

    ## Few-shot示例

    用户问题: 如何提高英语口语水平？
    标题: 英语口语提升

    用户问题: 最近上海有什么好玩的展览活动？
    标题: 上海展览推荐

    用户问题: 苹果手机电池不耐用怎么解决？
    标题: 苹果电池优化

    ## 用户的问题是：
	{question}
	## 模型回答是：
	{answer}
`
)
