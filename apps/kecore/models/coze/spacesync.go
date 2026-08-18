package coze

import (
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/logs"
)

const spaceSyncPath = "/api/internal/space_sync"

// SpaceSync 调用 coze 的内部同步接口(需登录状态) 同步空间和成员信息
func SpaceSync(ctx *gin.Context) error {
	// 接口无需请求体，传空 map 以保持一致的 JSON 结构
	err := CozeRequest(ctx, spaceSyncPath, map[string]interface{}{}, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "coze space sync error: %v", err)
		return err
	}
	return nil
}

// SpaceSyncWithToken 调用 coze 内部同步接口，显式指定 token
func SpaceSyncWithToken(ctx *gin.Context, token string) error {
	// 接口无需请求体，传空 map 以保持一致的 JSON 结构
	err := CozeRequestWithToken(ctx, spaceSyncPath, map[string]interface{}{}, nil, token)
	if err != nil {
		logs.ErrorContextf(ctx, "coze space sync error: %v", err)
		return err
	}
	return nil
}
