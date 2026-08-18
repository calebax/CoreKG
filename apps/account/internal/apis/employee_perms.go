package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ListPosition 职位列表
// @Tags 运营职位权限
// @Summary 职位列表
// @Description 职位列表
// @Router /account.ListPosition [post]
// @Param user body ListPositionRequest true "入参"
// @Success 200 {object} ListPositionResponse "返回值"
func ListPosition(ctx *gin.Context, req *ListPositionRequest, resp *ListPositionResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	err := employee.QueryPositionList(ctx, req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_list_position_failed" // 获取职位列表失败
		return
	}
}

// CreatePosition 添加职位
// @Tags 运营职位权限
// @Summary 添加职位
// @Description 添加职位
// @Router /account.CreatePosition [post]
// @Param user body CreatePositionRequest true "入参"
// @Success 200 {object} CreatePositionResponse "返回值"
func CreatePosition(ctx *gin.Context, req *CreatePositionRequest, resp *CreatePositionResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreatePosition validate params failed")
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	pos, err := employee.CreatePosition(&req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_position_failed" // 添加职位失败
		logs.WarnContextf(ctx, "CreatePosition failed ,err %s", err)
		return
	}

	resp.Response = pos
}

// GetPositionDetail 职位详情
// @Tags 运营职位权限
// @Summary 职位详情
// @Description 职位详情
// @Router /account.GetPositionDetail [post]
// @Param user body GetPositionDetailRequest true "入参"
// @Success 200 {object} GetPositionDetailResponse "返回值"
func GetPositionDetail(ctx *gin.Context, req *GetPositionDetailRequest, resp *GetPositionDetailResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	detail, err := employee.GetPositionDetailByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_position_detail_failed" // 获取职位详情失败
		return
	}
	resp.Response = *detail
}

// ModifyPositionInfo 修改职位信息
// @Tags 运营职位权限
// @Summary 修改职位基本信息
// @Description 修改职位基本信息
// @Router /account.ModifyPositionInfo [post]
// @Param user body ModifyPositionInfoRequest true "入参"
// @Success 200 {object} ModifyPositionInfoResponse "返回值"
func ModifyPositionInfo(ctx *gin.Context, req *ModifyPositionInfoRequest, resp *ModifyPositionInfoResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ModifyPositionInfo validate params failed")
		return
	}
	if req.Request.CompanyID == 0 || req.Request.CompanyID != runtime.CompanyID(ctx) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_modify_position_invalid_company" // 只能修改自己公司的权限
		return
	}

	position, err := employee.ModifyPosition(req.Request.ID, &req.Request.Position)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_modify_position_failed" // 编辑职位失败
		logs.WarnContextf(ctx, "ModifyPositionInfo failed ,err %s", err)
		return
	}

	resp.Response = *position
}

// ModifyPositionPrivilege 修改职位信息
// @Tags 运营职位权限
// @Summary 修改职位权限信息
// @Description 修改职位权限信息
// @Router /account.ModifyPositionPrivilege [post]
// @Param user body ModifyPositionPrivilegeRequest true "入参"
// @Success 200 {object} ModifyPositionPrivilegeResponse "返回值"
func ModifyPositionPrivilege(ctx *gin.Context, req *ModifyPositionPrivilegeRequest, resp *ModifyPositionPrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ModifyPositionPrivilege validate params failed")
		return
	}

	position, err := employee.ModifyPositionPrivilege(req.Request.ID, req.Request.PrivilegeIDs)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_modify_position_privilege_failed" // 编辑职位权限失败
		logs.WarnContextf(ctx, "ModifyPositionPrivilege failed ,err %s", err)
		return
	}

	resp.Response = *position
}

// DeletePosition 删除职位
// @Tags 运营职位权限
// @Summary 删除职位
// @Description 删除职位
// @Router /account.DeletePosition [post]
// @Param user body DeletePositionRequest true "入参"
// @Success 200 {object} DeletePositionResponse "返回值"
func DeletePosition(ctx *gin.Context, req *DeletePositionRequest, resp *DeletePositionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "DeletePosition validate params failed")
		return
	}
	err := employee.DeletePosition(req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_delete_position_failed" // 删除职位失败
		logs.WarnContextf(ctx, "DeletePosition failed ,err %s", err)
		return
	}
}

// ListPrivilege 权限列表
// @Tags 运营权限
// @Summary 权限列表
// @Description 权限列表
// @Router /account.ListPrivilege [post]
// @Param user body ListPrivilegeRequest true "入参"
// @Success 200 {object} ListPrivilegeResponse "返回值"
func ListPrivilege(ctx *gin.Context, req *ListPrivilegeRequest, resp *ListPrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)

	err := employee.QueryPrivilegeList(ctx, req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_list_privilege_failed" // 获取权限列表失败
		return
	}
}

// CreatePrivilege 添加权限
// @Tags 运营权限
// @Summary 添加权限
// @Description 添加权限
// @Router /account.CreatePrivilege [post]
// @Param user body CreatePrivilegeRequest true "入参"
// @Success 200 {object} CreatePrivilegeResponse "返回值"
func CreatePrivilege(ctx *gin.Context, req *CreatePrivilegeRequest, resp *CreatePrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreatePrivilege validate params failed")
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	pos, err := employee.CreatePrivilege(&req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_privilege_failed" // 添加权限失败
		logs.WarnContextf(ctx, "CreatePrivilege failed ,err %s", err)
		return
	}

	resp.Response = pos
}

// ModifyPrivilege 修改权限信息
// @Tags 运营权限
// @Summary 修改权限信息
// @Description 修改权限信息
// @Router /account.ModifyPrivilege [post]
// @Param user body ModifyPrivilegeRequest true "入参"
// @Success 200 {object} ModifyPrivilegeResponse "返回值"
func ModifyPrivilege(ctx *gin.Context, req *ModifyPrivilegeRequest, resp *ModifyPrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ModifyPrivilegeInfo validate params failed")
		return
	}
	if req.Request.CompanyID == 0 || req.Request.CompanyID != runtime.CompanyID(ctx) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_modify_privilege_invalid_company" // 只能修改自己公司的权限
		return
	}

	err := employee.ModifyPrivilege(&req.Request.Privilege)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_modify_privilege_failed" // 编辑权限失败
		logs.WarnContextf(ctx, "ModifyPrivilegeInfo failed ,err %s", err)
		return
	}
}

// DeletePrivilege 删除权限
// @Tags 运营权限
// @Summary 删除权限
// @Description 删除权限
// @Router /account.DeletePrivilege [post]
// @Param user body DeletePrivilegeRequest true "入参"
// @Success 200 {object} DeletePrivilegeResponse "返回值"
func DeletePrivilege(ctx *gin.Context, req *DeletePrivilegeRequest, resp *DeletePrivilegeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "DeletePrivilege validate params failed")
		return
	}
	err := employee.DeletePrivilege(req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_delete_privilege_failed" // 删除权限失败
		logs.WarnContextf(ctx, "DeletePrivilege failed ,err %s", err)
		return
	}
}
