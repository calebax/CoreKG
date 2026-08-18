package userctl

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/admin/models/company"
	"github.com/insmtx/corekg/apps/admin/models/user"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	apiobj.BaseRequest
	Request user.CreateUserOption
}

// Validity 验证参数
func (req *CreateUserRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户名不能为空"
		return
	}
	if req.Request.Password == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "密码不能为空"
		return
	}
	if req.Request.Email == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "邮箱不能为空"
		return
	}
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "手机号不能为空"
		return
	}
	if err := validate.IsEmail(req.Request.Email); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "邮箱格式不正确"
		return
	}
	if err := validate.IsPhone(req.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "手机号格式不正确"
		return
	}
	if req.Request.CompanyQuota < 2 || req.Request.CompanyQuota > 999 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "公司配额必须在2-99之间"
		return
	}
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	apiobj.BaseResponse
	Response struct {
		accounttype.User
	}
}

// ListUserRequest 查询用户请求
type ListUserRequest apiobj.QueryRequest

// Validity 验证参数
func (req *ListUserRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "created_at", "updated_at",
			"name desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "orderBy字段不支持"
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "phone":
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

// ListUserResponse 查询用户响应
type ListUserResponse struct {
	apiobj.BaseResponse
	Response user.QueryUserListResponse
}

// ModifyUserRequest 修改用户请求
type ModifyUserRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
		user.CreateUserOption
	}
}

// Validity 验证参数
func (req *ModifyUserRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户ID不能为空"
		return
	}
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户名不能为空"
		return
	}
	if req.Request.Email == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "邮箱不能为空"
		return
	}
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "手机号不能为空"
		return
	}
	if err := validate.IsEmail(req.Request.Email); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "邮箱格式不正确"
		return
	}
	if err := validate.IsPhone(req.Request.Phone); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "手机号格式不正确"
		return
	}
	if req.Request.CompanyQuota < 2 || req.Request.CompanyQuota > 999 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "公司配额必须在2-999之间"
		return
	}
}

// ModifyUserResponse 修改用户响应
type ModifyUserResponse struct {
	apiobj.BaseResponse
	Response struct {
		accounttype.User
	}
}

// ModifyUserPasswordRequest 修改用户请求
type ModifyUserPasswordRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID       uint   `json:"id"`
		Password string `json:"password"`
	}
}

// Validity 验证参数
func (req *ModifyUserPasswordRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户ID不能为空"
		return
	}
	if req.Request.Password == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户名不能为空"
		return
	}
}

// ModifyUserPasswordResponse 修改用户响应
type ModifyUserPasswordResponse struct {
	apiobj.BaseResponse
	Response struct {
		accounttype.User
	}
}

// GetUserDetailRequest 获取用户详情请求
type GetUserDetailRequest apiobj.DetailIdRequest

// Validity 验证参数
func (req *GetUserDetailRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户ID不能为空"
		return
	}
	return
}

// GetUserDetailResponse 修改用户响应
type GetUserDetailResponse struct {
	apiobj.BaseResponse
	Response struct {
		user.UserDetail
		EmployeeList []*company.QueryCompanyEmployeeListItem `json:"employee_list"`
	}
}
