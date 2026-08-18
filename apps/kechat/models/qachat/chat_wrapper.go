package qachat

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/settings"
)

type ChatWapper struct {
	ctx      *gin.Context
	model    *chattype.ChatModel
	question *chattype.ChatQuestion
	session  *chattype.ChatSession
}

func NewChatWrapper(ctx *gin.Context, question *chattype.ChatQuestion, session *chattype.ChatSession, model *chattype.ChatModel) *ChatWapper {
	return &ChatWapper{
		ctx:      ctx,
		model:    model,
		question: question,
		session:  session,
	}
}

// ChatAPIWrapper API 调用器
type ChatAPIWrapper struct {
	ctx       *gin.Context
	model     *chattype.ChatModel
	question  *chattype.ChatQuestion
	agentInfo *chatagent.AgentWithVersion
}

func NewChatAPIWrapper(ctx *gin.Context, question *chattype.ChatQuestion, model *chattype.ChatModel, agentInfo *chatagent.AgentWithVersion) *ChatAPIWrapper {
	return &ChatAPIWrapper{
		ctx:       ctx,
		model:     model,
		question:  question,
		agentInfo: agentInfo,
	}
}

type InternalChat struct {
	ctx    *gin.Context
	step   int
	reqid  string
	apiKey string
	req    *chattype.ChatRequestBody
}

func NewInternalChat(ctx *gin.Context, reqid, apiKey string, step int, req *chattype.ChatRequestBody) (*InternalChat, error) {
	if apiKey == "" {
		sysAPIKey, err := settings.GetText(global.SettingGroupKnowledge, global.SettingKeySystemLlmAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get system api key: %v", err)
		}
		apiKey = sysAPIKey
	}
	return &InternalChat{
		ctx:    ctx,
		step:   step,
		reqid:  reqid,
		apiKey: apiKey,
		req:    req,
	}, nil
}
