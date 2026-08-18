package migrator

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type DepartmentMigrator struct {
}

func (m *DepartmentMigrator) Migrator(ctx context.Context) error {
	logs.InfoContextf(ctx, "DepartmentMigrator start")
	//create top department
	var (
		cmps []accounttype.Company
	)
	if err := dbutil.Account().Table(accounttype.TableNameCompany).
		Where("deleted_at IS NULL").
		Find(&cmps).
		Error; err != nil {
		panic(err)
	}
	//batch add
	if err := dbutil.Account().Transaction(func(tx *gorm.DB) error {
		//get company's employee
		for _, cmp := range cmps {
			if len(cmp.Name) > 0 {
				logs.DebugContextf(ctx, "DepartmentMigrator start process company[id:%v|name:%v]: %s", cmp.ID, cmp.Name)
			} else {
				logs.WarnContextf(ctx, "company[id:%v]'s name is null", cmp.ID)
			}
			var c int64
			if err := tx.Table(accounttype.TableNameAccountDepartment).
				Where("deleted_at IS NULL").
				Where("company_id = ?", cmp.ID).
				Count(&c).Error; err != nil {
				logs.ErrorContextf(ctx, "get company department count error = %v", err)
				return err
			}
			if c > 0 {
				logs.InfoContextf(ctx, "get company department count = %d", c)
				logs.InfoContextf(ctx, "migrate department for company(id=%v) would be ignored", cmp.ID)
				continue
			}

			dep := accounttype.AccountDepartment{
				Name:      cmp.Name,
				ParentID:  0,
				CompanyID: cmp.ID,
				Sort:      1000,
			}

			if err := tx.Create(&dep).Error; err != nil {
				return err
			}

			emps, err := GetCompanyEmployeeInfo(ctx, cmp.ID)
			if err != nil {
				return err
			}
			var depRels []accounttype.AccountRelEmployeeDepartment
			for _, emp := range emps {
				depRels = append(depRels, accounttype.AccountRelEmployeeDepartment{
					Uin:          emp.Uin,
					DepartmentID: dep.ID,
					IsPrimary:    1,
					EmployeeID:   emp.ID,
					CompanyID:    cmp.ID,
				})
			}
			logs.DebugContextf(ctx, "migrate department rel[total:%v] for company(id=%v) would be created", len(depRels), cmp.ID)

			//create rel
			if err = tx.CreateInBatches(depRels, len(depRels)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "migrate department err: %v", err)
		return err
	}
	logs.InfoContextf(ctx, "DepartmentMigrator end")
	return nil
}
func GetCompanyEmployeeInfo(ctx context.Context, companyID uint) ([]employee.EmployeeSimpleList, error) {
	var res []employee.EmployeeSimpleList
	if err := dbutil.Account().WithContext(ctx).Table(accounttype.TableNameEmployee+" e").
		Select("e.uin, e.id, u.name AS user_name,us.phone as phone,us.email as email, e.created_at,e.sys_role as sys_role").
		Joins("LEFT JOIN user_identification u ON e.user_id = u.user_id "+
			"AND (u.subject_type = ? AND u.subject_id = ?) AND u.deleted_at IS NULL AND e.uin = u.id", accounttype.SubjectTypeCompany, companyID).
		Joins("INNER JOIN user us ON us.id = u.user_id AND us.deleted_at IS NULL").
		Where("e.company_id = ?", companyID).
		Where("e.deleted_at IS NULL").
		Find(&res).Error; err != nil {
		logs.ErrorContextf(ctx, "GetCompanyEmployeeInfo() error = %v", err)
		return nil, err
	}
	return res, nil
}
