package chatquestion

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

type StatisticsReq struct {
	AgentID   uint       `json:"agent_id"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

// SearchAgentAllHistory 获取指定时间段内的所以数据
func SearchAgentAllHistory(ctx context.Context, req *StatisticsReq) (*QuestionSearchResult, error) {
	// 初始查询
	var fullResult QuestionSearchResult
	batchSize := 1000
	sid := ""
	var err error
	for {
		var resp *esapi.Response
		if sid == "" {
			var buf bytes.Buffer
			mustMap := []map[string]interface{}{
				{
					"range": map[string]interface{}{
						"created_at": map[string]interface{}{
							"gte": req.StartTime,
							"lte": req.EndTime,
						},
					},
				},
			}
			if req.AgentID > 0 {
				mustMap = []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"base_agent_id": req.AgentID,
						},
					},
					{
						"range": map[string]interface{}{
							"created_at": map[string]interface{}{
								"gte": req.StartTime,
								"lte": req.EndTime,
							},
						},
					},
				}
			}
			query := map[string]interface{}{
				"size":    batchSize,
				"_source": []string{"created_at", "question", "answer"},
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": mustMap,
						"must_not": []map[string]interface{}{
							{
								"term": map[string]interface{}{
									"question.keyword": "",
								},
							},
						},
					},
				},
				"sort": []map[string]interface{}{
					{
						"created_at": map[string]interface{}{
							"order": "asc",
						},
					},
				},
			}
			logs.InfoContextf(ctx, "query: %v", query)
			if err := json.NewEncoder(&buf).Encode(query); err != nil {
				logs.ErrorContextf(ctx, "error encoding query: %s")
				return nil, err
			}
			resp, err = question_es_client.Search(
				question_es_client.Search.WithContext(ctx),
				question_es_client.Search.WithIndex(chattype.HistoryIndex),
				question_es_client.Search.WithBody(&buf),
				question_es_client.Search.WithScroll(time.Minute*10),
			)
			if err != nil {
				logs.ErrorContextf(ctx, "Error getting response: %s", err)
				return nil, err
			}
		} else {
			resp, err = question_es_client.Scroll(
				question_es_client.Scroll.WithScrollID(sid),
				question_es_client.Scroll.WithScroll(time.Minute*10))
			if err != nil {
				logs.ErrorContextf(ctx, "Error getting response: %s", err)
				return nil, err
			}
		}
		defer resp.Body.Close()
		// 读取完整响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "error reading body: %v", err)
			return nil, err
		}
		// 解析JSON响应
		var searchRes QuestionSearchResult
		if err := json.Unmarshal(body, &searchRes); err != nil {
			logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
			return nil, err
		}
		sid = searchRes.ScrollID
		fullResult.Hits.Hits = append(fullResult.Hits.Hits, searchRes.Hits.Hits...)
		if len(searchRes.Hits.Hits) < batchSize {
			break
		}
	}

	return &fullResult, nil
}

// GetAgentHistoryCount 获取count
func GetAgentHistoryCount(ctx context.Context, req *StatisticsReq) (int64, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"base_agent_id": req.AgentID,
						},
					},
					{
						"range": map[string]interface{}{
							"created_at": map[string]interface{}{
								"gte": req.StartTime,
								"lte": req.EndTime,
							},
						},
					},
				},
			},
		},
	}
	logs.InfoContextf(ctx, "query: %v", query)
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		logs.ErrorContextf(ctx, "error encoding query: %s")
		return 0, err
	}
	resp, err := question_es_client.Count(
		question_es_client.Count.WithContext(ctx),
		question_es_client.Count.WithIndex(chattype.HistoryIndex),
		question_es_client.Count.WithBody(&buf),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "Error getting response: %s", err)
		return 0, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return 0, err
	}
	// 解析JSON响应
	var searchRes AgentHistoryCount
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return 0, err
	}
	return searchRes.Count, nil
}

type AgentHistoryCount struct {
	Count int64 `json:"count"`
}
