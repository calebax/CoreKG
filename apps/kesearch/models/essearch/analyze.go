package essearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// Analyze 使用ES拆词器进行拆词
func Analyze(ctx context.Context, text string) (*AnalyzeResultList, error) {
	client, err := InitESClient(ctx)
	if err != nil {
		return nil, err
	}
	query := esquery.NewBuilder().Set("analyzer", "ik_max_word").
		Set("text", text)
	querybyte, err := query.BuildBytes()
	if err != nil {
		return nil, err
	}
	resp, err := client.Indices.Analyze(
		client.Indices.Analyze.WithBody(bytes.NewBuffer(querybyte)),
		client.Indices.Analyze.WithContext(context.Background()),
	)
	if err != nil {
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es query Analyze failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var result *AnalyzeResultList
	if err := json.Unmarshal(body, &result); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	if len(result.Tokens) == 0 {
		result.Tokens = append(result.Tokens, AnalyzeResult{
			Token: text,
		})
	}
	return result, nil
}

type AnalyzeResultList struct {
	Tokens []AnalyzeResult `json:"tokens"`
}

type AnalyzeResult struct {
	Token       string `json:"token"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Type        string `json:"type"`
	Position    int    `json:"position"`
}
