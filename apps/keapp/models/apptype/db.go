package apptype

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/apps/keapp/models/web"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	TableNamePrefix        = "ke_"
	TableNameKeApplication = TableNamePrefix + "application"
)

func InitDB() error {
	if err := dbtools.InitModel(dbutil.Knownow(),
		&KeApplication{},
	); err != nil {
		return err
	}
	return web.InitDB()
}
