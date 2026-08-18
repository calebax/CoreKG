package dtocommon

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetCommonInfoResponse struct {
	apiobj.BaseResponse
	Response GetCommonInfoEmbedResponse
}

type GetCommonInfoEmbedResponse struct {
	// CompanyQuota 公司配额信息
	CompanyQuota CompanyQuota `json:"company_quota"`
}

type CompanyQuota struct {
	// IsPurchased 是否已购买
	IsPurchased bool `json:"is_purchased"`
	// AgentQuota 可创建智能体数量
	AgentQuota int64 `json:"agent_quota"`
	// AgentQuotaUsed 已使用智能体数量
	AgentQuotaUsed int64 `json:"agent_quota_used"`
	// QaQuota 每日问答次数
	QaQuota int64 `json:"qa_quota"`
	// QaQuotaUsed 已使用问答次数
	QaQuotaUsed int64 `json:"qa_quota_used"`
	// ArticleQuota 可创建文档数量
	ArticleQuota int64 `json:"article_quota"`
	// ArticleQuotaUsed 已使用文档数量
	ArticleQuotaUsed int64 `json:"article_quota_used"`
	// DiskQuota 磁盘配额，单位字节
	DiskQuota int64 `json:"disk_quota"`
	// DiskQuotaUsed 已使用磁盘空间，单位字节
	DiskQuotaUsed int64 `json:"disk_quota_used"`
	// EmployeeQuota 成员数量
	EmployeeQuota int64 `json:"employee_quota"`
	// EmployeeQuotaUsed 已使用成员数量
	EmployeeQuotaUsed int64 `json:"employee_quota_used"`
	CompanyQuota      uint  `json:"company_quota"`
}
