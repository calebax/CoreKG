package forestctl

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	"github.com/insmtx/corekg/apps/kecore/services/svcforestnode"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

// CreateDir 创建文件夹
func CreateDir(ctx *gin.Context, req *dtokeapi.CreateDirRequest, resp *dtokeapi.CreateDirResponse) {
	if !req.ValidCreateDir(&resp.BaseResponse) {
		return
	}

	fileID, err := svcforestnode.CreateDir(ctx, &svcforestnode.CreateDirRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		ForestID:  req.Request.ForestID,
		ParentID:  req.Request.ParentID,
		Name:      req.Request.Name,
	})
	if err == nil {
		resp.Response.ForestFileID = fileID
		return
	}
	switch {
	case errors.Is(err, svcforestnode.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
	case errors.Is(err, svcforestnode.ErrNoPermission):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission"
	case errors.Is(err, svcforestnode.ErrGetParentNodeFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_parent_node_failed"
	case errors.Is(err, svcforestnode.ErrCheckFileExistsFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_check_file_exists_failed"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_dir_failed"
	}
}

// RenamePath 重命名文件夹
func RenamePath(ctx *gin.Context, req *dtokeapi.RenamePathRequest, resp *apiobj.BaseResponse) {
	if !req.ValidRenamePath(resp) {
		return
	}

	err := svcforestnode.RenamePath(ctx, &svcforestnode.RenamePathRequest{
		Uin:     runtime.Uin(ctx),
		FileID:  req.Request.ForestFileID,
		NewName: req.Request.Name,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforestnode.ErrGetSourceFileFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_source_file_failed"
	case errors.Is(err, svcforestnode.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
	case errors.Is(err, svcforestnode.ErrNoPermission):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission"
	case errors.Is(err, svcforestnode.ErrCheckNewNameFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_check_new_name_failed"
	case errors.Is(err, svcforestnode.ErrNewNameExists):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_new_name_exists"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_file_name_failed"
	}
}

// DeletePath 删除文件夹
func DeletePath(ctx *gin.Context, req *dtokeapi.DeletePathRequest, resp *apiobj.BaseResponse) {
	if !req.ValidDeletePath(resp) {
		return
	}

	err := svcforestnode.DeletePath(ctx, &svcforestnode.DeletePathRequest{
		Uin:     runtime.Uin(ctx),
		FileIDs: req.Request.ForestFileIDs,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforestnode.ErrGetFileOrDirFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_or_dir_failed"
	case errors.Is(err, svcforestnode.ErrUnknownFileList):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_file_list"
	case errors.Is(err, svcforestnode.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
	case errors.Is(err, svcforestnode.ErrNoPermission):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission"
	case errors.Is(err, svcforestnode.ErrTaskRunning):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_task_running"
	case errors.Is(err, svcforestnode.ErrFileStatusCheckFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_file_status_check_failed"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_file_or_dir_failed"
	}
}
