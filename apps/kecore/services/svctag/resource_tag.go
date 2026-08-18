package svctag

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtotag"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SetResourceTag(ctx *gin.Context, req *dtotag.SetResourceTagRequest) (res *dtotag.SetResourceTagResponse, err error) {
	res = &dtotag.SetResourceTagResponse{}
	// 如果没有传标签ID，则删除所有标签
	if len(req.Request.TagIDs) == 0 {
		if err := forest.NewResourceTagDao().DeleteByResource(ctx, req.Request.ResourceType, []uint{req.Request.ResourceID}); err != nil {
			return nil, err
		}
	}
	tagEntityList, err := forest.NewTagDao().GetListByCond(ctx, &forest.TagCond{
		IDs: req.Request.TagIDs,
	})
	if err != nil {
		return nil, err
	}
	tagMap := tagEntityList.ToMap()
	insertEntityList := make([]foresttype.ResourceTag, 0, len(req.Request.TagIDs))
	for _, tagID := range req.Request.TagIDs {
		insertEntityList = append(insertEntityList, foresttype.ResourceTag{
			ResourceType: req.Request.ResourceType,
			ResourceID:   req.Request.ResourceID,
			TagID:        tagID,
			GroupID:      tagMap[tagID].GroupID,
		})
	}

	tIDs := utils.Keys(tagMap)
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除所有标签
		if err := forest.NewResourceTagDao().WithTx(tx).DeleteByResource(ctx, req.Request.ResourceType, []uint{req.Request.ResourceID}); err != nil {
			return err
		}
		if len(insertEntityList) > 0 {
			// 再插入新标签
			if err := forest.NewResourceTagDao().WithTx(tx).BatchInsert(ctx, insertEntityList); err != nil {
				return err
			}
		}

		return BatchSaveRecentUsedTag(ctx, tx, tIDs, uin, companyID)
	})
	if txErr != nil {
		return nil, txErr
	}
	return res, nil
}

// BatchSaveRecentUsedTag 批量保存 RecentUsedTag，更新使用次数与上次使用时间
// tagIDs: 标签ID列表
// uin: 用户ID
// companyID: 公司ID
func BatchSaveRecentUsedTag(ctx context.Context, tx *gorm.DB, tagIDs []uint, uin uint, companyID uint) error {
	if len(tagIDs) == 0 {
		return nil
	}

	tagEntityList, err := forest.NewTagDao().GetListByCond(ctx, &forest.TagCond{
		IDs: tagIDs,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "failed to fetch tag info: %w", err)
		return err
	}
	tagMap := tagEntityList.ToMap()

	ts, err := forest.NewRecentUsedTagDao().GetListByCond(ctx, &forest.RecentUsedTagCond{
		TagIDs: tagIDs,
		BaseCond: forest.BaseCond{
			Uin:       uin,
			CompanyID: companyID,
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "failed to fetch recent used tag info: %w", err)
		return err
	}
	tUsedMap := make(map[uint]*foresttype.RecentUsedTag)
	for i := range ts {
		tUsedMap[ts[i].TagID] = &ts[i]
	}

	var toSaveUsedTag []*foresttype.RecentUsedTag
	for _, tagID := range tagIDs {
		if tag, ok := tagMap[tagID]; ok {
			if t, exists := tUsedMap[tagID]; !exists {
				// 不存在，创建新记录
				toSaveUsedTag = append(toSaveUsedTag, &foresttype.RecentUsedTag{
					CompanyID:  companyID,
					GroupID:    tag.GroupID,
					Uin:        uin,
					TagID:      tagID,
					LastUsedAt: time.Now(),
					UsageCount: 1,
				})
			} else {
				// 存在，更新使用次数和最后使用时间
				t.UsageCount++
				t.LastUsedAt = time.Now()
				toSaveUsedTag = append(toSaveUsedTag, t)
			}
		}
	}

	// 批量保存
	if len(toSaveUsedTag) > 0 {
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uin"}, {Name: "tag_id"}, {Name: "group_id"}, {Name: "company_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"usage_count", "last_used_at"}),
		}).Create(&toSaveUsedTag).Error; err != nil {
			logs.ErrorContextf(ctx, "failed to save recent used tag: %w", err)
			return err
		}
	}

	return nil
}
