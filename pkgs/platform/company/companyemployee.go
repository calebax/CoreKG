package company

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/account"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/platform/user"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateCompanyEmployeeOption 创建团队成员选项
type CreateCompanyEmployeeOption struct {
	CompanyID uint                `json:"company_id"`
	UserID    uint                `json:"user_id"`
	Role      accounttype.SysRole `json:"role"`
}

// IsExist 检查用户是否已经是该公司的成员
func (opt *CreateCompanyEmployeeOption) IsExist() (bool, error) {
	var count int64
	err := dbutil.Account().Table(accounttype.TableNameEmployee).
		Where("deleted_at is null").
		Where("company_id = ? AND user_id = ?", opt.CompanyID, opt.UserID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateCompanyEmployee 添加团队成员
func CreateCompanyEmployee(ctx context.Context, opt *CreateCompanyEmployeeOption) (*accounttype.Employee, error) {
	_, err := GetCompanyByID(opt.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] get company %d failed: %v", opt.CompanyID, err)
		return nil, err
	}
	u, err := user.GetUserByID(opt.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] get user %d failed: %v", opt.UserID, err)
		return nil, err
	}

	dep, err := account.NewAccountDepartmentDao().GetByCond(ctx, &account.DepartmentCond{
		CompanyID: opt.CompanyID,
		ParentID:  0,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] get department failed: %v", err)
		return nil, err
	} else if dep.ID == 0 {
		logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] get department failed: %v", err)
		return nil, fmt.Errorf("department not found")
	}

	// todo UIN封装,employee封装
	uin := &accounttype.UserIdentification{
		Name:        u.Name,
		UserID:      opt.UserID,
		SubjectType: accounttype.SubjectTypeCompany,
		SubjectID:   opt.CompanyID,
		UinStatus:   accounttype.UinStatusNormal,
		Issuer:      global.IssuerYYGU,
	}
	role := opt.Role
	if role == "" {
		role = accounttype.SysRoleSysEmployee
	}
	emp := &accounttype.Employee{
		CompanyID: opt.CompanyID,
		UserID:    opt.UserID,
		Uin:       0,
		SysRole:   role,
	}

	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(uin).Error; err != nil {
			logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] create uin failed: %v", err)
			return err
		}
		emp.Uin = uin.ID
		if err := tx.Create(emp).Error; err != nil {
			logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] create employee failed: %v", err)
			return err
		}

		rel := accounttype.AccountRelEmployeeDepartment{
			CompanyID:    opt.CompanyID,
			EmployeeID:   emp.ID,
			DepartmentID: dep.ID,
			IsPrimary:    1,
			Uin:          emp.Uin,
		}
		if err := tx.Create(&rel).Error; err != nil {
			logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] create rel failed: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		logs.ErrorContextf(ctx, "[company][CreateCompanyEmployee] create company employee failed: %v", err)
		return nil, err
	}
	return emp, nil
}

// QueryCompanyEmployeeListResponse 团队成员列表响应
type QueryCompanyEmployeeListResponse struct {
	apiobj.QueryResponse
	Data []*QueryCompanyEmployeeListItem `json:"data"`
}

// QueryCompanyEmployeeListItem 团队成员列表项
type QueryCompanyEmployeeListItem struct {
	accounttype.Employee
	CompanyID   uint   `gorm:"-" json:"company_id"` // todo 规范化输出适配，后续整理
	UserID      uint   `gorm:"-" json:"user_id"`    // todo 规范化输出适配，后续整理
	UserName    string `gorm:"column:user_name" json:"user_name"`
	UserPhone   string `gorm:"column:user_phone" json:"user_phone"`
	CompanyName string `gorm:"column:company_name" json:"company_name"`
}

// QueryCompanyEmployeeList 查询成员列表
func QueryCompanyEmployeeList(ctx context.Context, opt *apiobj.PageQuery, ret *QueryCompanyEmployeeListResponse) error {
	query := dbutil.Account().WithContext(ctx).Table(accounttype.TableNameEmployee).
		Joins("LEFT JOIN user on account_employee.user_id = user.id" +
			" and user.deleted_at is null").
		Joins("LEFT JOIN company on account_employee.company_id = company.id" +
			" and company.deleted_at is null").
		Where("account_employee.deleted_at is null")
	if opt.CompanyID > 0 {
		query = query.Where("account_employee.company_id = ?", opt.CompanyID)
	}

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "user_name", "name":
			query = query.Where("user.name like (?)", "%"+filter.Value[0]+"%")
		case "user_phone", "phone":
			query = query.Where("user.phone like (?)", "%"+filter.Value[0]+"%")
		case "user_id":
			query = query.Where("account_employee.user_id =?", filter.Value[0])
		case "company_id":
			query = query.Where("account_employee.company_id = ?", filter.Value[0])
		default:
			logs.WarnContextf(ctx, "[user][QueryCompanyEmployeeList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&ret.Total).Error; err != nil {
		return err
	}
	if ret.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	} else {
		query = query.Order("account_employee.id desc")
	}
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	} else {
		query = query.Limit(10)
	}

	err := query.Select("account_employee.*," +
		"user.name as user_name," +
		"user.phone as user_phone," +
		"company.name as company_name").
		Find(&ret.Data).Error
	for i, datum := range ret.Data {
		ret.Data[i].CompanyID = datum.Employee.CompanyID
		ret.Data[i].UserID = datum.Employee.UserID
	}
	if err != nil {
		return err
	}
	return nil
}

// UpdateCompanyEmployeeRole 更换成员角色
func UpdateCompanyEmployeeRole(id uint, role accounttype.SysRole) error {
	emp := &accounttype.Employee{}
	err := dbutil.Account().First(emp, id).Error
	if err != nil {
		return err
	}
	if emp.SysRole == role {
		return nil
	}
	emp.SysRole = role
	return dbutil.Account().Save(emp).Error
}
