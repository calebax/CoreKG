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
import { useEffect, useState } from 'react';

import {
  ActionKey,
  PluginType,
  ResType,
  type ResourceInfo,
} from '@coze-arch/idl/plugin_develop';
import { I18n } from '@coze-arch/i18n';
import { PluginDevelopApi, KnowledgeApi } from '@coze-arch/bot-api';
import { useBotCodeEditOutPlugin } from '@coze-agent-ide/bot-plugin/hook';
import { CreateFormPluginModal } from '@coze-agent-ide/bot-plugin/component';
import { IconCozPlugin } from '@coze-arch/coze-design/icons';
import { Menu, Tag, Toast, Table } from '@coze-arch/coze-design';

import { PermissionActionButton } from '@/components/resource-permission';
import { ResourceScopeType } from '@/services/corekg-api';

import { BaseLibraryItem } from '../../components/base-library-item';
import PluginDefaultIcon from '../../assets/plugin_default_icon.png';
import { type UseEntityConfigHook } from './types';

const { TableAction } = Table;

export const usePluginConfig: UseEntityConfigHook = ({
  spaceId,
  reloadList,
  getCommonActions,
  onOptimisticDelete,
  existingResourceNames = [],
}) => {
  const [showFormPluginModel, setShowFormPluginModel] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const url = new URL(window.location.href);
    if (url.searchParams.get('create') !== '1') {
      return;
    }
    url.searchParams.delete('create');
    window.history.replaceState(
      null,
      '',
      `${url.pathname}${url.search}${url.hash}`,
    );
    setShowFormPluginModel(true);
  }, [spaceId]);

  const { modal: editPluginCodeModal, open } = useBotCodeEditOutPlugin({
    modalProps: {
      onSuccess: reloadList,
    },
  });

  return {
    modals: (
      <>
        <CreateFormPluginModal
          isCreate={true}
          visible={showFormPluginModel}
          existingNames={existingResourceNames}
          onSuccess={pluginID => {
            navigate(`/space/${spaceId}/plugin/${pluginID}`);
            reloadList();
          }}
          onCancel={() => {
            setShowFormPluginModel(false);
          }}
        />
        {editPluginCodeModal}
      </>
    ),
    config: {
      typeFilter: {
        label: I18n.t('library_resource_type_plugin'),
        value: ResType.Plugin,
      },
      renderCreateMenu: () => (
        <Menu.Item
          data-testid="workspace.library.header.create.plugin"
          icon={<IconCozPlugin />}
          onClick={() => {
            setShowFormPluginModel(true);
          }}
        >
          {I18n.t('library_resource_type_plugin')}
        </Menu.Item>
      ),
      target: [ResType.Plugin],
      onItemClick: (item: ResourceInfo) => {
        if (
          item.res_type === ResType.Plugin &&
          item.res_sub_type === 2 //Plugin：1-Http; 2-App; 6-Local；
        ) {
          const disable = !item.actions?.find(
            action => action.key === ActionKey.Delete,
          )?.enable;
          open(item.res_id || '', disable);
        } else {
          navigate(`/space/${spaceId}/plugin/${item.res_id}`);
        }
      },
      renderItem: item => (
        <BaseLibraryItem
          resourceInfo={item}
          defaultIcon={PluginDefaultIcon}
          tag={
            item.res_type === ResType.Plugin &&
            item.res_sub_type === PluginType.LOCAL ? (
              <Tag
                data-testid="workspace.library.item.tag"
                color="cyan"
                size="mini"
                className="flex-shrink-0 flex-grow-0"
              >
                {I18n.t('local_plugin_label')}
              </Tag>
            ) : null
          }
        />
      ),
      renderActions: (
        item: ResourceInfo,
        reloadList: () => void,
        onOptimisticDelete?: (resourceId: string) => void,
      ) => {
        const canDelete = item.actions?.find(
          action => action.key === ActionKey.Delete,
        )?.enable;

        const handleDelete = async () => {
          if (canDelete) {
            onOptimisticDelete?.(item.res_id || '');
            await PluginDevelopApi.DelPlugin({ plugin_id: item.res_id });
            reloadList();
            Toast.success(I18n.t('Delete_success'));
          }
        };

        return (
          <div className="flex items-center gap-3">
            {/* 权限按钮 */}
            <PermissionActionButton
              resourceId={item.res_id || ''}
              resourceType={ResourceScopeType.Plugin}
              onSuccess={reloadList}
            />

            {/* 删除按钮 */}
            <span
              onClick={e => {
                e.stopPropagation();
                handleDelete();
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
