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

func CreateTagGroup(ctx *gin.Context, req *dtotag.CreateTagGroupRequest) (res *dtotag.CreateTagGroupResponse, err error) {
	res = &dtotag.CreateTagGroupResponse{}
	companyID := runtime.CompanyID(ctx)
	duplicatedNameCount, err := forest.NewTagGroupDao().CountByCond(ctx, &forest.TagGroupCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
		},
		Name: req.Request.Name,
	})
	if err != nil {
		return nil, err
	}
	if duplicatedNameCount > 0 {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_tag_group_name_duplicated"
		return res, nil
	}

	uin := runtime.Uin(ctx)
	insertEntity := &foresttype.TagGroup{
		CompanyID:  companyID,
		Name:       req.Request.Name,
		Status:     foresttype.TagGroupStatusEnable,
		CreatedUin: uin,
		UpdatedUin: uin,
	}
	if err := forest.NewTagGroupDao().Insert(ctx, insertEntity); err != nil {
		return nil, err
	}
	res.Response.TagGroupID = insertEntity.ID
	return res, nil
}

func ModifyTagGroup(ctx *gin.Context, req *dtotag.ModifyTagGroupRequest) (res *dtotag.ModifyTagGroupResponse, err error) {
	res = &dtotag.ModifyTagGroupResponse{}
	companyID := runtime.CompanyID(ctx)
	existGroupEntity, err := forest.NewTagGroupDao().GetByCond(ctx, &forest.TagGroupCond{
		BaseCond: forest.BaseCond{
			CompanyID: companyID,
		},
		Name: req.Request.Name,
	})
	if err != nil {
		return nil, err
	}
	if existGroupEntity != nil && existGroupEntity.ID != 0 && existGroupEntity.ID != req.Request.TagGroupID {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_tag_group_name_duplicated"
		return res, nil
	}
	uin := runtime.Uin(ctx)
	updateMap := map[string]any{
		"name":        req.Request.Name,
		"updated_uin": uin,
	}
	if err := forest.NewTagGroupDao().UpdateMap(ctx, req.Request.TagGroupID, updateMap); err != nil {
		return nil, err
	}

	res.Response.TagGroupID = req.Request.TagGroupID
	return res, nil
}

func DeleteTagGroup(ctx *gin.Context, req *dtotag.DeleteTagGroupRequest) (res *dtotag.DeleteTagGroupResponse, err error) {
	res = &dtotag.DeleteTagGroupResponse{}
	groupEntity, err := forest.NewTagGroupDao().GetByID(ctx, req.Request.TagGroupID)
	if err != nil {
		return nil, err
	}
	if groupEntity == nil || groupEntity.ID == 0 {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_tag_group_not_exist"
		return res, nil
	}

	tagEntityList, err := forest.NewTagDao().GetListByCond(ctx, &forest.TagCond{
		BaseCond: forest.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
		},
		GroupID: req.Request.TagGroupID,
	})
	if err != nil {
		return nil, err
	}
	var tagIDs []uint
	for _, v := range tagEntityList {
		tagIDs = append(tagIDs, v.ID)
	}
	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := forest.NewTagGroupDao().WithTx(tx).Delete(ctx, req.Request.TagGroupID); err != nil {
			return err
		}
		// 删除标签分组相关数据
		if len(tagIDs) > 0 {
			if err := forest.NewTagDao().WithTx(tx).DeleteByGroupID(ctx, req.Request.TagGroupID); err != nil {
				return err
			}
			if err := forest.NewResourceTagDao().WithTx(tx).DeleteByTagIDs(ctx, tagIDs); err != nil {
				return err
			}
			if err := forest.NewRecentUsedTagDao().WithTx(tx).DeleteByTagIDs(ctx, tagIDs); err != nil {
				return err
			}
		}

		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	res.Response.TagGroupID = req.Request.TagGroupID
	return res, nil
}

func ListTagGroup(ctx *gin.Context, req *dtotag.ListTagGroupRequest) (res *dtotag.ListTagGroupResponse, err error) {
	res = &dtotag.ListTagGroupResponse{}
	groupEntityList, total, err := forest.NewTagGroupDao().GetPageListByCond(ctx, &forest.TagGroupCond{
		BaseCond: forest.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
			Offset:    req.Request.Offset,
			Limit:     req.Request.Limit,
		},
		NameLike: req.Request.Name,
	})
	if err != nil {
		return nil, err
	}
	groupList := make([]dtotag.ListTagGroupItem, 0, len(groupEntityList))
	for _, v := range groupEntityList {
		groupList = append(groupList, dtotag.ListTagGroupItem{
			TagGroupID: v.ID,
			Name:       v.Name,
			CreateAt:   v.CreatedAt.Unix(),
		})
	}
	res.Response.QueryResponse.Total = total
	res.Response.QueryResponse.Limit = req.Request.Limit
	res.Response.QueryResponse.Offset = req.Request.Offset
	res.Response.List = groupList
	return res, nil
}
