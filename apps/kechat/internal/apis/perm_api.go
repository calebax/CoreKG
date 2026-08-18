package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kechat/models/agentperm"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// GetAgentPermSet 获取用户轻应用权限集
// @Tags 账户系统
// @Summary 获取用户权限集
// @Description 获取用户权限集
// @Router /chat.GetAgentPermSet [post]
// @Param user body GetAgentPermSetRequest true "入参"
// @Success 200 {object} GetAgentPermSetResponse "返回值"
func GetAgentPermSet(ctx *gin.Context, req *GetAgentPermSetRequest, resp *GetAgentPermSetResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "[GetAgentPermSet] validate params failed")
		return
	}
	//*get uin about do this action
	actUin := runtime.Uin(ctx)

	actEmp, err := employee.GetEmployeeByUin(actUin)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAgentPermSet] GetEmployeeByUin[uin=%v] failed ,err %v", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_operator_info_failed" // 获取操作者员工信息失败
		return
	}

	if actEmp.SysRole != accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[GetAgentPermSet] [uin=%v,sysRole=%v] failed", actUin, actEmp.SysRole)
		resp.Code = errcode.ErrCode_Unauthorized
		resp.Message = "kechat_no_permission" // 无操作权限
		return
	}

	ags, err := agentperm.GetAgentByCompanyID(ctx, actEmp.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAgentPermSet] [GetAgentsByCompanyID] failed ,err %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_company_forest_failed" // 获取公司森林失败
		return
	}

	if req.Request.Uin == 0 {
		resp.Response.PermSet = make([]*agentperm.PermSet, 0, len(ags))
		for _, f := range ags {
			resp.Response.PermSet = append(resp.Response.PermSet, &agentperm.PermSet{Agent: f, ManagePerm: false, UsePerm: false})
		}
		return
	}

	emp, err := employee.GetEmployeeByUin(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] GetEmployeeByUin[uin=%v] failed ,err %v", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_employee_info_failed" // 获取员工信息失败
		return
	}
	if emp.CompanyID != actEmp.CompanyID {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] Incorrect emp[actuin=%v/targuin=%v]", actUin, emp.Uin)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_unknown_company_user" // 未知公司用户无权限
		return
	}

	wps := agentperm.NewWrapperPermSet(ctx, req.Request.Uin, emp.CompanyID, nil, ags)
	if err = wps.BuildCurrentPermMap(); err != nil {
		logs.ErrorContextf(ctx, "[GetAgentPermSet] [BuildCurrentPermMap] failed ,err %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_build_agent_perm_failed" // 构建机器人权限集失败
		return
	}
	resp.Response.PermSet = wps.GetCurrPermSet()
}

// ModifyChatPermSet 更新用户机器人权限集
// @Tags 账户系统
// @Summary 更新用户机器人权限集
// @Description 更新用户机器人权限集
// @Router /forest.ModifyChatPermSet [post]
// @Param user body ModifyChatPermSetRequest true "入参"
// @Success 200 {object} ModifyChatPermSetResponse "返回值"
func ModifyChatPermSet(ctx *gin.Context, req *ModifyChatPermSetRequest, resp *ModifyChatPermSetResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "[ModifyChatPermSet] validate params failed")
		return
	}
	//*get uin about do this action
	actUin := runtime.Uin(ctx)

	actEmp, err := employee.GetEmployeeByUin(actUin)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] GetEmployeeByUin[uin=%v] failed ,err %v", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_operator_info_failed" // 获取操作者员工信息失败
		return
	}

	if actEmp.SysRole != accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] [uin=%v,sysRole=%v] failed", actUin, actEmp.SysRole)
		resp.Code = errcode.ErrCode_Unauthorized
		resp.Message = "kechat_no_permission" // 无操作权限
		return
	}

	emp, err := employee.GetEmployeeByUin(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] GetEmployeeByUin[uin=%v] failed ,err %v", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_employee_info_failed" // 获取员工信息失败
		return
	}
	if emp.CompanyID != actEmp.CompanyID {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] Incorrect emp[actuin=%v/targuin=%v]", actUin, emp.Uin)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_unknown_company_user" // 未知公司用户无权限
		return
	}

	//if actUin == req.Request.Uin {
	//	//* if the user is an admin (himself/herself/itself?),
	//	//* then this action will be no mean
	//	return
	//}

	wps := agentperm.NewWrapperPermSet(ctx, req.Request.Uin, emp.CompanyID, req.Request.PermSet, nil)
	if err = wps.BuildCurrentPermMap(); err != nil {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] BuildCurrentPermMap failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_build_user_agent_perm_failed" // 构建用户机器人权限集失败
		return
	}

	if err = wps.ApplyChanges(); err != nil {
		logs.ErrorContextf(ctx, "[ModifyChatPermSet] ApplyChanges failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_permission_change_failed" // 权限变更失败
		return
	}
}

// GetAgentWithPerm 获取机器人详情(新权限)
// @Tags 机器人
// @Summary 获取机器人详情(新权限)
// @Description 获取机器人详情(新权限)
// @Router /chat.GetAgentWithPerm [post]
// @Param request body apiobj.DetailIdRequest true "入参"
// @Success 200 {object} GetAgentInfoWithPermResponse
func GetAgentWithPerm(ctx *gin.Context, req *apiobj.DetailIdRequest, resp *GetAgentInfoWithPermResponse) {
	if req.Request.ID <= 0 {
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_invalid_id")) // 非法id
		return
	}
	// 获取机器人信息
	agentInfo, err := chatagent.GetChatAgentWithPermByID(ctx, req.Request.ID)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_agent_info_failed")) // 获取机器人基础信息失败
		logs.ErrorContextf(ctx, "GetChatAgentWithPermByID(%v) failed ,err %s", req.Request.ID, err)
		return
	}
	// 获取机器人类型额外信息
	promptAgent, err := chatagent.GetAgentTypeByID(ctx, agentInfo.Version)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_agent_type_failed")) // 获取机器人类型信息失败
		logs.ErrorContextf(ctx, "GetAgentTypeByID(%v) failed ,err %s", agentInfo.Version, err)
		return
	}
	if agentInfo.AgentType == chattype.AgentTypeWorkflow {
		workflowItem := coze.Workflowitem{
			WorkflowID: agentInfo.CozeWorkflowID,
			SpaceID:    agentInfo.CozeSpaceID,
			ExecuteID:  "",
		}
		cozeUrl, err := settings.GetText("corekg", "coze_url")
		if err != nil {
			logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
			return
		}
		sessionKey := runtime.LoginStatus(ctx).Token
		wf, err := workflowItem.GetWorkflowCanvas(ctx, cozeUrl, sessionKey)
		if err != nil {
			logs.ErrorContextf(ctx, "GetWorkflowCanvas failed ,err %s", err.Error())
			return
		}
		var pam chattype.ParamsList
		for _, workflowField := range wf {
			pa := chattype.Params{
				Name:       workflowField.Name,
				Input:      workflowField.Name,
				InputType:  chattype.InputTypeText,
				IsRequired: workflowField.Required,
			}
			pam = append(pam, pa)
		}
		promptAgent.Params = pam
	}

	resp.Response.AgentWithPerm = *agentInfo
	resp.Response.AgentItemInfo = *promptAgent
}

// UpdateAgentWithPerm 编辑指令型机器人(新权限)
// @Tags 机器人
// @Summary 编辑指令型机器人(新权限)
// @Description 编辑指令型机器人(新权限)
// @Router /chat.UpdateAgentWithPerm [post]
// @Param request body UpdateAgentRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func UpdateAgentWithPerm(ctx *gin.Context, req *UpdateAgentRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "UpdateCommandAgent validate params failed")
		return
	}

	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	agent, err := chatagent.GetChatAgentByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_info_failed" // 获取机器人信息失败
		logs.WarnContextf(ctx, "GetChatAgentByID failed, err %s", err)
		return
	}

	if !perm.HasManageAct(ctx, uin, agent.ID, foresttype.ResourceTypeAgent) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, agent.ID)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_no_permission_to_update")) // 无权限更新此资源
		return
	}

	//do not append manager action list with admins in company
	if req.Request.PublicScope == chattype.PublicScopeCompany {
		req.Request.ScopeIDs = types.NewUintArray([]uint{})
	}

	// 去重
	req.Request.ManagerIDs.RemoveDuplicates()
	req.Request.ScopeIDs.RemoveDuplicates()

	uins := types.NewUintArray(append(req.Request.ManagerIDs.Slice(), req.Request.ScopeIDs.Slice()...))
	uins.RemoveDuplicates()

	us := uins.Slice()
	if !employee.CheckUinsValid(ctx, us, companyID) {
		logs.ErrorContextf(ctx, "CheckUinsValid: exist no-local company[%v] uin in uins[%v]", companyID, us)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_invalid_employee_id")) // 存在非法员工id
		return
	}

	req.Request.CompanyID = companyID
	if err = chatagent.UpdateAgentWithPerm(ctx, &req.Request); err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_update_agent_failed")) // 更新机器人失败
		logs.WarnContextf(ctx, "UpdateCommandAgent failed, err %s", err)
		return
	}
	cozeMapping, err := chattype.GetCozeMappingByAgentID(ctx, agent.ID, chattype.ChatTypeWorkflow)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCozeMappingByAgentID error, %s", err.Error())
		return
	}
	switch cozeMapping.CozeID {
	case "":
	default:
		cozeUrl, err := settings.GetText("corekg", "coze_url")
		if err != nil {
			logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
			return
		}
		sessionKey := runtime.LoginStatus(ctx).Token
		wf := coze.Workflowitem{
			WorkflowID: agent.CozeWorkflowID,
			SpaceID:    agent.CozeSpaceID,
		}
		success, err := wf.WorkflowPublish(ctx, cozeUrl, sessionKey, cozeMapping.CozeID)
		if err != nil {
			logs.ErrorContextf(ctx, "WorkflowPublish failed ,err %s", err.Error())
			return
		}
		if !success {
			logs.ErrorContextf(ctx, "WorkflowPublish failed ,success %v", success)
			return
		}
		if err := chatagent.UpdateWorkflowVsrsion(ctx, agent.ID, wf.Version); err != nil {
			logs.ErrorContextf(ctx, "UpdateWorkflowVsrsion failed ,err %s", err.Error())
			return
		}
	}
}
