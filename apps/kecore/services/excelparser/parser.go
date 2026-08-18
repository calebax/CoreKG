package excelparser

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	"github.com/xuri/excelize/v2"
	"github.com/ygpkg/yg-go/logs"
)

var _ parser.Parser = (*ExcelParser)(nil)

// ExcelParser excel 解析器
type ExcelParser struct {
	f *excelize.File
}

func (p *ExcelParser) openExcelFile(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open excel: %v", err)
	}
	p.f = f
	return nil
}

// Parse 解析 excel 文件
func (p *ExcelParser) Parse(ctx context.Context, reader io.Reader, opts ...parser.Option) ([]*schema.Document, error) {
	xlsxFile, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel: %v", err)
	}

	p.f = xlsxFile
	for _, sheetName := range p.f.GetSheetList() {
		sheet := &Sheet{
			f:         p.f,
			ctx:       logs.WithContextFields(ctx, "sheet_name", sheetName),
			SheetName: sheetName,
		}
		sheet.parseSheet()
	}

	return nil, nil
}

func (p *ExcelParser) decodeSheetDSLs(ctx context.Context, sheetName string) ([]*schema.Document, error) {
	rows, err := p.f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %v", err)
	}

	for i := 0; rows.Next(); i++ {
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get columns: %v", err)
		}
		logs.DebugContextf(ctx, "cols %v: %v", i, cols)
	}
	return nil, nil
}

// Close 关闭 excel 文件
func (p *ExcelParser) Close() error {
	return p.f.Close()
}

// Cell excel 单元格
type Cell struct {
	excelize.Cell
}
