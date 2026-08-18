package perm

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

type Set struct {
	Forest     *foresttype.KnownowForest `json:"forest"`
	ManagePerm bool                      `json:"manage_perm"`
	UserPerm   bool                      `json:"use_perm"`
	ActOption  ActOption                 `json:"act_option"`
}

// GetForestsByCompanyID get all forests that a company can meet
func GetForestsByCompanyID(companyID uint) (res []*foresttype.KnownowForest, err error) {
	err = dbutil.Knownow().
		Table(foresttype.TableNameKnownowForest+" AS f").
		Where("f.deleted_at IS NULL").
		Where("f.public_scope != ?", foresttype.PublicScopePrivate).
		Where("f.company_id = ?", companyID).
		Find(&res).Error
	return
}

func GetScopesByForestIDs(frsIDs []uint, uin, companyID uint) (res []*foresttype.KeResourceScope, err error) {
	err = dbutil.Knownow().
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeForest).
		Where("resource_id IN ?", frsIDs).
		Where("(scope_type = ? AND scope_id = ?) OR"+
			"(scope_type = ? AND scope_id = ?)", foresttype.ScopeTypeUser, uin, foresttype.ScopeTypeCompany, companyID).
		Find(&res).Error
	return
}

type WrapperPermSet struct {
	Ctx context.Context

	Uin       uint `json:"uin"`
	CompanyID uint `json:"company_id"`

	TargPs []*Set `json:"targ_perm_set"`

	Forests   []*foresttype.KnownowForest `json:"forests"`
	ForestIDs []uint                      `json:"forest_ids"`

	CurrManageFm map[uint]struct{} `json:"curr_manage_fm"`
	CurrUseFm    map[uint]struct{} `json:"curr_use_fm"`
}

func NewWrapperPermSet(ctx context.Context, uin, companyID uint, tps []*Set, frss []*foresttype.KnownowForest) *WrapperPermSet {
	wps := &WrapperPermSet{
		Ctx: ctx,

		Uin:       uin,
		CompanyID: companyID,
		Forests:   frss,

		TargPs: tps,

		ForestIDs: make([]uint, 0, len(tps)),

		CurrManageFm: make(map[uint]struct{}, len(tps)),
		CurrUseFm:    make(map[uint]struct{}, len(tps)),
	}

	if len(tps) > 0 {
		for i := range tps {
			if tps[i].ActOption == ActOptionUpdate {
				wps.ForestIDs = append(wps.ForestIDs, tps[i].Forest.ID)
			}
		}
	} else {
		for i := range frss {
			wps.ForestIDs = append(wps.ForestIDs, frss[i].ID)
		}
	}

	if err := dbutil.Knownow().
		Where("company_id = ? ", companyID).
		Where("id in ?", wps.ForestIDs).Find(&wps.Forests).Error; err != nil {
		logs.ErrorContextf(ctx, "NewWrapperPermSet err: %v", err)
		return nil
	}
	wps.ForestIDs = make([]uint, 0)
	for _, v := range wps.Forests {
		wps.ForestIDs = append(wps.ForestIDs, v.ID)
	}

	tmp := make([]*Set, 0)
	for _, v := range tps {
		if slices.Contains(wps.ForestIDs, v.Forest.ID) {
			tmp = append(tmp, v)
		}
	}
	wps.TargPs = tmp
	return wps
}

func (w *WrapperPermSet) BuildCurrentPermMap() error {
	//*construct perm map
	scopes, err := GetScopesByForestIDs(w.ForestIDs, w.Uin, w.CompanyID)
	if err != nil {
		return err
	}

	userScopeForestIDs := make(map[uint]struct{})
	companyScopeForestIDs := make(map[uint]struct{})

	manageScopeForestIDs := make(map[uint]struct{})

	for _, scope := range scopes {
		if scope.ScopeType == foresttype.ScopeTypeUser {
			switch scope.Action {
			case foresttype.ActionView:
				userScopeForestIDs[scope.ResourceID] = struct{}{}
			case foresttype.ActionManage:
				manageScopeForestIDs[scope.ResourceID] = struct{}{}
			}
		} else if scope.ScopeType == foresttype.ScopeTypeCompany {
			companyScopeForestIDs[scope.ResourceID] = struct{}{}
		}
	}

	for _, f := range w.Forests {
		//*build manage forestID -> perm map
		if _, ok := manageScopeForestIDs[f.ID]; ok {
			w.CurrManageFm[f.ID] = struct{}{}
		}
		//*build use forestID -> perm map
		if f.PublicScope == foresttype.PublicScopeCompany {
			w.CurrUseFm[f.ID] = struct{}{}
		} else if f.PublicScope == foresttype.PublicScopeCustom {
			if _, ok := userScopeForestIDs[f.ID]; ok {
				w.CurrUseFm[f.ID] = struct{}{}
			}
		} else if f.PublicScope == foresttype.PublicScopePublic {
			if _, ok := companyScopeForestIDs[f.ID]; ok {
				w.CurrUseFm[f.ID] = struct{}{}
			}
		}
	}
	return nil
}

func (w *WrapperPermSet) GetCurrPermSet() []*Set {
	var result []*Set
	for _, f := range w.Forests {
		result = append(result, &Set{
			Forest:     f,
			ManagePerm: w.hasPermission(w.CurrManageFm, f.ID),
			UserPerm:   w.hasPermission(w.CurrUseFm, f.ID),
		})
	}
	return result
}

func (w *WrapperPermSet) hasPermission(permMap map[uint]struct{}, forestID uint) bool {
	_, ok := permMap[forestID]
	return ok
}

// ApplyChanges get diff and apply changes
func (w *WrapperPermSet) ApplyChanges() error {
	// --- 阶段1: 收集所有需要执行的变更 ---
	// 待处理"使用权限"的变更
	scopesToCreate := make([]*foresttype.KeResourceScope, 0)
	forestIDsToDeleteScope := make([]uint, 0)
	// 待处理"管理权限"的变更
	forestIDsToAddManager := make([]*foresttype.KeResourceScope, 0)
	forestIDsToRemoveManager := make([]uint, 0)
	for _, targetPerm := range w.TargPs {
		forestID := targetPerm.Forest.ID
		// 1. 收集【管理权限】的变更
		hasManagePerm := w.hasPermission(w.CurrManageFm, forestID)
		if targetPerm.ManagePerm && !hasManagePerm {
			forestIDsToAddManager = append(forestIDsToAddManager, &foresttype.KeResourceScope{
				ResourceID:   forestID,
				ResourceType: foresttype.ResourceTypeForest,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      w.Uin,
				Action:       foresttype.ActionManage,
			})
		} else if !targetPerm.ManagePerm && hasManagePerm {
			forestIDsToRemoveManager = append(forestIDsToRemoveManager, forestID)
		}
		// 2. 收集【使用权限】的变更
		hasUsePerm := w.hasPermission(w.CurrUseFm, forestID)
		if targetPerm.UserPerm && !hasUsePerm {
			scopesToCreate = append(scopesToCreate, &foresttype.KeResourceScope{
				ResourceID:   forestID,
				ResourceType: foresttype.ResourceTypeForest,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      w.Uin,
				Action:       foresttype.ActionView,
			})
		} else if !targetPerm.UserPerm && hasUsePerm {
			forestIDsToDeleteScope = append(forestIDsToDeleteScope, forestID)
		}
	}
	// --- 阶段2: 批量执行数据库操作 ---
	// 批量授予【管理权限】
	return dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if len(forestIDsToAddManager) > 0 {
			if err := tx.CreateInBatches(&forestIDsToAddManager, len(forestIDsToAddManager)).Error; err != nil {
				return fmt.Errorf("批量授予管理权限失败: %w", err)
			}
		}
		// 批量撤销【管理权限】
		if len(forestIDsToRemoveManager) > 0 {
			if len(forestIDsToRemoveManager) > 0 {
				if err := tx.Where("resource_id IN ?", forestIDsToRemoveManager).
					Where("resource_type = ?", foresttype.ResourceTypeForest).
					Where("scope_type = ? AND scope_id = ?", foresttype.ScopeTypeUser, w.Uin).
					Where("action = ?", foresttype.ActionManage).
					Delete(&foresttype.KeResourceScope{}).Error; err != nil {
					return fmt.Errorf("批量撤销管理权限失败: %w", err)
				}
			}
		}
		// 批量创建"使用权限"
		if len(scopesToCreate) > 0 {
			if err := tx.CreateInBatches(&scopesToCreate, len(scopesToCreate)).Error; err != nil {
				return fmt.Errorf("批量授予使用权限失败: %w", err)
			}
		}
		// 批量删除"使用权限"
		if len(forestIDsToDeleteScope) > 0 {
			if err := tx.Where("resource_id IN ?", forestIDsToDeleteScope).
				Where("resource_type = ?", foresttype.ResourceTypeForest).
				Where("scope_type = ? AND scope_id = ?", foresttype.ScopeTypeUser, w.Uin).
				Where("action = ?", foresttype.ActionView).
				Delete(&foresttype.KeResourceScope{}).Error; err != nil {
				return fmt.Errorf("批量撤销使用权限失败: %w", err)
			}
		}
		return nil
	})
}

func HasAct(ctx context.Context, uin, resourceID uint, resourceType foresttype.ResourceType, action foresttype.ActionType) bool {
	var c int64
	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("action = ?", action).
		Where("resource_id = ?", resourceID).
		Where("resource_type = ?", resourceType).
		Where("scope_type = ?", chattype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "check uin[%v] act[%v] for resource[type:%v|id:%v] err:%v", uin, action, resourceType, resourceID, err)
		return false
	}
	return c > 0
}

func HasManageAct(ctx context.Context, uin, resourceID uint, resourceType foresttype.ResourceType) bool {
	return HasAct(ctx, uin, resourceID, resourceType, foresttype.ActionManage)
}
func HasViewAct(ctx context.Context, uin, resourceID uint, resourceType foresttype.ResourceType) bool {
	return HasAct(ctx, uin, resourceID, resourceType, foresttype.ActionView)
}

func GetManageList(ctx context.Context, uin uint, resourceType foresttype.ResourceType) []uint {
	var manageScopesIDs []uint
	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("resource_type = ?", resourceType).
		Where("action = ?", foresttype.ActionManage).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Where("deleted_at IS NULL").
		Pluck("distinct resource_id", &manageScopesIDs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "GetManageList err:%v", err)
		return nil
	}
	return manageScopesIDs
}

// UpdateResourceScope 更新资源的权限范围
func UpdateResourceScope(ctx context.Context, tx *gorm.DB, resourceID uint, resourceType foresttype.ResourceType, newViewsIDs, newManagerIDs []uint, scope foresttype.PublicScope, companyID uint) error {
	if resourceID == 0 {
		logs.WarnContextf(ctx, "[UpdateResourceScope] resourceID is required")
		return nil
	}
	logs.InfoContextf(ctx, "UpdateResourceScope has accept newViewsIDs:%v newManagerIDs:%v", newViewsIDs, newManagerIDs)
	if len(newViewsIDs) == 0 && len(newManagerIDs) == 0 {
		logs.InfoContextf(ctx, "[UpdateResourceScope] no new viewers and managers")
		return nil
	}
	var scps []*foresttype.KeResourceScope

	if err := tx.WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("resource_type = ?", resourceType).
		Where("resource_id = ?", resourceID).
		Where("deleted_at IS NULL").
		Where("scope_type in (?)", []foresttype.ScopeType{foresttype.ScopeTypeUser, foresttype.ScopeTypeCompany}).
		Find(&scps).Error; err != nil {
		logs.ErrorContextf(ctx, "UpdateResourceScope Find scps err:%v", err)
		return err
	}

	var (
		viewers  = make(map[uint]struct{})
		managers = make(map[uint]struct{})

		newViewers  = make(map[uint]struct{})
		newManagers = make(map[uint]struct{})

		toAdd                          []*foresttype.KeResourceScope
		toDelUserView, toDelUserManage []uint

		cmpFlag uint
	)

	for _, v := range newViewsIDs {
		newViewers[v] = struct{}{}
	}

	for _, v := range newManagerIDs {
		newManagers[v] = struct{}{}
	}

	for _, v := range scps {
		if v.ScopeType == foresttype.ScopeTypeCompany {
			cmpFlag = v.ID
		} else {
			switch v.Action {
			case foresttype.ActionView:
				viewers[v.ScopeID] = struct{}{}
			case foresttype.ActionManage:
				managers[v.ScopeID] = struct{}{}
			}
		}

	}

	logs.InfoContextf(ctx, "UpdateResourceScope viewers:%v managers:%v", viewers, managers)

	//================================get diff=================================
	//TO ADD
	for _, v := range newViewsIDs {
		//don's exist in old -> need to add
		if _, exist := viewers[v]; !exist {
			toAdd = append(toAdd, &foresttype.KeResourceScope{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      v,
				Action:       foresttype.ActionView,
			})
		}
	}

	logs.InfoContextf(ctx, "UpdateResourceScope newManagers:%v", newManagers)

	for _, v := range newManagerIDs {
		//don's exist in old -> need to add
		if _, exist := managers[v]; !exist {
			toAdd = append(toAdd, &foresttype.KeResourceScope{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      v,
				Action:       foresttype.ActionManage,
			})
		}
	}

	logs.InfoContextf(ctx, "UpdateResourceScope newViewers:%v", newViewers)

	//TO DEL
	for k := range viewers {
		//don't exist in new -> need to del
		if _, exist := newViewers[k]; !exist {
			toDelUserView = append(toDelUserView, k)
		}
	}
	for k := range managers {
		//don't exist in new -> need to del
		if _, exist := newManagers[k]; !exist {
			toDelUserManage = append(toDelUserManage, k)
		}
	}

	//1. custom -> company
	if cmpFlag == 0 && scope == foresttype.PublicScopeCompany {
		toAdd = append(toAdd, &foresttype.KeResourceScope{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ScopeType:    foresttype.ScopeTypeCompany,
			ScopeID:      companyID,
			Action:       foresttype.ActionView,
		})
	}

	logs.InfoContextf(ctx, "UpdateResourceScope toAdd:%v toDelUserView:%v toDelUserManage:%v", toAdd, toDelUserView, toDelUserManage)

	return tx.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(toAdd) > 0 {
			if err := tx.WithContext(ctx).CreateInBatches(toAdd, len(toAdd)).Error; err != nil {
				logs.ErrorContextf(ctx, "UpdateResourceScope CreateInBatches err:%v", err)
				return err
			}
		}
		if len(toDelUserView)+len(toDelUserManage) > 0 {
			logs.InfoContextf(ctx, "UpdateResourceScope toDelUserView:%v toDelUserManage:%v", toDelUserView, toDelUserManage)
			if err := tx.WithContext(ctx).
				Where("scope_type = ?", foresttype.ScopeTypeUser).
				Where("resource_id = ?", resourceID).
				Where("(action = ? AND scope_id IN ?) OR (action = ? AND scope_id IN ?)",
					foresttype.ActionView, toDelUserView,
					foresttype.ActionManage, toDelUserManage).
				Where("resource_type = ?", resourceType).
				Delete(&foresttype.KeResourceScope{}).
				Error; err != nil {
				logs.ErrorContextf(ctx, "UpdateResourceScope Delete err:%v", err)
				return err
			}
		}

		//2. company -> custom
		if cmpFlag > 0 && scope != foresttype.PublicScopeCompany {
			logs.InfoContextf(ctx, "UpdateResourceScope toDelCompany:%v", cmpFlag)
			if err := tx.WithContext(ctx).Where("id = ?", cmpFlag).Delete(&foresttype.KeResourceScope{}).Error; err != nil {
				logs.ErrorContextf(ctx, "UpdateResourceScope Delete err:%v", err)
				return err
			}
		}

		return nil
	})
}

func DeleteResourceScope(ctx context.Context, id uint, resType foresttype.ResourceType, tx *gorm.DB) (err error) {
	if err := tx.WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", resType).
		Where("resource_id = ?", id).
		Delete(&foresttype.KeResourceScope{}).Error; err != nil {
		logs.ErrorContextf(ctx, "DeleteResourceScope err:%v", err)
	}
	return err
}

func GetViewableResourceIDs(ctx context.Context, uin, companyID uint, resourceType foresttype.ResourceType) []uint {
	var resourceIDs []uint
	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", resourceType).
		Where("(action = ? AND scope_type = ?) OR "+
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			"(action = ? AND scope_type = ? AND scope_id = ?)",
			foresttype.ActionView, foresttype.ScopeTypePublic,
			foresttype.ActionManage, foresttype.ScopeTypeUser, uin,
			foresttype.ActionView, foresttype.ScopeTypeUser, uin,
			foresttype.ActionView, foresttype.ScopeTypeCompany, companyID).
		Pluck("distinct resource_id", &resourceIDs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "GetViewableResourceIDs err:%v", err)
		return nil
	}
	return resourceIDs
}

// GetManageScopeList 获取管理权限范围列表
func GetManageScopeList(ctx context.Context, resourceType foresttype.ResourceType, resourceID uint) (managerIDs []uint, scopeIDs []uint, err error) {
	rss := []*foresttype.KeResourceScope{}
	if err := dbutil.Knownow().
		Where("deleted_at IS NULL").
		Where("resource_type = ?", resourceType).
		Where("resource_id = ?", resourceID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Find(&rss).Error; err != nil {
		logs.ErrorContextf(ctx, "GetForestWithPerm failed: %v", err)
		return nil, nil, err
	}

	for _, v := range rss {
		switch v.Action {
		case foresttype.ActionManage:
			managerIDs = append(managerIDs, v.ScopeID)
		case foresttype.ActionView:
			scopeIDs = append(scopeIDs, v.ScopeID)
		}
	}
	logs.InfoContextf(ctx, "GetManageScopeList managerIDs:%v scopeIDs:%v", managerIDs, scopeIDs)
	return managerIDs, scopeIDs, nil
}

// GetALLManageScopeList 获取管理权限范围列表
func GetALLManageScopeList(ctx context.Context, resourceType foresttype.ResourceType, resourceID uint) (managerIDs []uint, scopeIDs []uint, err error) {
	rss := []*foresttype.KeResourceScope{}
	if err := dbutil.Knownow().WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", resourceType).
		Where("resource_id = ?", resourceID).
		// Where("scope_type = ?", foresttype.ScopeTypeUser).
		Find(&rss).Error; err != nil {
		logs.ErrorContextf(ctx, "GetForestWithPerm failed: %v", err)
		return nil, nil, err
	}

	for _, v := range rss {
		switch v.Action {
		case foresttype.ActionManage:
			managerIDs = append(managerIDs, v.ScopeID)
		case foresttype.ActionView:
			scopeIDs = append(scopeIDs, v.ScopeID)
		}
	}
	return managerIDs, scopeIDs, nil
}

// ResourceScopePerm 单个资源的权限信息
type ResourceScopePerm struct {
	ManagerIDs []uint
	ViewIDs    []uint
}

// BatchGetManageScopeList 批量获取多个资源的权限范围列表
// 返回 resourceID -> ResourceScopePerm 的映射
func BatchGetManageScopeList(ctx context.Context, resourceType foresttype.ResourceType, resourceIDs []uint) (map[uint]*ResourceScopePerm, error) {
	if len(resourceIDs) == 0 {
		return make(map[uint]*ResourceScopePerm), nil
	}

	var rss []*foresttype.KeResourceScope
	if err := dbutil.Knownow().WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", resourceType).
		Where("resource_id IN ?", resourceIDs).
		Find(&rss).Error; err != nil {
		logs.ErrorContextf(ctx, "BatchGetManageScopeList failed: %v", err)
		return nil, err
	}

	result := make(map[uint]*ResourceScopePerm, len(resourceIDs))
	for _, id := range resourceIDs {
		result[id] = &ResourceScopePerm{
			ManagerIDs: make([]uint, 0),
			ViewIDs:    make([]uint, 0),
		}
	}

	for _, s := range rss {
		perm, ok := result[s.ResourceID]
		if !ok {
			perm = &ResourceScopePerm{
				ManagerIDs: make([]uint, 0),
				ViewIDs:    make([]uint, 0),
			}
			result[s.ResourceID] = perm
		}
		switch s.Action {
		case foresttype.ActionManage:
			perm.ManagerIDs = append(perm.ManagerIDs, s.ScopeID)
		case foresttype.ActionView:
			perm.ViewIDs = append(perm.ViewIDs, s.ScopeID)
		}
	}

	return result, nil
}

// ResourceScopeUpdate 资源权限更新请求
type ResourceScopeUpdate struct {
	ResourceID   uint
	ResourceType foresttype.ResourceType
	ViewIDs      []uint // 期望的查看权限用户列表
	ManagerIDs   []uint // 期望的管理权限用户列表
	PublicScope  foresttype.PublicScope
	CompanyID    uint
}

// BatchUpdateResourceScope 批量更新多个资源的权限范围
// 相比循环调用 UpdateResourceScope，此函数只查询一次数据库获取所有资源的当前权限
func BatchUpdateResourceScope(ctx context.Context, tx *gorm.DB, updates []*ResourceScopeUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 收集所有需要查询的资源 ID 和类型
	resourceKeys := make(map[string]struct{})
	resourceIDs := make([]uint, 0, len(updates))
	resourceTypes := make([]foresttype.ResourceType, 0)
	resourceTypeSet := make(map[foresttype.ResourceType]struct{})

	for _, u := range updates {
		if u.ResourceID == 0 {
			continue
		}
		key := fmt.Sprintf("%s_%d", u.ResourceType, u.ResourceID)
		if _, exists := resourceKeys[key]; !exists {
			resourceKeys[key] = struct{}{}
			resourceIDs = append(resourceIDs, u.ResourceID)
			if _, ok := resourceTypeSet[u.ResourceType]; !ok {
				resourceTypeSet[u.ResourceType] = struct{}{}
				resourceTypes = append(resourceTypes, u.ResourceType)
			}
		}
	}

	if len(resourceIDs) == 0 {
		return nil
	}

	// 一次性查询所有资源的当前权限
	var existingScopes []*foresttype.KeResourceScope
	if err := tx.WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("resource_type IN ?", resourceTypes).
		Where("resource_id IN ?", resourceIDs).
		Where("deleted_at IS NULL").
		Where("scope_type IN ?", []foresttype.ScopeType{foresttype.ScopeTypeUser, foresttype.ScopeTypeCompany}).
		Find(&existingScopes).Error; err != nil {
		logs.ErrorContextf(ctx, "BatchUpdateResourceScope query existing scopes err: %v", err)
		return err
	}

	// 构建 resourceKey -> scopes 的映射
	scopeMap := make(map[string][]*foresttype.KeResourceScope)
	for _, s := range existingScopes {
		key := fmt.Sprintf("%s_%d", s.ResourceType, s.ResourceID)
		scopeMap[key] = append(scopeMap[key], s)
	}

	// 收集所有待执行的变更
	var toAdd []*foresttype.KeResourceScope
	var toDelIDs []uint

	for _, update := range updates {
		if update.ResourceID == 0 {
			continue
		}
		key := fmt.Sprintf("%s_%d", update.ResourceType, update.ResourceID)
		scopes := scopeMap[key]

		// 构建当前权限的 map
		viewers := make(map[uint]uint)  // scopeID -> recordID
		managers := make(map[uint]uint) // scopeID -> recordID
		var companyRecordID uint

		for _, s := range scopes {
			switch s.ScopeType {
			case foresttype.ScopeTypeCompany:
				companyRecordID = s.ID
			case foresttype.ScopeTypeUser:
				switch s.Action {
				case foresttype.ActionView:
					viewers[s.ScopeID] = s.ID
				case foresttype.ActionManage:
					managers[s.ScopeID] = s.ID
				}
			}
		}

		// 构建期望权限的 set
		newViewers := make(map[uint]struct{})
		newManagers := make(map[uint]struct{})
		for _, v := range update.ViewIDs {
			newViewers[v] = struct{}{}
		}
		for _, v := range update.ManagerIDs {
			newManagers[v] = struct{}{}
		}

		// 计算差异 - 需要添加的
		for viewerID := range newViewers {
			if _, exists := viewers[viewerID]; !exists {
				toAdd = append(toAdd, &foresttype.KeResourceScope{
					ResourceType: update.ResourceType,
					ResourceID:   update.ResourceID,
					ScopeType:    foresttype.ScopeTypeUser,
					ScopeID:      viewerID,
					Action:       foresttype.ActionView,
				})
			}
		}
		for managerID := range newManagers {
			if _, exists := managers[managerID]; !exists {
				toAdd = append(toAdd, &foresttype.KeResourceScope{
					ResourceType: update.ResourceType,
					ResourceID:   update.ResourceID,
					ScopeType:    foresttype.ScopeTypeUser,
					ScopeID:      managerID,
					Action:       foresttype.ActionManage,
				})
			}
		}

		// 计算差异 - 需要删除的
		for viewerID, recordID := range viewers {
			if _, exists := newViewers[viewerID]; !exists {
				toDelIDs = append(toDelIDs, recordID)
			}
		}
		for managerID, recordID := range managers {
			if _, exists := newManagers[managerID]; !exists {
				toDelIDs = append(toDelIDs, recordID)
			}
		}

		// 处理公司权限切换
		if companyRecordID == 0 && update.PublicScope == foresttype.PublicScopeCompany {
			// 需要添加公司权限
			toAdd = append(toAdd, &foresttype.KeResourceScope{
				ResourceType: update.ResourceType,
				ResourceID:   update.ResourceID,
				ScopeType:    foresttype.ScopeTypeCompany,
				ScopeID:      update.CompanyID,
				Action:       foresttype.ActionView,
			})
		} else if companyRecordID > 0 && update.PublicScope != foresttype.PublicScopeCompany {
			// 需要删除公司权限
			toDelIDs = append(toDelIDs, companyRecordID)
		}
	}

	logs.InfoContextf(ctx, "BatchUpdateResourceScope toAdd: %d, toDel: %d", len(toAdd), len(toDelIDs))

	// 批量执行数据库操作
	if len(toAdd) > 0 {
		if err := tx.WithContext(ctx).CreateInBatches(toAdd, 100).Error; err != nil {
			logs.ErrorContextf(ctx, "BatchUpdateResourceScope CreateInBatches err: %v", err)
			return fmt.Errorf("批量创建权限失败: %w", err)
		}
	}

	if len(toDelIDs) > 0 {
		if err := tx.WithContext(ctx).
			Where("id IN ?", toDelIDs).
			Delete(&foresttype.KeResourceScope{}).Error; err != nil {
			logs.ErrorContextf(ctx, "BatchUpdateResourceScope Delete err: %v", err)
			return fmt.Errorf("批量删除权限失败: %w", err)
		}
	}

	return nil
}
