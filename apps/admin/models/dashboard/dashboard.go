package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

type DataSet struct {
	Total     int `json:"total"`
	Increased int `json:"increased"`
}

// GetForestData
// 知识库总数（包含所有类型的）
// 知识库增量（和昨天相比增加了多少）
func GetForestData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}
	query := dbutil.Knownow().WithContext(ctx).
		Model(&foresttype.KnownowForest{}).
		Select("COUNT(*) AS total, SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased").
		Where("deleted_at IS NULL")

	// 如果 companyID 大于 0，则添加筛选条件
	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(&ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetForestData query err: %v", err)
		return nil, err
	}
	return ds, nil
}

// GetDocData
// 上传文档总数
// 上传文档增量
func GetDocData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}
	query := dbutil.Knownow().WithContext(ctx).
		Model(&foresttype.KnownowForestFile{}).
		Select("COUNT(*) AS total, SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased").
		Where("deleted_at IS NULL")

	// 如果 companyID 大于 0，则添加筛选条件
	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(&ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetDocData query err: %v", err)
		return nil, err
	}
	return ds, nil
}

// GetParseData
// 成功解析文档总数
// 成功解析文档增量
func GetParseData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}
	query := dbutil.Knownow().WithContext(ctx).
		Model(&foresttype.KnownowForestFile{}).
		Select("COUNT(*) AS total, SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased").
		Where("deleted_at IS NULL").
		Where("parse_status = ? AND knowledge_status = ? AND graph_status = ? AND desc_status = ?", "success", "success", "success", "success")

	// 如果 companyID 大于 0，则添加筛选条件
	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(&ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetParseData query err: %v", err)
		return nil, err
	}
	return ds, nil
}

// GetSessionData
// 会话量（包含所有会话历史记录）
// 会话量增量
func GetSessionData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}
	query := dbutil.Chat().WithContext(ctx).
		Model(&chattype.ChatSession{}).
		Select("COUNT(*) AS total, SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased").
		Where("deleted_at IS NULL")

	// 如果 companyID 大于 0，则添加筛选条件
	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(&ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetSessionData query err: %v", err)
		return nil, err
	}
	return ds, nil
}

// GetQAData
// 获取问答数量
func GetQAData(ctx context.Context, companyID uint) (*DataSet, error) {
	cli, err := essearch.InitESClient(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "InitESClient failed err: %v", err)
		return nil, err
	}

	reqBody := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"total_unique_requests": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": "req_id",
				},
			},
		},
	}

	if companyID > 0 {
		reqBody["query"] = map[string]interface{}{
			"term": map[string]interface{}{
				"company_id": companyID,
			},
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("encode query failed: %v", err)
	}

	res, err := cli.Search(
		cli.Search.WithContext(ctx),
		cli.Search.WithIndex(chattype.HistoryIndex),
		cli.Search.WithBody(&buf),
		cli.Search.WithTrackTotalHits(false))
	if err != nil {
		logs.ErrorContextf(ctx, "ES search request failed: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		logs.ErrorContextf(ctx, "ES search response error: %s", string(bodyBytes))
		return nil, fmt.Errorf("es search returned error status: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logs.ErrorContextf(ctx, "Decode response failed: %v", err)
		return nil, err
	}
	var qaCount int64 = 0

	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		if bucket, ok := aggs["total_unique_requests"].(map[string]interface{}); ok {
			if val, ok := bucket["value"].(float64); ok {
				qaCount = int64(val)
			}
		}
	}

	return &DataSet{
		Total: int(qaCount),
	}, nil
}

// GetUserData
// 获取用户数量
func GetUserData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}
	selectSQL := `
    COUNT(DISTINCT user_id) AS total,
    COUNT(DISTINCT CASE WHEN DATE(created_at) = CURDATE() THEN user_id END) AS increased
`

	query := dbutil.Account().WithContext(ctx).
		Model(&accounttype.UserIdentification{}).
		Where("subject_type = ?", accounttype.SubjectTypeCompany).
		Select(selectSQL).
		Where("deleted_at IS NULL")

	if companyID > 0 {
		query = query.Where("subject_id = ?", companyID)
	}

	if err := query.Scan(ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetUserData query err: %v", err)
		return nil, err
	}
	return ds, nil
}

// GetGraphData
// 获取图谱数量
func GetGraphData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}

	selectSQL := `
        COUNT(*) AS total, 
        SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased
    `

	query := dbutil.Knownow().WithContext(ctx).
		Model(&foresttype.ForestGraph{}).
		Where("deleted_at IS NULL").
		Select(selectSQL)

	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetGraphData query err: %v", err)
		return nil, err
	}

	return ds, nil
}

// GetAgentData
// 获取轻应用数量
func GetAgentData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}

	selectSQL := `
        COUNT(*) AS total, 
        SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased
    `

	query := dbutil.Chat().WithContext(ctx).
		Model(&chattype.ChatAgent{}).
		Where("deleted_at IS NULL").
		Select(selectSQL)

	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetAgentData query err: %v", err)
		return nil, err
	}

	return ds, nil
}

// GetArticleData
// 获取写作空间数量
func GetArticleData(ctx context.Context, companyID uint) (*DataSet, error) {
	ds := &DataSet{}

	selectSQL := `
        COUNT(*) AS total, 
        SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END) AS increased
    `

	query := dbutil.Knownow().WithContext(ctx).
		Model(&foresttype.KeArticle{}).
		Where("deleted_at IS NULL").
		Select(selectSQL)

	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Scan(ds).Error; err != nil {
		logs.ErrorContextf(ctx, "GetArticleData query err: %v", err)
		return nil, err
	}

	return ds, nil
}
