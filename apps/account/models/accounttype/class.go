package accounttype

import "gorm.io/gorm"

// Class 学习班信息表
type Class struct {
	gorm.Model

	// CompanyID 公司id
	CompanyID uint `gorm:"column:company_id;type:bigint;not null;comment:公司ID" json:"company_id"`
	// Name 班名
	Name string `gorm:"column:name;type:varchar(32);not null;comment:班名" json:"name"`
	// Founder 创建人
	Founder uint `gorm:"column:founder;type:bigint;not null;comment:创建人" json:"founder"`
	// StudyYear 学年
	StudyYear uint `gorm:"column:study_year;type:bigint;not null;comment:学年" json:"study_year"`
	// Status 状态
	Status ClassStatus `gorm:"column:status;type:varchar(16);not null;comment:状态;default:'not_started'" json:"status"`
}

// TableName 表名
func (Class) TableName() string {
	return TableNameClass
}

// ClassStatus 状态
type ClassStatus string

const (
	ClassStatusNotStarted ClassStatus = "not_started"
	ClassStatusStarted    ClassStatus = "started"
	ClassStatusFinished   ClassStatus = "finished"
)

// ClassStudent 学生班级表
type ClassStudent struct {
	StudentID uint `gorm:"column:student_id;type:bigint;not null;comment:学生ID" json:"student_id"`
	ClassID   uint `gorm:"column:class_id;type:bigint;not null;comment:班级ID" json:"class_id"`
}

// TableName 表名
func (ClassStudent) TableName() string {
	return TableNameClassStudent
}

// ClassTeacher 老师班级表
type ClassTeacher struct {
	TeacherUin uint `gorm:"column:teacher_uin;type:bigint;not null;comment:教师uin" json:"teacher_uin"`
	ClassID    uint `gorm:"column:class_id;type:bigint;not null;comment:班级ID" json:"class_id"`
}

// TableName 表名
func (ClassTeacher) TableName() string {
	return TableNameClassTeacher
}
