package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtoorganize"
	"github.com/insmtx/corekg/apps/account/services/svcorganize"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateDepartment 创建部门
// @Tags 组织管理
// @Summary 创建部门
// @Description 创建部门
// @Router /account.CreateDepartment [post]
// @Param request body dtoorganize.CreateDepartmentRequest true "request"
// @Success 200 {object} dtoorganize.CreateDepartmentResponse "response"
func CreateDepartment(ctx *gin.Context, req *dtoorganize.CreateDepartmentRequest, resp *dtoorganize.CreateDepartmentResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateDepartment] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.CreateDepartment(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateDepartment] svcorganize.CreateDepartment failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_department_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteDepartment 删除部门
// @Tags 组织管理
// @Summary 删除部门
// @Description 删除部门
// @Router /account.DeleteDepartment [post]
// @Param request body dtoorganize.DeleteDepartmentRequest true "request"
// @Success 200 {object} dtoorganize.DeleteDepartmentResponse "response"
func DeleteDepartment(ctx *gin.Context, req *dtoorganize.DeleteDepartmentRequest, resp *dtoorganize.DeleteDepartmentResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteDepartment] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.DeleteDepartment(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteDepartment] svcorganize.DeleteDepartment failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_delete_department_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// RenameDepartment 重命名部门
// @Tags 组织管理
// @Summary 重命名部门
// @Description 重命名部门
// @Router /account.RenameDepartment [post]
// @Param request body dtoorganize.RenameDepartmentRequest true "request"
// @Success 200 {object} dtoorganize.RenameDepartmentResponse "response"
func RenameDepartment(ctx *gin.Context, req *dtoorganize.RenameDepartmentRequest, resp *dtoorganize.RenameDepartmentResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RenameDepartment] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.RenameDepartment(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[RenameDepartment] svcorganize.RenameDepartment failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_rename_department_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// MoveDepartment 移动部门
// @Tags 组织管理
// @Summary 移动部门
// @Description 移动部门
// @Router /account.MoveDepartment [post]
// @Param request body dtoorganize.MoveDepartmentRequest true "request"
// @Success 200 {object} dtoorganize.MoveDepartmentResponse "response"
func MoveDepartment(ctx *gin.Context, req *dtoorganize.MoveDepartmentRequest, resp *dtoorganize.MoveDepartmentResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[MoveDepartment] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.MoveDepartment(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[MoveDepartment] svcorganize.MoveDepartment failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_move_department_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetDepartmentTree 获取组织部门树
// @Tags 组织管理
// @Summary 获取组织部门树
// @Description 获取组织部门树
// @Router /account.GetDepartmentTree [post]
// @Param request body dtoorganize.GetDepartmentTreeRequest true "request"
// @Success 200 {object} dtoorganize.GetDepartmentTreeResponse "response"
func GetDepartmentTree(ctx *gin.Context, req *dtoorganize.GetDepartmentTreeRequest, resp *dtoorganize.GetDepartmentTreeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetDepartmentTree] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.GetDepartmentTree(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetDepartmentTree] svcorganize.GetDepartmentTree failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_department_data_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateDepartmentEmployee 新增部门员工
// @Tags 组织管理
// @Summary 新增部门员工
// @Description 新增部门员工
// @Router /account.CreateDepartmentEmployee [post]
// @Param request body dtoorganize.CreateDepartmentEmployeeRequest true "request"
// @Success 200 {object} dtoorganize.CreateDepartmentEmployeeResponse "response"
func CreateDepartmentEmployee(ctx *gin.Context, req *dtoorganize.CreateDepartmentEmployeeRequest, resp *dtoorganize.CreateDepartmentEmployeeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateDepartmentEmployee] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.CreateDepartmentEmployee(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateDepartmentEmployee] svcorganize.CreateDepartmentEmployee failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_employee_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// EditDepartmentEmployee 编辑部门员工
// @Tags 组织管理
// @Summary 编辑部门员工
// @Description 编辑部门员工
// @Router /account.EditDepartmentEmployee [post]
// @Param request body dtoorganize.EditDepartmentEmployeeRequest true "request"
// @Success 200 {object} dtoorganize.EditDepartmentEmployeeResponse "response"
func EditDepartmentEmployee(ctx *gin.Context, req *dtoorganize.EditDepartmentEmployeeRequest, resp *dtoorganize.EditDepartmentEmployeeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[EditDepartmentEmployee] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	//增量更新
	res, err := svcorganize.EditDepartmentEmployee(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditDepartmentEmployee] svcorganize.EditDepartmentEmployee failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_employee_info_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// EditCompanyInfo 更新组织信息
// @Tags 组织管理
// @Summary 更新组织信息
// @Description 更新组织信息
// @Router /account.EditCompanyInfo [post]
// @Param request body dtoorganize.EditCompanyInfoRequest true "request"
// @Success 200 {object} dtoorganize.EditCompanyInfoResponse "response"
func EditCompanyInfo(ctx *gin.Context, req *dtoorganize.EditCompanyInfoRequest, resp *dtoorganize.EditCompanyInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[EditCompanyInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.EditCompanyInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditCompanyInfo] svcorganize.EditCompanyInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_save_company_info_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// UploadOrganizeLogo 上传组织头像
// @Tags 组织管理
// @Summary 上传组织头像
// @Description 上传组织头像
// @Router /account.UploadOrganizeLogo [post]
// @Accept multipart/form-data
// @Param file formData file true "图片文件"
// @Success 200 {object} dtoorganize.UploadOrganizeLogoResponse "response"
func UploadOrganizeLogo(ctx *gin.Context) {

	res, err := svcorganize.UploadOrganizeLogo(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[UploadOrganizeLogo] svcorganize.UploadOrganizeLogo failed, err: %v", err)
		ctx.JSON(http.StatusOK, dtoorganize.UploadOrganizeLogoResponse{
			BaseResponse: apiobj.BaseResponse{
				Code:    errcode.ErrCode_InternalError,
				Message: "upload_organize_logo_failed",
			},
		})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// GetCompanyInfo 获取组织信息
// @Tags 组织管理
// @Summary 获取组织信息
// @Description 获取组织信息
// @Router /account.GetCompanyInfo [post]
// @Param request body dtoorganize.GetCompanyInfoRequest true "request"
// @Success 200 {object} dtoorganize.GetCompanyInfoResponse "response"
func GetCompanyInfo(ctx *gin.Context, req *dtoorganize.GetCompanyInfoRequest, resp *dtoorganize.GetCompanyInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetCompanyInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.GetCompanyInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetCompanyInfo] svcorganize.GetCompanyInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateDepartmentEmployeePrivate 创建组织员工(私有化版)
// @Tags 组织管理
// @Summary 创建组织员工(私有化版)
// @Description 创建组织员工(私有化版)
// @Router /account.CreateDepartmentEmployeePrivate [post]
// @Param request body dtoorganize.CreateDepartmentEmployeePrivateRequest true "request"
// @Success 200 {object} dtoorganize.CreateDepartmentEmployeePrivateResponse "response"
func CreateDepartmentEmployeePrivate(ctx *gin.Context, req *dtoorganize.CreateDepartmentEmployeePrivateRequest, resp *dtoorganize.CreateDepartmentEmployeePrivateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateDepartmentEmployeePrivate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.CreateDepartmentEmployeePrivate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateDepartmentEmployeePrivate] svcorganize.CreateDepartmentEmployeePrivate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_employee_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// EditDepartmentEmployeePrivate 编辑组织员工(私有化版)
// @Tags 组织管理
// @Summary 编辑组织员工(私有化版)
// @Description 编辑组织员工(私有化版)
// @Router /account.EditDepartmentEmployeePrivate [post]
// @Param request body dtoorganize.EditDepartmentEmployeePrivateRequest true "request"
// @Success 200 {object} dtoorganize.EditDepartmentEmployeePrivateResponse "response"
func EditDepartmentEmployeePrivate(ctx *gin.Context, req *dtoorganize.EditDepartmentEmployeePrivateRequest, resp *dtoorganize.EditDepartmentEmployeePrivateResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[EditDepartmentEmployeePrivate] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.EditDepartmentEmployeePrivate(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditDepartmentEmployeePrivate] svcorganize.EditDepartmentEmployeePrivate failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_employee_info_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ChangePasswordNotice 修改密码提醒
// @Tags 组织管理
// @Summary 修改密码提醒
// @Description 修改密码提醒
// @Router /account.ChangePasswordNotice [post]
// @Param request body dtoorganize.ChangePasswordNoticeRequest true "request"
// @Success 200 {object} dtoorganize.ChangePasswordNoticeResponse "response"
func ChangePasswordNotice(ctx *gin.Context, req *dtoorganize.ChangePasswordNoticeRequest, resp *dtoorganize.ChangePasswordNoticeResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ChangePasswordNotice] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.ChangePasswordNotice(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChangePasswordNotice] svcorganize.ChangePasswordNotice failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_change_password_notice_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ChangeDefaultPassword 修改默认密码
// @Tags 组织管理
// @Summary 修改默认密码
// @Description 修改默认密码
// @Router /account.ChangeDefaultPassword [post]
// @Param request body dtoorganize.ChangeDefaultPasswordRequest true "request"
// @Success 200 {object} dtoorganize.ChangeDefaultPasswordResponse "response"
func ChangeDefaultPassword(ctx *gin.Context, req *dtoorganize.ChangeDefaultPasswordRequest, resp *dtoorganize.ChangeDefaultPasswordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ChangeDefaultPassword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcorganize.ChangeDefaultPassword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChangeDefaultPassword] svcorganize.ChangeDefaultPassword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_password_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ResetPassword 重置密码
// @Tags 组织管理
// @Summary 重置密码
// @Description 重置密码
// @Router /account.ResetPassword [post]
// @Param request body dtoorganize.ResetPasswordRequest true "request"
// @Success 200 {object} dtoorganize.ResetPasswordResponse "response"
func ResetPassword(ctx *gin.Context, req *dtoorganize.ResetPasswordRequest, resp *dtoorganize.ResetPasswordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ResetPassword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svcorganize.ResetPassword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ResetPassword] svcorganize.ResetPassword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_reset_password_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// UploadWebSiteLogo 上传网站logo
// @Tags 组织管理
// @Summary 上传网站logo
// @Description 上传网站logo
// @Router /account.UploadWebSiteLogo [post]
// @Accept multipart/form-data
// @Param file formData file true "图片文件"
// @Success 200 {object} dtoorganize.UploadOrganizeLogoResponse "response"
func UploadWebSiteLogo(ctx *gin.Context) {
	res, err := svcorganize.UploadWebSiteLogo(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "[UploadOrganizeLogo] svcorganize.UploadOrganizeLogo failed, err: %v", err)
		ctx.JSON(http.StatusOK, dtoorganize.UploadOrganizeLogoResponse{
			BaseResponse: apiobj.BaseResponse{
				Code:    errcode.ErrCode_InternalError,
				Message: "upload_website_logo_failed",
			},
		})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

