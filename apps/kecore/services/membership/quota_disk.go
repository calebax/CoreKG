package membership

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
)

// diskUsageProvider 磁盘用量统计Provider
type diskUsageProvider struct{}

// newDiskUsageProvider 创建磁盘用量统计Provider
func newDiskUsageProvider() UsageProvider {
	return &diskUsageProvider{}
}

// GetResourceType 返回资源类型
func (p *diskUsageProvider) GetResourceType() QuotaResourceType {
	return QuotaResourceTypeDisk
}

// CalculateUsage 计算磁盘用量
func (p *diskUsageProvider) CalculateUsage(ctx context.Context, req *UsageQueryReq) (*ResourceUsageStatItem, error) {
	total, err := forest.NewForestFileDao().StatSizeByCond(ctx, &forest.ForestFileCond{
		CompanyID: req.CompanyID,
	})
	if err != nil {
		return nil, err
	}

	return &ResourceUsageStatItem{
		ResourceType: QuotaResourceTypeDisk,
		Usage:        total,
	}, nil
}
