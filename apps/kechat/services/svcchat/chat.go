package svcchat

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/insmtx/corekg/apps/einonodes/qachatnodes"
	"github.com/insmtx/corekg/apps/kechat/chat/core"
	agentwrapper "github.com/insmtx/corekg/apps/kechat/chat/wrapper"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochat"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kechat/models/qachat"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

func ChatQuestionStream(ctx *gin.Context, req *dtochat.ChatQuestionStreamRequest) (res *dtochat.ChatQuestionStreamResponse, err error) {
	res = &dtochat.ChatQuestionStreamResponse{}
	// 问题检查和更新
	questionEntity, err := chatquestion.GetQuetionByID(ctx, req.Request.QuestionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to get question by chatSession, questionID: %s, err: %v", req.Request.QuestionID, err)
		return nil, err
	}
	if questionEntity.ID == "" {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] invalid question, questionID: %s", req.Request.QuestionID)
		res.Code = errcode.ErrCode_NotFound
		res.Message = "kechat_question_not_found"
		return res, nil
	}

	// 会话检查和更新
	questionEntity.Source.ReqID = runtime.RequestID(ctx)
	if err := chatquestion.UpdateQuestion(ctx, questionEntity); err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to update question, questionEntity: %s, err: %v", logs.JSON(questionEntity), err)
		return nil, err
	}

	sessionEntity, err := chat.NewChatSessionsDao().GetByID(ctx, questionEntity.Source.SessionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to get chatSession by id, sessionID:%d, err: %v", questionEntity.Source.SessionID, err)
		return nil, err
	}
	if sessionEntity.ID == 0 {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] invalid chatSession, sessionID: %d", questionEntity.Source.SessionID)
		res.Code = errcode.ErrCode_NotFound
		res.Message = "kechat_session_not_found"
		return res, nil
	}
	userQuestion := questionEntity.Source.Question

	chatModelEntity, err := chat.NewChatModelDao().GetByID(ctx, sessionEntity.ModelID)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to get chatModel by id, modelId: %d, err: %v", sessionEntity.ModelID, err)
		return nil, err
	}
	if chatModelEntity.ID == 0 {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] invalid chatModel, modelId: %d", sessionEntity.ModelID)
		res.Code = errcode.ErrCode_NotFound
		res.Message = "kechat_model_not_found"
		return res, nil
	}

	defer func() {
		logs.InfoContextf(ctx, "[ChatQuestionStream] questionEntity: %s", logs.JSON(questionEntity))
		subquestioon, _ := chatquestion.GetLLmSubQuestion(ctx, questionEntity.Source.Question, questionEntity.Source.Answer)
		questionEntity.Source.SubQuestion = subquestioon
		if err := chatquestion.UpdateQuestion(ctx, questionEntity); err != nil {
			logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to update question, questionEntity: %s, err: %v", logs.JSON(questionEntity), err)
		}
		// chatsession.UpdateSessionName(ctx, sessionEntity, questionEntity.Source.Question)
		chatsession.UpdateSessionNameWithLLM(ctx, sessionEntity, questionEntity.Source.Question, questionEntity.Source.Answer)
	}()

	state := qachatnodes.NewState(ctx)
	state.UserInput = userQuestion
	state.QuestionEntity = questionEntity
	state.SessionEntity = sessionEntity
	state.ModelEntity = chatModelEntity
	state.QuestionDbDatasetEntity = &chattype.ChatQuestionDbDataset{
		SessionID:    sessionEntity.ID,
		QuestionID:   questionEntity.ID,
		RequestID:    runtime.RequestID(ctx),
		DatabaseType: string(dbplugins.DatabaseTypeMySQL),
	}

	genFunc := func(c context.Context) *qachatnodes.State {
		return state
	}

	wrapper := qachat.NewChatWrapper(ctx, questionEntity, sessionEntity, chatModelEntity)
	var runable compose.Runnable[nodebase.RecordList, nodebase.RecordList]
	switch sessionEntity.BaseType {
	case chattype.ResourceQASessionBaseModel,
		chattype.ResourceQASessionBaseTypeGraphSearch,
		chattype.ResourceQASessionBaseTypeStandard,
		chattype.ResourceQASessionBaseTypeReactExcel:
		// err = wrapper.LLmChat(true)
		// err = wrapper.ForestChat()
		{
			wrapper := agentwrapper.NewChatWrapper(ctx, &core.ChatContext{
				Session:  sessionEntity,
				Question: questionEntity,
				Model:    chatModelEntity,
			})
			_, err = wrapper.Run(ctx)
		}
	case chattype.ResourceQASessionBaseAgent:
		err = wrapper.AgentChat(true)
	case chattype.ResourceQASessionBaseTypeDbMYSQL:
		err = wrapper.MysqlChat()
	case chattype.ResourceQASessionBaseGraphQA: // 图谱问答
		err = wrapper.ForestGraphChat()
	case chattype.ResourceQASessionBaseTypeExcel:
		runable, err = qachatnodes.ExcelChatBuilder[nodebase.RecordList, nodebase.RecordList](ctx, genFunc)
		defer func() {
			if err := chatquestion.NewChatQuestionDbDatasetDao().Insert(ctx, state.QuestionDbDatasetEntity); err != nil {
				logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to insert chat question db dataset, err: %v", err)
			}
		}()
		// runable, err = qachatnodes.ExcelChatBuilder[nodebase.RecordList, nodebase.RecordList](ctx, genFunc)
	case chattype.ResourceQASessionBaseExternalData:
		runable, err = qachatnodes.ExternalChatBuilder[nodebase.RecordList, nodebase.RecordList](ctx, genFunc)
	default:
		logs.ErrorContextf(ctx, "[ChatQuestionStream] invalid session base type, baseType: %s", sessionEntity.BaseType)
		res.Code = errcode.ErrCode_NotFound
		res.Message = "kechat_session_base_type_not_support"
		return res, nil
	}

	if err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to build chat, err: %v", err)
		answer := questionEntity.Source.Answer
		if answer == "" {
			answer = i18n.T(runtime.GetLanguage(ctx), "kechat_question_answer_error")
			questionEntity.Source.Answer = answer
		}
		llmchat.WriteContent(ctx, questionEntity.Source.ReqID, answer)
		questionEntity.Source.Status = chattype.QuestionStatusError
	}
	if runable == nil {
		return res, nil
	}
	_, err = runable.Invoke(ctx, state.Records)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to invoke chat, err: %v", err)
		answer := questionEntity.Source.Answer
		if answer == "" {
			answer = i18n.T(runtime.GetLanguage(ctx), "kechat_question_answer_error")
			questionEntity.Source.Answer = answer
		}
		llmchat.WriteContent(ctx, questionEntity.Source.ReqID, answer)
		questionEntity.Source.Status = chattype.QuestionStatusError
	}

	return res, nil
}
