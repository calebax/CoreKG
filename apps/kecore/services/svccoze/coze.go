package svccoze

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/coze"
	"github.com/ygpkg/yg-go/logs"
)

// SpaceSync 调用 coze 的内部同步接口，同步空间与成员信息
func SpaceSync(ctx *gin.Context) error {
	if err := coze.SpaceSync(ctx); err != nil {
		logs.ErrorContextf(ctx, "[SpaceSync] coze.SpaceSync failed, err: %v", err)
		return err
	}
	return nil
}

// SpaceSyncWithToken 调用 coze 的内部同步接口，显式指定 token
func SpaceSyncWithToken(ctx *gin.Context, token string) error {
	if err := coze.SpaceSyncWithToken(ctx, token); err != nil {
		logs.ErrorContextf(ctx, "[SpaceSyncWithToken] coze.SpaceSyncWithToken failed, err: %v", err)
		return err
	}
	return nil
}
