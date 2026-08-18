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

package yyguauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/ygpkg/yg-go/logs"
	"github.com/dgrijalva/jwt-go"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
)

// AuthYYGuToken 校验yygu token有效性
func AuthYYGuToken(ctx context.Context, token string) (*auth.LoginStatus, error) {
	ls := &auth.LoginStatus{}
	token = strings.TrimPrefix(token, auth.AuthBearer)
	token = strings.TrimSpace(token)
	ls.Token = token
	claims := new(auth.UserClaims)
	_, err := jwt.ParseWithClaims(ls.Token, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Claims == nil {
			return nil, fmt.Errorf("token claims is nil")
		}
		c, ok := token.Claims.(*auth.UserClaims)
		if !ok {
			return nil, fmt.Errorf("token claims is not UserClaims")
		}

		return auth.GetJwtSecret(c.Issuer)
	})
	if err != nil {
		logs.ErrorContextf(ctx,"[manager_auth] parse claims failed.",
			"error", err, "token", ls.Token)
		ls.Err = err
		ls.State = auth.StateFailed
		return nil, err
	}
	ls.State = auth.StateSucc
	ls.Claim = claims
	return ls, nil
}
