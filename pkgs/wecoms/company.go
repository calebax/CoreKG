package wecoms

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Name      string `gorm:"column:name;type:varchar(32)" json:"name"`
	Namespace string `gorm:"column:namespace;type:varchar(16)" json:"namespace"`
}

func (*Company) TableName() string { return TableNameCompany }
