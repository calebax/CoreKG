package apis

import (
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/apps/corekg/internal/apis/licensectl"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("corekg")
	eng.PRequireLogin("corekg.CheckLicense", accountmds.RequireSysAdminRole, licensectl.CheckLicense)
	eng.P("corekg.GetLicenseInfo", licensectl.GetLicenseInfo)
	eng.PRequireLogin("corekg.RegisterLicense", accountmds.RequireSysAdminRole, licensectl.RegisterLicense)

	// HTTP 任务接入端点：供 pipeline / clients.task_worker 等以 HTTP 轮询拉取与回报任务。
	eng.P("knowledge.GetPendingTask", task.GetPendingTask)
	eng.P("knowledge.TaskCallBack", task.TaskCallBack)
}
