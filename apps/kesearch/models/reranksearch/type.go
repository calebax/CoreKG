package reranksearch

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

type SearchConfig struct {
	DescriptionWeight float32 `yaml:"description_weight" json:"description_weight"` // chunk权重
	EmbeddingWeight   float32 `yaml:"embedding_weight" json:"embedding_weight"`     // embedding权重
	EnabelAbstract    bool    `yaml:"enabel_abstract" json:"enabel_abstract"`       // 是否需要摘要
	EnableRerank      bool    `yaml:"enable_rerank" json:"enable_rerank"`           // 是否需要rerank
	Topn              int     `yaml:"topn" json:"topn"`                             // 每次rerank后保留的chunk数量
	Topm              int     `yaml:"topm" json:"topm"`                             // 整个检索过程最多返召回的chunk数量
	Topk              int     `yaml:"topk" json:"topk"`                             // 每次rerank后，如果使用过滤阈值，若无chunk，则默认提取前topk个
	NeighborSize      int     `yaml:"neighbor_size" json:"neighbor_size"`           // 邻居chunk数量
	RerankThreshold   float64 `yaml:"rerank_threshold" json:"rerank_threshold"`     // rerank 过滤阈值
	FetchFactor       int     `yaml:"fetch_factor" json:"fetch_factor"`             // 第一次检索 topn*fetch_factor的chunk
	FallBackToTopK    bool    `yaml:"fall_back_to_topk" json:"fall_back_to_topk"`   // 如果第一次 rerank 过滤后为空，则直接取 topk 返回
}

func (c *SearchConfig) Validity() error {
	if c.DescriptionWeight+c.EmbeddingWeight != 1 {
		return fmt.Errorf("description_weight + embedding_weight must be 1")
	}
	if c.Topn <= 0 {
		return fmt.Errorf("topn must be greater than 0")
	}
	if c.Topm <= 0 {
		return fmt.Errorf("topm must be greater than 0")
	}
	if c.Topk <= 0 {
		return fmt.Errorf("topk must be greater than 0")
	}
	if c.RerankThreshold < 0 || c.RerankThreshold > 1 {
		return fmt.Errorf("rerank_threshold must be between 0 and 1")
	}
	if c.FetchFactor <= 0 {
		return fmt.Errorf("fetch_factor must be greater than 0")
	}
	return nil
}

func GetDefaultConfig() *SearchConfig {
	if defaultConfig != nil {
		return defaultConfig
	}
	cfg := &SearchConfig{}
	err := settings.GetYaml("knowledge", "reranksearchcfg", &cfg)
	if err != nil {
		logs.Errorf("get es config failed: %s", err)
		defaultConfig = &SearchConfig{
			DescriptionWeight: 0.3,
			EmbeddingWeight:   0.7,
			EnabelAbstract:    true,
			EnableRerank:      true,
			Topn:              30,
			Topm:              15,
			Topk:              5,
			NeighborSize:      1,
			RerankThreshold:   0.5,
			FetchFactor:       2,
			FallBackToTopK:    true,
		}
		return defaultConfig
	}
	defaultConfig = cfg
	return cfg
}

var defaultConfig *SearchConfig

func GraphSearchConf() *SearchConfig {
	if graphSearchConfig != nil {
		return graphSearchConfig
	}
	cfg := &SearchConfig{}
	err := settings.GetYaml("knowledge", "graphsearchcfg", &cfg)
	if err != nil {
		logs.Errorf("get es config graphsearchcfg failed: %s", err)
		graphSearchConfig = &SearchConfig{
			DescriptionWeight: 0.3,
			EmbeddingWeight:   0.7,
			EnabelAbstract:    true,
			EnableRerank:      true,
			Topn:              100,
			Topm:              20,
			Topk:              50,
			NeighborSize:      1,
			RerankThreshold:   0.4,
			FetchFactor:       3,
			FallBackToTopK:    true,
		}
		return graphSearchConfig
	}
	graphSearchConfig = cfg
	return graphSearchConfig
}

var graphSearchConfig *SearchConfig

type SearchType struct {
	ChunkID     string  `json:"chunk_id"`
	Type        string  `json:"type"`
	Score       float64 `json:"score"`
	FileID      uint    `json:"file_id"`
	FileName    string  `json:"file_name"`
	ForestID    uint    `json:"forest_id"`
	Sequence    int     `json:"sequence"`
	Description string  `json:"description"`

	rawContent string

	ImageURL   string            `json:"image_url"`
	YGLocation string            `json:"yg_location"`
	Location   ragtypes.Location `json:"location"`

	RerankScore float64 `json:"rerank_score"`
}

func getSearchType(ctx context.Context, searchRes essearch.SearchResult) []*SearchType {
	res := []*SearchType{}
	for _, v := range searchRes.Hits.Hits {
		desc := v.Source.Description
		if version.DeployMode() != "" && version.DeployMode() != global.DeployModeOpenPO {
			v.Source.ImageUrl = fs.SplitHost(ctx, v.Source.ImageUrl)
		}
		if v.Source.Type == "image" {
			desc = fmt.Sprintf("![图片描述：%s] %s\n", v.Source.Description, v.Source.ImageUrl)
		}
		if v.Source.Type == "video" {
			desc = fmt.Sprintf("![视频帧描述：%s] %s\n", v.Source.Description, v.Source.ImageUrl)
		}
		if v.Source.Type == "table" {
			desc = fmt.Sprintf("表格数据：%s\n 表格描述：%s", v.Source.Table, v.Source.Description)
		}

		res = append(res, &SearchType{
			ChunkID:     v.ID,
			Type:        v.Source.Type,
			Score:       v.Score,
			FileID:      v.Source.FileID,
			ForestID:    v.Source.ForestID,
			Sequence:    v.Source.Sequence,
			Description: desc,
			rawContent:  desc,
			ImageURL:    v.Source.ImageUrl,
			YGLocation:  v.Source.YGLocation,
			Location:    Ternary(v.Source.YGLocation == "", v.Source.Location, parseYgPosString(ctx, v.Source.YGLocation)),
			RerankScore: v.Score,
			FileName:    v.Source.FileName,
		})
	}
	return res
}

func parseYgPosString(ctx context.Context, input string) [5]int {
	if input == "" {
		return [5]int{}
	}
	// 定义起始和结束标记
	prefix := "<!--yg_pos"
	suffix := "yg_pos-->"

	// 检查字符串是否以指定前缀开头
	if !strings.HasPrefix(input, prefix) {
		logs.WarnContextf(ctx, "err str prefix :%s", input)
		return [5]int{}
	}

	// 检查字符串是否以指定后缀结尾
	if !strings.HasSuffix(input, suffix) {
		logs.WarnContextf(ctx, "err str suffix :%s", input)
		return [5]int{}
	}

	// 提取标记之间的内容
	contentStartIndex := len(prefix)
	contentEndIndex := len(input) - len(suffix)
	if contentEndIndex <= contentStartIndex {
		// 标记之间没有内容或格式错误
		return [5]int{}
	}
	content := input[contentStartIndex:contentEndIndex]

	// 按逗号分割内容
	parts := strings.Split(content, ",")

	// 创建一个整数切片，容量为分割后的部分数量
	numbers := [5]int{}
	// 遍历每个分割后的字符串部分
	for i, part := range parts {
		// 去除可能存在的前后空白字符
		trimmedPart := strings.TrimSpace(part)
		// 将字符串转换为整数
		num, err := strconv.Atoi(trimmedPart)
		if err != nil {
			// 如果转换失败，返回错误
			logs.WarnContextf(ctx, "can not change '%s' int, str:'%s' err: %w", trimmedPart, input, err)
			return [5]int{}
		}
		// 将转换后的整数添加到切片中
		if i > 4 {
			return numbers
		}
		numbers[i] = num
	}

	return numbers
}

// Ternary 泛型三元表达式函数
func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
