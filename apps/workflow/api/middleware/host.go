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
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/insmtx/corekg/apps/workflow/pkg/ctxcache"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/types/consts"
	"github.com/insmtx/corekg/apps/workflow/utils/mysql/coresettings"
)

func SetHostMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		item, err := coresettings.GetCozeUrl()
		if err != nil {
			logs.Error(err.Error())
			return
		}
		host := item
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		ctxcache.Store(c, consts.HostKeyInCtx, host)
		ctxcache.Store(c, consts.RequestSchemeKeyInCtx, string(ctx.GetRequest().Scheme()))
		ctx.Next(c)
	}
}
