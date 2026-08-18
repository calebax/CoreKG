package svcexcelforest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/einodocument/parser/xlsxparser"
	"github.com/xuri/excelize/v2"
)

// customColumnNameTypeMap 用于覆盖自动类型识别中容易误判的字段
var customColumnNameTypeMap = map[string]string{
	"是否":   "VARCHAR(32)",
	"是否启用": "VARCHAR(32)",
	"是否删除": "VARCHAR(32)",
	"是否有效": "VARCHAR(32)",
	"是否公开": "VARCHAR(32)",
	"是否完成": "VARCHAR(32)",
	"是否通过": "VARCHAR(32)",
	"是否审核": "VARCHAR(32)",
	"是否激活": "VARCHAR(32)",
}

func getSheetList(ctx *gin.Context, filePublicURL string) ([]string, error) {
	resp, err := http.Get(filePublicURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download excel: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download excel, status code: %d", resp.StatusCode)
	}
	f, err := excelize.OpenReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel: %v", err)
	}
	defer f.Close()

	return f.GetSheetList(), nil
}

func getSheetDocs(ctx *gin.Context, filePublicURL, sheetName string, metaData map[string]any) ([]*schema.Document, error) {
	xlsxParser := xlsxparser.NewXlsxParser()
	loader, err := url.NewLoader(ctx, &url.LoaderConfig{
		Parser: xlsxParser,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get loader: %v", err)
	}
	loadOpts := document.WithParserOptions(parser.WithExtraMeta(metaData), xlsxparser.WithSheetName(sheetName))
	docs, err := loader.Load(ctx, document.Source{
		URI: filePublicURL,
	}, loadOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to load excel: %v", err)
	}
	return docs, nil
}

// dataTypeCounter 数据类型计数器
type dataTypeCounter struct {
	Text     int `json:"text"`
	Float    int `json:"float"`
	BigInt   int `json:"big_int"`
	DateTime int `json:"date_time"`
	Boolean  int `json:"boolean"`
	Total    int `json:"total"`
}

// DetectColumnTypes 判断列类型
// headers: 表头列表
// sampleData: 抽样数据，每个Document的Content包含该行的Tab分隔数据
// 返回: 表头名称 -> 数据类型的映射
func DetectColumnTypes(headers []string, sampleData []*schema.Document) (map[string]string, error) {
	if len(headers) == 0 {
		return make(map[string]string), nil
	}

	// 初始化计数器
	counters := make(map[string]*dataTypeCounter)
	for _, header := range headers {
		counters[header] = &dataTypeCounter{}
	}

	// 分析每个样本数据
	for _, doc := range sampleData {
		if doc == nil || doc.Content == "" {
			continue
		}

		// 解析Tab分隔的数据
		fields := strings.Split(doc.Content, "\t")

		// 分析每个字段
		for i, header := range headers {
			counter := counters[header]

			// 检查字段索引是否有效
			if i >= len(fields) {
				continue
			}

			value := strings.TrimSpace(fields[i])
			if value == "" {
				continue
			}

			counter.Total++
			dataType := analyzeValue(value)

			switch dataType {
			case "TEXT":
				counter.Text++
			case "FLOAT":
				counter.Float++
			case "BIGINT":
				counter.BigInt++
			case "DATETIME":
				counter.DateTime++
			case "BOOLEAN":
				counter.Boolean++
			default:
				counter.Text++
			}
		}
	}

	// 确定最终类型
	result := make(map[string]string)
	for _, header := range headers {
		counter := counters[header]
		finalType := determineFinalType(counter)
		if customType, ok := customColumnNameTypeMap[header]; ok {
			finalType = customType
		}
		result[header] = finalType
	}

	return result, nil
}

// analyzeValue 分析单个值的类型
func analyzeValue(value interface{}) string {
	// 直接处理字符串值（因为从Tab分隔中解析出来的都是字符串）
	strValue := strings.TrimSpace(fmt.Sprintf("%v", value))

	// 空值或占位符
	if strValue == "" || strValue == "-" || strValue == "N/A" || strValue == "NULL" {
		return "TEXT"
	}

	// 1. 检查布尔值
	if isBooleanValue(strValue) {
		return "BOOLEAN"
	}

	// 2. 检查日期时间
	if isDateTimeValue(strValue) {
		return "DATETIME"
	}

	// 3. 检查数值
	if isNumericValue(strValue) {
		if isPotentialBigInt(strValue) {
			return "BIGINT"
		}
		return "FLOAT"
	}

	// 4. 默认为文本
	return "TEXT"
}

// isBooleanValue 检查是否为布尔值
func isBooleanValue(s string) bool {
	lower := strings.ToLower(s)
	switch lower {
	case "true", "false", "yes", "no", "y", "n", "1", "0", "是", "否":
		return true
	}
	return false
}

// isDateTimeValue 检查是否为日期时间值
func isDateTimeValue(dateStr string) bool {
	// 先检查是否包含明显的日期分隔符
	if !strings.ContainsAny(dateStr, "/-:.年月日时分秒") {
		return false
	}

	formats := []string{
		// 标准格式
		"2006-01-02", "2006/01/02", "2006.01.02",
		"2006-01-02 15:04:05", "2006/01/02 15:04:05",
		"2006-01-02 15:04", "2006/01/02 15:04",
		"01-02-2006", "01/02/2006", "01.02.2006",
		"01-02-2006 15:04:05", "01/02/2006 15:04:05",
		"01-02-2006 15:04", "01/02/2006 15:04",
		"02-01-2006", "02/01/2006", "02.01.2006",
		"2-1-2006", "2/1/2006", "1-2-2006", "1/2/2006",

		// 中文格式
		"2006年01月02日", "2006年1月2日",
		"2006年01月02日 15时04分05秒", "2006年01月02日 15:04:05",
		"2006年01月02日 15时04分", "2006年01月02日 15:04",

		// ISO格式
		time.RFC3339, time.RFC3339Nano,
		time.RFC1123, time.RFC822,

		// 简化格式
		"01/02/06", "01-02-06", "2006/1/2", "2006-1-2",
		"15:04:05", "15:04", "3:04 PM", "3:04:05 PM",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return true
		}
	}
	return false
}

// isNumericValue 检查是否为数值
func isNumericValue(s string) bool {
	// 移除常见的货币符号和千位分隔符
	cleaned := strings.ReplaceAll(s, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, "¥", "")
	cleaned = strings.ReplaceAll(cleaned, "€", "")
	cleaned = strings.ReplaceAll(cleaned, "%", "")
	cleaned = strings.TrimSpace(cleaned)

	// 尝试解析为浮点数
	_, err := strconv.ParseFloat(cleaned, 64)
	return err == nil
}

// isPotentialBigInt 检查是否为长整数
func isPotentialBigInt(s string) bool {
	// 移除千位分隔符
	cleaned := strings.ReplaceAll(s, ",", "")
	cleaned = strings.TrimSpace(cleaned)

	// 排除科学计数法和小数点
	if strings.ContainsAny(cleaned, "eE.") {
		return false
	}

	// 检查是否为纯整数且长度超过15位（避免float64精度丢失）
	if len(cleaned) > 15 {
		// 验证是否为纯数字（可能包含正负号）
		if strings.HasPrefix(cleaned, "-") || strings.HasPrefix(cleaned, "+") {
			cleaned = cleaned[1:]
		}
		return strings.Trim(cleaned, "0123456789") == ""
	}

	// 检查是否超出float64安全整数范围
	if val, err := strconv.ParseInt(cleaned, 10, 64); err == nil {
		// JavaScript安全整数范围: -(2^53 - 1) to 2^53 - 1
		const maxSafeInt = 9007199254740991
		return val > maxSafeInt || val < -maxSafeInt
	}

	return false
}

// determineFinalType 确定最终的数据类型
func determineFinalType(counter *dataTypeCounter) string {
	if counter.Total == 0 {
		return "TEXT"
	}

	// 计算各类型占比
	floatRatio := float64(counter.Float) / float64(counter.Total)
	bigIntRatio := float64(counter.BigInt) / float64(counter.Total)
	dateTimeRatio := float64(counter.DateTime) / float64(counter.Total)
	booleanRatio := float64(counter.Boolean) / float64(counter.Total)

	// 置信度阈值：90%
	const confidenceThreshold = 0.9

	// 优先级：DATETIME > BOOLEAN > BIGINT > FLOAT > TEXT
	if dateTimeRatio >= confidenceThreshold {
		return "DATETIME"
	}
	if booleanRatio >= confidenceThreshold {
		return "BOOLEAN"
	}
	if bigIntRatio >= confidenceThreshold {
		return "BIGINT"
	}
	if floatRatio >= confidenceThreshold {
		return "DOUBLE" // 映射为数据库友好的类型
	}

	// 数值类型组合判断：如果 BIGINT + FLOAT >= 90%，优先选择精度更高的类型
	numericRatio := floatRatio + bigIntRatio
	if numericRatio >= confidenceThreshold {
		if bigIntRatio > floatRatio {
			return "BIGINT"
		}
		return "DOUBLE"
	}

	// 默认返回TEXT类型
	return "TEXT"
}
