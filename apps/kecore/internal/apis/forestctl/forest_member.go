package forestctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ListForestPublicScope
// Deprecated:api route,this func was replaced with forestctl.GetForest
// @Tags 知识森林
// @Summary 获取知识森林公开范围
// @Description 获取知识森林公开范围列表
// @Router /forest.ListForestPublicScope [post]
// @Param user body ListForestPublicScopeRequest true "入参"
// @Success 200 {object} ListForestPublicScopeResponse "返回值"
func ListForestPublicScope(ctx *gin.Context, req *ListForestPublicScopeRequest, resp *ListForestPublicScopeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "ListForestPublicScope validate params failed")
		return
	}
	_ = runtime.Uin(ctx)

	// 获取知识森林信息
	_, err := forest.GetForestByID(ctx, req.Request.ForestId)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed_data" // 获取知识森林信息失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "ListForestPublicScope GetPromptAgentDetail failed ,err %s", err)
		return
	}

	//// 校验权限
	//if !forestInfo.ManagerIDs.Contains(uin) {
	//	resp.Code = errcode.ErrCode_InternalError
	//	resp.Message = "没有权限查看此知识森林"
	//	logs.ErrorContextf(ctx, "ListForestPublicScope Permission denied, uin:%d is neither owner nor manager of forest:%d",
	//		uin, req.Request.ForestId)
	//	return
	//}

	err = forest.QueryResourceScopeList(ctx, req.Request.PageQuery, req.Request.ForestId, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_scope_failed_data" // 获取知识森林列表失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "ListForestPublicScope QueryResourceScopeList failed ,err %s", err)
		return
	}
}

// UpdateForestPublicScope
// Deprecated:api route,this was replaced with accountctl.UpdateForestWithPerm
// @Tags 知识森林
// @Summary 全量更新知识森林公开范围
// @Description 覆盖更新知识森林公开范围，只保留传入的 ScopeIDs
// @Router /forest.UpdateForestPublicScope [post]
// @Param user body UpdateForestPublicScopeRequest true "入参"
// @Success 200 {object} UpdateForestPublicScopeResponse "返回值"
func UpdateForestPublicScope(ctx *gin.Context, req *UpdateForestPublicScopeRequest, resp *UpdateForestPublicScopeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "UpdateForestPublicScope validate params failed")
		return
	}
	_ = runtime.Uin(ctx)

	// 权限校验
	_, err := forest.GetForestByID(ctx, req.Request.ForestId)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed_data" // 获取知识森林信息失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "GetPromptAgentDetail failed, err %s", err)
		return
	}
	//if !forestInfo.ManagerIDs.Contains(uin) {
	//	resp.Code = errcode.ErrCode_InternalError
	//	resp.Message = "没有权限修改此知识森林的公开范围"
	//	logs.ErrorContextf(ctx, "Permission denied, uin:%d is not manager of forest:%d", uin, req.Request.ForestId)
	//	return
	//}
	// 覆盖更新公开范围
	err = forest.CoverForestPublicScope(req.Request.ForestId, req.Request.ScopeType, req.Request.ScopeIDs)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_forest_scope_failed_data" // 更新公开范围失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "CoverForestPublicScope failed, err: %s", err)
		return
	}
}

// UpdateForestManager
// Deprecated:api route,this func was replaced with accountctl.UpdateForestWithPerm
// @Tags 知识森林
// @Summary 更新知识森林管理
// @Description 更新知识森林公开范围和管理员
// @Router /forest.UpdateForestManager [post]
// @Param user body UpdateForestManagerRequest true "入参"
// @Success 200 {object} UpdateForestManagerResponse "返回值"
func UpdateForestManager(ctx *gin.Context, req *UpdateForestManagerRequest, resp *UpdateForestManagerResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "UpdateForestManager validate params failed")
		return
	}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	// 权限校验
	_, err := forest.GetForestByID(ctx, req.Request.ForestId)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed_data" // 获取知识森林信息失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "GetPromptAgentDetail failed, err %s", err)
		return
	}
	//if !forestInfo.ManagerIDs.Contains(uin) {
	//	resp.Code = errcode.ErrCode_InternalError
	//	resp.Message = "没有权限修改此知识森林的公开范围"
	//	logs.ErrorContextf(ctx, "Permission denied, uin:%d is not manager of forest:%d", uin, req.Request.ForestId)
	//	return
	//}

	req.Request.ManagerIDs.RemoveDuplicates()

	// 验证所有管理员用户
	for _, managerID := range req.Request.ManagerIDs.Slice() {
		if managerID == uin {
			continue
		}
		// 验证管理员用户是否存在且属于同一公司
		_, err := employee.GetEmployeeByUinAndCompanyID(managerID, companyID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_invalid_manager_data" // 管理员用户 %d 不存在或不属于当前公司
			resp.MessageData = map[string]interface{}{
				"manager_id": managerID,
			}
			logs.ErrorContextf(ctx, "Invalid manager %d in CreatePromptAgent, uin:%d, err:%v",
				managerID, uin, err)
			return
		}
	}

	// 更新公开范围和管理员
	err = forest.UpdateForestManager(req.Request.ForestId, req.Request.PublicScope, req.Request.ManagerIDs)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_forest_manager_failed_data" // 更新公开范围和管理员失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "UpdateForestManager failed, err: %s", err)
		return
	}

}
