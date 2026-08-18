package membership

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

// QuotaManager 配额管理器接口
type QuotaManager interface {
	Grant(ctx context.Context, grantReq *QuotaGrantReq) error
	Query(ctx context.Context, queryReq *QuotaQueryReq) (*QuotaQueryRes, error)
	Check(ctx context.Context, checkReq *QuotaCheckReq) (bool, int64, error)
}

type quotaManager struct{}

func NewQuotaManager() QuotaManager {
	return &quotaManager{}
}

func (m *quotaManager) Grant(ctx context.Context, grantReq *QuotaGrantReq) error {
	return nil
}

func (m *quotaManager) Query(ctx context.Context, queryReq *QuotaQueryReq) (*QuotaQueryRes, error) {
	if queryReq.CompanyID == 0 {
		return nil, fmt.Errorf("[QuotaManager.Query] companyID is required")
	}
	res := &QuotaQueryRes{}

	usageManager := NewUsageManager([]QuotaResourceType{
		QuotaResourceTypeAgent,
		QuotaResourceTypeQA,
		QuotaResourceTypeDisk,
		QuotaResourceTypeEmployee,
		QuotaResourceTypeArticle,
	})
	usageReq := &UsageQueryReq{
		CompanyID: queryReq.CompanyID,
	}

	usageRes, err := usageManager.QueryUsage(ctx, usageReq)
	if err != nil {
		return nil, err
	}

	// 根据资源类型填充对应的用量字段
	for _, stat := range usageRes.ResourceStats {
		switch stat.ResourceType {
		case QuotaResourceTypeAgent:
			res.AgentQuotaUsed = stat.Usage
		case QuotaResourceTypeQA:
			res.QaQuotaUsed = stat.Usage
		case QuotaResourceTypeDisk:
			res.DiskQuotaUsed = stat.Usage
		case QuotaResourceTypeEmployee:
			res.EmployeeQuotaUsed = stat.Usage
		case QuotaResourceTypeArticle:
			res.ArticleQuotaUsed = stat.Usage
		}
	}
	// 获取原始配额信息
	rawQuota, err := m.getRawQuota(ctx, queryReq.CompanyID)
	if err != nil {
		return nil, err
	}

	res.AgentQuota = rawQuota.AgentQuota
	res.QaQuota = rawQuota.QaQuota
	res.DiskQuota = rawQuota.DiskQuota
	res.EmployeeQuota = rawQuota.EmployeeQuota
	res.ArticleQuota = rawQuota.ArticleQuota

	return res, nil
}

func (m *quotaManager) Check(ctx context.Context, checkReq *QuotaCheckReq) (bool, int64, error) {
	// 调用UsageManager获取指定资源类型的用量
	usageManager := NewUsageManager([]QuotaResourceType{checkReq.ResourceType})
	usageReq := &UsageQueryReq{
		CompanyID: checkReq.CompanyID,
	}

	usageRes, err := usageManager.QueryUsage(ctx, usageReq)
	if err != nil {
		return false, 0, err
	}

	// 获取该资源类型的用量
	var currentUsage int64
	var resourceQuota int64
	for _, stat := range usageRes.ResourceStats {
		if stat.ResourceType == checkReq.ResourceType {
			currentUsage = stat.Usage
			break
		}
	}

	// 获取原始配额信息
	rawQuota, err := m.getRawQuota(ctx, checkReq.CompanyID)
	if err != nil {
		return false, 0, err
	}
	switch checkReq.ResourceType {
	case QuotaResourceTypeAgent:
		resourceQuota = rawQuota.AgentQuota
	case QuotaResourceTypeQA:
		resourceQuota = rawQuota.QaQuota
	case QuotaResourceTypeDisk:
		resourceQuota = rawQuota.DiskQuota
	case QuotaResourceTypeEmployee:
		resourceQuota = rawQuota.EmployeeQuota
	case QuotaResourceTypeArticle:
		resourceQuota = rawQuota.ArticleQuota
	}

	balance := resourceQuota - currentUsage

	return balance > 0, balance, nil
}

// getRawQuota 获取原始配额信息
// 从数据库查询指定公司的所有有效配额记录，并聚合各资源类型的配额
func (m *quotaManager) getRawQuota(ctx context.Context, companyID uint) (*RawQuota, error) {
	if companyID == 0 {
		return nil, fmt.Errorf("[QuotaManager.getRawQuota] companyID is required")
	}

	systemPackageEntity, err := forest.NewKePackageDao().GetByCond(ctx, &forest.KePackageCond{
		SourceType: foresttype.PackageSourceTypeSystem,
		Status:     foresttype.PackageStatusOnline,
	})
	if err != nil {
		return nil, fmt.Errorf("[QuotaManager.getRawQuota] get system package fail, companyID:%d, err: %v", companyID, err)
	}

	// 获取有效期内的套餐配额
	now := time.Now()
	quotaList, err := forest.NewKeCompanyQuotaDao().GetListByCond(ctx, &forest.KeCompanyQuotaCond{
		CompanyID:     companyID,
		ExpireAtStart: &now,
	})
	if err != nil {
		return nil, fmt.Errorf("[QuotaManager.getRawQuota] get company quota list fail, companyID:%d, err: %v", companyID, err)
	}

	agentQuota := systemPackageEntity.AgentQuota
	qaQuota := systemPackageEntity.QaQuota
	diskQuota := systemPackageEntity.DiskQuota
	employeeQuota := systemPackageEntity.EmployeeQuota
	articleQuota := systemPackageEntity.ArticleQuota

	// 聚合各资源类型的配额
	for _, v := range quotaList {
		agentQuota = max(agentQuota, v.AgentQuota)
		qaQuota = max(qaQuota, v.QaQuota)
		diskQuota = max(diskQuota, v.DiskQuota)
		employeeQuota = max(employeeQuota, v.EmployeeQuota)
		articleQuota = max(articleQuota, v.ArticleQuota)
	}

	return &RawQuota{
		AgentQuota:    agentQuota,
		QaQuota:       qaQuota,
		DiskQuota:     diskQuota,
		EmployeeQuota: employeeQuota,
		ArticleQuota:  articleQuota,
	}, nil
}
