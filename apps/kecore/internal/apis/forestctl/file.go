package forestctl

import (
	"fmt"
	"net/http"
	stdurl "net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/decoupler"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kecore/services/svcexcelforest"
	"github.com/insmtx/corekg/pkgs/apis/code"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
	"gorm.io/gorm"
)

// UploadFile 上传文件
// @Tags 知识森林文件
// @Summary 上传文件
// @Description 上传文件
// @Router /forest.UploadFile [post]
// @Accept multipart/form-data
// @Param file formData file true "文件"
// @Param forest_id formData string true "知识森林id"
// @Param parent_id formData string false "父级id"
// @Success 200 {object} UploadFileResponse "返回值"
func UploadFile(ctx *gin.Context) {
	resp := &UploadFileResponse{}
	uin := runtime.Uin(ctx)
	forestIDStr := ctx.Request.FormValue("forest_id")
	forestID, err := strconv.Atoi(forestIDStr)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: UploadFile failed,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_upload_failed" // 上传失败
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	if forestID <= 0 {
		logs.ErrorContextf(ctx, "forest: UploadFile failed,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_invalid_forest_id" // 无效的知识森林ID
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	// 获取知识森林
	forests, err := forest.GetForestByID(ctx, uint(forestID))
	if err != nil {
		logs.ErrorContextf(ctx, "[UploadFile] get forest failed, err = %v", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识森林失败
		return
	}
	if !perm.HasManageAct(ctx, uin, forests.ID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, forests.ID)
		return
	}

	file, err := decoupler.UploadFile(ctx, forestID)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: UploadFile failed,err = %v", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_upload_failed" // 上传失败
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	if file.ParseStatus != foresttype.TaskStatusUnsupported {
		if err = coretask.CreateForestTask(ctx, file); err != nil {
			logs.ErrorContextf(ctx, "forest: CreateForestTask failed,err = %v", err.Error())
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_create_task_failed" // 创建任务失败
		}
	}
	resp.Code = 0
	resp.Message = "kecore_upload_success" // 文件上传成功
	resp.Response.ForestFileID = file.ID
	ctx.JSON(http.StatusOK, resp)
}

// Deprecated: 历史重构弃用
// PreUploadFile 获取文件上传预签名
// @Tags 知识森林文件
// @Summary 获取文件上传预签名
// @Description 获取文件上传预签名
// @Router /forest.PreUploadFile [post]
// @Param user body PreUploadFileRequest true "入参,当前只支持pdf预览"
// @Success 200 {object} PreUploadFileResponse "返回值"
func PreUploadFile(ctx *gin.Context, req *PreUploadFileRequest, resp *PreUploadFileResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	var (
		parent *foresttype.KnownowForestFile
		err    error
		url    string
	)
	if req.Request.ParentID > 0 {
		parent, err = forest.GetForestFileByID(req.Request.ParentID)
		if err != nil {
			logs.ErrorContextf(ctx, "forest: PreUploadFile failed,err = %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_query_parent_failed" // 查询父级目录失败
			return
		}
	}
	isExist, err := forest.IsExistForestFile(req.Request.ForestID, req.Request.ParentID, req.Request.FileName)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: PreUploadFile IsExistForestFile failed,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询同名文件失败
		return
	}
	if isExist {
		logs.WarnContextf(ctx, "该文件已经存在,创建同名文件_timestep")
		req.Request.FileName = fmt.Sprintf("%v_%v%v", filepath.Base(req.Request.FileName), time.Now().Format("20060102_150405"), filepath.Ext(req.Request.FileName))
	}
	// 创建file 对象
	finfo := &foresttype.KnownowForestFile{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       runtime.Uin(ctx),
		ForestID:  req.Request.ForestID,
		IsDir:     -1,
		Name:      req.Request.FileName,
		Ext:       strings.ToLower(filepath.Ext(req.Request.FileName)),
		Size:      req.Request.FileSize,
		Status:    foresttype.FileStatusPending,
		FileConfig: foresttype.FileConfig{
			SplitConfig: req.Request.SplitConfig,
		},
	}
	if !forest.PreViewAble(finfo) {
		finfo.PreViewAble = foresttype.PreViewAbleStatusUnsupported
		finfo.ParseStatus = foresttype.TaskStatusUnsupported
		finfo.KnowledgeStatus = foresttype.TaskStatusUnsupported
		finfo.AnalysisStatus = foresttype.TaskStatusUnsupported
		finfo.GraphStatus = foresttype.TaskStatusUnsupported
	}
	if parent == nil {
		// 知识森林根目录下创建
		finfo.Depth = 1
	} else {
		finfo.ParentID = parent.ID
		finfo.ParentIDs = fmt.Sprintf("%s%d/", parent.ParentIDs, parent.ID)
		finfo.Depth = parent.Depth + 1
	}
	var path string
	err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(finfo).Error; err != nil {
			return err
		}
		fi := &storage.FileInfo{
			CompanyID: runtime.CompanyID(ctx),
			Uin:       runtime.Uin(ctx),
			Filename:  finfo.Name,
			Size:      finfo.Size,
			FileExt:   finfo.Ext,
			// TODO 历史版本实现，弃用
			// StoragePath: finfo.GetForestFilePath(),
			Status: storage.FileStatusUploading,
		}
		url, err = fs.Forest.GetPresignedURL(http.MethodPut, fi.StoragePath)
		path = fi.StoragePath
		if err != nil {
			return err
		}
		fi.PublicURL = fs.Forest.GetPublicURL(fi.StoragePath, false)
		err = fs.CreateFileInfo(fi)
		if err != nil {
			return err
		}
		finfo.CoreFileID = fi.ID
		if err := tx.Save(finfo).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_core_upload_failed"
		return
	}
	resp.Response.FileID = finfo.ID
	resp.Response.UploadURL = url
	if version.DeployMode() != "" && version.DeployMode() != global.DeployModeOpenPO {
		var (
			cfg config.StorageConfig
		)
		if err = settings.GetYaml(settings.SettingGroupCore, storage.SettingPrefix+fs.PurposeKeFile, &cfg); err != nil {
			logs.ErrorContextf(ctx, "get storage config error: %v", err)
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
		url, err := st.GetPresignedURL(http.MethodPut, path)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_presigned_url_failed" // 获取预签名文件地址失败
			return
		}
		resp.Response.UploadURL = url
	}
}

// Deprecated: 历史重构弃用
// UploadFileCallBack 预签名上传回调
// @Tags 知识森林文件
// @Summary 预签名上传回调
// @Description 预签名上传回调
// @Router /forest.UploadFileCallBack [post]
// @Param user body UploadFileCallBackRequest true "入参,当前只支持pdf预览"
// @Success 200 {object} UploadFileCallBackResponse "返回值"
func UploadFileCallBack(ctx *gin.Context, req *UploadFileCallBackRequest, resp *UploadFileCallBackResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	file_info, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: UploadFileCallBack failed,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询文件失败
		return
	}
	if req.Request.Status != foresttype.FileStatusNormal {
		// 失败不用修改文件
		// 返回新的预签名
		if version.DeployMode() != "" && version.DeployMode() != global.DeployModeOpenPO {
			var (
				cfg config.StorageConfig
			)
			if err = settings.GetYaml(settings.SettingGroupCore, storage.SettingPrefix+fs.PurposeKeFile, &cfg); err != nil {
				logs.ErrorContextf(ctx, "get storage config error: %v", err)
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
			// st, err := storage.NewStorageWithCfg(cfg)
			// if err != nil {
			// 	logs.ErrorContextf(ctx, "[PreviewFileByURL] new storage error: %v cfg[+%v]", err, cfg)
			// 	runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_create_storage_failed")) // 创建存储器失败
			// 	return
			// }

			// TODO 历史版本实现，弃用
			// url, err := st.GetPresignedURL(http.MethodPut, file_info.GetForestFilePath())
			// if err != nil {
			// 	resp.Code = errcode.ErrCode_InternalError
			// 	resp.Message = "kecore_get_presigned_url_failed" // 获取预签名文件地址失败
			// 	return
			// }
			// resp.Response.UploadURL = url
		} else {

			// TODO 历史版本实现，弃用
			// resp.Response.UploadURL, err = fs.Forest.GetPresignedURL(http.MethodPut, file_info.GetForestFilePath())
			// if err != nil {
			// 	logs.ErrorContextf(ctx, "forest: UploadFileCallBack failed,err = %v", err)
			// 	resp.Code = errcode.ErrCode_InternalError
			// 	resp.Message = "kecore_get_presigned_failed" // 获取预签名失败
			// 	return
			// }
		}

		return
	}
	if file_info.Status == foresttype.FileStatusNormal {
		// 文件已经上传成功，不用再上传
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_duplicate_upload" // 请勿重复上传
		return
	}
	file, err := storage.GetFileByID(dbutil.Core(), file_info.CoreFileID)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: UploadFileCallBack failed,err = %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_info_failed" // 查询文件信息失败
		return
	}
	err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		file_info.PriviewFileID = file_info.CoreFileID
		file_info.PriviewExt = file_info.Ext
		// 如果是ppt或者word，转pdf
		switch file_info.Ext {
		case ".ppt", ".pptx", ".doc", ".docx", ".ofd":
			f, err := fs.Forest.ReadFile(file.StoragePath)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack ReadFile pdf failed %v", err)
				return err
			}
			pdf, err := decoupler.FileToPDF(ctx, f, file_info.Name)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack FileToPDF failed %v", err)
				return err
			}
			defer pdf.Close()
			file_info.PriviewExt = ".pdf"
			fi := &storage.FileInfo{
				CompanyID: runtime.CompanyID(ctx),
				Uin:       runtime.Uin(ctx),
				Filename:  file.Filename + file_info.PriviewExt,
				Size:      file_info.Size,
				FileExt:   file_info.PriviewExt,

				// TODO 历史版本实现，弃用
				// StoragePath: file_info.GetForestPriviewFilePath(),
			}
			err = fs.Forest.Save(ctx, fi, pdf)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack Save pdf failed %v", err)
				return err
			}
			fi.PublicURL = fs.Forest.GetPublicURL(fi.StoragePath, false)
			err = fs.CreateFileInfo(fi)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack CreateFileInfo failed %v", err)
				return err
			}
			file_info.PriviewFileID = fi.ID
		case ".csv":
			f, err := fs.Forest.ReadFile(file.StoragePath)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack ReadFile csv failed %v", err)
				return err
			}
			excel, err := decoupler.CSVToExcel(ctx, f)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack CSVToExcel failed %v", err)
				return err
			}
			file_info.PriviewExt = ".xlsx"
			fi := &storage.FileInfo{
				CompanyID: runtime.CompanyID(ctx),
				Uin:       runtime.Uin(ctx),
				Filename:  file.Filename + file_info.PriviewExt,
				Size:      file.Size,
				FileExt:   file_info.PriviewExt,

				// TODO 历史版本实现，弃用
				// StoragePath: file_info.GetForestPriviewFilePath(),
			}
			err = fs.Forest.Save(ctx, fi, excel)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack Save csv failed %v", err)
				return err
			}
			fi.PublicURL = fs.Forest.GetPublicURL(fi.StoragePath, false)
			err = fs.CreateFileInfo(fi)
			if err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack CreateFileInfo failed %v", err)
				return err
			}
			file_info.PriviewFileID = fi.ID
		}
		file_info.Status = foresttype.FileStatusNormal
		file.Status = storage.FileStatusNormal
		if err := tx.Save(file_info).Error; err != nil {
			logs.ErrorContextf(ctx, "forest: UploadFileCallBack Save(file_info) failed %v", err)
			return err
		}
		if err := dbutil.Core().Save(file).Error; err != nil {
			logs.ErrorContextf(ctx, "forest: UploadFileCallBack dbutil.Core().Save(file) failed %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		logs.ErrorContextf(ctx, "forest: UploadFileCallBack failed,err = %v,file_info: %s", err, logs.JSON(file_info))
		resp.Code = code.ErrCode_HideErr
		resp.Message = "kecore_update_file_status_failed" // 修改文件状态失败
		return
	}

	forestEntity, err := forest.NewForestDao().GetByID(ctx, file_info.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: GetByID failed,err = %v ,file_info:%+v", err.Error(), file_info)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}

	switch forestEntity.DataSourceType {
	// excel 知识库文件触发解析任务
	case foresttype.ForestDataSourceExcel:
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logs.ErrorContextf(ctx, "[UploadFileCallBack] svcexcelforest.AnalyzeXlsx panic: %v", err)
				}
			}()
			if err := svcexcelforest.AnalyzeXlsx(ctx.Copy(), &svcexcelforest.AnalyzeXlsxReq{
				ForestFileID: file_info.ID,
			}); err != nil {
				updateMap := map[string]interface{}{}
				for _, v := range svcexcelforest.TaskStatusFields {
					updateMap[v] = foresttype.TaskStatusFail
				}
				if err := forest.NewForestFileDao().UpdateMap(ctx, file_info.ID, updateMap); err != nil {
					logs.ErrorContextf(ctx, "[UploadFileCallBack] forest file updateMap failed, err = %v", err)
				}
				logs.ErrorContextf(ctx, "[UploadFileCallBack] svcexcelforest.AnalyzeXlsx failed, err = %v", err)
			}

		}()
		return
	default:
		// 生产任务
		if file_info.ParseStatus != foresttype.TaskStatusUnsupported {
			if err = coretask.CreateForestTask(ctx, file_info); err != nil {
				logs.ErrorContextf(ctx, "forest: CreateForestTask failed,err = %v", err.Error())
				resp.Code = errcode.ErrCode_InternalError
				resp.Message = "kecore_create_task_failed" // 创建任务失败
			}
		}
	}

	resp.Response.FileID = file_info.ID
}

// ListParseHistory 获取解析历史
// @Tags 知识森林文件
// @Summary 获取解析历史
// @Description 获取解析历史
// @Router /forest.ListParseHistory [post]
// @Param user body ListParseHistoryRequest true "入参"
// @Success 200 {object} ListParseHistoryResponse "返回值"
func ListParseHistory(ctx *gin.Context, req *ListParseHistoryRequest, resp *ListParseHistoryResponse) {
	if req.Validate(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "ListParseHistory Validate failed, req = %#v, resp = %#v", req, resp)
		return
	}

	req.Request.CompanyID = runtime.CompanyID(ctx)
	if err := forest.QueryParseHistory(ctx, req.Request, &resp.Response); err != nil {
		logs.ErrorContextf(ctx, "forest: ListParseHistory failed,err: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_parse_history_failed")) // 获取解析历史失败
	}
}

// RetryParse 解析重试
// @Tags 知识森林文件
// @Summary 解析重试
// @Description 解析重试
// @Router /forest.RetryParse [post]
// @Param user body RetryParseRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func RetryParse(ctx *gin.Context, req *RetryParseRequest, resp *apiobj.BaseResponse) {
	if req.Validate(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "RetryParse Validate failed, req = %#v, resp = %#v", req, resp)
		return
	}

	f, err := forest.GetForestFileByID(req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "RetryParse: forest.GetForestFileByID(%v) failed: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_file_failed")) // 获取文件失败
		return
	}

	var ts []task.Task
	if err = dbutil.Core().
		Where("deleted_at IS NULL").
		Where("subject_id = ?", f.ID).
		Where("task_status in ?", []task.TaskStatus{task.TaskStatusFail, task.TaskStatusTimeout}).
		Find(&ts).Error; err != nil {
		logs.ErrorContextf(ctx, "RetryParse: forest.GetForestFileByID(%v) failed: %v", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_file_task_list_failed")) // 获取文件任务列表失败
		return
	}

	for _, t := range ts {
		switch t.TaskType {
		case coretask.CopyTask, coretask.PraseVideoTask, coretask.PraseTask:
			f.ParseStatus = foresttype.TaskStatusPending
		case coretask.KnowledgeTask:
			f.KnowledgeStatus = foresttype.TaskStatusPending
		case coretask.DescriptionTask:
			f.DescStatus = foresttype.TaskStatusPending
		case coretask.GraphTask:
			f.GraphStatus = foresttype.TaskStatusPending
		}
	}

	if err := dbutil.Core().Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(task.TableNameCoreTask).
			Where("deleted_at IS NULL").
			Where("subject_id = ?", req.Request.ID).
			Where("task_status in ?", []task.TaskStatus{task.TaskStatusFail, task.TaskStatusTimeout}).
			Updates(map[string]interface{}{
				"task_status": task.TaskStatusPending,
				"redo":        0,
				"priority":    10,
			}).Error; err != nil {
			return err
		}
		return dbutil.Knownow().Save(f).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "RetryParse Updates failed subject_id[%v] err: %v ", req.Request.ID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_retry_parse_failed")) // 重试解析失败
	}
}

// ResplitChunk 重拆当前文件chunk
// @Tags 知识森林文件
// @Summary 重拆当前文件chunk
// @Description 重拆当前文件chunk
// @Router /forest.ResplitChunk [post]
// @Param user body ResplitChunkRequest true "入参"
// @Success 200 {object} ResplitChunkResponse "返回值"
func ResplitChunk(ctx *gin.Context, req *ResplitChunkRequest, resp *ResplitChunkResponse) {
	if req.Validate(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "RetryParse Validate failed, req = %#v, resp = %#v", req, resp)
		return
	}
	file, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: ResplitChunk failed,err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_file_info_failed" // 获取文件信息失败
		return
	}
	if file.KnowledgeStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_status_invalid" // 文件状态异常
		return
	}
	forestInfo, err := forest.GetForestByID(ctx, file.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "forest: ResplitChunk failed,err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_info_failed" // 获取知识库信息失败
		return
	}
	if forestInfo.ForestType != foresttype.ForestTypeCAD &&
		forestInfo.ForestType != foresttype.ForestTypeFile {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_type_not_supported" // 知识库类型不支持
		return
	}
	file.KnowledgeStatus = foresttype.TaskStatusPending
	file.GraphStatus = foresttype.TaskStatusPending
	file.FileConfig.SplitConfig = req.Request.SplitConfig
	// 删除chunk
	err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(file).Error; err != nil {
			logs.ErrorContextf(ctx, "forest: ResplitChunk failed,err: %v", err)
			return err
		}
		return dbutil.Core().Transaction(func(tx *gorm.DB) error {
			if err := coretask.DeleteChunkTask(ctx, tx, file.ID); err != nil {
				logs.ErrorContextf(ctx, "forest: DeleteChunkTask failed,err: %v", err)
				return err
			}
			if err := coretask.CreateReChunkTask(ctx, tx, file, forestInfo); err != nil {
				logs.ErrorContextf(ctx, "forest: CreateReChunkTask failed,err: %v", err)
				return err
			}
			return nil
		})
	})
	if err != nil {
		logs.ErrorContextf(ctx, "forest: ResplitChunk failed,err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_resplit_chunk_failed" // 重新插入拆分任务失败
		return
	}
}
