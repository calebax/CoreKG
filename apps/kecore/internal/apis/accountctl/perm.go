package accountctl

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

// GetForestPermSet 获取用户知识库权限集
// @Tags 账户系统
// @Summary 获取用户权限集
// @Description 获取用户权限集
// @Router /forest.GetForestPermSet [post]
// @Param user body GetForestPermSetRequest true "入参"
// @Success 200 {object} GetForestPermSetResponse "返回值"
func GetForestPermSet(ctx *gin.Context, req *GetForestPermSetRequest, resp *GetForestPermSetResponse) {
	if !req.Valid(resp) {
		logs.WarnContextf(ctx, "GetForestPermSet validate params failed")
		return
	}
	//*get uin about do this action
	actUin := runtime.Uin(ctx)
	actEmp, err := employee.GetEmployeeByUin(actUin)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetForestPermSet] GetEmployeeByUin[uin=%v] failed ,err %s", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_operator_employee_failed" // 获取操作者员工信息失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	if actEmp.SysRole != accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[GetForestPermSet] [uin=%v,sysRole=%v] failed ,err %s", actUin, actEmp.SysRole, err)
		resp.Code = errcode.ErrCode_Unauthorized
		resp.Message = "kecore_no_permission" // 无操作权限
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	//*get user forests set
	frss, err := perm.GetForestsByCompanyID(actEmp.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetForestPermSet] [GetForestsByCompanyID] failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_company_forests_failed" // 获取公司森林失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	if req.Request.Uin == 0 {
		resp.Response.PermSet = make([]*perm.Set, 0, len(frss))
		for _, f := range frss {
			resp.Response.PermSet = append(resp.Response.PermSet, &perm.Set{
				Forest:     f,
				ManagePerm: false,
				UserPerm:   false,
			})
		}
		return
	}

	emp, err := employee.GetEmployeeByUin(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetForestPermSet] GetEmployeeByUin[uin=%v] failed ,err %s", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_employee_failed" // 获取员工信息失败
		return
	}
	if emp.CompanyID != actEmp.CompanyID {
		logs.ErrorContextf(ctx, "[GetForestPermSet] Incorrect emp[actuin=%v/targuin=%v]", actUin, emp.Uin)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_company_user" // 未知公司用户无权限
		return
	}

	wps := perm.NewWrapperPermSet(ctx, req.Request.Uin, emp.CompanyID, nil, frss)
	if err = wps.BuildCurrentPermMap(); err != nil {
		logs.ErrorContextf(ctx, "[GetForestPermSet] BuildCurrentPermMap[uin=%v] failed ,err %s", req.Request.Uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_build_user_perm_set_failed" // 构建用户知识库权限集失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}
	resp.Response.PermSet = wps.GetCurrPermSet()
}

// ModifyForestPermSet 更新用户知识库权限集
// @Tags 账户系统
// @Summary 更新用户知识库权限集
// @Description 更新用户知识库权限集
// @Router /forest.ModifyForestPermSet [post]
// @Param user body ModifyForestPermSetRequest true "入参"
// @Success 200 {object} ModifyForestPermSetResponse "返回值"
func ModifyForestPermSet(ctx *gin.Context, req *ModifyForestPermSetRequest, resp *ModifyForestPermSetResponse) {
	if !req.Valid(resp) {
		logs.WarnContextf(ctx, "ModifyForestPermSet validate params failed")
		return
	}
	//*get uin about do this action
	actUin := runtime.Uin(ctx)

	actEmp, err := employee.GetEmployeeByUin(actUin)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] GetEmployeeByUin[uin=%v] failed ,err %s", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_operator_employee_failed" // 获取操作者员工信息失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	if actEmp.SysRole != accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] [uin=%v,sysRole=%v] failed ,err %s", actUin, actEmp.SysRole, err)
		resp.Code = errcode.ErrCode_Unauthorized
		resp.Message = "kecore_no_permission" // 无操作权限
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	emp, err := employee.GetEmployeeByUin(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] GetEmployeeByUin[uin=%v] failed ,err %s", actUin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_target_employee_failed" // 获取待变更员工信息失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}
	if emp.CompanyID != actEmp.CompanyID {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] Incorrect emp[actuin=%v/targuin=%v]", actUin, emp.Uin)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_company_user" // 未知公司用户无权限
		return
	}

	wps := perm.NewWrapperPermSet(ctx, req.Request.Uin, emp.CompanyID, req.Request.PermSet, nil)
	if err = wps.BuildCurrentPermMap(); err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] BuildCurrentPermMap failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_build_user_perm_set_failed" // 构建用户知识库权限集失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	if err = wps.ApplyChanges(); err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] ApplyChanges failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_perm_change_failed" // 权限变更失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	// 收集需要同步权限的 forest IDs
	forestIDs := make([]uint, 0)
	for _, p := range req.Request.PermSet {
		if p.ActOption == perm.ActOptionUpdate {
			forestIDs = append(forestIDs, p.Forest.ID)
		}
	}

	if len(forestIDs) == 0 {
		return
	}

	// 批量获取 forest 对应的 graph
	graphMap, err := graph.GetGraphsByForestIDs(ctx, forestIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] GetGraphsByForestIDs failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graphs_failed"
		return
	}

	// 批量获取所有 forest 当前的权限（已更新后的状态）
	forestPerms, err := perm.BatchGetManageScopeList(ctx, foresttype.ResourceTypeForest, forestIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] BatchGetManageScopeList failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_perm_failed"
		return
	}

	// 批量获取 forest 信息（获取 PublicScope）
	var forests []*foresttype.KnownowForest
	if err := dbutil.Knownow().WithContext(ctx).
		Where("id IN ?", forestIDs).
		Where("deleted_at IS NULL").
		Find(&forests).Error; err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] get forests failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forests_failed"
		return
	}
	forestInfoMap := make(map[uint]*foresttype.KnownowForest, len(forests))
	for _, f := range forests {
		forestInfoMap[f.ID] = f
	}

	// 构建批量更新请求
	graphUpdates := make([]*perm.ResourceScopeUpdate, 0)
	graphsToUpdate := make([]*foresttype.ForestGraphInfo, 0)

	for _, forestID := range forestIDs {
		graphInfo, ok := graphMap[forestID]
		if !ok {
			logs.WarnContextf(ctx, "[ModifyForestPermSet] graph not found for forest: %d", forestID)
			continue
		}

		forestInfo, ok := forestInfoMap[forestID]
		if !ok {
			logs.WarnContextf(ctx, "[ModifyForestPermSet] forest info not found: %d", forestID)
			continue
		}

		forestPerm, ok := forestPerms[forestID]
		if !ok {
			forestPerm = &perm.ResourceScopePerm{}
		}

		// 同步 PublicScope 到 graph
		graphInfo.PublicScope = forestInfo.PublicScope
		graphsToUpdate = append(graphsToUpdate, graphInfo)

		graphUpdates = append(graphUpdates, &perm.ResourceScopeUpdate{
			ResourceID:   graphInfo.ID,
			ResourceType: foresttype.ResourceTypeGraph,
			ViewIDs:      forestPerm.ViewIDs,
			ManagerIDs:   forestPerm.ManagerIDs,
			PublicScope:  forestInfo.PublicScope,
			CompanyID:    graphInfo.CompanyID,
		})
	}

	// 在事务中执行批量更新
	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量更新 graph 权限
		if err := perm.BatchUpdateResourceScope(ctx, tx, graphUpdates); err != nil {
			logs.ErrorContextf(ctx, "[ModifyForestPermSet] BatchUpdateResourceScope failed: %v", err)
			return err
		}

		// 批量更新 graph 的 PublicScope
		for _, g := range graphsToUpdate {
			if err := graph.UpdateGraph(ctx, g, tx); err != nil {
				logs.WarnContextf(ctx, "[ModifyForestPermSet] UpdateGraph failed: %v", err)
				return err
			}
		}
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "[ModifyForestPermSet] transaction failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_update_resource_scope_failed"
		return
	}
}

// UpdateForestWithPerm 更新知识库(新权限)
// @Tags 知识森林
// @Summary 更新知识库(新权限)
// @Description 更新知识库(新权限)
// @Router /forest.UpdateForestWithPerm [post]
// @Param user body UpdateForestWithPermRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func UpdateForestWithPerm(ctx *gin.Context, req *UpdateForestWithPermRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "UpdateForestWithPerm validate params failed")
		return
	}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	// 权限校验
	forests, err := forest.GetForestByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		logs.ErrorContextf(ctx, "GetForestByID(%v) failed, err %v", req.Request.ID, err)
		return
	}
	if !perm.HasManageAct(ctx, uin, req.Request.ID, foresttype.ResourceTypeForest) {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_no_permission_update_resource" // 无权限修改此资源
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, req.Request.ID)
		return
	}

	// do not append manager action list with admins in company
	if req.Request.PublicScope == foresttype.PublicScopeCompany {
		req.Request.ScopeIDs = types.NewUintArray([]uint{})
	}

	// 去重
	req.Request.ManagerIDs.RemoveDuplicates()
	req.Request.ScopeIDs.RemoveDuplicates()

	uins := types.NewUintArray(append(req.Request.ManagerIDs.Slice(), req.Request.ScopeIDs.Slice()...))
	uins.RemoveDuplicates()

	us := uins.Slice()
	if !employee.CheckUinsValid(ctx, us, companyID) {
		logs.ErrorContextf(ctx, "CheckUinsValid: exist no-local company[%v] uin in uins[%v]", companyID, us)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_employee_id" // 存在非法员工id
		return
	}

	if err = dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := forest.UpdateForestWithPerm(ctx, tx, &req.Request.WithPermItem, forests)
		if err != nil {
			logs.ErrorContextf(ctx, "UpdateForestWithPerm failed: %v", err)
			return err
		}
		// 同步修改图谱权限
		graphInfo, err := graph.GetForestGraph(ctx, forests.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.WarnContextf(ctx, "GetForestGraph not found, forestID: %v", forests.ID)
				return nil
			}
			logs.ErrorContextf(ctx, "GetForestGraph failed: %v", err)
			return err
		}
		graphInfo.PublicScope = forests.PublicScope
		err = graph.UpdateGraph(ctx, graphInfo, tx)
		if err != nil {
			logs.ErrorContextf(ctx, "Save graphinfo failed: %v", err)
			return err
		}
		return perm.UpdateResourceScope(ctx, tx, graphInfo.ID, foresttype.ResourceTypeGraph, req.Request.ScopeIDs.Slice(), req.Request.ManagerIDs.Slice(), graphInfo.PublicScope, graphInfo.CompanyID)

	}); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		if errors.Is(err, forest.ErrNameAlreadyExists) {
			resp.Message = "kecore_forest_name_exists"
		} else {
			resp.Message = "kecore_update_resource_failed"
		}
		logs.ErrorContextf(ctx, "UpdateForestWithPerm failed: %v", err)
		return
	}
}
