package foresttype

import (
	"github.com/insmtx/corekg/pkgs/types"
	"gorm.io/gorm"
)

// ForestHotWord 热词表结构体
type ForestHotWord struct {
	gorm.Model
	Uin       uint              `gorm:"column:uin;type:bigint;not null;default 0;comment:'用户ID'" json:"uin"`
	CompanyID uint              `gorm:"column:company_id;type:bigint;not null;default 0;comment:'公司ID'" json:"company_id"`
	HotWords  types.StringArray `gorm:"column:hot_words;type:text;comment:'热词列表'" json:"hot_words"`
}

type ForestHotWordList []ForestHotWord

func (ForestHotWord) TableName() string {
	return TableNameForestHotWord
}

func (l ForestHotWordList) ToMap() map[uint]ForestHotWord {
	m := make(map[uint]ForestHotWord)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}
