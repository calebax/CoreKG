package chatagent

import (
	"context"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

func CanViewAgent(ctx context.Context, agID, uin, companyID uint) bool {
	var c int64
	ag, err := GetChatAgentByID(ctx, agID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetChatAgentByID(%d) failed: %v", agID, err)
		return false
	}

	if ag.PublicScope == chattype.PublicScopePublic ||
		(ag.PublicScope == chattype.PublicScopeCompany && ag.CompanyID == companyID) ||
		(ag.PublicScope == chattype.PublicScopePrivate && ag.Uin == uin) ||
		(ag.PublishStatus == chattype.StatusDraft && ag.Uin == uin) {
		return true
	}

	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeAgent).
		Where("resource_id = ?", agID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "get resource_scope faild %v", err)
		return false
	}

	return c > 0
}

func CanManageAgent(ctx context.Context, agID, uin uint) bool {
	var c int64
	ag, err := GetChatAgentByID(ctx, agID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetChatAgentByID(%d) failed: %v", agID, err)
		return false
	}

	if (ag.PublicScope == chattype.PublicScopePrivate && ag.Uin == uin) ||
		(ag.PublishStatus == chattype.StatusDraft && ag.Uin == uin) {
		return true
	}

	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeAgent).
		Where("resource_id = ?", agID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Where("action = ?", foresttype.ActionManage).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "get resource_scope faild %v", err)
		return false
	}

	return c > 0
}
