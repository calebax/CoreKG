package dtoorganize

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

const (
	MaxDepartmentNameLen = 100
)

type CreateDepartmentRequest struct {
	apiobj.BaseRequest
	Request CreateDepartmentEmbedRequest
}

type CreateDepartmentEmbedRequest struct {
	//部门名
	Name string `json:"name"`
	//父级部门id
	ParentId uint `json:"parent_id"`
}

func (opt *CreateDepartmentRequest) Validity(resp *CreateDepartmentResponse) {
	if len(opt.Request.Name) <= 0 || len(opt.Request.Name) > MaxDepartmentNameLen {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_name_invalid"
		return
	}
	if opt.Request.ParentId < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_parent_id_invalid"
		return
	}
}

type DeleteDepartmentRequest struct {
	apiobj.BaseRequest
	Request DeleteDepartmentEmbedRequest
}
type DeleteDepartmentEmbedRequest struct {
	//部门id
	ID uint `json:"id"`
}

func (opt *DeleteDepartmentRequest) Validity(resp *DeleteDepartmentResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_id_invalid"
		return
	}
}

type RenameDepartmentRequest struct {
	apiobj.BaseRequest
	Request RenameDepartmentEmbedRequest
}
type RenameDepartmentEmbedRequest struct {
	//部门id
	ID uint `json:"id"`
	//部门名
	Name string `json:"name"`
}

func (opt *RenameDepartmentRequest) Validity(resp *RenameDepartmentResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_id_invalid"
		return
	}
	if len(opt.Request.Name) <= 0 || len(opt.Request.Name) > MaxDepartmentNameLen {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_name_invalid"
		return
	}
}

type MoveDepartmentRequest struct {
	apiobj.BaseRequest
	Request MoveDepartmentEmbedRequest
}
type MoveDepartmentEmbedRequest struct {
	//部门id(被移动部门id)
	DepartmentId uint `json:"department_id"`
	//目标前置部门id
	PreID uint `json:"pre_id"`
	//目标后置部门id
	PostID uint `json:"post_id"`
}

func (opt *MoveDepartmentRequest) Validity(resp *MoveDepartmentResponse) {
	if opt.Request.DepartmentId <= 0 || opt.Request.PreID < 0 || opt.Request.PostID < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_id_invalid"
		return
	}
	if opt.Request.DepartmentId == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_id_invalid"
		return
	}
	if opt.Request.PreID == 0 && opt.Request.PostID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_interval_invalid"
		return
	}
	if opt.Request.DepartmentId == opt.Request.PreID || opt.Request.DepartmentId == opt.Request.PostID {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_interval_invalid"
		return
	}
}

type GetDepartmentTreeRequest struct {
	apiobj.BaseRequest
	Request GetDepartmentTreeEmbedRequest
}
type GetDepartmentTreeEmbedRequest struct {
	//是否包含员工数据
	IncludeEmployee bool `json:"include_employee"`
}

func (opt *GetDepartmentTreeRequest) Validity(_ *GetDepartmentTreeResponse) {
}

type CreateDepartmentEmployeeRequest struct {
	apiobj.BaseRequest
	Request CreateDepartmentEmployeeEmbedRequest
}
type CreateDepartmentEmployeeEmbedRequest struct {
	//员工的基础信息
	EmployeeInfo `json:"employee,omitempty"`
}

func (opt *CreateDepartmentEmployeeRequest) Validity(resp *CreateDepartmentEmployeeResponse) {
	if len(opt.Request.Phone) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_empty"
		return
	}
	if len(opt.Request.Name) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_real_name_empty"
		return
	}
	if len(opt.Request.Email) > 0 {
		if err := validate.IsEmail(opt.Request.Email); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_email_format" // 错误的邮箱格式
			resp.MessageData = map[string]interface{}{
				"error": err.Error(),
			}
			return
		}
	}
	if err := validate.IsPhone(opt.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_phone_format" // 错误的手机格式
		return
	}
	if len(opt.Request.DepartmentIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_ids_empty"
		return
	}
	for _, id := range opt.Request.DepartmentIDs {
		if id <= 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_department_id_invalid"
			return
		}
	}
	switch opt.Request.SysRole {
	case accounttype.SysRoleSysAdmin, accounttype.SysRoleSysEmployee:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_role_parameter"
		return
	}
}

type EditDepartmentEmployeeRequest struct {
	apiobj.BaseRequest
	Request EditDepartmentEmployeeEmbedRequest
}
type EditDepartmentEmployeeEmbedRequest struct {
	//员工的基础信息
	EmployeeInfo `json:"employee,omitempty"`
}

func (opt *EditDepartmentEmployeeRequest) Validity(resp *EditDepartmentEmployeeResponse) {
	if opt.Request.Uin <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_edit_illegal_employee_uin"
		return
	}
	if len(opt.Request.DepartmentIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_ids_empty"
		return
	}

	if len(opt.Request.Email) > 0 {
		if err := validate.IsEmail(opt.Request.Email); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_email_format" // 错误的邮箱格式
			resp.MessageData = map[string]interface{}{
				"error": err.Error(),
			}
			return
		}
	}
	if len(opt.Request.Phone) > 0 {
		if err := validate.IsPhone(opt.Request.Phone); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_phone_format" // 错误的手机格式
			return
		}
	}
	if len(opt.Request.DepartmentIDs) > 0 {
		for _, id := range opt.Request.DepartmentIDs {
			if id <= 0 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_department_id_invalid"
				return
			}
		}
	}
	if opt.Request.SysRole != "" {
		switch opt.Request.SysRole {
		case accounttype.SysRoleSysAdmin, accounttype.SysRoleSysEmployee:
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_role_parameter"
			return
		}
	}
}

type EditCompanyInfoRequest struct {
	apiobj.BaseRequest
	Request EditCompanyInfoEmbedRequest
}
type EditCompanyInfoEmbedRequest struct {
	// Name 公司名称
	Name string `json:"name"`
	// Alias 公司别名
	Alias string `json:"alias"`
	// Description 公司描述
	Description string `json:"description"`
	// Logo 公司logo
	Logo string `json:"logo"`
	// Address 公司地址
	Address string `json:"address"`
	// Tel 公司电话
	Tel string `json:"tel"`
	// Email 公司邮箱
	Email string `json:"email"`
	// Website 公司网址
	Website string `json:"website"`
}

func (opt *EditCompanyInfoRequest) Validity(resp *EditCompanyInfoResponse) {
	if len(opt.Request.Name) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_company_name_empty"
		return
	}
}

type GetCompanyInfoRequest struct {
	apiobj.BaseRequest
	Request GetCompanyInfoEmbedRequest
}
type GetCompanyInfoEmbedRequest struct {
}

func (opt *GetCompanyInfoRequest) Validity(resp *GetCompanyInfoResponse) {
}

type CreateDepartmentEmployeePrivateRequest struct {
	apiobj.BaseRequest
	Request CreateDepartmentEmployeePrivateEmbedRequest
}
type CreateDepartmentEmployeePrivateEmbedRequest struct {
	//员工的基础信息
	EmployeeInfo `json:"employee,omitempty"`
}

func (opt *CreateDepartmentEmployeePrivateRequest) Validity(resp *CreateDepartmentEmployeePrivateResponse) {
	if len(opt.Request.Email) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_email_empty"
		return
	}
	if len(opt.Request.Name) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_real_name_empty"
		return
	}

	// if opt.Request.Email != "" || version.DeployMode() != "" {
	// 	if err := validate.IsEmail(opt.Request.Email); err != nil {
	// 		resp.Code = errcode.ErrCode_BadRequest
	// 		resp.Message = "account_invalid_email_format" // 错误的邮箱格式
	// 		resp.MessageData = map[string]interface{}{
	// 			"error": err.Error(),
	// 		}
	// 		return
	// 	}
	// }

	if err := validate.IsEmail(opt.Request.Email); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_email_format" // 错误的邮箱格式
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	if opt.Request.Phone != "" {
		if err := validate.IsPhone(opt.Request.Phone); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_phone_format" // 错误的手机格式
			return
		}
	}
	if len(opt.Request.DepartmentIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_ids_empty"
		return
	}
	for _, id := range opt.Request.DepartmentIDs {
		if id <= 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_department_id_invalid"
			return
		}
	}
	switch opt.Request.SysRole {
	case accounttype.SysRoleSysAdmin, accounttype.SysRoleSysEmployee:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_role_parameter"
		return
	}

	return
}

type EditDepartmentEmployeePrivateRequest struct {
	apiobj.BaseRequest
	Request EditDepartmentEmployeePrivateEmbedRequest
}
type EditDepartmentEmployeePrivateEmbedRequest struct {
	//员工的基础信息
	*EmployeeInfo `json:"employee,omitempty"`
}

func (opt *EditDepartmentEmployeePrivateRequest) Validity(resp *EditDepartmentEmployeePrivateResponse) {
	if len(opt.Request.DepartmentIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_department_ids_empty"
		return
	}
}

type ChangePasswordNoticeRequest struct {
	apiobj.BaseRequest
	Request ChangePasswordNoticeEmbedRequest
}
type ChangePasswordNoticeEmbedRequest struct {
	UserID       uint   `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
	AlwaysIgnore bool   `json:"always_ignore"`
}

func (opt *ChangePasswordNoticeRequest) Validity(resp *ChangePasswordNoticeResponse) {

}

type ChangeDefaultPasswordRequest struct {
	apiobj.BaseRequest
	Request ChangeDefaultPasswordEmbedRequest
}
type ChangeDefaultPasswordEmbedRequest struct {
	UserID       uint   `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
	OldPassword  string `json:"old_password"`
	NewPassword  string `json:"new_password"`
}

func (opt *ChangeDefaultPasswordRequest) Validity(resp *ChangeDefaultPasswordResponse) {
	if len(opt.Request.OldPassword) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_old_password_empty"
		return
	}
	if opt.Request.NewPassword == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_new_password_empty" // 新密码不能为空
		return
	}
	if len(opt.Request.NewPassword) < 8 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_password_too_short" // 密码长度最短8位
		return
	}
	if len(opt.Request.NewPassword) > 36 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_password_too_long" // 密码长度最长36位
		return
	}
	if validate.HasCNChar(opt.Request.NewPassword) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_password_no_chinese" // 密码不能包含中文
		return
	}
}

type ResetPasswordRequest struct {
	apiobj.BaseRequest
	Request ResetPasswordEmbedRequest
}
type ResetPasswordEmbedRequest struct {
	// 用户uin
	Uin uint `json:"uin"`
}

func (opt *ResetPasswordRequest) Validity(resp *ResetPasswordResponse) {
	if opt.Request.Uin == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_uin_invalid"
		return
	}
}
