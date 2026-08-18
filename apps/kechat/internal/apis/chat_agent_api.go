package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// ListChatAgent 机器人列表
// @Tags 机器人
// @Summary 机器人列表
// @Description 机器人列表
// @Router /chat.ListChatAgent [post]
// @Param request body ListChatAgentRequest true "入参"
// @Success 200 {object} ListChatAgentResponse
func ListChatAgent(ctx *gin.Context, req *ListChatAgentRequest, resp *ListChatAgentResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListChatAgent validate params failed")
		return
	}
	req.Request.Uin = runtime.Uin(ctx)
	req.Request.CompanyID = runtime.CompanyID(ctx)
	err := chatagent.QueryChatAgentList(ctx, req.Request, &resp.Response)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_agent_list_failed")) // 获取机器人列表失败
		logs.ErrorContextf(ctx, "ListChatAgent failed ,err %s", err)
		return
	}
	cozeResp, err := coze.GetCozeAgent(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "get coze soutce type failed, err %s", err)
		return
	}
	items, err := chattype.GetCozeMappingByID(ctx, req.Request.Uin, chattype.ChatTypeAgentApp)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCozeMappingByID error, %s", err.Error())
		return
	}
	corekgMap := make(map[uint]int)
	for i, s := range resp.Response.Data {
		corekgMap[s.ID] = i
		resp.Response.Data[i].IsSynced = false
	}
	mappingMap := make(map[string]uint)
	for _, mapping := range items {
		mappingMap[mapping.CozeID] = mapping.CoreKGID
	}
	for _, s := range cozeResp.ResourceList {
		if cozeId, ok := mappingMap[s.ResID]; ok {
			if v, ok := corekgMap[cozeId]; ok {
				resp.Response.Data[v].IsSynced = true
			}
		}
	}
}

// UpdateChatAgent 更新机器人
// Deprecated:this api func was deprecated try to use perm.UpdateAgentWithPerm to instead it
// @Tags 机器人
// @Summary 更新机器人
// @Description 更新机器人
// @Router /chat.UpdateChatAgent [post]
// @Param request body UpdateChatAgentRequest true "入参"
// @Success 200 {object} UpdateChatAgentResponse
func UpdateChatAgent(ctx *gin.Context, req *UpdateChatAgentRequest, resp *UpdateChatAgentResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "UpdateChatAgent validate params failed")
		return
	}
	// 获取机器人基本信息
	agentInfo, err := chatagent.GetChatAgentByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_info_failed" // 获取机器人信息失败
		logs.ErrorContextf(ctx, "ListSettingGroup failed ,err %s", err)
		return
	}
	//// 校验权限
	//if !agentInfo.ManagerIDs.Contains(uin) {
	//	resp.Code = errcode.ErrCode_InternalError
	//	resp.Message = "没有权限编辑此机器人"
	//	logs.Warnf("Permission denied, uin:%d is neither owner nor manager of agent:%d",
	//		uin, req.Request.ID)
	//	return
	//}
	// 更新机器人
	err = chatagent.UpdateChatAgent(ctx, agentInfo.ID, req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_agent_failed" // 更新机器人失败
		logs.ErrorContextf(ctx, "UpdateChatAgent failed ,err %s", err)
		return
	}
}

// DeleteChatAgent 删除机器人
// @Tags 机器人
// @Summary 删除机器人
// @Description 删除机器人
// @Router /chat.DeleteChatAgent [post]
// @Param request body DeleteChatAgentRequest true "入参"
// @Success 200 {object} DeleteChatAgentResponse
func DeleteChatAgent(ctx *gin.Context, req *DeleteChatAgentRequest, resp *DeleteChatAgentResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "DeleteChatAgent validate params failed")
		return
	}
	uin := runtime.Uin(ctx)
	// 获取机器人基本信息
	agentInfo, err := chatagent.GetChatAgentByID(ctx, req.Request.ID)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_agent_info_failed")) // 获取机器人信息失败
		logs.ErrorContextf(ctx, "GetChatAgentByID(%v) failed ,err %s", req.Request.ID, err)
		return
	}
	if !perm.HasManageAct(ctx, uin, agentInfo.ID, foresttype.ResourceTypeAgent) {
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_no_permission")) // 无权限更新此资源
		return
	}

	// 删除机器人
	if err = chatagent.DeleteChatAgent(ctx, req.Request.ID); err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_delete_agent_failed")) // 删除机器人失败
		logs.ErrorContextf(ctx, "DeleteChatAgent(%v) failed ,err %s", req.Request.ID, err)
		return
	}
	cozeMapping, err := chattype.GetCozeMappingByCoreKGID(ctx, req.Request.ID)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_coze_mapping_failed")) // 获取映射信息失败
		logs.ErrorContextf(ctx, "GetCozeMappingByCoreKGID(%v) failed ,err %s", req.Request.ID, err)
		return
	}
	token := runtime.LoginStatus(ctx).Token
	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
		return
	}
	if err = chattype.DeleteCozeMappingByCorekgID(ctx, req.Request.ID); err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_delete_mapping_failed")) // 删除映射失败
		logs.ErrorContextf(ctx, "DeleteChatAgentMapping(%v) failed ,err %s", req.Request.ID, err)
		return
	}
	for _, m := range cozeMapping {
		if m.Type == chattype.ChatTypeAgentApp {
			if err = coze.DeleteCozePluginAPI(ctx, m.CozeID, token, cozeUrl); err != nil {
				runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_delete_coze_plugin_failed")) // 删除coze数据失败
				logs.ErrorContextf(ctx, "DeleteCozePluginAPI(%v) failed ,err %s", req.Request.ID, err)
			}

		} else if m.Type == chattype.ChatTypeWorkflow {
			spaceID, code, err := coze.GetSpaceAPI(ctx, cozeUrl, token)
			if err != nil || code != 0 {
				runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_coze_space_failed")) // 删除coze工作流失败
				logs.ErrorContextf(ctx, "GetSpaceAPI(%v) failed ,err %s", req.Request.ID, err)
				return
			}
			if err = coze.DeleteCozeWorkflow(ctx, m.CozeID, spaceID, token, cozeUrl); err != nil {
				runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_delete_coze_workflow_failed")) // 删除coze工作流失败
				logs.ErrorContextf(ctx, "delete coze workflow(%v) failed ,err %s", req.Request.ID, err)
				return
			}
		}
	}

}

// GetAgentInfo 获取指令型机器人详情
// @Tags 机器人
// @Summary 获取指令型机器人详情
// @Description 获取指令型机器人详情
// @Router /chat.GetAgentInfo [post]
// @Param request body GetAgentInfoRequest true "入参"
// @Success 200 {object} GetAgentInfoResponse
func GetAgentInfo(ctx *gin.Context, req *GetAgentInfoRequest, resp *GetAgentInfoResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "GetAgentInfo validate params failed")
		return
	}

	// 获取机器人信息
	agentInfo, err := chatagent.GetAgentDetail(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_detail_failed" // 获取基本信息失败
		logs.ErrorContextf(ctx, "GetAgentInfo failed ,err %s", err)
		return
	}
	resp.Response.AgentInfo = agentInfo
}

// CreateAgentApp 创建应用
// @Tags 应用管理
// @Summary 创建应用
// @Description 创建应用
// @Router /chat.CreateAgentApp [post]
// @Param user body CreateAgentAppRequest true "入参"
// @Success 200 {object} CreateAgentAppResponse "返回值"
func CreateAgentApp(ctx *gin.Context, req *CreateAgentAppRequest, resp *CreateAgentAppResponse) {
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	agent, err := chatagent.CreateAgentApp(ctx, uin, companyID, req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_app_failed" // 创建应用失败
		logs.ErrorContextf(ctx, "CreateAgentApp failed ,err %s", err)
		return
	}
	resp.Response.ID = agent.ID
	resp.Response.SpaceID = agent.CozeSpaceID
	resp.Response.WorkflowID = agent.CozeWorkflowID
}

// ListCollectApp 我的收藏列表
// @Tags 收藏应用
// @Summary 我的收藏列表
// @Description 我的收藏列表
// @Router /chat.ListCollectApp [post]
// @Param user body ListCollectAppRequest true "入参"
// @Success 200 {object} ListCollectAppResponse "返回值"
func ListCollectApp(ctx *gin.Context, req *ListCollectAppRequest, resp *ListCollectAppResponse) {
	uin := runtime.Uin(ctx)
	req.Request.Uin = uin
	err := chatagent.QueryAgentCollectList(ctx, req.Request, &resp.Response)
	if err != nil {
		logs.ErrorContextf(ctx, "ListCollectApp failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_collect_list_failed" // 获取收藏应用列表失败
		logs.ErrorContextf(ctx, "ListCollectApp failed ,err %s", err)
		return
	}
}

// CollectApp 收藏应用
// @Tags 收藏应用
// @Summary 收藏应用
// @Description 收藏应用
// @Router /chat.CollectApp [post]
// @Param user body CollectAppRequest true "入参"
// @Success 200 {object} CollectAppResponse "返回值"
func CollectApp(ctx *gin.Context, req *CollectAppRequest, resp *CollectAppResponse) {
	uin := runtime.Uin(ctx)
	isExist, err := chatagent.IsExistAgentCollect(ctx, uin, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "CollectApp IsExistAgentCollect failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_collect_app_failed" // 收藏应用失败
		logs.ErrorContextf(ctx, "CollectApp IsExistAgentCollect failed ,err %s", err)
		return
	}
	if isExist {
		err = chatagent.DeleteAgentCollect(ctx, uin, req.Request.ID)
		if err != nil {
			logs.ErrorContextf(ctx, "CollectApp DeleteAgentCollect failed ,err %s", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_cancel_collect_failed" // 取消收藏应用失败
			logs.ErrorContextf(ctx, "CollectApp DeleteAgentCollect failed ,err %s", err)
			return
		}
		return
	}
	err = chatagent.CreateAgentCollect(ctx, uin, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "CollectApp CreateAgentCollect failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_collect_app_failed" // 收藏应用失败
		logs.ErrorContextf(ctx, "CollectApp CreateAgentCollect failed ,err %s", err)
		return
	}
}
