package dtoorganize

import (
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type CreateDepartmentResponse struct {
	apiobj.BaseResponse
	Response CreateDepartmentEmbedResponse
}

type CreateDepartmentEmbedResponse struct {
	*accounttype.AccountDepartment `json:"department,omitempty"`
}

type DeleteDepartmentResponse struct {
	apiobj.BaseResponse
	Response DeleteDepartmentEmbedResponse
}
type DeleteDepartmentEmbedResponse struct {
}

type RenameDepartmentResponse struct {
	apiobj.BaseResponse
	Response RenameDepartmentEmbedResponse
}
type RenameDepartmentEmbedResponse struct {
	*accounttype.AccountDepartment `json:"department,omitempty"`
}

type MoveDepartmentResponse struct {
	apiobj.BaseResponse
	Response MoveDepartmentEmbedResponse
}
type MoveDepartmentEmbedResponse struct {
}

type GetDepartmentTreeResponse struct {
	apiobj.BaseResponse
	Response GetDepartmentTreeEmbedResponse
}

type EmployeeInfo struct {
	//uin
	Uin uint `json:"uin"`
	//CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	//UserName 唯一用户名
	UserName string `json:"user_name"`
	//昵称
	Name string `json:"name"`
	//邮箱
	Email string `json:"email"`
	//手机号
	Phone string `json:"phone"`
	//员工id(employee_id)
	EmployeeID uint `json:"employee_id"`
	//角色(sys_admin/sys_employee)
	SysRole accounttype.SysRole `json:"role"`
	//部门列表ids 首个为主部门id
	DepartmentIDs []uint `json:"department_ids" gorm:"-"`
}

type GetDepartmentTreeEmbedResponse struct {
	Departments []accounttype.AccountDepartment `json:"departments"`
	Employees   []EmployeeInfo                  `json:"employees,omitempty"`
}

type CreateDepartmentEmployeeResponse struct {
	apiobj.BaseResponse
	Response CreateDepartmentEmployeeEmbedResponse
}
type CreateDepartmentEmployeeEmbedResponse struct {
	//员工的基础信息
	*EmployeeInfo `json:"employee,omitempty"`
}

type EditDepartmentEmployeeResponse struct {
	apiobj.BaseResponse
	Response EditDepartmentEmployeeEmbedResponse
}
type EditDepartmentEmployeeEmbedResponse struct {
	//员工的基础信息
	*EmployeeInfo `json:"employee,omitempty"`
}

type EditCompanyInfoResponse struct {
	apiobj.BaseResponse
	Response EditCompanyInfoEmbedResponse
}
type EditCompanyInfoEmbedResponse struct {
}

type UploadOrganizeLogoResponse struct {
	apiobj.BaseResponse
	Response UploadOrganizeLogoEmbedResponse
}
type UploadOrganizeLogoEmbedResponse struct {
	FileID    uint   `json:"file_id"`
	PublicUrl string `json:"public_url"`
}

type GetCompanyInfoResponse struct {
	apiobj.BaseResponse
	Response GetCompanyInfoEmbedResponse
}
type GetCompanyInfoEmbedResponse struct {
	accounttype.Company
}

type CreateDepartmentEmployeePrivateResponse struct {
	apiobj.BaseResponse
	Response CreateDepartmentEmployeePrivateEmbedResponse
}
type CreateDepartmentEmployeePrivateEmbedResponse struct {
	//员工的基础信息
	*EmployeeInfo `json:"employee,omitempty"`
}

type EditDepartmentEmployeePrivateResponse struct {
	apiobj.BaseResponse
	Response EditDepartmentEmployeePrivateEmbedResponse
}
type EditDepartmentEmployeePrivateEmbedResponse struct {
	*EmployeeInfo `json:"employee,omitempty"`
}

type ChangePasswordNoticeResponse struct {
	apiobj.BaseResponse
	Response ChangePasswordNoticeEmbedResponse
}
type ChangePasswordNoticeEmbedResponse struct {
	PasswordChanged bool `json:"password_changed"`
}

type ChangeDefaultPasswordResponse struct {
	apiobj.BaseResponse
	Response ChangeDefaultPasswordEmbedResponse
}
type ChangeDefaultPasswordEmbedResponse struct {
}

type ResetPasswordResponse struct {
	apiobj.BaseResponse
	Response ResetPasswordEmbedResponse
}
type ResetPasswordEmbedResponse struct {
}
