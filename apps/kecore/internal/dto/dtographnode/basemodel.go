package dtographnode

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/types"
)

type TagObject struct {
	// 实体类型id
	TagID uint `json:"tag_id,omitempty"`
	// 实体类型名
	TagName string `json:"tag_name,omitempty"`
	// 实体类型属性
	Properties foresttype.Properties `json:"properties,omitempty"`
	// 实体类型属性值
	PropertiesValues foresttype.PropertiesValues `json:"properties_values,omitempty"`
}

type EdgeObject struct {
	// 边id
	EdgeID uint `json:"edge_id,omitempty"`
	// 起点id
	SrcNodeID uint `json:"src_node_id,omitempty"`
	// 起点名称
	SrcNodeName string `json:"src_node_name,omitempty"`
	// 起点类型id
	SrcTagID uint `json:"src_tag_id,omitempty"`
	// 边名
	EdgeName string `json:"edge_name,omitempty"`
	// 终点id
	DstNodeID uint `json:"dst_node_id,omitempty"`
	// 终点名称
	DstNodeName string `json:"dst_node_name,omitempty"`
	// 终点类型id
	DstTagID uint `json:"dst_tag_id,omitempty"`
	// 文件id列表
	FileIDList types.UintArray `json:"file_id_list,omitempty"`
	// 分块id列表
	ChunkIDList types.StringArray `json:"chunk_id_list,omitempty"`
	// 边属性
	Properties foresttype.Properties `json:"properties,omitempty"`
	// 边属性值
	PropertiesValues foresttype.PropertiesValues `json:"properties_values,omitempty"`
}

type FileObject struct {
	// 文件id
	FileID uint `json:"file_id,omitempty"`
	// 文件名
	FileName string `json:"file_name,omitempty"`
	//文件访问地址
	FileURL string `json:"file_url,omitempty"`
	// 文件分块
	Chunks []ragtypes.Chunk `json:"chunks,omitempty"`
}
