package chat

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/chat/core"
	"github.com/insmtx/corekg/apps/kechat/chat/modes"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

type ChatWrapper struct {
	ctx     *gin.Context
	context *core.ChatContext
}

func NewChatWrapper(
	ctx *gin.Context,
	context *core.ChatContext,
) *ChatWrapper {
	return &ChatWrapper{
		ctx:     ctx,
		context: context,
	}
}

func (w *ChatWrapper) Run(ctx context.Context) (*core.ChatResult, error) {
	var mode core.ChatMode

	switch w.context.Session.BaseType {
	case chattype.ResourceQASessionBaseTypeStandard:
		mode = modes.NewForestChatMode(w.ctx)
	case chattype.ResourceQASessionBaseTypeGraphSearch:
		mode = modes.NewGraphSearchChatMode(w.ctx)
	case chattype.ResourceQASessionBaseModel:
		mode = modes.NewDirectModelChatMode(w.ctx)
	case chattype.ResourceQASessionBaseTypeReactExcel:
		mode = modes.NewExcelChatMode(w.ctx)
	case chattype.ResourceQASessionBaseTypeForestAgent:
		mode = modes.NewForestAgentChatMode(w.ctx)
	default:
		logs.ErrorContextf(ctx, "[ChatQuestionStream] invalid session base type, baseType: %s", w.context.Session.BaseType)
	}

	// 执行聊天模式
	chatResult, err := mode.Run(ctx, w.context)

	// 统一更新 question.Source 数据
	w.updateQuestionSource(chatResult, err)

	return chatResult, err
}

func (w *ChatWrapper) updateQuestionSource(chatResult *core.ChatResult, err error) {
	if err != nil {
		w.context.Question.Source.Status = chattype.QuestionStatusError
		logs.ErrorContextf(w.ctx, "[ChatWrapper run] error: %s", err.Error())
		return
	}

	if chatResult == nil {
		logs.ErrorContextf(w.ctx, "[ChatWrapper updateQuestionSource] chatResult is nil")
		return
	}

	w.context.Question.Source.Answer = chatResult.Answer
	w.context.Question.Source.Reasoning = chatResult.Reasoning
	w.context.Question.Source.ReasoningSeconds = chatResult.Performance.ReasoningSeconds
	w.context.Question.Source.CostSeconds = chatResult.Performance.CostSeconds
	w.context.Question.Source.OutToken = chatResult.Usage.CompletionTokens
	w.context.Question.Source.CacheHitToken = chatResult.Usage.CacheHitTokens
	w.context.Question.Source.CacheMissToken = chatResult.Usage.CacheMissTokens
	w.context.Question.Source.TotalTokens = chatResult.Usage.TotalTokens
	w.context.Question.Source.Status = chatResult.Status

	if chatResult.Meta != nil {
		if svc := chatResult.Meta.AgentService(); svc != nil {
			w.context.Question.Source.ReactAgentService = svc
		}
		if refs := chatResult.Meta.QueryReferences(); refs != nil {
			w.context.Question.Source.QueryReferenceList = refs
		}
		if refs := chatResult.Meta.ChatReferences(); refs != nil {
			w.context.Question.Source.ChatReferenceList = refs
		}
	}
}
