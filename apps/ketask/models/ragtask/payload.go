package ragtask

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/pkgs/task"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

// TaskPayload 任务载荷
type TaskPayload struct {
	task.CommonPayload `json:",inline"`
	FileID             uint   `json:"file_id"`  // The ID of the file to extract text from
	FileURL            string `json:"file_url"` // The URL of the file to extract text from
	Filename           string `json:"filename"`
	PreviewFileID      uint   `json:"preview_file_id"`
	SubjectID          uint   `json:"subject_id"`
	StoragePath        string `json:"storage_path"`
	UploadPath         string `json:"upload_path"` // file or dir
	CompanyID          uint   `json:"company_id"`
	ForestID           uint   `json:"forest_id"`
	Uin                uint   `json:"uin"`
	Bucket             string `json:"bucket"`
	FileExt            string `json:"file_ext"`
	ForestType         string `json:"forest_type"`
	// TODO remove after 2025.7
	ESIndex string `json:"es_index"` // es 索引

	FileName string `json:"file_name"`

	VideoExtractext *VideoExtractextOption `json:"video_extractext,omitempty"` // 视频提取文本
	ES              *ESIndexOption         `json:"es,omitempty"`
	Graph           *GraphDBOption         `json:"graph,omitempty"`
	Embedding       *config.LLMModelConfig `json:"embedding,omitempty"`
	VLLM            *config.LLMModelConfig `json:"vllm,omitempty"`
	LLM             *config.LLMModelConfig `json:"llm,omitempty"`

	// 拆分配置
	SplitConfig *SplitConfig `json:"split_config,omitempty"`
}

// VideoExtractextOption 解析视频提取文本的选项
type VideoExtractextOption struct {
	Threshold            float64 `json:"threshold"`              // 提取文本的阈值
	FrameIntervalSeconds float64 `json:"frame_interval_seconds"` // 视频帧间隔秒数
}

// GraphDBOption 图数据库选项
type GraphDBOption struct {
	// DBMode 图数据库 Nebula
	Mode string `json:"db_mode"`
	// DBName 图数据库名称
	Name string `json:"db_name"`
	// DBAddr 图数据库地址
	Addr     string `json:"db_addr"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ESIndexOption es 索引选项
type ESIndexOption struct {
	IndexName string `json:"index_name"`
	Addr      string `json:"addr"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (p *TaskPayload) String() string {
	jsonData, err := json.Marshal(p)
	if err != nil {
		logs.ErrorContextf(context.TODO(), "PDFToString json.Marshal payload error: %v, payload :%+v", err, p)
		return ""
	}
	return string(jsonData)
}

type SplitMode string

const (
	SplitAuto SplitMode = "auto"
	SplitRule SplitMode = "rule"
)

type SplitConfig struct {
	// SplitMode 分割模式
	SplitMode          SplitMode          `json:"split_mode"`
	ChunkSize          int                `json:"chunk_size"`
	SplitMark          []string           `json:"split_mark"`
	SplitOverlap       float64            `json:"split_overlap"`
	PreprocessingRules RreprocessingRules `json:"preprocessing_rules"`
}

type RreprocessingRules struct {
	RemoveEmptyLine bool `json:"remove_empty_line"`
	RemoveURL       bool `json:"remove_url"`
	RemoveEmail     bool `json:"remove_email"`
}
