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

import { useState, useEffect, useMemo, useCallback } from 'react';

import { Modal, Radio, Tag, Spin, Button, Toast } from '@coze-arch/coze-design';
import { I18n } from '@coze-arch/i18n';

import {
  CoreKGApiService,
  ResourceScopeType,
  type UinInfo,
} from '@/services/corekg-api';
import { useEmployeeList } from '@/hooks/useEmployeeList';

import AddMembersModal from './AddMembersModal';

interface ResourcePermissionModalProps {
  open: boolean;
  onClose: () => void;
  resourceId: string; // 改为字符串类型避免大整数精度丢失
  resourceType: ResourceScopeType;
  onSuccess?: () => void;
}

export const ResourcePermissionModal: React.FC<
  ResourcePermissionModalProps
> = ({ open, onClose, resourceId, resourceType, onSuccess }) => {
  // 获取员工列表（用于显示用户名）
  const { getUserName: getEmployeeName } = useEmployeeList();

  // 可管理人员ID列表
  const [managerIds, setManagerIds] = useState<number[]>([]);
  // 仅查看范围类型
  const [viewScopeType, setViewScopeType] = useState<'company' | 'user'>(
    'company',
  );
  // 仅查看人员ID列表
  const [viewerIds, setViewerIds] = useState<number[]>([]);
  // 用户信息映射（来自GetResourceScope接口）
  const [uinMap, setUinMap] = useState<Map<number, string>>(new Map());
  // 加载状态
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [saving, setSaving] = useState(false);
  // 弹窗控制
  const [showAddManagerModal, setShowAddManagerModal] = useState(false);
  const [showAddViewerModal, setShowAddViewerModal] = useState(false);

  // 获取资源权限数据
  const fetchPermission = useCallback(async () => {
    if (!resourceId) return;

    setFetching(true);
    try {
      const res = await CoreKGApiService.getResourceScope({
        resource_id_strs: [resourceId],
        resource_type: resourceType,
      });

      const scopeItem = res.resource_scope_list?.[0];
      if (scopeItem) {
        setManagerIds(scopeItem.manage_scope_ids || []);
        setViewScopeType(scopeItem.view_scope_type || 'company');
        setViewerIds(scopeItem.view_scope_ids || []);
      }

      // 构建用户信息映射
      const map = new Map<number, string>();
      res.uin_list?.forEach((user: UinInfo) => {
        map.set(user.uin, user.name);
      });
      setUinMap(map);
    } catch (error) {
      console.error('获取资源权限失败:', error);
      Toast.error(I18n.t('获取权限信息失败' as any, {}, '获取权限信息失败'));
    } finally {
      setFetching(false);
    }
  }, [resourceId, resourceType]);

  useEffect(() => {
    if (open) {
      fetchPermission();
    } else {
      // 重置状态
      setManagerIds([]);
      setViewScopeType('company');
      setViewerIds([]);
      setUinMap(new Map());
    }
  }, [open, fetchPermission]);

  // 获取用户名称（优先使用GetResourceScope返回的，否则使用ListEmployeeNickID的数据）
  const getUserName = useCallback(
    (uin: number) => {
      // 优先使用GetResourceScope接口返回的用户名
      const nameFromScope = uinMap.get(uin);
      if (nameFromScope) return nameFromScope;

      // 否则使用ListEmployeeNickID接口的数据
      return getEmployeeName(uin);
    },
    [uinMap, getEmployeeName],
  );

  // 可管理人员列表
  const managerList = useMemo(() => {
    return managerIds.map(id => ({
      id,
      name: getUserName(id),
    }));
  }, [managerIds, getUserName]);

  // 仅查看人员列表
  const viewerList = useMemo(() => {
    return viewerIds.map(id => ({
      id,
      name: getUserName(id),
    }));
  }, [viewerIds, getUserName]);

  // 移除可管理人员
  const handleRemoveManager = (id: number) => {
    // 至少保留一个可管理人员
    if (managerIds.length <= 1) return;
    setManagerIds(prev => prev.filter(i => i !== id));
  };

  // 移除仅查看人员
  const handleRemoveViewer = (id: number) => {
    // 如果在管理员列表中则不能移除
    if (managerIds.includes(id)) return;
    // 至少保留一个查看人员
    if (viewerIds.length <= 1) return;
    setViewerIds(prev => prev.filter(i => i !== id));
  };

  // 处理添加管理员
  const handleAddManagers = (newManagerIds: number[]) => {
    setManagerIds(newManagerIds);
    // 自定义模式下，自动把管理员并入自定义查看人员
    if (viewScopeType === 'user') {
      setViewerIds(prev => {
        const combined = new Set([...newManagerIds, ...prev]);
        return Array.from(combined);
      });
    }
    setShowAddManagerModal(false);
  };

  // 处理添加查看者
  const handleAddViewers = (newViewerIds: number[]) => {
    setViewerIds(newViewerIds);
    setShowAddViewerModal(false);
  };

  // 处理查看范围类型变化
  const handleViewScopeTypeChange = (type: 'company' | 'user') => {
    setViewScopeType(type);
    if (type === 'company') {
      setViewerIds([]);
    } else {
      // 切换到自定义时，确保管理员包含在查看列表中
      setViewerIds(prev => {
        const combined = new Set([...managerIds, ...prev]);
        return Array.from(combined);
      });
    }
  };

  // 保存权限
  const handleSave = async () => {
    setSaving(true);
    try {
      await CoreKGApiService.setResourceScope({
        resource_id_str: resourceId,
        resource_type: resourceType,
        manage_scope_ids: managerIds,
        view_scope_type: viewScopeType,
        view_scope_ids: viewScopeType === 'company' ? [] : viewerIds,
      });
      Toast.success(I18n.t('保存成功' as any, {}, '保存成功'));
      onSuccess?.();
      onClose();
    } catch (error) {
      console.error('设置资源权限失败:', error);
      Toast.error(I18n.t('保存失败' as any, {}, '保存失败'));
    } finally {
      setSaving(false);
    }
  };

  // 渲染用户标签
  const renderUserTag = (
    user: { id: number; name: string },
    onRemove?: () => void,
    isDisabled?: boolean,
  ) => {
    return (
      <Tag
        key={user.id}
        closable={!!onRemove && !isDisabled}
        onClose={onRemove}
        style={{
          backgroundColor: 'var(--coz-bg-tertiary)',
          border: '1px solid var(--coz-stroke-primary)',
          borderRadius: '2px',
          color: 'var(--coz-fg-primary)',
          fontSize: '12px',
          fontWeight: 400,
          lineHeight: '20px',
          padding: '0 6px',
          margin: 0,
        }}
      >
        {user.name}
      </Tag>
    );
  };

  return (
    <>
      <Modal
        visible={open}
        onCancel={onClose}
        footer={null}
        width={400}
        centered
        destroyOnClose
        closable={true}
        maskClosable={false}
        closeOnEsc={false}
        title={I18n.t('权限' as any, {}, '权限')}
        bodyStyle={{ padding: '24px' }}
        getPopupContainer={() => document.body}
      >
        <div className="flex flex-col gap-4">
          {fetching ? (
            <div className="flex items-center justify-center py-12">
              <Spin />
            </div>
          ) : (
            <>
              {/* 权限管理标题 */}
              {/* <div className="text-sm font-medium text-[var(--coz-fg-primary)]">
                {I18n.t('权限管理' as any, {}, '权限管理')}
              </div> */}

              {/* 可管理部分 */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-base text-[var(--coz-fg-primary)]">
                    {I18n.t('可管理' as any, {}, '可管理')}
                  </label>
                  <button
                    type="button"
                    className="flex items-center gap-1 text-sm cursor-pointer"
                    style={{
                      color: '#0C99FF',
                      border: 'none',
                      background: 'none',
                      padding: 0,
                    }}
                    onClick={e => {
                      e.stopPropagation();
                      setShowAddManagerModal(true);
                    }}
                  >
                    <span>+</span>
                    {I18n.t('添加' as any, {}, '添加')}
                  </button>
                </div>
                <div className="bg-[var(--coz-bg-secondary)] border border-solid border-[var(--coz-stroke-primary)] rounded p-2 min-h-[64px] max-h-[120px] overflow-y-auto">
                  {loading ? (
                    <div className="flex items-center justify-center h-16">
                      <Spin />
                    </div>
                  ) : (
                    <div className="flex flex-wrap gap-1">
                      {managerList.map(user =>
                        renderUserTag(
                          user,
                          () => handleRemoveManager(user.id),
                          managerIds.length <= 1,
                        ),
                      )}
                    </div>
                  )}
                </div>
              </div>

              {/* 仅查看部分 */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-base text-[var(--coz-fg-primary)]">
                    {I18n.t('仅查看' as any, {}, '仅查看')}
                  </label>
                  {viewScopeType === 'user' && (
                    <button
                      type="button"
                      className="flex items-center gap-1 text-sm cursor-pointer"
                      style={{
                        color: '#0C99FF',
                        border: 'none',
                        background: 'none',
                        padding: 0,
                      }}
                      onClick={e => {
                        e.stopPropagation();
                        setShowAddViewerModal(true);
                      }}
                    >
                      <span>+</span>
                      {I18n.t('添加' as any, {}, '添加')}
                    </button>
                  )}
                </div>

                {/* 范围选择 */}
                <div className="flex gap-6 items-center mb-2">
                  <Radio.Group
                    value={viewScopeType}
                    onChange={e =>
                      handleViewScopeTypeChange(
                        e.target.value as 'company' | 'user',
                      )
                    }
                  >
                    <Radio value="user">
                      {I18n.t('自定义' as any, {}, '自定义')}
                    </Radio>
                    <Radio value="company">
                      {I18n.t('组织' as any, {}, '组织')}
                    </Radio>
                  </Radio.Group>
                </div>

                {/* 自定义查看范围成员展示 */}
                {viewScopeType === 'user' && (
                  <div className="bg-[var(--coz-bg-secondary)] border border-solid border-[var(--coz-stroke-primary)] rounded p-2 min-h-[64px] max-h-[120px] overflow-y-auto">
                    {loading ? (
                      <div className="flex items-center justify-center h-16">
                        <Spin />
                      </div>
                    ) : (
                      <div className="flex flex-wrap gap-1">
                        {viewerList.map(user =>
                          renderUserTag(
                            user,
                            () => handleRemoveViewer(user.id),
                            managerIds.includes(user.id) ||
                              viewerIds.length <= 1,
                          ),
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </>
          )}

          {/* 底部按钮 */}
          <div className="flex justify-end gap-2 pt-4 mt-2">
            <Button
              onClick={e => {
                e.stopPropagation();
                onClose();
              }}
              disabled={saving}
              style={{
                padding: '9px 24.5px',
                border: 'none',
                borderRadius: '6px',
                backgroundColor: '#F5F5F5',
                color: '#0C1F17',
              }}
            >
              {I18n.t('Cancel')}
            </Button>
            <Button
              type="primary"
              loading={saving}
              disabled={fetching}
              onClick={e => {
                e.stopPropagation();
                handleSave();
              }}
              style={{
                padding: '9px 24.5px',
                border: 'none',
                borderRadius: '6px',
                backgroundColor:
                  saving || fetching ? 'rgba(0, 0, 0, 0.06)' : '#0C99FF',
                color: saving || fetching ? 'rgba(0, 0, 0, 0.25)' : '#ffffff',
              }}
            >
              {I18n.t('保存' as any, {}, '保存')}
            </Button>
          </div>
        </div>
      </Modal>

      {/* 添加管理员弹窗 */}
      <AddMembersModal
        open={showAddManagerModal}
        onClose={() => setShowAddManagerModal(false)}
        onConfirm={handleAddManagers}
        initialSelectedIds={managerIds}
        lockedIds={[]}
        minSelected={1}
      />

      {/* 添加查看者弹窗 */}
      <AddMembersModal
        open={showAddViewerModal}
        onClose={() => setShowAddViewerModal(false)}
        onConfirm={handleAddViewers}
        initialSelectedIds={viewerIds}
        lockedIds={managerIds}
        minSelected={1}
      />
    </>
  );
};

export default ResourcePermissionModal;
