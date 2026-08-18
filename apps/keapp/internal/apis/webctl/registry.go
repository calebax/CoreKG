package webctl

import (
	"github.com/insmtx/corekg/apps/keapp/internal/middleware"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func RegistryRouter(eng *server.Router) {
	eng.PRequireLogin("keapp.web.ListResources", middleware.AppContextMiddleware(), middleware.RequireAppViewPerm, ListResources)
	eng.PRequireLogin("keapp.web.GetResource", middleware.AppContextMiddleware(), middleware.RequireAppViewPerm, GetResource)
	eng.PRequireLogin("keapp.web.DeleteResource", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, DeleteResource)
	eng.PRequireLogin("keapp.web.RecrawlResource", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, RecrawlResource)

	eng.PRequireLogin("keapp.web.AddCrawlRule", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, AddCrawlRule)
	eng.PRequireLogin("keapp.web.ListCrawlRules", middleware.AppContextMiddleware(), middleware.RequireAppViewPerm, ListCrawlRules)
	eng.PRequireLogin("keapp.web.UpdateCrawlRule", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, UpdateCrawlRule)
	eng.PRequireLogin("keapp.web.DeleteCrawlRule", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, DeleteCrawlRule)

	eng.PRequireLogin("keapp.web.TriggerCrawl", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, TriggerCrawl)
	eng.PRequireLogin("keapp.web.GetCrawlTask", middleware.AppContextMiddleware(), middleware.RequireAppViewPerm, GetCrawlTask)
	eng.PRequireLogin("keapp.web.ListCrawlTasks", middleware.AppContextMiddleware(), middleware.RequireAppViewPerm, ListCrawlTasks)
	eng.PRequireLogin("keapp.web.CancelCrawlTask", middleware.AppContextMiddleware(), middleware.RequireAppManagePerm, CancelCrawlTask)
}
