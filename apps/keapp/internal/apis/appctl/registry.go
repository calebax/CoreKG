package appctl

import (
	"github.com/insmtx/corekg/apps/keapp/internal/middleware"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func RegistryRouter(eng *server.Router) {
	eng.PRequireLogin("keapp.CreateApplication", CreateApplication)
	eng.PRequireLogin("keapp.ListApplications", ListApplications)
	eng.PRequireLogin("keapp.GetApplication", middleware.RequireAppViewPerm, GetApplication)
	eng.PRequireLogin("keapp.UpdateApplication", middleware.RequireAppManagePerm, UpdateApplication)
	eng.PRequireLogin("keapp.DeleteApplication", middleware.RequireAppManagePerm, DeleteApplication)
}
