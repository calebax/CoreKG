package svctag

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtotag"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"gorm.io/gorm"
)

func ListTag(ctx *gin.Context, req *dtotag.ListTagRequest) (res *dtotag.ListTagResponse, err error) {
	res = &dtotag.ListTagResponse{}
	companyID := runtime.CompanyID(ctx)
	tagEntityList, total, err := forest.NewTagDao().GetPageListByCond(ctx, &forest.TagCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
			OrderBy:   []string{"id desc"},
		},
		GroupID:  req.Request.TagGroupID,
		NameLike: req.Request.Name,
	})
	if err != nil {
		return nil, err
	}
	list := make([]dtotag.ListTagItem, 0, len(tagEntityList))
	var tagGroupIDs []uint
	for _, v := range tagEntityList {
		tagGroupIDs = append(tagGroupIDs, v.GroupID)
		list = append(list, dtotag.ListTagItem{
			TagGroupID: v.GroupID,
			TagID:      v.ID,
			Name:       v.Name,
			CreateAt:   v.CreatedAt.Unix(),
		})
	}
	groupEntityList, err := forest.NewTagGroupDao().GetListByCond(ctx, &forest.TagGroupCond{
		IDs: tagGroupIDs,
	})
	if err != nil {
		return nil, err
	}
	groupMap := groupEntityList.ToMap()
	for i := range list {
		group, ok := groupMap[list[i].TagGroupID]
		if ok {
			list[i].TagGroupName = group.Name
		}
	}
	res.Response.Limit = req.Request.Limit
	res.Response.Offset = req.Request.Offset
	res.Response.Total = total
	res.Response.List = list

	return res, nil
}

func CreateTag(ctx *gin.Context, req *dtotag.CreateTagRequest) (res *dtotag.CreateTagResponse, err error) {
	res = &dtotag.CreateTagResponse{}
	companyID := runtime.CompanyID(ctx)
	duplicatedNameCount, err := forest.NewTagDao().CountByCond(ctx, &forest.TagCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
		},
		GroupID: req.Request.TagGroupID,
		Name:    req.Request.Name,
	})
	if err != nil {
		return nil, err
	}
	if duplicatedNameCount > 0 {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_tag_name_duplicated"
		return res, nil
	}
	uin := runtime.Uin(ctx)
	insertEntity := &foresttype.Tag{
		CompanyID:  companyID,
		GroupID:    req.Request.TagGroupID,
		Name:       req.Request.Name,
		Status:     foresttype.TagGroupStatusEnable,
		CreatedUin: uin,
		UpdatedUin: uin,
	}
	if err := forest.NewTagDao().Insert(ctx, insertEntity); err != nil {
		return nil, err
	}
	res.Response.TagID = insertEntity.ID
	return res, nil
}

func ModifyTag(ctx *gin.Context, req *dtotag.ModifyTagRequest) (res *dtotag.ModifyTagResponse, err error) {
	res = &dtotag.ModifyTagResponse{}
	companyID := runtime.CompanyID(ctx)
	existTagEntity, err := forest.NewTagDao().GetByCond(ctx, &forest.TagCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
		},
		GroupID: req.Request.TagGroupID,
		Name:    req.Request.Name,
	})
	if err != nil {
		return nil, err
	}
	if existTagEntity != nil && existTagEntity.ID != 0 && existTagEntity.ID != req.Request.TagID {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_tag_name_duplicated"
		return res, nil
	}
	uin := runtime.Uin(ctx)
	updateMap := map[string]any{
		"name":        req.Request.Name,
		"updated_uin": uin,
		"group_id":    req.Request.TagGroupID,
	}
	if err := forest.NewTagDao().UpdateMap(ctx, req.Request.TagID, updateMap); err != nil {
		return nil, err
	}
	res.Response.TagID = req.Request.TagID
	return res, nil
}

func DeleteTag(ctx *gin.Context, req *dtotag.DeleteTagRequest) (res *dtotag.DeleteTagResponse, err error) {
	res = &dtotag.DeleteTagResponse{}
	tagEntity, err := forest.NewTagDao().GetByID(ctx, req.Request.TagID)
	if err != nil {
		return nil, err
	}
	if tagEntity == nil || tagEntity.ID == 0 {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_tag_not_exist"
		return res, nil
	}
	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := forest.NewTagDao().WithTx(tx).Delete(ctx, req.Request.TagID); err != nil {
			return err
		}
		if err := forest.NewResourceTagDao().WithTx(tx).DeleteByTagIDs(ctx, []uint{req.Request.TagID}); err != nil {
			return err
		}
		if err := forest.NewRecentUsedTagDao().WithTx(tx).DeleteByTagIDs(ctx, []uint{req.Request.TagID}); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return res, nil
}

func GetTagTree(ctx *gin.Context, req *dtotag.GetTagTreeRequest) (res *dtotag.GetTagTreeResponse, err error) {
	res = &dtotag.GetTagTreeResponse{}
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)
	recentTagEntityList, err := forest.NewRecentUsedTagDao().GetListByCond(ctx, &forest.RecentUsedTagCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
			Uin:       uin,
			Limit:     5,
			OrderBy:   []string{"last_used_at desc"},
		},
	})
	if err != nil {
		return nil, err
	}

	recentTagIDs := make([]uint, 0, len(recentTagEntityList))
	for _, v := range recentTagEntityList {
		recentTagIDs = append(recentTagIDs, v.TagID)
	}

	tagEntityList, err := forest.NewTagDao().GetListByCond(ctx, &forest.TagCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
		},
	})
	if err != nil {
		return nil, err
	}
	tagEntityMap := tagEntityList.ToMap()
	recentTagList := make([]dtotag.TagTreeListTagItem, 0, len(recentTagIDs))
	for _, tagID := range recentTagIDs {
		if tag, ok := tagEntityMap[tagID]; ok {
			recentTagList = append(recentTagList, dtotag.TagTreeListTagItem{
				TagID:   tag.ID,
				TagName: tag.Name,
			})
		}
	}

	groupTagMap := make(map[uint][]dtotag.TagTreeListTagItem)
	for _, v := range tagEntityList {
		if _, ok := groupTagMap[v.GroupID]; !ok {
			groupTagMap[v.GroupID] = make([]dtotag.TagTreeListTagItem, 0)
		}
		groupTagMap[v.GroupID] = append(groupTagMap[v.GroupID], dtotag.TagTreeListTagItem{
			TagID:   v.ID,
			TagName: v.Name,
		})
	}

	groupEntityList, err := forest.NewTagGroupDao().GetListByCond(ctx, &forest.TagGroupCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
		},
	})
	if err != nil {
		return nil, err
	}
	groupList := make([]dtotag.TagTreeListGroupItem, 0, len(groupEntityList))
	for _, v := range groupEntityList {
		item := dtotag.TagTreeListGroupItem{
			TagGroupID:   v.ID,
			TagGroupName: v.Name,
			TagList:      groupTagMap[v.ID],
		}
		if _, ok := groupTagMap[v.ID]; !ok {
			item.TagList = make([]dtotag.TagTreeListTagItem, 0)
		}
		groupList = append(groupList, item)
	}

	res.Response.RecentTagList = recentTagList
	res.Response.GroupList = groupList
	return res, nil
}
