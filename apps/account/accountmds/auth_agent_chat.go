package accountmds

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// RequireAgentChatPrivilege chat.Agent权限
func RequireAgentChatPrivilege(ctx *gin.Context) {
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
	// 历史原因：兼容两种鉴权方式 TODO : 后续整理这块的鉴权
	switch keyInfo.ResourceType {
	case accounttype.ResourceTypeAgent:
		if !VerifyAPIKeyAgentPrivilege(ctx, keyInfo, ctx.Request.Body) {
			logs.WarnContextf(ctx, "[RequireAPIKeyPrivilege] API key %d unauthorized for %s", keyInfo.ID, apiPath)
			return
		}
	default:
		if !VerifyAPIKeyAPIPrivilege(ctx, keyInfo, apiPath) {
			logs.WarnContextf(ctx, "[RequireAPIKeyPrivilege] API key %d unauthorized for %s", keyInfo.ID, apiPath)
			return
		}
	}
	//TODO feat balance

	ls.SetID(global.CtxKeyUin, keyInfo.Uin)
	ls.SetID(global.CtxKeyCompanyID, keyInfo.CompanyID)
	ctx.Set(global.CtxKeyAPIKey, keyInfo.APIKey)

	//===========================================DONE===================================================
	ctx.Next()
}
