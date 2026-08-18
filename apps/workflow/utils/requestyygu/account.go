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
	"context"
	"time"

	"github.com/ygpkg/yg-go/logs"
)

const (
	DetailPersonalCenterPath = "/v2/account.DetailPersonalCenter"
)

// DetailPersonalCenter 个人中心详情
func DetailPersonalCenter(ctx context.Context) (*DetailPersonalCenterResponse, error) {
	resp := &DetailPersonalCenterResponse{}
	err := YyguRequest(ctx, DetailPersonalCenterPath, map[string]interface{}{}, resp)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to get personal center details: %v", err)
		return nil, err
	}
	return resp, nil
}

type DetailPersonalCenterResponse struct {
	UserInfo UserInfo `json:"user_info"`
}

type UserInfo struct {
	GithubID         uint   `json:"github_id,omitempty"`
	WorkWechatUserID string `json:"work_wechat_user_id,omitempty"`
	WechatUnionID    string `json:"wechat_union_id,omitempty"`
	WechatWebOpenID  string `json:"wechat_web_open_id,omitempty"`

	Identify  string    `json:"identify,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Name      string    `json:"name,omitempty"`
	RealName  string    `json:"real_name,omitempty"`
	IDcard    string    `json:"id_card,omitempty"`
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Uin       uint      `json:"uin,omitempty"`

	HasPassword int `json:"has_password"`
}
