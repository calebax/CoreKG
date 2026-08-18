package chatclient

import "regexp"

// ExtractCode 提取代码
func ExtractCode(codetype, text string) string {
	re := regexp.MustCompile("(?s)```" + codetype + "\\s*(.*?)\\s*```")
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	} else {
		return text
	}
}
