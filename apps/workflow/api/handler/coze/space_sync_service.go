package coze

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/insmtx/corekg/apps/workflow/application/spacesync"
	"github.com/ygpkg/yg-go/logs"
)

// SpaceSyncWebhook Webhook 用于从 ROC 同步空间/成员信息
// @router /api/internal/space_sync [POST]
func SpaceSyncWebhook(ctx context.Context, c *app.RequestContext) {

	// 使用 spacesync 包的 Sync 函数执行同步
	if _, err := spacesync.Sync(ctx); err != nil {
		logs.ErrorContextf(ctx, "sync spaces/members failed: %v", err)
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 0,
		"msg":  "ok",
	})
}
