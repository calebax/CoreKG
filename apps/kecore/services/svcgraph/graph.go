package svcgraph

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtograph"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateForestGraph 创建知识库专属图谱
func CreateForestGraph(ctx *gin.Context, req *dtograph.CreateForestGraphRequest) (res *dtograph.CreateForestGraphResponse, err error) {
	res = &dtograph.CreateForestGraphResponse{}
	forestInfo, err := forest.NewForestDao().GetByID(ctx, req.Request.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateForestGraph] get forest info failed, forest id: %d, err: %v", req.Request.ForestID, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_get_forest_failed"
		return res, nil
	}
	if forestInfo.ID == 0 {
		logs.WarnContextf(ctx, "[CreateForestGraph] forest not found, forest id: %d", req.Request.ForestID)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_forest_not_found"
		return res, nil
	}
	if forestInfo.KnowledgeStatus != foresttype.TaskStatusSuccess {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_knowledge_resource_not_generated"
		return
	}
	if forestInfo.GraphStatus != foresttype.GraphStatusUpdatable &&
		forestInfo.GraphStatus != foresttype.GraphStatusUnCreated &&
		forestInfo.GraphStatus != foresttype.GraphStatusSuccess {
		logs.WarnContextf(ctx, "ParseGraph failed graph status is not draft")
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_graph_not_draft" // 请勿重复操作图谱
		return
	}
	// 判断当前知识库是否已经存在图谱
	graphInfo, err := graph.GetForestGraph(ctx, req.Request.ForestID)
	if err != nil && err != gorm.ErrRecordNotFound {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_create_graph_failed"
		return res, nil
	}
	// 若不存在创建初始版本
	if err == gorm.ErrRecordNotFound {
		graphInfo = &foresttype.ForestGraphInfo{
			Uin:         forestInfo.Uin,
			CompanyID:   forestInfo.CompanyID,
			Name:        fmt.Sprintf("【%s】关联图谱", forestInfo.Name),
			Description: "暂无描述",
			PublicScope: forestInfo.PublicScope,
			ForestID:    req.Request.ForestID,
			AvatarUrl:   req.Request.AvatarUrl,
		}
		if err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
			managerIDS, scopeIDS, err := perm.GetALLManageScopeList(ctx, foresttype.ResourceTypeForest, forestInfo.ID)
			if err != nil {
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_get_permission_failed"
				logs.ErrorContextf(ctx, "GetManageScopeList failed: %v", err)
				return err
			}
			if len(managerIDS) == 0 {
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_get_permission_failed"
				logs.WarnContextf(ctx, "GetManageScopeList failed: managerIDS is empty, forest id: %d", req.Request.ForestID)
				return err
			}
			err = graph.CreateGraph(ctx, graphInfo, tx)
			if err != nil {
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_create_graph_failed" // 创建图谱失败
				logs.ErrorContextf(ctx, "CreateGraph.CreateGraph failed: %v", err)
				return err
			}
			forestInfo.GraphStatus = foresttype.GraphStatusDraft
			err = tx.WithContext(ctx).Save(forestInfo).Error
			if err != nil {
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_create_graph_failed" // 创建图谱失败
				logs.ErrorContextf(ctx, "CreateGraph.Save forestInfo failed: %v", err)
				return err
			}

			return perm.UpdateResourceScope(ctx, tx, graphInfo.ID, foresttype.ResourceTypeGraph, scopeIDS, managerIDS, graphInfo.PublicScope, graphInfo.CompanyID)
		}); err != nil {
			logs.ErrorContextf(ctx, "CreateGraph.CreateGraph failed: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_create_graph_failed"
			return res, nil
		}
		res.Response.Data = graphInfo
		return res, nil
	}
	// 若存在创建新版本
	if err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		err := graph.CreateGraphVersion(ctx, graphInfo, tx)
		if err != nil {
			return err
		}
		forestInfo.GraphStatus = foresttype.GraphStatusDraft
		err = tx.WithContext(ctx).Save(forestInfo).Error
		if err != nil {
			logs.ErrorContextf(ctx, "CreateGraph.Save forestInfo failed: %v", err)
			return err
		}
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateForestGraph.CreateGraphVersion failed: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_create_graph_failed"
		return res, nil
	}
	res.Response.Data = graphInfo
	return res, nil
}
