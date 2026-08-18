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

import { useState } from 'react';

import { I18n } from '@coze-arch/i18n';

import { type ResourceScopeType } from '@/services/corekg-api';

import ResourcePermissionModal from './ResourcePermissionModal';

interface PermissionActionButtonProps {
  resourceId: string; // 改为字符串类型避免大整数精度丢失
  resourceType: ResourceScopeType;
  onSuccess?: () => void;
  disabled?: boolean;
}

/**
 * 权限操作文字链接组件
 * 直接展示在资源库列表的操作列中
 */
export const PermissionActionButton: React.FC<PermissionActionButtonProps> = ({
  resourceId,
  resourceType,
  onSuccess,
  disabled = false,
}) => {
  const [modalVisible, setModalVisible] = useState(false);

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!disabled) {
      setModalVisible(true);
    }
  };

  return (
    <>
      <span
        onClick={handleClick}
        className={`text-sm ${
          disabled
            ? 'text-[var(--coz-fg-dim)] cursor-not-allowed'
            : 'text-[var(--coz-fg-primary)] cursor-pointer hover:text-[var(--coz-fg-hglt)]'
        }`}
      >
        {I18n.t('权限' as any, {}, '权限')}
      </span>

      <ResourcePermissionModal
        open={modalVisible}
        onClose={() => setModalVisible(false)}
        resourceId={resourceId}
        resourceType={resourceType}
        onSuccess={onSuccess}
      />
    </>
  );
};

export default PermissionActionButton;
