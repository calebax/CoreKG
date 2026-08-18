package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type FileInfo struct {
	FileName   string      `json:"fileName"`
	SheetCount int         `json:"sheetCount"`
	Sheets     []SheetInfo `json:"sheets"`
}

// SheetInfo 表示单个工作表的信息
type SheetInfo struct {
	SheetIndex int         `json:"sheetIndex"`
	SheetName  string      `json:"sheetName"`
	RowCount   int         `json:"rowCount"`
	ColCount   int         `json:"colCount"`
	DataBlocks []DataBlock `json:"dataBlocks"`
}

// SubTable 表示一个子表的信息
type DataBlock struct {
	Title           string     `json:"Title" jsonschema:"description=子表标题（首行存在合并单元格情况，视为title）"`         // 子表标题（首行存在合并单元格情况，视为title）
	CellRange       string     `json:"dataRange" jsonschema:"description=数据区域范围，如A1:C10,example=A1:C10"` // 数据区域范围，如"A1:C10"
	HeaderLabels    []string   `json:"headerLabels" jsonschema:"description=顶部标签，对应每列含义"`                // 列头
	SubHeaderLabels []string   `json:"subHeaderLabels" jsonschema:"description=二级顶部标签，如果存在"`             // 子列头（如果存在）
	RowHeaders      []string   `json:"rowHeaders" jsonschema:"description=左侧行头"`                         // 左侧行头
	SampleValues    [][]string `json:"sampleValues" jsonschema:"description=样本值（预览前5行数据）"`               // 样本值（预览前5行数据）
}

// Deprecated: use OpenFromPathOrURL instead
// openExcelFromPathOrURL 打开本地路径或远程 URL 的 Excel 文件，并返回文件句柄和文件名
func openExcelFromPathOrURL(pathOrURL string) (*excelize.File, string, error) {
	if pathOrURL == "" {
		return nil, "", fmt.Errorf("文件路径或URL不能为空")
	}

	// 尝试解析为 URL
	u, err := neturl.Parse(pathOrURL)
	if err == nil && u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			// 远程文件，通过 HTTP 获取并用 OpenReader 解析
			resp, err := http.Get(pathOrURL)
			if err != nil {
				return nil, "", fmt.Errorf("下载远程文件失败: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, "", fmt.Errorf("远程文件请求失败: %s", resp.Status)
			}

			// 解析文件名：优先 Content-Disposition，其次 URL 路径
			fileName := ""
			if cd := resp.Header.Get("Content-Disposition"); cd != "" {
				if _, params, err := mime.ParseMediaType(cd); err == nil {
					if fn, ok := params["filename"]; ok && fn != "" {
						fileName = fn
					}
				}
			}
			if fileName == "" {
				fileName = path.Base(u.Path)
			}

			// 读取到内存并打开
			// excelize.OpenReader 会读取传入的 Reader 数据
			f, err := excelize.OpenReader(resp.Body)
			if err != nil {
				return nil, "", fmt.Errorf("打开远程Excel失败: %v", err)
			}
			return f, fileName, nil

		case "file":
			// file:// URL，本地文件
			localPath := u.Path
			if localPath == "" {
				return nil, "", fmt.Errorf("无效的file URL: %s", pathOrURL)
			}
			// 转换为当前系统路径分隔符
			localPath = filepath.FromSlash(localPath)
			f, err := excelize.OpenFile(localPath)
			if err != nil {
				return nil, "", fmt.Errorf("打开本地Excel失败: %v", err)
			}
			return f, filepath.Base(localPath), nil

		default:
			// 其他 scheme 暂不支持，按本地路径处理失败
			return nil, "", fmt.Errorf("不支持的URL协议: %s", u.Scheme)
		}
	}

	// 非 URL，当作本地路径
	f, err := excelize.OpenFile(pathOrURL)
	if err != nil {
		return nil, "", fmt.Errorf("打开本地Excel失败: %v", err)
	}
	return f, filepath.Base(pathOrURL), nil
}

func ViewSheetRowDataFromReader(r io.Reader, sheetIndex int, rowStartIdx int, rowEndIdx int) ([][]string, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("解析Excel失败: %v", err)
	}
	defer f.Close()

	return ViewSheetRowData(f, sheetIndex, rowStartIdx, rowEndIdx)
}

// ViewSheetRowData
// rowEndIdx=-1 表示读取到最后一行
func ViewSheetRowData(f *excelize.File, sheetIndex int, rowStartIdx int, rowEndIdx int) ([][]string, error) {
	// 校验 sheetIndex 并获取 sheet 名称
	sheetNames := f.GetSheetList()
	if sheetIndex < 0 || sheetIndex >= len(sheetNames) {
		return nil, fmt.Errorf("sheet索引超出范围: %d (总数 %d)", sheetIndex, len(sheetNames))
	}
	sheetName := sheetNames[sheetIndex]

	if rowStartIdx < 0 {
		rowStartIdx = 0
	}

	// 使用流式迭代器，只读取需要的行
	rowsIter, err := f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("获取行迭代器失败: %v", err)
	}
	defer rowsIter.Close()

	// 预分配结果容量
	var result [][]string
	if rowEndIdx >= 0 && rowEndIdx >= rowStartIdx {
		result = make([][]string, 0, rowEndIdx-rowStartIdx+1)
	} else {
		// rowEndIdx=-1 时预分配一个合理的初始容量
		result = make([][]string, 0, 1000)
	}

	currentRow := 0
	// rowEndIdx=-1 表示读到最后
	readToEnd := rowEndIdx < 0

	for rowsIter.Next() {
		// 跳过起始行之前的行
		if currentRow < rowStartIdx {
			currentRow++
			continue
		}

		// 如果不是读到最后，检查是否超出结束行
		if !readToEnd && currentRow > rowEndIdx {
			break
		}

		// 读取列数据
		cols, err := rowsIter.Columns()
		if err != nil {
			return nil, fmt.Errorf("读取行 %d 失败: %v", currentRow, err)
		}

		result = append(result, cols)
		currentRow++
	}

	if err := rowsIter.Error(); err != nil {
		return nil, fmt.Errorf("迭代行数据失败: %v", err)
	}

	return result, nil
}

func GetXlsFileInfo(filePathOrURL string) (*FileInfo, error) {
	// 打开Excel文件（支持本地路径或远程URL）
	f, fileName, err := openExcelFromPathOrURL(filePathOrURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// 获取所有工作表名称
	sheetNames := f.GetSheetList()
	sheetCount := len(sheetNames)
	// 创建工作表信息数组
	sheets := make([]SheetInfo, 0, sheetCount)

	for i, sheetName := range sheetNames {
		// 读取获取行数据
		rows, colCount, err := getRowsStreaming(f, sheetName)
		if err != nil {
			// 如果获取失败，添加一个空的工作表信息
			sheets = append(sheets, SheetInfo{
				SheetIndex: i,
				SheetName:  sheetName,
				RowCount:   0,
				ColCount:   0,
			})
			continue
		}

		rowCount := len(rows)

		// 传递预读取的 rows 数据，避免重复读取
		subTables, err := getSheetSubTables(f, sheetName, rows, colCount)
		if err != nil {
			return nil, fmt.Errorf("获取子表失败: %v", err)
		}

		sheets = append(sheets, SheetInfo{
			SheetIndex: i,
			SheetName:  sheetName,
			RowCount:   rowCount,
			ColCount:   colCount,
			DataBlocks: *subTables,
		})
	}

	return &FileInfo{
		FileName:   fileName,
		SheetCount: sheetCount,
		Sheets:     sheets,
	}, nil
}

// getRowsStreaming 获取行数据并同时计算 colCount
func getRowsStreaming(f *excelize.File, sheetName string) ([][]string, int, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, 0, err
	}

	colCount := 0
	for _, row := range rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	return rows, colCount, nil
}

func FlattenRowToCSVLine(row []string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(row); err != nil {
		return "", err
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return "", err
	}

	// csv.Writer 会自动加 '\n'
	line := strings.TrimSuffix(buf.String(), "\n")

	return line, nil
}

// getSheetSubTables 使用预读取的 rows 数据检测子表，避免重复 I/O
func getSheetSubTables(f *excelize.File, sheetName string, rows [][]string, colCount int) (*[]DataBlock, error) {
	// 使用缓存的 rows 检测子表区域（传递 colCount 避免重复计算）
	sheetRects := detectSubTablesOptimized(rows, colCount)

	// 获取合并单元格信息
	mergedCells, err := f.GetMergeCells(sheetName)
	if err != nil {
		mergedCells = []excelize.MergeCell{}
	}

	subTables := make([]DataBlock, 0, len(sheetRects))

	for _, rect := range sheetRects {
		startRowIdx, endRowIdx := rect.RowStartIdx, rect.RowEndIdx
		startColIdx, endColIdx := rect.ColStartIdx, rect.ColEndIdx

		firstNonMergedRowIdx := findFirstNonMergedRow(mergedCells, startRowIdx, endRowIdx, startColIdx, endColIdx)
		title, colHeaders, subColHeaders := parseHeaderLayersFromCache(rows, startRowIdx, firstNonMergedRowIdx, startColIdx, endColIdx)

		// 从缓存获取 rowHeaders
		rowHeaders := make([]string, 0, endRowIdx-firstNonMergedRowIdx+1)
		for r := firstNonMergedRowIdx; r <= endRowIdx; r++ {
			val := getCellValueFromCache(rows, r, startColIdx)
			rowHeaders = append(rowHeaders, val)
		}

		// 长度超过 200，判断为常规纵向数据表格
		if len(rowHeaders) > 200 {
			rowHeaders = nil
		}

		// 数据预览（前5行）- 从缓存获取
		previewStart := firstNonMergedRowIdx + 1
		previewEnd := min(previewStart+5, endRowIdx)
		previewData := [][]string{}
		for r := previewStart; r <= previewEnd; r++ {
			rowData := getRowValuesFromCache(rows, r, startColIdx, endColIdx)
			if len(rowData) > 0 {
				previewData = append(previewData, rowData)
			}
		}

		subTables = append(subTables, DataBlock{
			Title:           title,
			HeaderLabels:    colHeaders,
			SubHeaderLabels: subColHeaders,
			RowHeaders:      rowHeaders,
			SampleValues:    previewData,
			CellRange:       rect.DataRange,
		})
	}

	return &subTables, nil
}

type Rect struct {
	DataRange   string `json:"dataRange"`
	RowStartIdx int    `json:"rowStartIdx"`
	RowEndIdx   int    `json:"rowEndIdx"`
	ColStartIdx int    `json:"colStartIdx"`
	ColEndIdx   int    `json:"colEndIdx"`
}

// detectSubTablesOptimized 优化的子表检测算法
// 使用单个 uint8 数组替代两个 bool 数组，减少内存分配
func detectSubTablesOptimized(rows [][]string, colCount int) []Rect {
	rowCount := len(rows)
	if rowCount == 0 || colCount == 0 {
		return nil
	}

	// 使用单个一维数组，每个元素存储: bit0=非空, bit1=已访问
	// 这比两个二维 bool 数组更高效
	const (
		flagNonEmpty = 1 << 0 // bit 0: 单元格非空
		flagVisited  = 1 << 1 // bit 1: 已访问
	)

	// 一维数组，索引 = row*colCount + col
	cells := make([]uint8, rowCount*colCount)

	// 标记非空单元格
	for r, row := range rows {
		for c, val := range row {
			if val != "" {
				cells[r*colCount+c] = flagNonEmpty
			}
		}
	}

	var regions []Rect
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	// 预分配队列容量，避免频繁扩容
	queue := make([][2]int, 0, 256)

	for r := 0; r < rowCount; r++ {
		for c := 0; c < colCount; c++ {
			idx := r*colCount + c
			// 跳过空单元格或已访问的
			if cells[idx]&flagNonEmpty == 0 || cells[idx]&flagVisited != 0 {
				continue
			}

			// Flood fill - 复用 queue
			queue = queue[:0]
			queue = append(queue, [2]int{r, c})
			cells[idx] |= flagVisited

			rowMin, rowMax := r, r
			colMin, colMax := c, c

			for len(queue) > 0 {
				// 从队尾取出（栈方式，更高效）
				last := len(queue) - 1
				cell := queue[last]
				queue = queue[:last]
				cr, cc := cell[0], cell[1]

				for _, d := range dirs {
					nr, nc := cr+d[0], cc+d[1]
					if nr >= 0 && nr < rowCount && nc >= 0 && nc < colCount {
						nidx := nr*colCount + nc
						if cells[nidx]&flagNonEmpty != 0 && cells[nidx]&flagVisited == 0 {
							cells[nidx] |= flagVisited
							queue = append(queue, [2]int{nr, nc})

							if nr < rowMin {
								rowMin = nr
							} else if nr > rowMax {
								rowMax = nr
							}
							if nc < colMin {
								colMin = nc
							} else if nc > colMax {
								colMax = nc
							}
						}
					}
				}
			}

			// 构造 Excel 坐标范围
			startCell, _ := excelize.CoordinatesToCellName(colMin+1, rowMin+1)
			endCell, _ := excelize.CoordinatesToCellName(colMax+1, rowMax+1)

			regions = append(regions, Rect{
				RowStartIdx: rowMin,
				RowEndIdx:   rowMax,
				ColStartIdx: colMin,
				ColEndIdx:   colMax,
				DataRange:   startCell + ":" + endCell,
			})
		}
	}

	return regions
}

// FindFirstNonMergedRow 从 rowStart 开始逐行检测，返回第一个不包含合并单元格的行号。
// 如果所有行都在合并区域中，则返回 rowEnd+1。
func findFirstNonMergedRow(merges []excelize.MergeCell, rowStartIdx, rowEndIdx, colStartIdx, colEndIdx int) int {

	if len(merges) == 0 {
		return rowStartIdx
	}

	type mergeArea struct {
		rowStart, rowEnd int
		colStart, colEnd int
	}

	areas := make([]mergeArea, 0, len(merges))
	for _, m := range merges {
		colS, rowS, _ := excelize.CellNameToCoordinates(m.GetStartAxis())
		colE, rowE, _ := excelize.CellNameToCoordinates(m.GetEndAxis())
		areas = append(areas, mergeArea{
			rowStart: rowS - 1, rowEnd: rowE - 1,
			colStart: colS - 1, colEnd: colE - 1,
		})
	}

	for r := rowStartIdx; r <= rowEndIdx; r++ {
		isMerged := false
		for _, a := range areas {
			if r >= a.rowStart && r <= a.rowEnd &&
				a.colEnd >= colStartIdx && a.colStart <= colEndIdx {
				isMerged = true
				break
			}
		}
		if !isMerged {
			return r
		}
	}
	return rowEndIdx + 1
}

// parseHeaderLayersFromCache 从缓存的 rows 数据解析表头
func parseHeaderLayersFromCache(allRows [][]string, startRow, firstNonMergedRow, startCol, endCol int) (title string, colHeaders, subColHeaders []string) {
	layerCount := max(0, firstNonMergedRow-startRow)

	// 从缓存取出可能用到的几行
	rows := [][]string{
		getRowValuesFromCache(allRows, startRow, startCol, endCol),
		getRowValuesFromCache(allRows, startRow+1, startCol, endCol),
		getRowValuesFromCache(allRows, startRow+2, startCol, endCol),
	}

	// 一行都没取到，直接返回空
	if len(rows[0]) == 0 {
		return
	}

	switch {
	case layerCount <= 0:
		colHeaders = rows[0]
	case layerCount == 1:
		title = joinTitle(rows[0])
		colHeaders = rows[1]
	case layerCount >= 2:
		title = joinTitle(rows[0])
		colHeaders = rows[1]
		subColHeaders = rows[2]
	}

	return
}

// 常用于合并单元格标题、表头拼接等。
func joinTitle(cells []string) string {
	if len(cells) == 0 {
		return ""
	}

	parts := make([]string, 0, len(cells))
	seen := make(map[string]struct{}, len(cells))

	for _, raw := range cells {
		if val := strings.TrimSpace(raw); val != "" {
			if _, exists := seen[val]; !exists {
				seen[val] = struct{}{}
				parts = append(parts, val)
			}
		}
	}

	return strings.Join(parts, " ")
}

// getRowValuesFromCache 从预读取的 rows 缓存中获取指定行的列值
func getRowValuesFromCache(rows [][]string, rowIdx, startCol, endCol int) []string {
	if rowIdx < 0 || rowIdx >= len(rows) || startCol > endCol {
		return nil
	}

	row := rows[rowIdx]
	rowVals := make([]string, 0, endCol-startCol+1)
	for c := startCol; c <= endCol; c++ {
		var val string
		if c < len(row) {
			val = strings.TrimSpace(row[c])
		}
		rowVals = append(rowVals, val)
	}
	return rowVals
}

// getCellValueFromCache 从预读取的 rows 缓存中获取单个单元格值
func getCellValueFromCache(rows [][]string, rowIdx, colIdx int) string {
	if rowIdx < 0 || rowIdx >= len(rows) {
		return ""
	}
	row := rows[rowIdx]
	if colIdx < 0 || colIdx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[colIdx])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
