package membership

import (
	"context"
	"sync"

	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
)

// UsageManager 用量统计管理器接口
// 作为对外统一入口，管理多个资源类型的用量统计
type UsageManager interface {
	QueryUsage(ctx context.Context, req *UsageQueryReq) (*UsageQueryRes, error)
	GetProvider(resourceType QuotaResourceType) (UsageProvider, bool)
}

type UsageProvider interface {
	GetResourceType() QuotaResourceType
	CalculateUsage(ctx context.Context, req *UsageQueryReq) (*ResourceUsageStatItem, error)
}

// usageManager 用量统计管理器实现
type usageManager struct {
	providers map[QuotaResourceType]UsageProvider
	mu        sync.RWMutex
}

// NewUsageManager 创建用量统计管理器实例
// resourceTypes: 支持的资源类型列表
func NewUsageManager(resourceTypes []QuotaResourceType) UsageManager {
	supportedTypes := make(map[QuotaResourceType]bool)
	for _, rt := range resourceTypes {
		supportedTypes[rt] = true
	}

	um := &usageManager{
		providers: make(map[QuotaResourceType]UsageProvider),
	}

	// 为每个资源类型创建并注册Provider
	for _, rt := range resourceTypes {
		provider := usageProviderFactory(rt)
		if provider != nil {
			um.providers[rt] = provider
		}
	}

	return um
}

// usageProviderFactory Provider工厂函数
func usageProviderFactory(resourceType QuotaResourceType) UsageProvider {
	switch resourceType {
	case QuotaResourceTypeAgent:
		return newAgentUsageProvider()
	case QuotaResourceTypeQA:
		return newQAUsageProvider()
	case QuotaResourceTypeDisk:
		return newDiskUsageProvider()
	case QuotaResourceTypeEmployee:
		return newEmployeeUsageProvider()
	case QuotaResourceTypeArticle:
		return newArticleUsageProvider()
	default:
		return nil
	}
}

// QueryUsage 查询用量统计
func (m *usageManager) QueryUsage(ctx context.Context, req *UsageQueryReq) (*UsageQueryRes, error) {
	m.mu.RLock()
	providers := make(map[QuotaResourceType]UsageProvider)
	for rt, p := range m.providers {
		providers[rt] = p
	}
	m.mu.RUnlock()

	res := &UsageQueryRes{
		CompanyID:     req.CompanyID,
		ResourceStats: make([]ResourceUsageStatItem, 0),
	}

	var statsMu sync.Mutex
	g, gCtx := errgroup.WithContext(ctx)

	for resourceType, provider := range providers {
		rType := resourceType // 避免闭包问题
		provider := provider

		g.Go(func() error {
			statItem, err := provider.CalculateUsage(gCtx, req)
			statsMu.Lock()
			defer statsMu.Unlock()

			if err != nil {
				// 计算失败，记录错误但继续处理其他资源类型
				res.ResourceStats = append(res.ResourceStats, ResourceUsageStatItem{
					ResourceType: rType,
					Usage:        0,
				})
				logs.ErrorContextf(ctx, "[usageManager.QueryUsage] calculate usage fail, resourceType:%s, err: %v", rType, err)
				return err
			}
			// 确保ResourceType一致
			statItem.ResourceType = rType
			res.ResourceStats = append(res.ResourceStats, *statItem)
			return nil
		})
	}

	// 等待所有goroutine完成
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

// GetProvider 获取指定资源类型的统计器
func (m *usageManager) GetProvider(resourceType QuotaResourceType) (UsageProvider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[resourceType]
	return provider, exists
}
