package login_setting

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	// LoginWayOpoEmail OPO业务邮箱登录
	LoginWayOpoEmail auth.LoginWay = 100
)

// InitDB 初始化数据库
func InitDB() error {
	err := dbtools.InitModel(dbutil.Account(), &LoginSetting{})
	if err != nil {
		return err
	}
	return nil
}
