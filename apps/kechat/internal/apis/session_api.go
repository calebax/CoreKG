package apis

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/project"
	"github.com/insmtx/corekg/pkgs/apis/code"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

// NewChatSession 新建聊天会话
// @Tags ChatSession
// @Summary 新建聊天会话
// @Description 新建聊天会话
// @Router /chat.NewChatSession [post]
// @Param request body NewChatSessionRequest true "request"
// @Success 200 {object} NewChatSessionResponse
func NewChatSession(ctx *gin.Context, req *NewChatSessionRequest, resp *NewChatSessionResponse) {
	if req.Request.ResourceType == chattype.ResourceTypeAgent && req.Request.ModelID == 0 {
		req.Request.ModelID = 1
	}
	defer func() {
		useModelEntity := &chattype.ChatRecentUsedModel{
			Uin:       runtime.Uin(ctx),
			CompanyID: runtime.CompanyID(ctx),
			ModelID:   req.Request.ModelID,
		}
		if err := chat.NewChatRecentUsedModelDao().Upsert(ctx, useModelEntity); err != nil {
			logs.ErrorContextf(ctx, "[ChatQuestionStream] Failed to upsert recent used model, entity: %s, err: %v", logs.JSON(useModelEntity), err)
		}
	}()
	if req.Validity(ctx, resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "NewChatSession validate params failed : %v", resp.Message)
		return
	}
	// 绑定知识库
	model, err := chatmodel.GetModelByID(ctx, req.Request.ModelID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed" // 查找模型失败
		return
	}
	var (
		Uin       = runtime.Uin(ctx)
		CompanyID = runtime.CompanyID(ctx)
	)

	sess := &chattype.ChatSession{
		CompanyID:    CompanyID,
		Uin:          Uin,
		ResourceType: req.Request.ResourceType,
		ModelID:      model.ID,
		ModelName:    model.ShowName,
		// TODO 支持多配置后动态选项es索引并且判断所选文件是否在一个索引中
		EsIndex:   "ke_0",
		SubjectID: req.Request.ProjectID,
	}

	if req.Request.SourceFrom == SourceFromAgent {
		// Create rel to the default project for agent
		pro, err := forest.NewKeProjectDao().GetByCond(ctx, &forest.KeProjectCond{
			Uin:             Uin,
			ProjectTypeList: []foresttype.ProjectType{foresttype.ProjectTypeAgentQA},
		})
		if err != nil {
			logs.ErrorContextf(ctx, "failed to fetch project info: %w", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_project_failed"
			return
		}
		sess.SubjectID = pro.ID
	}

	var agentInfo *chatagent.AgentWithVersion
	pro := &foresttype.KnownowProject{}
	if req.Request.ProjectID != 0 {
		pro, err = project.GetProjectByID(ctx, req.Request.ProjectID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_project_failed" // 查找项目失败
			return
		}
	}
	// TODO：前端传basetype
	switch req.Request.ResourceType {
	case chattype.ResourceTypeAgent:
		//Get Agent Detail
		agentInfo, err = chatagent.GetAgentDetail(ctx, req.Request.ResourceID)
		if err != nil {
			logs.ErrorContextf(ctx, "failed to fetch agent info: %w", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_server_error" // 服务器错误
			return
		}
		sess.AgentVersion = agentInfo.Version
		sess.BaseAgentID = agentInfo.ID
		sess.ForestIDList = types.NewUintArray(agentInfo.ForestOption.ForestIDs)
		sess.Input = req.Request.Input
		sess.BaseType = chattype.ResourceQASessionBaseAgent
	case chattype.ResourceTypeForest:
		foreseEntityList, err := forest.NewForestDao().GetListByCond(ctx, &forest.ForestCond{
			IDs: req.Request.IDS.Slice(),
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get forest failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_forest_failed" // 查找知识库失败
		}
		excelCount := 0
		mysqlCount := 0
		for _, v := range foreseEntityList {
			switch v.DataSourceSubType {
			case foresttype.ForestDataSourceSubtypeExcel:
				excelCount++
			case foresttype.ForestDataSourceSubtypeMySQL:
				mysqlCount++
			}
		}
		sess.BaseType = chattype.ResourceQASessionBaseTypeStandard
		if excelCount == len(foreseEntityList) {
			sess.BaseType = chattype.ResourceQASessionBaseTypeExcel
		}
		if mysqlCount == len(foreseEntityList) {
			sess.BaseType = chattype.ResourceQASessionBaseTypeDbMYSQL
		}
		sess.ForestIDList = req.Request.IDS
		for _, v := range req.Request.IDS.Slice() {
			pro.ForestIDList.Append(v)
		}
		sess.PromptMode = req.Request.PromptKey
	case chattype.ResourceTypeFileList, chattype.ResourceTypeDirList:
		file, err := forest.GetDirsFileByIDs(ctx, req.Request.IDS.Slice())
		if err != nil {
			logs.ErrorContextf(ctx, "[NewChatSession] get file failed, err: %v, resourceIDs: %s", err, logs.JSON(req.Request.IDS.Slice()))
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_file_failed" // 查找文件失败
		}
		forestIDs := []uint{}
		for _, v := range file {
			sess.FileIDList.Append(v.ID)
			forestIDs = append(forestIDs, v.ForestID)
		}
		sess.BaseType = chattype.ResourceQASessionBaseTypeStandard
		forests := types.NewUintArray(forestIDs)
		forests.RemoveDuplicates()
		for _, v := range forests.Slice() {
			pro.ForestIDList.Append(v)
		}
		sess.PromptMode = req.Request.PromptKey
	case chattype.ResourceTypeExcelList:
		forestFileEntityList, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
			IDs:         req.Request.IDS.Slice(),
			ParseStatus: foresttype.TaskStatusSuccess,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get file list failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_file_failed" // 查找文件失败
		}
		if len(forestFileEntityList) != len(req.Request.IDS.Slice()) {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_file_not_parsed" // 文件未解析完成
		}
		forestIDs := []uint{}
		for _, v := range forestFileEntityList {
			sess.ExcelIDList.Append(v.ID)
			forestIDs = append(forestIDs, v.ForestID)
		}
		forests := types.NewUintArray(forestIDs)
		forests.RemoveDuplicates()
		for _, v := range forests.Slice() {
			pro.ForestIDList.Append(v)
		}
		sess.BaseType = chattype.ResourceQASessionBaseTypeExcel
	case chattype.ResourceTypeReactExcelList:
		forestFileEntityList, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
			IDs:         req.Request.IDS.Slice(),
			ParseStatus: foresttype.TaskStatusSuccess,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get file list failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_file_failed" // 查找文件失败
		}
		if len(forestFileEntityList) != len(req.Request.IDS.Slice()) {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_file_not_parsed" // 文件未解析完成
		}
		forestIDs := []uint{}
		for _, v := range forestFileEntityList {
			sess.ExcelIDList.Append(v.ID)
			forestIDs = append(forestIDs, v.ForestID)
		}
		forests := types.NewUintArray(forestIDs)
		forests.RemoveDuplicates()
		for _, v := range forests.Slice() {
			pro.ForestIDList.Append(v)
		}
		sess.BaseType = chattype.ResourceQASessionBaseTypeReactExcel
	case chattype.ResourceTypeExcelSheetList:
		sheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, &forest.ForestExcelSheetCond{
			IDS: req.Request.IDS.Slice(),
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[CreateForestSession] get sheet list failed, err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_sheet_failed" // 查找sheet失败
		}
		forestIDs := []uint{}
		for _, v := range sheetEntityList {
			sess.ExcelSheetIDList.Append(v.ID)
			forestIDs = append(forestIDs, v.ForestID)
		}
		forests := types.NewUintArray(forestIDs)
		forests.RemoveDuplicates()
		for _, v := range forests.Slice() {
			pro.ForestIDList.Append(v)
		}
		sess.BaseType = chattype.ResourceQASessionBaseTypeExcel
	case chattype.ResourceTypeDBList:
		if req.Request.ResourceID == 0 {
			logs.ErrorContextf(ctx, "[CreateForestSession] mysql db chat, resource id is empty")
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_forest_id_required" // 知识库ID不能为空
			return
		}
		sess.ForestIDList = types.NewUintArray([]uint{req.Request.ResourceID})
		pro.ForestIDList.Append(req.Request.ResourceID)
		sess.BaseType = chattype.ResourceQASessionBaseTypeDbMYSQL
		sess.DBList = req.Request.Names
	case chattype.ResourceTypeDBTableList:
		if req.Request.ResourceID == 0 {
			logs.ErrorContextf(ctx, "[CreateForestSession] mysql table chat, resource id is empty")
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_forest_id_required" // 知识库ID不能为空
			return
		}
		pro.ForestIDList.Append(req.Request.ResourceID)
		sess.ForestIDList = types.NewUintArray([]uint{req.Request.ResourceID})
		sess.BaseType = chattype.ResourceQASessionBaseTypeDbMYSQL
		sess.DBTableList = req.Request.Names
	case chattype.ResourceTypeModel:
		// model 啥也不用干
		sess.BaseType = chattype.ResourceQASessionBaseModel
	case chattype.ResourceTypeExternalData:
		// 外部数据源问答
		sess.ExternalTokenIDList = req.Request.IDS
		sess.BaseType = chattype.ResourceQASessionBaseExternalData
	case chattype.ResourceTypeGraphSearch:
		// model 啥也不用干
		sess.BaseType = chattype.ResourceQASessionBaseTypeGraphSearch
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_unknown_session_type" // 未知的会话类型
		return
	}
	if req.Request.BaseType != "" {
		sess.BaseType = req.Request.BaseType
	}

	// 创建聊天会话
	err = chatsession.CreateSession(ctx, sess)
	if err != nil {
		logs.ErrorContext(ctx, "NewChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		return
	}
	// 保存项目
	if req.Request.ProjectID != 0 {
		err = project.UpdateProject(ctx, pro)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_update_project_failed" // 更新项目失败
			return
		}
	}
	// 如果是角色机器人插入一个问答记录用于展示问候语
	if req.Request.ResourceType == chattype.ResourceTypeAgent && agentInfo.GreetingMessage != "" {
		ques := &chattype.ChatQuestion{
			Source: &chattype.Question{
				CompanyID:    runtime.CompanyID(ctx),
				Uin:          runtime.Uin(ctx),
				ReqID:        runtime.RequestID(ctx),
				Answer:       agentInfo.GreetingMessage,
				Status:       chattype.QuestionStatusAnswered,
				SessionID:    sess.ID,
				BaseAgentID:  agentInfo.ID,
				AgentVersion: sess.AgentVersion,
				ModelID:      sess.ModelID,
			},
		}
		err = chatquestion.CreateQuestion(ctx, ques)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_create_greeting_failed" // 创建问候语失败
			logs.ErrorContextf(ctx, "NewChatQuestionStream CreateQuestion err: %v", err)
			return
		}
	}

	resp.Response.ChatSession = sess
}

// func AA(ctx context.Context, sess *chattype.ChatSession, projectID uint) error {
// 	pro, err := project.GetProjectByID(ctx, projectID)
// 	if err != nil {
// 		return err
// 	}
// 	forestIDList := []uint{}

// 	if sess.ForestIDList.Len() != 0 {
// 		forestIDList = append(forestIDList, sess.FileIDList.Slice()...)
// 	}
// 	if sess.FileIDList.Len() != 0 {
// 		forestIDList = append(forestIDList, sess.FileIDList.Slice()...)
// 	}
// 	if sess.ExcelIDList.Len() != 0 {
// 	}
// 	if sess.ExcelSheetIDList.Len() != 0 {
// 	}
// 	return nil
// }

// RenameChatSession 重命名聊天会话
// @Tags ChatSession
// @Summary 重命名聊天会话
// @Description 重命名聊天会话
// @Router /chat.RenameChatSession [post]
// @Param request body apiobj.DetailNameRequest true "request"
// @Success 200 {object} NewChatSessionResponse
func RenameChatSession(ctx *gin.Context, req *apiobj.DetailNameRequest, resp *NewChatSessionResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	if err := validate.IsTitle(req.Request.Name); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = err.Error()
		return
	}
	session, err := chatsession.GetChatSession(ctx, runtime.Uin(ctx), req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_session_failed" // 查找记录失败
		return
	}
	session.Name = req.Request.Name
	err = chatsession.UpdateChatSession(ctx, session)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_name_failed" // 修改名称失败
		return
	}

	resp.Response.ChatSession = session
}

// RemoveChatSession 删除聊天会话
// @Tags ChatSession
// @Summary 删除聊天会话
// @Description 删除聊天会话
// @Router /chat.RemoveChatSession [post]
// @Param request body apiobj.DetailIdRequest true "request"
// @Success 200 {object} apiobj.BaseResponse
func RemoveChatSession(ctx *gin.Context, req *apiobj.DetailIdRequest, resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	// 暂时不删除问答记录
	// err := chatquestion.DeleteSessionQuestions(ctx, runtime.Uin(ctx), req.Request.ID)
	// if err != nil {
	// 	logs.ErrorContext(ctx,"RemoveChatSession error: %v", err)
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kechat_server_error" // 服务器错误
	// 	return
	// }
	err := chatsession.DeleteSession(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContext(ctx, "RemoveChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		return
	}
}

// ListChatSession 列出聊天会话
// @Tags ChatSession
// @Summary 列出聊天会话
// @Description 列出聊天会话
// @Router /chat.ListChatSession [post]
// @Param request body ListChatSessionRequest true "request"
// @Success 200 {object} ListChatSessionResponse
func ListChatSession(ctx *gin.Context, req *ListChatSessionRequest, resp *ListChatSessionResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "ListChatSession validate params failed")
		return
	}
	req.Request.Uin = runtime.Uin(ctx)
	req.Request.CompanyID = runtime.CompanyID(ctx)
	err := chatsession.QueryListChatSessions(ctx, req.Request.PageQuery, req.Request.ProjectID, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		logs.ErrorContextf(ctx, "ListChatSession failed ,err %s", err)
		return
	}
}

// ListSessionChats 列出聊天会话的聊天记录
// @Tags ChatSession
// @Summary 列出聊天会话的聊天记录
// @Description 列出聊天会话的聊天记录
// @Router /chat.ListSessionChats [post]
// @Param request body apiobj.DetailIdRequest true "request"
// @Success 200 {object} ListSessionChatsResponse
func ListSessionChats(ctx *gin.Context, req *apiobj.DetailIdRequest, resp *ListSessionChatsResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}

	chats, err := chatquestion.ListSessionQuestionsByUin(ctx, runtime.Uin(ctx), req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "ListSessionChats error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		return
	}
	data := []*Question{}
	for _, v := range chats {
		q := &Question{}
		if v.Source.ReactAgentService != nil {
			q.Msg = utils.ConvertMsg2WriteResult(ctx, v.Source.ReactAgentService.Memory.Messages)
			if len(v.Source.ReactAgentService.Rresult) > 0 {
				q.Msg = append(q.Msg, v.Source.ReactAgentService.Rresult...)
			}
			v.Source.ReactAgentService = nil
		}

		q.ChatQuestion = v
		data = append(data, q)
	}

	resp.Response.Data = data
}

// SetTopChatSession 会话置顶
// @Tags ChatSession
// @Summary 会话置顶
// @Description 会话置顶
// @Router /chat.SetTopChatSession [post]
// @Param request body apiobj.DetailIdRequest true "request"
// @Success 200 {object} SetTopChatSessionResponse
func SetTopChatSession(ctx *gin.Context, req *apiobj.DetailIdRequest, resp *SetTopChatSessionResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}

	is_exist, err := chatsession.IsExistSetTopChatSession(ctx, runtime.Uin(ctx), req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "SetTopChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_set_top_failed" // 会话置顶失败
		return
	}
	if is_exist {
		err = chatsession.CancelSetTopChatSession(ctx, runtime.Uin(ctx), req.Request.ID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_cancel_set_top_failed" // 取消置顶会话失败
			return
		}
		return
	}
	err = chatsession.CreateSetTopChatSession(ctx, runtime.Uin(ctx), req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_set_top_failed" // 会话置顶失败
		return
	}
}

// GetSessionInfo 重命名聊天会话
// @Tags ChatSession
// @Summary 重命名聊天会话
// @Description 重命名聊天会话
// @Router /chat.GetSessionInfo [post]
// @Param request body apiobj.DetailIdRequest true "request"
// @Success 200 {object} GetSessionInfoResponse
func GetSessionInfo(ctx *gin.Context, req *apiobj.DetailIdRequest, resp *GetSessionInfoResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	sess, err := chatsession.GetSessionInfo(ctx, req.Request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Code = errcode.ErrCode_NotFound
			resp.Message = "kechat_session_not_found" // 会话未找到
			return
		}
		logs.ErrorContextf(ctx, "GetChatSession error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		return
	}

	resp.Response = sess
}

// WorkflowTestRun 工作流试运行接口
// @Tags workflow
// @Summary 工作流试运行接口
// @Description 工作流试运行接口
// @Router /chat.WorkflowTestRun [post]
// @Param request body WorkflowTestRunRequest true "request"
// @Success 200 {object} WorkflowTestRunResponse
func WorkflowTestRun(ctx *gin.Context, req *WorkflowTestRunRequest, resp *WorkflowTestRunResponse) {
	if req.Request.WorkflowID == "" || req.Request.SpaceID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}

	wf := &coze.Workflowitem{
		WorkflowID: req.Request.WorkflowID,
		SpaceID:    req.Request.SpaceID,
	}
	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
		return
	}
	sessionKey := runtime.LoginStatus(ctx).Token
	err = wf.WorkflowTestRun(ctx, cozeUrl, sessionKey, req.Request.Input)
	if err != nil {
		logs.WarnContextf(ctx, "WorkflowTestRun err %s", err.Error())
		resp.Code = code.ErrCode_HideErrShowMessage
		resp.Message = "工作流暂未成功试运行，请检查编排流程是否配置完整"
		return
	}
	output, code, err := wf.WorkflowGetProcess(ctx, cozeUrl, sessionKey)
	if err != nil {
		logs.ErrorContextf(ctx, "WorkflowGetProcess err %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "workflow_get_process_error"
		return
	}
	if code != 0 {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_coze_workflow_test_run_failed"
		return
	}
	resp.Response.Output = output
}

// UpdateWorkflowVersion 更新工作流版本
// @Tags workflow
// @Summary 更新工作流版本
// @Description 更新工作流版本
// @Router /chat.UpdateWorkflowVersion [post]
// @Param request body UpdateWorkflowVersionRequest true "request"
// @Success 200 {object} apiobj.BaseResponse
func UpdateWorkflowVersion(ctx *gin.Context, req *UpdateWorkflowVersionRequest, resp *apiobj.BaseResponse) {
	if req.Request.WorkflowID == "" || req.Request.Version == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	m, err := chattype.GetCozeMappingByCozeID(ctx, req.Request.WorkflowID, chattype.ChatTypeAgentApp)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCozeMappingByCozeID err %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		return
	}
	item, err := chatagent.GetChatAgentByID(ctx, m.CoreKGID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetChatAgentByID err %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error"
		return
	}
	item.WorkflowVersion = req.Request.Version
	err = chatagent.UpdateWorkflowVsrsion(ctx, item.ID, item.WorkflowVersion)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateCozeMapping err %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error" // 服务器错误
		return
	}
}

// UpdateChatAgentInput 根据coze工作流ID更新入参
// @Tags workflow
// @Summary 根据coze工作流ID更新入参
// @Description 根据coze工作流ID更新入参
// @Router /chat.UpdateChatAgentInput [post]
// @Param request body UpdateChatAgentInputRequest true "request"
// @Success 200 {object} apiobj.BaseResponse
func UpdateChatAgentInput(ctx *gin.Context, req *UpdateChatAgentInputRequest, resp *apiobj.BaseResponse) {
	if req.Request.AgentID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}

	agentInfo, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetChatAgentByID err %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error"
		return
	}

	version, err := chatagent.GetAgentDetailByName(ctx, agentInfo.Name)
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentDetailByName err %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error"
		return
	}
	item := chatagent.UpdateAgentItem{
		ID:              version.ID,
		AgentType:       version.AgentType,
		Description:     version.Description,
		ChatModelIDs:    version.ChatModelIDs,
		Temperature:     version.Temperature,
		PromptTemplate:  version.PromptTemplate,
		Params:          req.Request.Params,
		GreetingMessage: version.GreetingMessage,
	}
	err = chatagent.UpdateChatAgentInput(ctx, &item)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCozeMappingByCozeID err %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_server_error"
		return
	}
}
