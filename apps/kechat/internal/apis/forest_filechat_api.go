package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/qachat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// ListFileQA 问答对话记录列表
// @Tags 单文档智慧问答
// @Summary 问答对话记录列表
// @Description 问答对话记录列表
// @Router /chat.ListFileQA [post]
// @Param user body ListFileQARequest true "入参"
// @Success 200 {object} ListFileQAResponse "返回值"
func ListFileQA(ctx *gin.Context, req *ListFileQARequest, resp *ListFileQAResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContext(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_get_forest_failed"
		return
	}
	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContext(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_get_file_failed"
		return
	}
	if f.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_knowledge_not_generated" // 知识库未生成
		return
	}
	session, err := chatsession.GetFileLastSession(ctx, uin, f.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有创建一个会话
			session = &chattype.ChatSession{
				CompanyID:    runtime.CompanyID(ctx),
				Uin:          runtime.Uin(ctx),
				ResourceType: chattype.ResourceTypeFile,
				BaseType:     chattype.ResourceQASessionBaseTypeStandard,
				FileID:       f.ID,

				// TODO 支持多配置后动态选项es索引并且判断所选文件是否在一个索引中
				EsIndex: forest_info.EsIndex(),
			}
			err = chatsession.CreateSession(ctx, session)
			if err != nil {
				resp.Code = errcode.ErrCode_InternalError
				resp.Message = "kechat_create_session_failed"
				return
			}
			return
		}
		logs.ErrorContextf(ctx, "GetFileLastSession failed ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed"
		return
	}
	chats, err := chatquestion.ListSessionQuestionsByUin(ctx, runtime.Uin(ctx), session.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "QueryForestQAList failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_query_failed"
		return
	}
	resp.Response.Data = chats
}

// FileChat 问答对话
// @Tags 单文档智慧问答
// @Summary 问答对话
// @Description 问答对话
// @Router /chat.FileChat [post]
// @Param user body FileChatRequest true "入参"
// @Success 200 {object} FileChatResponse "返回值"
func FileChat(ctx *gin.Context, req *FileChatRequest, resp *FileChatResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	_, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContext(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_get_forest_failed" // 查询知识森林失败
		return
	}
	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContext(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_get_file_failed" // 未查找到文件信息
		return
	}
	if f.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_knowledge_not_generated" // 知识库未生成
		return
	}
	session, err := chatsession.GetFileLastSession(ctx, uin, f.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetFileLastSession failed ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed" // 查询会话失败
		return
	}

	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    runtime.CompanyID(ctx),
			Uin:          uin,
			ReqID:        runtime.RequestID(ctx),
			Question:     req.Request.Question,
			Status:       chattype.QuestionStatusPending,
			SessionID:    session.ID,
			ModelID:      session.ModelID,
			BaseAgentID:  session.BaseAgentID,
			AgentVersion: session.AgentVersion,
		},
	}
	err = chatquestion.CreateQuestion(ctx, ques)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_question_failed" // 创建问题失败
		logs.ErrorContextf(ctx, "NewChatQuestionStream CreateQuestion err: %v", err)
		return
	}
	session.FileIDList.Append(session.FileID)

	wrapper := qachat.NewChatWrapper(ctx, ques, session, &chattype.ChatModel{
		Model: gorm.Model{
			ID: session.ModelID,
		},
	})
	err = wrapper.ForestChat()
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed" // 问答失败
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream ForestChat failed ,err %s", err)
		ques.Source.Status = chattype.QuestionStatusError
	}
	err = chatquestion.UpdateQuestion(ctx, ques)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_question_failed" // 聊天失败
		logs.ErrorContextf(ctx, "SubmitChatQuestionStream UpdateQuestion failed ,err %s", err)
		return
	}
	resp.Response.Status = chattype.QuestionStatusAnswered
	resp.Response.Answer = ques.Source.Answer
}

// DeleteFileQA 单文档问答清除历史记录
// @Tags 单文档智慧问答
// @Summary 单文档问答清除历史记录
// @Description 单文档问答清除历史记录
// @Router /chat.DeleteFileQA [post]
// @Param user body DeleteFileQARequest true "入参"
// @Success 200 {object} DeleteFileQAResponse "返回值"
func DeleteFileQA(ctx *gin.Context, req *DeleteFileQARequest, resp *DeleteFileQAResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	uin := runtime.Uin(ctx)
	forestInfo, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContext(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_get_forest_failed" // 查询知识森林失败
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContext(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_get_file_failed" // 未查找到文件信息
		return
	}
	if f.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_knowledge_not_generated" // 知识库未生成
		return
	}
	session := &chattype.ChatSession{
		CompanyID:    runtime.CompanyID(ctx),
		Uin:          runtime.Uin(ctx),
		ResourceType: chattype.ResourceTypeFile,
		BaseType:     chattype.ResourceQASessionBaseTypeStandard,
		FileID:       f.ID,

		// TODO 支持多配置后动态选项es索引并且判断所选文件是否在一个索引中
		EsIndex: forestInfo.EsIndex(),
	}

	sessionOld, err := chatsession.GetFileLastSession(ctx, uin, f.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err := chatsession.CreateSession(ctx, session)
			if err != nil {
				resp.Code = errcode.ErrCode_InternalError
				resp.Message = "kechat_create_session_failed" // 创建会话失败
				return
			}
			return
		}
		logs.ErrorContextf(ctx, "GetFileLastSession failed ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed" // 查询会话失败
		return
	}
	chats, err := chatquestion.ListSessionQuestionsByUin(ctx, runtime.Uin(ctx), sessionOld.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "QueryForestQAList failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_query_failed" // 查询失败
		return
	}
	if len(chats) == 0 {
		return
	}
	err = chatsession.CreateSession(ctx, session)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_session_failed" // 创建会话失败
		return
	}
}
