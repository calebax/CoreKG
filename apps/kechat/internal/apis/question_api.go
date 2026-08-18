package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// NewChatQuestionStream 新建聊天问题流
// @Tags question
// @Summary 新建聊天问题流
// @Description 新建聊天问题流
// @Router /chat.NewChatQuestionStream [post]
// @Param request body NewChatQuestionRequest true "request"
// @Success 200 {object} NewChatQuestionResponse
func NewChatQuestionStream(ctx *gin.Context, req *NewChatQuestionRequest, resp *NewChatQuestionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "NewChatQuestionStream validate params failed : %v", resp.Message)
		return
	}
	uin := runtime.Uin(ctx)
	session, err := chatsession.GetChatSession(ctx, uin, req.Request.SessionID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed" // 查询会话记录失败
		logs.ErrorContextf(ctx, "NewChatQuestionStream GetChatSession err: %v", err)
		return
	}
	agentExtra := chattype.AgentExtra{}
	if req.Request.Options != nil {
		agentExtra.EnableWebSearch = req.Request.Options.EnableWebSearch
	}

	inputExtra := chattype.InputExtra{}
	if req.Request.Input != nil {
		for _, attachment := range req.Request.Input.Attachments {
			inputExtra.Attachments = append(inputExtra.Attachments, chattype.AttachmentInfo{
				Url:   attachment.URL,
				MdUrl: attachment.MdURL,
				Type:  attachment.Type,
				Name:  attachment.Name,
			})
		}
	}
	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    runtime.CompanyID(ctx),
			Uin:          uin,
			ImageUrlList: req.Request.ImageUrlList,
			ReqID:        runtime.RequestID(ctx),
			Question:     req.Request.Question,
			Status:       chattype.QuestionStatusPending,
			SessionID:    session.ID,
			ModelID:      session.ModelID,
			BaseAgentID:  session.BaseAgentID,
			AgentVersion: session.AgentVersion,
			Extra: &chattype.ExtraInfo{
				Agent: &agentExtra,
				Input: &inputExtra,
			},
		},
	}
	err = chatquestion.CreateQuestion(ctx, ques)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_question_failed" // 创建问题失败
		logs.ErrorContextf(ctx, "NewChatQuestionStream CreateQuestion err: %v", err)
		return
	}
	resp.Response.QuestionID = ques.ID
}

// GetQuestionInfo 获取问题详情
// @Tags question
// @Summary 获取问题详情
// @Description 获取问题详情
// @Router /chat.GetQuestionInfo [post]
// @Param request body GetQuestionInfoRequest true "request"
// @Success 200 {object} GetQuestionInfoResponse
func GetQuestionInfo(ctx *gin.Context, req *GetQuestionInfoRequest, resp *GetQuestionInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "NewChatQuestionStream validate params failed : %v", resp.Message)
		return
	}
	question, err := chatquestion.GetQuetionByID(ctx, req.Request.QuestionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] [SubmitChatQuestionStream] Failed to get question by chatSession key: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_question_failed" // 获取问题失败
		return
	}
	resp.Response.Question.ChatQuestion = question
	if question.Source.ReactAgentService != nil {
		resp.Response.Question.Msg = utils.ConvertMsg2WriteResult(ctx, question.Source.ReactAgentService.Memory.Messages)
		if len(question.Source.ReactAgentService.Rresult) > 0 {
			resp.Response.Question.Msg = append(resp.Response.Question.Msg, question.Source.ReactAgentService.Rresult...)
		}
	}
}
