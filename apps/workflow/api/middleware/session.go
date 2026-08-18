/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/insmtx/corekg/apps/workflow/api/internal/httputil"
	"github.com/insmtx/corekg/apps/workflow/domain/user/entity"
	"github.com/insmtx/corekg/apps/workflow/utils/yyguauth"

	"github.com/insmtx/corekg/apps/workflow/bizpkg/config"
	"github.com/insmtx/corekg/apps/workflow/pkg/ctxcache"
	"github.com/insmtx/corekg/apps/workflow/pkg/errorx"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/types/consts"
	"github.com/insmtx/corekg/apps/workflow/types/errno"
)

var noNeedSessionCheckPath = map[string]bool{
	"/api/public/agent/external_token": true,
}

func SessionAuthMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		requestAuthType := ctx.GetInt32(RequestAuthTypeStr)
		if requestAuthType != int32(RequestAuthTypeWebAPI) {
			ctx.Next(c)
			return
		}

		path := string(ctx.GetRequest().URI().Path())
		if strings.HasPrefix(path, "/api/common/upload/") {
			ctx.Next(c)
			return
		}
		if noNeedSessionCheckPath[path] {
			ctx.Next(c)
			return
		}

		token := ctx.Request.Header.Get(HeaderAuthorizationKey)
		// open api auth
		if len(token) == 0 {
			httputil.InternalError(c, ctx,
				errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "missing authorization in header")))
			return
		}

		session, err := getEntitySession(c, ctx, token)
		if err != nil {
			logs.Errorf("[SessionAuthMW] validate yygu token failed, err: %v", err)
			httputil.InternalError(c, ctx, err)
			return
		}

		if session != nil {
			ctxcache.Store(c, consts.SessionDataKeyInCtx, session)
		}

		ctx.Next(c)
	}
}

func getEntitySession(c context.Context, ctx *app.RequestContext, token string) (*entity.Session, error) {
	acceptLanguage := string(ctx.Request.Header.Get("Accept-Language"))
	locale := "zh-CN"
	if acceptLanguage != "" {
		languages := strings.Split(acceptLanguage, ",")
		if len(languages) > 0 {
			locale = languages[0]
		}
	}
	ls, err := yyguauth.AuthYYGuToken(c, token)
	if err != nil {
		logs.ErrorContextf(c, "auth yygu token err: %v", err)
		return nil, err
	}

	return &entity.Session{
		UserID:    int64(ls.Claim.Uin),
		Locale:    locale,
		Token:     ls.Token,
		CreatedAt: time.Unix(ls.Claim.IssuedAt, 0),
		ExpiresAt: time.Unix(ls.Claim.ExpiresAt, 0),
	}, err
}

func AdminAuthMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		session, ok := ctxcache.Get[*entity.Session](c, consts.SessionDataKeyInCtx)
		if !ok {
			logs.Errorf("[AdminAuthMW] session data is nil")
			httputil.InternalError(c, ctx,
				errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session data is nil")))
			return
		}

		adminUins := os.Getenv("ADMIN_UINS")
		if adminUins == "" {
			baseConf, err := config.Base().GetBaseConfig(c)
			if err != nil {
				logs.Errorf("[AdminAuthMW] get base config failed, err: %v", err)
				httputil.InternalError(c, ctx, err)
				return
			}
			adminUins = baseConf.AdminEmails
		}

		if adminUins == "" {
			logs.WarnContextf(c, "[AdminAuthMW] admin uins is empty")
			ctx.Next(c)
			return
		}

		uinStr := strconv.FormatInt(session.UserID, 10)
		for _, uin := range strings.Split(adminUins, ",") {
			if strings.TrimSpace(uin) == uinStr {
				ctx.Next(c)
				return
			}
		}

		httputil.Unauthorized(ctx, "the account does not have permission to access")
	}
}
