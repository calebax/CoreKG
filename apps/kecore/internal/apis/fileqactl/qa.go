package fileqactl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// ListFileQA 问答对话记录列表
// @Tags 单文档智慧问答
// @Summary 问答对话记录列表
// @Description 问答对话记录列表
// @Router /forest.ListFileQA [post]
// @Param user body ListFileQARequest true "入参"
// @Success 200 {object} ListFileQAResponse "返回值"
func ListFileQA(ctx *gin.Context, req *ListFileQARequest, resp *ListFileQAResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_forest_failed" // 查询知识森林失败
		return
	}
	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContextf(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_not_found" // 未查找到文件信息
		return
	}
	if f.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_knowledge_not_generated" // 知识库未生成
		return
	}
	session, err := keqa.GetFileLastSession(uin, f.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有创建一个会话
			session := &foresttype.KnownowQASession{
				CompanyID: runtime.CompanyID(ctx),
				Uin:       uin,
				Name:      foresttype.DefaultSessionName,
				// TODO 支持多配置后动态选项es索引并且判断所选文件是否在一个索引中
				EsIndex: forest_info.EsIndex(),
				Type:    foresttype.KnownowQASessionTypeFile,
				FileID:  f.ID,
			}
			_, err := keqa.CreateForestSession(session)
			if err != nil {
				resp.Code = errcode.ErrCode_InternalError
				resp.Message = "kecore_create_session_failed" // 创建会话失败
				return
			}
			return
		}
		logs.ErrorContextf(ctx, "GetFileLastSession failed ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_session_failed" // 查询会话失败
		return
	}
	err = keqa.QueryForestQAList(&keqa.QueryForestQAListOption{
		Uin:       uin,
		SessionID: session.ID,
	}, &resp.Response)
	if err != nil {
		logs.ErrorContextf(ctx, "QueryForestQAList failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_failed" // 查询失败
		return
	}
}

// DeleteFileQA 单文档问答清除历史记录
// @Tags 单文档智慧问答
// @Summary 单文档问答清除历史记录
// @Description 单文档问答清除历史记录
// @Router /forest.DeleteFileQA [post]
// @Param user body DeleteFileQARequest true "入参"
// @Success 200 {object} DeleteFileQAResponse "返回值"
func DeleteFileQA(ctx *gin.Context, req *DeleteFileQARequest, resp *DeleteFileQAResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	uin := runtime.Uin(ctx)
	forest_info, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_forest_failed" // 查询知识森林失败
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContextf(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_not_found" // 未查找到文件信息
		return
	}
	if f.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_knowledge_not_generated" // 知识库未生成
		return
	}
	session := &foresttype.KnownowQASession{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       uin,
		Name:      foresttype.DefaultSessionName,
		// TODO 支持多配置后动态选项es索引并且判断所选文件是否在一个索引中
		EsIndex: forest_info.EsIndex(),
		Type:    foresttype.KnownowQASessionTypeFile,
		FileID:  f.ID,
	}
	session_old, err := keqa.GetFileLastSession(uin, f.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			_, err := keqa.CreateForestSession(session)
			if err != nil {
				resp.Code = errcode.ErrCode_InternalError
				resp.Message = "kecore_create_session_failed" // 创建会话失败
				return
			}
			return
		}
		logs.ErrorContextf(ctx, "GetFileLastSession failed ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_session_failed" // 查询会话失败
		return
	}
	qa_list := &keqa.QueryForestQAListResponse{}
	err = keqa.QueryForestQAList(&keqa.QueryForestQAListOption{
		Uin:       uin,
		SessionID: session_old.ID,
	}, qa_list)
	if err != nil {
		logs.ErrorContextf(ctx, "QueryForestQAList failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_failed" // 查询失败
		return
	}
	if len(qa_list.Data) == 0 {
		return
	}
	_, err = keqa.CreateForestSession(session)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_session_failed" // 创建会话失败
		return
	}
}

// FileChat 问答对话
// @Tags 单文档智慧问答
// @Summary 问答对话
// @Description 问答对话
// @Router /forest.FileChat [post]
// @Param user body FileChatRequest true "入参"
// @Success 200 {object} FileChatResponse "返回值"
func FileChat(ctx *gin.Context, req *FileChatRequest, resp *FileChatResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	_, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_forest_failed" // 查询知识森林失败
		return
	}
	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContextf(ctx, "PrepareFileQA failed ,err = %v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_not_found" // 未查找到文件信息
		return
	}
	if f.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_knowledge_not_generated" // 知识库未生成
		return
	}
	session, err := keqa.GetFileLastSession(uin, f.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetFileLastSession failed ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_session_failed" // 查询会话失败
		return
	}
	qa, err := keqa.CreateForestQA(session, req.Request.Question, nil)
	if err != nil {
		logs.WarnContextf(ctx, "forestqa.CreateForestQA faied ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_question_failed" // 创建问题失败
		return
	}
	session.FileIDList.Append(session.FileID)
	qs, err := keqa.ForestChat(ctx, qa, session)
	if err != nil {
		logs.WarnContextf(ctx, "forestqa.ForestChat faied ,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_chat_failed" // 对话失败
		return
	}
	resp.Response.Status = foresttype.QAStatusAnswered
	resp.Response.Answer = qs.Answer
}
