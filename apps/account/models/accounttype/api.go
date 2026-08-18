package accounttype

import (
	"regexp"

	"gorm.io/gorm"
)

// APIStatus API授权状态
type APIStatus string

const (
	// APIStatusNormal 正常
	APIStatusNormal APIStatus = "normal"
	// APIStatusDisable 禁用
	APIStatusDisable APIStatus = "disable"
)

func (s APIStatus) IsNormal() bool {
	return s == APIStatusNormal
}

// APIPrivilege API权限表
type APIPrivilege struct {
	gorm.Model

	Description string `gorm:"column:description;type:varchar(255);not null;default:'';comment:'描述'" json:"description"`
	API         string `gorm:"column:api;type:varchar(64);not null;comment:'API';uniqueIndex;" json:"api"`
	Action      string `gorm:"column:action;type:varchar(64);not null;comment:'操作'" json:"action"`
	ActionPath  string `gorm:"column:action_path;type:varchar(64);not null;comment:'操作路径'" json:"action_path"`
	//
	ParentID   uint      `gorm:"column:parent_id;type:int;not null;default:0" json:"parent_id" `
	Type       string    `gorm:"column:type;type:varchar(255);not null;" json:"type"`
	Status     APIStatus `gorm:"column:status;type:varchar(32);not null;default:'normal';comment:'状态'" json:"status"`
	CallSchema string    `gorm:"column:call_schema;type:text;comment:'调用json schema'" json:"call_schema"`
}

func (APIPrivilege) TableName() string {
	return TableNameAPIPrivilege
}

func (api APIPrivilege) IsNormal() bool {
	return api.Status.IsNormal()
}

func (api APIPrivilege) Service() string {
	service, ok := FindAPIService(api.API)
	if !ok {
		return ""
	}
	return service
}

func FindAPIService(api string) (string, bool) {
	reg := regexp.MustCompile(`^(\w+)/\w+/\w+$`)
	if reg.MatchString(api) {
		return reg.FindStringSubmatch(api)[1], true
	} else {
		return "", false
	}
}

type APIService struct {
	gorm.Model
	Name   string `gorm:"column:name;type:varchar(32);not null;default:'';comment:'服务名称'" json:"name"`
	User   string `gorm:"column:user;type:varchar(32);not null;default:'';comment:'服务名称'" json:"user"`
	Passwd string `gorm:"column:passwd;type:varchar(64);not null;default:'';comment:'密码'" json:"-"`
}

func (APIService) TableName() string {
	return TableNameAPIService
}
