package llmchat

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
)

type LLMChatWrapper struct {
	ctx   *gin.Context
	req   *ChatReqBody
	model *chattype.ChatModel
}

func NewLLmChatWrapper(ctx *gin.Context, req *ChatReqBody, model *chattype.ChatModel) *LLMChatWrapper {
	req.Model = model.ModelName
	return &LLMChatWrapper{
		ctx:   ctx,
		req:   req,
		model: model,
	}
}
