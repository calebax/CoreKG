package apipath

import "regexp"

// ExtractAPIFromRequestURI 提取API路径，去掉版本号
func ExtractAPIFromRequestURI(requestURI string) string {
	return regexp.MustCompile(`^/v\d+/`).ReplaceAllString(requestURI, "")
}
