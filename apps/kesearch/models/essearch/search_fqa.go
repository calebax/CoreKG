package essearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// FindFQAByQuestion 搜索问答对
func (w *EsSearchWrapper) FindFQAByQuestion() (*SearchResult, error) {
	searchRes := SearchResult{}
	if len(w.forestIds) == 0 {
		return &searchRes, nil
	}
	client, err := InitESClient(w.ctx)
	if err != nil {
		return nil, err
	}
	mustQuery := []esquery.Map{
		esquery.BuildMap("terms", esquery.BuildMap("type", []string{"FQA"})),
		esquery.BuildMap("terms", esquery.BuildMap("forest_id", w.forestIds)),
		esquery.BuildMap("term", esquery.BuildMap("enable", 1)),
	}

	boolQuery := esquery.BuildMap("filter", mustQuery)

	boolQuery["should"] = []esquery.Map{
		esquery.BuildMap("script_score",
			esquery.BuildMap("query",
				esquery.BuildMap("match_all", esquery.Map{}),
				"script",
				esquery.BuildMap(
					"source", "double score = cosineSimilarity(params.query_vector, 'embedding') + 1; return score ;",
					"params",
					esquery.BuildMap(
						"query_vector", w.embedding)))),
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSize(EsResultSize).
		Set("min_score", 1+0.95)
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(w.ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(w.ctx, "querybyte: %v", string(querybyte))
	resp, err := client.Search(client.Search.WithIndex(w.indexName), client.Search.WithBody(bytes.NewBuffer(querybyte)), client.Search.WithContext(w.ctx))
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

	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(w.ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}

	return &searchRes, nil
}

type QueryFQAResponse struct {
	apiobj.QueryResponse
	Data []*FQAItem `json:"data"`
}

type Child struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	CreatedAt time.Time `json:"created_at"`
	IsDeleted bool      `json:"is_deleted"`
}

type FQAItem struct {
	Main  *ragtypes.FQA `json:"main"`
	Child []*Child      `json:"child"`
}

// SearchFQA 查询方法
func (w *EsSearchWrapper) SearchFQA() (*QueryFQAResponse, error) {
	boolQuery := w.buildBaseQuery()

	total, err := w.getCount(boolQuery)
	if err != nil {
		logs.ErrorContextf(w.ctx, "get count failed: %v", err)
		return nil, err
	}
	queryList := &QueryFQAResponse{
		QueryResponse: apiobj.QueryResponse{},
	}
	if total == 0 {
		return queryList, nil
	}

	searchQuery := w.buildSearchQuery()
	data, err := w.executeSearch(searchQuery)
	if err != nil {
		logs.ErrorContextf(w.ctx, "execute search failed: %v", err)
		return nil, err
	}
	queryList.Data = data
	queryList.Total = int64(w.Total)
	queryList.Offset = w.pageQuery.Offset
	queryList.Limit = w.pageQuery.Limit
	return queryList, nil
}

// DeleteByQuery 根据查询条件删除
func (w *EsSearchWrapper) DeleteByQuery() error {
	boolQuery := w.buildBaseQuery()

	deleteQuery := esquery.BuildMap("query", esquery.BuildMap("bool", boolQuery))

	queryBytes, err := json.Marshal(deleteQuery)
	if err != nil {
		return err
	}

	resp, err := w.cli.DeleteByQuery([]string{w.indexName}, bytes.NewReader(queryBytes), w.cli.DeleteByQuery.WithContext(w.ctx), w.cli.DeleteByQuery.WithRefresh(true))
	if err != nil {
		logs.ErrorContextf(w.ctx, "client.DeleteByQuery error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es DeleteByQuery failed: %s, error: %s", resp.Status(), string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	deleted, _ := result["deleted"].(float64)
	logs.InfoContextf(w.ctx, "DeleteByQuery completed, deleted: %.0f documents", deleted)

	return nil
}

// buildBaseQuery 构建基础查询条件
func (w *EsSearchWrapper) buildBaseQuery() esquery.Map {
	w.query = esquery.Map{
		"must": []esquery.Map{},
	}

	if len(w.embedding) > 0 {
		w.query["should"] = []esquery.Map{
			esquery.BuildMap("script_score", esquery.BuildMap("query", esquery.BuildMap("match_all", esquery.Map{}), "script", esquery.BuildMap("source", "double score = cosineSimilarity(params.query_vector, 'embedding') ; return score ;", "params", esquery.BuildMap("query_vector", w.embedding)))),
		}
	}

	if len(w.forestIds) > 0 {
		w.query["must"] = append(w.query["must"].([]esquery.Map), esquery.BuildMap("terms", esquery.BuildMap("forest_id", w.forestIds)))
	}

	if len(w.fileIds) > 0 {
		w.query["must"] = append(w.query["must"].([]esquery.Map), esquery.BuildMap("terms", esquery.BuildMap("file_id", w.fileIds)))
	}

	if len(w.questionIDs) > 0 {
		w.query["must"] = append(w.query["must"].([]esquery.Map), esquery.BuildMap("terms", esquery.BuildMap("_id", w.questionIDs)))
	}

	w.addFilters()

	return w.query
}

// addFilters 添加过滤条件
func (w *EsSearchWrapper) addFilters() {
	if w.pageQuery == nil {
		return
	}
	for _, filter := range w.pageQuery.Filters {
		field := filter.Field
		switch field {
		case "company_id", "qa_lable", "qa_main_id", "qa_question", "qa_answer", "uin":
			if len(filter.Value) > 0 {
				w.query["must"] = append(w.query["must"].([]esquery.Map), esquery.BuildMap("term", esquery.BuildMap(field, filter.Value[0])))
			}
		default:
			logs.ErrorContextf(w.ctx, "EsSearchWrapper.addFilters invalid filtering field: %s", filter.Field)
		}
	}
}

// buildSearchQuery 构建完整搜索查询
func (w *EsSearchWrapper) buildSearchQuery() map[string]interface{} {
	rule := "asc"
	ss := strings.Split(w.pageQuery.OrderBy[0], " ")
	if len(ss) > 1 && ss[1] == "desc" {
		rule = "desc"
	}

	if w.pageQuery.ListAll {
		w.pageQuery.Limit = 10000
	}

	return map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": w.query,
		},
		"aggs": map[string]interface{}{
			"qa_answer_groups": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "qa_answer_id",
					"size":  w.getAggSize(),
					"order": map[string]interface{}{
						"max_updated_at": rule,
					},
				},
				"aggs": map[string]interface{}{
					"max_updated_at": map[string]interface{}{
						"max": map[string]interface{}{
							"field": "updated_at",
						},
					},
					"main_question": map[string]interface{}{
						"filter": map[string]interface{}{
							"bool": map[string]interface{}{
								"must_not": []map[string]interface{}{
									{
										"exists": map[string]interface{}{
											"field": "qa_main_id",
										},
									},
								},
							},
						},
						"aggs": map[string]interface{}{
							"main_info": map[string]interface{}{
								"top_hits": map[string]interface{}{
									"size": 1,
									"_source": map[string]interface{}{
										"excludes": []string{"embedding"},
									},
									"sort": []map[string]interface{}{
										{
											"created_at": map[string]interface{}{
												"order": "desc",
											},
										},
									},
								},
							},
						},
					},
					"child_questions": map[string]interface{}{
						"filter": map[string]interface{}{
							"exists": map[string]interface{}{
								"field": "qa_main_id",
							},
						},
						"aggs": map[string]interface{}{
							"children_info": map[string]interface{}{
								"top_hits": map[string]interface{}{
									"size":    20,
									"_source": []string{"qa_question", "created_at"},
									"sort": []map[string]interface{}{
										{
											"created_at": map[string]interface{}{
												"order": "desc",
											},
										},
									},
								},
							},
						},
					},
					"bucket_sort": map[string]interface{}{
						"bucket_sort": map[string]interface{}{
							"from": w.pageQuery.Offset,
							"size": w.pageQuery.Limit,
						},
					},
				},
			},
		},
	}
}

// getCount 获取查询结果总数
func (w *EsSearchWrapper) getCount(boolQuery esquery.Map) (int64, error) {
	countQuery := esquery.BuildMap("query", esquery.BuildMap("bool", boolQuery))
	queryBytes, err := json.Marshal(countQuery)
	if err != nil {
		return 0, err
	}

	resp, err := w.cli.Count(w.cli.Count.WithContext(w.ctx), w.cli.Count.WithIndex(w.indexName), w.cli.Count.WithBody(bytes.NewReader(queryBytes)))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("es count failed: %s, error: %s", resp.Status(), string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	count, ok := result["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid count response")
	}

	return int64(count), nil
}

// executeSearch 执行搜索查询
func (w *EsSearchWrapper) executeSearch(searchQuery esquery.Map) ([]*FQAItem, error) {
	queryBytes, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}
	resp, err := w.cli.Search(w.cli.Search.WithContext(w.ctx), w.cli.Search.WithIndex(w.indexName), w.cli.Search.WithBody(bytes.NewReader(queryBytes)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es search failed: %s, error: %s, search: %v", resp.Status(), string(body), string(queryBytes))
	}
	return w.parseSearchResponse(resp.Body)
}

// parseSearchResponse 解析搜索响应
func (w *EsSearchWrapper) parseSearchResponse(body io.Reader) ([]*FQAItem, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}
	aggs, ok := result["aggregations"].(map[string]interface{})
	if !ok {
		return []*FQAItem{}, nil
	}
	qaGroups, ok := aggs["qa_answer_groups"].(map[string]interface{})
	if !ok {
		return []*FQAItem{}, nil
	}
	buckets, ok := qaGroups["buckets"].([]interface{})
	if !ok {
		return []*FQAItem{}, nil
	}
	var fqaWithChildren []*FQAItem
	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})

		// 解析主问题
		mainQuestion := w.parseMainQuestion(bucketMap)
		if mainQuestion == nil {
			continue
		}
		// 解析子问题
		childQuestions := w.parseChildQuestions(bucketMap)
		fqaWithChildren = append(fqaWithChildren, &FQAItem{
			Main:  mainQuestion,
			Child: childQuestions,
		})
	}
	w.Total = uint(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	return fqaWithChildren, nil
}

// getAggSize 获取聚合大小
func (w *EsSearchWrapper) getAggSize() int {
	if w.pageQuery.ListAll {
		return 10000
	}
	return w.pageQuery.Limit + w.pageQuery.Offset
}

// parseMainQuestion 解析主问题
func (w *EsSearchWrapper) parseMainQuestion(bucketMap map[string]interface{}) *ragtypes.FQA {
	mainQuestionAgg, exists := bucketMap["main_question"].(map[string]interface{})
	if !exists {
		return nil
	}
	mainInfo, exists := mainQuestionAgg["main_info"].(map[string]interface{})
	if !exists {
		return nil
	}
	mainHits, exists := mainInfo["hits"].(map[string]interface{})["hits"].([]interface{})
	if !exists || len(mainHits) == 0 {
		return nil
	}
	mainHit := mainHits[0].(map[string]interface{})
	source := mainHit["_source"]
	sourceBytes, _ := json.Marshal(source)
	var fqa ragtypes.FQA
	json.Unmarshal(sourceBytes, &fqa)
	fqa.ID = mainHit["_id"].(string)
	return &fqa
}

// parseChildQuestions 解析子问题
func (w *EsSearchWrapper) parseChildQuestions(bucketMap map[string]interface{}) []*Child {
	var childQuestions []*Child
	childQuestionAgg, exists := bucketMap["child_questions"].(map[string]interface{})
	if !exists {
		return childQuestions
	}
	childInfo, exists := childQuestionAgg["children_info"].(map[string]interface{})
	if !exists {
		return childQuestions
	}
	childHits, exists := childInfo["hits"].(map[string]interface{})["hits"].([]interface{})
	if !exists {
		return childQuestions
	}
	for _, childHit := range childHits {
		childHitMap := childHit.(map[string]interface{})
		source := childHitMap["_source"].(map[string]interface{})
		child := &Child{
			ID:       childHitMap["_id"].(string),
			Question: source["qa_question"].(string),
		}
		// 解析created_at
		if createdAtStr, ok := source["created_at"].(string); ok {
			if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				child.CreatedAt = createdAt
			}
		}
		childQuestions = append(childQuestions, child)
	}
	return childQuestions
}

func (w *EsSearchWrapper) ModifyQAPair(targ *FQAItem) error {
	qas, err := w.SearchFQA()
	if err != nil {
		return fmt.Errorf("ModifyQAPair SearchFQA err: %v", err)
	}
	if len(qas.Data) != 1 {
		return fmt.Errorf("ModifyQAPair 查询到无效qa数据对: %v", qas)
	}
	var (
		orig    = qas.Data[0]
		mainFQA = make(map[string]interface{})
		now     = time.Now()
	)
	if orig.Main.QAQuestion != targ.Main.QAQuestion {
		eb, err := GetEmbedding(targ.Main.QAQuestion)
		if err != nil {
			return fmt.Errorf("ModifyQAPair GetEmbedding err: %v", err)
		}
		targ.Main.Embedding = eb
		mainFQA["qa_question"] = targ.Main.QAQuestion
		mainFQA["embedding"] = eb
		mainFQA["updated_at"] = now
	}
	if orig.Main.QAAnswer != targ.Main.QAAnswer {
		mainFQA["qa_answer"] = targ.Main.QAAnswer
	}
	// 获取 FQA 子项的差异
	toAdd, toUpdate, toDelIDs := w.diff(targ, orig)
	bulkBody := new(bytes.Buffer)
	encoder := json.NewEncoder(bulkBody)
	var isUpdate bool
	// 更新主文档
	if len(mainFQA) > 0 {
		if err := encoder.Encode(map[string]interface{}{
			"update": map[string]interface{}{
				"_index": w.indexName,
				"_id":    orig.Main.ID,
			},
		}); err != nil {
			return err
		}
		if err = encoder.Encode(map[string]interface{}{"doc": mainFQA}); err != nil {
			return err
		}
		isUpdate = true
	}
	// 删除子项
	for _, id := range toDelIDs {
		if err := encoder.Encode(map[string]interface{}{
			"delete": map[string]interface{}{
				"_index": w.indexName,
				"_id":    id,
			},
		}); err != nil {
			return err
		}
		isUpdate = true
	}
	// 更新子项
	for _, child := range toUpdate {
		embedding, err := GetEmbedding(child.Question)
		if err != nil {
			return fmt.Errorf("ModifyQAPair GetEmbedding err: %v", err)
		}
		updateDoc := map[string]interface{}{
			"qa_question": child.Question,
			"updated_at":  now,
			"embedding":   embedding,
		}
		if err := encoder.Encode(map[string]interface{}{
			"update": map[string]interface{}{
				"_index": w.indexName,
				"_id":    child.ID,
			},
		}); err != nil {
			return err
		}
		if err := encoder.Encode(map[string]interface{}{"doc": updateDoc}); err != nil {
			return err
		}
		isUpdate = true
	}
	// 新增子项
	for _, q := range toAdd {
		if err := encoder.Encode(map[string]interface{}{
			"index": map[string]interface{}{
				"_index": w.indexName,
				"_id":    uuid.New().String(),
			},
		}); err != nil {
			return err
		}
		embedding, err := GetEmbedding(q)
		if err != nil {
			return fmt.Errorf("ModifyQAPair GetEmbedding err: %v", err)
		}
		if err := encoder.Encode(ragtypes.FQA{
			Common: ragtypes.Common{
				CreatedAt: now,
				UpdatedAt: now,
				ForestID:  orig.Main.ForestID,
				CompanyID: orig.Main.CompanyID,
				Uin:       orig.Main.Uin,
				Type:      orig.Main.Type,
			},
			QAQuestion: q,
			QALable:    orig.Main.QALable,
			QAAnswer:   orig.Main.QAAnswer,
			QAMainID:   orig.Main.ID,
			QAAnswerID: orig.Main.QAAnswerID,
			Embedding:  embedding,
		}); err != nil {
			return err
		}
		isUpdate = true
	}
	if !isUpdate {
		return nil
	}
	resp, err := w.cli.Bulk(
		bulkBody,
		w.cli.Bulk.WithContext(w.ctx),
		w.cli.Bulk.WithRefresh("true"))
	if err != nil {
		return fmt.Errorf("ES bulk request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ES bulk error: %s search: %v", string(body), bulkBody.String())
	}
	return nil
}

func (w *EsSearchWrapper) diffChild(targ, orig *FQAItem) (toAdd, toDelIDs []string) {
	var (
		origMap = make(map[string]*Child, len(orig.Child))
		targMap = make(map[string]*Child, len(targ.Child))
	)

	for _, v := range targ.Child {
		targMap[v.ID] = v
	}

	for _, c := range orig.Child {
		origMap[c.Question] = c
		if q, exist := targMap[c.ID]; q.IsDeleted ||
			(exist && q.Question != c.Question) {
			toDelIDs = append(toDelIDs, c.ID)
		}
	}

	for _, c := range targ.Child {
		if _, exists := origMap[c.Question]; !exists && !c.IsDeleted {
			toAdd = append(toAdd, c.Question)
		}
	}

	return
}

func (w *EsSearchWrapper) diff(targ, orig *FQAItem) (toAdd []string, toUpdate []*Child, toDelIDs []string) {
	var (
		origMap = make(map[string]*Child, len(orig.Child))
		targMap = make(map[string]*Child, len(targ.Child))
	)
	// 建立基于 ID 的映射
	for _, v := range orig.Child {
		origMap[v.ID] = v
	}
	for _, v := range targ.Child {
		targMap[v.ID] = v
	}
	// 处理原有的子项
	for _, origChild := range orig.Child {
		if targChild, exists := targMap[origChild.ID]; exists {
			// 如果目标中标记为删除，则删除
			if targChild.IsDeleted {
				toDelIDs = append(toDelIDs, origChild.ID)
			} else if targChild.Question != origChild.Question {
				// 如果问题内容有变化，则更新
				toUpdate = append(toUpdate, targChild)
			}
		} else {
			// 如果目标中不存在该 ID，表示需要删除
			toDelIDs = append(toDelIDs, origChild.ID)
		}
	}
	for _, targChild := range targ.Child {
		if _, exists := origMap[targChild.ID]; !exists && !targChild.IsDeleted {
			// 如果原有数据中不存在该 ID，且未标记删除，则新增
			toAdd = append(toAdd, targChild.Question)
		}
	}
	return
}
