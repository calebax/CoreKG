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
	"strconv"

	"github.com/ygpkg/yg-go/logs"
)

const (
	getResourceScopePath = "/v3/forest.GetResourceScope"
	setResourceScopePath = "/v3/forest.SetResourceScope"
)

// GetCoreKGResourceScope 获取资源权限范围
func GetCoreKGResourceScope(ctx context.Context, req *CoreKGGetResourceScopeRequest) (*CoreKGGetResourceScopeResponse, error) {
	if req == nil {
		req = &CoreKGGetResourceScopeRequest{}
	}
	payload := map[string]interface{}{
		"resource_ids":  req.ResourceIDs,
		"resource_type": strconv.FormatInt(req.ResourceType, 10),
	}
	resp := &CoreKGGetResourceScopeResponse{}
	if err := YyguRequest(ctx, getResourceScopePath, payload, resp); err != nil {
		logs.ErrorContextf(ctx, "failed to get resource scope from yygu: %v", err)
		return nil, err
	}
	return resp, nil
}

// SetCoreKGResourceScope 设置资源权限范围
func SetCoreKGResourceScope(ctx context.Context, req *CoreKGSetResourceScopeRequest) error {
	if req == nil {
		req = &CoreKGSetResourceScopeRequest{}
	}
	payload := map[string]interface{}{
		"resource_id":      req.ResourceID,
		"resource_type":    strconv.FormatInt(req.ResourceType, 10),
		"manage_scope_ids": req.ManageScopeIDs,
		"view_scope_ids":   req.ViewScopeIDs,
		"view_scope_type":  req.ViewScopeType,
	}
	resp := &CoreKGSetResourceScopeResponse{}
	if err := YyguRequest(ctx, setResourceScopePath, payload, resp); err != nil {
		logs.ErrorContextf(ctx, "failed to set resource scope from yygu: %v", err)
		return err
	}
	return nil
}

func HasScopePermissionByActionStr(operatorID int64, scopeItem *CoreKGResourceScopeItem, action string) bool {
	switch action {
	case string(ActionRead):
		return HasScopePermission(operatorID, scopeItem, ActionRead)
	case string(ActionWrite):
		return HasScopePermission(operatorID, scopeItem, ActionWrite)
	default:
		return false
	}
}

// FilterCoreKGResourceIDsByScopePermission 返回有权限操作的资源ID列表
func FilterCoreKGResourceIDsByScopePermission(ctx context.Context, operatorID int64, resourceType int64, resourceIDs []int64, action string) (map[int64]struct{}, error) {
	if len(resourceIDs) == 0 {
		return map[int64]struct{}{}, nil
	}

	resp, err := GetCoreKGResourceScope(ctx, &CoreKGGetResourceScopeRequest{
		ResourceIDs:  resourceIDs,
		ResourceType: resourceType,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.ResourceScopeList) == 0 {
		return map[int64]struct{}{}, nil
	}

	allowed := make(map[int64]struct{}, len(resp.ResourceScopeList))
	for i := range resp.ResourceScopeList {
		scopeItem := resp.ResourceScopeList[i]
		if HasScopePermissionByActionStr(operatorID, &scopeItem, action) {
			allowed[scopeItem.ResourceID] = struct{}{}
		}
	}

	return allowed, nil
}

func HasScopePermission(operatorID int64, scopeItem *CoreKGResourceScopeItem, action CorekgAction) bool {
	if scopeItem == nil {
		return false
	}
	if containsInt64(scopeItem.ManageScopeIDs, operatorID) {
		return true
	}
	if action == ActionWrite {
		return false
	}
	if scopeItem.ViewScopeType == "company" {
		return true
	}
	return containsInt64(scopeItem.ViewScopeIDs, operatorID)
}

type (
	CorekgAction string
)

const (
	ActionRead  CorekgAction = "read"
	ActionWrite CorekgAction = "write"
)

func containsInt64(values []int64, target int64) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

type CoreKGGetResourceScopeRequest struct {
	ResourceIDs  []int64 `json:"resource_ids"`
	ResourceType int64   `json:"resource_type"`
}

type CoreKGResourceScopeItem struct {
	ManageScopeIDs []int64 `json:"manage_scope_ids"`
	ResourceID     int64   `json:"resource_id"`
	ResourceType   string  `json:"resource_type"`
	ViewScopeIDs   []int64 `json:"view_scope_ids"`
	ViewScopeType  string  `json:"view_scope_type"`
}

type CoreKGGetResourceScopeResponse struct {
	ResourceScopeList []CoreKGResourceScopeItem `json:"resource_scope_list"`
}

type CoreKGSetResourceScopeRequest struct {
	ManageScopeIDs []int64 `json:"manage_scope_ids"`
	ResourceID     int64   `json:"resource_id"`
	ResourceType   int64   `json:"resource_type"`
	ViewScopeIDs   []int64 `json:"view_scope_ids"`
	ViewScopeType  string  `json:"view_scope_type"`
}

type CoreKGSetResourceScopeResponse struct{}
