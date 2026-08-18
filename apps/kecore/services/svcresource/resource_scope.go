package svcresource

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/account"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoresource"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func SetResourceScope(ctx *gin.Context, req *dtoresource.SetResourceScopeRequest) (res *dtoresource.SetResourceScopeResponse, err error) {
	res = &dtoresource.SetResourceScopeResponse{
		Response: dtoresource.SetResourceScopeEmbedResponse{
			ResourceType:  req.Request.ResourceType,
			ResourceID:    req.Request.ResourceID,
			ResourceIDStr: req.Request.ResourceIDStr,
		},
	}

	companyID := runtime.CompanyID(ctx)

	dbResourceType, resourceTypeExists := foresttype.ResourceTypeMap[req.Request.ResourceType]
	if !resourceTypeExists || dbResourceType == "" {
		logs.WarnContextf(ctx, "[SetResourceScope] invalid resourceType: %v", req.Request.ResourceType)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_permission_invalid_resource_type"
		return res, nil
	}

	// 获取对应资源类型的处理器（可能为 nil）
	handler, handlerExists := GetHandler(dbResourceType)
	if handlerExists {
		if handler != nil {
			if err := handler.BeforeSetScope(ctx, req); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceScope] BeforeSetScope failed: %v", err)
				return nil, err
			}
		}
	}

	// 构建需要添加的权限列表
	var addList foresttype.KeResourceScopeList

	// 构建查看权限列表
	switch req.Request.ViewScopeType {
	case foresttype.ScopeTypeUser:
		for _, scopeID := range req.Request.ViewScopeIDs {
			addList = append(addList, foresttype.KeResourceScope{
				ResourceType: dbResourceType,
				ResourceID:   req.Request.ResourceID,
				ScopeType:    req.Request.ViewScopeType,
				ScopeID:      scopeID,
				Action:       foresttype.ActionView,
			})
		}

	case foresttype.ScopeTypeCompany:
		addList = append(addList, foresttype.KeResourceScope{
			ResourceType: dbResourceType,
			ResourceID:   req.Request.ResourceID,
			ScopeType:    req.Request.ViewScopeType,
			ScopeID:      companyID,
			Action:       foresttype.ActionView,
		})
	}

	// 添加管理权限
	for _, scopeID := range req.Request.ManageScopeIDs {
		addList = append(addList, foresttype.KeResourceScope{
			ResourceType: dbResourceType,
			ResourceID:   req.Request.ResourceID,
			ScopeType:    req.Request.ViewScopeType,
			ScopeID:      scopeID,
			Action:       foresttype.ActionManage,
		})
	}

	// 在事务中执行权限更新
	err = dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除该资源的所有已存在权限
		deleteCond := &forest.KeResourceScopeCond{
			ResourceType: dbResourceType,
			ResourceID:   req.Request.ResourceID,
		}
		if err := forest.NewKeResourceScopeDao().WithTx(tx).DeleteByCond(ctx, deleteCond); err != nil {
			return err
		}

		// 批量插入新权限
		if len(addList) > 0 {
			if err := forest.NewKeResourceScopeDao().WithTx(tx).BatchInsert(ctx, addList); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if handlerExists {
		// 如果有处理器，执行后置钩子
		if handler != nil {
			if err := handler.AfterSetScope(ctx, req, addList); err != nil {
				logs.ErrorContextf(ctx, "[SetResourceScope] AfterSetScope failed: %v", err)
				// 后置钩子失败只记录日志，不影响主流程
			}
		}
	}

	return res, nil
}

func GetResourceScope(ctx *gin.Context, req *dtoresource.GetResourceScopeRequest) (res *dtoresource.GetResourceScopeResponse, err error) {
	res = &dtoresource.GetResourceScopeResponse{}

	dbResourceType, ok := foresttype.ResourceTypeMap[req.Request.ResourceType]
	if !ok || dbResourceType == "" {
		logs.WarnContextf(ctx, "[GetResourceScope] invalid resourceType: %v", req.Request.ResourceType)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_permission_invalid_resource_type"
		return res, nil
	}

	scopeCond := &forest.KeResourceScopeCond{
		ResourceType: dbResourceType,
		ResourceIDs:  req.Request.ResourceIDs,
	}

	// 一次性查询所有资源的权限数据
	resourceScopeEntityList, err := forest.NewKeResourceScopeDao().GetListByCond(ctx, scopeCond)
	if err != nil {
		return nil, err
	}
	validResourceScopeMap := make(map[uint]struct{})
	var uins []uint
	for _, v := range resourceScopeEntityList {
		validResourceScopeMap[v.ResourceID] = struct{}{}
		if v.ScopeType == foresttype.ScopeTypeUser {
			uins = append(uins, v.ScopeID)
		}
	}
	uins = utils.SliceDuplicate(uins)

	uinEntityList, err := account.NewUserIdentificationDao().GetListByCond(ctx, &account.UserIdentificationCond{
		IDs: uins,
	})
	if err != nil {
		return nil, err
	}
	uinList := make([]dtoresource.UinListItem, 0, len(uinEntityList))
	for _, v := range uinEntityList {
		uinList = append(uinList, dtoresource.UinListItem{
			Uin:  v.ID,
			Name: v.Name,
		})
	}

	// 按资源ID分组整理权限数据，使用指针避免值拷贝导致 append 未写回 map 的问题
	resourceScopeMap := make(map[uint]*dtoresource.ResourceScopeItem)

	for _, resourceID := range req.Request.ResourceIDs {
		if _, exists := validResourceScopeMap[resourceID]; !exists {
			continue
		}
		resourceScopeMap[resourceID] = &dtoresource.ResourceScopeItem{
			ResourceType:   req.Request.ResourceType,
			ResourceID:     resourceID,
			ResourceIDStr:  strconv.FormatUint(uint64(resourceID), 10),
			ViewScopeType:  "",
			ViewScopeIDs:   make([]uint, 0),
			ManageScopeIDs: make([]uint, 0),
		}
	}

	// 遍历查询结果，按资源ID聚合 ViewScopeIDs / ManageScopeIDs（直接修改 map 中的指针，无需写回）
	for _, v := range resourceScopeEntityList {
		item, exists := resourceScopeMap[v.ResourceID]
		if !exists {
			continue // 请求中未包含该资源ID，跳过
		}
		if item.ViewScopeType == "" && v.ScopeType != "" {
			item.ViewScopeType = v.ScopeType
		}
		switch v.Action {
		case foresttype.ActionView:
			item.ViewScopeIDs = append(item.ViewScopeIDs, v.ScopeID)
		case foresttype.ActionManage:
			item.ManageScopeIDs = append(item.ManageScopeIDs, v.ScopeID)
		}
	}

	// 按请求的 ResourceIDs 顺序组装列表
	resourceScopeList := make([]dtoresource.ResourceScopeItem, 0, len(req.Request.ResourceIDs))
	for _, resourceID := range req.Request.ResourceIDs {
		if item := resourceScopeMap[resourceID]; item != nil {
			resourceScopeList = append(resourceScopeList, *item)
		}
	}

	res.Response.ResourceScopeList = resourceScopeList
	res.Response.UinList = uinList

	return res, nil
}
