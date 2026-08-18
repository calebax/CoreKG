package svccompany

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtocompany"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func ExistCompany(ctx context.Context, name string) bool {
	var cnt int64
	if err := dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameCompany).
		Where("name = ?", name).
		Count(&cnt).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistCompany: get company count failed, %+v", err)
		return false
	}
	return cnt > 0
}

func CreateCompany(ctx *gin.Context, req *dtocompany.CreateCompanyRequest) (*dtocompany.CompanyInfo, *dtocompany.ErrorResponse) {
	setting, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "CreateCompany: get login setting failed, %s", err)
		return nil, &dtocompany.ErrorResponse{
			Code:    errcode.ErrCode_InternalError,
			Message: "account_get_login_setting_failed",
		}
	}

	if ExistCompany(ctx, req.Request.CompanyName) {
		logs.ErrorContextf(ctx, "CreateCompany: company name already exist, %s", req.Request.CompanyName)
		return nil, &dtocompany.ErrorResponse{
			Code:    errcode.ErrCode_BadRequest,
			Message: "account_company_name_exist",
		}
	}

	userId := req.Request.UserID

	var companyDTO *dtocompany.CompanyInfo
	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		// 创建公司
		comp, err := company.CreateCompany(ctx, tx, &company.CompanyInfo{
			Name:   req.Request.CompanyName,
			UserID: userId,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "CreateCompany failed, %+v", err)
			return err
		}
		// 创建公司uin身份
		uin, err := company.CreateEmployeeIdentification(ctx, tx, userId, comp.ID, setting.Issuer, req.Request.UserDisplayName)
		if err != nil {
			logs.ErrorContextf(ctx, " CreateCompany: CreateEmployeeIdentification failed, %+v", err)
			return err
		}

		dept := &accounttype.AccountDepartment{
			Name:      comp.Name,
			ParentID:  0,
			CompanyID: comp.ID,
			Sort:      1000,
		}
		if err = tx.Create(&dept).Error; err != nil {
			logs.ErrorContextf(ctx, "CreateCompany: create account department(name:%v) failed, %+v", comp.Name, err)
			return err
		}

		emp := &accounttype.Employee{
			CompanyID: comp.ID,
			UserID:    userId,
			Uin:       uin.ID,
			SysRole:   accounttype.SysRoleSysAdmin,
		}
		// 创建员工
		if err = employee.CreateEmployee(ctx, tx, emp); err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployee failed, %+v", err)
			return err
		}

		if err = tx.Create(&accounttype.AccountRelEmployeeDepartment{
			DepartmentID: dept.ID,
			Uin:          uin.ID,
			IsPrimary:    1,
			EmployeeID:   emp.ID,
			CompanyID:    comp.ID,
		}).Error; err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateAccountRelEmployeeDepartment(uin:%v|department_id:%v) failed, %+v", uin.ID, dept.ID, err)
			return err
		}

		companyDTO = &dtocompany.CompanyInfo{
			Uin:           *uin,
			CompanyLogo:   comp.Logo,
			CompanyName:   comp.Name,
			Role:          emp.SysRole,
			CompanyStatus: comp.CompanyStatus,
			LastLoginAt:   uin.LastLoginAt,
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "CreateCompany failed, %+v", err)
		return nil, &dtocompany.ErrorResponse{
			Code:    errcode.ErrCode_InternalError,
			Message: "account_create_company_failed",
		}
	}

	return companyDTO, nil
}
