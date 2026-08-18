package forestctl

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/dsnet/golib/memfile"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoforest"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kecore/services/svcforest"
	"github.com/insmtx/corekg/apps/kecore/services/svcforestfile"
	"github.com/insmtx/corekg/apps/kecore/services/svcforestnode"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/metrics"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

// CreateForest 创建知识森林
// @Tags 知识森林
// @Summary 创建知识森林
// @Description 创建知识森林
// @Router /forest.CreateForest [post]
// @Param user body CreateForestRequest true "入参"
// @Success 200 {object} CreateForestResponse "返回值"
func CreateForest(ctx *gin.Context, req *CreateForestRequest, resp *CreateForestResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	forestID, err := svcforest.CreateForest(ctx, &svcforest.CreateForestRequest{
		Uin:               runtime.Uin(ctx),
		CompanyID:         runtime.CompanyID(ctx),
		Name:              req.Request.Name,
		AvatarURL:         req.Request.AvatarUrl,
		Description:       req.Request.Decription,
		PublicScope:       req.Request.PublicScope,
		ForestType:        req.Request.ForestType,
		DataSourceType:    req.Request.DataSourceType,
		DataSourceSubtype: req.Request.DataSourceSubtype,
	})
	if err != nil {
		if errors.Is(err, svcforest.ErrForestNameExists) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_forest_name_exists" // 名称已存在
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_forest_failed" // 创建失败
		return
	}

	resp.Response.ForestID = forestID
}

// ListForest 知识森林列表
// @Tags 知识森林
// @Summary 知识森林列表
// @Description 知识森林列表
// @Router /forest.ListForest [post]
// @Param user body ListForestRequest true "入参"
// @Success 200 {object} ListForestResponse "返回值"
func ListForest(ctx *gin.Context, req *ListForestRequest, resp *ListForestResponse) {
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	req.Request.Uin = uin
	req.Request.CompanyID = companyID
	// TODO: 指标上报测试，要下掉
	metrics.Counter("forest_list_count").
		With(prometheus.Labels{
			"company_id": strconv.Itoa(int(companyID)),
			"uin":        strconv.Itoa(int(uin)),
		}).Add(float64(1))
	logs.InfoContextf(ctx, "[ListForest] uin = %v", uin)
	if req.Validity(&resp.BaseResponse); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListForest validate params failed")
		return
	}

	out, err := svcforest.ListForest(ctx, &svcforest.ListForestRequest{
		Uin:               uin,
		CompanyID:         companyID,
		Query:             req.Request,
		PresetWhenListing: true,
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_list_failed" // 查询知识森林列表失败
		return
	}
	resp.Response = *out
}

// ModifyForest 知识森林编辑
// Deprecated:api route,this func be replaced with accountctl.UpdateForestWithPerm
// @Tags 知识森林
// @Summary 知识森林编辑
// @Description 知识森林编辑
// @Router /forest.ModifyForest [post]
// @Param user body ModifyForestRequest true "入参"
// @Success 200 {object} ModifyForestResponse "返回值"
func ModifyForest(ctx *gin.Context, req *ModifyForestRequest, resp *ModifyForestResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := svcforest.UpdateForest(ctx, &svcforest.UpdateForestRequest{
		ForestID:    req.Request.ID,
		Name:        &req.Request.Name,
		AvatarURL:   &req.Request.AvatarUrl,
		Description: &req.Request.Decription,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforest.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_failed" // 查询知识森林失败
	case errors.Is(err, svcforest.ErrForestNameExists):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_name_exists" // 名称已存在
	case errors.Is(err, svcforest.ErrModifyForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_modify_forest_failed" // 修改失败
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_modify_forest_failed"
	}
}

// GetForest 获取知识森林详情
// @Tags 知识森林
// @Summary 获取知识森林详情
// @Description 获取知识森林详情
// @Router /forest.GetForest [post]
// @Param user body GetForestRequest true "入参"
// @Success 200 {object} GetForestResponse "返回值"
func GetForest(ctx *gin.Context, req *GetForestRequest, resp *GetForestResponse) {
	if req.Request.ID <= 0 {
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_invalid_forest_id")) // 非法id
		return
	}

	frs, err := forest.GetForestByID(ctx, req.Request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_query_forest_failed"
			return
		}
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_forest_failed")) // 获取知识森林失败
		logs.ErrorContextf(ctx, "GetForestByID(%v) faild: %v", req.Request.ID, err)
		return
	}

	var (
		scopeIDs, managerIDs []uint
		rss                  []*foresttype.KeResourceScope
	)

	if err := dbutil.Knownow().
		Where("deleted_at IS NULL").
		Where("resource_type", foresttype.ResourceTypeForest).
		Where("resource_id = ?", frs.ID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Find(&rss).Error; err != nil {
		logs.ErrorContextf(ctx, "GetForestWithPerm failed: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_permission_list_failed")) // 获取权限列表失败
		return
	}

	for _, v := range rss {
		switch v.Action {
		case foresttype.ActionManage:
			managerIDs = append(managerIDs, v.ScopeID)
		case foresttype.ActionView:
			scopeIDs = append(scopeIDs, v.ScopeID)
		}
	}

	var isAdmin bool
	if slices.Contains(managerIDs, runtime.Uin(ctx)) {
		isAdmin = true
	}

	var calcType svcforest.ResourceCalcType
	switch (*frs).ForestType {
	case foresttype.ForestTypeFile:
		calcType = svcforest.ResourceCalcTypeFile
	case foresttype.ForestTypeData:
		switch (*frs).DataSourceType {
		case foresttype.ForestDataSourceDB:
			calcType = svcforest.ResourceCalcTypeMysql
		case foresttype.ForestDataSourceExcel:
			calcType = svcforest.ResourceCalcTypeFile
		default:
			logs.ErrorContextf(ctx, "[GetForest] Unknown data source type: %v", *frs)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_unknown_data_source_type"
			return
		}
	case foresttype.ForestTypeQA:
		calcType = svcforest.ResourceCalcTypeQAPair
	default:
		logs.ErrorContextf(ctx, "[GetForest] Unknown forest source type: %v", *frs)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_unknown_data_source_type"
		return

	}
	metrics, err := svcforest.NewResourceCalculator(calcType).Metrics(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForest failed: %v", err)
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_resource_calculator_metrics_failed" // 资源计算器指标获取失败
		//return
	}

	resp.Response.Data = &forest.ForestInfoItemstruct{
		KnownowForest: *frs,
		IsAdmin:       isAdmin,
		ManagerIDs:    types.NewUintArray(managerIDs),
		ScopeIDs:      types.NewUintArray(scopeIDs),
	}
	if metrics != nil {
		//计量数据
		resp.Response.Data.FileCount = metrics.Count
		resp.Response.Data.TotalSize = metrics.SizeBytes
	}

	// 判断当前知识库是否已经存在图谱
	graphInfo, err := graph.GetForestGraph(ctx, req.Request.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.ErrorContextf(ctx, "GetForestGraph failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查找图谱失败"
		return
	}
	if graphInfo != nil {
		count, err := coretask.GetGraphTaskCount(ctx, graphInfo.VersionID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_forest_list_failed"
			logs.ErrorContextf(ctx, "ListForestGraph.ListGraphTaskCount failed: %v", err)
			return
		}
		graphInfo.TaskCount = count.Count
		graphInfo.SuccessTaskCount = count.SuccessCount
	}
	resp.Response.GraphInfo = graphInfo
}

// DeleteForest 知识森林删除
// @Tags 知识森林
// @Summary 知识森林删除
// @Description 知识森林删除
// @Router /forest.DeleteForest [post]
// @Param user body DeleteForestRequest true "入参"
// @Success 200 {object} DeleteForestResponse "返回值"
func DeleteForest(ctx *gin.Context, req *DeleteForestRequest, resp *DeleteForestResponse) {
	uin := runtime.Uin(ctx)
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := svcforest.DeleteForest(ctx, &svcforest.DeleteForestRequest{
		Uin:      uin,
		ForestID: req.Request.ID,
		Token:    runtime.LoginStatus(ctx).Token,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforest.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_failed"
	case errors.Is(err, svcforest.ErrNoPermission):
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, req.Request.ID)
	case errors.Is(err, svcforest.ErrForestInUse):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_in_use"
	case errors.Is(err, svcforest.ErrStatusCheckFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_status_check_failed"
	case errors.Is(err, svcforest.ErrGraphInfoFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed"
	case errors.Is(err, svcforest.ErrTaskRunning):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "任务正在运行，请稍候再试"
	case errors.Is(err, svcforest.ErrDeleteForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_forest_failed"
	case errors.Is(err, svcforest.ErrCozeMappingFailed):
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_coze_mapping_failed"))
		logs.ErrorContextf(ctx, "GetCozeMappingByCoreKGID(%v) failed ,err %s", req.Request.ID, err)
	case errors.Is(err, svcforest.ErrDeleteMappingFailed):
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_delete_mapping_failed"))
		logs.ErrorContextf(ctx, "DeleteChatAgentMapping(%v) failed ,err %s", req.Request.ID, err)
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_forest_failed"
		logs.ErrorContextf(ctx, "DeleteForest err: %v", err)
	}
}

// ListFile 知识森林文件列表
// @Tags 知识森林
// @Summary 知识森林文件列表
// @Description 知识森林文件列表
// @Router /forest.ListFile [post]
// @Param user body ListFileRequest true "入参"
// @Success 200 {object} ListFileResponse "返回值"
func ListFile(ctx *gin.Context, req *ListFileRequest, resp *ListFileResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	out, err := svcforestfile.ListFile(ctx, &svcforestfile.ListFileRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		ForestID:  req.Request.ForestID,
		ImageURL:  req.Request.ImageUrl,
		PageQuery: req.Request.PageQuery,
	})
	if err != nil {
		switch {
		case errors.Is(err, svcforestfile.ErrUserIDEmpty):
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_user_id_empty"
		case errors.Is(err, svcforestfile.ErrGetFileFailed):
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_file_failed"
		default:
			logs.ErrorContextf(ctx, "ListFile err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_forest_list_failed"
		}
		return
	}

	resp.BaseResponse = out.BaseResponse
	resp.Response = out.Response
	resp.Response.QueryResponse = out.Response.QueryResponse
}

// CreateDir 创建目录
// @Tags 知识森林
// @Summary 创建目录
// @Description 创建目录
// @Router /forest.CreateDir [post]
// @Param user body CreateDirRequest true "入参"
// @Success 200 {object} CreateDirResponse "返回值"
func CreateDir(ctx *gin.Context, req *CreateDirRequest, resp *CreateDirResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	id, err := svcforestnode.CreateDir(ctx, &svcforestnode.CreateDirRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		ForestID:  req.Request.ForestID,
		ParentID:  req.Request.ParentID,
		Name:      req.Request.Name,
	})
	if err == nil {
		resp.Response.ID = id
		return
	}
	switch {
	case errors.Is(err, svcforestnode.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = i18n.T(runtime.GetLanguage(ctx), "kecore_get_forest_failed")
	case errors.Is(err, svcforestnode.ErrNoPermission):
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", runtime.Uin(ctx), foresttype.ResourceTypeAgent, req.Request.ForestID)
	case errors.Is(err, svcforestnode.ErrGetParentNodeFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_parent_node_failed"
	case errors.Is(err, svcforestnode.ErrCheckFileExistsFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_check_file_exists_failed"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_dir_failed"
		logs.ErrorContextf(ctx, "[CreateDir] failed to create directory: %v", err)
	}
}

// DeleteDir 删除目录或文件
// @Tags 知识森林
// @Summary 删除目录或文件
// @Description 删除目录或文件
// @Router /forest.DeletePath [post]
// @Param user body DeleteDirRequest true "入参"
// @Success 200 {object} DeleteDirResponse "返回值"
func DeleteDir(ctx *gin.Context, req *DeleteDirRequest, resp *DeleteDirResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := svcforestnode.DeletePath(ctx, &svcforestnode.DeletePathRequest{
		Uin:     runtime.Uin(ctx),
		FileIDs: req.Request.FileIDs,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforestnode.ErrGetFileOrDirFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_or_dir_failed"
		logs.ErrorContextf(ctx, "[DeleteDir] failed to get file or directory: %v", err)
	case errors.Is(err, svcforestnode.ErrUnknownFileList):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_file_list"
	case errors.Is(err, svcforestnode.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
	case errors.Is(err, svcforestnode.ErrNoPermission):
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", runtime.Uin(ctx), foresttype.ResourceTypeAgent, req.Request.FileIDs)
	case errors.Is(err, svcforestnode.ErrTaskRunning):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_task_running"
	case errors.Is(err, svcforestnode.ErrFileStatusCheckFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_file_status_check_failed"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_file_or_dir_failed"
		logs.ErrorContextf(ctx, "[DeleteDir] failed to delete file or directory: %v", err)
	}
}

// RenamePath 文件重命名
// @Tags 知识森林
// @Summary 文件重命名
// @Description 文件重命名
// @Router /forest.RenamePath [post]
// @Param user body RenamePathRequest true "入参"
// @Success 200 {object} RenamePathResponse "返回值"
func RenamePath(ctx *gin.Context, req *RenamePathRequest, resp *RenamePathResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	err := svcforestnode.RenamePath(ctx, &svcforestnode.RenamePathRequest{
		Uin:     runtime.Uin(ctx),
		FileID:  req.Request.FileID,
		NewName: req.Request.NewName,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforestnode.ErrGetSourceFileFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_source_file_failed"
		logs.ErrorContextf(ctx, "[RenamePath] failed to get source file or directory info: %v", err)
	case errors.Is(err, svcforestnode.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
	case errors.Is(err, svcforestnode.ErrNoPermission):
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", runtime.Uin(ctx), foresttype.ResourceTypeAgent, req.Request.FileID)
	case errors.Is(err, svcforestnode.ErrCheckNewNameFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_check_new_name_failed"
		logs.ErrorContextf(ctx, "[RenamePath] failed to check if the file exists: %v", err)
	case errors.Is(err, svcforestnode.ErrNewNameExists):
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_new_name_exists"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_file_name_failed"
	}
}

// MovePath 文件移动
// @Tags 知识森林
// @Summary 文件移动
// @Description 文件移动
// @Router /forest.MovePath [post]
// @Param user body MovePathRequest true "入参"
// @Success 200 {object} MovePathResponse "返回值"
func MovePath(ctx *gin.Context, req *MovePathRequest, resp *MovePathResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	file, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_source_file_failed" // 获取源文件或目录信息失败
		logs.ErrorContextf(ctx, "[MovePath] failed to get source file or directory info: %v", err)
		return
	}
	// 获取知识森林
	forests, err := forest.GetForestByID(ctx, file.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识森林失败
		return
	}
	if !perm.HasManageAct(ctx, uin, forests.ID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, forests.ID)
		return
	}

	var dstparent *foresttype.KnownowForestFile
	if req.Request.DstParentID != 0 {
		dstparent, err = forest.GetForestFileByID(req.Request.DstParentID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_target_dir_failed" // 获取目标目录信息失败
			logs.ErrorContextf(ctx, "[MovePath] failed to get target directory info: %v", err)
			return
		}
	}
	isExist, err := forest.IsExistForestFile(file.ForestID, req.Request.DstParentID, file.Name)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_check_target_dir_failed" // 检查目标目录是否存在失败
		logs.ErrorContextf(ctx, "[MovePath] failed to check if the file exists: %v", err)
		return
	}
	if isExist {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_exists_in_target_dir" // 目标目录中文件已存在
		return
	}
	// 移动操作
	if err := forest.HandleMove(ctx, file, dstparent); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_move_file_failed" // 移动文件或目录失败
		return
	}
}

// PreviewFile 文件预览
// @Tags 知识森林
// @Summary 文件预览
// @Description 文件预览
// @Router /forest.PreviewFile [post]
// @Param user body PreviewFileRequest true "入参,当前只支持pdf预览"
// @Success 200 {object} PreviewFileResponse "返回值"
func PreviewFile(ctx *gin.Context, req *PreviewFileRequest, resp *PreviewFileResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	// 获取文件
	File, err := forest.GetForestFileByPath(req.Request.ForestID, req.Request.Path)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_source_file_failed" // 获取源文件或目录信息失败
		return
	}

	// 将 File.ID 转换为字符串并拼接 ".pdf"
	filePath, err := File.GetForestFilePath()
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_path_failed" // 获取文件路径失败
		return
	}

	content, err := forest.PreviewFile(*filePath)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_preview_failed" // 预览失败
		return
	}
	filename := filepath.Base(req.Request.Path)

	buf := memfile.New([]byte{})
	buf.Write(content)
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%q", filename))
	ctx.Writer.Header().Set("Content-Type", "application/pdf;charset=utf-8")
	ctx.Writer.Header().Set("Content-Transfer-Encoding", "binary")
	http.ServeContent(ctx.Writer, ctx.Request, filename, time.Now(), buf)
}

// GetFileInfo 获取文件信息
// @Tags 知识森林
// @Summary 获取文件信息
// @Description 获取文件信息
// @Router /forest.GetFileInfo [post]
// @Param user body GetFileInfoRequest true "入参"
// @Success 200 {object} GetFileInfoResponse "返回值"
func GetFileInfo(ctx *gin.Context, req *GetFileInfoRequest, resp *GetFileInfoResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	file, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_or_dir_failed" // 获取文件或目录失败
		logs.ErrorContextf(ctx, "[GetFileInfo] failed to get file or directory: %v", err)
		return
	}

	// 判断是否是文件
	if file.IsDir.Value() {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_not_a_file" // 不是文件,请选择文件
		return
	}

	frs, err := forest.GetForestByID(ctx, file.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_info_failed" // 获取知识库信息失败
		return
	}
	pathIds, pathStrs, err := forest.GetPathString(file)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_parent_path_failed" // 获取上级路径失败
		return
	}
	resp.Response = struct {
		*foresttype.KnownowForestFile
		Forest        *foresttype.KnownowForest
		ParentIDArr   []uint   `json:"parent_id_arr"`
		ParentPathArr []string `json:"parent_path_arr"`
	}{
		KnownowForestFile: file,
		Forest:            frs,
		ParentIDArr:       pathIds,
		ParentPathArr:     pathStrs,
	}
}

// ListExcelSheet 获取excel的sheet列表
// @Tags 知识森林
// @Summary 获取excel的sheet列表
// @Description 获取excel的sheet列表
// @Router /forest.ListExcelSheet [post]
// @Param user body ListExcelSheetRequest true "入参"
// @Success 200 {object} ListExcelSheetResponse "返回值"
func ListExcelSheet(ctx *gin.Context, req *ListExcelSheetRequest, resp *ListExcelSheetResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	sheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, &forest.ForestExcelSheetCond{
		ForestFileIDs: req.Request.ForestFileIDs,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[ListExcelSheet] GetListByCond failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_sheet_list_failed" // 获取 sheet 列表失败
		return
	}

	sheetList := make([]ExcelSheetListItem, 0, len(sheetEntityList))
	for _, v := range sheetEntityList {
		sheetList = append(sheetList, ExcelSheetListItem{
			ForestFileID: v.ForestFileID,
			ExcelSheetID: v.ID,
			SheetName:    v.SheetName,
		})
	}

	resp.Response.SheetList = sheetList
}

// PreviewFileByURL 通过url文件预览
// @Tags 知识森林
// @Summary 通过url文件预览
// @Description 通过url文件预览
// @Router /forest.PreviewFileByURL [post]
// @Param user body PreviewFileByURLRequest true "入参,当前只支持pdf预览"
// @Success 200 {object} PreviewFileByURLResponse "返回值"
func PreviewFileByURL(ctx *gin.Context, req *PreviewFileByURLRequest, resp *PreviewFileByURLResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	url, err := svcforestfile.PreviewFileByURL(ctx, &svcforestfile.PreviewFileByURLRequest{
		FileID:     req.Request.FileID,
		IsDownload: req.Request.IsDownLoad,
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
			logs.ErrorContextf(ctx, "get storage config error: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_upload_config_failed"))
		case errors.Is(err, svcforestfile.ErrParseURLFailed):
			logs.ErrorContextf(ctx, "get referer error: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_parse_url_failed"))
		case errors.Is(err, svcforestfile.ErrCreateStorageFailed):
			logs.ErrorContextf(ctx, "[PreviewFileByURL] new storage error: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_create_storage_failed"))
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

// RecentlyForest 最近知识库
// @Tags 知识森林
// @Summary 最近知识库
// @Description 最近知识库
// @Router /forest.RecentlyForest [post]
// @Param user body RecentlyForestRequest true "入参"
// @Success 200 {object} RecentlyForestResponse "返回值"
func RecentlyForest(ctx *gin.Context, req *RecentlyForestRequest, resp *RecentlyForestResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	company_id := runtime.CompanyID(ctx)
	if uin == 0 || company_id == 0 {
		return
	}
	var err error
	resp.Response.Forests, err = forest.GetRecentlyForest(uin)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_recent_forest_failed" // 获取最近知识森林失败
		logs.ErrorContextf(ctx, "[RecentlyForest] failed to get recent forests: %v", err)
		return
	}
	resp.Response.ForestCount, err = forest.GetForestCount(company_id)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_count_failed" // 获取知识森林数量失败
		logs.ErrorContextf(ctx, "[RecentlyForest] failed to get forest count: %v", err)
		return
	}
	resp.Response.FileCount, err = forest.GetFileCount(company_id)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_count_failed" // 获取文件数量失败
		logs.ErrorContextf(ctx, "[RecentlyForest] failed to get file count: %v", err)
		return
	}
}

// GetFilePath 获取文件路径
// @Tags 知识森林
// @Summary 获取文件路径
// @Description 获取文件路径
// @Router /forest.GetFilePath [post]
// @Param user body GetFilePathRequest true "入参"
// @Success 200 {object} GetFilePathResponse "返回值"
func GetFilePath(ctx *gin.Context, req *GetFilePathRequest, resp *GetFilePathResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_list_failed" // 获取文件列表失败
		logs.ErrorContextf(ctx, "[GetForestFileByID] failed to get file: %v", err)
		return
	}
	if f.ForestID != req.Request.ForestID {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_file_not_in_forest" // 森林中文件不存在
		logs.ErrorContextf(ctx, "[GetForestFileByID] unknown forest file: %v", err)
		return
	}
	pids, ps, err := forest.GetPathString(f)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_path_failed" // 获取文件路径失败
		logs.ErrorContextf(ctx, "[GetPathString] 获取路径失败: %v", err)
		return
	}
	resp.Response.PathIds = pids
	resp.Response.PathStrings = ps
}

// DeleteFile 删除知识森林文件
// @Tags 知识森林
// @Summary 知识森林文件删除
// @Description 知识森林文件删除
// @Router /forest.DeleteFile [post]
// @Param user body DeleteFileRequest true "入参"
// @Success 200 {object} DeleteFileResponse "返回值"
func DeleteFile(ctx *gin.Context, req *DeleteFileRequest, resp *DeleteFileResponse) {
	uin := runtime.Uin(ctx)
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFile GetForestFileByID err: %v", err)
		return
	}

	ff, err := forest.GetForestByID(ctx, f.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFile GetForestByID err: %v", err)
		return
	}
	if !perm.HasManageAct(ctx, uin, ff.ID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, ff.ID)
		return
	}

	if err = forest.DeleteFilesStatusCheck(ctx, []uint{f.ID}); err != nil {
		if errors.Is(err, forest.ErrHasRunningTask) {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_file_in_use" // 正在解析无法删除
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_file_status_check_failed" // 文件删除状态检查失败
		return
	}

	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {

		graphInfo, err := graph.GetForestGraph(ctx, f.ForestID)
		if err == nil {
			if err = graph.DeleteFileGraph(ctx, tx, graphInfo, f); err != nil {
				logs.ErrorContextf(ctx, "DeleteFile DeleteFileGraph err: %v", err)
				return err
			}
		}

		// delete the file related all tasks
		if err := coretask.DeleteTasksByFileIDs(ctx, []uint{f.ID}); err != nil {
			return err
		}

		if err = essearch.DeleteFileReferences(ctx, ff.EsIndex(), []uint{f.ID}); err != nil {
			return err
		}
		if err = tx.Delete(&foresttype.KnownowForestFile{}, f.ID).Error; err != nil {
			return err
		}
		//nbgraph.DeleteFiles(ctx, ff.ID, []uint{f.ID}, ff.EsIndex())
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "DeleteFile err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_file_failed" // 删除文件资源失败
		return
	}
}

// GetNameByModuleIDs 通过模块ids获取名称
// @Tags 知识森林
// @Summary 通过模块ids获取名称
// @Description 通过模块ids获取名称
// @Router /forest.GetNameByModuleIDs [post]
// @Param user body GetNameByModuleIDsRequest true "入参"
// @Success 200 {object} GetNameByModuleIDsResponse "返回值"
func GetNameByModuleIDs(ctx *gin.Context, req *GetNameByModuleIDsRequest, resp *GetNameByModuleIDsResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	moduleIdsMap := make(map[foresttype.ForestModule][]uint)
	for _, v := range req.Request.ModuleIDList {
		moduleIdsMap[v.Module] = v.IDs
	}

	nameRes, err := forest.GetNameByModuleIDs(ctx, moduleIdsMap)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetNameByModuleIDs] failed to get name by module ids: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_name_failed" // 获取名称失败
		return
	}
	nameList := make([]GetNameByModuleIDsNameListItem, 0, len(nameRes.NameList))
	for _, v := range nameRes.NameList {
		nameList = append(nameList, GetNameByModuleIDsNameListItem{
			ID:     v.ID,
			Module: v.Module,
			Name:   v.Name,
		})
	}
	resp.Response.NameList = nameList
}

// GetResourceBaseInfo 获取资源基础信息
// @Tags 知识森林
// @Summary 获取资源基础信息
// @Description 获取资源基础信息
// @Router /forest.GetResourceBaseInfo [post]
// @Param request body dtoforest.GetResourceBaseInfoRequest true "request"
// @Success 200 {object} dtoforest.GetResourceBaseInfoResponse "response"
func GetResourceBaseInfo(ctx *gin.Context, req *dtoforest.GetResourceBaseInfoRequest, resp *dtoforest.GetResourceBaseInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetResourceBaseInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcforest.GetResourceBaseInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetResourceBaseInfo] svcforest.GetResourceBaseInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_resource_base_info_failed" // 获取资源基础信息失败
		return
	}
	resp.Code = res.Code
	resp.Message = "message_id"
	resp.Response = res.Response
}

// RenameForest 知识库重命名
// @Tags 知识森林
// @Summary 知识库重命名
// @Description 知识库重命名
// @Router /forest.RenameForest [post]
// @Param request body RenameForestRequest true "request"
// @Success 200 {object} apiobj.BaseResponse "response"
func RenameForest(ctx *gin.Context, req *RenameForestRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "[RenameForest] request invalid, req: %+v", req)
		return
	}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	if !perm.HasManageAct(ctx, uin, req.Request.ID, foresttype.ResourceTypeForest) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeForest, req.Request.ID)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_no_permission_to_update")) // 无权限更新此资源
		return
	}

	if forest.CheckForestNameExists(ctx, req.Request.ID, req.Request.Name, companyID) {
		logs.ErrorContextf(ctx, "[RenameForest] renamed %s already exists", req.Request.Name)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_name_exists"
		return
	}

	if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
		Where("deleted_at IS NULL").
		Where("id = ?", req.Request.ID).
		Update("name", req.Request.Name).
		Error; err != nil {
		logs.ErrorContextf(ctx, "update forestname failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_forest_name_update_failed"
		return
	}
}

// SetResourceEnable 设置资源启用状态
// @Tags 知识库资源管理
// @Summary 设置资源启用状态
// @Description 设置资源启用状态
// @Router /forest.SetResourceEnable [post]
// @Param request body dtoforest.SetResourceEnableRequest true "request"
// @Success 200 {object} dtoforest.SetResourceEnableResponse "response"
func SetResourceEnable(ctx *gin.Context, req *dtoforest.SetResourceEnableRequest, resp *dtoforest.SetResourceEnableResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SetResourceEnable] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcforest.SetResourceEnable(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SetResourceEnable] svcforest.SetResourceEnable failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_set_resource_enable_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
