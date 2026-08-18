package xlsxparser

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/xuri/excelize/v2"
)

type TableColumnType int

const (
	TableColumnTypeUnknown TableColumnType = 0
	TableColumnTypeString  TableColumnType = 1
	TableColumnTypeInteger TableColumnType = 2
	TableColumnTypeTime    TableColumnType = 3
	TableColumnTypeNumber  TableColumnType = 4
	TableColumnTypeBoolean TableColumnType = 5
	TableColumnTypeImage   TableColumnType = 6
)

const (
	SubSheetID = "_sub_sheet_id"
)

func (t TableColumnType) String() string {
	switch t {
	case TableColumnTypeString:
		return "varchar"
	case TableColumnTypeInteger:
		return "bigint"
	case TableColumnTypeTime:
		return "timestamp"
	case TableColumnTypeNumber:
		return "double"
	case TableColumnTypeBoolean:
		return "boolean"
	case TableColumnTypeImage:
		return "image"
	default:
		return "unknown"
	}
}

type TableStatus string

// TableRegion 表区域结构
type TableRegion struct {
	// SheetName 原工作表名称
	SheetName string `json:"sheet_name"`
	// IdxInSheet 原工作表中区域的索引
	IdxInSheet int `json:"idx_in_sheet"`

	// StartRow 起始行（从1开始）
	StartRow int `json:"start_row"`
	// EndRow 结束行
	EndRow int `json:"end_row"`
	// StartCol 起始列（从1开始）
	StartCol int `json:"start_col"`
	// EndCol 结束列
	EndCol int `json:"end_col"`

	// Data 区域数据
	Data [][]string `json:"-"`
}

// HeaderLocation 表头位置 相对表格区域的位置，从1开始 0 表示不存在
type HeaderLocation struct {
	// HeaderRow 表头行
	HeaderRow int `json:"header_row"`
	// FirstDataRow 数据开始行
	FirstDataRow int `json:"first_data_row"`
}

// TableSchema 表结构 一个包含表头和数据的表
type TableSchema struct {
	TableRegion    `json:",inline"`
	HeaderLocation `json:",inline"`

	Comment string             `json:"comment"`
	Docs    []*schema.Document `json:"docs"`
	Columns []*Column          `json:"-"`

	TableStatus TableStatus `json:"table_status"`
}

// Column 表列结构
type Column struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Type        TableColumnType `json:"type"`
	Description string          `json:"description"`
	Nullable    bool            `json:"nullable"`
	IsPrimary   bool            `json:"is_primary"`

	Data []*ColumnData
}

// ColumnData 表列数据
type ColumnData struct {
	ColumnID   int64
	ColumnName string
	Type       TableColumnType
	ValString  *string
	ValInteger *int64
	ValTime    *time.Time
	ValNumber  *float64
	ValBoolean *bool
	ValImage   *string // base64 / url
}

// String 返回区域的字符串表示
func (tr *TableRegion) String() string {
	startCell, _ := excelize.CoordinatesToCellName(tr.StartCol, tr.StartRow)
	endCell, _ := excelize.CoordinatesToCellName(tr.EndCol, tr.EndRow)
	return fmt.Sprintf("%v (%s:%s) [%dx%d]",
		tr.IdxInSheet, startCell, endCell,
		tr.EndRow-tr.StartRow+1, tr.EndCol-tr.StartCol+1)
}

// GetCellRange 返回区域的单元格范围字符串
func (tr *TableRegion) GetCellRange() string {
	startCell, _ := excelize.CoordinatesToCellName(tr.StartCol, tr.StartRow)
	endCell, _ := excelize.CoordinatesToCellName(tr.EndCol, tr.EndRow)
	return fmt.Sprintf("%s:%s", startCell, endCell)
}

// IdentifyHeader 识别表头
func (ts *TableSchema) IdentifyHeader() HeaderLocation {
	hl := HeaderLocation{
		HeaderRow:    1,
		FirstDataRow: 2,
	}

	if len(ts.Data) == 0 {
		return hl
	}

	return hl
}

// Format 返回表格的字符串表示
func (ts *TableSchema) Format() error {
	var docs []*schema.Document
	for i := 0; i < len(ts.Data); i++ {
		row := ts.Data[i]
		if len(row) == 0 {
			continue
		}
		contentParts := make([]string, len(row))
		for j, cell := range row {
			contentParts[j] = strings.TrimSpace(cell)
		}
		content := strings.Join(contentParts, "\t")
		meta := make(map[string]any)
		meta[SubSheetID] = fmt.Sprintf("%s_%d", ts.SheetName, ts.IdxInSheet)
		doc := &schema.Document{
			ID:       fmt.Sprintf("%d", i),
			Content:  content,
			MetaData: meta,
		}
		docs = append(docs, doc)
	}
	ts.Docs = docs
	return nil
}
