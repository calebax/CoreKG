import { useState, useEffect } from 'react'
import {
  Form,
  FormInstance,
  Select,
  Tag,
  Button,
  Radio,
  Spin,
  message,
} from 'antd'
import { PlusOutlined, LoadingOutlined } from '@ant-design/icons'
import { Agent } from 'Agent'
import { useTranslation } from 'react-i18next'
import { uniqueArray } from '@/utils'
import useUserData from '@/hooks/data/useUserData'
import useLocalStore from '@/stores/local'
import { useAdmin } from '@/utils/useAdmin'
import { AgentEditValue } from '../..'
import AddMembersModal from './AddMembersModal'
import { usePersonnelData } from 'Personnel'

// 自定义绿色loading图标
const GreenLoadingIcon = (
  <LoadingOutlined
    spin
    style={{
      fontSize: '16px',
      color: '#0C99FF',
    }}
  />
)

interface PermissionManagementProps {
  form: FormInstance<AgentEditValue>
  className?: string
  style?: React.CSSProperties
}

export default function PermissionManagement({
  form,
  className,
  style,
}: PermissionManagementProps) {
  const { t } = useTranslation(['pages', 'messages', 'common'])
  const { adminIds } = useAdmin()
  const { uinId } = useLocalStore((state) => state.userInfo)
  const { userList, loading: userLoading } = useUserData()
  const type = Form.useWatch<Agent['type']>('type', { preserve: true })
  const manager_ids: number[] = Form.useWatch('manager_ids', form) || []
  // 修复类型定义
  const public_scope: 'company' | 'custom' | 'private' | 'public' =
    Form.useWatch('public_scope', form) || 'custom'
  const scope_ids: number[] = Form.useWatch('scope_ids', form) || []

  const [showAddManagerModal, setShowAddManagerModal] = useState(false)
  const [showAddViewerModal, setShowAddViewerModal] = useState(false)

  // 检查是否可以渲染内容（确保数据已加载）
  const canRenderContent = !userLoading && userList.length > 0

  // 处理添加管理员
  const handleAddManagers = (newManagerIds: number[]) => {
    const currentManagerIds = form.getFieldValue('manager_ids') || []
    const updatedManagerIds = uniqueArray(newManagerIds)
    form.setFieldValue('manager_ids', updatedManagerIds)

    // 自动同步到scope_ids
    const currentScopeIds = form.getFieldValue('scope_ids') || []
    const newScopeIds = uniqueArray(
      currentScopeIds,
      updatedManagerIds,
    )
    form.setFieldValue('scope_ids', newScopeIds)

    setShowAddManagerModal(false)
    message.success(t('messages:addSuccess', { defaultValue: '修改成功' }))
  }

  // 处理添加查看者
  const handleAddViewers = (newViewerIds: number[]) => {
    const currentManagerIds = form.getFieldValue('manager_ids') || []
    const updatedScopeIds = uniqueArray(currentManagerIds, newViewerIds)
    form.setFieldValue('scope_ids', updatedScopeIds)
    setShowAddViewerModal(false)
    message.success(t('messages:addSuccess', { defaultValue: '修改成功' }))
  }

  // 处理公开类型变更
  const handlePublicScopeChange = (e: any) => {
    const value = e.target.value
    form.setFieldValue('public_scope', value)

    if (value === 'company') {
      form.setFieldValue('scope_type', 'company')
      form.setFieldValue('scope_ids', [])
    } else if (value === 'custom') {
      form.setFieldValue('scope_type', 'user')
      // 切换到自定义时，确保管理员包含在scope_ids中
      const currentManagerIds = form.getFieldValue('manager_ids') || []
      const currentScopeIds = form.getFieldValue('scope_ids') || []
      const newScopeIds = uniqueArray(
        adminIds,
        currentManagerIds,
        currentScopeIds,
      )
      form.setFieldValue('scope_ids', newScopeIds)
    }
  }

  // 移除管理员
  const handleRemoveManager = (userId: number) => {
    const currentManagerIds = form.getFieldValue('manager_ids') || []
    // 保护：至少保留1个可管理成员
    const newManagerIds =
      currentManagerIds.length <= 1
        ? currentManagerIds
        : currentManagerIds.filter((id: number) => id !== userId)
    form.setFieldValue('manager_ids', newManagerIds)

    // 如果该用户不在scope_ids中作为查看者，也需要移除
    const currentScopeIds = form.getFieldValue('scope_ids') || []
    const newScopeIds =
      currentScopeIds.length <= 1
        ? currentScopeIds
        : currentScopeIds.filter((id: number) => id !== userId)
    form.setFieldValue('scope_ids', newScopeIds)
  }

  // 移除查看者
  const handleRemoveViewer = (userId: number) => {
    const currentManagerIds = form.getFieldValue('manager_ids') || []
    const currentScopeIds = form.getFieldValue('scope_ids') || []

    // 允许移除任意查看者
    // 保护：至少保留1个查看成员
    const newScopeIds =
      currentScopeIds.length <= 1
        ? currentScopeIds
        : currentScopeIds.filter((id: number) => id !== userId)
    form.setFieldValue('scope_ids', newScopeIds)
  }

  // 获取用户信息的辅助函数
  const getUserInfo = (userId: number) => {
    if (userLoading || userList.length === 0) {
      return {
        id: userId,
        name: t('common:status.loading'),
        avatar: '',
      }
    }

    const user = userList.find(
      (u: { value: number; label: string }) => u.value === userId,
    )
    if (user) {
      return {
        id: user.value,
        name: user.label,
        avatar: '',
      }
    }
    //有些ID可能来自employee_id
    try {
      const { data } = usePersonnelData.getState()
      const employee = (data as any)?.employee?.find((emp: any) => Number(emp.id) === Number(userId))
      if (employee) {
        return {
          id: userId,
          name: employee.name,
          avatar: '',
        }
      }
    } catch {}
    return {
      id: userId,
      name: `用户${userId}`,
      avatar: '',
    }
  }

  // 渲染用户标签
  const renderUserTag = (
    userId: number,
    onRemove?: () => void,
    isDisabled?: boolean,
  ) => {
    const user = getUserInfo(userId)
    return (
      <Tag
        key={userId}
        closable={!!onRemove && !isDisabled}
        onClose={onRemove}
        className='mb-1'
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

  return (
    <div>
      {/* 可管理部分 */}
      <div className='mb-6'>
        <div className='flex items-center justify-between mb-3'>
          <label className='text-base font-medium text-[#3C4149]'>可管理</label>
          {type === 'workflow' ? null : (
            <Button
              type='text'
              size='small'
              className='flex items-center gap-1 text-[#0C99FF] text-base font-medium px-[2px]'
              icon={<PlusOutlined />}
              onClick={() => setShowAddManagerModal(true)}
            >
              添加
            </Button>
          )}
        </div>
        <div className='bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-3 min-h-[100px]'>
          {canRenderContent ? (
            <div className='flex flex-wrap gap-2'>
                {manager_ids.map((userId) =>
                renderUserTag(
                  userId,
                  () => handleRemoveManager(userId),
                    adminIds.includes(userId) || type === 'workflow' || manager_ids.length <= 1,
                ),
              )}
            </div>
          ) : (
            <div className='flex items-center justify-center h-20'>
              <Spin indicator={GreenLoadingIcon} />
            </div>
          )}
        </div>
      </div>

      {/* 仅查看部分 */}
      <div>
        <div className='flex items-center justify-between mb-3'>
          <label className='text-base font-medium text-[#3C4149]'>仅查看</label>
          {public_scope === 'custom' && (
            <Button
              type='text'
              size='small'
              className='flex items-center gap-1 text-[#0C99FF] text-base font-medium px-[2px]'
              icon={<PlusOutlined />}
              onClick={() => setShowAddViewerModal(true)}
            >
              添加
            </Button>
          )}
        </div>

        <div className='mb-3'>
          <Radio.Group value={public_scope} onChange={handlePublicScopeChange}>
            <Radio value='custom'>自定义</Radio>
            <Radio value='company'>组织</Radio>
          </Radio.Group>
        </div>

        {/* 自定义查看范围成员展示 */}
        {public_scope === 'custom' && (
          <div className='bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-3 min-h-[100px]'>
            {canRenderContent ? (
              <div className='flex flex-wrap gap-2'>
                {scope_ids.map((userId) =>
                  renderUserTag(
                    userId,
                    () => handleRemoveViewer(userId),
                    manager_ids.includes(userId) || adminIds.includes(userId) || scope_ids.length <= 1,
                  ),
                )}
              </div>
            ) : (
              <div className='flex items-center justify-center h-20'>
                <Spin indicator={GreenLoadingIcon} />
              </div>
            )}
          </div>
        )}
      </div>

      {/* 添加管理员弹窗 */}
      <AddMembersModal
        open={showAddManagerModal}
        onClose={() => setShowAddManagerModal(false)}
        onConfirm={handleAddManagers}
        initialSelectedIds={manager_ids}
        lockedIds={[]}
        minSelected={1}
      />

      {/* 添加查看者弹窗 */}
      <AddMembersModal
        open={showAddViewerModal}
        onClose={() => setShowAddViewerModal(false)}
        onConfirm={handleAddViewers}
        initialSelectedIds={scope_ids}
        lockedIds={Array.from(new Set([...manager_ids, ...adminIds]))}
        minSelected={1}
      />
    </div>
  )
}
