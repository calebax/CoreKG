package excelparser

import (
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"
	"github.com/ygpkg/yg-go/logs"
)

// Sheet excel 解析器
type Sheet struct {
	f         *excelize.File
	ctx       context.Context
	SheetName string

	// rows *excelize.Rows

	// TableSchemas 单个sheet 中的多个表结构
	TableSchemas []*TableSchema
}

func (s *Sheet) parseSheet() error {
	if err := s.UnmergeCellsAndFill(); err != nil {
		return err
	}

	regions, err := s.SplitSheetByEmptyRowsAndCols()
	if err != nil {
		return err
	}

	for _, region := range regions {
		table := &TableSchema{
			TableRegion: *region,
		}
		table.Format()
	}

	return nil
}

// UnmergeCellsAndFill 取消所有合并单元格并用合并前的内容进行填充
func (s *Sheet) UnmergeCellsAndFill() error {
	logs.DebugContextf(s.ctx, "start UnmergeCellsAndFill")

	// 获取该工作表的所有合并单元格
	mergedCells, err := s.f.GetMergeCells(s.SheetName)
	if err != nil {
		logs.ErrorContextf(s.ctx, "GetMergeCells failed: %v", err)
		return fmt.Errorf("GetMergeCells failed: %v", err)
	}

	// 处理每个合并单元格
	for _, mergeCell := range mergedCells {
		// 获取合并单元格的值（通常在左上角单元格）
		value, err := s.f.GetCellValue(s.SheetName, mergeCell.GetStartAxis())
		if err != nil {
			logs.ErrorContextf(s.ctx, "GetCellValue failed: %v", err)
			continue
		}
		logs.DebugContextf(s.ctx, "mergeCell: %s, value: %s", mergeCell.GetCellValue(), value)

		// 取消合并
		err = s.f.UnmergeCell(s.SheetName, mergeCell.GetStartAxis(), mergeCell.GetEndAxis())
		if err != nil {
			logs.ErrorContextf(s.ctx, "UnmergeCell failed: %v", err)
			continue
		}

		// 填充所有原合并区域的单元格
		err = fillMergedArea(s.f, s.SheetName, mergeCell.GetStartAxis(), mergeCell.GetEndAxis(), value)
		if err != nil {
			logs.ErrorContextf(s.ctx, "fillMergedArea failed: %v", err)
			continue
		}
	}

	return nil
}

// fillMergedArea 填充合并区域的所有单元格
func fillMergedArea(f *excelize.File, sheetName, startAxis, endAxis, value string) error {
	// 解析起始和结束坐标
	startCol, startRow, err := excelize.CellNameToCoordinates(startAxis)
	if err != nil {
		return fmt.Errorf("Parse start axis failed: %v", err)
	}

	endCol, endRow, err := excelize.CellNameToCoordinates(endAxis)
	if err != nil {
		return fmt.Errorf("Parse end axis failed: %v", err)
	}

	// 填充区域内的所有单元格
	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			cellName, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				return fmt.Errorf("Parse cell name failed: %v", err)
			}

			err = f.SetCellValue(sheetName, cellName, value)
			if err != nil {
				return fmt.Errorf("SetCellValue failed: %v", err)
			}
		}
	}

	return nil
}

// fillMergedAreaWithStyle 填充区域并保持格式
func fillMergedAreaWithStyle(f *excelize.File, sheetName, startAxis, endAxis, value string, styleID int) error {
	startCol, startRow, err := excelize.CellNameToCoordinates(startAxis)
	if err != nil {
		return err
	}

	endCol, endRow, err := excelize.CellNameToCoordinates(endAxis)
	if err != nil {
		return err
	}

	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			cellName, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				continue
			}

			// 设置值
			f.SetCellValue(sheetName, cellName, value)

			// 设置样式
			if styleID > 0 {
				f.SetCellStyle(sheetName, cellName, cellName, styleID)
			}
		}
	}

	return nil
}
