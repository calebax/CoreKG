package membership

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/account"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
)

// employeeUsageProvider 成员用量统计Provider
type employeeUsageProvider struct{}

// newEmployeeUsageProvider 创建成员用量统计Provider
func newEmployeeUsageProvider() UsageProvider {
	return &employeeUsageProvider{}
}

// GetResourceType 返回资源类型
func (p *employeeUsageProvider) GetResourceType() QuotaResourceType {
	return QuotaResourceTypeEmployee
}

// CalculateUsage 计算成员数量
func (p *employeeUsageProvider) CalculateUsage(ctx context.Context, req *UsageQueryReq) (*ResourceUsageStatItem, error) {
	total, err := account.NewUserIdentificationDao().CountByCond(ctx, &account.UserIdentificationCond{
		SubjectType: accounttype.SubjectTypeCompany,
		SubjectID:   req.CompanyID,
	})
	if err != nil {
		return nil, err
	}
	return &ResourceUsageStatItem{
		ResourceType: QuotaResourceTypeEmployee,
		Usage:        total,
	}, nil
}
