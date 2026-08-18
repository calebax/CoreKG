package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListAgentVersion Agent版本变更记录列表
// @Tags Agent版本管理
// @Summary Agent版本变更记录列表
// @Description Agent版本变更记录列表
// @Router /chat.ListAgentVersion [post]
// @Param request body ListAgentVersionRequest true "入参"
// @Success 200 {object} ListAgentVersionResponse
func ListAgentVersion(ctx *gin.Context, req *ListAgentVersionRequest, resp *ListAgentVersionResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListAgentVersionHistory validate params failed")
		return
	}
	// 获取机器人基本信息
	_, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_info_failed" // 获取信息失败
		logs.ErrorContextf(ctx, "GetRolePlayAgentDetail failed ,err %s", err)
		return
	}
	err = chatagent.QueryAgentVersionList(ctx, req.Request.PageQuery, req.Request.AgentID, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_version_list_failed" // 获取版本列表失败
		logs.ErrorContextf(ctx, "ListSettingGroup failed ,err %s", err)
		return
	}
}

// ChooseAgentVersion 选择机器人版本
// @Tags Agent版本管理
// @Summary 选择机器人版本
// @Description 选择机器人版本
// @Router /chat.ChooseAgentVersion [post]
// @Param request body ChooseAgentVersionRequest true "入参"
// @Success 200 {object} ChooseAgentVersionResponse
func ChooseAgentVersion(ctx *gin.Context, req *ChooseAgentVersionRequest, resp *ChooseAgentVersionResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "ChooseAgentVersion validate params failed")
		return
	}
	// 获取机器人信息
	_, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_info_failed" // 获取信息失败
		logs.ErrorContextf(ctx, "GetRolePlayAgentDetail failed ,err %s", err)
		return
	}

	// 获取版本信息
	_, err = chatagent.GetChatAgentVersionByID(ctx, req.Request.AgentID, req.Request.VersionID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_version_info_failed" // 获取版本信息失败
		logs.ErrorContextf(ctx, "GetRolePlayAgentDetail failed ,err %s", err)
		return
	}
	// 选择版本
	err = chatagent.ChooseAgentVersion(ctx, req.Request.AgentID, req.Request.VersionID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_choose_version_failed" // 选择版本失败
		logs.ErrorContextf(ctx, "GetRolePlayAgentDetail failed ,err %s", err)
		return
	}
}
