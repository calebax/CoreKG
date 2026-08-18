package chunk

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/dbtools/estool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

const DefaultIndex = "ke_0"

// InitESClient 初始化es客户端s
func InitESClient(ctx context.Context) error {
	// cfg := config.ESConfig{
	// 	Addresses:     []string{"http://example.com:53082/"},
	// 	SlowThreshold: time.Millisecond,
	// 	Username:      "elastic",
	// 	Password:      "CHANGE_ME_PASSWORD",
	// }
	cfg := config.ESConfig{}
	err := settings.GetYaml("knowledge", "es", &cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "get es config failed: %s", err)
		return err
	}
	client, err := estool.InitES(cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "init es client failed: %s", err)
		return err
	}
	escli = client
	return nil
}

type Chunk struct {
	ID     string     `json:"_id"`
	Score  float64    `json:"_score"`
	Source *ChunkInfo `json:"_source"`
}

// Chunk 文本分块结构体
type ChunkInfo struct {
	// ID        string `json:"-"`          // 分块ID
	ForestID  uint `json:"forest_id"`  // 分块所属的森林ID
	CompanyID uint `json:"company_id"` // 公司ID
	Uin       uint `json:"uin"`        // 用户ID
	// 分块类型，例如 "chunk", "entity", "relationship" 等
	Type string `json:"type"`

	Level int `json:"level,omitempty"` // 分块的层级

	Embedding ragtypes.Embedding `json:"embedding,omitempty"` // 分块的向量表示
	// 分块的token数量
	Tokens int `json:"tokens,omitempty"`

	FileID          uint     `json:"file_id"`
	Location        Location `json:"location,omitempty"`
	Sequence        int      `json:"sequence,omitempty"`         // 分块在文件中的顺序
	Description     string   `json:"description,omitempty"`      // 分块的描述信息
	DescriptionHash string   `json:"description_hash,omitempty"` // 分块的描述信息hash
	ImageUrl        string   `json:"image_url,omitempty"`        // 分块的图片URL
	Formula         string   `json:"formula,omitempty"`          // 分块的公式

	TitleLevel    []string `json:"title_level,omitempty"`     // 标题层级
	TitleLevelIDs []string `json:"title_level_ids,omitempty"` // 标题层级ID

	ChunkSize      int                `json:"chunk_size,omitempty"`      // 分块的大小
	YGLocation     string             `json:"yg_location,omitempty"`     // 新位置
	ImageEmbedding ragtypes.Embedding `json:"image_embedding,omitempty"` // 图片向量表示

	ISDisable bool `json:"is_disable"` // 是否禁用

	Table string `json:"table,omitempty"` // 分块所属的表格，JSON格式
}

func (c ChunkInfo) String() string {
	ctx := context.TODO()
	jsonPayload, err := json.Marshal(c)
	if err != nil {
		logs.ErrorContextf(ctx, "[ChunkInfo] Failed to convert request body to JSON: %s", err.Error())
		return ""
	}
	return string(jsonPayload)
}

// Location 定义分块在文件中的位置
type Location [5]int

// createRespnse 创建问题返回值获取id结构
type createRespnse struct {
	ID string `json:"_id"`
}

// ChunkSearchResult es 查询问题结构体
type ChunkSearchResult struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		MaxScore float64 `json:"max_score"`
		Total    struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []*Chunk `json:"hits"`
	} `json:"hits"`
}
