package llmchat

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

func TestInternalChatResponse(t *testing.T) {
	// testutils.Initialize(testutils.AppNameKechat)
	// defer testutils.Close()
	ctx := &gin.Context{}
	model, err := chatmodel.GetModelByID(ctx, 1)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateChatModel GetModelByID failed ,err %s", err)
		return
	}
	request := ChatReqBody{
		Stream:        true,
		StreamOptions: NewStreamOptions(),
	}
	messages := []*Message{
		{Role: MessageRoleSystem, Content: "你的名字是言小古"},
		{Role: MessageRoleUser, Content: "你是谁"},
	}
	request.Messages = messages
	wrapper := NewLLmChatWrapper(ctx, &request, model)
	res, err := wrapper.InternalChatResponse(func(chunk *chattype.ChatStreamResponseBody) error {
		for _, v := range chunk.Choices {
			println(v.Delta.ReasoningContent)
			println(v.Delta.Content)
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "InternalChatResponse failed ,err %s", err)
		return
	}
	logs.InfoContextf(ctx, "InternalChatResponse res %s", res)
}
