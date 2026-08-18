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

package coresettings

import (
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// GetCoreKGUrl 获取配置中的corekg_url
func GetCoreKGUrl() (string, error) {
	corekg_url, err := settings.GetText("corekg", "corekg_url")
	if err != nil {
		logs.Errorf("failed to get corekg url settings: %w", err)
		return "", err
	}

	return corekg_url, nil
}

// GetCozeUrl 获取配置中的coze_url
func GetCozeUrl() (string, error) {
	coze_url, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.Errorf("failed to get coze url settings: %w", err)
		return "", err
	}
	return coze_url, nil
}
