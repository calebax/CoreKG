package validate

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// type Func func(v interface{}) error

// IsEmail 是否是合法邮箱
func IsEmail(value string) error {
	emailRegexp := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if !emailRegexp.MatchString(value) {
		return fmt.Errorf("invalid format email: %s", value)
	}
	return nil
}

// IsPhone 是否是国内合法手机号
func IsPhone(value string) error {
	reg := `^1([38796][0-9]|14[57]|5[^4])\d{8}$`
	rgx := regexp.MustCompile(reg)
	if rgx.MatchString(value) {
		return nil
	}
	return fmt.Errorf("手机号码 %s 格式错误", value)
}

// IsCardNumber 是否是国内合法身份证号
func IsCardNumber(value string) error {
	if len(value) != 18 && len(value) != 15 {
		return fmt.Errorf("身份证 %s 长度错误", value)
	}
	reg := `(^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx])|([1−9]\d5\d2((0[1−9])|(10|11|12))(([0−2][1−9])|10|20|30|31)\d2[0−9Xx])`
	rgx := regexp.MustCompile(reg)
	if rgx.MatchString(value) {
		return nil
	}
	return fmt.Errorf("身份证 %s 格式错误", value)
}

// IsBankAccountNumber 是否是国内合法银行卡号
func IsBankAccountNumber(value string) error {
	if len(value) != 16 && len(value) != 19 {
		return fmt.Errorf("银行卡号 %s 长度错误", value)
	}
	reg := `^(\d+)$`
	rgx := regexp.MustCompile(reg)
	if rgx.MatchString(value) {
		return nil
	}
	return fmt.Errorf("银行卡号 %s 格式错误", value)
}

// IsLetterNumber 是否是国内合法手机号
func IsLetterNumber(value string) error {
	reg := `^[A-Za-z0-9]+$`
	rgx := regexp.MustCompile(reg)
	if rgx.MatchString(value) {
		return nil
	}
	return fmt.Errorf("内容格式错误 %s", value)
}

// IsUsername 是否为合法用户名
func IsUsername(username string) error {
	// 用户名必须为 3-32 个字符
	runes := []rune(username)
	if len(runes) < 2 || len(runes) > 32 {
		return fmt.Errorf("用户名长度必须为 3-32 个字符")
	}
	// 用户名只能包含字母、数字、中文和破折号（-）
	for _, ch := range runes {
		if !isValidUsernameChar(ch) {
			return fmt.Errorf("用户名只能包含字母、数字、中文和破折号（-）")
		}
	}
	// 用户名不能以破折号（-）开头或结尾
	if runes[0] == '-' || runes[len(runes)-1] == '-' {
		return fmt.Errorf("用户名不能以破折号（-）开头或结尾")
	}
	// 用户名不能包含连续的破折号（--）
	if strings.Contains(username, "--") {
		return fmt.Errorf("用户名不能包含连续的破折号（--）")
	}
	return nil
}

// isValidUsernameChar 检查字符是否为合法的用户名字符
func isValidUsernameChar(ch rune) bool {
	// 字母和数字
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
		return true
	}

	// 破折号
	if ch == '-' {
		return true
	}

	// 中文字符（CJK统一汉字）
	if ch >= 0x4E00 && ch <= 0x9FFF {
		return true
	}

	return false
}

// IsTitle 是否为合法session标题
func IsTitle(value string) error {
	length := utf8.RuneCountInString(value) // 按字符数统计
	if length < 1 {
		return fmt.Errorf("标题长度最短为1个字")
	}
	if length > 50 {
		return fmt.Errorf("标题长度最长50字")
	}
	return nil
}

// IsPassword 是否为合法密码
func IsPassword(value string) error {
	if len(value) < 6 {
		return fmt.Errorf("密码长度过短")
	}
	return nil
}

var (
	defaultNearbyTimeRange = time.Hour * 24 * 365 * 8
)

// IsNearbyTime 是否为附近的时间
func IsNearbyTime(t time.Time, intervals ...time.Duration) error {
	var (
		now        = time.Now()
		begin, end time.Time
	)
	if len(intervals) > 0 {
		begin = now.Add(intervals[0] * -1)
	} else {
		begin = now.Add(defaultNearbyTimeRange * -1)
	}
	if len(intervals) > 1 {
		end = now.Add(intervals[1])
	} else {
		end = now.Add(defaultNearbyTimeRange)
	}

	if t.Before(begin) {
		return fmt.Errorf("invalid time %s, is before %s", t, begin)
	}
	if t.After(end) {
		return fmt.Errorf("invalid time %s, is after %s", t, end)
	}
	return nil
}

// HasCNChar 校验字符串是否包含中文字符
func HasCNChar(str string) bool {
	for _, r := range str {
		// unicode.Is(unicode.Han, r) 用于判断字符 r 是否是汉字
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
