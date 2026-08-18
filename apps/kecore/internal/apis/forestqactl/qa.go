package forestqactl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ListSession 会话列表
// @Tags 知识问答
// @Summary 会话列表
// @Description 会话列表
// @Router /forest.ListSession [post]
// @Param user body ListSessionRequest true "入参"
// @Success 200 {object} ListSessionResponse "返回值"
func ListSession(ctx *gin.Context, req *ListSessionRequest, resp *ListSessionResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	req.Request.Filters = []apiobj.Filter{
		{
			Field: "type",
			Value: []string{
				string(foresttype.KnownowQASessionTypeForest),
				string(foresttype.KnownowQASessionTypeFileList),
				string(foresttype.KnownowQASessionTypeDirList),
				string(foresttype.KnownowQASessionTypeExcelList),
				string(foresttype.KnownowQASessionTypeExcelSheetList),
			},
			ExactMatch: false,
		},
	}
	req.Request.Uin = runtime.Uin(ctx)
	err := keqa.QueryQASessionList(ctx, &req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_failed" // 查询失败
		return
	}
}

// ListSessionQA 会话对话记录
// @Tags 知识问答
// @Summary 会话对话记录
// @Description 会话对话记录
// @Router /forest.ListSessionQA [post]
// @Param user body ListSessionQARequest true "入参"
// @Success 200 {object} ListSessionQAResponse "返回值"
func ListSessionQA(ctx *gin.Context, req *ListSessionQARequest, resp *ListSessionQAResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	_, err := keqa.GetCustomerSession(uin, req.Request.SessionID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_session_not_found" // 该会话不存在
		return
	}
	err = keqa.QueryForestQAList(&keqa.QueryForestQAListOption{
		Uin:       uin,
		SessionID: req.Request.SessionID,
	}, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_failed" // 查询失败
		return
	}
}

// CreateForestSession 创建知识森林会话
// @Tags 知识问答
// @Summary 创建知识森林会话
// @Description 创建知识森林会话
// @Router /forest.CreateForestSession [post]
// @Param user body CreateForestSessionRequest true "入参"
// @Success 200 {object} CreateForestSessionResponse "返回值"
func CreateForestSession(ctx *gin.Context, req *CreateForestSessionRequest, resp *CreateForestSessionResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	session := &foresttype.KnownowQASession{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       runtime.Uin(ctx),
		Name:      foresttype.DefaultSessionName,
		Type:      req.Request.SessionType,
		// TODO 支持多配置后动态选项es索引并且判断所选文件是否在一个索引中
		EsIndex:    "ke_0",
		LLMModelID: req.Request.LLMModelID,
	}
	switch req.Request.SessionType {
	case foresttype.KnownowQASessionTypeForest:
		foreseEntityList, err := forest.NewForestDao().GetListByCond(ctx, &forest.ForestCond{
			IDs: req.Request.IDS.Slice(),
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get forest failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_forest_failed" // 查找知识库失败
		}
		excelCount := 0
		for _, v := range foreseEntityList {
			switch v.DataSourceSubType {
			case foresttype.ForestDataSourceSubtypeExcel:
				excelCount++
			}
		}
		session.BaseType = foresttype.KnownowQASessionBaseTypeStandard
		if excelCount == len(foreseEntityList) {
			session.BaseType = foresttype.KnownowQASessionBaseTypeExcel
		}
		session.ForestIDList = req.Request.IDS
	case foresttype.KnownowQASessionTypeFileList:
		file, err := forest.GetDirsFileByIDs(ctx, req.Request.IDS.Slice())
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_file_failed" // 查找文件失败
		}
		for _, v := range file {
			session.FileIDList.Append(v.ID)
		}
		session.BaseType = foresttype.KnownowQASessionBaseTypeStandard

	case foresttype.KnownowQASessionTypeExcelList:
		forestFileEntityList, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
			IDs:         req.Request.IDS.Slice(),
			ParseStatus: foresttype.TaskStatusSuccess,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get file list failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_file_failed" // 查找文件失败
		}
		if len(forestFileEntityList) != len(req.Request.IDS.Slice()) {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_file_not_parsed" // 文件未解析完成
		}
		for _, v := range forestFileEntityList {
			session.ExcelIDList.Append(v.ID)
		}
		session.BaseType = foresttype.KnownowQASessionBaseTypeExcel
	case foresttype.KnownowQASessionTypeExcelSheetList:
		sheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, &forest.ForestExcelSheetCond{
			IDS: req.Request.IDS.Slice(),
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get sheet list failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_sheet_failed" // 查找sheet失败
		}
		for _, v := range sheetEntityList {
			session.ExcelSheetIDList.Append(v.ID)
		}
		session.BaseType = foresttype.KnownowQASessionBaseTypeExcel
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_session_type" // 未知的会话类型
		return
	}

	//session.Type = foresttype.KnownowQASessionTypeFileList
	out, err := keqa.CreateForestSession(session)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_session_failed" // 创建会话失败
		return
	}
	resp.Response = out
}

// ModifySession 修改会话
// @Tags 知识问答
// @Summary 修改会话
// @Description 修改会话
// @Router /forest.ModifySession [post]
// @Param user body ModifySessionRequest true "入参"
// @Success 200 {object} ModifySessionResponse "返回值"
func ModifySession(ctx *gin.Context, req *ModifySessionRequest, resp *ModifySessionResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	session, err := keqa.GetCustomerSession(uin, req.Request.SessionID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_session_not_found" // 该会话不存在
		return
	}
	session.Name = req.Request.Name

	if err := keqa.ModifySession(ctx, session); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_modify_session_failed" // 修改会话失败
		return
	}
}

// DeleteSession 删除会话
// @Tags 知识问答
// @Summary 删除会话
// @Description 删除会话
// @Router /forest.DeleteSession [post]
// @Param user body DeleteSessionRequest true "入参"
// @Success 200 {object} DeleteSessionResponse "返回值"
func DeleteSession(ctx *gin.Context, req *DeleteSessionRequest, resp *DeleteSessionResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	session, err := keqa.GetCustomerSession(runtime.Uin(ctx), req.Request.SessionID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_session_not_found" // 该会话不存在
		return
	}

	if err := keqa.DeleteSession(ctx, session.ID); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_session_failed" // 删除会话失败
		return
	}
}

// CreateForestQA 2.创建知识森林问答
// @Tags 知识问答
// @Summary 2.创建知识森林问答
// @Description 2.创建知识森林问答
// @Router /forest.CreateForestQA [post]
// @Param user body CreateForestQARequest true "入参"
// @Success 200 {object} CreateForestQAResponse "返回值"
func CreateForestQA(ctx *gin.Context, req *CreateForestQARequest, resp *CreateForestQAResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	session, err := keqa.GetCustomerSession(uin, req.Request.SessionID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_session_not_found" // 该会话不存在
		return
	}

	ret, err := keqa.CreateForestQA(session, req.Request.Question, req.Request.ImageUrlList)
	if err != nil {
		logs.ErrorContextf(ctx, "forestqa.CreateForestQA failed, err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_question_failed" // 创建问题失败
		return
	}
	resp.Response = ret
}

// ForestQAChat 5.知识森林问答
// Deprecated: 该接口已废弃，使用 chat.NewChatQuestionStream 路由接口
// @Tags 知识问答
// @Summary 5.知识森林问答（流式输出）
// @Description 5.知识森林问答
// @Router /forest.ForestQAChat [post]
// @Param user body ForestQAChatRequest true "入参"
// @Success 200 {object} ForestQAChatResponse "返回值"
func ForestQAChat(ctx *gin.Context, req *ForestQAChatRequest, resp *ForestQAChatResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	session, err := keqa.GetCustomerSession(uin, req.Request.SessionID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_session_not_found" // 该会话不存在
		return
	}

	qa, err := keqa.GetForestQA(req.Request.QuestionID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_query_failed" // 查询问题失败
		return
	}
	switch session.BaseType {
	case foresttype.KnownowQASessionBaseTypeExcel:
		qs, err := keqa.NewQA(foresttype.KnownowQASessionBaseTypeExcel).Chat(ctx, qa, session)
		if err != nil {
			logs.ErrorContextf(ctx, "excelqa.ForestChat failed, err = %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_chat_failed" // 对话失败
			return
		}
		resp.Response.Status = foresttype.QAStatusAnswered
		resp.Response.Answer = qs.Answer
	default:
		qs, err := keqa.ForestChat(ctx, qa, session)
		if err != nil {
			logs.ErrorContextf(ctx, "forestqa.ForestChat failed, err = %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_chat_failed" // 对话失败
			return
		}
		resp.Response.Status = foresttype.QAStatusAnswered
		resp.Response.Answer = qs.Answer
	}
}

// ChatGetMessage 知识森林问答恢复问答
// @Tags 知识问答
// @Summary 知识森林问答恢复问答
// @Description 知识森林问答恢复问答
// @Router /forest.ChatGetMessage [post]
// @Param user body ChatGetMessageRequest true "入参"
// @Success 200 {object} ChatGetMessageResponse "返回值"
func ChatGetMessage(ctx *gin.Context, req *ChatGetMessageRequest, resp *ChatGetMessageResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := keqa.GetStreamMessage(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_message_failed" // 获取消息失败
		return
	}
}

// StopChat 知识森林问答恢复问答
// @Tags 知识问答
// @Summary 知识森林问答恢复问答
// @Description 知识森林问答恢复问答
// @Router /forest.StopChat [post]
// @Param user body ChatGetMessageRequest true "入参"
// @Success 200 {object} ChatGetMessageResponse "返回值"
func StopChat(ctx *gin.Context, req *ChatGetMessageRequest, resp *ChatGetMessageResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := keqa.StopChatStream(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_message_failed" // 获取消息失败
		return
	}
}

// GetSessionInfo 获取session信息
// @Tags 知识问答
// @Summary 获取session信息
// @Description 获取session信息
// @Router /forest.GetSessionInfo [post]
// @Param user body GetSessionInfoRequest true "入参"
// @Success 200 {object} GetSessionInfoResponse "返回值"
func GetSessionInfo(ctx *gin.Context, req *GetSessionInfoRequest, resp *GetSessionInfoResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	si, err := keqa.GetSessionInfo(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_session_info_failed" // 获取会话信息失败
		return
	}
	resp.Response = si
}
