package admin

import (
	"sync"

	"github.com/insmtx/corekg/apps/admin/internal/apis"
	_ "github.com/insmtx/corekg/apps/admin/internal/docs"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"gorm.io/gorm"
)

// @title Roc API
// @host yygu.cn
// @BasePath /v2
// @schemes https http
// @accept json
// @produce json

// @param Env header string true "test"

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description					Description for what is this security definition being used

// Routers 路由注册
func Routers(eng *server.Router) error {
	// @host tapi.example.com
	apis.RegistryRouter(eng)
	return nil
}

// Migrates 补全数据表及数据库索引
func Migrates(db *gorm.DB) error {
	if err := admintype.InitCoreDB(); err != nil {
		return err
	}
	return nil
}

var onceStart sync.Once

// RunJob 启动定时任务
func RunJob() error {
	onceStart.Do(func() {
	})
	return nil
}
