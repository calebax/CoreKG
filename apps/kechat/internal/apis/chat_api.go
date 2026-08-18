package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/einonodes/qachatnodes"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochat"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kechat/models/qachat"
	"github.com/insmtx/corekg/apps/kechat/services/svcchat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

// SubmitChatQuestionStream Chat聊天问题流
// @Tags Chat
// @Summary 聊天问题流
// @Description 聊天问题流
// @Router /chat.SubmitChatQuestionStream [post]
// @Param request body SubmitChatQuestionStreamRequest true "request"
// @Success 200 {object} SubmitChatQuestionStreamResponse
func SubmitChatQuestionStream(ctx *gin.Context, req *SubmitChatQuestionStreamRequest, resp *SubmitChatQuestionStreamResponse) {
	question, err := chatquestion.GetQuetionByID(ctx, req.Request.QuestionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] [SubmitChatQuestionStream] Failed to get question by chatSession key: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_question_failed" // 获取问题失败
		return
	}
	// 只能修改否则停止不了
	question.Source.ReqID = runtime.RequestID(ctx)
	err = chatquestion.UpdateQuestion(ctx, question)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] [SubmitChatQuestionStream] Failed to update question: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_question_failed" // 修改问题失败
		return
	}
	// 获取会话
	session, err := chatsession.GetChatSession(ctx, question.Source.Uin, question.Source.SessionID)
	if err != nil {
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream get chat chatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed" // 获取会话记录失败
		return
	}
	// chatsession.UpdateSessionName(ctx, session, question.Source.Question)
	// 获取模型
	model, err := chatmodel.GetModelByID(ctx, session.ModelID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed" // 查询模型失败
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream GetModelByID failed ,err %s", err)
		return
	}

	wrapper := qachat.NewChatWrapper(ctx, question, session, model)
	switch session.BaseType {
	case chattype.ResourceQASessionBaseModel:
		err = wrapper.LLmChat(true)
	case chattype.ResourceQASessionBaseAgent:
		err = wrapper.AgentChat(true)
	case chattype.ResourceQASessionBaseTypeStandard:
		err = wrapper.ForestChat()
	case chattype.ResourceQASessionBaseTypeExcel:
		err = wrapper.ExcelChat()
	case chattype.ResourceQASessionBaseTypeDbMYSQL:
		err = wrapper.MysqlChat()
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_unsupported_chat_type" // 不支持的聊天类型
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream unsupported resource type: %s", session.ResourceType)
		return
	}

	// 保存聊天记录
	if err != nil {
		answer := question.Source.Answer
		if answer == "" {
			answer = i18n.T(runtime.GetLanguage(ctx), "kechat_chat_failed_retry") // 问答失败，请稍后重试
			question.Source.Answer = answer
		}
		llmchat.WriteContent(ctx, question.Source.ReqID, answer)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed" // 聊天失败
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream LLmChat failed ,err %s", err)
		question.Source.Status = chattype.QuestionStatusError
	}
	defer chatsession.UpdateSessionNameWithLLM(ctx, session, question.Source.Question, question.Source.Answer)
	subquestioon, _ := chatquestion.GetLLmSubQuestion(ctx, question.Source.Question, question.Source.Answer)
	question.Source.SubQuestion = subquestioon
	err = chatquestion.UpdateQuestion(ctx, question)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_question_failed" // 聊天失败
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream UpdateQuestion failed ,err %s", err)
		return
	}
}

// AgentChat 机器人api接口
// @Tags 机器人
// @Summary 机器人api调用
// @Description 机器人api接口
// @Router /chat.Agent/chat/completions [post]
func AgentChat(ctx *gin.Context) {
	var resp AgentChatResponse
	reqBodyData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] AgentChat ioutil.ReadAll failed, %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_read_request_body_failed" // error to read req body
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	req := &chattype.ChatRequestBody{}
	if err := json.Unmarshal(reqBodyData, req); err != nil {
		logs.ErrorContextf(ctx, "[chat] AgentChat BindJSON failed, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_bind_json_failed" // error to bind json
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	var (
		uin      = runtime.Uin(ctx)
		cmpID    = runtime.CompanyID(ctx)
		apikeyID = runtime.APIKeyID(ctx)
	)
	// 获取机器人信息
	agentInfo, err := chatagent.GetAgentDetailByName(ctx, req.Model)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to fetch agent info: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_fetch_agent_info_failed" // failed to fetch agent info
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	var modelID uint
	if req.LLMModelID != 0 {
		modelID = req.LLMModelID
	} else {
		if agentInfo.AgentType == chattype.AgentTypeWorkflow {
			modelID = 1
		} else {
			modelID = agentInfo.ChatModelIDs.Slice()[0]
		}
	}
	// 获取模型信息
	model, err := chatmodel.GetModelByID(ctx, modelID)
	if err != nil {
		logs.ErrorContextf(ctx, "AgentChat GetModelByID failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed" // get model failed
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    cmpID,
			Uin:          uin,
			ReqID:        runtime.RequestID(ctx),
			Status:       chattype.QuestionStatusPending,
			ApiKeyID:     apikeyID,
			BaseAgentID:  agentInfo.ID,
			AgentVersion: agentInfo.Version,
			ModelID:      model.ID,
			UserInput:    req,
			AgentName:    agentInfo.Name,
		},
	}
	wrapper := qachat.NewChatAPIWrapper(ctx, ques, model, agentInfo)
	err = wrapper.AgentChat()
	if err != nil {
		logs.ErrorContextf(ctx, "AgentChat AgentAPIChat failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed" // chat failed
		ctx.JSON(http.StatusBadRequest, resp)
		// return
		ques.Source.Status = chattype.QuestionStatusError
	}
	err = chatquestion.CreateQuestion(ctx, ques)
	if err != nil {
		logs.ErrorContextf(ctx, "AgentChat CreateQuestion err: %v", err)
		return
	}
}

// ChatGetMessage 知识森林问答恢复问答
// @Tags 知识问答
// @Summary 知识森林问答恢复问答
// @Description 知识森林问答恢复问答
// @Router /chat.ChatGetMessage [post]
// @Param user body ChatGetMessageRequest true "入参"
// @Success 200 {object} ChatGetMessageResponse "返回值"
func ChatGetMessage(ctx *gin.Context, req *ChatGetMessageRequest, resp *ChatGetMessageResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	question, err := chatquestion.GetQuetionByID(ctx, req.Request.QuestionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] [ChatGetMessage] Failed to get question by chatSession key: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_question_failed"
		return
	}
	err = llmchat.GetStreamMessage(ctx, question.Source.ReqID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_message_failed"
		return
	}
}

// StopChat 知识森林问答恢复问答
// @Tags 知识问答
// @Summary 知识森林问答恢复问答
// @Description 知识森林问答恢复问答
// @Router /chat.StopChat [post]
// @Param user body ChatGetMessageRequest true "入参"
// @Success 200 {object} ChatGetMessageResponse "返回值"
func StopChat(ctx *gin.Context, req *ChatGetMessageRequest, resp *ChatGetMessageResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	question, err := chatquestion.GetQuetionByID(ctx, req.Request.QuestionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] [StopChat] Failed to stop question by chatSession key: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_question_failed"
		return
	}
	err = llmchat.StopChatStream(ctx, question.Source.ReqID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_stop_chat_failed"
		return
	}
}

// SubmitChatQuestion Chat聊天问答
// @Tags Chat
// @Summary 聊天问答
// @Description 聊天问答
// @Router /chat.SubmitChatQuestion [post]
// @Param request body SubmitChatQuestionRequest true "request"
// @Success 200 {object} SubmitChatQuestionResponse
func SubmitChatQuestion(ctx *gin.Context, req *SubmitChatQuestionRequest, resp *SubmitChatQuestionResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.WarnContextf(ctx, "[SubmitChatQuestion] request invalid, req: %s, err: %s", logs.JSON(req.Request), resp.Message)
		return
	}
	sessionEntity, err := getChatQuestionSession(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SubmitChatQuestion] getChatQuestionSession failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed"
		return
	}

	questionEntity, err := createChatQuestion(ctx, req, sessionEntity)
	if err != nil {
		logs.ErrorContextf(ctx, "[SubmitChatQuestion] createChatQuestion failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_question_failed"
		return
	}

	// 开始问答
	questionEntity.Source.ReqID = runtime.RequestID(ctx)
	// chatsession.UpdateSessionName(ctx, sessionEntity, questionEntity.Source.Question)
	chatModelEntity, err := chatmodel.GetModelByID(ctx, sessionEntity.ModelID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed"
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream GetModelByID failed ,err %s", err)
		return
	}
	logs.InfoContextf(ctx, "SubmitChatQuestionStream chatModelEntity %s", logs.JSON(chatModelEntity))
	genFunc := func(c context.Context) *qachatnodes.State {
		state := qachatnodes.NewState(ctx)
		state.UserInput = req.Request.Question
		state.QuestionEntity = questionEntity
		state.SessionEntity = sessionEntity
		state.ModelEntity = chatModelEntity

		return state
	}
	r, err := qachatnodes.ChunkChatBuilder[string, string](ctx, genFunc)
	if err != nil {
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream ChunkChatBuilder failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed"
		return
	}
	res, err := r.Invoke(ctx, req.Request.Question)
	if err != nil {
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream Invoke failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed"
		return
	}
	logs.InfoContextf(ctx, "SubmitChatQuestionStream res %s", res)
	defer chatsession.UpdateSessionNameWithLLM(ctx, sessionEntity, questionEntity.Source.Question, questionEntity.Source.Answer)
	subquestioon, _ := chatquestion.GetLLmSubQuestion(ctx, questionEntity.Source.Question, questionEntity.Source.Answer)
	questionEntity.Source.SubQuestion = subquestioon
	if err := chatquestion.UpdateQuestion(ctx, questionEntity); err != nil {
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream UpdateQuestion failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_question_failed"
		return
	}
	resp.Response.Answer = res
	resp.Response.QuestionID = questionEntity.ID
	resp.Response.QueryReferenceList = questionEntity.Source.QueryReferenceList
}

func getChatQuestionSession(ctx *gin.Context, req *SubmitChatQuestionRequest) (*chattype.ChatSession, error) {

	if req.Request.SessionID != 0 {
		sessionEntity, err := chatsession.GetChatSession(ctx, runtime.Uin(ctx), req.Request.SessionID)
		if err != nil {
			logs.ErrorContextf(ctx, "[getChatQuestionSession] GetChatSession failed, err %v, sessionID: %d", err, req.Request.SessionID)
			return nil, err
		}
		return sessionEntity, nil
	}
	sessionEntity := chattype.ChatSession{
		Uin:          runtime.Uin(ctx),
		CompanyID:    runtime.CompanyID(ctx),
		BaseType:     req.Request.BaseType,
		ResourceType: req.Request.ResourceType,
		ModelID:      req.Request.ModelID,
		EsIndex:      chunk.DefaultIndex,
	}

	switch req.Request.ResourceType {
	case chattype.ResourceTypeForest:
		sessionEntity.ForestIDList = req.Request.ResourceIDs
	case chattype.ResourceTypeFileList:
		forestFileEntityList, err := forest.GetDirsFileByIDs(ctx, req.Request.ResourceIDs.Slice())
		if err != nil {
			logs.ErrorContextf(ctx, "[getChatQuestionSession] get file failed, err: %v, resourceIDs: %s", err, logs.JSON(req.Request.ResourceIDs.Slice()))
			return nil, err
		}
		for _, v := range forestFileEntityList {
			sessionEntity.FileIDList.Append(v.ID)
		}
	default:
		return nil, fmt.Errorf("resourceType %s not support", req.Request.ResourceType)
	}
	if err := chatsession.CreateSession(ctx, &sessionEntity); err != nil {
		logs.ErrorContextf(ctx, "[getChatQuestionSession] CreateSession failed, err: %v, sessionEntity: %s", err, logs.JSON(sessionEntity))
		return nil, err
	}
	return &sessionEntity, nil
}

func createChatQuestion(ctx *gin.Context, req *SubmitChatQuestionRequest, sessionEntity *chattype.ChatSession) (*chattype.ChatQuestion, error) {
	questionEntity := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    runtime.CompanyID(ctx),
			Uin:          runtime.Uin(ctx),
			ReqID:        runtime.RequestID(ctx),
			Question:     req.Request.Question,
			Status:       chattype.QuestionStatusPending,
			SessionID:    sessionEntity.ID,
			ModelID:      sessionEntity.ModelID,
			BaseAgentID:  sessionEntity.BaseAgentID,
			AgentVersion: sessionEntity.AgentVersion,
		},
	}
	if err := chatquestion.CreateQuestion(ctx, questionEntity); err != nil {
		logs.ErrorContextf(ctx, "[createChatQuestion] CreateQuestion failed, err: %v, questionEntity: %s", err, logs.JSON(questionEntity))
		return nil, err
	}
	return questionEntity, nil
}

// ChatQuestionStream 流式问答v2
// @Tags Chat
// @Summary 流式问答v2
// @Description 流式问答v2
// @Router /chat.ChatQuestionStream [post]
// @Param request body dtochat.ChatQuestionStreamRequest true "request"
// @Success 200 {object} dtochat.ChatQuestionStreamResponse
func ChatQuestionStream(ctx *gin.Context, req *dtochat.ChatQuestionStreamRequest, resp *dtochat.ChatQuestionStreamResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.WarnContextf(ctx, "[ChatQuestionStream] request invalid, req: %s, err: %s", logs.JSON(req.Request), resp.Message)
		return
	}
	res, err := svcchat.ChatQuestionStream(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChatQuestionStream] ChatQuestionStream failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
