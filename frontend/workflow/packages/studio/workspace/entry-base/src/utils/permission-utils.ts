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

import CoreKGApiService, {
  type ResourceScopeItem,
  ResourceScopeType,
} from '@/services/corekg-api';

/**
 * 获取当前用户ID
 * 优先从userStore获取，如果获取失败则从localStorage获取
 */
export const getCurrentUserId = (): number | null => {
  try {
    // 尝试从 localStorage 获取 coze_user_info
    const userInfoStr = localStorage.getItem('coze_user_info');
    if (userInfoStr) {
      const userInfo = JSON.parse(userInfoStr);
      // uinId 是用户ID字段
      if (userInfo.uinId) {
        const userId = Number(userInfo.uinId);
        return userId;
      }
    }
  } catch (error) {
    console.error('Failed to get current user id:', error);
  }
  return null;
};

/**
 * 检查用户对某个资源是否有管理权限
 * @param userId 当前用户ID
 * @param scopeItem 资源权限信息
 * @returns true表示有管理权限，false表示仅查看权限
 */
export const hasManagePermission = (
  userId: number,
  scopeItem: ResourceScopeItem | undefined,
): boolean => {
  if (!scopeItem) {
    // 如果没有权限信息，默认有管理权限（保持现有行为）
    return true;
  }

  // 检查用户是否在管理权限列表中
  return scopeItem.manage_scope_ids.includes(userId);
};

/**
 * 批量获取资源权限
 * @param resourceIds 资源ID数组
 * @param resourceType 资源类型
 * @returns 资源ID到权限信息的映射
 */
export const fetchResourcePermissions = async (
  resourceIds: string[],
  resourceType: ResourceScopeType,
): Promise<Map<string, ResourceScopeItem>> => {
  if (resourceIds.length === 0) {
    return new Map();
  }

  try {
    const response = await CoreKGApiService.getResourceScope({
      resource_id_strs: resourceIds,
      resource_type: resourceType,
    });

    const permissionMap = new Map<string, ResourceScopeItem>();
    response.resource_scope_list.forEach(item => {
      // 使用 resource_id_str
      // 这样可以避免精度丢失
      const resourceIdStr = item.resource_id_str;
      if (resourceIdStr) {
        permissionMap.set(resourceIdStr, item);
      }
    });

    return permissionMap;
  } catch (error) {
    console.error('Failed to fetch resource permissions:', error);
    // 返回空Map，这样所有资源都会默认有管理权限
    return new Map();
  }
};

/**
 * 检查用户对某个资源是否有管理权限（通过资源ID）
 * @param userId 当前用户ID
 * @param resourceId 资源ID
 * @param permissionMap 权限映射表
 * @returns true表示有管理权限，false表示仅查看权限
 */
export const hasManagePermissionForResource = (
  userId: number,
  resourceId: string,
  permissionMap: Map<string, ResourceScopeItem>,
): boolean => {
  const scopeItem = permissionMap.get(resourceId);
  return hasManagePermission(userId, scopeItem);
};
