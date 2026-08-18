package membership

import (
	"context"

	"github.com/insmtx/corekg/apps/kechat/models/chat"
)

// agentUsageProvider 智能体用量统计Provider
type agentUsageProvider struct{}

// newAgentUsageProvider 创建智能体用量统计Provider
func newAgentUsageProvider() UsageProvider {
	return &agentUsageProvider{}
}

// GetResourceType 返回资源类型
func (p *agentUsageProvider) GetResourceType() QuotaResourceType {
	return QuotaResourceTypeAgent
}

// CalculateUsage 计算智能体用量
func (p *agentUsageProvider) CalculateUsage(ctx context.Context, req *UsageQueryReq) (*ResourceUsageStatItem, error) {
	total, err := chat.NewChatAgentDao().CountByCond(ctx, &chat.ChatAgentCond{
		CompanyID: req.CompanyID,
	})
	if err != nil {
		return nil, err
	}
	return &ResourceUsageStatItem{
		ResourceType: QuotaResourceTypeAgent,
		Usage:        total,
	}, nil
}
