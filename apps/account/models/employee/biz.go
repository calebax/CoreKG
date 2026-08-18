package employee

import (
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

// EmployeeInfoItemList 用户信息列表
type EmployeeInfoItemList struct {
	apiobj.QueryResponse
	Data []*EmployeeInfoItem
}

// EmployeeInfoItem 包含员工信息及其关联的用户、UIN 和实名信息
type EmployeeInfoItem struct {
	accounttype.Employee
	UserName       string                       `json:"user_name"`
	UserBio        string                       `json:"user_bio"`
	UserEmail      *string                      `json:"user_email"`
	UserPhone      *string                      `json:"user_phone"`
	UinStatus      accounttype.UinStatus        `json:"uin_status"`
	Issuer         string                       `json:"issuer"`
	RealName       string                       `json:"real_name"`
	AvatarURL      string                       `json:"avatar_url"`
	IDCard         string                       `json:"id_card"`
	RealNameStatus accounttype.IndividualStatus `json:"real_name_status"`
}

// CreateUserItem 创建用户
type CreateEmployeeItem struct {
	Username    string                 `json:"username"`
	RealName    string                 `json:"real_name"`
	Gender      accounttype.UserGender `json:"gender"`
	Phone       string                 `json:"phone"`
	Email       string                 `json:"email"`
	PositionIDs []uint                 `json:"position_ids" gorm:"-"`
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
	EmployeeID  uint                   `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Mobile      string                 `json:"phone"`
	Gender      accounttype.UserGender `json:"gender"`
	PositionIDs []uint                 `json:"position_ids" gorm:"-"`
}

type EmployeeDetail struct {
	accounttype.Employee
	UserName    string                  `json:"user_name"`
	RealName    string                  `json:"real_name"`
	Phone       string                  `json:"phone"`
	Email       string                  `json:"email"`
	Positions   []*accounttype.Position `json:"positions" gorm:"-"`
	ActionPaths []string                `json:"action_paths" gorm:"-"`
}

type EmployeeSimpleInfo struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

func (opt *EmployeeSimpleInfo) IsExist(exceptID ...uint) (bool, error) {
	var count int64
	query := dbutil.Account().Table(accounttype.TableNameEmployee).Where("deleted_at is null").
		Where("account_employee.email = ? or account_employee.mobile = ?", opt.Email, opt.Phone)
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

// CompanyEmployeeInfo 公司员工信息
type CompanyEmployeeInfo struct {
	accounttype.Employee
	CompanyLogo   string                    `json:"company_logo"`
	CompanyName   string                    `json:"company_name"`
	CompanyStatus accounttype.CompanyStatus `json:"company_status"`
	CompanyUserID uint                      `json:"company_user_id"`
}

type EmployeeSimpleInfoItemList struct {
	apiobj.QueryResponse
	Data []*EmployeeSimpleList
}

type EmployeeSimpleList struct {
	Uin       uint                `json:"uin"`
	ID        uint                `json:"id"`
	UserName  string              `json:"user_name"`
	Email     string              `json:"email"`
	Phone     string              `json:"phone"`
	CreatedAt time.Time           `json:"created_at"`
	SysRole   accounttype.SysRole `json:"sys_role"`
}
