package kecore

import (
	"context"
	"sync"

	"github.com/insmtx/corekg/apps/kecore/internal/apis"
	_ "github.com/insmtx/corekg/apps/kecore/internal/docs"
	"github.com/insmtx/corekg/apps/kecore/jobs"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"gorm.io/gorm"
)

// @title kecore API
// @description doc.json
// @host tapi.example.com
// @BasePath /apis/p
// @schemes http
// @accept json
// @produce json

// @param Env header string true "test"

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description					Description for what is this security definition being used

// Routers 路由注册
func Routers(eng *server.Router) error {
	apis.RegistryRouter(eng)
	return nil
}

// Migrates 补全数据表及数据库索引
func Migrates(db *gorm.DB) error {
	return nil
}

var onceStart sync.Once

// RunJob 启动定时任务
func RunJob(ctx context.Context) error {
	onceStart.Do(func() {
		go jobs.CoreCornJob(ctx)
	})
	return nil
}
