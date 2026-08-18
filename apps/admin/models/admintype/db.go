package admintype

import (
	"github.com/insmtx/corekg/apps/admin/models/prompt"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/storage"
)

const (
	TableNameEmployee             = "admin_employee"
	TableNamePosition             = "admin_position"
	TableNamePrivilege            = "admin_privilege"
	TableNameRelPositionPrivilege = "admin_rel_position_privilege"
	TableNameRelEmployeePosition  = "admin_rel_employee_position"
	//TableNameCompany              = "admin_company"

	TableNameLicense           = "admin_license"
	TableNameAdminAnnouncement = "admin_announcement"

	TableNameLkxCustomerInfo = "lkx_customer_info"
)

// InitModel init db tables
func InitDB() error {
	err := dbtools.InitModel(dbutil.Account(),
		&Employee{},
		// &EmployeeThirdBinding{},
		&Position{},
		&Privilege{},
		&RelEmployeePosition{},
		&RelPositionPrivilege{},
		&LkxCustomerInfo{},
	)
	if err != nil {
		return err
	}

	{
		// db_pre
		if err := presetDatabase(); err != nil {
			return err
		}
	}
	return nil
}

// 数据库初始化准备 生成职位权限
func presetDatabase() error {
	return nil
}

func InitCoreDB() error {
	err := dbtools.InitModel(dbutil.Core(),
		&storage.FileInfo{},
	)
	if err != nil {
		return err
	}

	if err := prompt.InitModel(dbutil.Core()); err != nil {
		return err
	}

	{ // db_pre
		if err := presetDatabase(); err != nil {
			return err
		}
	}
	return nil
}
