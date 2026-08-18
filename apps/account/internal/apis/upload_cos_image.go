package apis

import (
	"fmt"
	"net/http"
	stdurl "net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
)

// UploadCustomerImage 上传管理员图片
// @Tags Upload
// @Summary 上传管理员图片
// @Description 上传管理员图片
// @Router /account.UploadCustomerImage [post]
// @Accept multipart/form-data
// @Param file formData file true "图片文件"
// @Success 200 {object} UploadImageResponse "返回值"
func UploadCustomerImage(ctx *gin.Context) {
	var (
		resp = &UploadImageResponse{
			Response: FileInfo{},
		}
	)
	purpose := ctx.Request.FormValue("purpose")
	// 白名单
	if purpose != "login-bgd" && purpose != "cu-image" {
		logs.WarnContextf(ctx, "upload image: invalid purpose: %s", purpose)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_purpose")) // 无效的上传目的
		return
	}
	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_parameters")) // 参数错误
		return
	}
	defer f.Close()

	fi := &storage.FileInfo{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       runtime.Uin(ctx),
		Filename:  fh.Filename,
		Size:      fh.Size,
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	fi.FileExt = ext

	fi.StoragePath = storage.GenerateFileStoragePath(purpose, fi.Uin, ext)

	st, err := storage.LoadStorager(purpose)
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_storage_load_failed")) // 加载存储服务失败
		return
	}
	if err = st.Save(ctx, fi, f); err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_storage_save_failed")) // 保存文件失败
		return
	}
	fi.PublicURL = st.GetPublicURL(fi.StoragePath, false)
	//deploy mode switch
	if version.DeployMode() != "" && version.DeployMode() != global.DeployModeOpenPO {
		var (
			cfg config.StorageConfig
		)
		if err = settings.GetYaml(settings.SettingGroupCore, storage.SettingPrefix+purpose, &cfg); err != nil {
			logs.ErrorContextf(ctx, "get storage config [group:%v|key:%v] error: %v", settings.SettingGroupCore, storage.SettingPrefix+purpose, err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_upload_config_failed")) // 获取上传配置失败
			return
		}
		referer := ctx.GetHeader("Referer")
		// 解析 URL
		parsedURL, err := stdurl.Parse(referer)
		if err != nil {
			logs.ErrorContextf(ctx, "get referer error: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_parse_url_failed")) // 解析url失败
			return
		}

		cfg.S3.EndPoint = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		st, err := storage.NewStorageWithCfg(cfg)
		if err != nil {
			logs.ErrorContextf(ctx, "[PreviewFileByURL] new storage error: %v cfg[+%v]", err, cfg)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_create_storage_failed")) // 创建存储器失败
			return
		}
		url := st.GetPublicURL(fi.StoragePath, false)

		fi.PublicURL = url
		resp.Response.URL = url
	}
	if err := dbutil.Core().Create(fi).Error; err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_database_save_failed")) // 数据库保存失败
		return
	}

	resp = &UploadImageResponse{
		Response: FileInfo{
			FileID: fi.ID,
			URL:    fi.PublicURL,
		},
	}

	ctx.JSON(http.StatusOK, resp)
}
