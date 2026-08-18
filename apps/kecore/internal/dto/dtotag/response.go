package dtotag

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type CreateTagGroupResponse struct {
	apiobj.BaseResponse
	Response CreateTagGroupEmbedResponse `json:"response"`
}

type CreateTagGroupEmbedResponse struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
}

type ModifyTagGroupResponse struct {
	apiobj.BaseResponse
	Response ModifyTagGroupEmbedResponse `json:"response"`
}
type ModifyTagGroupEmbedResponse struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
}

type DeleteTagGroupResponse struct {
	apiobj.BaseResponse
	Response DeleteTagGroupEmbedResponse `json:"response"`
}
type DeleteTagGroupEmbedResponse struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
}

type ListTagGroupResponse struct {
	apiobj.BaseResponse
	Response ListTagGroupEmbedResponse `json:"response"`
}
type ListTagGroupEmbedResponse struct {
	apiobj.QueryResponse
	// List 标签分组列表
	List []ListTagGroupItem `json:"list"`
}

type ListTagGroupItem struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
	// Name 标签分组名称
	Name string `json:"name"`
	// CreateAt 创建时间(unix时间戳)
	CreateAt int64 `json:"create_at"`
}

type ListTagResponse struct {
	apiobj.BaseResponse
	Response ListTagEmbedResponse `json:"response"`
}
type ListTagEmbedResponse struct {
	apiobj.QueryResponse
	// List 标签列表
	List []ListTagItem `json:"list"`
}

type ListTagItem struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
	// TagGroupName 标签分组名称
	TagGroupName string `json:"tag_group_name"`
	// TagID 标签ID
	TagID uint `json:"tag_id"`
	// Name 标签名称
	Name string `json:"name"`
	// CreateAt 创建时间(unix时间戳)
	CreateAt int64 `json:"create_at"`
}

type CreateTagResponse struct {
	apiobj.BaseResponse
	Response CreateTagEmbedResponse `json:"response"`
}
type CreateTagEmbedResponse struct {
	// TagID 标签ID
	TagID uint `json:"tag_id"`
}

type ModifyTagResponse struct {
	apiobj.BaseResponse
	Response ModifyTagEmbedResponse `json:"response"`
}
type ModifyTagEmbedResponse struct {
	// TagID 标签ID
	TagID uint `json:"tag_id"`
}

type DeleteTagResponse struct {
	apiobj.BaseResponse
	Response DeleteTagEmbedResponse `json:"response"`
}
type DeleteTagEmbedResponse struct {
	// TagID 标签ID
	TagID uint `json:"tag_id"`
}

type SetResourceTagResponse struct {
	apiobj.BaseResponse
	Response SetResourceTagEmbedResponse `json:"response"`
}
type SetResourceTagEmbedResponse struct {
}

type GetTagTreeResponse struct {
	apiobj.BaseResponse
	Response GetTagTreeEmbedResponse `json:"response"`
}
type GetTagTreeEmbedResponse struct {
	// RecentTagList 最近使用的标签
	RecentTagList []TagTreeListTagItem `json:"recent_tag_list"`
	// GroupList 标签分组列表
	GroupList []TagTreeListGroupItem `json:"group_list"`
}

type TagTreeListGroupItem struct {
	// TagGroupID 标签分组ID
	TagGroupID uint `json:"tag_group_id"`
	// TagGroupName 标签分组名称
	TagGroupName string `json:"tag_group_name"`
	// TagList 标签列表
	TagList []TagTreeListTagItem `json:"tag_list"`
}

type TagTreeListTagItem struct {
	// TagID 标签ID
	TagID uint `json:"tag_id"`
	// TagName 标签名称
	TagName string `json:"tag_name"`
}
