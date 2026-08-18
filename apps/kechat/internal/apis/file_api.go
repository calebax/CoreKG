package apis

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/services/svcfile"
	taskpkg "github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/storage"
)

var (
	ErrFileEmpty         = errors.New("file is empty")
	ErrFileSizeExceeded  = errors.New("file size exceeded")
	ErrInvalidFileFormat = errors.New("invalid file format")
)

type uploadConfig struct {
	AllowedExts []string
	MaxSize     int64
	LogLabel    string
}

// UploadImage 上传图片
// @Summary 上传图片
// @Description 上传图片
// @Router /chat.UploadImage [post]
// @Accept multipart/form-data
// @Param file formData file true "上传图片"
// @Param purpose formData string true "用途，固定值为 yg-chat"
// @Success 200 {object} UploadFileResponse "返回值"
func UploadImage(ctx *gin.Context) {
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	lang := runtime.GetLanguage(ctx)
	purpose := ctx.Request.FormValue("purpose")

	if uin == 0 {
		logs.ErrorContextf(ctx, "UploadImage error uin=0")
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_invalid_request")) // 非法请求
		return
	}

	if purpose != "yg-chat" {
		logs.ErrorContextf(ctx, "UploadImage error: %s", purpose)
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_invalid_request")) // 非法请求
		return
	}

	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "UploadImage error: %v", err)
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_invalid_parameters")) // 参数错误
		return
	}
	defer f.Close()

	resp, err := uploadFile(ctx, uin, companyID, purpose, f, fh, uploadConfig{
		AllowedExts: []string{".jpg", ".jpeg", ".png"},
		MaxSize:     int64(5 << 20),
		LogLabel:    "upload image",
	})
	if err != nil {
		handleUploadError(ctx, err, lang)
		return
	}

	ctx.JSON(200, &UploadFileResponse{Response: *resp})
}

// UploadAttachment 上传解析附件
// @Summary 上传解析附件
// @Description 上传解析附件
// @Router /chat.UploadAttachment [post]
// @Accept multipart/form-data
// @Param file formData file true "上传附件"
// @Param purpose formData string true "用途，固定值为 yg-chat"
// @Success 200 {object} UploadFileResponse "返回值"
func UploadAttachment(ctx *gin.Context) {
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	lang := runtime.GetLanguage(ctx)
	purpose := ctx.Request.FormValue("purpose")

	if uin == 0 {
		logs.ErrorContextf(ctx, "UploadAttachment error uin=0")
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_invalid_request")) // 非法请求
		return
	}

	if purpose != "yg-chat" {
		logs.ErrorContextf(ctx, "UploadAttachment error: %s", purpose)
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_invalid_request")) // 非法请求
		return
	}

	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "UploadAttachment error: %v", err)
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_invalid_parameters")) // 参数错误
		return
	}
	defer f.Close()

	resp, err := uploadFile(ctx, uin, companyID, purpose, f, fh, uploadConfig{
		AllowedExts: []string{".pdf", ".txt", ".doc", ".docx", ".ofd", ".ppt", ".pptx", ".md", ".log", ".json", ".csv", ".png", ".jpg", ".jpeg", ".mp4", ".mov"},
		MaxSize:     int64(500 << 20),
		LogLabel:    "UploadAttachment",
	})
	if err != nil {
		handleUploadError(ctx, err, lang)
		return
	}

	parseResult, err := svcfile.ParseToMarkdown(ctx, &svcfile.ParseRequest{
		SourceID:  resp.FileID,
		SourceURL: resp.URL,
		Name:      fh.Filename,
		Purpose:   purpose,
		File:      f,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[UploadAttachment] parse attachment failed, fileID: %d, err: %v", resp.FileID, err)
		runtime.InternalError(ctx, i18n.T(lang, "kechat_internal_error"))
		return
	}
	resp.TaskID = parseResult.TaskID
	resp.ParseStatus = parseResult.Status
	resp.MDURL = parseResult.URL

	ctx.JSON(200, &UploadFileResponse{Response: *resp})
}

func handleUploadError(ctx *gin.Context, err error, lang string) {
	if errors.Is(err, ErrFileEmpty) {
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_empty_file")) // 文件为空
	} else if errors.Is(err, ErrFileSizeExceeded) {
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_file_size_exceeded")) // 文件大小超过限制
	} else if errors.Is(err, ErrInvalidFileFormat) {
		runtime.BadRequest(ctx, i18n.T(lang, "kechat_unsupported_file_format")) // 不支持的文件格式
	} else {
		runtime.InternalError(ctx, i18n.T(lang, "kechat_internal_error")) // 服务器错误
	}
}

func uploadFile(ctx context.Context, uin, companyID uint, purpose string, f multipart.File, fh *multipart.FileHeader, cfg uploadConfig) (*FileInfo, error) {
	logLabel := cfg.LogLabel
	if logLabel == "" {
		logLabel = "upload file"
	}

	if fh.Size == 0 {
		logs.WarnContextf(ctx, "%s size is zero: %d bytes", logLabel, fh.Size)
		return nil, ErrFileEmpty
	}

	if cfg.MaxSize > 0 && fh.Size > cfg.MaxSize {
		logs.WarnContextf(ctx, "%s size exceeds the limit: %d bytes", logLabel, fh.Size)
		return nil, ErrFileSizeExceeded
	}

	fileInfo := &storage.FileInfo{
		Uin:       uin,
		CompanyID: companyID,
		Filename:  fh.Filename,
		Size:      fh.Size,
		Purpose:   purpose,
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	fileInfo.FileExt = ext

	if len(cfg.AllowedExts) > 0 {
		validExt := false
		for _, allowedExt := range cfg.AllowedExts {
			if ext == allowedExt {
				validExt = true
				break
			}
		}

		if !validExt {
			logs.ErrorContextf(ctx, "%s format error: %s", logLabel, ext)

			return nil, ErrInvalidFileFormat
		}
	}

	fileInfo.StoragePath = storage.GenerateFileStoragePath(purpose, uin, ext)

	st, err := storage.LoadStorager(purpose)
	if err != nil {
		logs.ErrorContextf(ctx, "%s error: %v", logLabel, err)
		return nil, err
	}

	if err = st.Save(ctx, fileInfo, f); err != nil {
		logs.ErrorContextf(ctx, "%s error: %v", logLabel, err)
		return nil, err
	}

	if err := dbutil.Core().Create(fileInfo).Error; err != nil {
		logs.ErrorContextf(ctx, "%s error: %v", logLabel, err)
		return nil, err
	}

	fileInfo.PublicURL = st.GetPublicURL(fileInfo.StoragePath, false)

	return &FileInfo{
		FileID:   fileInfo.ID,
		URL:      fileInfo.PublicURL,
		Uin:      uin,
		Filename: fileInfo.Filename,
	}, nil
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	apiobj.BaseResponse
	Response FileInfo
}

// FileInfo 文件信息
type FileInfo struct {
	// FileID 表示上传文件在 core_upload_files 中的 ID。
	FileID uint `json:"file_id,omitempty"`
	// Uin 表示上传文件所属用户的 UIN。
	Uin uint `json:"uin"`
	// TaskID 表示附件解析任务的 core_task ID。
	TaskID uint `json:"task_id,omitempty"`
	// ParseStatus 表示附件解析任务在 core_task 中的任务状态。
	ParseStatus taskpkg.TaskStatus `json:"parse_status,omitempty"`
	// URL 表示上传文件的公开访问地址。
	URL string `json:"url,omitempty"`
	// MDURL 表示附件解析后生成的 Markdown 文件公开访问地址。
	MDURL string `json:"md_url,omitempty"`
	// Filename 原始文件名
	Filename string `json:"filename,omitempty"`
}
