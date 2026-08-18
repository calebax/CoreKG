package ragtypes

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type ChunkType = string

// ChunkType 定义分块的类型
const (
	ChunkTypeChunk           ChunkType = "chunk"
	ChunkTypeEntity          ChunkType = "entity"
	ChunkTypeRelationship    ChunkType = "relationship"
	ChunkTypeFQA             ChunkType = "FQA"
	ChunkTypeFileDescription ChunkType = "file_description"
	ChunkTypeImage           ChunkType = "image"
	ChunkTypeVideo           ChunkType = "video"
	ChunkTypeTable           ChunkType = "table"
)

// Embedding 定义分块的向量表示
type Embedding []float32

func (e Embedding) String() string {
	tmp := []string{}
	for i, v := range e {
		if i >= 2 {
			break
		}
		tmp = append(tmp, fmt.Sprintf("%v", v))
	}
	str := fmt.Sprintf("[%v]{%v,。。。}", len(e), strings.Join(tmp, ","))
	return str
}

// Location 定义分块在文件中的位置
type Location [5]int

// Chunk 文本分块结构体
type Chunk struct {
	// ID        string `json:"-"`          // 分块ID
	ForestID  uint `json:"forest_id"`  // 分块所属的森林ID
	CompanyID uint `json:"company_id"` // 公司ID
	Uin       uint `json:"uin"`        // 用户ID
	// 分块类型，例如 "chunk", "entity", "relationship" 等
	Type    ChunkType `json:"type"`
	Version string    `json:"version"`
	Level   int       `json:"level,omitempty"` // 分块的层级

	// 分块的内容
	Content string `json:"content,omitempty"`

	ContentSource string `json:"content_source,omitempty"`
	ContentTarget string `json:"content_target,omitempty"`

	Embedding Embedding `json:"embedding,omitempty"` // 分块的向量表示
	// 分块的token数量
	Tokens int `json:"tokens,omitempty"`

	FileID          uint     `json:"file_id"`
	Location        Location `json:"location,omitempty"`
	Sequence        int      `json:"sequence,omitempty"`         // 分块在文件中的顺序
	Description     string   `json:"description,omitempty"`      // 分块的描述信息
	DescriptionHash string   `json:"description_hash,omitempty"` // 分块的描述信息hash
	ImageUrl        string   `json:"image_url,omitempty"`        // 分块的图片URL
	Formula         string   `json:"formula,omitempty"`          // 分块的公式

	MindMap   string   `json:"mind_map,omitempty"`  // 思维导图
	Abstract  string   `json:"abstract,omitempty"`  // 摘要
	Questions []string `json:"questions,omitempty"` // 问题

	QAQuestion string   `json:"qa_question,omitempty"`  // 问答问题
	QAAnswer   string   `json:"qa_answer,omitempty"`    // 问答答案
	QALable    []string `json:"qa_lable,omitempty"`     // 问答标签
	QAMainID   string   `json:"qa_main_id,omitempty"`   // 问答问题ID
	QAAnswerID string   `json:"qa_answer_id,omitempty"` // 问答答案ID

	TitleLevel    []string  `json:"title_level,omitempty"`     // 标题层级
	TitleLevelIDs []string  `json:"title_level_ids,omitempty"` // 标题层级ID
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`

	References []ChunkReference `json:"references,omitempty"` // 分块的描述信息

	ChunkSize      int       `json:"chunk_size,omitempty"`      // 分块的大小
	YGLocation     string    `json:"yg_location,omitempty"`     // 新位置
	ImageEmbedding Embedding `json:"image_embedding,omitempty"` // 图片向量表示

	//SourceFrom
	SourceFrom SourceFromType `json:"source_from,omitempty"`
	FileName   string         `json:"file_name,omitempty"`

	Table string `json:"table,omitempty"` // 分块所属的表格，JSON格式
}

type SourceFromType string

var (
	SourceFromTypeManualImport SourceFromType = "manul_import"
	SourceFromTypeBatchUpload  SourceFromType = "batch_upload"
)

// ChunkReference 定义分块的引用信息
type ChunkReference struct {
	// 分块所属的文件ID
	FileID         uint   `json:"file_id"`
	Description    string `json:"description,omitempty"`     // 分块的描述信息
	ChunkID        string `json:"chunk_id,omitempty"`        // 分块的ID
	RelationshipID string `json:"relationship_id,omitempty"` // 关系ID，可能用于实体关系图
}

// ToMapping 返回分块的映射信息
func (c Chunk) ToMapping() string {
	full := map[string]interface{}{
		"mappings": c.mappings(),
		"settings": c.settings(),
	}

	data, err := json.Marshal(full)
	if err != nil {
		panic(fmt.Errorf("failed to marshal chunk mapping: %w", err))
	}
	return string(data)
}

func (c Chunk) settings() map[string]interface{} {
	return map[string]interface{}{
		"index": map[string]interface{}{
			"number_of_shards":   1, // 分片数
			"number_of_replicas": 1, // 副本数
		},
	}
}

func (c Chunk) mappings() map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"forest_id": map[string]interface{}{
				"type": "integer",
			},
			"company_id": map[string]interface{}{
				"type": "integer",
			},
			"uin": map[string]interface{}{
				"type": "integer",
			},
			"type": map[string]interface{}{
				"type": "keyword",
			},
			"version": map[string]interface{}{
				"type": "keyword",
			},
			"content": map[string]interface{}{
				"type":            "text",
				"analyzer":        "ik_max_word",
				"search_analyzer": "ik_smart",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
				},
			},
			"content_source": map[string]interface{}{
				"type":            "text",
				"analyzer":        "ik_max_word",
				"search_analyzer": "ik_smart",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
				},
			},
			"content_target": map[string]interface{}{
				"type":            "text",
				"analyzer":        "ik_max_word",
				"search_analyzer": "ik_smart",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
				},
			},
			"created_at": map[string]interface{}{
				"type": "date",
			},
			"updated_at": map[string]interface{}{
				"type": "date",
			},
			"embedding": map[string]interface{}{
				"type":  "dense_vector",
				"dims":  1024, // 假设向量维度为1536，根据实际情况调整
				"index": true, // 启用索引以支持向量搜索
			},
			"tokens": map[string]interface{}{
				"type": "integer",
			},
			"mind_map": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
				},
			},
			"abstract": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
				},
			},
			"questions": map[string]interface{}{
				"type": "keyword",
			},
			"file_id": map[string]interface{}{
				"type": "integer",
			},
			"location": map[string]interface{}{
				"type":  "integer",
				"index": false, // 不需要索引位置数组
			},
			"sequence": map[string]interface{}{
				"type": "integer",
			},
			"description": map[string]interface{}{
				"type":            "text",
				"analyzer":        "ik_max_word",
				"search_analyzer": "ik_smart",
			},
			"image_url": map[string]interface{}{
				"type": "keyword",
			},
			"formula": map[string]interface{}{
				"type": "keyword",
			},
			"references": map[string]interface{}{
				"type": "nested", // 使用嵌套类型以支持数组中的对象
				"properties": map[string]interface{}{
					"file_id": map[string]interface{}{
						"type": "integer",
					},
					"description": map[string]interface{}{
						"type":            "text",
						"analyzer":        "ik_max_word",
						"search_analyzer": "ik_smart",
					},
					"chunk_id": map[string]interface{}{
						"type": "keyword", // 使用 keyword 类型以支持精确匹配
					},
				},
			},
		},
		"dynamic": "strict", // 严格限制只能有这几个字段
	}
}

// WriteUpdateDSL 将分块转换为更新DSL格式
func (c *Chunk) WriteUpdateDSL(w io.Writer) error {
	// if c.ID == "" {
	// 	return fmt.Errorf("chunk ID is required for WriteUpdateDSL")
	// }
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	// meta := []byte(fmt.Sprintf(`{ "update": { "_id": "%s" } }%s`, c.ID, "\n"))
	// if _, err := w.Write(meta); err != nil {
	// 	return fmt.Errorf("failed to write metadata: %w", err)
	// }
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline after chunk data: %w", err)
	}

	return nil
}
