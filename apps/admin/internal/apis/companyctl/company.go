package companyctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/company"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateCompany 创建团队
// @Tags Admin团队管理
// @Summary 创建团队
// @Description 创建团队
// @Router /admin.CreateCompany [post]
// @Param user body CreateCompanyRequest true "入参"
// @Success 200 {object} CreateCompanyResponse "返回值"
func CreateCompany(ctx *gin.Context, req *CreateCompanyRequest, resp *CreateCompanyResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		return
	}
	exist, err := company.IsExistCompanyByName(req.Request.Name)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建团队失败"
		return
	}
	if exist {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "团队名称已存在"
		return
	}
	c, err := company.CreateCompany(ctx, dbutil.Account(), &req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建团队失败"
		return
	}
	resp.Response.Company = *c
}

// ListCompany 团队列表
// @Tags Admin团队管理
// @Summary 团队列表
// @Description 团队列表
// @Router /admin.ListCompany [post]
// @Param user body ListCompanyRequest true "入参"
// @Success 200 {object} ListCompanyResponse "返回值"
func ListCompany(ctx *gin.Context, req *ListCompanyRequest, resp *ListCompanyResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		return
	}
	err := company.QueryCompanyList(ctx, &req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询团队列表失败"
		return
	}
}

// ModifyCompany 修改团队
// @Tags Admin团队管理
// @Summary 修改团队
// @Description 修改团队
// @Router /admin.ModifyCompany [post]
// @Param user body ModifyCompanyRequest true "入参"
// @Success 200 {object} ModifyCompanyResponse "返回值"
func ModifyCompany(ctx *gin.Context, req *ModifyCompanyRequest, resp *ModifyCompanyResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		return
	}

	exist, err := company.IsExistCompanyByName(req.Request.Name, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "团队名称已存在"
		return
	}
	if exist {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "团队名称已存在"
		return
	}

	c, err := company.ModifyCompany(req.Request.ID, &req.Request.ModifyCompanyOption)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "修改团队失败"
		return
	}
	resp.Response.Company = *c
}
