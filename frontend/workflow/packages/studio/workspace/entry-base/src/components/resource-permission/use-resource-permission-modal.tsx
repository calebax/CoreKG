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

import { useState, useCallback } from 'react';

import { type ResourceScopeType } from '@/services/corekg-api';

import ResourcePermissionModal from './ResourcePermissionModal';

interface UseResourcePermissionModalParams {
  onSuccess?: () => void;
}

/**
 * 资源权限弹窗 Hook
 * 用于在资源库列表页中快速集成权限设置功能
 */
export const useResourcePermissionModal = (
  params?: UseResourcePermissionModalParams,
) => {
  const [visible, setVisible] = useState(false);
  const [resourceId, setResourceId] = useState<string>(''); // 改为字符串类型避免大整数精度丢失
  const [resourceType, setResourceType] = useState<ResourceScopeType>(
    '17' as ResourceScopeType,
  );

  // 打开权限弹窗
  const openPermissionModal = useCallback(
    (id: string, type: ResourceScopeType) => {
      setResourceId(id);
      setResourceType(type);
      setVisible(true);
    },
    [],
  );

  // 关闭权限弹窗
  const closePermissionModal = useCallback(() => {
    setVisible(false);
    setResourceId('');
  }, []);

  // 权限弹窗组件
  const permissionModal = (
    <ResourcePermissionModal
      open={visible}
      onClose={closePermissionModal}
      resourceId={resourceId}
      resourceType={resourceType}
      onSuccess={params?.onSuccess}
    />
  );

  return {
    visible,
    openPermissionModal,
    closePermissionModal,
    permissionModal,
  };
};

export default useResourcePermissionModal;
