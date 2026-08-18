package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapp/services/svcapp"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

type contextKey string

const appIDContextKey contextKey = "app_id"

type appIDRequest struct {
	Request struct {
		AppID uint `json:"app_id"`
	} `json:"request"`
}

func AppContextMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil {
			ctx.Next()
			return
		}

		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "AppContextMiddleware read body err: %v", err)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			ctx.Next()
			return
		}

		var req appIDRequest
		if err = json.Unmarshal(bodyBytes, &req); err == nil && req.Request.AppID != 0 {
			ctx.Set(string(appIDContextKey), req.Request.AppID)
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		ctx.Next()
	}
}

func GetAppID(ctx *gin.Context) uint {
	val, exists := ctx.Get(string(appIDContextKey))
	if !exists {
		return 0
	}
	appID, ok := val.(uint)
	if !ok {
		return 0
	}
	return appID
}

func RequireAppViewPerm(ctx *gin.Context) {
	appID := GetAppID(ctx)
	if appID == 0 {
		ctx.Next()
		return
	}

	uin := runtime.Uin(ctx)
	if !svcapp.CheckAppPermission(ctx, uin, appID, foresttype.ActionView) {
		logs.WarnContextf(ctx, "uin[%v] desire to view app[%v] but doesn't have perm", uin, appID)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    errcode.ErrCode_NoPermission,
			Message: "keapp_no_view_permission",
		})
		return
	}

	ctx.Next()
}

func RequireAppManagePerm(ctx *gin.Context) {
	appID := GetAppID(ctx)
	if appID == 0 {
		ctx.Next()
		return
	}

	uin := runtime.Uin(ctx)
	if !svcapp.CheckAppManagePermission(ctx, uin, appID) {
		logs.WarnContextf(ctx, "uin[%v] desire to manage app[%v] but doesn't have perm", uin, appID)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    errcode.ErrCode_NoPermission,
			Message: "keapp_no_manage_permission",
		})
		return
	}

	ctx.Next()
}
