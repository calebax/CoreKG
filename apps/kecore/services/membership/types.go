package membership

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

type QuotaResourceType string

const (
	QuotaResourceTypeAgent    QuotaResourceType = "agent"    // 智能体
	QuotaResourceTypeQA       QuotaResourceType = "qa"       // 问答
	QuotaResourceTypeDisk     QuotaResourceType = "disk"     // 磁盘
	QuotaResourceTypeEmployee QuotaResourceType = "employee" // 成员
	QuotaResourceTypeArticle  QuotaResourceType = "article"  // 文章
)

// QuotaGrantReq 配额发放请求
type QuotaGrantReq struct {
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
	// SourceType 来源类型
	SourceType foresttype.CompanyQuotaSourceType `json:"source_type"`
	// SourceID 来源ID（订单ID或发放记录ID）
	SourceID uint `json:"source_id"`
	// OperatorID 操作人ID（手动发放时记录操作人）
	OperatorID uint `json:"operator_id"`
}

// QuotaGrantRes 配额发放结果
type QuotaGrantRes struct {
	// GrantID 发放记录ID
	GrantID uint `json:"grant_id"`
}

type QuotaQueryReq struct {
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
}

type RawQuota struct {
	// AgentQuota 可创建智能体数量
	AgentQuota int64 `json:"agent_quota"`
	// QaQuota 每日问答次数
	QaQuota int64 `json:"qa_quota"`
	// DiskQuota 磁盘配额，单位字节
	DiskQuota int64 `json:"disk_quota"`
	// EmployeeQuota 成员数量
	EmployeeQuota int64 `json:"employee_quota"`
	// ArticleQuota 可创建文章数量
	ArticleQuota int64 `json:"article_quota"`
}
type QuotaQueryRes struct {
	RawQuota
	// AgentQuotaUsed 已使用智能体数量
	AgentQuotaUsed int64 `json:"agent_quota_used"`
	// QaQuotaUsed 已使用问答次数
	QaQuotaUsed int64 `json:"qa_quota_used"`
	// DiskQuotaUsed 已使用磁盘空间，单位字节
	DiskQuotaUsed int64 `json:"disk_quota_used"`
	// EmployeeQuotaUsed 已使用成员数量
	EmployeeQuotaUsed int64 `json:"employee_quota_used"`
	// ArticleQuotaUsed 已使用文章数量
	ArticleQuotaUsed int64 `json:"article_quota_used"`
}

// QuotaCheckReq 配额检查请求
type QuotaCheckReq struct {
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
	// ResourceType 资源类型
	ResourceType QuotaResourceType `json:"resource_type"`
}

// UsageQueryReq 用量查询请求
type UsageQueryReq struct {
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
}

// UsageQueryRes 用量查询响应
type UsageQueryRes struct {
	// CompanyID 公司ID
	CompanyID uint `json:"company_id"`
	// ResourceStats 资源用量统计列表，每个资源类型对应一个统计结果
	ResourceStats []ResourceUsageStatItem `json:"resource_stats"`
}

// ResourceUsageStatItem 资源用量统计项
// 每个资源类型对应一个统计项，不同资源类型可能有不同的统计方式
type ResourceUsageStatItem struct {
	// ResourceType 资源类型
	ResourceType QuotaResourceType `json:"resource_type"`
	// Usage 用量数量
	Usage int64 `json:"usage"`
}

type PackageQuota struct {
	PackageID     uint
	PackageLevel  foresttype.PackageLevel
	Quantity      int64
	Days          int64
	AgentQuota    int64
	QaQuota       int64
	DiskQuota     int64
	EmployeeQuota int64
	ArticleQuota  int64
}
