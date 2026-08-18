package svclikes

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtolikes"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func ListLikes(ctx *gin.Context, req *dtolikes.ListLikesRequest) (res *dtolikes.ListLikesResponse, err error) {
	res = &dtolikes.ListLikesResponse{}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	resourceType := foresttype.ResourceTypeForestFile

	cond := &forest.UinLikesCond{
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
				logs.ErrorContextf(ctx, "[ListLikes] invalid resource id: %s", v)
				return 0
			})
			cond.ResourceIDs = ids
		case "resource_type":
			cond.ResourceType = foresttype.ResourceType(v.Value[0])
			resourceType = cond.ResourceType
		default:
			logs.WarnContextf(ctx, "[ListLikes] invalid filter field: %s", v.Field)
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
		cond.NoResourceIDs = ap.BanList
	}

	viewableForests, err := forest.ViewAbleForests(uin, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListLikes] ViewAbleForests failed, err: %v", err)
		return nil, err
	}
	if len(viewableForests) == 0 {
		logs.InfoContextf(ctx, "[ListLikes] no viewable forests")
		return res, nil
	}

	frsIDMap := make(map[uint]struct{})
	for _, frsID := range viewableForests {
		frsIDMap[frsID] = struct{}{}
	}

	list, c, err := forest.NewUinLikesDao().GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListLikes] NewUinLikesDao().GetPageListByCond failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_list_likes_fail"
		return res, err
	}

	res.Response.QueryResponse.Total = c
	res.Response.QueryResponse.Limit = req.Request.Limit
	res.Response.QueryResponse.Offset = req.Request.Offset

	if c == 0 {
		logs.InfoContextf(ctx, "[ListLikes] NewUinLikesDao().GetPageListByCond no data")
		return res, nil
	}

	resourceIDs := utils.Map(list, func(like foresttype.UinLikes) uint {
		return like.ResourceID
	})

	res.Response.Data = make([]dtolikes.LiItem, 0, len(list))

	switch resourceType {
	case foresttype.ResourceTypeForestFile:
		ffs, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
			BaseCond: forest.BaseCond{
				CompanyID: companyID,
			},
			IDs: resourceIDs,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[ListLikes] NewForestFileDao().GetListByCond failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_list_likes_fail"
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
			logs.ErrorContextf(ctx, "[ListLikes] NewForestDao().GetListByCond failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_list_likes_fail"
			return res, err
		}
		frsMap := frs.ToMap()

		var tags []dtolikes.ResourceTag
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

		fTagMap := make(map[uint][]dtolikes.ResourceTag)
		for _, tag := range tags {
			fTagMap[tag.ResourceID] = append(fTagMap[tag.ResourceID], tag)
		}

		for _, like := range list {
			_, ffOk := ffsMap[like.ResourceID]
			_, frsOk := frsMap[ffsMap[like.ResourceID].ForestID]
			if ffOk && frsOk {
				res.Response.Data = append(res.Response.Data, dtolikes.LiItem{
					ID:           like.ID,
					CreatedAt:    like.CreatedAt.Unix(),
					ResourceID:   like.ResourceID,
					ResourceType: like.ResourceType,
					ResourceName: ffsMap[like.ResourceID].Name,
					ForestID:     frsMap[ffsMap[like.ResourceID].ForestID].ID,
					ForestName:   frsMap[ffsMap[like.ResourceID].ForestID].Name,
					TagList:      fTagMap[like.ResourceID],
					FileConfig:   ffsMap[like.ResourceID].FileConfig,
				})
			}
		}
	}

	return res, nil
}

func MarkResourceLike(ctx *gin.Context, req *dtolikes.MarkResourceLikeRequest) (res *dtolikes.MarkResourceLikeResponse, err error) {
	res = &dtolikes.MarkResourceLikeResponse{}

	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	like, err := forest.NewUinLikesDao().GetByCond(ctx, &forest.UinLikesCond{
		BaseCond: forest.BaseCond{
			Uin:       uin,
			CompanyID: companyID,
		},
		ResourceIDs:  []uint{req.Request.ResourceID},
		ResourceType: req.Request.ResourceType,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[MarkResourceLike] NewUinLikesDao().GetByCond failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_mark_resource_like_fail"
		return res, err
	}
	if req.Request.Enable {
		// ? make it like
		if like.ID == 0 {
			// ? correct branch
			if err := dbutil.Knownow().Create(&foresttype.UinLikes{
				Uin:          uin,
				CompanyID:    companyID,
				ResourceID:   req.Request.ResourceID,
				ResourceType: req.Request.ResourceType,
			}).Error; err != nil {
				logs.ErrorContextf(ctx, "[MarkResourceLike] dbutil.Knownow().Create failed, err: %v", err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "kecore_mark_resource_like_fail"
				return res, err
			}
			return res, nil
		}
		logs.WarnContextf(ctx, "[MarkResourceLike] resource already liked")
		return res, nil
	}

	// ? make it unlike
	if like.ID != 0 {
		if err := dbutil.Knownow().Where("id = ?", like.ID).Delete(&foresttype.UinLikes{}).Error; err != nil {
			logs.ErrorContextf(ctx, "[MarkResourceLike] dbutil.Knownow().Delete failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_mark_resource_like_fail"
			return res, err
		}
	}
	return res, nil
}
