package forest

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type QueryForestPublicScopeListResponse struct {
	apiobj.QueryResponse
	Data []*ResourceScopeWithName
}

type ResourceScopeWithName struct {
	foresttype.KeResourceScope
	UserName string `json:"user_name"`
}

// QueryResourceScopeList 查询 forest 的作用域列表，并附加用户名称
func QueryResourceScopeList(ctx context.Context, opt apiobj.PageQuery, forestID uint, result *QueryForestPublicScopeListResponse) error {
	// 第一步：查询 KeResourceScope 列表
	var rawData []*foresttype.KeResourceScope

	query := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeForest).
		Where("resource_id = ?", forestID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("action = ?", foresttype.ActionView)

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "scope_type":
			query = query.Where(foresttype.TableNameKeResourceScope+".`scope_type` = ?", filter.Value[0])
		default:
			logs.WarnContextf(ctx, "[knownow-forest][QueryResourceScopeList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&result.Total).Error; err != nil {
		return err
	}
	if result.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	if err := query.Find(&rawData).Error; err != nil {
		return err
	}

	// 第二步：提取 scope_id，去 account 数据库查用户名
	var scopeIDs []uint
	for _, item := range rawData {
		scopeIDs = append(scopeIDs, item.ScopeID)
	}

	var idNamePairs []struct {
		ID   uint
		Name string
	}
	if len(scopeIDs) > 0 {
		err := dbutil.Account().WithContext(ctx).
			Table("user_identification AS ui").
			Select("ui.id AS id, ui.name").
			Where("ui.subject_type = ?", "company").
			Where("ui.uin_status = ?", "normal").
			Where("ui.id IN ?", scopeIDs).
			Scan(&idNamePairs).Error
		if err != nil {
			return err
		}
	}

	nameMap := make(map[uint]string, len(idNamePairs))
	for _, pair := range idNamePairs {
		nameMap[pair.ID] = pair.Name
	}

	// 带用户名的结果
	for _, item := range rawData {
		result.Data = append(result.Data, &ResourceScopeWithName{
			KeResourceScope: *item,
			UserName:        nameMap[item.ScopeID],
		})
	}

	return nil
}

func CoverForestPublicScope(forestID uint, scopeType foresttype.ScopeType, newScopeIDs []uint) error {
	return dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		var existingScopeIDs []uint
		err := tx.
			Model(&foresttype.KnownowForestPublicScope{}).
			Where("forest_id = ? AND scope_type = ?", forestID, scopeType).
			Pluck("scope_id", &existingScopeIDs).Error
		if err != nil {
			return err
		}

		// 构造集合对比
		newScopeMap := make(map[uint]struct{}, len(newScopeIDs))
		for _, id := range newScopeIDs {
			newScopeMap[id] = struct{}{}
		}

		existingMap := make(map[uint]struct{}, len(existingScopeIDs))
		for _, id := range existingScopeIDs {
			existingMap[id] = struct{}{}
		}

		// 找出需要添加的
		var toAdd []uint
		for id := range newScopeMap {
			if _, exists := existingMap[id]; !exists {
				toAdd = append(toAdd, id)
			}
		}

		// 找出需要删除的
		var toDelete []uint
		for id := range existingMap {
			if _, exists := newScopeMap[id]; !exists {
				toDelete = append(toDelete, id)
			}
		}

		// 执行插入
		if len(toAdd) > 0 {
			var insertList []foresttype.KnownowForestPublicScope
			for _, id := range toAdd {
				insertList = append(insertList, foresttype.KnownowForestPublicScope{
					ForestID:  forestID,
					ScopeType: scopeType,
					ScopeID:   id,
				})
			}
			if err := tx.Create(&insertList).Error; err != nil {
				return err
			}
		}

		// 执行删除
		if len(toDelete) > 0 {
			if err := tx.
				Where("forest_id = ? AND scope_type = ? AND scope_id IN ?", forestID, scopeType, toDelete).
				Delete(&foresttype.KnownowForestPublicScope{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func UpdateForestManager(forestID uint, publicScope foresttype.PublicScope, managers types.UintArray) error {
	return dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		// 更新 public_scope 和 manager_ids
		if err := tx.Model(&foresttype.KnownowForest{}).
			Where("id = ?", forestID).
			Updates(map[string]interface{}{
				"public_scope": publicScope,
				"manager_ids":  managers,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
