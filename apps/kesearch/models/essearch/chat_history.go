package essearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

type StatChatHistoryReq struct {
	CompanyID      uint
	CreatedAtStart time.Time
	CreatedAtEnd   time.Time
}

// StatChatHistory 计算问答用量（统计指定时间范围内的唯一请求数量）
func StatChatHistory(ctx context.Context, req *StatChatHistoryReq) (int64, error) {
	esCli, err := InitESClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("[StatChatHistory] init es client fail, err: %v", err)
	}

	// 构建查询条件列表
	mustQuery := []esquery.Map{}

	// 添加 question 字段存在的条件
	mustQuery = append(mustQuery, esquery.BuildMap("exists", esquery.BuildMap("field", "question")))

	// 添加 question 字段值不为空字符串的条件
	mustNotQuery := []esquery.Map{
		esquery.BuildMap("term", esquery.BuildMap("question.keyword", "")),
	}

	// 只有当 CompanyID 有值时才添加 company_id 查询条件
	if req.CompanyID > 0 {
		mustQuery = append(mustQuery, esquery.BuildMap("term", esquery.BuildMap("company_id", req.CompanyID)))
	}

	// 构建时间范围查询：只有当至少有一个时间有效时才添加
	hasStartTime := !req.CreatedAtStart.IsZero()
	hasEndTime := !req.CreatedAtEnd.IsZero()

	if hasStartTime || hasEndTime {
		rangeMap := make(esquery.Map)
		if hasStartTime {
			rangeMap["gte"] = req.CreatedAtStart
		}
		if hasEndTime {
			rangeMap["lt"] = req.CreatedAtEnd
		}
		// 只有当 rangeMap 不为空时才添加时间范围查询
		if len(rangeMap) > 0 {
			mustQuery = append(mustQuery, esquery.BuildMap("range", esquery.BuildMap("created_at", rangeMap)))
		}
	}

	boolQuery := esquery.BuildMap("bool", esquery.BuildMap(
		"must", mustQuery,
		"must_not", mustNotQuery,
	))

	// 构建查询体
	builder := esquery.NewBuilder().
		SetSize(0).
		SetQuery(boolQuery)

	queryBytes, err := builder.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "[StatChatHistory] build query fail: %v", err)
		return 0, fmt.Errorf("[StatChatHistory] build query fail: %v", err)
	}

	// 执行 ES 查询
	res, err := esCli.Search(
		esCli.Search.WithContext(ctx),
		esCli.Search.WithIndex(chattype.HistoryIndex),
		esCli.Search.WithBody(bytes.NewBuffer(queryBytes)),
		esCli.Search.WithTrackTotalHits(true))
	if err != nil {
		logs.ErrorContextf(ctx, "[StatChatHistory] ES search request failed: %v", err)
		return 0, fmt.Errorf("[StatChatHistory] ES search request failed: %v", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		logs.ErrorContextf(ctx, "[StatChatHistory] ES search response error: status=%s, body=%s", res.Status(), string(bodyBytes))
		return 0, fmt.Errorf("[StatChatHistory] ES search returned error status: %s", res.Status())
	}

	// 解析响应结果
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logs.ErrorContextf(ctx, "[StatChatHistory] decode response failed: %v", err)
		return 0, fmt.Errorf("[StatChatHistory] decode response failed: %v", err)
	}

	// 提取文档总数
	qaCount := int64(0)
	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if val, ok := total["value"].(float64); ok {
				qaCount = int64(val)
			}
		}
	}

	return qaCount, nil
}
