import { useState } from 'react'
import { Tag, Radio, Button, Spin } from 'antd'
import { PlusOutlined, LoadingOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { uniqueArray } from '@/utils'
import useUserData from '@/hooks/data/useUserData'
import { useAdmin } from '@/utils/useAdmin'
import SelectMembersModal from '../SelectMembersModal'
import styles from './PermissionSection.module.scss'

// 自定义绿色loading图标
const GreenLoadingIcon = (
  <LoadingOutlined
    spin
    style={{
      fontSize: '16px',
      color: '#0C99FF'
    }}
  />
)

interface PermissionSectionProps {
  // 是否为编辑模式
  isEdit?: boolean
  // 管理员ID列表
  managerIds: number[]
  // 公开范围
  publicScope: 'company' | 'custom'
  // 自定义公开范围的ID列表
  scopeIds: number[]
  // 管理员变更回调
  onManagerIdsChange?: (ids: number[]) => void
  // 公开范围变更回调
  onPublicScopeChange?: (scope: 'company' | 'custom') => void
  // 自定义公开范围ID变更回调
  onScopeIdsChange?: (ids: number[]) => void
  // 表单实例
  form?: any
}

export default function PermissionSection({
  isEdit = false,
  managerIds,
  publicScope,
  scopeIds,
  onManagerIdsChange,
  onPublicScopeChange,
  onScopeIdsChange,
  form,
}: PermissionSectionProps) {
  const { t } = useTranslation(['pages', 'messages', 'common'])
  const { adminIds } = useAdmin()
  const { userList, loading: userLoading } = useUserData()

  const [showAddManagerModal, setShowAddManagerModal] = useState(false)
  const [showAddViewerModal, setShowAddViewerModal] = useState(false)

  // 检查是否可以渲染内容（确保数据已加载）
  const canRenderContent = !userLoading && userList.length > 0

  // 处理公开范围变更
  const handlePublicScopeChange = (e: any) => {
    const value = e.target.value as 'company' | 'custom'
    onPublicScopeChange?.(value)
    if (value === 'company') {
      onScopeIdsChange?.([])
    } else {
      // 切换到自定义时，确保管理员包含在scope_ids中（仅保留当前管理员/自己）
      const newScopeIds = uniqueArray(managerIds, scopeIds)
      onScopeIdsChange?.(newScopeIds)
    }
  }

  // 获取用户信息的辅助函数
  const getUserInfo = (userId: number) => {
    // 如果用户列表还没加载完成，返回加载中状态
    if (userLoading || userList.length === 0) {
      return {
        id: userId,
        name: t('common:status.loading'),
        avatar: ''
      }
    }

    // 从用户列表中查找用户信息
    const user = userList.find((u: { value: number; label: string }) => u.value === userId)
    if (user) {
      return {
        id: user.value,
        name: user.label,
        avatar: ''
      }
    }
    // 如果找不到用户，返回一个默认的用户信息对象
    return {
      id: userId,
      name: `用户${userId}`,
      avatar: ''
    }
  }

  // 渲染用户标签
  const renderUserTag = (userId: number, onRemove?: () => void, isDisabled?: boolean) => {
    const user = getUserInfo(userId)
    return (
      <Tag
        key={userId}
        closable={!!onRemove && !isDisabled}
        onClose={onRemove}
        className="mb-1"
        style={{
          backgroundColor: '#F5F5F5',
          border: '1px solid #E3E6ED',
          borderRadius: '4px',
          color: '#000000D9',
          fontSize: '12px',
          fontWeight: 400,
          lineHeight: '20px',
          padding: '2px 8px',
          margin: '0',
        }}
      >
        {user.name}
      </Tag>
    )
  }

  // 处理添加管理员
  const handleAddManagers = (newManagerIds: number[]) => {
    onManagerIdsChange?.(newManagerIds)
    setShowAddManagerModal(false)

    // 新增管理员时，同步更新查看权限
    if (publicScope === 'custom') {
      const newScopeIds = uniqueArray(adminIds, newManagerIds, scopeIds)
      onScopeIdsChange?.(newScopeIds)
    }
  }

  // 处理添加查看者
  const handleAddViewers = (newViewerIds: number[]) => {
    onScopeIdsChange?.(newViewerIds)
    setShowAddViewerModal(false)
  }

  return (
    <>
      {/* 可管理部分 - 仅编辑模式显示 */}
      {isEdit && (
        <div className="mb-6">
          <div className='flex items-center justify-between mb-3'>
            <label className='text-base font-medium text-[#3C4149]'>
              {t('app.docs.detail.manageable')}
            </label>
            <Button
              type="text"
              icon={<PlusOutlined />}
              className='text-[#0C99FF] text-base font-medium p-0 h-auto'
              onClick={() => setShowAddManagerModal(true)}
            >
              {t('app.docs.detail.add')}
            </Button>
          </div>
          <div
            className={`bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-1 pr-[6px] min-h-[100px] max-h-[200px] overflow-y-auto ${styles.scrollContainer}`}
          >
            {canRenderContent ? (
              <div className='flex flex-wrap gap-1'>
                {managerIds.map(userId =>
                  renderUserTag(
                    userId,
                    () => {
                      if (!adminIds.includes(userId)) {
                        const newManagerIds = managerIds.filter(id => id !== userId)
                        onManagerIdsChange?.(newManagerIds)
                      }
                    },
                    adminIds.includes(userId) // 管理员不能删除
                  )
                )}
              </div>
            ) : (
              <div className="flex items-center justify-center h-20">
                <Spin indicator={GreenLoadingIcon} />
              </div>
            )}
          </div>
        </div>
      )}

      {/* 仅查看部分 */}
      <div>
        <div className='flex items-center justify-between mb-3'>
          <label className='text-base font-medium text-[#3C4149]'>
            {isEdit ? t('app.docs.detail.viewOnly') : '权限设置'}
          </label>
          {publicScope === 'custom' && (
            <Button
              type="text"
              icon={<PlusOutlined />}
              className='text-[#0C99FF] text-base font-medium p-0 h-auto'
              onClick={() => setShowAddViewerModal(true)}
            >
              {t('app.docs.detail.add')}
            </Button>
          )}
        </div>
        <div className="flex gap-3 items-center">
          <Radio.Group
            value={publicScope}
            onChange={handlePublicScopeChange}
          >
            <Radio value='custom'>{t('app.docs.detail.customize')}</Radio>
            <Radio value='company'>{t('app.docs.detail.organization')}</Radio>
          </Radio.Group>
        </div>

        {/* 自定义查看范围成员展示 */}
        {publicScope === 'custom' && (
          <div className='mt-3'>
            <div
              className={`bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-3 min-h-[100px] max-h-[200px] overflow-y-auto ${styles.scrollContainer}`}
            >
              {canRenderContent ? (
                <div className='flex flex-wrap gap-2'>
                  {scopeIds.map(userId =>
                    renderUserTag(
                      userId,
                      () => {
                        if (!managerIds.includes(userId) && !adminIds.includes(userId)) {
                          const newScopeIds = scopeIds.filter(id => id !== userId)
                          onScopeIdsChange?.(newScopeIds)
                        }
                      },
                      managerIds.includes(userId) || adminIds.includes(userId) // 管理员或当前知识库管理员不能删除
                    )
                  )}
                </div>
              ) : (
                <div className="flex items-center justify-center h-20">
                  <Spin indicator={GreenLoadingIcon} />
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* 添加管理员弹窗 */}
      <SelectMembersModal
        open={showAddManagerModal}
        onClose={() => setShowAddManagerModal(false)}
        onConfirm={handleAddManagers}
        initialSelectedIds={managerIds}
        title={t('app.docs.detail.selectMembers')}
      />

      {/* 添加查看者弹窗 */}
      <SelectMembersModal
        open={showAddViewerModal}
        onClose={() => setShowAddViewerModal(false)}
        onConfirm={handleAddViewers}
        initialSelectedIds={scopeIds}
        title={t('app.docs.detail.selectMembers')}
      />
    </>
  )
}