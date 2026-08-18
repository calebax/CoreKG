package essearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

const (
	EsResultSize = 50
)

// SearchQuestionChunk 根据问题搜索对应的chunk，返回检索结果
func (w *EsSearchWrapper) SearchQuestionChunk() (*SearchResult, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("bool", esquery.BuildMap("should", []esquery.Map{
			esquery.BuildMap("term", esquery.BuildMap("is_disable", false)),
			esquery.BuildMap("bool", esquery.BuildMap("must_not", esquery.BuildMap("exists", esquery.BuildMap("field", "is_disable")))),
		})),
		esquery.BuildMap("exists", esquery.BuildMap("field", "embedding")),
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video"})),
		esquery.BuildMap("script",
			esquery.BuildMap("script",
				esquery.BuildMap("source", "(doc.containsKey('description.keyword') && !doc['description.keyword'].empty && doc['description.keyword'].value.length() > 15)  || doc['description.keyword'].empty"))),
	}
	if len(w.forestIds) != 0 {
		mustQuery = append(mustQuery, esquery.BuildMap("terms", esquery.BuildMap("forest_id", w.forestIds)))
	}
	if len(w.fileIds) != 0 {
		// mustQuery = append(mustQuery, esquery.BuildMap("term", esquery.BuildMap("references.file_id", file_id)))
		// nested := esquery.BuildMap("nested", esquery.BuildMap("path", "references",
		// 	"query", esquery.BuildMap("bool", esquery.BuildMap("must", []esquery.Map{
		// 		esquery.BuildMap("terms", esquery.BuildMap("references.file_id", fileId)),
		// 	}))))
		mustQuery = append(mustQuery, esquery.BuildMap("terms", esquery.BuildMap("file_id", w.fileIds)))
	}
	boolQuery := esquery.BuildMap("must", mustQuery)

	//construct should map
	shouldMap, err := ChatSubQuestion(w.ctx, w.question)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatVersionSchedule error: %v", err)
		return nil, err
	}
	logs.DebugContextf(w.ctx, "kws: %v", shouldMap)
	boolQuery["should"] = []esquery.Map{
		esquery.BuildMap("script_score",
			esquery.BuildMap(
				"query", esquery.BuildMap("match_all", esquery.Map{}),
				"script",
				esquery.BuildMap(
					"source", "double score = cosineSimilarity(params.query_vector, 'embedding') + 1.0; return score > 1.0 ? score * _score : 0;",
					"params", esquery.BuildMap("query_vector", w.embedding)),
			)),
		// esquery.BuildMap("nested", esquery.BuildMap("path", "references",
		// 	"query", esquery.BuildMap("bool", esquery.BuildMap("should", shouldMap)))),

	}
	boolQuery["should"] = append(boolQuery["should"].([]esquery.Map), shouldMap...)
	funcs := []esquery.Map{
		esquery.BuildMap("filter", esquery.BuildMap("terms", esquery.BuildMap("type", []string{"image", "video"})), "weight", 0.65),
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("function_score", esquery.BuildMap("query",
			esquery.BuildMap("bool", boolQuery),
			"functions", funcs,
			"boost_mode", "multiply", // 用乘法降低 image 类型得分
			"score_mode", "sum",
		),
		)).
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

// SearchSubQChunk 根据子问题搜索对应的chunk，返回检索结果
func SearchSubQChunk(ctx context.Context, indexName string, subQuestions []string, uin, forestID uint) ([]byte, error) {
	client, err := InitESClient(ctx)
	if err != nil {
		return nil, err
	}
	var subQs []esquery.Map
	for _, v := range subQuestions {
		subQs = append(subQs, esquery.BuildMap("match", esquery.BuildMap("reference.description", v)))
	}
	queryBytes, err := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("query",
			esquery.BuildMap("bool",
				esquery.BuildMap("filter"), []esquery.Map{
					{"term": esquery.BuildMap("uin", uin)},
					{"term": esquery.BuildMap("forest_id", forestID)},
				}),
			esquery.BuildMap("nested",
				esquery.BuildMap("query", []esquery.Map{
					esquery.BuildMap("path", "reference"),
					esquery.BuildMap("query",
						esquery.BuildMap("should", subQs)),
				})))).SetSize(EsResultSize).
		BuildBytes()
	if err != nil {
		return nil, err
	}
	logs.InfoContextf(ctx, "queryBytes: %v", string(queryBytes))
	resp, err := client.Search(
		client.Search.WithIndex(indexName),
		client.Search.WithBody(bytes.NewBuffer(queryBytes)),
		client.Search.WithContext(context.TODO()),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// SearchChunkSequence 搜索chunk上下文
func (w *EsSearchWrapper) SearchChunkSequence(chunks *SearchResult) (*SearchResult, error) {
	// step 搜索上下文多少个
	var step = 1
	var searchCount = len(chunks.Hits.Hits) * (2*step + 1)

	shoudMap := []esquery.Map{}
	for _, hit := range chunks.Hits.Hits {
		sequence := []int{hit.Source.Sequence}
		for i := 1; i <= step; i++ {
			sequence = append(sequence, hit.Source.Sequence+i)
			sequence = append(sequence, hit.Source.Sequence-i)
		}
		shoud_item := esquery.BuildMap("bool",
			esquery.BuildMap("must", []esquery.Map{
				esquery.BuildMap("bool", esquery.BuildMap("should", []esquery.Map{
					esquery.BuildMap("term", esquery.BuildMap("is_disable", false)),
					esquery.BuildMap("bool", esquery.BuildMap("must_not", esquery.BuildMap("exists", esquery.BuildMap("field", "is_disable")))),
				})),
				esquery.BuildMap("term", esquery.BuildMap("file_id", hit.Source.FileID)),
				esquery.BuildMap("terms", esquery.BuildMap("sequence", sequence))}))
		shoudMap = append(shoudMap, shoud_item)
	}
	mustQuery := []esquery.Map{
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"chunk", "image", "table", "video"})),
		esquery.BuildMap("bool",
			esquery.BuildMap(
				"should", shoudMap,
				"minimum_should_match", 1)),
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool",
			esquery.BuildMap("must", mustQuery))).
		SetSize(searchCount).
		Set("_source", esquery.BuildMap("excludes", []string{"embedding"}))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(w.ctx, "SearchChunkSequence esquery.BuildMap error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "querybyte: %v", string(querybyte))
	resp, err := w.cli.Search(
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewBuffer(querybyte)),
		w.cli.Search.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "SearchChunkSequence client.Search error: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(w.ctx, "SearchChunkSequence resp.StatusCode error: %s, body: %s", resp.Status(), string(body))
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

func (w *EsSearchWrapper) DeleteType(tp ...ragtypes.ChunkType) error {
	queryJSON, err := json.Marshal(map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{"terms": map[string]interface{}{"type": tp}},
					{"terms": map[string]interface{}{"forest_id": w.forestIds}},
					{"terms": map[string]interface{}{"file_id": w.fileIds}},
				},
			},
		},
	})
	if err != nil {
		logs.ErrorContextf(w.ctx, "EsSearchWrapper.DeleteType to marshal delete query: %v", err)
		return fmt.Errorf("failed to marshal delete query: %w", err)
	}

	res, err := w.cli.DeleteByQuery(
		[]string{w.indexName},
		bytes.NewReader(queryJSON),
		w.cli.DeleteByQuery.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "Error during DeleteByQuery request: %v", err)
		return fmt.Errorf("error during delete_by_query request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		logs.ErrorContextf(w.ctx, "DeleteByQuery failed with status: %s, response: %s", res.Status(), string(bodyBytes))
		return fmt.Errorf("delete_by_query failed with status: %s", res.Status())
	}

	var deleteResp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&deleteResp); err == nil {
		logs.InfoContextf(w.ctx, "DeleteByQuery successful. Documents deleted: %v", deleteResp["deleted"])
	} else {
		logs.InfoContextf(w.ctx, "DeleteByQuery successful.")
	}
	return nil
}

func (w *EsSearchWrapper) Insert(v interface{}) error {
	if v == nil {
		return fmt.Errorf("v is nil")
	}
	//*store record to es
	data, err := json.Marshal(v)
	if err != nil {
		logs.ErrorContextf(w.ctx, "PostRun json.Marshal error: %v", err)
		return err
	}

	resp, err := w.cli.Index(
		w.indexName,
		bytes.NewReader(data),
		w.cli.Index.WithDocumentID(uuid.NewString()),
		w.cli.Index.WithContext(w.ctx),
	)
	if err != nil {
		logs.ErrorContextf(w.ctx, "EsSearchWrapper.Insert error: %v", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.IsError() {
		logs.ErrorContextf(w.ctx, "Insert failed with status: %s, response: %s", resp.Status(), string(bodyBytes))
		return fmt.Errorf("insert failed with status: %s", resp.Status())
	}
	logs.DebugContextf(w.ctx, "EsSearchWrapper.Insert response: %v", bodyBytes)

	return nil
}

func (w *EsSearchWrapper) GetFileDesc() (*ragtypes.FileDescription, error) {
	query := map[string]interface{}{
		"size": 1,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"type": ragtypes.ChunkTypeFileDescription}},
					{"terms": map[string]interface{}{"forest_id": w.forestIds}},
					{"terms": map[string]interface{}{"file_id": w.fileIds}},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	res, err := w.cli.Search(
		w.cli.Search.WithContext(w.ctx),
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	if res.IsError() {
		return nil, fmt.Errorf("search failed with status: %s, response: %s", res.Status(), string(bodyBytes))
	}

	var result FileDescResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	if len(result.Hits.Hits) > 0 {
		return result.Hits.Hits[0].Source, nil
	}

	logs.DebugContextf(w.ctx, "No matching document found.")
	return nil, nil
}

func (w *EsSearchWrapper) GetFileChunk(tps ...ragtypes.ChunkType) ([]Hits, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{"terms": map[string]interface{}{"type": tps}},
					{"terms": map[string]interface{}{"forest_id": w.forestIds}},
					{"terms": map[string]interface{}{"file_id": w.fileIds}},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}
	logs.DebugContextf(w.ctx, "PostRun json.Marshal request: %v", string(queryJSON))
	res, err := w.cli.Search(
		w.cli.Search.WithContext(w.ctx),
		w.cli.Search.WithIndex(w.indexName),
		w.cli.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	logs.DebugContextf(w.ctx, "response: %s", string(bodyBytes))
	if res.IsError() {
		return nil, fmt.Errorf("search failed with status: %s, response: %s", res.Status(), string(bodyBytes))
	}

	var result SearchResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	logs.DebugContextf(w.ctx, "resp:%v", result)
	if len(result.Hits.Hits) > 0 {
		return result.Hits.Hits, nil
	}

	logs.DebugContextf(w.ctx, "No matching document found.")
	return nil, nil
}
