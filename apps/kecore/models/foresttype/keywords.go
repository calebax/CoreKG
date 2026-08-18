package foresttype

import (
	"gorm.io/gorm"
)

type WordType string

const (
	WordTypeSynonym WordType = "synonym"
	WordTypeMajor   WordType = "major"
)

// Keywords 同义词关键词表表结构体
type Keywords struct {
	gorm.Model
	CompanyID   uint     `gorm:"column:company_id;type:bigint;not null;default 0;comment:'公司ID'" json:"company_id"`
	Description string   `gorm:"column:description;type:varchar(255);;;comment:description 专业名词描述" json:"description"`
	SubjectID   uint     `gorm:"column:subject_id;type:bigint;not null;;comment:subject id" json:"subject_id"`
	Uin         uint     `gorm:"column:uin;type:bigint;not null;default 0;comment:'用户ID'" json:"uin"`
	Word        string   `gorm:"column:word;type:varchar(255);;;comment:word 词内容" json:"word"`
	WordType    WordType `gorm:"column:word_type;type:varchar(32);not null;;comment:词典类型 synonym:同义词；major：专业" json:"word_type"`
}

type KeywordsList []Keywords

func (Keywords) TableName() string {
	return TableNameKeywords
}

func (l KeywordsList) ToMap() map[uint]Keywords {
	m := make(map[uint]Keywords)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}
