package web

type CrawlRuleType string

const (
	CrawlRuleInclude CrawlRuleType = "include"
	CrawlRuleExclude CrawlRuleType = "exclude"
)

type CrawlTaskType string

const (
	CrawlTaskFull        CrawlTaskType = "full"
	CrawlTaskIncremental CrawlTaskType = "incremental"
	CrawlTaskSingle      CrawlTaskType = "single"
)

type CrawlTaskStatus string

const (
	CrawlTaskPending   CrawlTaskStatus = "pending"
	CrawlTaskRunning   CrawlTaskStatus = "running"
	CrawlTaskSuccess   CrawlTaskStatus = "success"
	CrawlTaskFailed    CrawlTaskStatus = "failed"
	CrawlTaskCancelled CrawlTaskStatus = "cancelled"
)

type IndexStatus string

const (
	IndexPending IndexStatus = "pending"
	IndexIndexed IndexStatus = "indexed"
	IndexFailed  IndexStatus = "failed"
)
