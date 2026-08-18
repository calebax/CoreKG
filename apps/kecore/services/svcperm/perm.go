package svcperm

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoperm"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

func SetResourcePerm(ctx *gin.Context, req *dtoperm.SetResourcePermRequest) (res *dtoperm.SetResourcePermResponse, err error) {
	res = &dtoperm.SetResourcePermResponse{}

	if err := forest.NewAccessProvider(ctx, &forest.ContextModel{
		ResourceID:   req.Request.ResourceID,
		ResourceType: req.Request.ResourceType,
		Opt:          &req.Request.PermOption,
	}).Apply(ctx); err != nil {
		logs.ErrorContextf(ctx, "SetResourcePerm: Apply failed: %w", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_set_resource_perm_failed"
		return res, nil
	}

	return res, nil
}

func GetResourcePerm(ctx *gin.Context, req *dtoperm.GetResourcePermRequest) (res *dtoperm.GetResourcePermResponse, err error) {
	res = &dtoperm.GetResourcePermResponse{}

	var r *forest.AccessResult
	r, err = forest.NewAccessProvider(ctx, &forest.ContextModel{
		ResourceID:   req.Request.ResourceID,
		ResourceType: req.Request.ResourceType,
	}).Get(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "GetResourcePerm: Get failed: %w", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_get_resource_perm_failed"
		return res, nil
	}

	res.Response = dtoperm.GetResourcePermEmbedResponse{
		AccessResult: dtoperm.AccessResult{
			ManagerList: r.ManagerList,
			ViewerList:  r.ViewerList,
			BanList:     r.BanList,
			ScopeType:   r.ScopeType,
		},
	}
	return res, nil
}
