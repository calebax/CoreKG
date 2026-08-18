package forest

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/types"
)

// ForestInfoItem 知识森林信息
type ForestInfoItemList struct {
	apiobj.QueryResponse
	Data []*ForestInfoItemstruct
}

// ForestInfoItem 知识森林信息
type ForestInfoItemstruct struct {
	foresttype.KnownowForest
	IsAdmin     bool            `json:"is_admin"`
	FileCount   int64           `json:"file_count"`
	TotalSize   int64           `json:"total_size"`
	DiskStorage string          `json:"disk_storage"`
	ManagerIDs  types.UintArray `json:"manager_ids"`
	ScopeIDs    types.UintArray `json:"scope_ids"`
	AppCount    uint            `json:"app_count"`
	IsSynced    bool            `json:"is_synced"`
}

// WithPerm forest with perm ids
type WithPerm struct {
	foresttype.KnownowForest
	ManagerIDs  types.UintArray `json:"manager_ids"`
	ScopeIDs    types.UintArray `json:"scope_ids"`
	IsAdmin     bool            `json:"is_admin"`
	FileCount   int64           `json:"file_count"`
	TotalSize   int64           `json:"total_size"`
	DiskStorage string          `json:"disk_storage"`
}

type GetNameByIDsRes struct {
	// NameList 名称列表
	NameList []GetNameByIDsNameListItem `json:"name_list"`
}

type GetNameByIDsNameListItem struct {
	// ID 主键 id
	ID uint `json:"id"`
	// Module 模块名称
	Module foresttype.ForestModule
	// Name 名称
	Name string `json:"name"`
}
