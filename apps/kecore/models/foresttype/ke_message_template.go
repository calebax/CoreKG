package foresttype

import (
	"gorm.io/gorm"
)

type MessageTemplateType string

const (
	MessageTemplateTypeAnnouncement MessageTemplateType = "announcement"
	MessageTemplateTypeSystem       MessageTemplateType = "system"
)

type MessageModule string

const (
	MessageModuleUserManagement  MessageModule = "announcement"
	MessageModuleOrderManagement MessageModule = "payment"
)

type MessageTemplateName string

const (
	MessageTemplateNameAnnouncementNewRelease MessageTemplateName = "announcement_new_release"
	MessageTemplateNameOrderAboutToExpire     MessageTemplateName = "order_about_to_expire"
	MessageTemplateNameAlreadyUploadSameFile  MessageTemplateName = "already_upload_same_file"
)

// KeMessageTemplate 消息模板表结构体
type KeMessageTemplate struct {
	gorm.Model
	Name            MessageTemplateName `gorm:"column:name;type:varchar(64);not null;;comment:模板名称，如公告、下单等，具有唯一性"`
	Description     string              `gorm:"column:description;type:varchar(128);;;comment:模板描述"`
	Type            MessageTemplateType `gorm:"column:type;type:varchar(32);not null;;comment:模板类型：system-系统消息，announcement-公告消息"`
	TitleTemplate   string              `gorm:"column:title_template;type:varchar(256);not null;;comment:标题模板，支持 {{variable}} 占位符"`
	ContentTemplate string              `gorm:"column:content_template;type:varchar(256);not null;;comment:内容模板，支持 {{variable}} 占位符"`
	Module          MessageModule       `gorm:"column:module;type:varchar(64);;;comment:功能模块，表示所属哪个功能模块"`
	RoutePath       string              `gorm:"column:route_path;type:varchar(256);;;comment:前端路由路径模板，支持 {{variable}} 占位符，如：/order/detail?id={{order_id}}"`
	Status          string              `gorm:"column:status;type:varchar(16);not null;default draft;comment:模板状态：draft-草稿，enable-启用，disable-禁用"`
}

type KeMessageTemplateList []KeMessageTemplate

func (KeMessageTemplate) TableName() string {
	return TableNameKeMessageTemplate
}

func (l KeMessageTemplateList) ToMap() map[uint]KeMessageTemplate {
	m := make(map[uint]KeMessageTemplate)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}

func (l KeMessageTemplateList) ToNameMap() map[MessageTemplateName]KeMessageTemplate {
	m := make(map[MessageTemplateName]KeMessageTemplate)
	for _, v := range l {
		m[v.Name] = v
	}
	return m
}
