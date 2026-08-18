package admintype

import (
	"gorm.io/gorm"
)

// Position 运营端职位表
type Position struct {
	gorm.Model

	Name        string `gorm:"column:name;type:varchar(64);not null;uniqueIndex;comment:'姓名'" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:'';comment:'描述'" json:"description"`
}

func (Position) TableName() string {
	return TableNamePosition
}

// Privilege 运营端权限表
type Privilege struct {
	gorm.Model

	//Name        string `gorm:"column:name;type:varchar(32);not null;default:'';comment:'姓名'" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:'';comment:'描述'" json:"description"`
	API         string `gorm:"column:api;type:varchar(64);not null;comment:'API'" json:"api"`
	Action      string `gorm:"column:action;type:varchar(64);not null;comment:'操作'" json:"action"`
	ActionPath  string `gorm:"column:action_path;type:varchar(64);not null;comment:'操作路径'" json:"action_path"`

	ParentID uint   `gorm:"column:parent_id;type:int;not null;default:0" json:"parent_id" `
	Type     string `gorm:"column:type;type:varchar(255);not null;" json:"type"`
}

func (Privilege) TableName() string {
	return TableNamePrivilege
}

type RelPositionPrivilege struct {
	gorm.Model

	PositionID  uint `gorm:"column:position_id;type:int;not null;uniqueIndex:uniidx_position_privilege" json:"position_id" `
	PrivilegeID uint `gorm:"column:privilege_id;type:int;not null;uniqueIndex:uniidx_position_privilege" json:"privilege_id" `
}

func (RelPositionPrivilege) TableName() string {
	return TableNameRelPositionPrivilege
}

type RelEmployeePosition struct {
	gorm.Model
	EmployeeID uint `gorm:"column:employee_id;type:int;not null;uniqueIndex:uniidx_employee_position" json:"employee_id" `
	PositionID uint `gorm:"column:position_id;type:int;not null;uniqueIndex:uniidx_employee_position" json:"position_id" `
}

func (RelEmployeePosition) TableName() string {
	return TableNameRelEmployeePosition
}
