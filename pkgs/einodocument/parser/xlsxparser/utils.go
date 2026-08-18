package xlsxparser

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"github.com/ygpkg/yg-go/random"
)

// HasChinese 判断字符串中是否有中文
func HasChinese(s string) bool {
	for _, r := range s {
		// 判断是否属于汉字范围
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// ConvertChinese 将中文转化为拼音
func ConvertChinese(s string, sep string) string {
	arr := []string{}
	strBuf := ""
	for _, r := range s {
		// 判断是否属于汉字范围
		if unicode.Is(unicode.Han, r) {
			if strBuf != "" {
				arr = append(arr, strBuf)
			}
			strBuf = ""
			pys := pinyin.SinglePinyin(r, pinyin.NewArgs())
			if len(pys) > 0 {
				arr = append(arr, pys[0])
			}
		} else {
			strBuf += string(r)
		}
	}
	if strBuf != "" {
		arr = append(arr, strBuf)
	}
	return strings.Join(arr, sep)
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

// NormalizeHeader 将任意字符串转化为合法的 MySQL 列名: snake_case, 字母数字和下划线
func NormalizeHeader(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// 替换空白为下划线
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "_")
	// 去掉非法字符
	// 判断是否有中文
	s = ConvertChinese(s, "_")
	s = strings.TrimPrefix(s, "_")
	s = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(s, "")
	if s == "" {
		s = "col" + random.Alphabet(8)
	}
	// 如果以数字开头，加前缀
	if regexp.MustCompile(`^[0-9]`).MatchString(s) {
		s = "c_" + s
	}
	if strings.ToLower(strings.TrimSpace(s)) == "id" {
		s = "r_id"
	}
	return s
}
