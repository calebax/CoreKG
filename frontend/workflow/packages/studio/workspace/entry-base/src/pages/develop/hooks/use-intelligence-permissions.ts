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

import { type IntelligenceData } from '@coze-arch/idl/intelligence_api';
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
 * Hook to manage permissions for intelligence list
 * Returns a map of intelligence ID to whether the user can manage it
 */
export const useIntelligencePermissions = (
  intelligenceList: IntelligenceData[] | undefined,
) => {
  const [permissionMap, setPermissionMap] = useState<
    Map<string, ResourceScopeItem>
  >(new Map());
  const [loading, setLoading] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);

  useEffect(() => {
    // Get current user ID
    const userId = getCurrentUserId();
    setCurrentUserId(userId);
  }, []);

  useEffect(() => {
    if (!intelligenceList || intelligenceList.length === 0) {
      setPermissionMap(new Map());
      return;
    }

    const fetchPermissions = async () => {
      setLoading(true);
      try {
        // Extract all intelligence IDs
        const resourceIds = intelligenceList
          .map(item => item.basic_info?.id)
          .filter((id): id is string => Boolean(id));

        if (resourceIds.length === 0) {
          setPermissionMap(new Map());
          return;
        }

        // Fetch permissions for all intelligence items
        // Use Agent type as intelligence type maps to Agent in resource scope
        const permissions = await fetchResourcePermissions(
          resourceIds,
          ResourceScopeType.Agent,
        );

        setPermissionMap(permissions);
      } catch (error) {
        console.error('Failed to fetch intelligence permissions:', error);
        // Keep empty map so all items default to manageable
        setPermissionMap(new Map());
      } finally {
        setLoading(false);
      }
    };

    fetchPermissions();
  }, [intelligenceList]);

  /**
   * Check if current user can manage a specific intelligence
   */
  const canManageIntelligence = (intelligenceId: string): boolean => {
    if (!currentUserId) {
      // If we don't have user ID, default to allowing management
      return true;
    }

    return hasManagePermissionForResource(
      currentUserId,
      intelligenceId,
      permissionMap,
    );
  };

  return {
    canManageIntelligence,
    permissionsLoading: loading,
  };
};
