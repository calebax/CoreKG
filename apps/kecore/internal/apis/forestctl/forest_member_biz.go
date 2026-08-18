package forestctl

import (
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ListForestPublicScopeRequest
type ListForestPublicScopeRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestId uint `json:"forest_id"`
		apiobj.PageQuery
	}
}

func (req *ListForestPublicScopeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "scope_type", "created_at", "updated_at",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "scope_type", "created_at", "updated_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field_data" // 查询条件中的字段不存在, {{.field}}
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ListForestPublicScopeResponse
type ListForestPublicScopeResponse struct {
	apiobj.BaseResponse
	Response forest.QueryForestPublicScopeListResponse
}

// UpdateForestPublicScopeRequest
type UpdateForestPublicScopeRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestId  uint                 `json:"forest_id"`
		ScopeType foresttype.ScopeType `json:"scope_type"`
		ScopeIDs  []uint               `json:"scope_ids"`
	}
}

// Validity 验证有效性
func (req *UpdateForestPublicScopeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ForestId == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 知识库id不能为空
		return
	}
	if req.Request.ScopeType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_public_scope_empty" // 公开范围不能为空
		return
	}
	if req.Request.ScopeIDs == nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_scope_ids_empty" // 范围id不能为空
		return
	}
}

// UpdateForestPublicScopeResponse
type UpdateForestPublicScopeResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// UpdateForestManagerRequest
type UpdateForestManagerRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestId    uint                   `json:"forest_id"`
		PublicScope foresttype.PublicScope `json:"public_scope"`
		ManagerIDs  types.UintArray        `json:"manager_ids"`
	}
}

func (req *UpdateForestManagerRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ForestId == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 知识库id不能为空
		return
	}
	if req.Request.PublicScope == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_public_scope_empty" // 公开范围不能为空
		return
	}
	if req.Request.ManagerIDs == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_manager_ids_empty" // 管理员id不能为空
		return
	}
}

// UpdateForestManagerResponse
type UpdateForestManagerResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
