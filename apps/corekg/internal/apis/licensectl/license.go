package licensectl

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/license"
	"github.com/insmtx/corekg/apps/corekg/internal/jobs"
	"github.com/insmtx/corekg/apps/corekg/mds"
	corekglicense "github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

// CheckLicense 检查license状态
// @Tags license自助
// @Summary 检查license状态
// @Description 检查license状态
// @Router /corekg.CheckLicense [post]
// @Param user body apiobj.BaseRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func CheckLicense(ctx *gin.Context, _ *apiobj.BaseRequest, _ *apiobj.BaseResponse) {
	if err := mds.CheckValidLogEntry(ctx); err != nil {
		logs.ErrorContextf(ctx, "CheckLicense error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "corekg_license_check_failed")) // License认证检查失败
		return
	}
}

// RegisterLicense 注册license
// @Tags license自助
// @Summary 注册license
// @Description 注册license
// @Router /corekg.RegisterLicense [post]
// @Param user body RegisterLicenseRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func RegisterLicense(ctx *gin.Context, req *RegisterLicenseRequest, resp *apiobj.BaseResponse) {
	if req.Valid(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "RegisterLicense param valid error: %v", resp.Message)
		return
	}
	if err := settings.UpsertSetting(&settings.SettingItem{
		Group:     "corekg",
		Key:       "raw_license",
		ValueType: settings.ValueText,
		Describe:  fmt.Sprintf("license_%v", time.Now().UTC()),
		Value:     req.Request.License,
	}); err != nil {
		logs.ErrorContextf(ctx, "Set corekg license error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "corekg_license_register_failed")) // License注册失败
		return
	}

	// 1. 从环境变量或配置中获取当前环境类型
	// e.g., "kubernetes"
	envTypeStr := os.Getenv(jobs.DeploymentEnv)

	// 2. 使用工厂创建对应的 Environment 实例
	env, err := corekglicense.NewEnvironment(corekglicense.EnvType(envTypeStr))
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to create environment: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "corekg_license_env_check_failed")) // License环境检查失败
		return
	}

	// 3. 创建校验器 Checker
	checker := license.NewChecker(dbutil.Core(), env)

	// 4. 执行检查
	status, msg := checker.PureCheck(ctx)
	if status != corekglicense.StatusValid {
		logs.ErrorContextf(ctx, "License pure check status [%v] error: %v", status, msg)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "corekg_license_check_failed")) // License检查失败
		return
	}
	// 5. 清除旧记录
	if err = dbutil.Core().
		Where("deleted_at IS NULL").
		Delete(&admintype.DailyLog{}).
		Error; err != nil {
		logs.ErrorContextf(ctx, "License delete error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "corekg_license_migration_failed")) // License迁移失败
		return
	}
	// 6. 初始化哈希
	checker.PerformCheck(ctx)

	// 7. 更新公司信息
	var company accounttype.Company
	companyID := runtime.CompanyID(ctx)
	if err := dbutil.Account().WithContext(ctx).Table(accounttype.TableNameCompany).First(&company, "id = ?", companyID).Error; err != nil {
		logs.ErrorContextf(ctx, "Failed to get company: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_company_info_get_failed")) // 公司信息获取失败
		return
	}
	company.Name = req.Request.CompanyName
	company.Logo = req.Request.CompanyLogo
	if err := dbutil.Account().WithContext(ctx).Save(&company).Error; err != nil {
		logs.ErrorContextf(ctx, "Failed to update company: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_company_info_update_failed")) // 公司信息更新失败
		return
	}

	// 8. 更新网站logo
	if err := settings.SetYaml("core", "website-info", req.Request.WebsiteInfo); err != nil {
		logs.ErrorContextf(ctx, "Failed to update website info: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_website_info_update_failed")) // 网站信息更新失败
		return
	}
}

// GetLicenseInfo 获取license元信息
// @Tags license自助
// @Summary 获取license元信息
// @Description 获取license元信息
// @Router /corekg.GetLicenseInfo [post]
// @Param user body RegisterLicenseRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func GetLicenseInfo(ctx *gin.Context, _ *apiobj.BaseRequest, resp *GetLicenseInfoResponse) {
	// 1. 从环境变量或配置中获取当前环境类型
	// e.g., "kubernetes"
	envTypeStr := os.Getenv(jobs.DeploymentEnv)

	// 2. 使用工厂创建对应的 Environment 实例
	env, err := corekglicense.NewEnvironment(corekglicense.EnvType(envTypeStr))
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to create environment: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "corekg_license_env_check_failed")) // License环境检查失败
		return
	}

	// 3. 创建校验器 Checker
	checker := license.NewChecker(dbutil.Core(), env)

	// 4. 执行检查并返回原信息
	meta, status := checker.Meta(ctx)
	if meta != nil {
		logs.DebugContextf(ctx, "License meta: %v", *meta)
		resp.Response.Meta = meta
		resp.Response.ValidDays = int(meta.ExpiredAt.Sub(time.Now().UTC()).Hours() / 24)

		if len(meta.VersionKey) <= 0 {
			resp.Response.Modules = global.VersionKeyMap["all"]
		} else {
			resp.Response.Modules = global.VersionKeyMap[meta.VersionKey]
		}
	}

	// 5. 获取网站信息
	var websiteInfo WebsiteInfo
	if err := settings.GetYaml("core", "website-info", &websiteInfo); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logs.ErrorContextf(ctx, "Failed to get website info: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_website_info_get_failed")) // 网站信息获取失败
			return
		}
		logs.WarnContextf(ctx, "Failed to get website info: %v", err)
	}
	resp.Response.WebsiteInfo = websiteInfo

	// 6. 返回状态
	resp.Response.Status = status
}
