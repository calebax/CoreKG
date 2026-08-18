package account

import (
	"context"
	"testing"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

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

func TestMigrateDepartment(t *testing.T) {
	//get company
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
	//create top department
	var (
		ctx  = context.TODO()
		cmps []accounttype.Company
	)
	if err := dbutil.Account().Table(accounttype.TableNameCompany).
		Where("deleted_at IS NULL").
		Find(&cmps).
		Error; err != nil {
		panic(err)
	}
	//batch add
	dbutil.Account().Transaction(func(tx *gorm.DB) error {
		//get company's employee
		for _, cmp := range cmps {
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
			//create rel
			if err = tx.CreateInBatches(depRels, len(depRels)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
