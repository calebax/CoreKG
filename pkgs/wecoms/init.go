package wecoms

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	tableNamePrefix      = "wecom_"
	TableNameCompany     = tableNamePrefix + "company"
	TableNameApp         = tableNamePrefix + "app"
	TableNameDept        = tableNamePrefix + "dept"
	TableNameUser        = tableNamePrefix + "user"
	TableNameRelDeptUser = tableNamePrefix + "rel_dept_user"
)

func InitDB() error {
	return dbtools.InitModel(dbutil.Account(),
		&Company{},
		&App{},
		&Department{},
		&RelDeptUser{},
	)
}
