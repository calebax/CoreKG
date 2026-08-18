import { useState, useEffect } from 'react'
import { Drawer, Select, Tag, message, Button, Spin, Radio } from 'antd'
import { CloseOutlined, PlusOutlined, LoadingOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { uniqueArray } from '@/utils'
import { updateForestWithPerm } from '@/api/knowledge'
import useLocalStore from '@/stores/local'
import useUserData from '@/hooks/data/useUserData'
import { usePersonnelData } from 'Personnel'
import { useAdmin } from '@/utils/useAdmin'
import AddMembersModal from './AddMembersModal'
import styles from './PermissionSidebar.module.scss'
import scrollStyles from '@/styles/scroll/styles.module.scss'

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

interface PermissionSidebarProps {
  open: boolean
  onClose: () => void
  knowledgeBaseId: number
  initialData: {
    manager_ids: number[]
    public_scope: 'company' | 'custom'
    scope_ids: number[]
    name?: string
    description?: string
  }
  onSuccess?: () => void
}

export default function PermissionSidebar({
  open,
  onClose,
  knowledgeBaseId,
  initialData,
  onSuccess,
}: PermissionSidebarProps) {
  const { t } = useTranslation(['pages', 'messages', 'common'])
  const { adminIds } = useAdmin()
  const { uinId } = useLocalStore((state) => state.userInfo)
  const { userList, loading: userLoading } = useUserData() // 获取加载状态

  const [managerIds, setManagerIds] = useState<number[]>([])
  const [publicScope, setPublicScope] = useState<'company' | 'custom'>('custom')
  const [scopeIds, setScopeIds] = useState<number[]>([])
  const [loading, setLoading] = useState(false)
  const [showAddManagerModal, setShowAddManagerModal] = useState(false)
  const [showAddViewerModal, setShowAddViewerModal] = useState(false)

  // 临时状态，用于取消时重置
  const [tempManagerIds, setTempManagerIds] = useState<number[]>([])
  const [tempPublicScope, setTempPublicScope] = useState<'company' | 'custom'>('custom')
  const [tempScopeIds, setTempScopeIds] = useState<number[]>([])

  // 检查是否可以渲染内容（确保数据已加载）
  const canRenderContent = !userLoading && userList.length > 0

  // 初始化数据
  useEffect(() => {
    if (open && initialData) {
      setManagerIds(initialData.manager_ids || [])
      setPublicScope(initialData.public_scope || 'custom')
      // 自定义范围下，默认确保管理员也在可查看列表中
      const initScopeIds =
        (initialData.public_scope || 'custom') === 'custom'
          ? uniqueArray(initialData.manager_ids || [], initialData.scope_ids || [])
          : (initialData.scope_ids || [])
      setScopeIds(initScopeIds)

      // 设置临时状态
      setTempManagerIds(initialData.manager_ids || [])
      setTempPublicScope(initialData.public_scope || 'custom')
      setTempScopeIds(initScopeIds)
    }
  }, [open, initialData])

  // 检查是否有变更
  const hasChanges =
    JSON.stringify(tempManagerIds) !== JSON.stringify(managerIds) ||
    tempPublicScope !== publicScope ||
    JSON.stringify(tempScopeIds) !== JSON.stringify(scopeIds)

  // 保存权限更改
  const handleSavePermissions = async () => {
    // 如果没有更改，直接返回，不执行任何操作
    if (!hasChanges) {
      return
    }

    setLoading(true)
    try {
      await updateForestWithPerm({
        id: knowledgeBaseId,
        name: initialData.name || '',
        description: initialData.description || '',
        manager_ids: tempManagerIds,
        public_scope: tempPublicScope,
        scope_ids: tempPublicScope === 'company' ? [] : tempScopeIds,
        data_source_type: 'standard',
        data_source_subtype: 'standard',
      })

      // 更新实际状态
      setManagerIds(tempManagerIds)
      setPublicScope(tempPublicScope)
      setScopeIds(tempScopeIds)

      message.success(t('messages:permissionUpdateSuccess'))
      onSuccess?.()
    } catch (error) {
      console.error('更新权限失败:', error)
      // message.error(t('app.docs.detail.permissionUpdateFailed'))
    } finally {
      setLoading(false)
    }
  }

  // 取消修改
  const handleCancel = () => {
    // 重置临时状态
    setTempManagerIds(managerIds)
    setTempPublicScope(publicScope)
    setTempScopeIds(scopeIds)
    onClose()
  }

  // 处理添加管理员
  const handleAddManagers = (newManagerIds: number[]) => {
    setTempManagerIds(newManagerIds)
    // 自定义模式下，自动把管理员并入自定义查看人员
    setTempScopeIds((prev) =>
      tempPublicScope === 'custom' ? uniqueArray(newManagerIds, prev) : prev,
    )
    setShowAddManagerModal(false)
  }

  // 处理添加查看者
  const handleAddViewers = (newViewerIds: number[]) => {
    setTempScopeIds(newViewerIds)
    setShowAddViewerModal(false)
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
    //有些ID可能来自employee_id
    try {
      const { data } = usePersonnelData.getState()
      console.log(data);
      const employee = data.employee?.find((emp: any) => Number(emp.id) === Number(userId))
      if (employee) {
        return {
          id: userId,
          name: employee.name,
          avatar: '',
        }
      }
    } catch {}
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

  return (
    <>
      <Drawer
        title={null}
        placement='right'
        onClose={onClose}
        open={open}
        width={400}
        styles={{
          body: { padding: 0 },
          header: { display: 'none' },
        }}
      >
        <div className={`h-full flex flex-col ${styles.permissionSidebar}`}>
          {/* 头部 */}
          <div className='flex items-center justify-between p-4 pb-0 border-b border-[#EFF1F4]'>
            <div className='flex-1'>
              <h3 className='text-sm font-medium text-[#0C1F17] mb-3'>
                {t('app.docs.detail.permission')}
              </h3>
              <div className='relative'>
                <div className='absolute top-0 left-0 h-[2px] bg-black w-15'></div>
              </div>
            </div>
            <button
              onClick={onClose}
              className='text-[#626466] p-[3px] cursor-pointer hover:bg-[#F5F5F5] hover:text-[#626466] transition-colors'
            >
              <CloseOutlined className='text-base' />
            </button>
          </div>

          {/* 内容区域 */}
          <div className={`flex-1 p-4 space-y-6 overflow-y-auto ${scrollStyles.scroll}`}>
            {/* 可管理部分 */}
            <div>
              <div className='flex items-center justify-between mb-3'>
                <label className='text-base font-medium text-[#3C4149]'>
                  {t('app.docs.detail.manageable')}
                </label>
                <button
                  className='flex items-center gap-1 text-[#0C99FF] text-base font-medium rounded px-[2px] cursor-pointer'
                  onClick={() => setShowAddManagerModal(true)}
                >
                  <PlusOutlined />
                  {t('app.docs.detail.add')}
                </button>
              </div>
              <div
                className={`bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-1 pr-[6px] min-h-[100px] max-h-[200px] overflow-y-auto ${scrollStyles.scroll}`}
              >
                {canRenderContent ? (
                  <div className='flex flex-wrap gap-1'>
                    {tempManagerIds.map(userId =>
                      renderUserTag(
                        userId,
                        () => {
                          // 保护：至少保留1个可管理成员
                          setTempManagerIds(prev =>
                            prev.length <= 1 ? prev : prev.filter(id => id !== userId),
                          )
                        },
                        tempManagerIds.length <= 1
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

            {/* 仅查看部分 */}
            <div>
              <div className='flex items-center justify-between mb-3'>
                <label className='text-base font-medium text-[#3C4149]'>
                  {t('app.docs.detail.viewOnly')}
                </label>
                {tempPublicScope === 'custom' && (
                  <button
                    className='flex items-center gap-1 text-[#0C99FF] text-base font-medium rounded px-[2px] cursor-pointer'
                    onClick={() => setShowAddViewerModal(true)}
                  >
                    <PlusOutlined />
                    {t('app.docs.detail.add')}
                  </button>
                )}
              </div>
              <div className="flex gap-3 items-center">
                <Radio.Group
                  value={tempPublicScope}
                  onChange={(e) => {
                    const value = e.target.value
                    setTempPublicScope(value)
                    if (value === 'company') {
                      setTempScopeIds([])
                    } else {
                      // 切换到自定义时，确保管理员包含在scope_ids中
                      setTempScopeIds(prev => uniqueArray(tempManagerIds, prev))
                    }
                  }}
                >
                  <Radio value='custom'>{t('app.docs.detail.customize')}</Radio>
                  <Radio value='company'>{t('app.docs.detail.organization')}</Radio>
                </Radio.Group>
              </div>

              {/* 自定义查看范围成员展示 */}
              {tempPublicScope === 'custom' && (
                <div className='mt-3'>
                  <div
                    className={`bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-3 min-h-[100px] max-h-[200px] overflow-y-auto ${scrollStyles.scroll}`}
                  >
                    {canRenderContent ? (
                      <div className='flex flex-wrap gap-2'>
                        {tempScopeIds.map(userId =>
                          renderUserTag(
                            userId,
                            () => {
                              // 保护：至少保留1个查看成员
                              setTempScopeIds(prev =>
                                prev.length <= 1 ? prev : prev.filter(id => id !== userId),
                              )
                            },
                            tempManagerIds.includes(userId) || tempScopeIds.length <= 1
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
          </div>

          {/* 底部按钮 */}
          <div className='flex gap-[6px] justify-end items-center px-5 py-2.5 border-t border-gray-200'>
            <button
              className='px-6 py-2 bg-[#F5F5F5] text-[#0C1F17] rounded-md text-sm cursor-pointer font-medium hover:bg-[#F5F5F5] disabled:cursor-not-allowed'
              onClick={handleCancel}
              disabled={loading}
            >
              {t('app.docs.detail.cancel')}
            </button>
            <button
              className={`px-6 py-2 rounded-md text-sm font-medium flex items-center cursor-pointer gap-2 ${loading ? 'bg-[#0C99FF] text-[#ffffff] opacity-50 cursor-not-allowed' : 'bg-[#0C99FF] text-[#ffffff] hover:bg-[#0C99FF]'}`}
              onClick={handleSavePermissions}
              disabled={loading}
            >
              {loading ? <Spin size='small' /> : null}
              {t('app.docs.detail.confirm')}
            </button>
          </div>
        </div>
      </Drawer>

      {/* 添加管理员弹窗 */}
      <AddMembersModal
        open={showAddManagerModal}
        onClose={() => setShowAddManagerModal(false)}
        onConfirm={handleAddManagers}
        initialSelectedIds={tempManagerIds}
        lockedIds={[]}
        minSelected={1}
      />

      {/* 添加查看者弹窗 */}
      <AddMembersModal
        open={showAddViewerModal}
        onClose={() => setShowAddViewerModal(false)}
        onConfirm={handleAddViewers}
        initialSelectedIds={tempScopeIds}
        lockedIds={Array.from(new Set([...tempManagerIds]))}
        minSelected={1}
      />
    </>
  )
}