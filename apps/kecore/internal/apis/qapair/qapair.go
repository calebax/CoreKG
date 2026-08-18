package qapair

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/xuri/excelize/v2"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

// CreateQAPair 创建问答对
// @Tags 知识库问答对
// @Summary 创建问答对
// @Description 创建问答对
// @Router /forest.CreateQAPair [post]
// @Param user body CreateQAPairRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func CreateQAPair(ctx *gin.Context, req *CreateQAPairRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)

	frs, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}
	if !perm.HasManageAct(ctx, uin, frs.ID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, frs.ID)
		return
	}

	now := time.Now()
	fqa := &ragtypes.FQA{
		Common: ragtypes.Common{
			CreatedAt:  now,
			UpdatedAt:  now,
			ForestID:   req.Request.ForestID,
			Uin:        uin,
			CompanyID:  cmpID,
			Type:       ragtypes.ChunkTypeFQA,
			SourceFrom: ragtypes.SourceFromTypeManualImport,
			Enable:     1,
		},
		QALable:    req.Request.Label,
		QAQuestion: req.Request.Question,
		QAAnswer:   req.Request.Answer,
	}
	qas, err := essearch.GeneratFQA(ctx, frs.EsIndex(), fqa, req.Request.SubQuestion)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_generate_qa_failed" // 生成问答对失败
		return
	}
	if err = essearch.InsertFQA(ctx, frs.EsIndex(), qas); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_insert_qa_failed" // 插入问答对失败
		return
	}
}

// DeleteQAPair 删除问答对
// @Tags 知识库问答对
// @Summary 删除问答对
// @Description 删除问答对
// @Router /forest.DeleteQAPair [post]
// @Param user body DeleteQAPairRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func DeleteQAPair(ctx *gin.Context, req *DeleteQAPairRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)

	frs, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}
	if !perm.HasManageAct(ctx, uin, frs.ID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, frs.ID)
		return
	}

	w, err := essearch.NewWrapper(ctx, frs.EsIndex(),
		"",
		req.Request.QuestionIDs, []uint{req.Request.ForestID},
		nil, nil)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_wrapper_init_failed" // 初始化wrapper失败
		return
	}
	if err = w.DeleteByQuery(); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_failed" // 删除操作失败
		return
	}
}

// ListQAPair 查询问答对列表
// @Tags 知识库问答对
// @Summary 查询问答对列表
// @Description 查询问答对列表
// @Router /forest.ListQAPair [post]
// @Param user body ListQAPairRequest true "入参"
// @Success 200 {object} ListQAPairResponse "返回值"
func ListQAPair(ctx *gin.Context, req *ListQAPairRequest, resp *ListQAPairResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	frs, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}

	w, err := essearch.NewWrapper(
		ctx, frs.EsIndex(),
		"", nil, []uint{req.Request.ForestID},
		nil, &req.Request.PageQuery)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_wrapper_init_failed" // 初始化wrapper失败
		return
	}
	rs, err := w.SearchFQA()
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_qa_failed" // 查询QA对失败
		return
	}
	resp.Response = *rs
}

// ModifyQAPair 更新问答对
// @Tags 知识库问答对
// @Summary 更新问答对
// @Description 更新问答对
// @Router /forest.ModifyQAPair [post]
// @Param user body ModifyQAPairRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func ModifyQAPair(ctx *gin.Context, req *ModifyQAPairRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	uin := runtime.Uin(ctx)

	frs, err := forest.GetForestByID(ctx, req.Request.Main.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_qa_failed" // 获取问答对失败
		return
	}
	if !perm.HasManageAct(ctx, uin, frs.ID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, frs.ID)
		return
	}

	var qIDs = []string{req.Request.Main.ID}
	for _, v := range req.Request.Child {
		if len(v.ID) > 0 {
			qIDs = append(qIDs, v.ID)
		}
	}
	w, err := essearch.NewWrapper(
		ctx, frs.EsIndex(),
		"", qIDs, []uint{frs.ID},
		nil, &apiobj.PageQuery{
			Limit:   0,
			ListAll: true,
			Offset:  0,
			OrderBy: []string{"updated"},
		})
	if err != nil {
		logs.ErrorContextf(ctx, "NewWrapper err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_wrapper_init_failed" // 初始化wrapper失败
		return
	}

	if err = w.ModifyQAPair(req.Request.FQAItem); err != nil {
		logs.ErrorContextf(ctx, "EsSearchWrapper.ModifyQAPair err: %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_qa_failed" // 更新问答对失败
		return
	}

	return
}

// UploadQAPair 导入问答对
// @Tags 知识库问答对
// @Summary 导入问答对
// @Description 导入问答对
// @Router /forest.UploadQAPair [post]
// @Accept multipart/form-data
// @Param file formData file true "问答对文件"
// @Param forest_id formData string true "知识库id"
// @Success 200 {object} UploadQAPairResponse "返回值"
func UploadQAPair(ctx *gin.Context) {
	frsIDStr := ctx.Request.FormValue("forest_id")
	frsID, err := strconv.Atoi(frsIDStr)
	if err != nil || frsID <= 0 {
		runtime.InternalError(ctx, i18n.TWithData(runtime.GetLanguage(ctx), "kecore_get_forest_id_failed_data", map[string]interface{}{
			"forest_id": frsIDStr,
		}))
		return
	}

	// 获取文件
	f, _, err := ctx.Request.FormFile("file")
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_qa_file_failed"))
		return
	}
	defer f.Close()

	uin := runtime.Uin(ctx)
	if !perm.HasManageAct(ctx, uin, uint(frsID), foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeForest, frsID)
		return
	}

	frs, err := forest.GetForestByID(ctx, uint(frsID))
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID(%v): failed %v", frsID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_forest_failed"))
		return
	}

	if frs.ForestType != foresttype.ForestTypeQA {
		logs.WarnContextf(ctx, "user uin[%v] desire import qa but forest type[%v] no meet", uin, frs.ForestType)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_invalid_forest_type"))
		return
	}

	// 打开Excel文件
	exc, err := excelize.OpenReader(f)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_read_qa_file_failed"))
		return
	}
	// 获取第一个工作表的名字
	sheetName := exc.GetSheetName(0)
	// 获取该工作表的所有行
	rows, err := exc.GetRows(sheetName)
	if err != nil {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_qa_rows_failed"))
	}

	resp := UploadQAPairResponse{}

	// 遍历每一行
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if i == keqa.MaxUploadQAItem+1 {
			logs.WarnContextf(ctx, "QApairs uploaded[%v lines] have overflow MaxParseCount[%d lines] ", len(rows), keqa.MaxUploadQAItem)
			break
		}
		//遍历每一列
		item := &keqa.PureQAItem{}
		for j, col := range row {
			if j == 0 && len(col) > 0 {
				item.Question = col
			} else if j == 1 && len(col) > 0 {
				item.Answer = col
			}
		}
		if len(item.Question)*len(item.Answer) > 0 {
			resp.Response.QAList = append(resp.Response.QAList, item)
		}
	}
	resp.Response.TotalLines = uint(len(rows) - 1)
	resp.Response.ValidLines = uint(len(resp.Response.QAList))

	ctx.JSON(http.StatusOK, resp)
}

// CommitQAPair 提交上传的问答对
// @Tags 知识库问答对
// @Summary 提交上传的问答对
// @Description 提交上传的问答对
// @Router /forest.CommitQAPair [post]
// @Param user body CommitQAPairRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func CommitQAPair(ctx *gin.Context, req *CommitQAPairRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)
	if !perm.HasManageAct(ctx, uin, req.Request.ForestID, foresttype.ResourceTypeForest) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource"))
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeForest, req.Request.ForestID)
		return
	}

	frs, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByID(%v) failed: %v", req.Request.ForestID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_forest_failed"))
		return
	}

	if frs.ForestType != foresttype.ForestTypeQA {
		logs.WarnContextf(ctx, "user uin[%v] desire import qa but forest type[%v] no meet", uin, frs.ForestType)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_invalid_forest_type"))
		return
	}

	if len(req.Request.QAList) == 0 {
		return
	}

	qas := make([]*ragtypes.FQA, 0)
	for _, v := range req.Request.QAList {
		now := time.Now()
		embed, err := essearch.GetEmbedding(v.Question)
		if err != nil {
			logs.ErrorContextf(ctx, "GetEmbedding(%v) err: %v", v.Question, err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_generate_qa_failed"))
			return
		}
		qas = append(qas, &ragtypes.FQA{
			Common: ragtypes.Common{
				ID:         uuid.NewString(),
				CreatedAt:  now,
				UpdatedAt:  now,
				ForestID:   req.Request.ForestID,
				Uin:        uin,
				CompanyID:  cmpID,
				Type:       ragtypes.ChunkTypeFQA,
				SourceFrom: ragtypes.SourceFromTypeBatchUpload,
			},
			QAAnswerID: uuid.NewString(),
			Embedding:  embed,
			QAQuestion: v.Question,
			QAAnswer:   v.Answer,
		})
	}

	if err = essearch.InsertFQA(ctx, frs.EsIndex(), qas); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_insert_qa_failed"
		return
	}
}
