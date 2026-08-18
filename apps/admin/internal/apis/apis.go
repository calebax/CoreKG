package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/apps/admin/adminmds"
	"github.com/insmtx/corekg/apps/admin/internal/apis/companyctl"
	"github.com/insmtx/corekg/apps/admin/internal/apis/dashboard"
	"github.com/insmtx/corekg/apps/admin/internal/apis/license"
	"github.com/insmtx/corekg/apps/admin/internal/apis/lkxctl"
	"github.com/insmtx/corekg/apps/admin/internal/apis/loginctl"
	"github.com/insmtx/corekg/apps/admin/internal/apis/promptctl"
	"github.com/insmtx/corekg/apps/admin/internal/apis/userctl"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime/middleware"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

// RegistryRouter .
func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("admin")
	eng.AuthInject(global.IssuerYYGUAdmin, adminmds.InjectEmployeeLoginStatus)
	eng.AuthInject("", accountmds.InjectLoginStatus)

	eng.PRequireEmployee("admin.ModifyMyUserInfo", ModifyMyUserInfo)
	eng.PRequireEmployee("admin. ", ChangeMyWechat)
	eng.PRequireEmployee("admin.GetMyAction", GetMyAction)
	// 员工
	eng.PRequireEmployee("admin.ListEmployee", adminmds.RequireOpPrivilege, ListEmployee)
	eng.PRequireEmployee("admin.GetEmployeeDetail", adminmds.RequireOpPrivilege, GetEmployeeDetail)
	eng.PRequireEmployee("admin.CreateEmployee", adminmds.RequireOpPrivilege, CreateEmployee)
	eng.PRequireEmployee("admin.ModifyEmployeeInfo", adminmds.RequireOpPrivilege, UpdateEmployee)
	eng.PRequireEmployee("admin.ModifyEmployeePassword", UpdateEmployeePassword)
	eng.PRequireEmployee("admin.DeleteEmployee", adminmds.RequireOpPrivilege, DeleteEmployee)
	eng.PRequireEmployee("admin.GetEmployeeBindKey", adminmds.RequireOpPrivilege, GetEmployeeBindKey)
	eng.P("admin.BindEmployeeWechat", BindEmployeeWechat)

	// 职位权限
	eng.PRequireEmployee("admin.ListPosition", adminmds.RequireOpPrivilege, ListPosition)
	eng.PRequireEmployee("admin.CreatePosition", adminmds.RequireOpPrivilege, CreatePosition)
	eng.PRequireEmployee("admin.GetPositionDetail", adminmds.RequireOpPrivilege, GetPositionDetail)
	eng.PRequireEmployee("admin.ModifyPositionInfo", adminmds.RequireOpPrivilege, ModifyPositionInfo)
	eng.PRequireEmployee("admin.ModifyPositionPrivilege", adminmds.RequireOpPrivilege, ModifyPositionPrivilege)
	eng.PRequireEmployee("admin.DeletePosition", adminmds.RequireOpPrivilege, DeletePosition)

	eng.PRequireEmployee("admin.ListPrivilege", adminmds.RequireOpPrivilege, ListPrivilege)
	eng.PRequireEmployee("admin.CreatePrivilege", adminmds.RequireOpPrivilege, CreatePrivilege)
	eng.PRequireEmployee("admin.ModifyPrivilege", adminmds.RequireOpPrivilege, ModifyPrivilege)
	eng.PRequireEmployee("admin.DeletePrivilege", adminmds.RequireOpPrivilege, DeletePrivilege)

	// 运营配置管理
	eng.PRequireEmployee("admin.ListSetting", adminmds.RequireOpPrivilege, ListSetting)
	eng.PRequireEmployee("admin.CreateSetting", adminmds.RequireOpPrivilege, CreateSetting)
	eng.PRequireEmployee("admin.UpdateSetting", adminmds.RequireOpPrivilege, UpdateSetting)

	// 登录页配置
	eng.PRequireEmployee("admin.ListLoginSetting", adminmds.RequireOpPrivilege, ListLoginSetting)
	eng.PRequireEmployee("admin.CreateLoginSetting", adminmds.RequireOpPrivilege, CreateLoginSetting)
	eng.PRequireEmployee("admin.UpdateLoginSetting", adminmds.RequireOpPrivilege, UpdateLoginSetting)
	eng.PRequireEmployee("admin.DeleteLoginSetting", adminmds.RequireOpPrivilege, DeleteLoginSetting)
	eng.PRequireEmployee("admin.GetLoginSettingByID", adminmds.RequireOpPrivilege, GetLoginSettingByID)

	// API密钥管理
	//eng.PRequireEmployee("admin.ListAPIKey", adminmds.RequireOpPrivilege, ListAPIKey)
	//eng.PRequireEmployee("admin.CreatAPIKey", adminmds.RequireOpPrivilege, CreatAPIKey)
	//eng.PRequireEmployee("admin.DeleteAPIKey", adminmds.RequireOpPrivilege, DeleteAPIKey)
	//eng.PRequireEmployee("admin.ListAPIKeyPrivilege", adminmds.RequireOpPrivilege, ListAPIKeyPrivilege)
	//eng.PRequireEmployee("admin.AddAPIKeyPrivilege", adminmds.RequireOpPrivilege, AddAPIKeyPrivilege)
	//eng.PRequireEmployee("admin.DeleteAPIKeyPrivilege", adminmds.RequireOpPrivilege, DeleteAPIKeyPrivilege)

	//license generate
	eng.PRequireEmployee("admin.GenerateLicense", adminmds.RequireOpPrivilege, license.GenerateLicense)
	eng.PRequireEmployee("admin.ListLicense", adminmds.RequireOpPrivilege, license.ListLicense)
	eng.PRequireEmployee("admin.DownloadLicense", adminmds.RequireOpPrivilege, license.DownloadLicense)

	eng.P("admin.DistributeLicense", accountmds.RequireAPIKeyPrivilege, license.DistributeLicense)
	eng.PRequireEmployee("admin.ApplyLicense", license.ApplyLicense)

	// 上传文件
	eng.P("admin.UploadEmployeeImage", adminmds.RequireOpPrivilege, UploadEmployeeImage)
	eng.P("admin.PanicTest", func(ctx *gin.Context) {
		panic("test")
	})

	// 团队管理
	eng.PRequireEmployee("admin.CreateCompany", adminmds.RequireOpPrivilege, companyctl.CreateCompany)
	eng.PRequireEmployee("admin.ListCompany", adminmds.RequireOpPrivilege, companyctl.ListCompany)
	eng.PRequireEmployee("admin.ModifyCompany", adminmds.RequireOpPrivilege, companyctl.ModifyCompany)
	eng.PRequireEmployee("admin.CreateCompanyEmployee", adminmds.RequireOpPrivilege, companyctl.CreateCompanyEmployee)
	eng.PRequireEmployee("admin.ListCompanyEmployee", adminmds.RequireOpPrivilege, companyctl.ListCompanyEmployee)
	eng.PRequireEmployee("admin.UpdateCompanyEmployeeRole", adminmds.RequireOpPrivilege, companyctl.UpdateCompanyEmployeeRole)
	// 用户管理
	eng.PRequireEmployee("admin.CreateUser", adminmds.RequireOpPrivilege, userctl.CreateUser)
	eng.PRequireEmployee("admin.ListUser", adminmds.RequireOpPrivilege, userctl.ListUser)
	eng.PRequireEmployee("admin.ModifyUser", adminmds.RequireOpPrivilege, userctl.ModifyUser)
	eng.PRequireEmployee("admin.ModifyUserPassword", adminmds.RequireOpPrivilege, userctl.ModifyUserPassword)
	eng.PRequireEmployee("admin.GetUserDetail", adminmds.RequireOpPrivilege, userctl.GetUserDetail)
	// 登录
	eng.P("admin.GetLoginSetting", GetLoginSetting)
	eng.P("admin.LoginByPassword", loginctl.LoginByPassword)
	eng.P("admin.LoginThird", loginctl.LoginThird)
	//数据看板
	eng.PRequireEmployee("admin.GetDashboardData", dashboard.GetDashboardData)

	eng.PRequireEmployee("admin.ListAnnouncement", adminmds.RequireOpPrivilege, ListAnnouncement)
	eng.PRequireEmployee("admin.GetAnnouncement", adminmds.RequireOpPrivilege, GetAnnouncement)
	eng.PRequireEmployee("admin.ModifyAnnouncement", adminmds.RequireOpPrivilege, ModifyAnnouncement)
	eng.PRequireEmployee("admin.CreateAnnouncement", adminmds.RequireOpPrivilege, CreateAnnouncement)
	eng.PRequireEmployee("admin.DeleteAnnouncement", adminmds.RequireOpPrivilege, DeleteAnnouncement)

	eng.PRequireEmployee("admin.ListPaymentOrderRecord", adminmds.RequireOpPrivilege, ListPaymentOrderRecord)
	eng.PRequireEmployee("admin.GetDashboardOverview", adminmds.RequireOpPrivilege, GetDashboardOverview)

	// Prompt管理
	eng.PRequireEmployee("admin.CreatePrompt", adminmds.RequireOpPrivilege, promptctl.CreatePrompt)
	eng.PRequireEmployee("admin.AddPromptVersion", adminmds.RequireOpPrivilege, promptctl.AddPromptVersion)
	eng.PRequireEmployee("admin.SwitchPromptVersion", adminmds.RequireOpPrivilege, promptctl.SwitchPromptVersion)
	eng.PRequireEmployee("admin.GetPromptDetail", adminmds.RequireOpPrivilege, promptctl.GetPromptDetail)
	eng.PRequireEmployee("admin.ListPromptVersions", adminmds.RequireOpPrivilege, promptctl.ListPromptVersions)
	eng.PRequireEmployee("admin.ListPrompts", adminmds.RequireOpPrivilege, promptctl.ListPrompts)
	eng.PRequireEmployee("admin.EditPrompt", adminmds.RequireOpPrivilege, promptctl.EditPrompt)
	eng.PRequireEmployee("admin.DeletePrompt", adminmds.RequireOpPrivilege, promptctl.DeletePrompt)
	eng.PRequireEmployee("admin.RenderPromptPreview", adminmds.RequireOpPrivilege, promptctl.RenderPromptPreview)

	eng.Any("admin.ProxyHTTP", middleware.AuthMiddleWareEmployee, adminmds.RequireOpPrivilege, ProxyHTTP)

	// 丽科星发送验证码
	eng.P("lkxadmin.SendVerifyCode", lkxctl.SendVerifyCode)
	// 丽科星验证验证码并入库
	eng.P("lkxadmin.VerifyCodeAndSave", lkxctl.VerifyCodeAndSave)
}
