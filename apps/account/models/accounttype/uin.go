package accounttype

import (
	"time"

	"gorm.io/gorm"
)

// SubjectType 主体类型
type SubjectType string

const (
	// SubjectTypeIndividual 个人
	SubjectTypeIndividual SubjectType = "individual"
	// SubjectTypeCompany 公司
	SubjectTypeCompany SubjectType = "company"
)

// UinStatus 状态
type UinStatus string

const (
	// UinStatusNormal 正常
	UinStatusNormal UinStatus = "normal"
	// UinStatusDisabled 禁用
	UinStatusDisabled UinStatus = "disabled"
	// UinStatusDeleted 删除
	UinStatusDeleted UinStatus = "deleted"
)

// UserIdentification 用户标识 UIN
type UserIdentification struct {
	gorm.Model

	// UserID 用户标识
	UserID uint `gorm:"column:user_id;type:bigint;not null;comment:用户ID"`
	// SubjectType 主体类型
	SubjectType SubjectType `gorm:"column:subject_type;type:varchar(16);not null;comment:主体类型"`
	// SubjectID 主体ID
	SubjectID uint `gorm:"column:subject_id;type:bigint;not null;comment:主体ID"`
	// UinStatus 状态
	UinStatus UinStatus `gorm:"column:uin_status;type:varchar(16);not null;comment:状态;default:'normal'"`
	Issuer    string    `gorm:"column:issuer;type:varchar(128);not null;comment:颁发者"`
	Name      string    `gorm:"column:name;type:varchar(128);comment:用户uin关联名"`
	// LastLoginAt 最后登录时间
	LastLoginAt *time.Time `gorm:"column:last_login_at;comment:最后登录时间"`
}

type UserIdentificationList []UserIdentification

// TableName 表名
func (UserIdentification) TableName() string {
	return "user_identification"
}
