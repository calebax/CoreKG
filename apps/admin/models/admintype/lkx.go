package admintype

import (
	"gorm.io/gorm"
)

type LkxCustomerInfo struct {
	gorm.Model
	CompanyID   uint   `gorm:"column:company_id;type:bigint;not null;comment:公司ID" json:"company_id"`
	Province    string `gorm:"column:province;type:varchar(255);not null;default:'';comment:省份" json:"province"`
	City        string `gorm:"column:city;type:varchar(255);not null;default:'';comment:城市" json:"city"`
	Name        string `gorm:"column:name;type:varchar(255);not null;default:'';comment:名称" json:"name"`
	Phone       string `gorm:"column:phone;type:varchar(255);not null;default:'';comment:电话" json:"phone"`
	CompanyName string `gorm:"column:company_name;type:varchar(255);not null;default:'';comment:公司名称" json:"company_name"`
	Industry    string `gorm:"column:industry;type:varchar(255);not null;default:'';comment:行业" json:"industry"`
	Produce     string `gorm:"column:produce;type:varchar(255);not null;default:'';comment:产品" json:"produce"`
	Description string `gorm:"column:description;type:text;not null;comment:描述" json:"description"`
	Email       string `gorm:"column:email;type:varchar(255);not null;default:'';comment:邮箱" json:"email"`
	Position    string `gorm:"column:position;type:varchar(255);not null;default:'';comment:职位" json:"position"`
}

func (*LkxCustomerInfo) TableName() string {
	return TableNameLkxCustomerInfo
}
