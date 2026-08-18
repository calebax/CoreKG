package ragtypes

import (
	"time"
)

// Common 结构体
type Common struct {
	ID        string    `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ForestID  uint      `json:"forest_id"`  // 分块所属的森林ID
	CompanyID uint      `json:"company_id"` // 公司ID
	Uin       uint      `json:"uin"`        // 用户ID
	Type      ChunkType `json:"type"`
	Version   string    `json:"version"` // 版本号

	//SourceFrom
	SourceFrom SourceFromType `json:"source_from,omitempty"`
	//	Enable 是否启用 1: 启用, -1: 禁用
	Enable int `json:"enable"`
}

// Source ES结构体
func (Common) Source() []string {
	return []string{"id", "created_at", "updated_at", "forest_id", "company_id", "type", "version", "source_from", "enable"}
}

// FQA 问答对结构体， 从Chunk中拆分
type FQA struct {
	Common     `json:",inline"`
	Embedding  Embedding `json:"embedding,omitempty"`
	QAQuestion string    `json:"qa_question,omitempty"`  // 问答问题
	QAAnswer   string    `json:"qa_answer,omitempty"`    // 问答答案
	QALable    []string  `json:"qa_lable,omitempty"`     // 问答标签
	QAMainID   string    `json:"qa_main_id,omitempty"`   // 问答问题ID
	QAAnswerID string    `json:"qa_answer_id,omitempty"` // 问答答案ID
}

// Source 问答对结构体的数据源
func (c FQA) Source() []string {
	return append(c.Common.Source(), "embedding", "qa_question", "qa_answer", "qa_lable", "qa_main_id", "qa_answer_id")
}

// ChunkInfo 分块信息结构体，从Chunk中拆分
type ChunkInfo struct {
	Common    `json:",inline"`
	Embedding Embedding `json:"embedding,omitempty"`
	// 分块的token数量
	Tokens int `json:"tokens,omitempty"`

	FileID      uint     `json:"file_id"`
	Location    Location `json:"location,omitempty"`
	Sequence    int      `json:"sequence,omitempty"`    // 分块在文件中的顺序
	Description string   `json:"description,omitempty"` // 分块的描述信息
	ImageUrl    string   `json:"image_url,omitempty"`   // 分块的图片URL
	Formula     string   `json:"formula,omitempty"`     // 分块的公式

	TitleLevel    []string `json:"title_level,omitempty"`     // 标题层级
	TitleLevelIDs []string `json:"title_level_ids,omitempty"` // 标题层级ID
}

// Source 分块信息结构体的数据源
func (c ChunkInfo) Source() []string {
	return append(c.Common.Source(), "embedding", "tokens", "file_id", "location", "sequence", "description", "image_url", "formula", "title_level", "title_level_ids")
}

// Entity 实体结构体，从Chunk中拆分
type Entity struct {
	Common     `json:",inline"`
	Embedding  Embedding        `json:"embedding,omitempty"`
	References []ChunkReference `json:"references,omitempty"` // 分块的描述信息
}

// Source 实体结构体的数据源
func (c Entity) Source() []string {
	return append(c.Common.Source(), "embedding", "references")
}

// Relationship 关系结构体，从Chunk中拆分
type Relationship struct {
	Common     `json:",inline"`
	Embedding  Embedding        `json:"embedding,omitempty"`
	References []ChunkReference `json:"references,omitempty"` // 分块的描述信息
}

// Source 关系结构体的数据源
func (c Relationship) Source() []string {
	return append(c.Common.Source(), "embedding", "references")
}

// FileDescription 文件简述结构体，从Chunk中拆分
type FileDescription struct {
	Common `json:",inline"`
	FileID uint `json:"file_id"`
	// Description 文件智能摘要的简短描述
	Description string `json:"description,omitempty"`
	// Embedding 文件智能摘要的简短描述的向量
	Embedding Embedding `json:"embedding,omitempty"`
	// MindMap 文件的思维导图
	MindMap string `json:"mind_map,omitempty"` // 思维导图
	// Abstract 文件内容的智能摘要
	Abstract string `json:"abstract,omitempty"` // 摘要
	// Questions 文档的推荐问题
	Questions []string `json:"questions,omitempty"`
}

// Source 文件简述结构体的数据源
func (c FileDescription) Source() []string {
	return append(c.Common.Source(), "file_id", "description", "embedding", "mind_map", "abstract")
}
