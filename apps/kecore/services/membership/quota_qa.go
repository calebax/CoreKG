package membership

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
)

// qaUsageProvider 问答用量统计Provider
type qaUsageProvider struct{}

// newQAUsageProvider 创建问答用量统计Provider
func newQAUsageProvider() UsageProvider {
	return &qaUsageProvider{}
}

// GetResourceType 返回资源类型
func (p *qaUsageProvider) GetResourceType() QuotaResourceType {
	return QuotaResourceTypeQA
}

// CalculateUsage 计算问答用量（统计今天的唯一请求数量）
func (p *qaUsageProvider) CalculateUsage(ctx context.Context, req *UsageQueryReq) (*ResourceUsageStatItem, error) {
	// 计算当天的开始时间（00:00:00）
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 计算当天的结束时间（第二天的 00:00:00，使用 lt 查询）
	endOfDay := startOfDay.AddDate(0, 0, 1)

	total, err := essearch.StatChatHistory(ctx, &essearch.StatChatHistoryReq{
		CompanyID:      req.CompanyID,
		CreatedAtStart: startOfDay,
		CreatedAtEnd:   endOfDay,
	})
	if err != nil {
		return nil, err
	}
	return &ResourceUsageStatItem{
		ResourceType: QuotaResourceTypeQA,
		Usage:        total,
	}, nil
}
