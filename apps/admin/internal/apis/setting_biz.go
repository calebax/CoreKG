package apis

import (
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/settings"
)

type ListSettingRequest apiobj.QueryRequest

func (req *ListSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "key", "name", "group", "created_at", "updated_at",
			"key desc", "name desc", "group desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "orderBy字段不支持"
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "group", "key":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的字段只能有一个值"
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的值不能为空"
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "查询条件中的字段不存在, " + v.Field
			return
		}
	}
}

type ListSettingResponse struct {
	apiobj.BaseResponse

	Response employee.QuerySettingListResponse
}

type CreateSettingRequest struct {
	apiobj.BaseRequest

	Request settings.SettingItem
}

func (req *CreateSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Group == "" || req.Request.Key == "" || req.Request.Name == "" ||
		req.Request.Value == "" || req.Request.ValueType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "group、key、name、value、value_type不能为空"
		return
	}

	switch req.Request.ValueType {
	case settings.ValueSecret, settings.ValuePassword, settings.ValueText, settings.ValueInt64, settings.ValueBool,
		settings.ValueJSON, settings.ValueYaml:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "value_type字段不支持"
		return
	}
}

type CreateSettingResponse struct {
	apiobj.BaseResponse

	Response struct{}
}

type UpdateSettingRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID        uint               `json:"id"`
		Value     string             `json:"value"`
		Name      string             `json:"name"`
		Describe  string             `json:"describe"`
		ValueType settings.ValueType `json:"value_type"`
		Default   string             `json:"default"`
	}
}

func (req *UpdateSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "id不能为空"
		return
	}
	if req.Request.Value == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "value不能为空"
		return
	}
}

type UpdateSettingResponse struct {
	apiobj.BaseResponse
}
