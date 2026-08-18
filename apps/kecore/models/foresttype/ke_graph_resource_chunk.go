package foresttype

import (
	"gorm.io/gorm"
)

type KeGraphResourceChunkType string

const (
	KeGraphResourceChunkTypeNode KeGraphResourceChunkType = "node"
	KeGraphResourceChunkTypeEdge KeGraphResourceChunkType = "edge"
)

// KeGraphResourceChunk chunk 和图谱节点、边的关系表结构体
type KeGraphResourceChunk struct {
	gorm.Model
	ChunkID        string                   `gorm:"column:chunk_id;type:varchar(64);not null;;comment:chunk id"`
	GraphID        uint                     `gorm:"column:graph_id;type:bigint;not null;default 0;comment:图谱id"`
	GraphVersionID uint                     `gorm:"column:graph_version_id;type:bigint;not null;default 0;comment:图谱版本id"`
	ResourceID     uint                     `gorm:"column:resource_id;type:bigint;not null;default 0;comment:资源id"`
	ResourceType   KeGraphResourceChunkType `gorm:"column:resource_type;type:varchar(24);not null;;comment:资源类型，node：图谱节点，edge：图谱边"`
}

type KeGraphResourceChunkList []KeGraphResourceChunk

func (KeGraphResourceChunk) TableName() string {
	return TableNameKeGraphResourceChunk
}

func (l KeGraphResourceChunkList) ToMap() map[uint]KeGraphResourceChunk {
	m := make(map[uint]KeGraphResourceChunk)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}

// DeduplicateByChunkID 原地去重，根据 chunk id 去重
func (list *KeGraphResourceChunkList) DeduplicateByChunkID() {
	seen := make(map[string]struct{}) // 使用 struct{} 作为值，节省空间
	j := 0                            // j 指向去重后切片的下一个写入位置
	for i, chunk := range *list {
		if _, exists := seen[chunk.ChunkID]; !exists {
			seen[chunk.ChunkID] = struct{}{}
			(*list)[j] = (*list)[i]
			j++
		}
	}
	// 截取切片，保留去重后的部分
	*list = (*list)[:j]
}
