package agentperm

import (
	"context"
	"fmt"
	"slices"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ActOption string

var (
	ActOptionUpdate ActOption = "update"
	ActOptionFetch  ActOption = "fetch"
)

type PermSet struct {
	Agent      *chattype.ChatAgent `json:"agent"`
	ManagePerm bool                `json:"manage_perm"`
	UsePerm    bool                `json:"use_perm"`
	ActOption  ActOption           `json:"act_option"`
}

// GetAgentByCompanyID get all agents that a company can meet
func GetAgentByCompanyID(ctx context.Context, companyID uint) (res []*chattype.ChatAgent, err error) {
	err = dbutil.Chat().WithContext(ctx).
		Table(chattype.TableNameAgent+" AS f").
		Where("f.deleted_at IS NULL").
		Where("f.public_scope != ?", chattype.PublicScopePrivate).
		Where("f.publish_status != ?", chattype.StatusDraft).
		Where("f.company_id = ?", companyID).
		Find(&res).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAgentByCompanyID] failed ,err %v", err)
	}
	return
}

func GetScopesByAgentIDs(agentsIDs []uint, uin, companyID uint) (res []*foresttype.KeResourceScope, err error) {
	err = dbutil.Knownow().
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeAgent).
		Where("resource_id IN ?", agentsIDs).
		Where("(scope_type = ? AND scope_id = ?) OR (scope_type = ? AND scope_id = ?)",
			chattype.ScopeTypeUser, uin,
			chattype.PublicScopeCompany, companyID,
		).
		Find(&res).Error
	return
}

type WrapperPermSet struct {
	Ctx context.Context

	Uin       uint `json:"uin"`
	CompanyID uint `json:"company_id"`

	TargPs []*PermSet `json:"targ_perm_set"`

	Agents   []*chattype.ChatAgent `json:"agents"`
	AgentIDs []uint                `json:"agent_ids"`

	CurrManageFm map[uint]struct{} `json:"curr_manage_fm"`
	CurrUseFm    map[uint]struct{} `json:"curr_use_fm"`
}

func NewWrapperPermSet(
	ctx context.Context, uin, companyID uint, tps []*PermSet, ags []*chattype.ChatAgent) *WrapperPermSet {
	wps := &WrapperPermSet{
		Ctx: ctx,

		Uin:       uin,
		CompanyID: companyID,
		Agents:    ags,
		AgentIDs:  make([]uint, 0, len(tps)),

		TargPs: tps,

		CurrManageFm: make(map[uint]struct{}, len(tps)),
		CurrUseFm:    make(map[uint]struct{}, len(tps)),
	}

	if len(tps) > 0 {
		for i := range tps {
			if tps[i].ActOption == ActOptionUpdate {
				wps.AgentIDs = append(wps.AgentIDs, tps[i].Agent.ID)
			}
		}
	} else {
		for i := range ags {
			wps.AgentIDs = append(wps.AgentIDs, ags[i].ID)
		}
	}

	if err := dbutil.Chat().
		Where("company_id = ? ", companyID).
		Where("id in ?", wps.AgentIDs).
		Find(&wps.Agents).Error; err != nil {
		logs.ErrorContextf(ctx, "NewWrapperPermSet err: %v", err)
		return nil
	}

	wps.AgentIDs = make([]uint, 0)
	for _, v := range wps.Agents {
		wps.AgentIDs = append(wps.AgentIDs, v.ID)
	}

	tmp := make([]*PermSet, 0)
	for _, v := range tps {
		if slices.Contains(wps.AgentIDs, v.Agent.ID) {
			tmp = append(tmp, v)
		}
	}
	wps.TargPs = tmp

	return wps
}

func (w *WrapperPermSet) BuildCurrentPermMap() error {
	//*construct perm map
	scopes, err := GetScopesByAgentIDs(w.AgentIDs, w.Uin, w.CompanyID)
	if err != nil {
		return err
	}

	userScopeAgentIDs := make(map[uint]struct{})
	companyScopeAgentIDs := make(map[uint]struct{})

	managerScopeAgentIDs := make(map[uint]struct{})
	for _, scope := range scopes {
		if scope.ScopeType == foresttype.ScopeTypeUser {
			switch scope.Action {
			case foresttype.ActionView:
				userScopeAgentIDs[scope.ResourceID] = struct{}{}
			case foresttype.ActionManage:
				managerScopeAgentIDs[scope.ResourceID] = struct{}{}
			}
		} else if scope.ScopeType == foresttype.ScopeTypeCompany {
			companyScopeAgentIDs[scope.ResourceID] = struct{}{}
		}
	}

	for _, f := range w.Agents {
		//*build manage agentID -> perm map
		if _, ok := managerScopeAgentIDs[f.ID]; ok {
			w.CurrManageFm[f.ID] = struct{}{}
		}

		if f.PublishStatus == chattype.StatusDraft &&
			f.Uin == w.Uin {
			//For user, self's draft should have both manage and use perm
			w.CurrUseFm[f.ID] = struct{}{}
			continue
		}

		//*build use agentID -> perm map
		if f.PublicScope == chattype.PublicScopeCompany {
			w.CurrUseFm[f.ID] = struct{}{}
		} else if f.PublicScope == chattype.PublicScopeCustom {
			if _, ok := userScopeAgentIDs[f.ID]; ok {
				w.CurrUseFm[f.ID] = struct{}{}
			}
		} else if f.PublicScope == chattype.PublicScopePublic {
			if _, ok := companyScopeAgentIDs[f.ID]; ok {
				w.CurrUseFm[f.ID] = struct{}{}
			}
		}
	}
	return nil
}

func (w *WrapperPermSet) GetCurrPermSet() []*PermSet {
	var result []*PermSet
	for _, f := range w.Agents {
		result = append(result, &PermSet{
			Agent:      f,
			ManagePerm: w.hasPermission(w.CurrManageFm, f.ID),
			UsePerm:    w.hasPermission(w.CurrUseFm, f.ID),
		})
	}
	return result
}

func (w *WrapperPermSet) hasPermission(permMap map[uint]struct{}, agentID uint) bool {
	_, ok := permMap[agentID]
	return ok
}

// ApplyChanges get diff and apply changes
func (w *WrapperPermSet) ApplyChanges() error {
	// --- 阶段1: 收集所有需要执行的变更 ---
	// 待处理"使用权限"的变更
	scopesToCreate := make([]*foresttype.KeResourceScope, 0)
	ToDeleteScope := make([]uint, 0)
	// 待处理"管理权限"的变更
	ToAddManager := make([]*foresttype.KeResourceScope, 0)
	ToRemoveManager := make([]uint, 0)
	for _, targetPerm := range w.TargPs {
		forestID := targetPerm.Agent.ID
		// 1. 收集【管理权限】的变更
		hasManagePerm := w.hasPermission(w.CurrManageFm, forestID)
		if targetPerm.ManagePerm && !hasManagePerm {
			ToAddManager = append(ToAddManager, &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeAgent,
				ResourceID:   forestID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      w.Uin,
				Action:       foresttype.ActionManage,
			})
		} else if !targetPerm.ManagePerm && hasManagePerm {
			ToRemoveManager = append(ToRemoveManager, forestID)
		}
		// 2. 收集【使用权限】的变更
		hasUsePerm := w.hasPermission(w.CurrUseFm, forestID)
		if targetPerm.UsePerm && !hasUsePerm {
			scopesToCreate = append(scopesToCreate, &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeAgent,
				ResourceID:   forestID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      w.Uin,
				Action:       foresttype.ActionView,
			})
		} else if !targetPerm.UsePerm && hasUsePerm {
			ToDeleteScope = append(ToDeleteScope, forestID)
		}
	}
	// --- 阶段2: 批量执行数据库操作 ---
	// 批量授予【管理权限】
	return dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if len(ToAddManager) > 0 {
			if err := tx.CreateInBatches(&ToAddManager, len(ToAddManager)).Error; err != nil {
				return fmt.Errorf("批量授予管理权限失败: %w", err)
			}
		}
		// 批量撤销【管理权限】
		if len(ToRemoveManager) > 0 {
			if err := tx.Where("resource_id IN ?", ToRemoveManager).
				Where("resource_type = ?", foresttype.ResourceTypeAgent).
				Where("scope_type = ? AND scope_id = ?", chattype.ScopeTypeUser, w.Uin).
				Where("action = ?", foresttype.ActionManage).
				Delete(&foresttype.KeResourceScope{}).Error; err != nil {
				return fmt.Errorf("批量撤销使用权限失败: %w", err)
			}
		}
		// 批量创建"使用权限"
		if len(scopesToCreate) > 0 {
			if err := tx.CreateInBatches(&scopesToCreate, len(scopesToCreate)).Error; err != nil {
				return fmt.Errorf("批量授予使用权限失败: %w", err)
			}
		}
		// 批量删除"使用权限"
		if len(ToDeleteScope) > 0 {
			if err := tx.Where("resource_id IN ?", ToDeleteScope).
				Where("resource_type = ?", foresttype.ResourceTypeAgent).
				Where("scope_type = ? AND scope_id = ?", chattype.ScopeTypeUser, w.Uin).
				Where("action = ?", foresttype.ActionView).
				Delete(&foresttype.KeResourceScope{}).Error; err != nil {
				return fmt.Errorf("批量撤销使用权限失败: %w", err)
			}
		}
		return nil
	})
}
