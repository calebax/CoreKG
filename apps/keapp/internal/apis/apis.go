package apis

import (
	"github.com/insmtx/corekg/apps/keapp/internal/apis/appctl"
	"github.com/insmtx/corekg/apps/keapp/internal/apis/webctl"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("keapp")

	appctl.RegistryRouter(eng)
	webctl.RegistryRouter(eng)
}
