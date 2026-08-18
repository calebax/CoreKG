package apis

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/storage"
)

// UploadCustomerImage 上位管理员图片
// @Tags Upload
// @Summary 上位管理员图片
// @Description 上位管理员图片
// @Router /admin.UploadCustomerImage [post]
// @Accept multipart/form-data
// @Param file formData file true "图片文件"
// @Success 200 {object} UploadImageResponse "返回值"
func UploadCustomerImage(ctx *gin.Context) {
	var (
		logger = runtime.Logger(ctx)
	)
	purpose := ctx.Request.FormValue("purpose")
	// 白名单
	if purpose != "login-bgd" && purpose != "cu-image" {
		return
	}
	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.BadRequest(ctx, "参数错误")
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

	fi.StoragePath = "/" + storage.GenerateFileStoragePath(purpose, fi.Uin, ext)

	st, err := storage.LoadStorager(purpose)

	if err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.InternalError(ctx, "服务器错误")
		return
	}
	err = st.Save(ctx, fi, f)
	if err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.InternalError(ctx, "服务器错误")
		return
	}
	fi.PublicURL = st.GetPublicURL(fi.StoragePath, false)

	if err := dbutil.Core().Create(fi).Error; err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.InternalError(ctx, "服务器错误")
		return
	}

	resp := &UploadImageResponse{
		Response: FileInfo{
			FileID: fi.ID,
			URL:    fi.PublicURL,
		},
	}
	ctx.JSON(200, resp)
}

// UploadEmployeeImage 上位用户图片
// @Tags Upload
// @Summary 上位用户图片
// @Description 上位用户图片
// @Router /admin.UploadEmployeeImage [post]
// @Accept multipart/form-data
// @Param file formData file true "图片文件"
// @Success 200 {object} UploadImageResponse "返回值"
func UploadEmployeeImage(ctx *gin.Context) {
	var (
		logger = runtime.Logger(ctx)
	)
	purpose := ctx.Request.FormValue("purpose")
	// 白名单
	if purpose != "em-image" {
		return
	}
	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.BadRequest(ctx, "参数错误")
		return
	}
	defer f.Close()

	fi := &storage.FileInfo{
		CompanyID: 1,
		Uin:       runtime.Uin(ctx),
		Filename:  fh.Filename,
		Size:      fh.Size,
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	fi.FileExt = ext

	fi.StoragePath = "/" + storage.GenerateFileStoragePath(purpose, fi.Uin, ext)

	st, err := storage.LoadStorager(purpose)
	if err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.InternalError(ctx, "服务器错误")
		return
	}
	err = st.Save(ctx, fi, f)
	if err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.InternalError(ctx, "服务器错误")
		return
	}
	fi.PublicURL = st.GetPublicURL(fi.StoragePath, false)

	if err := dbutil.Core().Create(fi).Error; err != nil {
		logger.Errorf("upload image error: %v", err)
		runtime.InternalError(ctx, "服务器错误")
		return
	}

	resp := &UploadImageResponse{
		Response: FileInfo{
			FileID: fi.ID,
			URL:    fi.PublicURL,
		},
	}
	ctx.JSON(200, resp)
}
