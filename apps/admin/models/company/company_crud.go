package company

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateCompanyOption 创建团队Option
type CreateCompanyOption struct {
	Name string `json:"name"`
}

// CreateCompany 创建团队
func CreateCompany(ctx context.Context, tx *gorm.DB, opt *CreateCompanyOption) (*accounttype.Company, error) {
	c := &accounttype.Company{
		Name:          opt.Name,
		CompanyStatus: accounttype.CompanyStatusPending,
		Quota:         company.DefaultCompanyQuota,
	}

	dep := &accounttype.AccountDepartment{
		Name:     opt.Name,
		ParentID: 0,
		Sort:     1000,
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(c).Error; err != nil {
			logs.ErrorContextf(ctx, "[company][CreateCompany] create company failed: %v", err)
			return err
		}

		dep.CompanyID = c.ID
		if err := tx.Create(dep).Error; err != nil {
			logs.ErrorContextf(ctx, "[company][CreateCompany] create department failed: %v", err)
			return err
		}

		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "[company][CreateCompany] create company failed: %v", err)
		return nil, err
	}

	return c, nil
}

// GetCompanyByName 通过名称获取团队
func GetCompanyByName(name string) (*accounttype.Company, error) {
	out := &accounttype.Company{}
	if err := dbutil.Account().Where("name = ?", name).First(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetCompanyByID 通过ID获取团队
func GetCompanyByID(id uint) (*accounttype.Company, error) {
	out := &accounttype.Company{}
	if err := dbutil.Account().First(out, id).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// IsExistCompanyByName 判断团队是否存在
func IsExistCompanyByName(name string, exceptIDs ...uint) (bool, error) {
	query := dbutil.Account().Table(accounttype.TableNameCompany).Where("deleted_at is null").
		Where("name = ?", name)
	if len(exceptIDs) > 0 {
		query = query.Where("id not in (?)", exceptIDs)
	}
	err := query.First(&accounttype.Company{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// QueryCompanyListResponse 团队列表响应
type QueryCompanyListResponse struct {
	apiobj.QueryResponse
	Data []*QueryCompanyListItem
}

// QueryCompanyListItem 团队列表项
type QueryCompanyListItem struct {
	accounttype.Company
	EmployeeCount int64 `gorm:"column:employee_count" json:"employee_count"`
}

// QueryCompanyList 查询团对列表
func QueryCompanyList(ctx context.Context, opt *apiobj.PageQuery, ret *QueryCompanyListResponse) error {
	query := dbutil.Account().WithContext(ctx).Table(accounttype.TableNameCompany).
		Joins("LEFT JOIN (select company_id,count(*) as c from account_employee where deleted_at is null " +
			"group by company_id) as emp on emp.company_id=company.id").
		Where("company.deleted_at is null")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("company.name like (?)", "%"+filter.Value[0]+"%")
		case "id":
			query = query.Where("company.id like (?)", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[user][QueryCompanyList] invalid filter field: %s", filter.Field)
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
		query = query.Order("company.id desc")
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	} else {
		query = query.Limit(10)
	}

	err := query.Select("company.*,emp.c as employee_count").
		Find(&ret.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// ModifyCompanyOption 修改团队信息
type ModifyCompanyOption struct {
	Name string `json:"name"`
}

// ModifyCompany 修改团队信息
func ModifyCompany(id uint, opt *ModifyCompanyOption) (*accounttype.Company, error) {
	c, err := GetCompanyByID(id)
	if err != nil {
		return nil, err
	}
	if c.Name == opt.Name {
		return c, nil
	}
	c.Name = opt.Name
	if err := dbutil.Account().Save(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}
