package apis

import (
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/apps/corekg/internal/apis/licensectl"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("corekg")
	eng.PRequireLogin("corekg.CheckLicense", accountmds.RequireSysAdminRole, licensectl.CheckLicense)
	eng.P("corekg.GetLicenseInfo", licensectl.GetLicenseInfo)
	eng.PRequireLogin("corekg.RegisterLicense", accountmds.RequireSysAdminRole, licensectl.RegisterLicense)
}
