package forestctl

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kecore/services/svcforest"
	"github.com/insmtx/corekg/apps/kecore/services/svcforestfile"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/storage"
	"gorm.io/gorm"
)

// ListFile 文档列表
func ListFile(ctx *gin.Context, req *dtokeapi.ListFileRequest, resp *dtokeapi.ListFileResponse) {
	if !req.ValidListFile(&resp.BaseResponse) {
		return
	}

	out, err := svcforestfile.ListFile(ctx, &svcforestfile.ListFileRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		ForestID:  req.Request.ForestID,
		PageQuery: req.Request.PageQuery,
	})
	if err != nil {
		switch {
		case errors.Is(err, svcforestfile.ErrUserIDEmpty):
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_user_id_empty"
		default:
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_forest_list_failed"
		}
		return
	}

	resp.BaseResponse = out.BaseResponse
	resp.Response.Total = out.Response.Total
	resp.Response.Offset = req.Request.Offset
	resp.Response.Limit = req.Request.Limit
	resp.Response.Data = make([]*dtokeapi.ForestFileItem, 0, len(out.Response.Data))
	for _, item := range out.Response.Data {
		resp.Response.Data = append(resp.Response.Data, dtokeapi.NewForestFileItem(item))
	}
}

// BatchGetFile 批量查询文档信息
func BatchGetFile(ctx *gin.Context, req *dtokeapi.BatchGetFileRequest, resp *dtokeapi.BatchGetFileResponse) {
	if !req.ValidBatchGetFile(&resp.BaseResponse) {
		return
	}

	filterValues := make([]string, 0, len(req.Request.ForestFileIDs))
	for _, forestFileID := range req.Request.ForestFileIDs {
		if forestFileID == 0 {
			continue
		}
		filterValues = append(filterValues, strconv.FormatUint(uint64(forestFileID), 10))
	}
	if len(filterValues) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_file_ids"
		return
	}

	out, err := svcforestfile.ListFile(ctx, &svcforestfile.ListFileRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		PageQuery: apiobj.PageQuery{
			ListAll: true,
			Filters: []apiobj.Filter{{
				Field: "id",
				Value: filterValues,
			}},
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, svcforestfile.ErrUserIDEmpty):
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_user_id_empty"
		default:
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_forest_list_failed"
		}
		return
	}

	fileMap := make(map[uint]*forest.File, len(out.Response.Data))
	forestIDSet := make(map[uint]struct{}, len(out.Response.Data))
	for _, item := range out.Response.Data {
		fileMap[item.ID] = item
		forestIDSet[item.ForestID] = struct{}{}
	}

	forestFilterValues := make([]string, 0, len(forestIDSet))
	for forestID := range forestIDSet {
		forestFilterValues = append(forestFilterValues, strconv.FormatUint(uint64(forestID), 10))
	}

	allowedForestIDs := make(map[uint]struct{}, len(forestFilterValues))
	if len(forestFilterValues) > 0 {
		forestOut, err := svcforest.ListForest(ctx, &svcforest.ListForestRequest{
			Uin:       runtime.Uin(ctx),
			CompanyID: runtime.CompanyID(ctx),
			Query: apiobj.PageQuery{
				ListAll: true,
				Filters: []apiobj.Filter{{
					Field: "id",
					Value: forestFilterValues,
				}},
			},
			PresetWhenListing: false,
		})
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_forest_list_failed"
			return
		}
		for _, item := range forestOut.Data {
			allowedForestIDs[item.ID] = struct{}{}
		}
	}

	resp.Response.Offset = 0
	resp.Response.Limit = len(req.Request.ForestFileIDs)
	resp.Response.Data = make([]*dtokeapi.ForestFileItem, 0, len(req.Request.ForestFileIDs))
	for _, forestFileID := range req.Request.ForestFileIDs {
		item, ok := fileMap[forestFileID]
		if !ok {
			continue
		}
		if _, ok := allowedForestIDs[item.ForestID]; !ok {
			continue
		}
		resp.Response.Data = append(resp.Response.Data, dtokeapi.NewForestFileItem(item))
	}
	resp.Response.Total = int64(len(resp.Response.Data))
}

// UploadFile 上传文件
func UploadFile(ctx *gin.Context) {
	resp := &dtokeapi.UploadFileResponse{}
	uin := runtime.Uin(ctx)

	// keapi 进程内懒加载知识库文件存储器，供后续服务端直传使用。
	if err := ensureForestStorage(); err != nil {
		logs.ErrorContextf(ctx, "keapi: ensure forest storage failed, err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_upload_storager_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	forestID, err := strconv.Atoi(ctx.Request.FormValue("forest_id"))
	if err != nil || forestID <= 0 {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_invalid_forest_id"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	forestInfo, err := forest.GetForestByID(ctx, uint(forestID))
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	if !perm.HasManageAct(ctx, uin, forestInfo.ID, foresttype.ResourceTypeForest) {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// 读取外部上传文件，并先计算完整文件哈希，供 kecore 新上传链路复用。
	parentID, err := parseParentID(ctx)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_invalid_parent_id"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	fileReader, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "keapi: get upload file failed, err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_upload_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	defer fileReader.Close()

	hashValue, err := calcUploadHash(fileReader)
	if err != nil {
		logs.ErrorContextf(ctx, "keapi: calc upload hash failed, err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_upload_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// 通过 bridge 调用 kecore 的预上传逻辑，生成上传记录。
	files, errInfo := svcforestfile.PreUploadFileForAPI(ctx, &svcforestfile.APIPreUploadFileRequest{
		ForestID: uint(forestID),
		ParentID: parentID,
		Files: []svcforestfile.APIUploadFileItem{{
			Filename:    fileHeader.Filename,
			Hash:        hashValue,
			Size:        fileHeader.Size,
			SplitConfig: defaultSplitConfig(),
			ContentType: fileHeader.Header.Get("Content-Type"),
		}},
	})
	if errInfo != nil {
		logs.ErrorContextf(ctx, "keapi: pre upload file failed, err = %+v", errInfo)
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	if len(files) != 1 {
		logs.ErrorContextf(ctx, "keapi: unexpected pre upload response count = %d", len(files))
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_file_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	preUpload := files[0]
	coreFile, err := storage.GetFileByID(nilSafeCoreDB(ctx), preUpload.UploadID)
	if err != nil {
		logs.ErrorContextf(ctx, "keapi: get core upload file failed, upload_id = %d, err = %v", preUpload.UploadID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_info_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// 命中秒传时跳过文件写入；否则将上传记录切换为单文件直传并直接写入目标存储路径。
	if !preUpload.Exists {
		coreFile, errInfo = svcforestfile.PrepareCoreFileForDirectUpload(ctx, preUpload.UploadID)
		if errInfo != nil {
			logs.ErrorContextf(ctx, "keapi: prepare direct upload failed, err = %+v", errInfo)
			resp.Code = errInfo.Code
			resp.Message = errInfo.Message
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}

		if _, err := fileReader.Seek(0, io.SeekStart); err != nil {
			logs.ErrorContextf(ctx, "keapi: reset upload reader failed, err = %v", err)
			markCoreUploadFailed(ctx, coreFile.ID)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_upload_failed"
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}

		if err := fs.Forest.Save(ctx, coreFile, fileReader); err != nil {
			logs.ErrorContextf(ctx, "keapi: direct save upload file failed, err = %v", err)
			markCoreUploadFailed(ctx, coreFile.ID)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_upload_failed"
			ctx.JSON(http.StatusInternalServerError, resp)
			return
		}
	}

	// 文件写入完成后，继续走 kecore 的上传完成回调，落知识库文件和后续任务。
	fileID, errInfo := svcforestfile.UploadFileCallBackForAPI(ctx, &svcforestfile.APIUploadFileCallBackRequest{
		ForestID: uint(forestID),
		UploadID: preUpload.UploadID,
		Filename: coreFile.Filename,
		Hash:     coreFile.Hash,
	})
	if errInfo != nil {
		logs.ErrorContextf(ctx, "keapi: upload callback failed, err = %+v", errInfo)
		markCoreUploadFailed(ctx, coreFile.ID)
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	if fileID == nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_file_failed"
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Message = "kecore_upload_success"
	resp.Response.ForestFileID = *fileID
	ctx.JSON(http.StatusOK, resp)
}

// PreviewFileByURL 下载文档
func PreviewFileByURL(ctx *gin.Context, req *dtokeapi.PreviewFileByURLRequest, resp *dtokeapi.PreviewFileByURLResponse) {
	if !req.ValidPreviewFile(resp) {
		return
	}

	url, err := svcforestfile.PreviewFileByURL(ctx, &svcforestfile.PreviewFileByURLRequest{
		FileID:     req.Request.ForestFileID,
		IsDownload: req.Request.IsDownload,
		Referer:    ctx.GetHeader("Referer"),
	})
	if err != nil {
		switch {
		case errors.Is(err, svcforestfile.ErrGetSourceFileFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_source_file_failed"
		case errors.Is(err, svcforestfile.ErrFileNotPreviewable):
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_file_not_previewable"
		case errors.Is(err, svcforestfile.ErrGetFilePathFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_file_path_failed"
		case errors.Is(err, svcforestfile.ErrGetUploadConfigFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_upload_config_failed"
		case errors.Is(err, svcforestfile.ErrParseURLFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_parse_url_failed"
		case errors.Is(err, svcforestfile.ErrCreateStorageFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_create_storage_failed"
		case errors.Is(err, svcforestfile.ErrGetPresignedURLFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_presigned_url_failed"
		default:
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_file_url_failed"
		}
		return
	}
	resp.Response.URL = url
}

func ensureForestStorage() error {
	if fs.Forest != nil {
		return nil
	}
	return fs.InitForestStorage()
}

func parseParentID(ctx *gin.Context) (uint, error) {
	parentIDStr := ctx.Request.FormValue("parent_id")
	if parentIDStr == "" {
		return 0, nil
	}

	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil || parentID < 0 {
		return 0, fmt.Errorf("invalid parent_id: %s", parentIDStr)
	}
	return uint(parentID), nil
}

func calcUploadHash(file io.ReadSeeker) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return "sha256:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

func defaultSplitConfig() *ragtask.SplitConfig {
	return &ragtask.SplitConfig{
		SplitMode:    ragtask.SplitAuto,
		ChunkSize:    500,
		SplitMark:    []string{"\n\n", "\n", "。", "；", ";", "，", ",", " "},
		SplitOverlap: 0.2,
	}
}

func markCoreUploadFailed(ctx *gin.Context, uploadID uint) {
	if uploadID == 0 {
		return
	}
	if err := storage.UpdateByID(nilSafeCoreDB(ctx), uploadID, map[string]interface{}{"status": storage.FileStatusFailed}); err != nil {
		logs.ErrorContextf(ctx, "keapi: mark core upload failed, upload_id = %d, err = %v", uploadID, err)
	}
}

func nilSafeCoreDB(ctx *gin.Context) *gorm.DB {
	return dbutil.Core().WithContext(ctx)
}
