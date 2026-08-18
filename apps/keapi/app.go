package keapi

import (
	"sync"

	"github.com/insmtx/corekg/apps/keapi/internal/apis"
	keapimcp "github.com/insmtx/corekg/apps/keapi/internal/mcp"
	_ "github.com/insmtx/corekg/apps/keapi/internal/docs"
	"github.com/insmtx/corekg/apps/keapi/internal/migrate"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"gorm.io/gorm"
)

// @title keapi API
// @description external knowledge api service
// @host 127.0.0.1:8086
// @BasePath /v3
// @schemes http
// @accept json
// @produce json
//
// @param Env header string false "test"
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer API Key

// Routers 路由注册
func Routers(eng *server.Router) error {
	apis.RegistryRouter(eng)
	keapimcp.RegistryRouter(eng)
	return nil
}

// Migrates 补全数据表及数据库索引
func Migrates(_ *gorm.DB) error {
	return migrate.InitDB()
}

var onceStart sync.Once

// RunJob 启动定时任务
func RunJob() error {
	onceStart.Do(func() {})
	return nil
}
