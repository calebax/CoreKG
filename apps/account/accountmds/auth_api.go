package accountmds

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// RequireAPIKeyPrivilege 需要拥有API权限
func RequireAPIKeyPrivilege(ctx *gin.Context) {
	ls, apiKeyID, apiPath, err := MustLoginAPIKey(ctx)
	if err != nil {
		return
	}

	keyInfo, err := VerifyAPIKey(ctx, apiKeyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] verify API key failed: %v", err)
		runtime.InternalError(ctx, "Unauthorized")
		return
	}
	if !VerifyAPIKeyAPIPrivilege(ctx, keyInfo, apiPath) {
		logs.WarnContextf(ctx, "[RequireAPIKeyPrivilege] API key [id:%v|key:%v] unauthorized for %s", keyInfo.ID, keyInfo.APIKey, apiPath)
		return
	}
	ls.SetID(global.CtxKeyUin, keyInfo.Uin)
	ls.SetID(global.CtxKeyCompanyID, keyInfo.CompanyID)
	ctx.Set(global.CtxKeyAPIKey, keyInfo.APIKey)
	ctx.Next()
}
