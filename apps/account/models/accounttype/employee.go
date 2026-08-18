package accounttype

import (
	"gorm.io/gorm"
)

const SystemUserID = "system"

// UserGender 用户性别
type UserGender int

const (
	// UserGenderUnspecified 性别未定义
	UserGenderUnspecified UserGender = 0
	// UserGenderMale 男性
	UserGenderMale UserGender = 1
	// UserGenderFemale 女性
	UserGenderFemale UserGender = 2
)

// SysRole 系统角色
type SysRole string

const (
	// SysRoleSysAdmin 系统管理员
	SysRoleSysAdmin SysRole = "sys_admin"
	// SysRoleSysEmployee 公司员工
	SysRoleSysEmployee SysRole = "sys_employee"
	// SysRoleTeacher 教师
	SysRoleTeacher SysRole = "sys_teacher"
	// SysRoleStudent 学生
	SysRoleStudent SysRole = "sys_student"
)

// Employee 运营端员工表
type Employee struct {
	gorm.Model
	// 公司id
	CompanyID uint `gorm:"column:company_id;type:bigint;not null;comment:公司ID"` // todo 规范化
	// 员工id
	UserID uint `gorm:"column:user_id;type:bigint;not null;comment:用户ID"` // todo json规范化
	// 员工uin
	Uin     uint    `gorm:"column:uin;type:bigint;not null;comment:uin"`
	SysRole SysRole `gorm:"column:sys_role;type:varchar(16);default:'';comment:系统角色" json:"sys_role"`
}

func (Employee) TableName() string {
	return TableNameEmployee
}
