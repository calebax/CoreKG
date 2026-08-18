package svccollections

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtocollections"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func ListCollection(ctx *gin.Context, req *dtocollections.ListCollectionRequest) (res *dtocollections.ListCollectionResponse, err error) {
	res = &dtocollections.ListCollectionResponse{}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	resourceType := foresttype.ResourceTypeForestFile

	cond := &forest.UinCollectionsCond{
		BaseCond: forest.BaseCond{
			Uin:       uin,
			CompanyID: companyID,
			Offset:    req.Request.Offset,
			Limit:     req.Request.Limit,
			OrderBy:   req.Request.OrderBy,
			BeginTime: req.Request.BeginTime,
			EndTime:   req.Request.EndTime,
			Filters:   req.Request.Filters,
		},
	}

	for _, v := range req.Request.Filters {
		switch v.Field {
		case "resource_ids":
			ids := utils.Map(v.Value, func(v string) uint {
				if id, err := strconv.ParseUint(v, 10, 64); err == nil {
					return uint(id)
				}
				logs.ErrorContextf(ctx, "[ListCollection] invalid resource id: %s", v)
				return 0
			})
			cond.ResourceIDs = ids
		case "resource_type":
			cond.ResourceType = foresttype.ResourceType(v.Value[0])
			resourceType = cond.ResourceType
		default:
			logs.WarnContextf(ctx, "[ListCollection] invalid filter field: %s", v.Field)
			return res, fmt.Errorf("invalid filter field: %s", v.Field)
		}
	}

	ap, err := forest.NewAccessProvider(ctx, &forest.ContextModel{
		ResourceType: resourceType,
		ScopeType:    foresttype.ScopeTypeUser,
		ScopeID:      uin,
		Action:       foresttype.ActionBan,
	}).Action(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "filter ban list failed: %v", err)
		return nil, err
	}

	if len(ap.BanList) > 0 {
		cond.NotResourceIDs = ap.BanList
	}

	viewableForests, err := forest.ViewAbleForests(uin, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListCollection] ViewAbleForests failed, err: %v", err)
		return nil, err
	}
	if len(viewableForests) == 0 {
		logs.InfoContextf(ctx, "[ListCollection] no viewable forests")
		return res, nil
	}

	frsIDMap := make(map[uint]struct{})
	for _, frsID := range viewableForests {
		frsIDMap[frsID] = struct{}{}
	}

	list, c, err := forest.NewUinCollectionsDao().GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListCollection] NewUinCollectionsDao().GetPageListByCond failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_list_collection_fail"
		return res, err
	}

	res.Response.QueryResponse.Total = c
	res.Response.QueryResponse.Limit = req.Request.Limit
	res.Response.QueryResponse.Offset = req.Request.Offset

	if c == 0 {
		logs.InfoContextf(ctx, "[ListCollection] NewUinCollectionsDao().GetPageListByCond no data")
		return res, nil
	}

	resourceIDs := utils.Map(list, func(collection foresttype.UinCollections) uint {
		return collection.ResourceID
	})

	res.Response.Data = make([]dtocollections.ColItem, 0, len(list))

	switch resourceType {
	case foresttype.ResourceTypeForestFile:
		ffs, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
			BaseCond: forest.BaseCond{
				CompanyID: companyID,
			},
			IDs: resourceIDs,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[ListCollection] NewForestFileDao().GetListByCond failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_list_collection_fail"
			return res, err
		}

		ffsMap := ffs.ToMap()

		frsIDs := utils.Map(ffs, func(ff foresttype.KnownowForestFile) uint {
			if _, ok := frsIDMap[ff.ForestID]; ok {
				return ff.ForestID
			}
			return 0
		})

		frs, err := forest.NewForestDao().GetListByCond(ctx, &forest.ForestCond{
			BaseCond: forest.BaseCond{
				CompanyID: companyID,
			},
			IDs: frsIDs,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[ListCollection] NewForestDao().GetListByCond failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_list_collection_fail"
			return res, err
		}
		frsMap := frs.ToMap()

		var tags []dtocollections.ResourceTag
		if err := dbutil.Knownow().Table(foresttype.TableNameResourceTag+" AS rt").
			Where("rt.deleted_at IS NULL").
			Where("rt.resource_id IN (?)", resourceIDs).
			Where("rt.resource_type = ?", foresttype.TagResourceTypeFile).
			Where("tg.status = ?", foresttype.TagGroupStatusEnable).
			Joins("LEFT JOIN " + foresttype.TableNameTag + " AS t ON rt.tag_id = t.id AND t.deleted_at IS NULL").
			Joins("LEFT JOIN " + foresttype.TableNameTagGroup + " AS tg ON t.group_id = tg.id AND tg.deleted_at IS NULL").
			Select("rt.id, rt.resource_id, rt.resource_type, rt.tag_id, t.name as tag_name, tg.name as tag_group_name").
			Find(&tags).
			Error; err != nil {
			logs.ErrorContextf(ctx, "QueryForestFile failed to get tags: %v", err)
			return nil, err
		}

		fTagMap := make(map[uint][]dtocollections.ResourceTag)
		for _, tag := range tags {
			fTagMap[tag.ResourceID] = append(fTagMap[tag.ResourceID], tag)
		}

		for _, collection := range list {
			_, ffOk := ffsMap[collection.ResourceID]
			_, frsOk := frsMap[ffsMap[collection.ResourceID].ForestID]
			if ffOk && frsOk {
				res.Response.Data = append(res.Response.Data, dtocollections.ColItem{
					ID:           collection.ID,
					CreatedAt:    collection.CreatedAt.Unix(),
					ResourceID:   collection.ResourceID,
					ResourceType: collection.ResourceType,
					ResourceName: ffsMap[collection.ResourceID].Name,
					ForestID:     frsMap[ffsMap[collection.ResourceID].ForestID].ID,
					ForestName:   frsMap[ffsMap[collection.ResourceID].ForestID].Name,
					TagList:      fTagMap[collection.ResourceID],
					FileConfig:   ffsMap[collection.ResourceID].FileConfig,
				})
			}
		}
	}

	return res, nil
}

func MarkResourceCollection(ctx *gin.Context, req *dtocollections.MarkResourceCollectionRequest) (res *dtocollections.MarkResourceCollectionResponse, err error) {
	res = &dtocollections.MarkResourceCollectionResponse{}

	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	collection, err := forest.NewUinCollectionsDao().GetByCond(ctx, &forest.UinCollectionsCond{
		BaseCond: forest.BaseCond{
			Uin:       uin,
			CompanyID: companyID,
		},
		ResourceIDs:  []uint{req.Request.ResourceID},
		ResourceType: req.Request.ResourceType,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[MarkResourceCollection] NewUinCollectionsDao().GetByCond failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_mark_resource_collection_fail"
		return res, err
	}
	if req.Request.Enable {
		// ? make it collect
		if collection.ID == 0 {
			// ? correct branch
			if err := dbutil.Knownow().Create(&foresttype.UinCollections{
				Uin:          uin,
				CompanyID:    companyID,
				ResourceID:   req.Request.ResourceID,
				ResourceType: req.Request.ResourceType,
			}).Error; err != nil {
				logs.ErrorContextf(ctx, "[MarkResourceCollection] dbutil.Knownow().Create failed, err: %v", err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_mark_resource_collection_fail"
				return res, err
			}
			return res, nil
		}
		logs.WarnContextf(ctx, "[MarkResourceCollection] resource already collected")
		return res, nil
	}

	// ? make it uncollect
	if collection.ID != 0 {
		if err := dbutil.Knownow().Where("id = ?", collection.ID).Delete(&foresttype.UinCollections{}).Error; err != nil {
			logs.ErrorContextf(ctx, "[MarkResourceCollection] dbutil.Knownow().Delete failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_mark_resource_collection_fail"
			return res, err
		}
	}
	return res, nil
}
