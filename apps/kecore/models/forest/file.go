package forest

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type FileAccessProvider struct {
	Model *ContextModel
}

func (f *FileAccessProvider) Apply(ctx context.Context) error {
	var entityList []foresttype.KeResourceScope

	for _, v := range f.Model.Opt.BanList {
		entityList = append(entityList, foresttype.KeResourceScope{
			ResourceID:   f.Model.ResourceID,
			ResourceType: f.Model.ResourceType,
			ScopeType:    foresttype.ScopeTypeUser,
			Action:       foresttype.ActionBan,
			ScopeID:      v,
		})
	}

	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := NewKeResourceScopeDao().WithTx(tx).DeleteByCond(ctx, &KeResourceScopeCond{
			ResourceID:   f.Model.ResourceID,
			ResourceType: f.Model.ResourceType,
			Action:       foresttype.ActionBan,
		}); err != nil {
			logs.ErrorContextf(ctx, "FileAccessProvider: Delete ban list failed: %w", err)
			return fmt.Errorf("delete ban list failed: %w", err)
		}

		if len(entityList) > 0 {
			if err := tx.CreateInBatches(entityList, 50).Error; err != nil {
				logs.ErrorContextf(ctx, "FileAccessProvider: Batch insert ban list failed: %w", err)
				return fmt.Errorf("batch insert ban list failed: %w", err)
			}
		}

		return nil

	}); err != nil {
		logs.ErrorContextf(ctx, "FileAccessProvider: Apply failed: %w", err)
		return err
	}
	return nil
}

func (f *FileAccessProvider) Get(ctx context.Context) (*AccessResult, error) {
	rsc, err := NewKeResourceScopeDao().GetListByCond(ctx, &KeResourceScopeCond{
		ResourceID:   f.Model.ResourceID,
		ResourceType: f.Model.ResourceType,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "FileAccessProvider: Get failed: %w", err)
		return nil, err
	}

	result := &AccessResult{
		ManagerList: make([]uint, 0),
		ViewerList:  make([]uint, 0),
		BanList:     make([]uint, 0),
	}

	for _, v := range rsc {
		if v.Action == foresttype.ActionBan {
			result.BanList = append(result.BanList, v.ScopeID)
		}
		if v.Action == foresttype.ActionManage {
			result.ManagerList = append(result.ManagerList, v.ScopeID)
		}
		if v.Action == foresttype.ActionView {
			result.ViewerList = append(result.ViewerList, v.ScopeID)
		}
	}
	return result, nil
}

func (f *FileAccessProvider) Action(ctx context.Context) (*AccessResult, error) {
	result := &AccessResult{
		ManagerList: make([]uint, 0),
		ViewerList:  make([]uint, 0),
		BanList:     make([]uint, 0),
	}
	cond := &KeResourceScopeCond{}

	if f.Model.ResourceID != 0 {
		cond.ResourceID = f.Model.ResourceID
	}
	if f.Model.ResourceType != "" {
		cond.ResourceType = f.Model.ResourceType
	}

	if f.Model.ScopeType != "" {
		cond.ScopeTypeList = []foresttype.ScopeType{f.Model.ScopeType}
	}
	if f.Model.ScopeID != 0 {
		cond.ScopeID = f.Model.ScopeID
	}
	if f.Model.Action != "" {
		cond.Action = f.Model.Action
	}
	rsc, err := NewKeResourceScopeDao().GetListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "FileAccessProvider: Action failed: %w", err)
		return nil, err
	}

	for _, v := range rsc {
		switch v.Action {
		case foresttype.ActionBan:
			result.BanList = append(result.BanList, v.ResourceID)
		case foresttype.ActionManage:
			result.ManagerList = append(result.ManagerList, v.ResourceID)
		case foresttype.ActionView:
			result.ViewerList = append(result.ViewerList, v.ResourceID)
		}
	}

	return result, nil
}
