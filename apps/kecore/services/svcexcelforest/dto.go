package svcexcelforest

import (
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

type AnalyzeXlsxReq struct {
	ForestFileID uint `json:"forest_file_id"` // 知识库文件 id
}

type CleanSheetReq struct {
	ForestID     uint   // 知识库 id
	ForestFileID uint   // 知识库文件 id
	FileID       uint   // 原始文件 id
	FileUrl      string // 文件 url
	SheetName    string // sheet 名称
}

type CleanSheetRes struct {
	SubSheetList []CleanSubSheet
}

type CleanSubSheet struct {
	SubSheetID   string
	HeaderMode   foresttype.ExcelHeaderMode // 表头模式，row_title：行表头，column_title：列表头
	HeaderRowNum int
	TableName    string
	Summary      string
	Remark       string
	Docs         []*schema.Document
}

type BatchInsertToDBReq struct {
	CleanSubSheet
}

type BatchInsertToDBRes struct {
	SheetMetadata SheetMetadata `json:"sheet_metadata"`
}

type SheetMetadata struct {
	HeaderRowNum       int                              `json:"header_row_num"`       // 标题行号
	DataStartRowNum    int                              `json:"data_start_row_num"`   // 数据起始行
	DataEndRowNum      int                              `json:"data_end_row_num"`     // 数据结束行
	TotalRow           int                              `json:"total_row"`            // 总行数
	HeaderMode         foresttype.ExcelHeaderMode       `json:"header_mode"`          // 表头模式，row_title：行表头，column_title：列表头
	ColumnMetaDataList []foresttype.ExcelColumnMetadata `json:"column_metadata_list"` // 列元数据
	TableName          string                           `json:"table_name"`
	Summary            string                           `json:"summary"`
	Remark             string                           `json:"remark"`
	IsValid            bool                             `json:"is_valid"`
}
