package es

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type RequestInfo struct {
	Method string
	Path   string
	Body   string
}

func parseMultipleDSLFile(filename string) ([]*RequestInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseMultipleDSL(file)
}

func parseMultipleDSL(reader io.Reader) ([]*RequestInfo, error) {
	scanner := bufio.NewScanner(reader)
	var requests []*RequestInfo
	var currentRequest *RequestInfo
	var bodyLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否是新的请求开始（PUT/POST/GET/DELETE 开头）
		if isRequestStart(line) {
			// 如果之前有请求在处理，保存它
			if currentRequest != nil && len(bodyLines) > 0 {
				currentRequest.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
				requests = append(requests, currentRequest)
			}

			// 开始新的请求
			method, path := parseRequestLine(line)
			currentRequest = &RequestInfo{
				Method: method,
				Path:   path,
			}
			bodyLines = []string{}
		} else if currentRequest != nil {
			// 收集body行
			bodyLines = append(bodyLines, line)
		}
	}

	// 保存最后一个请求
	if currentRequest != nil && len(bodyLines) > 0 {
		currentRequest.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		requests = append(requests, currentRequest)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func isRequestStart(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return false
	}

	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
		"PATCH":  true,
	}

	return validMethods[parts[0]]
}

func parseRequestLine(line string) (method, path string) {
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
