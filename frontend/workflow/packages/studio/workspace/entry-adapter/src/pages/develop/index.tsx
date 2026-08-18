/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* eslint-disable max-lines-per-function */
/* eslint @coze-arch/max-line-per-function: ["error", {"max": 500}] */
/* eslint-disable complexity */
import { type FC, useEffect, useMemo, useState } from 'react';

import { useNavigate, useParams } from 'react-router-dom';

import classNames from 'classnames';
import {
  highlightFilterStyle,
  WorkspaceEmpty,
  DevelopCustomPublishStatus,
  isPublishStatus,
  isRecentOpen,
  isSearchScopeEnum,
  getPublishRequestParam,
  getTypeRequestParams,
  isEqualDefaultFilterParams,
  isFilterHighlight,
  CREATOR_FILTER_OPTIONS,
  FILTER_PARAMS_DEFAULT,
  STATUS_FILTER_OPTIONS,
  TYPE_FILTER_OPTIONS,
  ORDER_BY_FILTER_OPTIONS,
  BotCard,
  Content,
  Header,
  HeaderActions,
  HeaderTitle,
  Layout,
  SubHeader,
  SubHeaderFilters,
  SubHeaderSearch,
  useIntelligenceList,
  useIntelligenceActions,
  useCachedQueryParams,
  useGlobalEventListeners,
  type DevelopProps,
  useProjectCopyPolling,
  useCardActions,
  useIntelligencePermissions,
} from '@coze-studio/workspace-base/develop';
import { useSpaceStore } from '@coze-foundation/space-store-adapter';
import {
  IntelligenceType,
  search,
  SearchScope,
} from '@coze-arch/idl/intelligence_api';
import { I18n, type I18nKeysNoOptionsType } from '@coze-arch/i18n';
import {
  IconCozLoading,
  IconCozPlus,
  IconCozKnowledgeFill,
} from '@coze-arch/coze-design/icons';
import {
  Button,
  IconButton,
  Input,
  Search,
  Select,
  Spin,
} from '@coze-arch/coze-design';
import { EVENT_NAMES, sendTeaEvent } from '@coze-arch/bot-tea';
import { SpaceType } from '@coze-arch/bot-api/developer_api';

import { useEmployeeList } from '../../../../entry-base/src/hooks/useEmployeeList';

import bgImg from './images/bg.png';
import SearchIcon from './images/search.svg?react';
import SearchIconActive from './images/search-copy.svg?react';
import LibraryIcon from './images/library.svg?react';

const WelcomeBanner: FC = () => {
  return (
    <div className="rounded-2xl mx-12 mt-[10px] relative">
      <img src={bgImg} className="w-full" alt="Welcome Banner" />
      <div className="text-[#2A4C95] z-10 absolute ml-14 top-[46%] -translate-y-1/2 flex flex-col gap-4">
        <p className="text-[40px] font-semibold">欢迎使用智能体</p>
        <p className="text-[16px] font-medium">
          可配置的任务型助手，理解上下文并持续学习
        </p>
      </div>
    </div>
  );
};

export const Develop: FC<DevelopProps> = ({ spaceId }) => {
  const navigate = useNavigate();
  const isPersonal = useSpaceStore(
    state => state.space.space_type === SpaceType.Personal,
  );

  // 搜索框展开状态
  const [searchExpanded, setSearchExpanded] = useState(false);
  // 搜索输入框的本地值（用于回车搜索）
  const [searchInputValue, setSearchInputValue] = useState('');

  // 加载员工列表（用于权限弹窗显示用户名）
  useEmployeeList();

  // Keyword Search & Filtering
  const [filterParams, setFilterParams, debouncedSetSearchValue] =
    useCachedQueryParams();

  const {
    isIntelligenceTypeFilterHighlight,
    isOwnerFilterHighlight,
    isPublishAndOpenFilterHighlight,
  } = isFilterHighlight(filterParams);

  const {
    listResp: { loading, data, loadingMore, mutate, noMore, reload },
    containerRef,
  } = useIntelligenceList({
    params: {
      spaceId,
      searchValue: filterParams.searchValue,
      types: getTypeRequestParams({
        type: filterParams.searchType,
      }),
      hasPublished: getPublishRequestParam(filterParams.isPublish),
      recentlyOpen: filterParams.recentlyOpen,
      searchScope: filterParams.searchScope,
      orderBy: filterParams.orderBy,
    },
  });

  // 获取权限信息
  const { canManageIntelligence } = useIntelligencePermissions(data?.list);

  useGlobalEventListeners({ reload, spaceId });

  useEffect(() => {
    setFilterParams(prev => ({
      ...prev,
      searchValue: '',
    }));
  }, [spaceId]);

  /**
   * report tea event
   */
  useEffect(() => {
    sendTeaEvent(EVENT_NAMES.view_bot, { tab: 'my_bots' });
  }, []);

  useProjectCopyPolling({
    listData: data?.list,
    spaceId,
    mutate,
  });

  const existingIntelligenceNames = useMemo(
    () =>
      data?.list
        ?.map(item => item.basic_info?.name)
        .filter((name): name is string => Boolean(name)) ?? [],
    [data?.list],
  );

  const { contextHolder: cardActionsContextHolder, actions: cardActions } =
    useCardActions({
      isPersonalSpace: isPersonal,
      mutate,
    });

  /**
   * Create project
   */
  const { contextHolder, actions } = useIntelligenceActions({
    spaceId,
    mutateList: mutate,
    reloadList: reload,
    existingNames: existingIntelligenceNames,
  });

  return (
    <>
      {contextHolder}
      {cardActionsContextHolder}
      {/* 整个页面容器 - 固定高度，不滚动 */}
      <div className="w-full h-full flex flex-col overflow-hidden bg-white">
        {/* 顶部header - 固定，不滚动 */}
        <div className="w-full h-[56px] flex items-center px-6 py-0 shrink-0">
          <div className="flex items-center gap-2 text-sm">
            <span className="coz-fg-primary font-medium cursor-default">
              智能体
            </span>
          </div>
        </div>

        {/* Banner区域 - 固定，不滚动 */}
        <div>
          <WelcomeBanner />
        </div>

        {/* 可滚动区域 - Banner下面的筛选和列表 */}
        <div
          className="flex-1 overflow-auto px-[100px] pb-12 pt-8"
          ref={containerRef}
        >
          <Spin spinning={loading} wrapperClassName="w-full">
            {/* 列表容器 - 使用grid布局 */}
            <div className="w-full flex justify-center">
              <div
                className="w-full grid gap-x-10 gap-y-8 justify-center"
                style={{ gridTemplateColumns: 'repeat(auto-fill, 300px)' }}
              >
                {/* 筛选和操作区域 - 放在grid第一行，与列表对齐 */}
                <div
                  className="flex items-center justify-between whitespace-nowrap"
                  style={{ gridColumn: '1/-1' }}
                >
                  {/* 左侧筛选区域 */}
                  <div className="flex gap-4 items-center">
                    {/* 排序方式筛选 */}
                    <div className="flex gap-[6px] items-center">
                      <div className="font-[500] text-[14px] coz-fg-secondary">
                        排序方式
                      </div>
                      <Select
                        className="min-w-[128px]"
                        value={filterParams.orderBy}
                        onChange={val => {
                          setFilterParams(prev => ({
                            ...prev,
                            orderBy: val as search.OrderBy,
                          }));
                          // Tea event tracking
                          sendTeaEvent(EVENT_NAMES.workspace_action_front, {
                            space_id: spaceId,
                            space_type: isPersonal ? 'personal' : 'teamspace',
                            tab_name: 'develop',
                            action: 'filter',
                            filter_type: 'order_by',
                            filter_name: I18n.t(
                              ORDER_BY_FILTER_OPTIONS.find(
                                opt => opt.value === val,
                              )?.labelI18NKey as I18nKeysNoOptionsType,
                            ),
                          });
                        }}
                      >
                        {ORDER_BY_FILTER_OPTIONS.map(opt => (
                          <Select.Option key={opt.value} value={opt.value}>
                            {I18n.t(opt.labelI18NKey)}
                          </Select.Option>
                        ))}
                      </Select>
                    </div>

                    {/* 智能体类型筛选 - 暂时隐藏 */}
                    {/* <div className="flex gap-[6px] items-center">
                      <div className="font-[500] text-[14px] coz-fg-secondary">
                        智能体类型
                      </div>
                      <Select
                        className="min-w-[128px]"
                        style={
                          isIntelligenceTypeFilterHighlight
                            ? highlightFilterStyle
                            : {}
                        }
                        value={filterParams.searchType}
                        onChange={val => {
                          setFilterParams(prev => ({
                            ...prev,
                            searchType:
                              val as (typeof TYPE_FILTER_OPTIONS)[number]['value'],
                          }));

                          // Tea event tracking
                          sendTeaEvent(EVENT_NAMES.workspace_action_front, {
                            space_id: spaceId,
                            space_type: isPersonal ? 'personal' : 'teamspace',
                            tab_name: 'develop',
                            action: 'filter',
                            filter_type: 'types',
                            filter_name: I18n.t(
                              TYPE_FILTER_OPTIONS.find(opt => opt.value === val)
                                ?.labelI18NKey as I18nKeysNoOptionsType,
                            ),
                          });
                        }}
                      >
                        {TYPE_FILTER_OPTIONS.map(opt => (
                          <Select.Option key={opt.value} value={opt.value}>
                            {I18n.t(opt.labelI18NKey)}
                          </Select.Option>
                        ))}
                      </Select>
                    </div> */}

                    {!isPersonal ? (
                      <div className="flex gap-[6px] items-center">
                        <div className="font-[500] text-[14px] coz-fg-secondary">
                          创建者
                        </div>
                        <Select
                          className="min-w-[128px]"
                          style={
                            isOwnerFilterHighlight ? highlightFilterStyle : {}
                          }
                          value={filterParams.searchScope}
                          onChange={val => {
                            if (!isSearchScopeEnum(val)) {
                              return;
                            }
                            setFilterParams(p => {
                              if (
                                val === SearchScope.CreateByMe &&
                                p.recentlyOpen
                              ) {
                                return {
                                  ...p,
                                  recentlyOpen: false,
                                  isPublish: DevelopCustomPublishStatus.All,
                                  searchScope: val,
                                };
                              }
                              return {
                                ...p,
                                searchScope: val,
                              };
                            });
                            // Tea event tracking
                            sendTeaEvent(EVENT_NAMES.workspace_action_front, {
                              space_id: spaceId,
                              space_type: isPersonal ? 'personal' : 'teamspace',
                              tab_name: 'develop',
                              action: 'filter',
                              filter_type: 'creators',
                              filter_name: I18n.t(
                                CREATOR_FILTER_OPTIONS.find(
                                  opt => opt.value === val,
                                )?.labelI18NKey as I18nKeysNoOptionsType,
                              ),
                            });
                          }}
                        >
                          {CREATOR_FILTER_OPTIONS.map(opt => (
                            <Select.Option key={opt.value} value={opt.value}>
                              {I18n.t(opt.labelI18NKey)}
                            </Select.Option>
                          ))}
                        </Select>
                      </div>
                    ) : null}

                    {/* 状态筛选 */}
                    <div className="flex gap-[6px] items-center">
                      <div className="font-[500] text-[14px] coz-fg-secondary">
                        状态
                      </div>
                      <Select
                        className="min-w-[128px]"
                        style={
                          isPublishAndOpenFilterHighlight
                            ? highlightFilterStyle
                            : {}
                        }
                        value={
                          filterParams.recentlyOpen
                            ? 'recentOpened'
                            : filterParams.isPublish
                        }
                        onChange={val => {
                          setFilterParams(p => ({
                            ...p,
                            searchScope: SearchScope.All,
                            recentlyOpen: isRecentOpen(val),
                            isPublish: isPublishStatus(val)
                              ? val
                              : DevelopCustomPublishStatus.All,
                          }));
                          // Tea event tracking
                          sendTeaEvent(EVENT_NAMES.workspace_action_front, {
                            space_id: spaceId,
                            space_type: isPersonal ? 'personal' : 'teamspace',
                            tab_name: 'develop',
                            action: 'filter',
                            filter_type: 'status',
                            filter_name: I18n.t(
                              STATUS_FILTER_OPTIONS.find(
                                opt => opt.value === val,
                              )?.labelI18NKey as I18nKeysNoOptionsType,
                            ),
                          });
                        }}
                      >
                        {STATUS_FILTER_OPTIONS.map(opt => (
                          <Select.Option key={opt.value} value={opt.value}>
                            {I18n.t(opt.labelI18NKey)}
                          </Select.Option>
                        ))}
                      </Select>
                    </div>
                  </div>

                  {/* 右侧操作区域 */}
                  <div className="ml-auto flex items-center gap-3">
                    {/* 可展开搜索框 */}
                    {searchExpanded ? (
                      <div className="relative">
                        <style>
                          {`
                            .custom-search-input input::placeholder {
                              color: #0C99FF !important;
                            }
                          `}
                        </style>
                        <Input
                          autoFocus
                          disabled={filterParams.recentlyOpen}
                          className="w-[200px] h-[30px] transition-all duration-300 custom-search-input"
                          style={{
                            border: '1px solid #0C99FF',
                            borderRadius: '6px',
                            backgroundColor: 'transparent',
                            ...(filterParams.searchValue
                              ? highlightFilterStyle
                              : {}),
                          }}
                          placeholder="搜索"
                          prefix={<SearchIconActive className="w-4 h-4" />}
                          value={searchInputValue}
                          onChange={val => {
                            setSearchInputValue(val);
                          }}
                          onBlur={() => {
                            if (!searchInputValue) {
                              setSearchExpanded(false);
                            }
                          }}
                          onKeyDown={e => {
                            if (e.key === 'Enter') {
                              e.preventDefault();
                              setFilterParams(prev => ({
                                ...prev,
                                searchValue: searchInputValue,
                              }));
                            }
                          }}
                        />
                      </div>
                    ) : (
                      <Button
                        icon={<SearchIcon />}
                        onClick={() => {
                          setSearchExpanded(true);
                          setSearchInputValue(filterParams.searchValue || '');
                        }}
                        style={{
                          height: '30px',
                          padding: '0 12px',
                          border: '1px solid #0C99FF',
                          color: '#0C99FF',
                          backgroundColor: 'transparent',
                          borderRadius: '6px',
                          boxShadow: 'none',
                          fontWeight: '400',
                        }}
                        onMouseEnter={e => {
                          e.currentTarget.style.borderColor = '#0C99FF';
                          e.currentTarget.style.color = '#40a9ff';
                        }}
                        onMouseLeave={e => {
                          e.currentTarget.style.borderColor = '#0C99FF';
                          e.currentTarget.style.color = '#0C99FF';
                        }}
                        onMouseDown={e => {
                          e.currentTarget.style.color = '#096dd9';
                        }}
                        onMouseUp={e => {
                          e.currentTarget.style.color = '#40a9ff';
                        }}
                      >
                        搜索
                      </Button>
                    )}
                    {/* 资源库按钮 */}
                    <Button
                      icon={<LibraryIcon />}
                      onClick={() => {
                        navigate(`/space/${spaceId}/library`);
                      }}
                      style={{
                        height: '30px',
                        padding: '0 12px',
                        border: '1px solid #0C99FF',
                        color: '#0C99FF',
                        backgroundColor: 'transparent',
                        borderRadius: '6px',
                        boxShadow: 'none',
                        fontWeight: '400',
                      }}
                      onMouseEnter={e => {
                        e.currentTarget.style.borderColor = '#0C99FF';
                        e.currentTarget.style.color = '#40a9ff';
                      }}
                      onMouseLeave={e => {
                        e.currentTarget.style.borderColor = '#0C99FF';
                        e.currentTarget.style.color = '#0C99FF';
                      }}
                      onMouseDown={e => {
                        e.currentTarget.style.color = '#096dd9';
                      }}
                      onMouseUp={e => {
                        e.currentTarget.style.color = '#40a9ff';
                      }}
                    >
                      {I18n.t('navigation_workspace_library')}
                    </Button>
                    {/* 创建按钮 */}
                    <Button
                      icon={<IconCozPlus />}
                      onClick={actions.createIntelligence}
                      style={{
                        height: '30px',
                        padding: '0 12px',
                        border: '1px solid #0C99FF',
                        color: '#0C99FF',
                        backgroundColor: 'transparent',
                        borderRadius: '6px',
                        boxShadow: 'none',
                        fontWeight: '400',
                      }}
                      onMouseEnter={e => {
                        e.currentTarget.style.borderColor = '#0C99FF';
                        e.currentTarget.style.color = '#40a9ff';
                      }}
                      onMouseLeave={e => {
                        e.currentTarget.style.borderColor = '#0C99FF';
                        e.currentTarget.style.color = '#0C99FF';
                      }}
                      onMouseDown={e => {
                        e.currentTarget.style.color = '#096dd9';
                      }}
                      onMouseUp={e => {
                        e.currentTarget.style.color = '#40a9ff';
                      }}
                    >
                      {I18n.t('workspace_create')}
                    </Button>
                  </div>
                </div>

                {/* 卡片列表 */}
                {data?.list.length
                  ? data.list.map((project, index) => (
                      <BotCard
                        key={`${project.basic_info?.id}-${index}`}
                        intelligenceInfo={project}
                        onRetryCopy={cardActions.onRetryCopy}
                        onCancelCopyAfterFailed={
                          cardActions.onCancelCopyAfterFailed
                        }
                        onClick={() => {
                          cardActions.onClick(project);
                        }}
                        onUpdateIntelligenceInfo={cardActions.onUpdate}
                        onPermissionUpdate={reload}
                        onDelete={({ name, id, type }) => {
                          if (type === IntelligenceType.Bot) {
                            actions.deleteIntelligence({
                              name,
                              spaceId,
                              agentId: id,
                            });
                            return;
                          }
                          if (type === IntelligenceType.Project) {
                            actions.deleteIntelligence({
                              name,
                              projectId: id,
                            });
                            return;
                          }
                        }}
                        onCopyAgent={cardActions.onCopyAgent}
                        onCopyProject={params => {
                          cardActions.onCopyProject({
                            initialValue: {
                              project_id: params.id ?? '',
                              to_space_id: spaceId,
                              name: params.name ?? '',
                              description: params.description,
                              icon_uri: [
                                {
                                  uid: params.icon_uri,
                                  url: params.icon_url ?? '',
                                },
                              ],
                            },
                          });
                        }}
                        timePrefixType={
                          filterParams.recentlyOpen
                            ? 'recentOpen'
                            : filterParams.isPublish
                            ? 'publish'
                            : 'edit'
                        }
                        actionsMenuVisible={
                          project.basic_info?.id
                            ? canManageIntelligence(project.basic_info.id)
                            : true
                        }
                      />
                    ))
                  : null}

                {/* 空状态 */}
                {!data?.list?.length && !loading ? (
                  <div
                    className="h-60 flex flex-col items-center justify-center coz-fg-secondary"
                    style={{ gridColumn: '1/-1' }}
                  >
                    <WorkspaceEmpty
                      onClear={() => {
                        setFilterParams(FILTER_PARAMS_DEFAULT);
                      }}
                      hasFilter={
                        !isEqualDefaultFilterParams({
                          filterParams,
                        })
                      }
                    />
                  </div>
                ) : null}

                {/* Show loading at the bottom. */}
                {data?.list.length && loadingMore ? (
                  <div
                    className="flex items-center justify-center w-full h-[38px] my-[20px] coz-fg-secondary text-[12px]"
                    style={{ gridColumn: '1/-1' }}
                  >
                    <IconButton
                      icon={<IconCozLoading />}
                      loading
                      color="secondary"
                    />
                    <div>{I18n.t('Loading')}...</div>
                  </div>
                ) : null}
                {/* Show a placeholder when there is no more data */}
                {noMore && data?.list.length ? (
                  <div
                    className="h-[38px] my-[20px]"
                    style={{ gridColumn: '1/-1' }}
                  ></div>
                ) : null}
              </div>
            </div>
          </Spin>
        </div>
      </div>
    </>
  );
};
