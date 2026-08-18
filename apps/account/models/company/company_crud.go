package company

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var DefaultCompanyQuota = &accounttype.ResourceQuota{DiskQuota: 10 * forest.GB, QAQuota: 1000, AgentQuota: 5, EmployeeQuota: 5, GraphQuota: 5, ArticleQuota: 5}

const (
	DefaultCompanyTemplate = "%v的组织%v"
)

// CreateCompany 创建公司
func CreateCompany(ctx context.Context, tx *gorm.DB, companyinfo *CompanyInfo) (*accounttype.Company, error) {
	company := &accounttype.Company{}
	company.Name = companyinfo.Name
	company.Alias = companyinfo.Alias
	company.Address = companyinfo.Address
	company.Tel = companyinfo.Tel
	company.Email = companyinfo.Email
	company.Website = companyinfo.Website
	company.Description = companyinfo.Description
	company.Logo = companyinfo.Logo
	company.CompanyStatus = accounttype.CompanyStatusPending
	// company.Quota = DefaultCompanyQuota
	company.UserID = companyinfo.UserID

	if err := tx.WithContext(ctx).Create(company).Error; err != nil {
		logs.ErrorContextf(ctx, "Create company fail: %v", err)
		return nil, err
	}
	return company, nil
}

// GetCompany 通过ID获取公司
func GetCompany(id uint) (*accounttype.Company, error) {
	var company accounttype.Company
	if err := dbutil.Account().Table(accounttype.TableNameCompany).Where("id = ?", id).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// GetCompanyByName 通过公司名称获取公司
func GetCompanyByName(name string) (*accounttype.Company, error) {
	var company accounttype.Company
	if err := dbutil.Account().Table(accounttype.TableNameCompany).Where("name = ?", name).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// UpdateCompany 修改公司信息
func UpdateCompany(company *accounttype.Company) error {
	return dbutil.Account().Table(accounttype.TableNameCompany).Save(company).Error
}

// QueryCompanyListResponse .
type QueryCompanyListResponse struct {
	apiobj.BaseResponse
	Response struct {
		apiobj.QueryResponse
		Data []*accounttype.Company
	}
}

// QueryCompanyList 查询公司列表
func QueryCompanyList(ctx context.Context, opt apiobj.PageQuery) (*QueryCompanyListResponse, error) {
	resp := &QueryCompanyListResponse{}
	query := dbutil.Account().Table(accounttype.TableNameCompany).WithContext(ctx).
		Where("deleted_at is null")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "alias":
			query = query.Where("alias LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "company_status":
			query = query.Where("company_status = ?", filter.Value[0])
		default:
			return nil, fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}
	// 处理 BeginTime 和 EndTime
	if !opt.BeginTime.IsZero() {
		query = query.Where("created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		query = query.Where("created_at <= ?", opt.EndTime)
	}

	if err := query.Count(&resp.Response.Total).Error; err != nil {
		return nil, err
	}
	if resp.Response.Total == 0 {
		return resp, nil
	}

	// 排序
	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	// 查询数据
	err := query.Find(&resp.Response.Data).Error
	if err != nil {
		logs.ErrorContextf(ctx, "QueryCompanyList err: %v", err)
		return nil, err
	}
	resp.Response.Limit = opt.Limit
	resp.Response.Offset = opt.Offset
	return resp, nil
}

// CreateEmployeeIdentification 创建员工标识
func CreateEmployeeIdentification(ctx context.Context, tx *gorm.DB, userID, companyID uint, issuer, name string) (*accounttype.UserIdentification, error) {
	// 初始化用户标识信息
	empIdentification := &accounttype.UserIdentification{
		UserID:      userID,
		Name:        name,
		SubjectType: accounttype.SubjectTypeCompany,
		UinStatus:   accounttype.UinStatusNormal,
		SubjectID:   companyID,
		Issuer:      issuer,
	}
	// 创建用户标识
	if err := tx.WithContext(ctx).Create(empIdentification).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: create uin failed, %+v", err)
		return nil, err
	}

	return empIdentification, nil
}

func ExistCompanyByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameCompany).
		Where("name = ?", name).
		Count(&count).
		Error
	if err != nil {
		logs.ErrorContextf(ctx, "ExistCompanyByName faild err: %v", err)
		return false, err
	}
	return count > 0, nil
}

func ExistEmployee(ctx context.Context, userID, companyID uint, uinStatus accounttype.UinStatus, issuer string) bool {
	var count int64
	if err := dbutil.Account().
		Table(accounttype.TableNameUserIdentification).WithContext(ctx).
		Where("user_id = ? AND subject_type = ? AND subject_id = ?",
			userID, accounttype.SubjectTypeCompany, companyID).
		Where("uin_status = ? AND issuer = ?",
			uinStatus, issuer).
		Where("deleted_at is null").
		Count(&count).Error; err != nil {
		logs.ErrorContextf(ctx,
			"ExistEmployee with user_id[%v] company_id[%v] uin_status[%v] issuer[%v] err[%v]",
			userID, companyID, uinStatus, issuer, err)
		return true
	}
	return count > 0
}

// UpdateUinLoginTime 更新用户标识最后登录时间
func UpdateUinLoginTime(ctx context.Context, tx *gorm.DB, uin uint) error {
	return tx.WithContext(ctx).
		Model(&accounttype.UserIdentification{}).
		Where("id = ?", uin).
		Update("last_login_at", time.Now()).Error
}
