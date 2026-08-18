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

/* eslint @coze-arch/max-line-per-function: ["error", {"max": 500}] */
/* eslint-disable complexity */
import React, { useMemo, useState, type ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';

import classNames from 'classnames';
import {
  type IntelligenceBasicInfo,
  type IntelligenceData,
  IntelligenceStatus,
  IntelligenceType,
} from '@coze-arch/idl/intelligence_api';
import { I18n } from '@coze-arch/i18n';
import { IconCozMore } from '@coze-arch/coze-design/icons';
import { Avatar, IconButton, Menu, Tooltip } from '@coze-arch/coze-design';
import { formatDate, getFormatDateType } from '@coze-arch/bot-utils';
import { useSpaceStore } from '@coze-arch/bot-studio-store';

import { useResourcePermissionModal } from '@/components/resource-permission';
import { ResourceScopeType } from '@/services/corekg-api';

import Name from './name';
import { type AgentCopySuccessCallback, MenuCopyBot } from './menu-actions';
import Description from './description';
import { CopyProcessMask } from './copy-process-mask';

export interface BotCardProps {
  intelligenceInfo: IntelligenceData;
  timePrefixType?: 'recentOpen' | 'publish' | 'edit';
  /**
   * Returning true interrupts the default jump behavior
   */
  onClick?: (() => true) | (() => void);
  onDelete?: (param: {
    name: string;
    id: string;
    type: IntelligenceType;
  }) => void;
  onCopyProject?: (basicInfo: IntelligenceBasicInfo) => void;
  onCopyAgent?: AgentCopySuccessCallback;
  onUpdateIntelligenceInfo: (info: IntelligenceData) => void;
  onRetryCopy: (basicInfo: IntelligenceBasicInfo) => void;
  onCancelCopyAfterFailed: (basicInfo: IntelligenceBasicInfo) => void;
  onPermissionUpdate?: () => void; // 权限更新成功后的回调
  extraMenu?: ReactNode;
  headerExtra?: ReactNode;
  statusExtra?: ReactNode;
  actionsMenuVisible?: boolean;
}

// eslint-disable-next-line max-lines-per-function
export const BotCard: React.FC<BotCardProps> = ({
  intelligenceInfo,
  timePrefixType,
  onClick,
  onDelete,
  onCopyProject,
  onCopyAgent,
  onUpdateIntelligenceInfo,
  onCancelCopyAfterFailed,
  onRetryCopy,
  onPermissionUpdate,
  extraMenu,
  actionsMenuVisible = true,
  headerExtra,
  statusExtra,
}) => {
  const navigate = useNavigate();
  const { permissionModal, openPermissionModal } = useResourcePermissionModal({
    onSuccess: () => {
      // 权限更新成功后刷新列表
      onPermissionUpdate?.();
    },
  });

  const {
    basic_info,
    type,
    permission_info: { can_delete } = {},
    publish_info: { publish_time, has_published } = {},
    other_info: { recently_open_time } = {},
  } = intelligenceInfo;

  const { id, name, icon_url, space_id, description, update_time, status } =
    basic_info ?? {};

  const hideOperation = useSpaceStore(store => store.space.hide_operation);

  if (!id || !space_id) {
    // The id and space id are necessary for the bot card. Here are the constraints on the ts type
    throw Error('No botID or no spaceID which are necessary');
  }

  const isBanned = status === IntelligenceStatus.Banned;
  const isAgent = type === IntelligenceType.Bot;
  const isProject = type === IntelligenceType.Project;

  const timePrefix = useMemo(() => {
    switch (timePrefixType) {
      case 'recentOpen':
        return I18n.t('develop_list_rank_tag_opened');
      case 'publish':
        return I18n.t('bot_list_rank_tag_published');
      case 'edit':
        return I18n.t('bot_list_rank_tag_edited');
      default:
    }
  }, [timePrefixType]);

  const time = useMemo(() => {
    let timestamp: string | undefined;

    switch (timePrefixType) {
      case 'recentOpen':
        timestamp = recently_open_time;
        break;
      case 'publish':
        timestamp = publish_time;
        break;
      case 'edit':
        timestamp = update_time;
        break;
      default:
    }

    return formatDate(Number(timestamp), getFormatDateType(Number(timestamp)));
  }, [timePrefixType, publish_time, update_time, recently_open_time]);

  // Whether to display the card layering operation button
  const [showActions, setShowActions] = useState(false);
  // Whether to display the menu menu, there are other components actively calling here, which need to be controlled
  const [showMenu, setShowMenu] = useState(false);

  return (
    <>
      <div
        className={classNames([
          'flex-grow h-[180px] min-w-[300px]',
          'rounded-[8px]',
          'relative',
          'overflow-hidden transition duration-150 ease-out',
          'shadow-[0_2px_8px_0_rgba(0,0,0,0.08)] hover:shadow-[0_4px_16px_0_rgba(0,0,0,0.12)]',
          'coz-mg-card',
        ])}
      >
        <div
          className="h-full w-full cursor-pointer flex flex-col gap-6 px-[24px] py-[16px]"
          onClick={() => {
            if (onClick?.()) {
              return;
            }
            if (isBanned) {
              return;
            }
            // 点击卡片根据发布状态决定跳转页面
            if (isAgent) {
              // 如果已发布，进入问答页；否则进入配置页
              if (has_published) {
                navigate(`/space/${space_id}/bot/${id}/chat`);
              } else {
                navigate(`/space/${space_id}/bot/${id}`);
              }
              return;
            }
            if (isProject) {
              navigate(`/space/${space_id}/project-ide/${id}`);
              return;
            }
          }}
          onMouseEnter={() => {
            // 只有在有权限时才显示操作菜单
            if (!hideOperation && actionsMenuVisible) {
              setShowActions(true);
            }
          }}
          onMouseLeave={() => {
            setShowActions(false);
            // 鼠标离开卡片时，重置菜单状态，避免再次进入时菜单直接显示
            setShowMenu(false);
          }}
          data-testid="bot-list-page.bot-card"
        >
          {/* Display migration failure status icon */}
          {statusExtra}

          {/* 上半部分：图标和基本信息 */}
          <div className="flex gap-[12px] mb-[16px]">
            {/* 左侧图标 */}
            <Avatar
              className="w-[60px] h-[60px] rounded-[8px] flex-shrink-0"
              shape="square"
              src={icon_url}
            />

            {/* 右侧信息：名称和时间 */}
            <div className="flex-1 flex flex-col justify-between min-w-0 pr-[32px]">
              {/* 名称 - 限制最大宽度避免和三点菜单重叠 */}
              <Name name={name} />

              {/* 最近编辑时间 */}
              <div className="text-[12px] coz-fg-secondary leading-[18px] mt-[4px]">
                {timePrefix} {time}
              </div>
            </div>
          </div>

          {/* 下半部分：描述 */}
          <Description description={description} />

          {/* Actions 右上角三点菜单 */}
          {!hideOperation && actionsMenuVisible ? (
            <div
              className="absolute top-[12px] right-[12px]"
              onClick={e => {
                // 阻止点击事件冒泡
                e.stopPropagation();
              }}
            >
              {showActions ? (
                <Menu
                  keepDOM
                  className="w-fit mt-4px mb-4px"
                  position="bottomRight"
                  trigger="custom"
                  visible={showMenu}
                  onVisibleChange={setShowMenu}
                  render={
                    <Menu.SubMenu mode="menu">
                      {/* 配置 - 进入原来的调试页面 */}
                      <Menu.Item
                        onClick={() => {
                          setShowActions(false);
                          if (isBanned) {
                            return;
                          }
                          if (isAgent) {
                            navigate(`/space/${space_id}/bot/${id}`);
                            return;
                          }
                          if (isProject) {
                            navigate(`/space/${space_id}/project-ide/${id}`);
                            return;
                          }
                        }}
                      >
                        配置
                      </Menu.Item>
                      {/* 创建副本 */}
                      {isAgent ? (
                        <MenuCopyBot
                          id={id}
                          spaceID={space_id}
                          disabled={isBanned}
                          onCopySuccess={onCopyAgent}
                          onClose={() => setShowActions(false)}
                        />
                      ) : null}
                      {isProject ? (
                        <Tooltip content={I18n.t('coze_copy_to_tips_1')}>
                          <Menu.Item
                            onClick={() => {
                              if (!basic_info) {
                                return;
                              }
                              onCopyProject?.(basic_info);
                            }}
                            data-testid="bot-card.copy"
                          >
                            {I18n.t('project_ide_create_duplicate')}
                          </Menu.Item>
                        </Tooltip>
                      ) : null}
                      {/* 权限 */}
                      <Menu.Item
                        onClick={() => {
                          if (id) {
                            openPermissionModal(id, ResourceScopeType.Agent);
                          }
                          setShowActions(false);
                        }}
                      >
                        权限
                      </Menu.Item>
                      {/* 删除 */}
                      <Tooltip
                        position="left"
                        trigger={can_delete ? 'custom' : 'hover'}
                        content={I18n.t('project_delete_permission_tooltips')}
                      >
                        <Menu.Item
                          type="danger"
                          disabled={!can_delete}
                          onClick={() => {
                            if (!name || !type) {
                              return;
                            }
                            onDelete?.({ name, id, type });
                          }}
                        >
                          <span>{I18n.t('Delete')}</span>
                        </Menu.Item>
                      </Tooltip>
                    </Menu.SubMenu>
                  }
                >
                  <IconButton
                    className="rotate-90"
                    data-testid="bot-card.icon-more-button"
                    color="primary"
                    size="default"
                    icon={<IconCozMore />}
                    onClick={() => setShowMenu(true)}
                  />
                </Menu>
              ) : null}
            </div>
          ) : null}
        </div>
        {basic_info ? (
          <CopyProcessMask
            intelligenceBasicInfo={basic_info}
            onRetry={changedStatus => {
              onRetryCopy({
                ...basic_info,
                status: changedStatus,
              });
            }}
            onCancelCopyAfterFailed={changedStatus => {
              onCancelCopyAfterFailed({
                ...basic_info,
                status: changedStatus,
              });
            }}
          />
        ) : null}
        {/* 权限弹窗 */}
        {permissionModal}
      </div>
    </>
  );
};
