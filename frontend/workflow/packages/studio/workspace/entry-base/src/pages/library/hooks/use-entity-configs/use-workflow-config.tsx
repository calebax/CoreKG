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

import { useWorkflowResourceAction } from '@coze-workflow/components';
import { useUserInfo } from '@coze-foundation/account-adapter';
import {
  ResType,
  WorkflowMode,
  type ResourceInfo,
} from '@coze-arch/idl/plugin_develop';
import { I18n } from '@coze-arch/i18n';
import { IconCozChat, IconCozWorkflow } from '@coze-arch/coze-design/icons';
import { Menu, Tag, Table } from '@coze-arch/coze-design';

import { PermissionActionButton } from '@/components/resource-permission';
import { ResourceScopeType } from '@/services/corekg-api';
import { createWorkflowDuplicateNameValidator } from '@/utils/duplicate-name-validator';

import { BaseLibraryItem } from '../../components/base-library-item';
import WorkflowDefaultIcon from '../../assets/workflow_default_icon.png';
import ImageFlowDefaultIcon from '../../assets/image_flow_default_icon.png';
import { type UseEntityConfigHook } from './types';

const { TableAction } = Table;

const defaultIconMap: { [key in ResType]?: string } = {
  [ResType.Workflow]: WorkflowDefaultIcon,
  [ResType.Imageflow]: ImageFlowDefaultIcon,
};

export const useWorkflowConfig: UseEntityConfigHook = ({
  spaceId,
  reloadList,
  getCommonActions,
  onOptimisticDelete,
  existingResourceNames = [],
}) => {
  const userInfo = useUserInfo();
  const nameValidators = existingResourceNames.length
    ? [createWorkflowDuplicateNameValidator(existingResourceNames)]
    : undefined;
  const {
    workflowResourceModals,
    handleWorkflowResourceClick,
    renderWorkflowResourceActions,
    openCreateModal,
  } = useWorkflowResourceAction({
    spaceId,
    userId: userInfo?.user_id_str,
    refreshPage: reloadList,
    getCommonActions,
    onOptimisticDelete,
    nameValidators,
  });

  return {
    modals: workflowResourceModals,
    config: {
      typeFilter: {
        label: I18n.t('library_resource_type_workflow'),
        value: ResType.Workflow,
      },
      parseParams: params => {
        // After the workflow image stream is merged, the selected workflow needs to also pull out the image stream
        if (params?.res_type_filter?.[0] === ResType.Workflow) {
          return {
            ...params,
            is_get_imageflow: true,
          };
        }
        return params;
      },
      renderCreateMenu: () => (
        <>
          <Menu.Item
            data-testid="workspace.library.header.create.workflow"
            icon={<IconCozWorkflow />}
            onClick={() => {
              openCreateModal(WorkflowMode.Workflow);
            }}
          >
            {I18n.t('library_resource_type_workflow')}
          </Menu.Item>
          {/* <Menu.Item
            data-testid="workspace.library.header.create.chatflow"
            icon={<IconCozChat />}
            onClick={() => {
              openCreateModal(WorkflowMode.ChatFlow);
            }}
          >
            {I18n.t('wf_chatflow_76')}
          </Menu.Item> */}
        </>
      ),
      target: [ResType.Workflow, ResType.Imageflow],
      onItemClick: handleWorkflowResourceClick,
      renderItem: item => (
        <BaseLibraryItem
          resourceInfo={item}
          defaultIcon={
            item.res_type !== undefined
              ? defaultIconMap[item.res_type]
              : undefined
          }
          tag={
            item.collaboration_enable === true ? (
              <Tag
                data-testid="workspace.library.item.tag"
                color="brand"
                size="mini"
                className="flex-shrink-0 flex-grow-0"
              >
                {I18n.t('library_filter_tags_collaboration')}
              </Tag>
            ) : null
          }
        />
      ),
      renderResType: item => {
        if (
          item.res_type === ResType.Workflow &&
          item.res_sub_type === WorkflowMode.ChatFlow
        ) {
          return I18n.t('wf_chatflow_76');
        }

        if (item.res_sub_type === 11) {
          return I18n.t('library_resource_type_workflow');
        }
        if (item.res_sub_type === 12) {
          return I18n.t('library_resource_type_workflow');
        }
        return I18n.t('library_resource_type_workflow');
      },
      // 包装工作流的 renderActions，添加权限按钮
      renderActions: (
        item: ResourceInfo,
        reloadListFn: () => void,
        onOptimisticDeleteFn?: (resourceId: string) => void,
      ) => {
        // 渲染原有的工作流操作按钮
        const originalActions = renderWorkflowResourceActions(
          item,
          reloadListFn,
          onOptimisticDeleteFn,
        );

        return (
          <div className="flex items-center gap-3">
            {/* 权限按钮 */}
            <PermissionActionButton
              resourceId={item.res_id || ''}
              resourceType={ResourceScopeType.Workflow}
              onSuccess={reloadListFn}
            />
            {/* 原有的操作按钮 */}
            {originalActions}
          </div>
        );
      },
    },
  };
};
