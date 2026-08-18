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

func (w *EsSearchWrapper) SearchChunkWithTitle() (*SearchResult, error) {
	titles, err := w.SearchTitle()
	if err != nil {
		logs.ErrorContextf(w.ctx, "SearchTitle err : %v", titles)
		return nil, err
	}
	if len(titles.Hits.Hits) == 0 {
		return titles, nil
	}
	chunks, err := w.SearchTitleChunk(titles)
	if err != nil {
		logs.ErrorContextf(w.ctx, "SearchTitleChunk err : %v", titles)
		return nil, err
	}
	return chunks, nil
}

// SearchTitle 搜索问题相关标题
func (w *EsSearchWrapper) SearchTitle() (*SearchResult, error) {
	filterQuery := []esquery.Map{
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"title"})),
	}
	if len(w.forestIds) != 0 {
		filterQuery = append(filterQuery, esquery.BuildMap("terms", esquery.BuildMap("forest_id", w.forestIds)))
	}
	if len(w.fileIds) != 0 {
		filterQuery = append(filterQuery, esquery.BuildMap("terms", esquery.BuildMap("file_id", w.fileIds)))
	}
	boolQuery := esquery.BuildMap("filter", filterQuery)

	boolQuery["must"] = []esquery.Map{
		esquery.BuildMap("match", esquery.BuildMap("content", w.question)),
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery))

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

func (w *EsSearchWrapper) SearchTitleChunk(titles *SearchResult) (*SearchResult, error) {
	// 获取id列表
	chunk_ids := []string{}
	for _, v := range titles.Hits.Hits {
		chunk_ids = append(chunk_ids, v.ID)
	}

	// DSL
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("terms", esquery.BuildMap("title_level_ids", chunk_ids))).
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
