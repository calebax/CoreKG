package dtotag

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateTagGroupRequest struct {
	apiobj.BaseRequest
	Request CreateTagGroupEmbedRequest `json:"request"`
}

type CreateTagGroupEmbedRequest struct {
	// Name 标签分组名称
	Name string `json:"name" validate:"required"`
}

func (opt *CreateTagGroupRequest) Validity(resp *CreateTagGroupResponse) {
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_group_name_empty"
		return
	}
}

type ModifyTagGroupRequest struct {
	apiobj.BaseRequest
	Request ModifyTagGroupEmbedRequest `json:"request"`
}
type ModifyTagGroupEmbedRequest struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id" validate:"required"`
	// Name 标签分组名称
	Name string `json:"name" validate:"required"`
}

func (opt *ModifyTagGroupRequest) Validity(resp *ModifyTagGroupResponse) {
	if opt.Request.TagGroupID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_group_id_empty"
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_group_name_empty"
		return
	}
}

type DeleteTagGroupRequest struct {
	apiobj.BaseRequest
	Request DeleteTagGroupEmbedRequest `json:"request"`
}
type DeleteTagGroupEmbedRequest struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id" validate:"required"`
}

func (opt *DeleteTagGroupRequest) Validity(resp *DeleteTagGroupResponse) {
	if opt.Request.TagGroupID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_group_id_empty"
		return
	}
}

type ListTagGroupRequest struct {
	apiobj.BaseRequest
	Request ListTagGroupEmbedRequest `json:"request"`
}
type ListTagGroupEmbedRequest struct {
	apiobj.PageQuery
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
	// Name 标签分组名称
	Name string `json:"name"`
}

func (opt *ListTagGroupRequest) Validity(resp *ListTagGroupResponse) {
}

type ListTagRequest struct {
	apiobj.BaseRequest
	Request ListTagEmbedRequest `json:"request"`
}
type ListTagEmbedRequest struct {
	apiobj.PageQuery
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
	// Name 标签名称
	Name string `json:"name"`
}

func (opt *ListTagRequest) Validity(resp *ListTagResponse) {
}

type CreateTagRequest struct {
	apiobj.BaseRequest
	Request CreateTagEmbedRequest `json:"request"`
}
type CreateTagEmbedRequest struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id" validate:"required"`
	// Name 标签名称
	Name string `json:"name" validate:"required"`
}

func (opt *CreateTagRequest) Validity(resp *CreateTagResponse) {
	if opt.Request.TagGroupID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_group_id_empty"
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_name_empty"
		return
	}
}

type ModifyTagRequest struct {
	apiobj.BaseRequest
	Request ModifyTagEmbedRequest `json:"request"`
}
type ModifyTagEmbedRequest struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id" validate:"required"`
	// TagID 标签ID
	TagID uint `json:"tag_id" validate:"required"`
	// Name 标签名称
	Name string `json:"name" validate:"required"`
}

func (opt *ModifyTagRequest) Validity(resp *ModifyTagResponse) {
	if opt.Request.TagGroupID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_group_id_empty"
		return
	}
	if opt.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_id_empty"
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_name_empty"
		return
	}
}

type DeleteTagRequest struct {
	apiobj.BaseRequest
	Request DeleteTagEmbedRequest `json:"request"`
}
type DeleteTagEmbedRequest struct {
	// TagID 标签ID
	TagID uint `json:"tag_id" validate:"required"`
}

func (opt *DeleteTagRequest) Validity(resp *DeleteTagResponse) {
	if opt.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_id_empty"
		return
	}
}

type SetResourceTagRequest struct {
	apiobj.BaseRequest
	Request SetResourceTagEmbedRequest `json:"request"`
}
type SetResourceTagEmbedRequest struct {
	// ResourceType 资源类型,file:文件
	ResourceType foresttype.TagResourceType `json:"resource_type" validate:"required"`
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id" validate:"required"`
	// TagIDs 标签ID列表
	TagIDs []uint `json:"tag_ids" validate:"required"`
}

func (opt *SetResourceTagRequest) Validity(resp *SetResourceTagResponse) {
	if opt.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_resource_type_empty"
		return
	}
	if opt.Request.ResourceType != foresttype.TagResourceTypeFile {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_resource_type_invalid"
		return
	}
	if opt.Request.ResourceID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_resource_id_empty"
		return
	}
}

type GetTagTreeRequest struct {
	apiobj.BaseRequest
	Request GetTagTreeEmbedRequest `json:"request"`
}
type GetTagTreeEmbedRequest struct {
}

func (opt *GetTagTreeRequest) Validity(resp *GetTagTreeResponse) {
}
