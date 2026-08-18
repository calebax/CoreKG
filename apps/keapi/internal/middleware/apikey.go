package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

// RequireAPIKeyPrivilege 仅要求 API Key 存在且有效，并注入用户上下文。
func RequireAPIKeyPrivilege(ctx *gin.Context) {
	ls, apiKeyID, _, err := accountmds.MustLoginAPIKey(ctx)
	if err != nil {
		return
	}

	keyInfo, err := accountmds.VerifyAPIKey(ctx, apiKeyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] verify API key failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	ls.SetID(global.CtxKeyUin, keyInfo.Uin)
	ls.SetID(global.CtxKeyCompanyID, keyInfo.CompanyID)
	ctx.Set(global.CtxKeyAPIKey, keyInfo.APIKey)
	ctx.Next()
}
