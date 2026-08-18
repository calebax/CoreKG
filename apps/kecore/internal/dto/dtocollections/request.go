package dtocollections

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type MarkResourceCollectionRequest struct {
	apiobj.BaseRequest
	Request MarkResourceCollectionEmbedRequest `json:"request"`
}

type MarkResourceCollectionEmbedRequest struct {
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id"`
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// Enable 是否启用
	Enable bool `json:"enable"`
}

func (opt *MarkResourceCollectionRequest) Validity(resp *MarkResourceCollectionResponse) {
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

type ListCollectionRequest struct {
	apiobj.BaseRequest
	Request ListCollectionEmbedRequest `json:"request"`
}
type ListCollectionEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListCollectionRequest) Validity(resp *ListCollectionResponse) {
}
