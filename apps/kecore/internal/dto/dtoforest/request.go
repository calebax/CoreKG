package dtoforest

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetResourceBaseInfoRequest struct {
	apiobj.BaseRequest
	Request GetResourceBaseInfoEmbedRequest
}

type GetResourceBaseInfoEmbedRequest struct {
	// ResourceName 资源名称
	ResourceName string `json:"resource_name" validate:"required"`
	// FileTagIDs 文件标签ID列表
	FileTagIDs []uint `json:"file_tag_ids"`
}

func (opt *GetResourceBaseInfoRequest) Validity(resp *GetResourceBaseInfoResponse) {
	// if opt.Request.ResourceName == "" {
	// 	resp.Code = errcode.ErrCode_BadRequest
	// 	resp.Message = "kecore_require_resource_name"
	// 	return
	// }
}

type SetResourceEnableRequest struct {
	apiobj.BaseRequest
	Request SetResourceEnableEmbedRequest
}
type SetResourceEnableEmbedRequest struct {
	//	ForestID 知识库ID
	ForestID uint `json:"forest_id"`
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// ResourceID 资源ID
	ResourceIDs []string `json:"resource_ids"`
	// Enable 是否启用 1: 启用, -1: 禁用
	Enable int `json:"enable"`
}

const (
	ResourceTypeFile    = "file"
	ResourceTypeExcel   = "excel"
	ResourceTypeMysqlDB = "mysql_db"
	ResourceTypeQAPair  = "qa_pair"
)

func (opt *SetResourceEnableRequest) Validity(resp *SetResourceEnableResponse) {
	if opt.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_forest_id"
		return
	}
	if len(opt.Request.ResourceIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_resource_id"
		return
	}
	switch opt.Request.ResourceType {
	case
		ResourceTypeFile,
		ResourceTypeExcel,
		ResourceTypeMysqlDB,
		ResourceTypeQAPair:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_resource_type"
		return
	}
}

type UpdateForestDescriptionRequest struct {
	apiobj.BaseRequest
	Request UpdateForestDescriptionEmbedRequest
}
type UpdateForestDescriptionEmbedRequest struct {
	// ForestID 知识库ID
	ForestID uint `json:"forest_id"`
	// Description 知识库描述
	Description string `json:"description"`
}

func (opt *UpdateForestDescriptionRequest) Validity(resp *UpdateForestDescriptionResponse) {
	if opt.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_forest_id"
		return
	}
}

type GetOriginResourceRequest struct {
	apiobj.BaseRequest
	Request GetResourceURLListEmbedRequest
}
type GetResourceURLListEmbedRequest struct {
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// ResourceIDs 资源ID列表
	ResourceIDs []uint `json:"resource_ids"`
}

func (opt *GetOriginResourceRequest) Validity(resp *GetOriginResourceResponse) {
	switch opt.Request.ResourceType {
	case foresttype.ResourceTypeForest:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_resource_type"
		return
	}
	if len(opt.Request.ResourceIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_require_resource_id"
		return
	}
}
