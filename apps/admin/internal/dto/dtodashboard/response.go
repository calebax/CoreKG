package dtodashboard

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetDashboardOverviewResponse struct {
	apiobj.BaseResponse
	Response GetDashboardOverviewEmbedResponse
}

type GetDashboardOverviewEmbedResponse struct {
	// BasicStatistics 基础统计
	BasicStatistics BasicStatistics `json:"basic_statistics"`
	// MultimodalStatistics 多模态资源统计
	MultimodalStatistics MultimodalStatistics `json:"multimodal_statistics"`
	// GraphStatistics 图谱统计
	GraphStatistics GraphStatistics `json:"graph_statistics"`
}

type Stats struct {
	EmployeeCount int `json:"employee_count"`
	QACount       int `json:"qa_count"`
	ForestCount   int `json:"forest_count"`
	GraphCount    int `json:"graph_count"`
	AgentCount    int `json:"agent_count"`
	ArticleCount  int `json:"article_count"`
}

type MultimodalStatistics struct {
	// UploadCount 上传文件数量
	UploadCount int64 `json:"upload_count"`
	// ReadyCount 资源就绪数量
	ReadyCount int64 `json:"ready_count"`
	// ParseSuccessCount 资源解析成功数
	ParseSuccessCount int64 `json:"parse_success_count"`
	// ParseFailCount 资源解析失败数
	ParseFailCount int64 `json:"parse_fail_count"`
	// ParseAvgCost 解析平均耗时
	ParseAvgCost int64 `json:"parse_avg_cost"`
	// ParseSuccessRate 解析成功率
	ParseSuccessRate string `json:"parse_success_rate"`
	// IndexSuccessCount 索引成功文件数量
	IndexSuccessCount int64 `json:"index_success_count"`
	// IndexFailCount 索引失败文件数量
	IndexFailCount int64 `json:"index_fail_count"`
	// IndexAvgCost 索引平均耗时
	IndexAvgCost int64 `json:"index_avg_cost"`
	// IndexSuccessRate 索引成功率
	IndexSuccessRate string `json:"index_success_rate"`
	// SummarySuccessCount 摘要成功文件数量
	SummarySuccessCount int64 `json:"summary_success_count"`
	// SummaryFailCount 摘要失败文件数量
	SummaryFailCount int64 `json:"summary_fail_count"`
	// SummaryAvgCost 摘要平均耗时
	SummaryAvgCost int64 `json:"summary_avg_cost"`
	// SummarySuccessRate 摘要成功率
	SummarySuccessRate string `json:"summary_success_rate"`
}

type BasicStatistics struct {
	// UserCount 用户总数
	UserCount int64 `json:"user_count"`
	// NewUserCount 新增用户数量
	NewUserCount int64 `json:"new_user_count"`
	// ForestCount 知识库数量
	ForestCount int64 `json:"forest_count"`
	// QACount 问答数量
	QACount int64 `json:"qa_count"`
	// GraphCount 图谱数量
	GraphCount int64 `json:"graph_count"`
	// AgentCount 智能体数量
	AgentCount int64 `json:"agent_count"`
}

type GraphStatistics struct {
	// SuccessCount 图谱成功数量
	SuccessCount int64 `json:"success_count"`
	// FailCount 图谱失败数量
	FailCount int64 `json:"fail_count"`
	// SuccessRate 图谱成功率
	SuccessRate string `json:"success_rate"`
}
