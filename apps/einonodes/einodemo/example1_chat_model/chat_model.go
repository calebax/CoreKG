package example1chatmodel

/*
eino ChatModel 简单示例
参考:
https://www.cloudwego.io/zh/docs/eino/core_modules/components/chat_model_guide/
https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/chat_model/
*/

import (
	"context"
	"io"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func ChatByStream(ctx context.Context, chatModel model.BaseChatModel, messages []*schema.Message, callback func(string)) error {
	// 获取流式回复
	reader, err := chatModel.Stream(ctx, messages)
	if err != nil {
		panic(err)
	}
	// 注意要关闭
	defer reader.Close()

	// 处理流式内容
	for {
		chunk, err := reader.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// 回调返回内容
		callback(chunk.Content)
	}
	return nil
}

func ChatByGenerate(ctx context.Context, chatModel model.BaseChatModel, messages []*schema.Message) (string, string, error) {
	// 生成回复
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", "", err
	}

	//  Token 使用情况
	if usage := response.ResponseMeta.Usage; usage != nil {
		println("提示 Tokens:", usage.PromptTokens)
		println("生成 Tokens:", usage.CompletionTokens)
		println("总 Tokens:", usage.TotalTokens)
	}

	return response.ReasoningContent, response.Content, nil
}
