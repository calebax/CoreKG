package forestctl

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/apps/kecore/services/svcforest"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

// ListForest 知识库列表
func ListForest(ctx *gin.Context, req *dtokeapi.ListForestRequest, resp *dtokeapi.ListForestResponse) {
	if !req.ValidListForest(&resp.BaseResponse) {
		return
	}

	out, err := svcforest.ListForest(ctx, &svcforest.ListForestRequest{
		Uin:               runtime.Uin(ctx),
		CompanyID:         runtime.CompanyID(ctx),
		Query:             req.Request,
		PresetWhenListing: false,
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_list_failed"
		return
	}
	resp.Response.Total = out.Total
	resp.Response.Offset = req.Request.Offset
	resp.Response.Limit = req.Request.Limit
	resp.Response.Data = make([]*dtokeapi.ForestItem, 0, len(out.Data))
	for _, item := range out.Data {
		resp.Response.Data = append(resp.Response.Data, dtokeapi.NewForestItem(item))
	}
}

// BatchGetForest 批量查询知识库信息
func BatchGetForest(ctx *gin.Context, req *dtokeapi.BatchGetForestRequest, resp *dtokeapi.BatchGetForestResponse) {
	if !req.ValidBatchGetForest(&resp.BaseResponse) {
		return
	}

	filterValues := make([]string, 0, len(req.Request.ForestIDs))
	for _, forestID := range req.Request.ForestIDs {
		if forestID == 0 {
			continue
		}
		filterValues = append(filterValues, strconv.FormatUint(uint64(forestID), 10))
	}
	if len(filterValues) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapi_invalid_forest_ids"
		return
	}

	out, err := svcforest.ListForest(ctx, &svcforest.ListForestRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		Query: apiobj.PageQuery{
			ListAll: true,
			Filters: []apiobj.Filter{{
				Field: "id",
				Value: filterValues,
			}},
		},
		PresetWhenListing: false,
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_list_failed"
		return
	}

	forestMap := make(map[uint]*forest.ForestInfoItemstruct, len(out.Data))
	for _, item := range out.Data {
		forestMap[item.ID] = item
	}

	resp.Response.Offset = 0
	resp.Response.Limit = len(req.Request.ForestIDs)
	resp.Response.Data = make([]*dtokeapi.ForestItem, 0, len(req.Request.ForestIDs))
	for _, forestID := range req.Request.ForestIDs {
		item, ok := forestMap[forestID]
		if !ok {
			continue
		}
		resp.Response.Data = append(resp.Response.Data, dtokeapi.NewForestItem(item))
	}
	resp.Response.Total = int64(len(resp.Response.Data))
}

// CreateForest 创建知识库
func CreateForest(ctx *gin.Context, req *dtokeapi.CreateForestRequest, resp *dtokeapi.CreateForestResponse) {
	if !req.ValidCreateForest(&resp.BaseResponse) {
		return
	}

	dataSourceType := foresttype.ForestDataSourceStandard
	dataSourceSubtype := foresttype.ForestDataSourceSubtypeStandard
	if req.Request.ForestType == foresttype.ForestTypeData {
		dataSourceType = foresttype.ForestDataSourceExcel
		dataSourceSubtype = foresttype.ForestDataSourceSubtypeExcel
	}

	forestID, err := svcforest.CreateForest(ctx, &svcforest.CreateForestRequest{
		Uin:               runtime.Uin(ctx),
		CompanyID:         runtime.CompanyID(ctx),
		Name:              req.Request.Name,
		AvatarURL:         req.Request.AvatarURL,
		Description:       req.Request.Description,
		PublicScope:       foresttype.PublicScopePrivate,
		ForestType:        req.Request.ForestType,
		DataSourceType:    dataSourceType,
		DataSourceSubtype: dataSourceSubtype,
	})
	if err != nil {
		if errors.Is(err, svcforest.ErrForestNameExists) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_forest_name_exists"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_forest_failed"
		return
	}
	resp.Response.ForestID = forestID
}

// UpdateForest 更新知识库
func UpdateForest(ctx *gin.Context, req *dtokeapi.UpdateForestRequest, resp *apiobj.BaseResponse) {
	if !req.ValidUpdateForest(resp) {
		return
	}

	uin := runtime.Uin(ctx)
	_, err := forest.GetForestByID(ctx, req.Request.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
		return
	}
	if !perm.HasManageAct(ctx, uin, req.Request.ForestID, foresttype.ResourceTypeForest) {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission_update_resource"
		return
	}

	var name *string
	var description *string
	if req.Request.Name != "" {
		name = &req.Request.Name
	}
	if req.Request.Description != "" {
		description = &req.Request.Description
	}

	err = svcforest.UpdateForest(ctx, &svcforest.UpdateForestRequest{
		ForestID:    req.Request.ForestID,
		Name:        name,
		Description: description,
	})
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, svcforest.ErrGetForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed"
	case errors.Is(err, svcforest.ErrForestNameExists):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_forest_name_exists"
	case errors.Is(err, svcforest.ErrModifyForestFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_resource_failed"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_resource_failed"
	}
}

// DeleteForest 删除知识库
func DeleteForest(ctx *gin.Context, req *dtokeapi.DeleteForestRequest, resp *apiobj.BaseResponse) {
	if !req.ValidDeleteForest(resp) {
		return
	}
	err := svcforest.DeleteForest(ctx, &svcforest.DeleteForestRequest{
		Uin:      runtime.Uin(ctx),
		ForestID: req.Request.ForestID,
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
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission"
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
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_coze_mapping_failed"
	case errors.Is(err, svcforest.ErrDeleteMappingFailed):
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_delete_mapping_failed"
	default:
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_forest_failed"
	}
}
