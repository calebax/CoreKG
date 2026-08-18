package apis

import (
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/apps/kecore/services/deployhandle"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/metrics"
)

// RegistryRouter 注册路由
func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("knowledge")
	eng.G("knowledge.metrics", metrics.GetHandler())
	eng.AuthInject("yygu", accountmds.InjectLoginStatus)

	eng.P("knowledge.SwitchPrivateEvn", deployhandle.SwitchPrivateEvn)
	eng.P("knowledge.NowDeployMode", deployhandle.NowDeployMode)
}
