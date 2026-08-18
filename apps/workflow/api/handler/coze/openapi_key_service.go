package coze

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/insmtx/corekg/apps/workflow/api/internal/httputil"
	"github.com/insmtx/corekg/apps/workflow/application/base/ctxutil"
	"github.com/ygpkg/yg-go/logs"
)

type getCurrentUserAPIKeyResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		APIKey   string `json:"api_key"`
		ExpireAt int64  `json:"expire_at"`
	} `json:"data"`
}

// GetOrCreateCurrentUserAPIKey ensures the current user has an API key valid for at least 3 days.
// @router /api/permission_api/pat/get_personal_access_token [POST]
func GetOrCreateCurrentUserAPIKey(ctx context.Context, c *app.RequestContext) {
	userID := ctxutil.GetUIDFromCtx(ctx)
	if userID == nil || *userID == 0 {
		logs.WarnContextf(ctx, "GetOrCreateCurrentUserAPIKey missing user id in context")
		httputil.Unauthorized(c, "session required")
		return
	}

	const (
		minValidDuration = 72 * time.Hour
		newTokenValidFor = 72 * time.Hour
		tokenNameForUser = "corekg_default_api_key"
	)

	apiKey, expireAt, err := getValidOrCreateApiKey(ctx, *userID, minValidDuration, newTokenValidFor, tokenNameForUser)
	if err != nil {
		logs.ErrorContextf(ctx, "GetOrCreateCurrentUserAPIKey failed user_id=%d: %v", *userID, err)
		internalServerErrorResponse(ctx, c, err)
		return
	}

	resp := getCurrentUserAPIKeyResponse{
		Code: 0,
		Msg:  "success",
	}
	resp.Data.APIKey = apiKey
	resp.Data.ExpireAt = expireAt

	c.JSON(consts.StatusOK, resp)
}
