package example2chattemplate

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/einonodes/einodemo/utils"
)

/*
eino ChatTemplate 简单示例
参考:
https://www.cloudwego.io/zh/docs/eino/core_modules/components/chat_template_guide/
*/

func ChatTemplate(ctx context.Context) {
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.MessagesPlaceholder("history_key", false),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮帮我，史瓦罗先生，{task}",
		},
	)
	params := map[string]any{
		"role": "机器人史瓦罗先生",
		"task": "写一首诗",
		"history_key": []*schema.Message{
			{Role: schema.User, Content: "告诉我油画是什么?"},
			{Role: schema.Assistant, Content: "油画是xxx"},
		},
	}
	messages, err := template.Format(ctx, params)
	if err != nil {
		panic(err)
	}

	model, err := utils.NewOpenAiChatModel(ctx, utils.DsDefaultConfigByYYGU)
	if err != nil {
		panic(err)
	}

	response, err := model.Generate(ctx, messages)
	if err != nil {
		panic(err)
	}
	println(response.Content)

}
