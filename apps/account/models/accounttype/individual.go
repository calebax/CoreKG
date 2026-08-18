package accounttype

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"gorm.io/gorm"
)

type IndividualItemList struct {
	apiobj.QueryResponse
	Data []*Individual
}

// Individual 个人实名信息
type Individual struct {
	gorm.Model

	// // Uin 用户唯一标识
	// Uin uint `gorm:"column:uin;type:bigint;uniqueIndex;not null;comment:用户唯一标识" json:"uin"`

	// UserID userid
	UserID uint `gorm:"column:user_id;type:bigint;not null;uniqueIndex;comment:用户ID"`

	// RealName 真实姓名
	RealName string `gorm:"column:real_name;type:varchar(32);comment:真实姓名" json:"real_name"`
	// // AvatarURL 头像URL
	// AvatarURL string `gorm:"column:avatar_url;type:varchar(256);comment:头像URL" json:"avatar_url"`
	// IDCard 身份证号
	IDCard string `gorm:"column:id_card;type:varchar(18);comment:身份证号" json:"id_card"`
	// RealNameStatus 实名状态
	RealNameStatus IndividualStatus `gorm:"column:real_name_status;type:varchar(32);default:pending;comment:实名状态" json:"real_name_status"`
}

// TableName 表名
func (Individual) TableName() string {
	return "individual"
}

type IndividualStatus string

const (
	IndividualStatuPassed  IndividualStatus = "passed"
	IndividualStatuPending IndividualStatus = "pending"
	IndividualStatuFialed  IndividualStatus = "failed"
)
