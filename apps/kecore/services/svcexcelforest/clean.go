package svcexcelforest

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/einodocument/parser/xlsxparser"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"gorm.io/gorm"
)

const (
	KeyForestID      = "forest_id"
	KeyForestFileID  = "forest_file_id"
	KeyFileID        = "file_id"
	KeySheetName     = "sheet_name"
	KeyTempTableName = "temp_table_name"
	KeyIsTransposed  = "_transposed_flag" // 是否转置过
)

type dataClean struct {
}

func NewDataClean() *dataClean {
	return &dataClean{}
}

func (dc *dataClean) CleanSheet(ctx *gin.Context, req *CleanSheetReq) (*CleanSheetRes, error) {
	metaData := map[string]any{
		KeyForestID:     req.ForestID,
		KeyForestFileID: req.ForestFileID,
		KeyFileID:       req.FileID,
		KeySheetName:    req.SheetName,
	}
	sheetDocs, err := getSheetDocs(ctx, req.FileUrl, req.SheetName, metaData)
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet docs: %v", err)
	}
	docsGroup := dc.groupSheetDocs(ctx, sheetDocs)
	var subSheetList []CleanSubSheet

	for subSheetID, docs := range docsGroup {
		if len(docs) == 0 {
			continue
		}
		totalRow := len(docs)
		startRowNum, endRowNum := dc.calcHeaderSampledRange(0, totalRow, 5, 5)
		sampledDocs := docs[startRowNum-1 : endRowNum]
		sampledRows := make([]string, 0, len(sampledDocs))
		for _, v := range sampledDocs {
			sampledRows = append(sampledRows, v.Content)
		}
		headerNumbers, err := RequestGetHeaderNumbers(ctx, sampledRows)
		if err != nil {
			logs.ErrorContextf(ctx, "[dc.CleanSheet] get header numbers failed, sheetName: %s, subSheetID: %s, err: %v", req.SheetName, subSheetID, err)
			continue
		}
		if len(headerNumbers) == 0 {
			logs.WarnContextf(ctx, "[dc.CleanSheet] get empty header numbers, sheetName: %s, subSheetID: %s, err: %v", req.SheetName, subSheetID, err)
			continue
		}
		headerRowNum := 0
		for i, v := range headerNumbers {
			if v < 0 {
				continue
			}
			if v == 0 {
				headerNumbers[i] = 1
			}
			headerRowNum = max(headerRowNum, headerNumbers[i])
		}
		// 目前只支持3行表头，如果超过3行，则取3行
		if headerRowNum > 3 {
			headerRowNum = 3
		}
		var headerSummaryList []string
		headerSummaryDocs := docs[:headerRowNum-1]
		for _, v := range headerSummaryDocs {
			headerSummaryList = append(headerSummaryList, v.Content)
		}

		headerSummary, headerRowNum := dc.getHeaderSummary(headerSummaryList)

		dataDocs := docs[headerRowNum:]
		lastRow := dataDocs[len(dataDocs)-1]
		remark := dc.getRemark(lastRow.Content)
		if remark != "" {
			dataDocs = dataDocs[:len(dataDocs)-1]
		}

		headerInFirstColumn, err := dc.isHeaderInFirstColumn(ctx, sampledDocs)
		if err != nil {
			return nil, fmt.Errorf("failed to detect header is in first column: %v", err)
		}

		headerMode := foresttype.ExcelHeaderModeRowTitle
		if headerInFirstColumn {
			headerMode = foresttype.ExcelHeaderModeColumnTitle
			sheetDocs = dc.transposeDocuments(sheetDocs)
		}

		subSheet := CleanSubSheet{
			TableName:    dc.buildTableName(),
			SubSheetID:   subSheetID,
			HeaderMode:   headerMode,
			HeaderRowNum: headerRowNum,
			Docs:         dataDocs,
			Summary:      headerSummary,
			Remark:       remark,
		}
		subSheetList = append(subSheetList, subSheet)

		fmt.Println(headerSummary)
	}

	return &CleanSheetRes{
		SubSheetList: subSheetList,
	}, nil
}

func (dc *dataClean) BatchInsertToDB(ctx *gin.Context, db *gorm.DB, req *BatchInsertToDBReq) (*BatchInsertToDBRes, error) {
	totalRow := len(req.Docs)

	if totalRow <= req.HeaderRowNum {
		return &BatchInsertToDBRes{
			SheetMetadata: SheetMetadata{
				IsValid: false,
			},
		}, nil
	}

	// 获取excel表头和数据表列名数据
	startRowNum, endRowNum := dc.calcHeaderSampledRange(req.HeaderRowNum, totalRow, 5, 5)
	headerSampledDocs := req.Docs[startRowNum-1 : endRowNum]
	tableColumnNames, headerRowNum, excelColumnNames, err := dc.getHeaders(ctx, headerSampledDocs)
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %v", err)
	}

	// 获取数据表列名类型
	sampledDocs := dc.getSampledRows(headerRowNum, req.Docs)
	excelColumnNameTypeMap, err := DetectColumnTypes(excelColumnNames, sampledDocs)
	if err != nil {
		return nil, fmt.Errorf("failed to detect column types: %v", err)
	}

	excelColumnMetaDataList := dc.buildExcelColumnMetaData(excelColumnNames, tableColumnNames, excelColumnNameTypeMap)

	if len(excelColumnNameTypeMap) == 0 {
		// 说明没有数据
		return &BatchInsertToDBRes{
			SheetMetadata: SheetMetadata{
				IsValid: false,
			},
		}, nil
	}

	createTableDDL := dc.buildTableDDL(req.TableName, excelColumnMetaDataList, req.Summary)

	if err := db.WithContext(ctx).Exec(createTableDDL).Error; err != nil {
		return nil, fmt.Errorf("failed to create table: %v, ddl: %s", err, createTableDDL)
	}

	dataDocs := req.Docs[headerRowNum:]
	var tableRecordList []map[string]any
	for _, v := range dataDocs {
		record, err := dc.buildTableRecord(v, excelColumnMetaDataList)
		if err != nil {
			return nil, fmt.Errorf("failed to build table record: %v", err)
		}
		tableRecordList = append(tableRecordList, record)
	}

	if err := dc.batchInsertDocs(ctx, db, req.TableName, tableRecordList); err != nil {
		return nil, fmt.Errorf("failed to insert documents: %v", err)
	}

	return &BatchInsertToDBRes{
		SheetMetadata: SheetMetadata{
			TableName:          req.TableName,
			HeaderRowNum:       headerRowNum,
			TotalRow:           totalRow,
			ColumnMetaDataList: excelColumnMetaDataList,
			Summary:            req.Summary,
			Remark:             req.Remark,
			IsValid:            true,
		},
	}, nil
}

// calcHeaderSampledRange 计算围绕表头的采样起止行范围
func (dc *dataClean) calcHeaderSampledRange(headerRow, totalRows, targetBefore, targetAfter int) (int, int) {
	before := min(targetBefore, headerRow-1)
	after := min(targetAfter, totalRows-headerRow)

	// 前面不足，补给后面
	if before < targetBefore {
		extra := targetBefore - before
		after = min(after+extra, totalRows-headerRow)
	}

	// 后面不足，补给前面
	if after < targetAfter {
		extra := targetAfter - after
		before = min(before+extra, headerRow-1)
	}

	startRow := headerRow - before
	endRow := headerRow + after
	return startRow, endRow
}

func (dc *dataClean) isHeaderInFirstColumn(ctx *gin.Context, sampledDocs []*schema.Document) (bool, error) {
	sampledRows := make([][]string, 0, len(sampledDocs))
	for _, v := range sampledDocs {
		sampledRows = append(sampledRows, strings.Split(v.Content, "\t"))
	}

	sampledRowsBytes, _ := json.Marshal(sampledRows)
	// TODO :换为内部调用，方便统计费用 @宋浩
	return RequestIsFirstColumnRowTitleAgent(ctx, string(sampledRowsBytes))
}

// TransposeDocuments 行和列转置，保留原始 metadata，新的表头使用原始第一列的数据
func (dc *dataClean) transposeDocuments(docs []*schema.Document) []*schema.Document {
	if len(docs) == 0 {
		return nil
	}

	var transposedRows [][]string

	// 先转置内容，拆分列后行列对调
	for rowIdx, doc := range docs {
		cols := strings.Split(doc.Content, "\t")

		for colIdx, val := range cols {
			for len(transposedRows) <= colIdx {
				transposedRows = append(transposedRows, []string{})
			}
			// 确保当前转置行有足够的列
			for len(transposedRows[colIdx]) <= rowIdx {
				transposedRows[colIdx] = append(transposedRows[colIdx], "")
			}
			transposedRows[colIdx][rowIdx] = val
		}
	}

	result := make([]*schema.Document, len(transposedRows))

	// 收集原始第一列作为新的表头
	var newHeaders []string
	if len(docs) > 0 {
		newHeaders = make([]string, len(docs))
		for i, doc := range docs {
			cols := strings.Split(doc.Content, "\t")
			if len(cols) > 0 {
				newHeaders[i] = cols[0]
			} else {
				newHeaders[i] = fmt.Sprintf("row_%d", i)
			}
		}
	}

	// 生成转置后的文档
	for i, row := range transposedRows {
		newMetaData := make(map[string]any)

		// 保留第一个文档的非 _row metadata（如 _ext 等）
		if len(docs) > 0 && docs[0].MetaData != nil {
			for k, v := range docs[0].MetaData {
				if k != "_row" { // 不直接复制 _row，因为要重新构建
					newMetaData[k] = v
				}
			}
		}

		// 构建转置后的 _row metadata
		// 使用原来的第一列的值作为 key，当前转置行的值作为 value
		transposedRowMeta := make(map[string]any)
		for j, val := range row {
			if j < len(newHeaders) {
				transposedRowMeta[newHeaders[j]] = val
			} else {
				transposedRowMeta[fmt.Sprintf("col_%d", j)] = val
			}
		}
		newMetaData["_row"] = transposedRowMeta

		// 添加标记，表明已经转置
		newMetaData[KeyIsTransposed] = true

		result[i] = &schema.Document{
			ID:       strconv.Itoa(i + 1),
			Content:  strings.Join(row, "\t"),
			MetaData: newMetaData,
		}
	}

	return result
}

func (dc *dataClean) getHeaders(ctx *gin.Context, sampledDocs []*schema.Document) ([]string, int, []string, error) {
	sampledRows := make([][]string, 0, len(sampledDocs))
	for _, v := range sampledDocs {
		sampledRows = append(sampledRows, strings.Split(v.Content, "\t"))
	}

	sampledRowsBytes, _ := json.Marshal(sampledRows)
	return RequestHeaderRowAgent(ctx, string(sampledRowsBytes))
}

func (dc *dataClean) buildExcelColumnMetaData(excelColumnNames []string, tableColumnNames []string, excelColumnNameTypeMap map[string]string) []foresttype.ExcelColumnMetadata {
	duplicateColumn := make(map[string]int)
	for _, v := range excelColumnNames {
		duplicateColumn[v]++
	}
	tableColumList := make([]foresttype.ExcelColumnMetadata, 0, len(excelColumnNames))
	for _, v := range excelColumnNames {
		tableColumnName := xlsxparser.NormalizeHeader(v)
		if len(tableColumnName) > 40 {
			tableColumnName = tableColumnName[:40]
		}
		if duplicateColumn[v] > 1 {
			tableColumnName = fmt.Sprintf("%s_%s", tableColumnName, random.Alphabet(8))
		}
		tableColumList = append(tableColumList, foresttype.ExcelColumnMetadata{
			ExcelColumnName: v,
			TableColumnName: tableColumnName,
			TableColumnType: excelColumnNameTypeMap[v],
		})
	}
	return tableColumList
}

func (dc *dataClean) getSampledRows(headerRowNum int, sheetDocs []*schema.Document) []*schema.Document {
	totalRow := len(sheetDocs)
	if headerRowNum < 1 || headerRowNum >= totalRow || len(sheetDocs) == 0 {
		return nil
	}

	// 计算采样行范围
	totalSampledRow := totalRow - headerRowNum
	if totalSampledRow > 300 {
		totalSampledRow = totalSampledRow / 2
	}
	if totalSampledRow > 1000 {
		totalSampledRow = 1000
	}

	dataDocs := sheetDocs[headerRowNum:]

	// 数据采样
	totalDataRow := len(dataDocs)
	sampleInterval := totalDataRow / totalSampledRow
	if sampleInterval < 1 {
		sampleInterval = 1
	}

	sampledRows := 0
	sampledDocs := make([]*schema.Document, 0, totalSampledRow)

	// 随机采样
	for i := 0; i < totalDataRow && sampledRows < totalSampledRow; i++ {
		if rand.Intn(sampleInterval) == 0 {
			sampledDocs = append(sampledDocs, dataDocs[i])
			sampledRows++
		}
	}

	return sampledDocs
}

func (dc *dataClean) buildTableName() string {
	return fmt.Sprintf("ke_excel_temp_%s_%s", time.Now().Format("20060102"), strings.ToLower(random.Alphabet(8)))
}

func (dc *dataClean) buildTableDDL(tableName string, excelColumnMetaDataList []foresttype.ExcelColumnMetadata, tableComment string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))
	sb.WriteString("  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',\n")
	sb.WriteString("  `excel_row_number` INT COMMENT 'Excel行号',\n")

	for _, v := range excelColumnMetaDataList {
		// 格式化列定义
		sb.WriteString(fmt.Sprintf("  `%s` %s COMMENT '%s',\n", v.TableColumnName, v.TableColumnType, v.ExcelColumnName))
		sb.WriteString(fmt.Sprintf("  `excel_origin_%s` TEXT COMMENT '%s_原始值',\n", v.TableColumnName, v.ExcelColumnName))
	}

	sb.WriteString("  `excel_abnormal` TEXT COMMENT '异常列',\n")
	sb.WriteString("  PRIMARY KEY (`id`)\n")
	sb.WriteString(fmt.Sprintf(") COMMENT='%s'", tableComment))

	return sb.String()
}

func (dc *dataClean) buildTableRecord(doc *schema.Document, tableColumnMetaDataList []foresttype.ExcelColumnMetadata) (map[string]any, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	if len(tableColumnMetaDataList) == 0 {
		return nil, fmt.Errorf("tableColumnMetaDataList is empty")
	}

	// 解析content，按制表符分割
	content := strings.TrimSpace(doc.Content)
	if content == "" {
		return nil, fmt.Errorf("document content is empty")
	}

	row := strings.Split(content, "\t")

	// 构建record
	record := make(map[string]any)
	var abnormalColumns []string

	// 从MetaData中获取行号，如果没有则默认为1

	// 设置行号
	excelRowNum, _ := strconv.Atoi(doc.ID)
	record["excel_row_number"] = excelRowNum + 1

	// 处理每一列
	for col, columnMeta := range tableColumnMetaDataList {
		// 表列名
		colName := columnMeta.TableColumnName
		// 列类型
		colType := columnMeta.TableColumnType
		// 原始数据列名
		realColName := "excel_origin_" + colName

		if col < len(row) {
			valStr := strings.TrimSpace(row[col])
			// 原始数据
			record[realColName] = valStr

			// 转换值
			convertedValue := dc.convertValueFromString(valStr, colType)
			record[colName] = convertedValue

			// 检查类型匹配
			if !dc.isValueTypeMatch(convertedValue, colType) {
				abnormalColumns = append(abnormalColumns, colName)
			}
		} else {
			// 列数不足，设置默认值
			record[realColName] = ""
			record[colName] = dc.getTypeDefaultValue(colType)
			abnormalColumns = append(abnormalColumns, colName)
		}
	}

	// 设置异常列字段
	if len(abnormalColumns) > 0 {
		record["excel_abnormal"] = strings.Join(abnormalColumns, ",")
	} else {
		record["excel_abnormal"] = ""
	}

	return record, nil
}

// convertValueFromString 从字符串转换值（简化版，不需要Excel文件对象）
func (dc *dataClean) convertValueFromString(valStr string, colType string) interface{} {
	if valStr == "" || valStr == "-" {
		return dc.getTypeDefaultValue(colType)
	}

	switch colType {
	case "DOUBLE", "FLOAT":
		// 尝试解析为浮点数
		if floatVal, err := strconv.ParseFloat(valStr, 64); err == nil {
			if floatVal == math.Trunc(floatVal) {
				return int64(floatVal) // 如果是整数返回int64
			}
			return floatVal
		}
		return 0.0

	case "DATETIME":
		// 首先尝试作为Excel日期数字解析
		if excelDate, err := strconv.ParseFloat(valStr, 64); err == nil {
			return dc.convertExcelDate(excelDate)
		}
		// 尝试解析为日期字符串
		if parsedDate, err := dc.parseDateString(valStr); err == nil {
			return parsedDate
		}
		return nil

	default: // TEXT
		return valStr
	}
}

// isValueTypeMatch 判断值与推断类型是否匹配
func (dc *dataClean) isValueTypeMatch(value interface{}, colType string) bool {
	switch colType {
	case "DOUBLE", "FLOAT":
		// 检查是否是float64或int64
		_, isFloat := value.(float64)
		_, isInt := value.(int64)
		return isFloat || isInt

	case "DATETIME":
		// 检查是否是time.Time
		_, isTime := value.(time.Time)
		return isTime

	case "TEXT":
		// 检查是否是string
		_, isString := value.(string)
		return isString

	default:
		return false
	}
}

// getTypeDefaultValue 当原始值与列推断类型不匹配，则返回默认值
func (dc *dataClean) getTypeDefaultValue(colType string) interface{} {
	switch colType {
	case "DOUBLE", "FLOAT":
		return 0.0
	case "DATETIME":
		return nil
	default: // TEXT
		return nil
	}
}

// convertExcelDate 将Excel日期转换为时间，浮点数
// Excel日期是一个基于1900年1月1日的自增天数
func (dc *dataClean) convertExcelDate(excelDate float64) time.Time {
	baseDate := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	days := int(excelDate)
	fraction := excelDate - float64(days)
	date := baseDate.AddDate(0, 0, days)

	totalSeconds := int(fraction * 86400)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return date.Add(time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second)
}

// parseDateString 解析日期字符串
func (dc *dataClean) parseDateString(dateStr string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	dateStr = strings.TrimSpace(dateStr)

	// 首先尝试解析为Excel日期数字
	if excelDate, err := strconv.ParseFloat(dateStr, 64); err == nil {
		return dc.convertExcelDate(excelDate), nil
	}

	formats := []string{
		"2/1/2006",            // 新增日/月/年格式
		"2/1/06 15:04",        // D/M/YY HH:mm
		"2/1/2006 15:04",      // D/M/YYYY HH:mm
		"2/1/2006 15:04:05",   // D/M/YYYY HH:mm:ss
		"02/01/06 15:04",      // DD/MM/YY HH:mm
		"02/01/06 15:04:05",   // DD/MM/YY HH:mm:ss
		"02/01/2006 15:04",    // DD/MM/YYYY HH:mm
		"02/01/2006 15:04:05", // DD/MM/YYYY HH:mm:ss
		"1/2/2006",
		"1/2/06 15:04",        // M/D/YY HH:mm (如 4/7/24 23:58)
		"1/2/2006 15:04",      // M/D/YYYY HH:mm
		"1/2/2006 15:04:05",   // M/D/YYYY HH:mm:ss
		"01/02/06 15:04",      // MM/DD/YY HH:mm
		"01/02/06 15:04:05",   // MM/DD/YY HH:mm:ss
		"01/02/2006 15:04",    // MM/DD/YYYY HH:mm
		"01/02/2006 15:04:05", // MM/DD/YYYY HH:mm:ss
		"2006/01/02 15:04",    // YYYY/MM/DD HH:mm
		"1-2-06 15:04",        // M-D-YY HH:mm
		"1-2-2006 15:04",      // M-D-YYYY HH:mm
		"01-02-06 15:04",      // MM-DD-YY HH:mm (新增)
		"01-02-2006 15:04",    // MM-DD-YYYY HH:mm
		"01-02-06",            // MM-DD-YY (新增)
		"01-02-2006",          // MM-DD-YYYY
		"2006-01-02 15:04:05", // YYYY-MM-DD HH:mm:ss
		"2006-01-02 15:04",    // YYYY-MM-DD HH:mm
		"2006-01-02",          // YYYY-MM-DD
		// 中国常用格式
		"2006年01月02日 15时04分05秒",
		"2006年01月02日",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"2006/1/2",
		"2006.1.2",
		// 美国常用格式
		"01/02/2006 03:04:05 PM", // 12小时制带AM/PM
		"01/02/2006 15:04:05",    // 24小时制
		"1/2/2006 3:04:05 PM",
		// 欧洲常用格式
		"02.01.2006 15:04:05", // 德语区格式
		"02/01/2006 15:04:05", // 英式格式
		"02-Jan-2006 15:04:05",
		// 其他国际格式
		"02-Jan-06 15:04:05",
		"Jan 02, 2006 15:04:05",
		"January 02, 2006 15:04:05",
		// 无分隔符格式
		"20060102150405", // 紧凑格式
		"060102150405",   // 短年紧凑格式
		"15:04:05",       // 仅时间
		"3:04:05 PM",     // 12小时制时间
		// 特殊格式
		time.RFC3339,  // 2006-01-02T15:04:05Z07:00
		time.RFC1123,  // Mon, 02 Jan 2006 15:04:05 MST
		time.ANSIC,    // Mon Jan _2 15:04:05 2006
		time.UnixDate, // Mon Jan _2 15:04:05 MST 2006
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (dc *dataClean) batchInsertDocs(ctx *gin.Context, db *gorm.DB, tableName string, recordList []map[string]any) error {
	if len(recordList) == 0 {
		return nil
	}

	// 每批次插入的行数
	const batchSize = 100

	// 分批处理
	for start := 0; start < len(recordList); start += batchSize {
		end := start + batchSize
		if end > len(recordList) {
			end = len(recordList)
		}

		// 取当前批次
		group := recordList[start:end]
		if len(group) == 0 {
			break
		}

		// 执行插入
		if err := db.WithContext(ctx).Table(tableName).Create(group).Error; err != nil {
			return err
		}
	}

	return nil
}

func (dc *dataClean) groupSheetDocs(ctx context.Context, docs []*schema.Document) map[string][]*schema.Document {
	docsGroup := make(map[string][]*schema.Document)
	for _, doc := range docs {
		subSheetIDVal := doc.MetaData[xlsxparser.SubSheetID]
		subSheetID, ok := subSheetIDVal.(string)
		if !ok {
			logs.ErrorContextf(ctx, "[dc.groupSheetDocs] assert subSheetID failed, subSheetIDVal: %v")
			continue
		}

		docsGroup[subSheetID] = append(docsGroup[subSheetID], doc)
	}
	return docsGroup
}

func (dc *dataClean) getHeaderSummary(headerDocs []string) (string, int) {
	if len(headerDocs) == 0 {
		return "", 0
	}

	seen := make(map[string]bool)
	var summary []string
	maxDuplicateRowNum := 0
	var prevLine string

	for i, line := range headerDocs {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		val, ok := IsMergedRow(line)
		isDuplicate := (prevLine != "" && line == prevLine)

		// 不是合并单元格行，也不是重复行 -> 停止
		if !ok && !isDuplicate {
			maxDuplicateRowNum = i + 1
			break
		}

		if ok {
			// 去重逻辑：如果当前合并单元格内容没出现过
			if !seen[val] {
				seen[val] = true
				summary = append(summary, val)
			}
		}

		prevLine = line
		maxDuplicateRowNum = i + 1
	}

	return strings.Join(summary, ","), maxDuplicateRowNum
}

func (dc *dataClean) getRemark(content string) string {
	val, ok := IsMergedRow(content)
	if !ok {
		return ""
	}
	return val
}

// IsMergedRow 判断一行是否为“合并单元格”行
// 如果所有非空单元格内容都相同，返回 (该值, true)，否则 ("", false)
func IsMergedRow(line string) (string, bool) {
	cells := strings.Split(line, "\t")

	var firstVal string
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}
		if firstVal == "" {
			firstVal = cell
		} else if cell != firstVal {
			// 出现不同内容，说明不是合并行
			return "", false
		}
	}

	if firstVal == "" {
		// 全为空，视为不是有效的合并单元格
		return "", false
	}

	return firstVal, true
}
