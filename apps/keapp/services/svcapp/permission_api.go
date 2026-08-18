package svcapp

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
)

func CheckAppPermission(ctx context.Context, uin, appID uint, action foresttype.ActionType) bool {
	return perm.HasAct(ctx, uin, appID, foresttype.ResourceTypeApp, action)
}

func CheckAppManagePermission(ctx context.Context, uin, appID uint) bool {
	return perm.HasManageAct(ctx, uin, appID, foresttype.ResourceTypeApp)
}
