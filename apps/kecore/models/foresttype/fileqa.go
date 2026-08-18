package foresttype

import (
	"gorm.io/gorm"
)

type QAStatus string

const (
	QAStatusPending  QAStatus = "pending"  // 待处理
	QAStatusAnswered QAStatus = "answered" // 已回答
	QAStatusFailed   QAStatus = "failed"   // 回答失败
	QAStatusDivide   QAStatus = "divide"   // 知识森林拆分子问题和思维导图
)

// KnownowFileQA 单文档智慧问答
type KnownowFileQA struct {
	gorm.Model

	Uin       uint `gorm:"column:uin;type:bigint;not null;default:0;comment:'用户ID';index" json:"uin"`
	CompanyID uint `gorm:"column:company_id;type:bigint;not null;default:0;comment:'公司ID';index" json:"company_id"`
	ForestID  uint `gorm:"type:int;column:forest_id;not null;comment:forest_id" json:"forest_id"`
	FileID    uint `gorm:"type:int;column:file_id;not null;comment:forest_id" json:"file_id"`
	// 聊天信息
	// Question 问题
	Question string `gorm:"column:question;type:text;comment:'问题'" json:"question"`
	// Answer 回答
	Answer string `gorm:"column:answer;type:text;comment:'回答'" json:"answer"`
	// Status 状态
	Status QAStatus `gorm:"column:status;type:varchar(8);not null;default:'pending';comment:'状态'" json:"status"`
}

func (KnownowFileQA) TableName() string {
	return TableNameKnownowFileQA
}
