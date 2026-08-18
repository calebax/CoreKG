package corekg

import (
	"sync"
	"time"

	"github.com/insmtx/corekg/apps/account"
	"github.com/insmtx/corekg/apps/corekg/internal/apis"
	"github.com/insmtx/corekg/apps/keapi"
	"github.com/insmtx/corekg/apps/keapp"
	"github.com/insmtx/corekg/apps/kechat"
	"github.com/insmtx/corekg/apps/kecore"
	"github.com/insmtx/corekg/apps/ketask"
	"github.com/insmtx/corekg/apps/kesearch"
	"github.com/insmtx/corekg/pkgs/apis/wecom"
	"github.com/insmtx/corekg/pkgs/global"
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
	kecore.Routers(eng)
	kechat.Routers(eng)
	account.Routers(eng)
	keapi.Routers(eng)
	wecom.RegistryRouter(eng.GinEngine().Group(global.PrefixAPIV2 + "account"))
	wecom.RegistryRouter(eng.GinEngine().Group(global.PrefixAPIV3 + "account"))
	ketask.Routers(eng)
	kesearch.Routers(eng)
	keapp.Routers(eng)

	apis.RegistryRouter(eng)
	return nil
}

// Migrates 补全数据表及数据库索引
func Migrates(db *gorm.DB) error {
	return nil
}

var onceStart sync.Once

// RunJob 启动定时任务
func RunJob(db *gorm.DB) error {
	onceStart.Do(func() {
		go func() {
			ticker := time.NewTicker(1 * time.Minute) // 每分钟执行一次
			defer ticker.Stop()

			for {
				select {

				case <-ticker.C:
					// 每分钟执行一次函数 a
					//fileqa.CleanTimeoutApi()
				}
			}
		}()
	})
	return nil
}
