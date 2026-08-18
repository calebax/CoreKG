package companyctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/company"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateCompanyEmployee
// @Tags Admin团队管理
// @Summary 创建团队成员
// @Description 创建团队成员
// @Router /admin.CreateCompanyEmployee [post]
// @Param user body CreateCompanyEmployeeRequest true "入参"
// @Success 200 {object} CreateCompanyEmployeeResponse "返回值"
func CreateCompanyEmployee(ctx *gin.Context, req *CreateCompanyEmployeeRequest, resp *CreateCompanyEmployeeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	exist, err := req.Request.IsExist()
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateCompanyEmployee]: IsExist failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询团队成员失败"
		return
	}
	if exist {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "团队成员已存在"
		return
	}
	ret, err := company.CreateCompanyEmployee(ctx, &req.Request)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateCompanyEmployee]: CreateCompanyEmployee failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建团队成员失败"
		return
	}
	resp.Response.Employee = *ret
}

// ListCompanyEmployee 团队成员列表
// @Tags Admin团队管理
// @Summary 团队成员列表
// @Description 团队成员列表
// @Router /admin.ListCompanyEmployee [post]
// @Param user body ListCompanyEmployeeRequest true "入参"
// @Success 200 {object} ListCompanyEmployeeResponse "返回值"
func ListCompanyEmployee(ctx *gin.Context, req *ListCompanyEmployeeRequest, resp *ListCompanyEmployeeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	err := company.QueryCompanyEmployeeList(ctx, &req.Request, &resp.Response)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListCompanyEmployee]: QueryCompanyEmployeeList failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询团队成员失败"
		return
	}
}

// UpdateCompanyEmployeeRole 更换成员角色
// @Tags Admin团队管理
// @Summary 团队成员列表
// @Description 团队成员列表
// @Router /admin.UpdateCompanyEmployeeRole [post]
// @Param user body UpdateCompanyEmployeeRoleRequest true "入参"
// @Success 200 {object} UpdateCompanyEmployeeRoleResponse "返回值"
func UpdateCompanyEmployeeRole(ctx *gin.Context, req *UpdateCompanyEmployeeRoleRequest, resp *UpdateCompanyEmployeeRoleResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	err := company.UpdateCompanyEmployeeRole(req.Request.ID, req.Request.Role)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListCompanyEmployee]: QueryCompanyEmployeeList failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询团队成员失败"
		return
	}
}
