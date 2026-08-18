package dtolikes

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListLikesResponse struct {
	apiobj.BaseResponse
	Response ListLikesEmbedResponse `json:"response"`
}

type ResourceTag struct {
	// ID 资源标签ID
	ID uint `json:"id"`
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id"`
	// ResourceType 资源类型
	ResourceType foresttype.TagResourceType `json:"resource_type"`
	// TagID 标签ID
	TagID uint `json:"tag_id"`
	// TagName 标签名称
	TagName string `json:"tag_name"`
	// TagGroupName 标签组名称
	TagGroupName string `json:"tag_group_name"`
}

type LiItem struct {
	// ID 资源ID
	ID uint `json:"id"`
	// CreatedAt 创建时间(unix时间戳)
	CreatedAt int64 `json:"created_at"`
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id"`
	// ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
	// ResourceName 资源名称
	ResourceName string `json:"resource_name"`
	// ForestID 知识森林ID
	ForestID uint `json:"forest_id"`
	// ForestName 知识森林名称
	ForestName string `json:"forest_name"`
	// TagList 标签列表
	TagList []ResourceTag `json:"tag_list"`
	// FileConfig 分段规则
	FileConfig foresttype.FileConfig `json:"file_config"`
}

type ListLikesEmbedResponse struct {
	apiobj.QueryResponse
	// Data 资源列表
	Data []LiItem `json:"data"`
}

type MarkResourceLikeResponse struct {
	apiobj.BaseResponse
	Response MarkResourceLikeEmbedResponse `json:"response"`
}
type MarkResourceLikeEmbedResponse struct {
}
