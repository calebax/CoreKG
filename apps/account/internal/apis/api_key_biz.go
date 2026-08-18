package apis

import (
	"time"

	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateAPIKeyRequest 创建API key 请求
type CreateAPIKeyRequest struct {
	apiobj.BaseRequest
	Request struct {
		Name      string     `json:"name"`
		Purpose   string     `json:"purpose"`
		ExpiredAt *time.Time `json:"expired_at"`
	}
}

// Validity 验证参数
func (req *CreateAPIKeyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
	if req.Request.ExpiredAt != nil && req.Request.ExpiredAt.Before(time.Now()) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_expired_time" // 过期时间无效
		return
	}
}

// CreateAPIKeyResponse 创建API key 响应
type CreateAPIKeyResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// ListAPIKeyRequest 查询API key 请求
type ListAPIKeyRequest apiobj.QueryRequest

// Validity 验证参数
func (req *ListAPIKeyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "expired_at", "api_key", "created_at", "updated_at",
			"name desc", "expired_at desc", "api_key desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_orderby_field" // 排序字段无效
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
			resp.Message = "account_invalid_filter_field_data" // 查询条件中的字段不存在, " + v.Field
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ListAPIKeyResponse 查询API key 响应
type ListAPIKeyResponse struct {
	apiobj.BaseResponse
	Response apikey.QueryListAPIKeyResponse
}

// DeleteAPIKeyRequest 删除API key 请求
type DeleteAPIKeyRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

// Validity 验证参数
func (req *DeleteAPIKeyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
}

// DeleteAPIKeyResponse 删除API key 响应
type DeleteAPIKeyResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// AddAPIKeyPrivilegeRequest 为API key添加权限
type AddAPIKeyPrivilegeRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID           uint   `json:"id"`
		PrivilegeIDs []uint `json:"privilege_ids"`
	}
}

// Validity 验证参数
func (req *AddAPIKeyPrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
	if req.Request.PrivilegeIDs == nil || len(req.Request.PrivilegeIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
}

// AddAPIKeyPrivilegeResponse 为API key添加权限响应
type AddAPIKeyPrivilegeResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// ListAPIKeyPrivilegeRequest API key权限列表
type ListAPIKeyPrivilegeRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
		apiobj.PageQuery
	}
}

// Validity 验证参数
func (req *ListAPIKeyPrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
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
			resp.Message = "account_invalid_filter_field_data" // 查询条件中的字段不存在, " + v.Field
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ListAPIKeyPrivilegeResponse API key权限列表响应
type ListAPIKeyPrivilegeResponse struct {
	apiobj.BaseResponse
	Response apikey.QueryListAPIKeyPrivilegeResponse
}

// DeleteAPIKeyPrivilegeRequest API key删除权限请求
type DeleteAPIKeyPrivilegeRequest struct {
	apiobj.BaseRequest
	Request struct {
		KeyID        uint   `json:"key_id"`
		PrivilegeIDs []uint `json:"privilege_ids"`
	}
}

// Validity 验证参数
func (req *DeleteAPIKeyPrivilegeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.KeyID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
	if req.Request.PrivilegeIDs == nil || len(req.Request.PrivilegeIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
}

// DeleteAPIKeyPrivilegeResponse API key删除权限响应
type DeleteAPIKeyPrivilegeResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
