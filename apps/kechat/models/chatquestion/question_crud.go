package chatquestion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/dbtools/estool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

var question_es_client *elasticsearch.Client

// InitHistoryESClient 初始化ES 历史记录查询client
func InitHistoryESClient(ctx context.Context) error {
	// cfg := config.ESConfig{
	// 	Addresses:     []string{"http://example.com:53082/"},
	// 	SlowThreshold: time.Millisecond,
	// 	Username:      "elastic",
	// 	Password:      "CHANGE_ME_PASSWORD",
	// }
	cfg := config.ESConfig{}
	err := settings.GetYaml("knowledge", "es", &cfg)
	if err != nil {
		logs.ErrorContext(ctx, "get es config failed: %s", err)
		return err
	}
	client, err := estool.InitES(cfg)
	if err != nil {
		logs.ErrorContext(ctx, "init es client failed: %s", err)
		return err
	}
	question_es_client = client
	return nil
}

// CreateQuestion 创建问题
func CreateQuestion(ctx context.Context, question *chattype.ChatQuestion) error {
	question.Source.CreatedAt = time.Now()
	resp, err := question_es_client.Index(
		chattype.HistoryIndex,
		strings.NewReader(question.Source.String()),
		question_es_client.Index.WithContext(ctx),
		question_es_client.Index.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateQuestion init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "CreateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("CreateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateQuestion error reading body: %v", err)
		return err
	}
	// 解析JSON响应
	var searchRes createRespnse
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal createRespnse error: %v", err)
		return err
	}
	question.ID = searchRes.ID
	return nil
}

// CreateQuestion 创建问题
func CreateQuestionWithID(ctx context.Context, question *chattype.ChatQuestion) error {
	question.Source.CreatedAt = time.Now()
	resp, err := question_es_client.Index(
		chattype.HistoryIndex,
		strings.NewReader(question.Source.String()),
		question_es_client.Index.WithDocumentID(question.ID),
		question_es_client.Index.WithContext(ctx),
		question_es_client.Index.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateQuestion init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "CreateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("CreateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
	}

	return nil
}

// DeleteQuestion 删除问题
func DeleteQuestion(ctx context.Context, id string) error {
	resp, err := question_es_client.Delete(
		chattype.HistoryIndex,
		id,
		question_es_client.Delete.WithContext(ctx),
		question_es_client.Delete.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteQuestion init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "DeleteQuestion es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("DeleteQuestion es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// UpdateQuestion 更新问题
func UpdateQuestion(ctx context.Context, question *chattype.ChatQuestion) error {
	updateBody := map[string]interface{}{
		"doc": question.Source,
	}
	// 将数据转换为 JSON
	body, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal update body: %w", err)
	}
	resp, err := question_es_client.Update(
		chattype.HistoryIndex,
		question.ID,
		strings.NewReader(string(body)),
		question_es_client.Update.WithContext(ctx),
		question_es_client.Update.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateQuestion init history es client failed: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "UpdateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("UpdateQuestion es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// ListSessionQuestions 获取会话问题
func ListSessionQuestionsByUin(ctx context.Context, uin, session_id uint) ([]*chattype.ChatQuestion, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("session_id", session_id)),
		esquery.BuildMap("term", esquery.BuildMap("uin", uin)),
	}
	sort := []esquery.Map{
		esquery.BuildMap("created_at", esquery.BuildMap("order", "asc")),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSort(sort).
		SetSize(1000)
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := question_es_client.Search(
		question_es_client.Search.WithIndex(chattype.HistoryIndex),
		question_es_client.Search.WithBody(bytes.NewBuffer(querybyte)),
		question_es_client.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
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
	return searchRes.Hits.Hits, nil
}

// GetQuetionByID 根据id获取问题
func GetQuetionByID(ctx context.Context, id string) (*chattype.ChatQuestion, error) {
	resp, err := question_es_client.Get(
		chattype.HistoryIndex,
		id,
		question_es_client.Get.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var question chattype.ChatQuestion
	if err := json.Unmarshal(body, &question); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return nil, err
	}
	return &question, nil
}

// DeleteSessionQuestions 删除会话问题
func DeleteSessionQuestions(ctx context.Context, uin, session_id uint) error {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("session_id", session_id)),
		esquery.BuildMap("term", esquery.BuildMap("uin", uin)),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := question_es_client.DeleteByQuery(
		[]string{chattype.HistoryIndex},
		bytes.NewBuffer(querybyte),
		question_es_client.DeleteByQuery.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "DeleteSessionQuestions es query failed: %s, error: %s", resp.Status(), string(body))
		return fmt.Errorf("DeleteSessionQuestions es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// ListExternalSessionQuestions 获取会话问题
func ListSessionQuestions(ctx context.Context, session_id uint) ([]*chattype.ChatQuestion, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("session_id", session_id)),
	}
	sort := []esquery.Map{
		esquery.BuildMap("created_at", esquery.BuildMap("order", "asc")),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSort(sort).
		SetSize(1000)
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := question_es_client.Search(
		question_es_client.Search.WithIndex(chattype.HistoryIndex),
		question_es_client.Search.WithBody(bytes.NewBuffer(querybyte)),
		question_es_client.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
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
	return searchRes.Hits.Hits, nil
}

// GetUnscopedQAByCompanyID 获取当日问答个数
func GetUnscopedQAByCompanyID(ctx context.Context, companyID uint) (int64, error) {
	mustQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("company_id", companyID)),
		esquery.BuildMap("range", esquery.BuildMap("created_at", esquery.BuildMap("gte", "now/d", "lt", "now/d+1d"))),
	}
	// sort := []esquery.Map{
	// 	esquery.BuildMap("created_at", esquery.BuildMap("order", "asc")),
	// }
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		// SetSort(sort).
		SetSize(0)
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return 0, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := question_es_client.Search(
		question_es_client.Search.WithIndex(chattype.HistoryIndex),
		question_es_client.Search.WithBody(bytes.NewBuffer(querybyte)),
		question_es_client.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return 0, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return 0, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return 0, err
	}
	// 解析JSON响应
	var searchRes QuestionSearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "unmarshal ChatResponse error: %v", err)
		return 0, err
	}
	return searchRes.Hits.Total.Value, nil
}

// ListUinSevenDayQuestions 获取会话问题
func ListUinSevenDayQuestions(ctx context.Context, uin uint) ([]*chattype.ChatQuestion, error) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	mustQuery := []esquery.Map{
		// esquery.BuildMap("term", esquery.BuildMap("session_id", session_id)),
		esquery.BuildMap("term", esquery.BuildMap("uin", uin)),
		// 添加日期范围查询，获取七天内的数据
		esquery.BuildMap("range", esquery.BuildMap("created_at",
			esquery.BuildMap("gte", sevenDaysAgo,
				"lte", now))),
		esquery.BuildMap("term", esquery.BuildMap("base_agent_id", 0)),
	}
	sort := []esquery.Map{
		esquery.BuildMap("created_at", esquery.BuildMap("order", "desc")),
	}
	boolQuery := esquery.BuildMap("must", mustQuery)
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", boolQuery)).
		SetSort(sort).
		SetSize(1000)
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "querybyte: %v", string(querybyte))
	resp, err := question_es_client.Search(
		question_es_client.Search.WithIndex(chattype.HistoryIndex),
		question_es_client.Search.WithBody(bytes.NewBuffer(querybyte)),
		question_es_client.Search.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return nil, err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "es query failed: %s, error: %s", resp.Status(), string(body))
		return nil, fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}
	// 解析JSON响应
	var searchRes QuestionSearchResult
	if err := json.Unmarshal(body, &searchRes); err != nil {
		logs.ErrorContextf(ctx, "es Unmarshal failed: %s, error: %v", string(body), err)
		return nil, err
	}
	return searchRes.Hits.Hits, nil
}
