package graphctl

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

// CreateGraph 创建知识图谱
// @Tags 知识森林知识图谱
// @Summary 创建知识图谱
// @Description 创建知识图谱
// @Router /forest.CreateGraph [post]
// @Param user body CreateGraphRequest true "入参"
// @Success 200 {object} CreateGraphResponse "返回值"
func CreateGraph(ctx *gin.Context, req *CreateGraphRequest, resp *CreateGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "CreateGraph.Validity failed: %s", resp.Message)
		return
	}
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	graphInfo := &foresttype.ForestGraphInfo{
		Uin:         uin,
		CompanyID:   companyID,
		Name:        req.Request.Name,
		Description: req.Request.Description,
		PublicScope: req.Request.PublicScope,
		AvatarUrl:   req.Request.AvatarUrl,
	}

	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		err := graph.CreateGraph(ctx, graphInfo, tx)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_create_graph_failed" // 创建图谱失败
			logs.ErrorContextf(ctx, "CreateGraph.CreateGraph failed: %v", err)
			return err
		}
		resp.Response.ForestGraphInfo = graphInfo

		//do not append manager action list with admins in company
		if req.Request.PublicScope == foresttype.PublicScopeCompany {
			req.Request.ScopeIDs = types.NewUintArray([]uint{})
		}

		// 去重
		req.Request.ManagerIDs.RemoveDuplicates()
		req.Request.ScopeIDs.RemoveDuplicates()

		uins := types.NewUintArray(append(req.Request.ManagerIDs.Slice(), req.Request.ScopeIDs.Slice()...))
		uins.RemoveDuplicates()

		us := uins.Slice()
		if !employee.CheckUinsValid(ctx, us, companyID) {
			logs.ErrorContextf(ctx, "CheckUinsValid: exist no-local company[%v] uin in uins[%v]", companyID, us)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_invalid_employee_id")) // 存在非法员工id
			return err
		}

		return perm.UpdateResourceScope(ctx, tx, graphInfo.ID, foresttype.ResourceTypeGraph, req.Request.ScopeIDs.Slice(), req.Request.ManagerIDs.Slice(), req.Request.PublicScope, companyID)
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateGraph.CreateGraph failed: %v", err)
		return
	}
}

// UpdateGraph 更新图谱
// @Tags 知识森林知识图谱
// @Summary 更新图谱
// @Description 更新图谱
// @Router /forest.UpdateGraph [post]
// @Param user body UpdateGraphRequest true "入参"
// @Success 200 {object} UpdateGraphResponse "返回值"
func UpdateGraph(ctx *gin.Context, req *UpdateGraphRequest, resp *UpdateGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "UpdateGraph.Validity failed: %s", resp.Message)
		return
	}
	// 查看图谱状态
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraph err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图谱信息失败
		return
	}
	if req.Request.Name != "" {
		graphInfo.Name = req.Request.Name
	}
	if req.Request.Description != "" {
		graphInfo.Description = req.Request.Description
	}
	graphInfo.ParseMode = req.Request.ParseMode
	graphInfo.PublicScope = req.Request.PublicScope
	if len(req.Request.FileIDList) > 0 {
		graphInfo.FileIDList = types.NewUintArray(req.Request.FileIDList)
	}
	// companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	if !perm.HasManageAct(ctx, uin, req.Request.GraphID, foresttype.ResourceTypeGraph) {
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_permission_update_resource")) // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, req.Request.GraphID)
		return
	}

	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err = graph.UpdateGraph(ctx, graphInfo, tx); err != nil {
			logs.ErrorContextf(ctx, "UpdateGraph failed err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_update_graph_failed" // 更新图谱状态失败
			return err
		}
		// ss := req.Request.ScopeIDs.Slice()
		// ms := req.Request.ManagerIDs.Slice()
		// if len(ss) != 0 && len(ms) != 0 {
		// 	//do not append manager action list with admins in company
		// 	if req.Request.PublicScope == foresttype.PublicScopeCompany {
		// 		req.Request.ScopeIDs = types.NewUintArray([]uint{})
		// 	}

		// 	// 去重
		// 	req.Request.ManagerIDs.RemoveDuplicates()
		// 	req.Request.ScopeIDs.RemoveDuplicates()

		// 	uins := types.NewUintArray(append(req.Request.ManagerIDs.Slice(), req.Request.ScopeIDs.Slice()...))
		// 	uins.RemoveDuplicates()

		// 	us := uins.Slice()
		// 	if !employee.CheckUinsValid(ctx, us, companyID) {
		// 		logs.ErrorContextf(ctx, "CheckUinsValid: exist no-local company[%v] uin in uins[%v]", companyID, us)
		// 		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_invalid_employee_id")) // 存在非法员工id
		// 		return err
		// 	}

		// 	return perm.UpdateResourceScope(tx, graphInfo.ID, foresttype.ResourceTypeGraph, req.Request.ScopeIDs.Slice(), req.Request.ManagerIDs.Slice(), req.Request.PublicScope, companyID)
		// }
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "UpdateGraph failed: %v", err)
		return
	}
}

// CreateTag 创建实体类型
// @Tags 知识森林知识图谱
// @Summary 创建实体类型
// @Description 创建实体类型
// @Router /forest.CreateTag [post]
// @Param user body CreateTagRequest true "入参"
// @Success 200 {object} CreateTagResponse "返回值"
func CreateTag(ctx *gin.Context, req *CreateTagRequest, resp *CreateTagResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "CreateTag.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}

	tag := &foresttype.GraphTag{
		Uin:            runtime.Uin(ctx),
		CompanyID:      runtime.CompanyID(ctx),
		GraphVersionID: graphInfo.VersionID,
		GraphID:        req.Request.GraphID,
		TagName:        req.Request.TagName,
		TagType:        foresttype.TagTypeNode,
		Description:    req.Request.Description,
		Properties:     req.Request.Properties,
	}
	err = graph.CreateTag(ctx, graphInfo.SpaceName, tag)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_tag_failed" // 创建失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
	resp.Response.GraphTag = tag
}

// UpdateTag 修改实体类型
// @Tags 知识森林知识图谱
// @Summary 修改实体类型
// @Description 修改实体类型
// @Router /forest.UpdateTag [post]
// @Param user body UpdateTagRequest true "入参"
// @Success 200 {object} UpdateTagResponse "返回值"
func UpdateTag(ctx *gin.Context, req *UpdateTagRequest, resp *UpdateTagResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "UpdateTag.Validity failed: %s", resp.Message)
		return
	}
	// 暂不修改图谱Tag
	// graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	// if err != nil {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
	// 	logs.ErrorContextf(ctx, "UpdateTag.GetGraph failed: %v", err)
	// 	return
	// }
	// if graphInfo.Status != foresttype.GraphStatusDraft {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_graph_not_draft" // 图谱不是草稿状态无法修改
	// 	return
	// }
	// cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "获取图信息失败"
	// 	return
	// }
	// defer cli.Release()
	tag, err := graph.GetTagByID(ctx, req.Request.TagID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_tag_failed" // 获取实体类型失败
		logs.ErrorContextf(ctx, "UpdateTag.GetTagByID failed: %v", err)
		return
	}
	if req.Request.TagName != "" {
		tag.TagName = req.Request.TagName
	}
	if req.Request.Description != "" {
		tag.Description = req.Request.Description
	}

	tag.Properties = req.Request.Properties

	//err = cli.AlterTag(tag, true)
	//if err != nil {
	//	resp.Code = errcode.ErrCode_InternalError
	//	resp.Message = "修改实体失败"
	//	logs.ErrorContextf(ctx, "UpdateTag.AlterTag failed: %v", err)
	//	return
	//}

	err = graph.UpdateTag(ctx, tag)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_tag_failed" // 修改实体类型失败
		logs.ErrorContextf(ctx, "UpdateTag.UpdateTag failed: %v", err)
		return
	}
	// time.Sleep(time.Second * 22)
	resp.Response.GraphTag = tag
}

// DeleteTag 删除实体类型
// @Tags 知识森林知识图谱
// @Summary 删除实体类型
// @Description 删除实体类型
// @Router /forest.DeleteTag [post]
// @Param user body DeleteTagRequest true "入参"
// @Success 200 {object} DeleteTagResponse "返回值"
func DeleteTag(ctx *gin.Context, req *DeleteTagRequest, resp *DeleteTagResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "DeleteTag.Validity failed: %s", resp.Message)
		return
	}
	// graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	// if err != nil {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
	// 	logs.ErrorContextf(ctx, "DeleteTag.GetGraph failed: %v", err)
	// 	return
	// }
	// if graphInfo.Status != foresttype.GraphStatusDraft {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_graph_not_draft" // 图谱不是草稿状态无法修改
	// 	return
	// }
	// cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "获取图信息失败"
	// 	return
	// }
	// defer cli.Release()
	tag, err := graph.GetTagByID(ctx, req.Request.TagID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_tag_failed" // 获取实体类型失败
		logs.ErrorContextf(ctx, "DeleteTag.GetTagByID failed: %v", err)
		return
	}

	// err = cli.DropGraphTag(tag)
	// if err != nil {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "修改实体失败"
	// 	logs.ErrorContextf(ctx, "DeleteTag.AlterTag failed: %v", err)
	// 	return
	// }
	err = graph.DeleteTag(ctx, tag.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_tag_failed" // 修改实体类型失败
		logs.ErrorContextf(ctx, "DeleteTag.UpdateTag failed: %v", err)
		return
	}
	resp.Response.GraphTag = tag
}

// CreateEdge 创建边
// @Tags 知识森林知识图谱
// @Summary 创建边
// @Description 创建边
// @Router /forest.CreateEdge [post]
// @Param user body CreateEdgeRequest true "入参"
// @Success 200 {object} CreateEdgeResponse "返回值"
func CreateEdge(ctx *gin.Context, req *CreateEdgeRequest, resp *CreateEdgeResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "CreateTag.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
	// if graphInfo.Status != foresttype.GraphStatusDraft {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_graph_not_draft" // 图谱不是草稿状态无法修改
	// 	return
	// }
	edge, err := graph.GetEdgeByName(ctx, req.Request.GraphID, graphInfo.VersionID, req.Request.EdgeName)
	if err != nil && err != gorm.ErrRecordNotFound {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_edge_failed" // 获取边信息失败
		logs.ErrorContextf(ctx, "CreateEdge.CreateTag failed: %v", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		edge = &foresttype.GraphTag{
			Uin:            runtime.Uin(ctx),
			CompanyID:      runtime.CompanyID(ctx),
			GraphID:        req.Request.GraphID,
			GraphVersionID: graphInfo.VersionID,
			TagName:        req.Request.EdgeName,
			TagType:        foresttype.TagTypeEdge,
		}
		err = graph.CreateTag(ctx, graphInfo.SpaceName, edge)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_create_edge_failed" // 创建边失败
			logs.ErrorContextf(ctx, "CreateEdge.CreateTag failed: %v", err)
			return
		}
	}

	_, err = graph.GetEdgeTag(ctx, edge.ID, req.Request.SrcTagID, req.Request.DstTagID)
	if err != nil && err != gorm.ErrRecordNotFound {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_edge_failed" // 获取边信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		et := &foresttype.GraphEdgeTag{
			GraphID:        req.Request.GraphID,
			GraphVersionID: graphInfo.VersionID,
			EdgeTypeID:     edge.ID,
			SrcTagID:       req.Request.SrcTagID,
			DstTagID:       req.Request.DstTagID,
		}
		err = dbutil.Knownow().Create(et).Error
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_create_edge_failed" // 创建边失败
			logs.ErrorContextf(ctx, "CreateEdge.CreateGraphEdgeTag failed: %v", err)
			return
		}
		return
	}
	resp.Code = errcode.ErrCode_InternalError
	resp.Message = "kecore_edge_exists" // 该关系已存在
}

// DeleteEdge 删除边
// @Tags 知识森林知识图谱
// @Summary 删除边
// @Description 删除边
// @Router /forest.DeleteEdge [post]
// @Param user body DeleteEdgeRequest true "入参"
// @Success 200 {object} DeleteEdgeResponse "返回值"
func DeleteEdge(ctx *gin.Context, req *DeleteEdgeRequest, resp *DeleteEdgeResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "DeleteEdge.Validity failed: %s", resp.Message)
		return
	}
	// graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	// if err != nil {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
	// 	logs.ErrorContextf(ctx, "DeleteEdge.GetGraph failed: %v", err)
	// 	return
	// }
	// if graphInfo.Status != foresttype.GraphStatusDraft {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_graph_not_draft" // 图谱不是草稿状态无法修改
	// 	return
	// }
	et, err := graph.GetEdgeTagByID(ctx, req.Request.EgdeID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_edge_failed" // 获取边信息失败
		logs.ErrorContextf(ctx, "DeleteEdge.GetEdgeTagByID failed: %v", err)
		return
	}
	err = graph.DeleteEdgeTag(ctx, et.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_edge_failed" // 删除边失败
		logs.ErrorContextf(ctx, "DeleteEdge.DeleteEdgeTag failed: %v", err)
		return
	}
	etlist, err := graph.ListEdgeTagByEdgeID(ctx, et.EdgeTypeID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_edge_failed" // 查询实体边失败
		logs.ErrorContextf(ctx, "DeleteEdge.ListEdgeTagByEdgeID failed: %v", err)
		return
	}
	if len(etlist) == 0 {
		// 删除边
		// cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
		// if err != nil {
		// 	logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
		// 	resp.Code = errcode.ErrCode_InternalError
		// 	resp.Message = "kecore_get_graph_info_failed" // 获取图谱信息失败
		// 	return
		// }
		// defer cli.Release()
		tag, err := graph.GetTagByID(ctx, et.EdgeTypeID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_edge_type_failed" // 获取边类型失败
			logs.ErrorContextf(ctx, "DeleteEdge.GetTagByID failed: %v", err)
			return
		}
		// err = cli.DropGraphTag(tag)
		// if err != nil {
		// 	resp.Code = errcode.ErrCode_InternalError
		// 	resp.Message = "kecore_update_edge_failed" // 修改边失败
		// 	logs.ErrorContextf(ctx, "DeleteTag.AlterTag failed: %v", err)
		// 	return
		// }
		err = graph.DeleteTag(ctx, tag.ID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_update_edge_type_failed" // 修改边类型失败
			logs.ErrorContextf(ctx, "DeleteTag.UpdateTag failed: %v", err)
			return
		}
	}
}

// UpdateEdge 修改边
// @Tags 知识森林知识图谱
// @Summary 修改边
// @Description 修改边
// @Router /forest.UpdateEdge [post]
// @Param user body UpdateEdgeRequest true "入参"
// @Success 200 {object} UpdateEdgeResponse "返回值"
func UpdateEdge(ctx *gin.Context, req *UpdateEdgeRequest, resp *UpdateEdgeResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "UpdateEdge.Validity failed: %s", resp.Message)
		return
	}
	// graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	// if err != nil {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
	// 	logs.ErrorContextf(ctx, "UpdateEdge.GetGraph failed: %v", err)
	// 	return
	// }
	// if graphInfo.Status != foresttype.GraphStatusDraft {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_graph_not_draft" // 图谱不是草稿状态无法修改
	// 	return
	// }
	et, err := graph.GetEdgeTagByID(ctx, req.Request.EdgeID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_edge_failed" // 获取边信息失败
		logs.ErrorContextf(ctx, "UpdateEdge.GetEdgeTagByID failed: %v", err)
		return
	}
	_, err = graph.GetEdgeTag(ctx, et.EdgeTypeID, req.Request.SrcTagID, req.Request.DstTagID)
	if err != nil && err != gorm.ErrRecordNotFound {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_edge_failed" // 获取边信息失败
		logs.ErrorContextf(ctx, "UpdateEdge.GetEdgeTag failed: %v", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		et.SrcTagID = req.Request.SrcTagID
		et.DstTagID = req.Request.DstTagID
		err = graph.UpdateEdgeTag(ctx, et)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_update_edge_failed" // 修改边失败
			logs.ErrorContextf(ctx, "UpdateEdge.UpdateEdgeTag failed: %v", err)
			return
		}
		return
	}
	resp.Code = errcode.ErrCode_InternalError
	resp.Message = "kecore_edge_exists" // 当前类型的边已存在
}

// ListForestGraph 获取图谱列表
// @Tags 知识森林知识图谱
// @Summary 获取图谱列表
// @Description 获取图谱列表
// @Router /forest.ListForestGraph [post]
// @Param user body ListForestGraphRequest true "入参"
// @Success 200 {object} ListForestGraphResponse "返回值"
func ListForestGraph(ctx *gin.Context, req *ListForestGraphRequest, resp *ListForestGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListForestGraph.Validity failed: %s", resp.Message)
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	req.Request.Uin = runtime.Uin(ctx)
	res, err := graph.ListForestGraph(ctx, req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_list_failed" // 查询知识森林列表失败
		logs.ErrorContextf(ctx, "ListForestGraph.ListForestGraph failed: %v", err)
		return
	}
	versionIDs := []uint{}
	for _, v := range res.Data {
		versionIDs = append(versionIDs, v.VersionID)
	}
	countMap, err := coretask.ListGraphTaskCount(ctx, versionIDs)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_list_failed"
		logs.ErrorContextf(ctx, "ListForestGraph.ListGraphTaskCount failed: %v", err)
		return
	}
	for i, v := range res.Data {
		if _, ok := countMap[v.VersionID]; !ok {
			continue
		}
		res.Data[i].TaskCount = countMap[v.VersionID].Count
		res.Data[i].SuccessTaskCount = countMap[v.VersionID].SuccessCount
	}
	resp.Response = res
}

// GetGraphInfo 获取图谱详情
// @Tags 知识森林知识图谱
// @Summary 获取图谱详情
// @Description 获取图谱详情
// @Router /forest.GetGraphInfo [post]
// @Param user body GetGraphInfoRequest true "入参"
// @Success 200 {object} GetGraphInfoResponse "返回值"
func GetGraphInfo(ctx *gin.Context, req *GetGraphInfoRequest, resp *GetGraphInfoResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListForestGraph.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "GetGraphInfo.GetGraph failed: %v", err)
		return
	}
	count, err := coretask.GetGraphTaskCount(ctx, graphInfo.VersionID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_task_count_failed"
		logs.ErrorContextf(ctx, "GetGraphInfo.GetGraphTaskCount failed: %v", err)
		return
	}
	graphInfo.TaskCount = count.Count
	graphInfo.SuccessTaskCount = count.SuccessCount
	resp.Response.ForestGraphInfo = graphInfo
	var (
		scopeIDs, managerIDs []uint
		rss                  []*foresttype.KeResourceScope
	)

	if err := dbutil.Knownow().
		Where("deleted_at IS NULL").
		Where("resource_type", foresttype.ResourceTypeGraph).
		Where("resource_id = ?", graphInfo.ID).
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

	resp.Response.ManagerIDs = types.NewUintArray(managerIDs)
	resp.Response.ScopeIDs = types.NewUintArray(scopeIDs)
	resp.Response.IsAdmin = isAdmin
}

// DeleteGraph 删除图谱
// @Tags 知识森林知识图谱
// @Summary 删除图谱
// @Description 删除图谱
// @Router /forest.DeleteGraph [post]
// @Param user body DeleteGraphRequest true "入参"
// @Success 200 {object} DeleteGraphResponse "返回值"
func DeleteGraph(ctx *gin.Context, req *DeleteGraphRequest, resp *DeleteGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListForestGraph.Validity failed: %s", resp.Message)
		return
	}
	// 判断当前知识库是否已经存在图谱
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed"
		return
	}
	// 运行中的图谱可以删除
	// if graphInfo.Status == foresttype.GraphStatusRunning || graphInfo.Status == foresttype.GraphStatusPending {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "任务正在运行，请稍候再试"
	// 	return
	// }
	forestInfo, err := forest.GetForestByID(ctx, graphInfo.ForestID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_faileds"
		return
	}
	forestInfo.GraphStatus = foresttype.GraphStatusUnCreated
	err = dbutil.Knownow().Save(forestInfo).Error
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_forest_failed"
		return
	}
	err = coretask.DeleteTasksByGraphVersion(ctx, graphInfo.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "coretask.DeleteTasksByGraphVersion failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_forest_failed"
		return
	}
	err = graph.DeleteGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
}

// ListGraphTag 获取图谱Tag
// @Tags 知识森林知识图谱
// @Summary 获取图谱Tag
// @Description 获取图谱Tag
// @Router /forest.ListGraphTag [post]
// @Param user body ListGraphTagRequest true "入参"
// @Success 200 {object} ListGraphTagResponse "返回值"
func ListGraphTag(ctx *gin.Context, req *ListGraphTagRequest, resp *ListGraphTagResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListGraphTag.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	res, err := graph.ListTag(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.PageQuery)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_forest_list_failed" // 查询知识森林列表失败
		logs.ErrorContextf(ctx, "ListGraphTag.ListGraphTag failed: %v", err)
		return
	}
	resp.Response = res
}

// ListGraphNode 获取图谱实体列表
// @Tags 知识森林知识图谱
// @Summary 获取图谱实体列表
// @Description 获取图谱实体列表
// @Router /forest.ListGraphNode [post]
// @Param user body ListGraphNodeRequest true "入参"
// @Success 200 {object} ListGraphNodeResponse "返回值"
func ListGraphNode(ctx *gin.Context, req *ListGraphNodeRequest, resp *ListGraphNodeResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "[ListGraphNode] Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}

	type ListNodeItem struct {
		NodeID   uint   `gorm:"column:node_id"`
		NodeName string `gorm:"column:node_name"`
	}

	var nodeEntityList []ListNodeItem

	// query := dbutil.Knownow().WithContext(ctx).Table(fmt.Sprintf("%s AS n", foresttype.TableNameKeGraphNode)).
	// 	Select("DISTINCT n.id AS node_id, n.name AS node_name").
	// 	Joins(fmt.Sprintf("JOIN %s AS t ON t.node_id = n.id", foresttype.TableNameKeGraphTagNode)).
	// 	Where("n.deleted_at IS NULL").
	// 	Where("t.deleted_at IS NULL").
	// 	Where("t.graph_id = ?", req.Request.GraphID).
	// 	Where("t.graph_version_id = ?", graphInfo.VersionID).
	// 	Where("n.graph_id = ?", req.Request.GraphID).
	// 	Where("n.graph_version_id = ?", graphInfo.VersionID)

	query := dbutil.Knownow().WithContext(ctx).
		Table(fmt.Sprintf("%s AS n", foresttype.TableNameKeGraphTagNode)).
		Select("DISTINCT n.id AS node_id, n.name AS node_name").
		Where("n.deleted_at IS NULL").
		Where("n.graph_id = ?", req.Request.GraphID).
		Where("n.graph_version_id = ?", graphInfo.VersionID)

	if req.Request.GraphTagID > 0 {
		query = query.Where("n.tag_id = ?", req.Request.GraphTagID)
	}
	if req.Request.GraphNodeName != "" {
		query = query.Where("n.name LIKE ?", "%"+req.Request.GraphNodeName+"%")
	}

	if err := query.Find(&nodeEntityList).Error; err != nil {
		logs.ErrorContextf(ctx, "[ListGraphNode] query.Find failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_edge_failed" // 查询实体边失败
		return
	}

	list := make([]ListGraphNodeItem, 0, len(nodeEntityList))
	for _, v := range nodeEntityList {
		list = append(list, ListGraphNodeItem{
			GraphNodeID:   v.NodeID,
			GraphNodeName: v.NodeName,
		})
	}

	resp.Response.List = list
	resp.Response.Total = int64(len(list))
}

// GetKnowledgeGraph 获取知识图谱
// @Tags 知识森林知识图谱
// @Summary 获取知识图谱
// @Description 获取知识图谱
// @Router /forest.GetKnowledgeGraph [post]
// @Param user body GetKnowledgeGraphRequest true "入参"
// @Success 200 {object} GetKnowledgeGraphResponse "返回值"
func GetKnowledgeGraph(ctx *gin.Context, req *GetKnowledgeGraphRequest, resp *GetKnowledgeGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListGraphTag.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		return
	}
	defer cli.Release()
	knowledgeGraph, err := cli.GetKnowledgeGraph(req.Request.KnowledgeGraphReq)
	if err != nil {
		logs.ErrorContextf(ctx, "GetKnowledgeGraph error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		return
	}
	resp.Response.KnowledgeGraph = knowledgeGraph
}

// GetTagEdge 获取tag的边
// @Tags 知识森林知识图谱
// @Summary 获取tag的边
// @Description 获取tag的边
// @Router /forest.GetTagEdge [post]
// @Param user body GetTagEdgeRequest true "入参"
// @Success 200 {object} GetTagEdgeResponse "返回值"
func GetTagEdge(ctx *gin.Context, req *GetTagEdgeRequest, resp *GetTagEdgeResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListGraphTag.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "CreateTag.CreateTag failed: %v", err)
		return
	}
	res, err := graph.GetEdgeTagInfoByTagID(ctx, graphInfo.ID, req.Request.TagID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_edge_failed" // 获取边信息失败
		logs.ErrorContextf(ctx, "ListGraphTag.ListGraphTag failed: %v", err)
		return
	}
	resp.Response.Data = res
}

// SubmitTemplate 提交模板
// @Tags 知识森林知识图谱
// @Summary 提交模板
// @Description 提交模板
// @Router /forest.SubmitTemplate [post]
// @Param user body SubmitTemplateRequest true "入参"
// @Success 200 {object} SubmitTemplateResponse "返回值"
func SubmitTemplate(ctx *gin.Context, req *SubmitTemplateRequest, resp *SubmitTemplateResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ListGraphTag.Validity failed: %s", resp.Message)
		return
	}
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图信息失败
		logs.ErrorContextf(ctx, "SubmitTemplate.GetGraph failed: %v", err)
		return
	}
	err = graph.DeleteGraphTempData(ctx, graphInfo)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_graph_failed" // 更新图谱状态失败
		logs.ErrorContextf(ctx, "SubmitTemplate.InsertTemplate failed: %v", err)
		return
	}

	err = graph.InsertTemplate(ctx, graphInfo, req.Request.Template)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_graph_failed" // 创建图谱失败
		logs.ErrorContextf(ctx, "SubmitTemplate.InsertTemplate failed: %v", err)
		return
	}
}
