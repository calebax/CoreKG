package admintype

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/types"
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

// UserStatus 用户状态
type UserStatus string

const (
	// UserStatusNormal 正常
	UserStatusNormal UserStatus = "normal"
	// UserStatusDisabled 已禁用
	UserStatusDisabled UserStatus = "disabled"
)

// SysRole 系统角色
type SysRole string

const (
	// SysRoleSysAdmin 系统管理员
	SysRoleSysAdmin SysRole = "sys_admin"
)

// Employee 运营端员工表
type Employee struct {
	gorm.Model

	// Uin 员工唯一ID
	Uin uint `gorm:"column:uin;type:int unsigned;not null" json:"uin"`
	// Username 用户名，用于登录、显示，唯一，不可修改
	Username string `gorm:"column:username;type:varchar(16);not null;uniqueIndex:udx_username" json:"username"`
	// RealName 真实姓名
	RealName       string `gorm:"column:real_name;type:varchar(32);not null;comment:'姓名'" json:"real_name"`
	RealnamePinYin string `gorm:"column:realname_pinyin;type:varchar(64)" json:"realname_pinyin"`
	// NickName 微信昵称
	NickName    string         `gorm:"column:nick_name;type:varchar(32);not null;comment:'昵称'" json:"nick_name"`
	Gender      UserGender     `gorm:"column:gender;type:varchar(64);not null;comment:'性别'" json:"gender"`
	Email       *string        `gorm:"column:email;type:varchar(64);comment:邮箱" json:"email"`
	Mobile      *string        `gorm:"column:mobile;type:varchar(16);comment:手机" json:"mobile"`
	Password    types.Password `gorm:"column:password;type:varchar(128)" json:"password"`
	AvatarURL   string         `gorm:"column:avatar_url;type:varchar(256);comment:客户头像" json:"avatar_url"`
	UnionID     *string        `gorm:"column:union_id;type:varchar(64);comment:微信用户UnionID" json:"union_id"`
	WebOpenID   *string        `gorm:"column:web_open_id;type:varchar(64);comment:微信用户WebOpenID" json:"web_open_id"`
	WeComUserID *string        `gorm:"column:wecom_user_id;type:varchar(64);comment:企业微信用户ID" json:"wecom_user_id"`

	Status       UserStatus `gorm:"column:status;type:varchar(16);default:'normal'" json:"status"`
	SearchFilter string     `gorm:"column:search_filter;type:varchar(255)" json:"-"`

	SysRole SysRole `gorm:"column:sys_role;type:varchar(16);default:'';comment:系统角色" json:"sys_role"`
}

func (Employee) TableName() string {
	return TableNameEmployee
}

func (u *Employee) FillSearchFilter() {
	u.RealnamePinYin = strings.Join(pinyin.LazyPinyin(u.RealName, pinyin.NewArgs()), "")
	sli := []string{u.Username, u.RealName, u.RealnamePinYin}
	if u.Email != nil {
		sli = append(sli, *u.Email)
	}
	if u.Mobile != nil {
		sli = append(sli, *u.Mobile)
	}

	u.SearchFilter = strings.Join(sli, " ")
}

type CreatePositionOption struct {
	Position
	PrivilegeIDs []uint `json:"privilege_ids"`
}

type ListPositionResponse struct {
	apiobj.BaseResponse

	Response QueryPositionListResponse
}
type QueryPositionListResponse struct {
	apiobj.QueryResponse
	Data []*Position
}

// type BindType = string

// const (
// 	BindTypeWechat    = "wechat_unionid"
// 	BindTypeWechatCom = "wechat_com"
// )

// // EmployeeThirdBinding 第三方账号绑定
// type EmployeeThirdBinding struct {
// 	gorm.Model

// 	EmployeeID uint `gorm:"column:employee_id;type:int;not null;index:idx_employee_id" json:"employee_id"`
// 	// 绑定的类型（微信web和企业微信）及对应的ID
// 	BindType  BindType `gorm:"column:bind_type;type:varchar(64);unique_index:idx_bind_type_value"`
// 	BindValue string   `gorm:"column:bind_value;type:varchar(64);unique_index:idx_bind_type_value"`
// }

// func (*EmployeeThirdBinding) TableName() string { return TableNameEmployeeThirdBinding }
