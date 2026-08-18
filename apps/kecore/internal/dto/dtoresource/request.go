package dtoresource

import (
	"strconv"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type SetResourceScopeRequest struct {
	apiobj.BaseRequest
	Request SetResourceScopeEmbedRequest `json:"request"`
}

type SetResourceScopeEmbedRequest struct {
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id"`
	// ResourceIDStr 资源ID字符串形式
	ResourceIDStr string `json:"resource_id_str"`
	// ViewScopeType 查看权限范围的作用域类型
	ViewScopeType foresttype.ScopeType `json:"view_scope_type"`
	// ViewScopeIDs 查看权限范围的作用域ID列表
	ViewScopeIDs []uint `json:"view_scope_ids"`
	// 管理权限范围的作用域ID列表
	ManageScopeIDs []uint `json:"manage_scope_ids"`
}

func (opt *SetResourceScopeRequest) Validity(resp *SetResourceScopeResponse) {
	if opt.Request.ResourceID == 0 && opt.Request.ResourceIDStr == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_resource_id_required"
		return
	}
	if opt.Request.ResourceIDStr != "" {
		// 优先使用字符串形式的 ResourceID
		resourceID, _ := strconv.ParseUint(opt.Request.ResourceIDStr, 10, 64)
		opt.Request.ResourceID = uint(resourceID)
	}
	if opt.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_resource_type_required"
		return
	}
	if _, ok := foresttype.ResourceTypeMap[opt.Request.ResourceType]; !ok {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_invalid_resource_type"
		return
	}
	if len(opt.Request.ManageScopeIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_no_manage_scope_ids"
		return
	}
	switch opt.Request.ViewScopeType {
	case foresttype.ScopeTypeUser:
		if len(opt.Request.ViewScopeIDs) == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_permission_no_view_scope_ids"
			return
		}
	case foresttype.ScopeTypeCompany:
		// 公司类型不需要校验 ViewScopeIDs
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_invalid_view_scope_type"
		return
	}
}

type GetResourceScopeRequest struct {
	apiobj.BaseRequest
	Request GetResourceScopeEmbedRequest `json:"request"`
}
type GetResourceScopeEmbedRequest struct {
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// ResourceIDs 资源 ID 列表
	ResourceIDs []uint `json:"resource_ids"`
	// ResourceIDStrs 资源 ID 字符串形式列表
	ResourceIDStrs []string `json:"resource_id_strs"`
}

func (opt *GetResourceScopeRequest) Validity(resp *GetResourceScopeResponse) {
	if len(opt.Request.ResourceIDs) == 0 && len(opt.Request.ResourceIDStrs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_no_resource_ids"
		return
	}
	if len(opt.Request.ResourceIDStrs) > 0 {
		// 优先使用字符串形式的 ResourceID 列表
		opt.Request.ResourceIDs = make([]uint, 0, len(opt.Request.ResourceIDStrs))
		for _, idStr := range opt.Request.ResourceIDStrs {
			resourceID, _ := strconv.ParseUint(idStr, 10, 64)
			opt.Request.ResourceIDs = append(opt.Request.ResourceIDs, uint(resourceID))
		}
	}
	if opt.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_resource_type_required"
		return
	}
	if _, ok := foresttype.ResourceTypeMap[opt.Request.ResourceType]; !ok {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_permission_invalid_resource_type"
		return
	}
}
