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

import { useEffect, useState } from 'react';

import { type ResourceInfo } from '@coze-arch/idl/plugin_develop';
import { ResType } from '@coze-arch/idl/resource';
import {
  type ResourceScopeItem,
  ResourceScopeType,
} from '@/services/corekg-api';
import {
  getCurrentUserId,
  fetchResourcePermissions,
  hasManagePermissionForResource,
} from '@/utils/permission-utils';

/**
 * Map resource type to resource scope type
 */
const getResourceScopeType = (resType: ResType): ResourceScopeType | null => {
  switch (resType) {
    case ResType.Plugin:
      return ResourceScopeType.Plugin;
    case ResType.Workflow:
      return ResourceScopeType.Workflow;
    case ResType.Prompt:
      return ResourceScopeType.Prompt;
    default:
      return null;
  }
};

/**
 * Hook to manage permissions for library resources
 * Returns a function to check if user can manage a resource
 */
export const useLibraryResourcePermissions = (
  resourceList: ResourceInfo[] | undefined,
) => {
  const [permissionMaps, setPermissionMaps] = useState<
    Map<ResourceScopeType, Map<string, ResourceScopeItem>>
  >(new Map());
  const [loading, setLoading] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);

  useEffect(() => {
    // Get current user ID
    const userId = getCurrentUserId();
    setCurrentUserId(userId);
  }, []);

  useEffect(() => {
    if (!resourceList || resourceList.length === 0) {
      setPermissionMaps(new Map());
      return;
    }

    const fetchPermissions = async () => {
      setLoading(true);
      try {
        // Group resources by type
        const resourcesByType = new Map<ResType, string[]>();

        resourceList.forEach(resource => {
          if (!resource.res_id || resource.res_type === undefined) {
            return;
          }

          const scopeType = getResourceScopeType(resource.res_type);
          if (!scopeType) {
            return;
          }

          if (!resourcesByType.has(resource.res_type)) {
            resourcesByType.set(resource.res_type, []);
          }
          resourcesByType.get(resource.res_type)!.push(resource.res_id);
        });

        // Fetch permissions for each resource type
        const newPermissionMaps = new Map<
          ResourceScopeType,
          Map<string, ResourceScopeItem>
        >();

        await Promise.all(
          Array.from(resourcesByType.entries()).map(async ([resType, ids]) => {
            const scopeType = getResourceScopeType(resType);
            if (!scopeType) {
              return;
            }

            const permissions = await fetchResourcePermissions(ids, scopeType);
            newPermissionMaps.set(scopeType, permissions);
          }),
        );

        setPermissionMaps(newPermissionMaps);
      } catch (error) {
        console.error('Failed to fetch library resource permissions:', error);
        // Keep empty maps so all items default to manageable
        setPermissionMaps(new Map());
      } finally {
        setLoading(false);
      }
    };

    fetchPermissions();
  }, [resourceList]);

  /**
   * Check if current user can manage a specific resource
   */
  const canManageResource = (resource: ResourceInfo): boolean => {
    if (!currentUserId || !resource.res_id || resource.res_type === undefined) {
      // If we don't have user ID or resource info, default to allowing management
      return true;
    }

    const scopeType = getResourceScopeType(resource.res_type);
    if (!scopeType) {
      // Unknown resource type, default to allowing management
      return true;
    }

    const permissionMap = permissionMaps.get(scopeType);
    if (!permissionMap) {
      // No permission data yet, default to allowing management
      return true;
    }

    return hasManagePermissionForResource(
      currentUserId,
      resource.res_id,
      permissionMap,
    );
  };

  return {
    canManageResource,
    permissionsLoading: loading,
  };
};
