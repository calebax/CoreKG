package companyctl

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/admin/models/company"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateCompanyEmployeeRequest 创建团队成员请求
type CreateCompanyEmployeeRequest struct {
	apiobj.BaseRequest
	Request company.CreateCompanyEmployeeOption
}

// Validity 校验
func (req *CreateCompanyEmployeeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.CompanyID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请选择团队"
		return
	}
	if req.Request.UserID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请选择用户"
		return
	}
	switch req.Request.Role {
	case accounttype.SysRoleSysEmployee, accounttype.SysRoleSysAdmin:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "角色不合法"
		return
	}
}

// CreateCompanyEmployeeResponse 创建团队成员响应
type CreateCompanyEmployeeResponse struct {
	apiobj.BaseResponse
	Response struct {
		accounttype.Employee
	}
}

type ListCompanyEmployeeRequest apiobj.QueryRequest

// Validity 校验
func (req *ListCompanyEmployeeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "created_at", "updated_at", "user_name", "user_phone",
			"name desc", "created_at desc", "updated_at desc", "phone":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "orderBy字段不支持"
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "company_id", "user_id", "phone", "user_name", "user_phone":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的字段只能有一个值"
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "查询条件中的字段不支持"
			return
		}
	}
	return
}

// ListCompanyEmployeeResponse 团队成员列表响应
type ListCompanyEmployeeResponse struct {
	apiobj.BaseResponse
	Response company.QueryCompanyEmployeeListResponse
}

// UpdateCompanyEmployeeRoleRequest 更换成员角色请求
type UpdateCompanyEmployeeRoleRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID   uint                `json:"id"`
		Role accounttype.SysRole `json:"role"`
	}
}

// Validity 校验
func (req *UpdateCompanyEmployeeRoleRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请选择团队成员"
		return
	}
	switch req.Request.Role {
	case accounttype.SysRoleSysEmployee, accounttype.SysRoleSysAdmin:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "角色不合法"
		return
	}
}

// UpdateCompanyEmployeeRoleResponse 响应
type UpdateCompanyEmployeeRoleResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
