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

import { useNavigate } from 'react-router-dom';
import { useRef } from 'react';

import { useRequest } from 'ahooks';
import {
  ActionKey,
  ResType,
  type ResourceInfo,
} from '@coze-arch/idl/plugin_develop';
import { type IntelligenceData } from '@coze-arch/idl/intelligence_api';
import { I18n } from '@coze-arch/i18n';
import { IconCozLightbulb } from '@coze-arch/coze-design/icons';
import { Table, Menu, Toast, Button, Space } from '@coze-arch/coze-design';
import { EVENT_NAMES, sendTeaEvent } from '@coze-arch/bot-tea';
import { useFlags } from '@coze-arch/bot-flags';
import { PlaygroundApi } from '@coze-arch/bot-api';
import { useModal as useSelectIntelligenceModal } from '@coze-common/biz-components/select-intelligence-modal';
import { usePromptConfiguratorModal } from '@coze-common/prompt-kit-adapter/create-prompt';

import { PermissionActionButton } from '@/components/resource-permission';
import { ResourceScopeType } from '@/services/corekg-api';

import { type UseEntityConfigHook } from './types';

const { TableAction } = Table;

export const usePromptConfig: UseEntityConfigHook = ({
  spaceId,
  isPersonalSpace = true,
  reloadList,
  getCommonActions,
  onOptimisticDelete,
  existingResourceNames = [],
}) => {
  const navigate = useNavigate();
  const [FLAGS] = useFlags();
  const recordRef = useRef<ResourceInfo | null>(null);

  const { open: openSelectIntelligenceModal, node: selectIntelligenceModal } =
    useSelectIntelligenceModal({
      spaceId,
      onSelect: (intelligence: IntelligenceData) => {
        const targetId = intelligence.basic_info?.id;
        const diffPromptResourceId = recordRef.current?.res_id;
        navigate(`/space/${spaceId}/bot/${targetId}`, {
          replace: true,
          state: {
            mode: 'diff',
            diffPromptResourceId,
            targetId,
          },
        });
        sendTeaEvent(EVENT_NAMES.compare_mode_front, {
          bot_id: targetId,
          compare_type: 'prompts',
          from: 'prompt_resource',
          source: 'bot_detail_page',
          action: 'start',
        });
      },
    });

  const { open: openCreatePrompt, node: promptConfiguratorModal } =
    usePromptConfiguratorModal({
      spaceId,
      source: 'resource_library',
      existingNames: existingResourceNames,
      // Support soon, so stay tuned.
      enableDiff: FLAGS['bot.studio.prompt_diff'],
      onUpdateSuccess: reloadList,
      onDiff: ({ libraryId }) => {
        recordRef.current = {
          res_id: libraryId,
        };
        openSelectIntelligenceModal();
      },
    });

  // delete
  const { run: delPrompt } = useRequest(
    (promptId: string) =>
      PlaygroundApi.DeletePromptResource({
        prompt_resource_id: promptId,
      }),
    {
      manual: true,
      onSuccess: () => {
        Toast.success(I18n.t('Delete_success'));
        // 删除成功后不刷新列表，避免后端删除延迟导致的闪烁问题
        // 乐观删除已经从UI移除了该项，用户切换筛选条件或刷新页面时会自然更新列表
      },
    },
  );

  return {
    modals: (
      <>
        {selectIntelligenceModal}
        {promptConfiguratorModal}
      </>
    ),
    config: {
      typeFilter: {
        label: I18n.t('library_resource_type_prompt'),
        value: ResType.Prompt,
      },
      renderCreateMenu: () => (
        <Menu.Item
          data-testid="workspace.library.header.create.prompt"
          icon={<IconCozLightbulb />}
          onClick={() => {
            sendTeaEvent(EVENT_NAMES.widget_create_click, {
              source: 'menu_bar',
              workspace_type: isPersonalSpace
                ? 'personal_workspace'
                : 'team_workspace',
            });
            openCreatePrompt({
              mode: 'create',
            });
          }}
        >
          {I18n.t('creat_new_prompt_prompt')}
        </Menu.Item>
      ),
      target: [ResType.Prompt],
      onItemClick: (record: ResourceInfo) => {
        recordRef.current = record;
        const canEdit = record.actions?.find(
          action => action.key === ActionKey.Edit,
        )?.enable;
        openCreatePrompt({
          mode: 'info',
          canEdit,
          editId: record.res_id || '',
        });
      },
      renderActions: (
        libraryResource: ResourceInfo,
        reloadList: () => void,
        onOptimisticDelete?: (resourceId: string) => void,
      ) => {
        const canEdit = libraryResource.actions?.find(
          action => action.key === ActionKey.Edit,
        )?.enable;
        const canDelete = libraryResource.actions?.find(
          action => action.key === ActionKey.Delete,
        )?.enable;

        return (
          <div className="flex items-center gap-3">
            {/* 编辑按钮 */}
            <span
              onClick={e => {
                e.stopPropagation();
                if (canEdit) {
                  openCreatePrompt({
                    mode: 'edit',
                    editId: libraryResource.res_id || '',
                  });
                }
              }}
              className={`text-sm ${
                canEdit
                  ? 'text-[var(--coz-fg-primary)] cursor-pointer hover:text-[var(--coz-fg-hglt)]'
                  : 'text-[var(--coz-fg-dim)] cursor-not-allowed'
              }`}
            >
              {I18n.t('Edit')}
            </span>

            {/* 权限按钮 */}
            <PermissionActionButton
              resourceId={libraryResource.res_id || ''}
              resourceType={ResourceScopeType.Prompt}
              onSuccess={reloadList}
            />

            {/* 删除按钮 */}
            <span
              onClick={e => {
                e.stopPropagation();
                if (canDelete) {
                  onOptimisticDelete?.(libraryResource.res_id || '');
                  delPrompt(libraryResource.res_id || '');
                }
              }}
              className={`text-sm ${
                canDelete
                  ? 'text-[var(--coz-fg-primary)] cursor-pointer hover:text-[var(--coz-fg-hglt)]'
                  : 'text-[var(--coz-fg-dim)] cursor-not-allowed'
              }`}
            >
              {I18n.t('Delete')}
            </span>
          </div>
        );
      },
    },
  };
};
