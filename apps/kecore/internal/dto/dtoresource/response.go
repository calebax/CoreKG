package dtoresource

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type SetResourceScopeResponse struct {
	apiobj.BaseResponse
	Response SetResourceScopeEmbedResponse `json:"response"`
}

type SetResourceScopeEmbedResponse struct {
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id"`
	// ResourceIDStr 资源ID字符串形式
	ResourceIDStr string `json:"resource_id_str"`
}

type GetResourceScopeResponse struct {
	apiobj.BaseResponse
	Response GetResourceScopeEmbedResponse `json:"response"`
}
type GetResourceScopeEmbedResponse struct {
	// ResourceScopeList 资源权限范围列表
	ResourceScopeList []ResourceScopeItem `json:"resource_scope_list"`
	// UinList 资源权限涉及的用户UIN列表
	UinList []UinListItem `json:"uin_list"`
}

type ResourceScopeItem struct {
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

type UinListItem struct {
	// Uin 用户UIN
	Uin uint `json:"uin"`
	// Name 用户名称
	Name string `json:"name"`
}
