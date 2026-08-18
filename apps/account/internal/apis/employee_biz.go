package apis

import (
	"regexp"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// 登陆相关

// LoginPasswordRequest 企业微信登录
type LoginPasswordRequest struct {
	apiobj.BaseRequest

	Request struct {
		DomainName string `json:"domain_name"`
		// Username 用户名 不能为空, 邮箱或者用户名
		Username string `json:"username"`
		// Password 原密码md5加密后的密码
		Password string `json:"password"`
	}
}

// LoginPasswordRequest 企业微信登录
type LoginByEmailRequest struct {
	apiobj.BaseRequest

	Request struct {
		DomainName string `json:"domain_name"`
		// Username 用户名 邮箱
		Email string `json:"email"`
	}
}

// LoginWechatRequest 微信用户登录
type LoginWechatRequest struct {
	apiobj.BaseRequest

	Request struct {
		Code    string `json:"code"`
		AppID   string `json:"appid"`
		AgentID string `json:"agentid"`
	}
}

// ListMenuRequest 浏览菜单请求
type ListMenuRequest struct {
	apiobj.BaseRequest

	Request struct {
		Platform string `json:"platform"`
		AppID    string `json:"app_id"`
	}
}

// ListMenuResponse 浏览菜单响应
type ListMenuResponse struct {
	apiobj.BaseResponse
	Response struct {
		// MenuList []employee.Menu `json:"menu_list"`
	}
}

// SendCodeForgotPasswordRequest 发送找回密码验证码请求
type SendCodeForgotPasswordRequest struct {
	apiobj.BaseRequest
	Request struct {
		// By 找回密码方式，可以为空，默认为email
		By string `json:"by"`
		// Value 如果by为email，则为邮箱；如果by为mobile，则为手机号
		Value string `json:"value"`
	}
}

// type CreatePositionRequest struct {
// 	apiobj.BaseRequest

// 	Request accounttype.CreatePositionOption
// }

// type ListPositionRequest apiobj.QueryRequest

// // Validity 验证有效性
// func (req *ListPositionRequest) Validity(resp *apiobj.BaseResponse) {
// 	if req.Request.Offset < 0 || req.Request.Limit < 0 {
// 		resp.Code = errcode.ErrCode_BadRequest
// 		resp.Message = "offset和limit必须大于0"
// 		return
// 	}
// 	for _, v := range req.Request.OrderBy {
// 		switch v {
// 		case "name", "created_at", "updated_at",
// 			"name desc", "created_at desc", "updated_at desc":
// 		default:
// 			resp.Code = errcode.ErrCode_BadRequest
// 			resp.Message = "orderBy字段不支持"
// 			return
// 		}
// 	}
// 	for _, v := range req.Request.Filters {
// 		switch v.Field {
// 		case "name":
// 			if len(v.Value) != 1 {
// 				resp.Code = errcode.ErrCode_BadRequest
// 				resp.Message = "查询条件中的字段只能有一个值"
// 				return
// 			}
// 			if v.Value[0] == "" {
// 				resp.Code = errcode.ErrCode_BadRequest
// 				resp.Message = "查询条件中的值不能为空"
// 				return
// 			}

// 		default:
// 			resp.Code = errcode.ErrCode_BadRequest
// 			resp.Message = "查询条件中的字段不存在, " + v.Field
// 			return
// 		}
// 	}
// }

// Validity 验证有效期
func (in *SendCodeForgotPasswordRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.By == "mobile" {
		out.Code = errcode.ErrCode_NotSupportMobileForgotPassword
		out.Message = "account_not_support_mobile_forgot_password" // 不支持通过手机号找回密码
		return
	}
	if in.Request.By != "email" && in.Request.By != "" {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_unsupported_method" // 暂不支持的方式
		out.MessageData = map[string]interface{}{
			"method": in.Request.By,
		}
		return
	}
	if err := validate.IsEmail(in.Request.Value); err != nil {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_invalid_email_format" // 错误的邮箱格式
		out.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}
}

// ResetPasswordForgotPasswordRequest 重置密码请求
type ResetPasswordForgotPasswordRequest struct {
	apiobj.BaseRequest
	Request struct {
		// By 找回密码方式，可以为空，默认为email
		By string `json:"by"`
		// Value 如果by为email，则为邮箱；如果by为mobile，则为手机号
		Value string `json:"value"`
		// Code 验证码
		Code string `json:"code"`
		// Password 新密码
		Password string `json:"password"`
	}
}

// Validity 验证有效期
func (in *ResetPasswordForgotPasswordRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.By == "mobile" {
		out.Code = errcode.ErrCode_NotSupportMobileForgotPassword
		out.Message = "account_not_support_mobile_forgot_password" // 不支持通过手机号找回密码
		return
	}
	if in.Request.By != "email" && in.Request.By != "" {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_unsupported_method" // 暂不支持的方式
		out.MessageData = map[string]interface{}{
			"method": in.Request.By,
		}
		return
	}
	if err := validate.IsEmail(in.Request.Value); err != nil {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_invalid_email_format" // 错误的邮箱格式
		out.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}
	if len(in.Request.Password) < 6 {
		out.Code = errcode.ErrCode_PasswordTooShort
		out.Message = "account_password_too_short" // 密码长度不能小于6位
		return
	}
}

// ********** 用户管理 **********

// ListEmployeeRequest 浏览用户请求
type ListEmployeeRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

// GetEmployeeRequest 获取用户信息请求
type GetEmployeeRequest struct {
	apiobj.BaseRequest
	Request struct {
		EmployeeID uint `json:"employee_id"`
	}
}

// ListEmployeeResponse 浏览用户响应
type ListEmployeeResponse struct {
	apiobj.BaseResponse
	Response employee.EmployeeInfoItemList
}

// EmployeeResponse 单用户信息响应
type EmployeeResponse struct {
	apiobj.BaseResponse
	Response struct {
		employee.EmployeeDetail
		// employee.EmployeeInfoItem
		// Positions   []*accounttype.Position `json:"positions"`
		// ActionPaths []string                `json:"action_paths"`
	}
}

// Validity 验证有效性
func (in *ListEmployeeRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.Offset < 0 || in.Request.Limit < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range in.Request.OrderBy {
		switch v {
		case "name", "sys_role", "mobile", "created_at", "updated_at",
			"username desc", "email desc", "mobile desc", "created_at desc", "updated_at desc":
		default:
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range in.Request.Filters {
		switch v.Field {
		case "auto", "name", "sys_role", "email", "mobile", "status":
			if len(v.Value) != 1 {
				out.Code = errcode.ErrCode_BadRequest
				out.Message = "account_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				out.Code = errcode.ErrCode_BadRequest
				out.Message = "account_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_filter_field_data_data" // 查询条件中的字段不存在
			out.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// isValidUsername 校验用户名
func isValidUsername(username string) bool {
	//校验用户名是否只包含英文或数字
	regex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return regex.MatchString(username)
}

// CreateEmployeeRequest 创建用户请求
type CreateEmployeeRequest struct {
	apiobj.BaseRequest

	Request employee.CreateEmployeeItem
}

//// Validity 验证有效性
//func (in *CreateEmployeeRequest) Validity(out *apiobj.BaseResponse) {
//	in.Request.Username = strings.TrimSpace(in.Request.Username)
//	in.Request.Email = strings.TrimSpace(in.Request.Email)
//	in.Request.Mobile = strings.TrimSpace(in.Request.Mobile)
//	in.Request.RealName = strings.TrimSpace(in.Request.RealName)
//
//	if in.Request.Username == "" {
//		out.Code = errcode.ErrCode_BadRequest
//		out.Message = "account_username_empty" // 用户名不能为空
//		return
//	}
//	if len(in.Request.Username) < 3 {
//		out.Code = errcode.ErrCode_BadRequest
//		out.Message = "account_username_too_short" // 用户名长度不能小于3位
//		return
//	}
//
//	if in.Request.Email == "" {
//		out.Code = errcode.ErrCode_BadRequest
//		out.Message = "account_email_empty" // 邮箱不能为空
//		return
//	}
//	if err := validate.IsEmail(in.Request.Email); err != nil {
//		out.Code = errcode.ErrCode_BadRequest
//		out.Message = "account_invalid_email_format" // 错误的邮箱格式
//		out.MessageData = map[string]interface{}{
//			"error": err.Error(),
//		}
//		return
//	}
//	if in.Request.Mobile != "" {
//		if err := validate.IsPhone(in.Request.Mobile); err != nil {
//			out.Code = errcode.ErrCode_BadRequest
//			out.Message = "account_invalid_phone_format" // 错误的手机号格式
//			return
//		}
//	}
//	if in.Request.Password != "" && len(in.Request.Password) < 6 {
//		out.Code = errcode.ErrCode_BadRequest
//		out.Message = "account_password_too_short" // 密码长度不能小于6位
//		return
//	}
//}

type EmployeeValidityOption employee.CreateEmployeeItem

func (opt *EmployeeValidityOption) Validity(resp *apiobj.BaseResponse) {
	opt.RealName = strings.TrimSpace(opt.RealName)
	opt.Email = strings.TrimSpace(opt.Email)
	opt.Phone = strings.TrimSpace(opt.Phone)
	opt.RealName = strings.TrimSpace(opt.RealName)

	if opt.RealName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_real_name_empty" // 姓名不能为空
		return
	}
	if opt.Username != "" {
		if !isValidUsername(opt.Username) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_username" // 用户名必须为英文或数字
			return
		}
	}

	if opt.Email == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_email_empty" // 邮箱不能为空
		return
	}

	if opt.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_empty" // 手机号不能为空
		return
	}

	if err := validate.IsEmail(opt.Email); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_email_format" // 错误的邮箱格式
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}
	if opt.Phone != "" {
		if err := validate.IsPhone(opt.Phone); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_phone_format" // 错误的手机号格式
			return
		}
	}
	if len(opt.PositionIDs) < 1 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_position_empty" // 职位不能为空
		return
	}
}

func (req *CreateEmployeeRequest) Validity(resp *apiobj.BaseResponse) {
	in := EmployeeValidityOption(req.Request)
	in.Validity(resp)
}

// UpdateEmployeeRequest 更新用户请求
type UpdateEmployeeRequest struct {
	apiobj.BaseRequest
	Request employee.UpdateEmployeeItem
}

// Validity 验证有效性
func (in *UpdateEmployeeRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.EmployeeID == 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_user_id_empty" // user_id不能为空
		return
	}
	if in.Request.Email != "" {
		if err := validate.IsEmail(in.Request.Email); err != nil {
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_email_format" // 错误的邮箱格式
			out.MessageData = map[string]interface{}{
				"error": err.Error(),
			}
			return
		}
	}
	if in.Request.Mobile != "" {
		if err := validate.IsPhone(in.Request.Mobile); err != nil {
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_phone_format" // 错误的手机号格式
			return
		}
	}
	if in.Request.Username != "" {
		if !isValidUsername(in.Request.Username) {
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_username" // 用户名必须为英文或数字
			return
		}
	}
}

// DeleteEmployeeRequest 删除用户请求
type DeleteEmployeeRequest struct {
	apiobj.BaseRequest
	Request struct {
		EmployeeID uint `json:"employee_id"`
		// DeleteReason 删除原因
		DeleteReason string `json:"delete_reason"`
	}
}

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	apiobj.BaseResponse
	Response accounttype.Employee
}

// ListEmployeeNickIDRequest 获取员工昵称和ID列表请求
type ListEmployeeNickIDRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

// Validity 验证有效性
func (in *ListEmployeeNickIDRequest) Validity(out *apiobj.BaseResponse) {
	if in.Request.Offset < 0 || in.Request.Limit < 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range in.Request.OrderBy {
		switch v {
		case "created_at", "updated_at",
			"user_name desc", "created_at desc", "updated_at desc":
		default:
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range in.Request.Filters {
		switch v.Field {
		case "auto", "user_name":
			if len(v.Value) != 1 {
				out.Code = errcode.ErrCode_BadRequest
				out.Message = "account_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				out.Code = errcode.ErrCode_BadRequest
				out.Message = "account_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			out.Code = errcode.ErrCode_BadRequest
			out.Message = "account_invalid_filter_field_data" // 查询条件中的字段不存在
			out.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ListEmployeeNickIDResponse 获取员工昵称和ID列表响应
type ListEmployeeNickIDResponse struct {
	apiobj.BaseResponse
	Response employee.EmployeeSimpleInfoItemList
}

/************** 员工个人信息 **************/

type GetMyUserInfoRequest apiobj.BaseRequest

type GetMyUserInfoResponse struct {
	apiobj.BaseResponse

	Response struct {
		employee.EmployeeDetail
	}
}

type ModifyMyUserInfoRequest struct {
	apiobj.BaseRequest
	Request *employee.EmployeeSimpleInfo
}

type ModifyMyUserInfoResponse apiobj.BaseResponse

func (opt *ModifyMyUserInfoRequest) Validity(resp *ModifyMyUserInfoResponse) {
	if opt.Request.Email == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_email_empty" // 邮箱不能为空
		return
	}

	if opt.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_empty" // 手机号不能为空
		return
	}

	if opt.Request.Email != "" {
		if err := validate.IsEmail(opt.Request.Email); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_email_format" // 错误的邮箱格式
			resp.MessageData = map[string]interface{}{
				"error": err.Error(),
			}
			return
		}
	}

	if opt.Request.Phone != "" {
		if err := validate.IsPhone(opt.Request.Phone); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_phone_format" // 错误的手机号格式
			return
		}
	}
}

type ChangeMyWechatRequest struct {
	apiobj.BaseRequest
	Request struct {
		Code string `json:"code"`
	}
}

type ChangeMyWechatResponse struct {
	apiobj.BaseResponse
}

// GetMyActionRequest 获取我的权限
type GetMyActionRequest apiobj.BaseRequest

// GetMyActionResponse 获取我的权限
type GetMyActionResponse struct {
	apiobj.BaseResponse
	Response struct {
		ActionPaths []string `json:"action_paths"`
	}
}

/************** 权限管理 **************/
type ListPositionRequest apiobj.QueryRequest

type ListPositionResponse struct {
	apiobj.BaseResponse

	Response employee.QueryPositionListResponse
}

// Validity 验证有效性
func (req *ListPositionRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "created_at", "updated_at",
			"name desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_filter_field_data" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

type CreatePositionRequest struct {
	apiobj.BaseRequest

	Request employee.CreatePositionOption
}

type CreatePositionResponse struct {
	apiobj.BaseResponse

	Response *accounttype.Position
}

func (req *CreatePositionRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_position_name_empty" // 请先写名称
		return
	}
	//if len(req.Request.PrivilegeIDs) == 0 {
	//	resp.Code = errcode.ErrCode_BadRequest
	//	resp.Message = "account_position_privileges_empty" // 请选择相应的权限
	//	return
	//}
}

type ModifyPositionInfoRequest struct {
	apiobj.BaseRequest

	Request struct {
		ID uint `json:"id"`
		accounttype.Position
	}
}

func (req *ModifyPositionInfoRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty" // ID不能为空
		return
	}
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_position_name_empty" // 请先写名称
		return
	}
}

type ModifyPositionInfoResponse struct {
	apiobj.BaseResponse

	Response accounttype.Position
}

type DeletePositionRequest apiobj.DetailIdRequest

func (req *DeletePositionRequest) Validity(resp *DeletePositionResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty" // ID不能为空
		return
	}
}

type DeletePositionResponse apiobj.BaseResponse

type GetPositionDetailRequest apiobj.DetailIdRequest

func (req *GetPositionDetailRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty" // ID不能为空
		return
	}
}

type GetPositionDetailResponse struct {
	apiobj.BaseResponse

	Response employee.PositionDetail
}

type ModifyPositionPrivilegeRequest struct {
	apiobj.BaseRequest

	Request struct {
		ID           uint   `json:"id"`
		PrivilegeIDs []uint `json:"privilege_ids"`
	}
}

func (req *ModifyPositionPrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty" // ID不能为空
		return
	}
	//if len(req.Request.PrivilegeIDs) == 0 {
	//	resp.Code = errcode.ErrCode_BadRequest
	//	resp.Message = "account_position_privileges_empty" // 请选择相应的权限
	//	return
	//}
}

type ModifyPositionPrivilegeResponse struct {
	apiobj.BaseResponse

	Response accounttype.Position
}

type ListPrivilegeRequest apiobj.QueryRequest

type ListPrivilegeResponse struct {
	apiobj.BaseResponse

	Response employee.QueryPrivilegeListResponse
}

// Validity 验证有效性
func (req *ListPrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "created_at", "updated_at",
			"name desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_filter_field_data" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

type CreatePrivilegeRequest struct {
	apiobj.BaseRequest

	Request accounttype.Privilege
}

type CreatePrivilegeResponse struct {
	apiobj.BaseResponse

	Response *accounttype.Privilege
}

func (req *CreatePrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if strings.HasPrefix(req.Request.API, global.PrefixAPIV2) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_api_version_not_required" // api不需要输入版本号
		return
	}
}

type ModifyPrivilegeRequest struct {
	apiobj.BaseRequest
	Request struct {
		accounttype.Privilege
	}
}

func (req *ModifyPrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty" // ID不能为空
		return
	}
	if strings.HasPrefix(req.Request.API, global.PrefixAPIV2) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_api_version_not_required" // api不需要输入版本号
		return
	}
}

type ModifyPrivilegeResponse struct {
	apiobj.BaseResponse
}

type DeletePrivilegeRequest apiobj.DetailIdRequest

func (req *DeletePrivilegeRequest) Validity(resp *DeletePrivilegeResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_empty" // ID不能为空
		return
	}
}

type DeletePrivilegeResponse apiobj.BaseResponse

type GetEmployeeBindKeyRequest apiobj.DetailIdRequest

type GetEmployeeBindKeyResponse struct {
	apiobj.BaseResponse
	Response struct {
		EmployeeID uint   `json:"employee_id"`
		BindKey    string `json:"bind_key"`
	}
}

type BindEmployeeWechatRequest struct {
	apiobj.BaseRequest

	Request struct {
		BindKey    string `json:"bind_key"`
		Code       string `json:"code"`
		EmployeeID uint   `json:"-"`
	}
}

func (req *BindEmployeeWechatRequest) Validity(resp *employee.BindEmployeeWechatResponse) {
	if req.Request.Code == "" || req.Request.EmployeeID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" //参数错误
		return
	}
}

type RegisterCompanyAdminRequest struct {
	apiobj.BaseRequest
	Request struct {
		Way string `json:"way"`
		// Code     string         `json:"code"`
		UserInfo    *user.UserInfo       `json:"user_info"`
		CompanyInfo *company.CompanyInfo `json:"company_info"`
		Issuer      string               `json:"issuer"`
	}
}

type GetCompanyAdminsResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data []*employee.AdminEmpItem
	}
}
