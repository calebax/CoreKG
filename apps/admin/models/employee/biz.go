package employee

import (
	"fmt"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"gorm.io/gorm"
)

// EmployeeInfoItemList 用户信息列表
type EmployeeInfoItemList struct {
	apiobj.QueryResponse
	Data []*EmployeeInfoItem
}

// EmployeeInfoItem 用户信息
type EmployeeInfoItem struct {
	gorm.Model

	Username string `json:"username"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Status   string `json:"status"`
	NickName string `json:"nick_name"`
	Gender   string `json:"gender"`
	UnionID  string `json:"union_id"`
}

// CreateUserItem 创建用户
type CreateEmployeeItem struct {
	Username    string               `json:"username"`
	RealName    string               `json:"real_name"`
	Gender      admintype.UserGender `json:"gender"`
	Phone       string               `json:"phone"`
	Email       string               `json:"email"`
	PositionIDs []uint               `json:"position_ids" gorm:"-"`
	Password    string               `json:"password"`
}

func (opt *CreateEmployeeItem) Check(exceptID ...uint) error {
	info := EmployeeSimpleInfo{
		Phone: opt.Phone,
		Email: opt.Email,
	}
	isExist, err := info.IsExist(exceptID...)
	if err != nil {
		return err
	}
	if isExist {

		return fmt.Errorf("用户信息已经存在")
	}
	return nil
}

// UpdateEmployeeItem 更新用户
type UpdateEmployeeItem struct {
	EmployeeID  uint                 `json:"id"`
	Username    string               `json:"username"`
	Email       string               `json:"email"`
	Mobile      string               `json:"phone"`
	RealName    string               `json:"real_name"`
	Gender      admintype.UserGender `json:"gender"`
	PositionIDs []uint               `json:"position_ids" gorm:"-"`
}

type EmployeeDetail struct {
	admintype.Employee

	Positions   []*admintype.Position `json:"positions"`
	ActionPaths []string              `json:"action_paths"`
}

type EmployeeSimpleInfo struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

func (opt *EmployeeSimpleInfo) IsExist(exceptID ...uint) (bool, error) {
	var count int64
	query := dbutil.Account().Table(admintype.TableNameEmployee).Where("deleted_at is null").
		Where("admin_employee.email = ? or admin_employee.mobile = ?", opt.Email, opt.Phone)
	if len(exceptID) > 0 {
		query.Where("id not in (?)", exceptID)
	}
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

type EmployeeWechatInfo struct {
	NickName  string `json:"nick_name"`
	AvatarURL string `json:"avatar_url"`
	UnionID   string `json:"union_id"`
	WebOpenID string `json:"web_open_id"`
}
