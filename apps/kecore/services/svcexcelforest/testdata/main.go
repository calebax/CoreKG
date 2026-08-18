package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
)

// 配置参数（可改为从 config 文件/环境变量读取）
var (
	MySQLDSN       = "user:password@tcp(127.0.0.1:3306)/testdb?parseTime=true&charset=utf8mb4&loc=Local"
	BatchSize      = 500               // 每次批量插入行数
	SampleRows     = 50                // 推断类型时采样多少行
	MaxVarcharLen  = 1024              // 最大 varchar 长度（超出改用 TEXT）
	TableNameTmpl  = "sheet_%s_row_%d" // 生成表名模板：sheet name + 起始行号
	AllowDropTable = false             // 是否允许覆盖已存在表（危险）
	ConnTimeout    = 30 * time.Second
)

type ColumnType int

const (
	ColUnknown ColumnType = iota
	ColInt
	ColFloat
	ColDatetime
	ColVarchar
)

func (ct ColumnType) String() string {
	switch ct {
	case ColInt:
		return "INT"
	case ColFloat:
		return "DOUBLE"
	case ColDatetime:
		return "DATETIME"
	case ColVarchar:
		return "VARCHAR"
	default:
		return "VARCHAR"
	}
}

// normalizeHeader 将任意字符串转化为合法的 MySQL 列名: snake_case, 字母数字和下划线
func normalizeHeader(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// 替换空白为下划线
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "_")
	// 去掉非法字符
	s = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(s, "")
	if s == "" {
		s = "col"
	}
	// 如果以数字开头，加前缀
	if regexp.MustCompile(`^[0-9]`).MatchString(s) {
		s = "c_" + s
	}
	return s
}

// detectBlocks 将一张 sheet 的行划分成多个表块（以全空行为分隔）
func detectBlocks(rows [][]string) [][2]int {
	var blocks [][2]int
	n := len(rows)
	i := 0
	for i < n {
		// skip empty rows
		for i < n && isRowEmpty(rows[i]) {
			i++
		}
		if i >= n {
			break
		}
		// start of block
		start := i
		for i < n && !isRowEmpty(rows[i]) {
			i++
		}
		end := i - 1
		blocks = append(blocks, [2]int{start, end})
	}
	return blocks
}

func isRowEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// trimTrailingEmptyCols 从表块中去掉尾部全空列
func trimTrailingEmptyCols(rows [][]string, start, end int) (int, int) {
	// find max cols present
	maxCols := 0
	for r := start; r <= end; r++ {
		if len(rows[r]) > maxCols {
			maxCols = len(rows[r])
		}
	}
	if maxCols == 0 {
		return 0, 0
	}
	// determine last non-empty col index
	last := -1
	for c := 0; c < maxCols; c++ {
		nonEmpty := false
		for r := start; r <= end; r++ {
			if c < len(rows[r]) && strings.TrimSpace(rows[r][c]) != "" {
				nonEmpty = true
				break
			}
		}
		if nonEmpty {
			last = c
		}
	}
	if last < 0 {
		return 0, 0
	}
	return 0, last
}

// inferColumnTypes 对给定列的数据样本推断类型
func inferColumnTypes(rows [][]string, headerIdx int, startRow int, endRow int, lastCol int) ([]ColumnType, []int) {
	cols := lastCol + 1
	types := make([]ColumnType, cols)
	maxLen := make([]int, cols)

	// default start: unknown
	for c := 0; c < cols; c++ {
		types[c] = ColUnknown
	}

	counted := 0
	for r := headerIdx + 1; r <= endRow && counted < SampleRows; r++ {
		for c := 0; c < cols; c++ {
			var cell string
			if c < len(rows[r]) {
				cell = strings.TrimSpace(rows[r][c])
			}
			if cell == "" {
				continue
			}
			if len(cell) > maxLen[c] {
				maxLen[c] = len(cell)
			}
			// try parse int
			if types[c] == ColUnknown || types[c] == ColInt {
				if isInt(cell) {
					types[c] = ColInt
					continue
				}
			}
			// try float
			if types[c] == ColUnknown || types[c] == ColInt || types[c] == ColFloat {
				if isFloat(cell) {
					types[c] = ColFloat
					continue
				}
			}
			// try datetime
			if types[c] == ColUnknown || types[c] == ColDatetime {
				if isDatetime(cell) {
					types[c] = ColDatetime
					continue
				}
			}
			// fallback to varchar
			types[c] = ColVarchar
		}
		counted++
	}

	// For any remaining unknown cols (all blank in sample), mark as varchar
	for c := 0; c < cols; c++ {
		if types[c] == ColUnknown {
			types[c] = ColVarchar
		}
		// if varchar and maxLen>MaxVarcharLen -> TEXT or use TEXT later
	}
	return types, maxLen
}

var intRegex = regexp.MustCompile(`^-?\d+$`)
var floatRegex = regexp.MustCompile(`^-?(\d+)(\.\d+)?([eE][+-]?\d+)?$`)

func isInt(s string) bool {
	return intRegex.MatchString(s)
}

func isFloat(s string) bool {
	return floatRegex.MatchString(s)
}

// isDatetime 尝试多种格式解析为时间
func isDatetime(s string) bool {
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		// pure numeric could be Excel serial date --> but here we ignore Excel's numeric date special-case in simple mode
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"02/01/2006",
		"2006.01.02",
		"02-Jan-2006",
	}
	for _, l := range layouts {
		if _, err := time.Parse(l, s); err == nil {
			return true
		}
	}
	// also try RFC3339
	if _, err := time.Parse(time.RFC3339, s); err == nil {
		return true
	}
	return false
}

// buildCreateTableSQL 根据推断的列类型生成 CREATE TABLE 语句
func buildCreateTableSQL(tableName string, headers []string, types []ColumnType, maxLens []int) (string, error) {
	if len(headers) != len(types) {
		return "", errors.New("headers and types length mismatch")
	}
	cols := make([]string, 0, len(headers))
	seen := map[string]int{}
	for i, h := range headers {
		col := normalizeHeader(h)
		if existing, ok := seen[col]; ok {
			seen[col] = existing + 1
			col = fmt.Sprintf("%s_%d", col, existing+1)
		} else {
			seen[col] = 0
		}
		sqlType := ""
		switch types[i] {
		case ColInt:
			sqlType = "BIGINT"
		case ColFloat:
			sqlType = "DOUBLE"
		case ColDatetime:
			sqlType = "DATETIME"
		default:
			if maxLens[i] > MaxVarcharLen {
				sqlType = "TEXT"
			} else {
				n := maxLens[i]
				if n < 1 {
					n = 255
				}
				if n > MaxVarcharLen {
					n = MaxVarcharLen
				}
				sqlType = fmt.Sprintf("VARCHAR(%d)", n)
			}
		}
		cols = append(cols, fmt.Sprintf("`%s` %s", col, sqlType))
	}
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;", tableName, strings.Join(cols, ",\n"))
	return createSQL, nil
}

// prepareInsertSQL 生成批量插入的占位符和 SQL 语句
func prepareInsertSQL(tableName string, headers []string, batchSize int) (string, error) {
	if len(headers) == 0 {
		return "", errors.New("no headers")
	}
	colNames := make([]string, len(headers))
	for i, h := range headers {
		colNames[i] = "`" + normalizeHeader(h) + "`"
	}
	placeholderRow := "(" + strings.TrimRight(strings.Repeat("?,", len(headers)), ",") + ")"
	placeholders := strings.TrimRight(strings.Repeat(placeholderRow+",", batchSize), ",")
	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", tableName, strings.Join(colNames, ","), placeholders)
	return sql, nil
}

// safeGet returns cell value or empty
func safeGet(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func main() {
	// 示例： 从命令行或硬编码读取文件路径和 sheet
	// xlsxPath := "言古研发_需求_20250914104431.xlsx"
	xlsxPath := "3466.xlsx"
	sheetName := "全市和分区"

	// open excel
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		log.Fatalf("open excel fail: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Fatalf("get rows fail: %v", err)
	}
	if len(rows) == 0 {
		log.Fatalf("sheet empty")
	}

	blocks := detectBlocks(rows)
	log.Printf("detected %d table blocks on sheet %s", len(blocks), sheetName)

	// open DB
	// db, err := sql.Open("mysql", MySQLDSN)
	// if err != nil {
	// 	log.Fatalf("open db fail: %v", err)
	// }
	ctx, cancel := context.WithTimeout(context.Background(), ConnTimeout)
	defer cancel()
	_ = ctx
	// if err := db.PingContext(ctx); err != nil {
	// 	log.Fatalf("db ping fail: %v", err)
	// }

	// process each block
	for bi, b := range blocks {
		start, end := b[0], b[1]
		log.Printf("block %d rows %d..%d", bi, start, end)
		// determine col range
		_, lastCol := trimTrailingEmptyCols(rows, start, end)
		if lastCol < 0 {
			log.Printf("block %d is empty columns, skip", bi)
			continue
		}
		// header: assume first row of the block
		headerRowIdx := start
		header := make([]string, lastCol+1)
		for c := 0; c <= lastCol; c++ {
			header[c] = safeGet(rows[headerRowIdx], c)
			if strings.TrimSpace(header[c]) == "" {
				header[c] = fmt.Sprintf("col_%d", c+1)
			}
		}
		types, maxLens := inferColumnTypes(rows, headerRowIdx, start, end, lastCol)
		// build table name
		tabName := fmt.Sprintf(TableNameTmpl, sanitizeForName(sheetName), start+1)
		createSQL, err := buildCreateTableSQL(tabName, header, types, maxLens)
		if err != nil {
			log.Printf("build create table fail: %v", err)
			continue
		}
		log.Printf("create table sql:\n%s", createSQL)
		if AllowDropTable {
			// optional: drop table then create
		}
		// create table
		// if _, err := db.ExecContext(ctx, createSQL); err != nil {
		// 	log.Printf("create table exec fail: %v", err)
		// 	continue
		// }
		// insert data rows in batches
		dataRows := [][]string{}
		for r := headerRowIdx + 1; r <= end; r++ {
			// build a row with exactly lastCol+1 columns
			rec := make([]string, lastCol+1)
			for c := 0; c <= lastCol; c++ {
				rec[c] = safeGet(rows[r], c)
			}
			// skip fully empty rows
			if isRowEmpty(rec) {
				continue
			}
			dataRows = append(dataRows, rec)
		}
		if len(dataRows) == 0 {
			log.Printf("no data rows in block %d", bi)
			continue
		}
		// perform batched inserts
		colsCount := len(header)
		// pre-generate base insert sql with batchSize = 1 placeholder, we'll adapt per batch
		for i := 0; i < len(dataRows); i += BatchSize {
			j := i + BatchSize
			if j > len(dataRows) {
				j = len(dataRows)
			}
			batch := dataRows[i:j]
			// build placeholders and args
			placeholderRow := "(" + strings.TrimRight(strings.Repeat("?,", colsCount), ",") + ")"
			placeholders := strings.TrimRight(strings.Repeat(placeholderRow+",", len(batch)), ",")
			colNames := make([]string, colsCount)
			for ci, h := range header {
				colNames[ci] = "`" + normalizeHeader(h) + "`"
			}
			insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", tabName, strings.Join(colNames, ","), placeholders)
			args := []interface{}{}
			for _, rec := range batch {
				for c := 0; c < colsCount; c++ {
					val := strings.TrimSpace(rec[c])
					// basic conversion per inferred type
					switch types[c] {
					case ColInt:
						if val == "" {
							args = append(args, nil)
						} else {
							iv, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								// fallback to null or original string
								args = append(args, nil)
							} else {
								args = append(args, iv)
							}
						}
					case ColFloat:
						if val == "" {
							args = append(args, nil)
						} else {
							fv, err := strconv.ParseFloat(val, 64)
							if err != nil {
								args = append(args, nil)
							} else {
								args = append(args, fv)
							}
						}
					case ColDatetime:
						if val == "" {
							args = append(args, nil)
						} else {
							// try parse with several layouts
							t, ok := tryParseTime(val)
							if ok {
								args = append(args, t)
							} else {
								args = append(args, val)
							}
						}
					default:
						if val == "" {
							args = append(args, nil)
						} else {
							args = append(args, val)
						}
					}
				}
			}
			log.Printf("batch %d..%d: %d rows", i, j-1, len(batch))
			log.Printf("insert sql:\n%s", insertSQL)
			log.Printf("args: %v", args)
			// exec in transaction
			// tx, err := db.BeginTx(ctx, nil)
			// if err != nil {
			// 	log.Printf("begin tx fail: %v", err)
			// 	continue
			// }
			// _, err = tx.ExecContext(ctx, insertSQL, args...)
			// if err != nil {
			// 	tx.Rollback()
			// 	log.Printf("batch insert fail: %v, sql=%s", err, insertSQL)
			// } else {
			// 	if err := tx.Commit(); err != nil {
			// 		log.Printf("tx commit fail: %v", err)
			// 	}
			// }
		}
		log.Printf("block %d inserted %d rows into %s", bi, len(dataRows), tabName)
	}

	log.Printf("done")
}

// sanitizeForName 把 sheet 名/其它字符转成表名安全字符
func sanitizeForName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "_")
	s = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(s, "")
	if s == "" {
		s = "sheet"
	}
	return s
}

func tryParseTime(s string) (time.Time, bool) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"02/01/2006",
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	// Excel may store dates as floats (serial), but excelize's GetRows already converts formatted dates to strings,
	// so handling Excel serial numbers requires using GetCellValue with styles or StreamReader — omitted for brevity.
	return time.Time{}, false
}
