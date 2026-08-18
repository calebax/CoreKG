package dtolikes

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListLikesRequest struct {
	apiobj.BaseRequest
	Request ListLikesEmbedRequest `json:"request"`
}

type ListLikesEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListLikesRequest) Validity(resp *ListLikesResponse) {
}

type MarkResourceLikeRequest struct {
	apiobj.BaseRequest
	Request MarkResourceLikeEmbedRequest `json:"request"`
}
type MarkResourceLikeEmbedRequest struct {
	// ResourceID 资源ID
	ResourceID   uint                    `json:"resource_id"`
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// Enable 是否启用
	Enable       bool                    `json:"enable"`
}

func (opt *MarkResourceLikeRequest) Validity(resp *MarkResourceLikeResponse) {
	if opt.Request.ResourceID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_resource_ids"
		return
	}
	if opt.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_resource_type"
		return
	}
}
