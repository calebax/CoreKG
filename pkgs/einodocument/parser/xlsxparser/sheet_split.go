package xlsxparser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ygpkg/yg-go/logs"
)

// SplitSheetByEmptyRowsAndCols 通过空行和空列拆分工作表为多个表区域
func (s *Sheet) SplitSheetByEmptyRowsAndCols() ([]*TableRegion, error) {
	logs.DebugContextf(s.ctx, "start SplitSheetByEmptyRowsAndCols")
	// 获取工作表的使用范围
	rows, err := s.f.GetRows(s.SheetName)
	if err != nil {
		logs.ErrorContextf(s.ctx, "GetRows failed: %v", err)
		return nil, fmt.Errorf("GetRows failed: %v", err)
	}

	if len(rows) == 0 {
		return []*TableRegion{}, nil
	}

	// 获取最大列数
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	logs.DebugContextf(s.ctx, "Sheet %s: rows=%d, cols=%d", s.SheetName, len(rows), maxCols)

	// 第一步：通过空列划分
	columnGroups := findColumnGroups(rows, maxCols)
	logs.InfoContextf(s.ctx, "Sheet %s: found %d column groups", s.SheetName, len(columnGroups))

	var allRegions []*TableRegion

	// 第二步：对每个列组按空行划分
	regionIndex := 1
	for groupIndex, colGroup := range columnGroups {
		logs.DebugContextf(s.ctx, "Handling column group %d: cols %d-%d", groupIndex+1, colGroup.StartCol, colGroup.EndCol)

		rowGroups := findRowGroupsInColumnRange(rows, colGroup.StartCol, colGroup.EndCol)
		logs.DebugContextf(s.ctx, "  Found %d row groups", len(rowGroups))

		for _, rowGroup := range rowGroups {
			region := &TableRegion{
				StartRow:   rowGroup.StartRow + 1, // 转换为Excel的1基索引
				EndRow:     rowGroup.EndRow + 1,
				StartCol:   colGroup.StartCol + 1,
				EndCol:     colGroup.EndCol + 1,
				IdxInSheet: regionIndex,
				SheetName:  s.SheetName,
			}

			// 提取区域数据
			region.Data = extractRegionData(rows, rowGroup.StartRow, rowGroup.EndRow,
				colGroup.StartCol, colGroup.EndCol)

			allRegions = append(allRegions, region)
			regionIndex++
		}
	}

	return allRegions, nil
}

// ColumnGroup 列组
type ColumnGroup struct {
	StartCol int
	EndCol   int
}

// RowGroup 行组
type RowGroup struct {
	StartRow int
	EndRow   int
}

// findColumnGroups 查找列组（通过空列分隔）
func findColumnGroups(rows [][]string, maxCols int) []ColumnGroup {
	var groups []ColumnGroup

	// 标记每列是否为空
	emptyColumns := make([]bool, maxCols)
	for col := 0; col < maxCols; col++ {
		isEmpty := true
		for row := 0; row < len(rows); row++ {
			if col < len(rows[row]) && strings.TrimSpace(rows[row][col]) != "" {
				isEmpty = false
				break
			}
		}
		emptyColumns[col] = isEmpty
	}

	// 查找连续的非空列组
	inGroup := false
	var currentGroup ColumnGroup

	for col := 0; col < maxCols; col++ {
		if !emptyColumns[col] { // 非空列
			if !inGroup {
				// 开始新组
				currentGroup.StartCol = col
				inGroup = true
			}
			currentGroup.EndCol = col
		} else { // 空列
			if inGroup {
				// 结束当前组
				groups = append(groups, currentGroup)
				inGroup = false
			}
		}
	}

	// 处理最后一组
	if inGroup {
		groups = append(groups, currentGroup)
	}

	return groups
}

// findRowGroupsInColumnRange 在指定列范围内查找行组（通过空行分隔）
func findRowGroupsInColumnRange(rows [][]string, startCol, endCol int) []RowGroup {
	var groups []RowGroup

	// 标记每行在指定列范围内是否为空
	emptyRows := make([]bool, len(rows))
	for row := 0; row < len(rows); row++ {
		isEmpty := true
		for col := startCol; col <= endCol; col++ {
			if col < len(rows[row]) && strings.TrimSpace(rows[row][col]) != "" {
				isEmpty = false
				break
			}
		}
		emptyRows[row] = isEmpty
	}

	// 查找连续的非空行组
	inGroup := false
	var currentGroup RowGroup

	for row := 0; row < len(rows); row++ {
		if !emptyRows[row] { // 非空行
			if !inGroup {
				// 开始新组
				currentGroup.StartRow = row
				inGroup = true
			}
			currentGroup.EndRow = row
		} else { // 空行
			if inGroup {
				// 结束当前组
				groups = append(groups, currentGroup)
				inGroup = false
			}
		}
	}

	// 处理最后一组
	if inGroup {
		groups = append(groups, currentGroup)
	}

	return groups
}

// extractRegionData 提取区域数据
func extractRegionData(rows [][]string, startRow, endRow, startCol, endCol int) [][]string {
	var regionData [][]string

	for row := startRow; row <= endRow; row++ {
		var rowData []string
		for col := startCol; col <= endCol; col++ {
			var cellValue string
			if row < len(rows) && col < len(rows[row]) {
				cellValue = rows[row][col]
			}
			rowData = append(rowData, cellValue)
		}
		regionData = append(regionData, rowData)
	}

	return regionData
}

// PrintRegionsSummary 打印区域摘要信息
func PrintRegionsSummary(regions []*TableRegion) {
	fmt.Printf("\n=== 区域拆分结果 ===\n")
	fmt.Printf("总共找到 %d 个表区域:\n\n", len(regions))

	for i, region := range regions {
		fmt.Printf("%d. %s\n", i+1, region.String())

		// 显示前几行数据预览
		fmt.Printf("   数据预览:\n")
		previewRows := 10
		if len(region.Data) < previewRows {
			previewRows = len(region.Data)
		}

		for j := 0; j < previewRows; j++ {
			fmt.Printf("   Row %d: ", j+1)
			previewCols := 15
			if len(region.Data[j]) < previewCols {
				previewCols = len(region.Data[j])
			}
			data, _ := json.Marshal(region.Data[j])
			fmt.Printf("   %s", data)
			// for k := 0; k < previewCols; k++ {
			// 	value := region.Data[j][k]
			// 	if len(value) > 10 {
			// 		value = value[:10] + "..."
			// 	}
			// 	fmt.Printf("[%s] ", value)
			// }
			if len(region.Data[j]) > previewCols {
				fmt.Printf("... (%d more cols)", len(region.Data[j])-previewCols)
			}
			fmt.Println()
		}

		if len(region.Data) > previewRows {
			fmt.Printf("   ... (%d more rows)\n", len(region.Data)-previewRows)
		}
		fmt.Println()
	}
}
