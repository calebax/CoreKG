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

package requestyygu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/insmtx/corekg/apps/workflow/application/base/ctxutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/utils/mysql/coresettings"
)

// YyguRequest 发起请求到 core-kg 服务
func YyguRequest(ctx context.Context, path string, request map[string]interface{}, targetStruct interface{}) error {
	baseurl, err := coresettings.GetCoreKGUrl()
	if err != nil {
		logs.ErrorContextf(ctx, "get core kg url error, %s", err.Error())
		return err
	}
	// baseurl := "http://CHANGE_ME_HOST/"
	apiurl, err := url.JoinPath(baseurl, path)
	if err != nil {
		logs.ErrorContextf(ctx, "join  url err %s", err.Error())
		return err
	}

	reqBody := map[string]interface{}{
		"request": request,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal request body error, %s", err.Error())
		return err
	}

	req, err := http.NewRequest("POST", apiurl, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	token, err := resolveCoreKGToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		var body string
		if resp != nil && resp.Body != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			body = string(bodyBytes)
		}
		logs.ErrorContextf(ctx, "client.Do error, %s,body:%s", err.Error(), body)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es query failed: %v, error: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return err
	}
	// 解析JSON响应
	var res yyguBaseResponse
	if err := json.Unmarshal(body, &res); err != nil {
		logs.ErrorContextf(ctx, "unmarshal yyguBaseResponse error: %v", err)
		return err
	}
	if res.Code != 0 {
		logs.ErrorContextf(ctx, "request yygu error code: %d, message: %s", res.Code, res.Message)
		return fmt.Errorf("request yygu error code: %d, message: %s", res.Code, res.Message)
	}
	responseBytes := res.Response
	if len(responseBytes) == 0 {
		responseBytes = []byte("null")
	}
	if err := json.Unmarshal(responseBytes, targetStruct); err != nil {
		logs.ErrorContextf(ctx, "unmarshal yyguBaseResponse error: %v", err)
		return err
	}
	return nil
}

func resolveCoreKGToken(ctx context.Context) (string, error) {

	if token := sanitizeToken(ctxutil.GetUTokenFromCtx(ctx)); token != "" {
		return token, nil
	}

	apiKeyInfo := ctxutil.GetApiAuthFromCtx(ctx)
	if apiKeyInfo == nil {
		return "", errors.New("corekg token missing: no session token or openapi auth context")
	}

	shortToken, err := getOrExchangeCoreKGToken(ctx, apiKeyInfo.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "exchange corekg token failed, user_id=%d err=%v", apiKeyInfo.UserID, err)
		return "", err
	}
	return sanitizeToken(shortToken), nil
}

func sanitizeToken(token string) string {
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")
	return strings.TrimSpace(token)
}

type yyguBaseResponse struct {
	Code      int             `json:"code"`
	Env       string          `json:"env"`
	RequestID string          `json:"request_id"`
	Message   string          `json:"message"`
	Response  json.RawMessage `json:"Response"`
}
