package essearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// DescriptionSearch 或者整个文件的简短描述
func (w *EsSearchWrapper) DescriptionSearch() (*SearchResult, error) {
	searchMap := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"type": "file_description",
						},
					},
				},
			},
		},
		"size": 10000,
		"_source": map[string]interface{}{
			"excludes": []string{"embedding", "mind_map", "abstract"},
		},
	}
	if len(w.forestIds) > 0 {
		searchMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"forest_id": w.forestIds,
				},
			},
		)
	}
	if len(w.fileIds) > 0 {
		searchMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"file_id": w.fileIds,
				},
			},
		)
	}

	querybyte, err := json.Marshal(searchMap)
	if err != nil {
		logs.ErrorContextf(w.ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "querybyte: %v", string(querybyte))
	resp, err := w.cli.Search(
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewBuffer(querybyte)),
		w.cli.Search.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	// 读取完整响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes SearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(w.ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	return &searchRes, nil
}

// SummarizeSearch 搜索对应的摘要
func (w *EsSearchWrapper) SummarizeSearch() (*SearchResult, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("exists", esquery.BuildMap("field", "embedding")),
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"file_description"})),
	}
	if len(w.forestIds) != 0 {
		mustQuery = append(mustQuery, esquery.BuildMap("terms", esquery.BuildMap("forest_id", w.forestIds)))
	}
	if len(w.fileIds) != 0 {
		mustQuery = append(mustQuery, esquery.BuildMap("terms", esquery.BuildMap("file_id", w.fileIds)))
	}

	//construct should map
	shouldMap := []esquery.Map{
		esquery.BuildMap("script_score",
			esquery.BuildMap(
				"query", esquery.BuildMap("match_all", esquery.Map{}),
				"script",
				esquery.BuildMap(
					"source", "double score = cosineSimilarity(params.query_vector, 'embedding') + 1.0; return score > 1.0 ? score * _score : 0;",
					"params", esquery.BuildMap("query_vector", w.embedding)),
			)),
		esquery.BuildMap("match", esquery.BuildMap("description", w.question)),
	}

	boolQuery := esquery.BuildMap("bool", esquery.BuildMap(
		"must", mustQuery,
		"should", shouldMap,
	))
	query := esquery.NewBuilder().
		SetQuery(boolQuery).
		SetSize(EsResultSize).
		Set("_source", esquery.BuildMap("excludes", []string{"embedding", "references"}))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(w.ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "querybyte: %v", string(querybyte))
	resp, err := w.cli.Search(
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewBuffer(querybyte)),
		w.cli.Search.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	// 转换结果
	// 读取完整响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes SearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(w.ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}

	return &searchRes, nil
}
