package keqa

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// 用户意图识别
func IntentionRecognition(ctx *gin.Context, question *chattype.ChatQuestion) (string, error) {
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentInetentionRecognition),
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: question.Source.Question},
			},
		},
	}
	w, err := chatclient.NewInternalChat(ctx, question.Source.ReqID, "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return "", err
	}
	return res.Content, nil
}
