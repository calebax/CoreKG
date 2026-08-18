package apptype

type AppStatus string

const (
	AppStatusDraft  AppStatus = "draft"
	AppStatusOnline AppStatus = "online"
	AppStatusPaused AppStatus = "paused"
)

type SyncStatus string

const (
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
	SyncStatusSyncing SyncStatus = "syncing"
)

type AppTemplateType string

const (
	AppTemplateWebsite    AppTemplateType = "website"
	AppTemplateProduct    AppTemplateType = "product"
	AppTemplateAftersales AppTemplateType = "aftersales"
	AppTemplateTraining   AppTemplateType = "training"
	AppTemplatePolicy     AppTemplateType = "policy"
)
