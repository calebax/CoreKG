import { useState, useEffect } from 'react'
import { Drawer, Tag, message, Spin } from 'antd'
import { CloseOutlined, PlusOutlined, LoadingOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getResourcePerm, setResourcePerm } from '@/api/knowledge'
import useUserData from '@/hooks/data/useUserData'
import useLocalStore from '@/stores/local'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import AddMembersModal from './AddMembersModal'
import styles from './PermissionSidebar.module.scss'

// 自定义加载图标
const LoadingIcon = (
  <LoadingOutlined
    spin
    style={{
      fontSize: '16px',
      color: '#0C99FF',
    }}
  />
)

interface FilePermissionSidebarProps {
  open: boolean
  onClose: () => void
  fileId: number
}

export default function FilePermissionSidebar({
  open,
  onClose,
  fileId,
}: FilePermissionSidebarProps) {
  const { t } = useTranslation(['pages', 'messages', 'common'])
  const { userList, loading: userLoading, loadUserList } = useUserData()
  const { userInfo } = useLocalStore()

  const [banList, setBanList] = useState<number[]>([])
  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(false)
  const [showAddModal, setShowAddModal] = useState(false)

  // 获取文件权限数据
  const fetchPermission = async () => {
    if (!fileId) return
    setFetching(true)
    try {
      const res = await getResourcePerm({
        resource_id: fileId,
        resource_type: 'forest_file',
      })
      setBanList(res.access_result?.BanList || [])
    } catch (error) {
      console.log('获取文件权限失败:', error)
    } finally {
      setFetching(false)
    }
  }

  useEffect(() => {
    if (open) {
      loadUserList()
      fetchPermission()
    } else {
      setBanList([])
    }
  }, [open, fileId])

  // 保存权限更改
  const handleSave = async () => {
    setLoading(true)
    try {
      await setResourcePerm({
        resource_id: fileId,
        resource_type: 'forest_file',
        perm_option: {
          ban_list: banList,
        },
      })
      message.success('保存成功')
      onClose()
    } catch (error) {
      console.log('设置文件权限失败:', error)
    } finally {
      setLoading(false)
    }
  }

  // 获取用户信息
  const getUserName = (uin: number) => {
    const user = userList.find((u) => u.value === uin)
    return user ? user.label : `用户${uin}`
  }

  // 渲染用户标签
  const renderUserTag = (uin: number) => {
    return (
      <Tag
        key={uin}
        closable
        onClose={() => setBanList((prev) => prev.filter((id) => id !== uin))}
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
        {getUserName(uin)}
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
                权限管理
              </h3>
              <div className='relative'>
                <div className='absolute top-0 left-0 h-[2px] bg-black w-15'></div>
              </div>
            </div>
            <button
              onClick={onClose}
              className='text-[#626466] p-[3px] cursor-pointer hover:bg-[#F5F5F5] transition-colors'
            >
              <CloseOutlined className='text-base' />
            </button>
          </div>

          {/* 内容区域 */}
          <div
            className={`flex-1 p-4 space-y-6 overflow-y-auto ${scrollStyles.scroll}`}
          >
            <div>
              <div className='flex items-center justify-between mb-3'>
                <label className='text-base font-medium text-[#3C4149]'>
                  不可查看
                </label>
                <button
                  className='flex items-center gap-1 text-[#0C99FF] text-base font-medium rounded px-[2px] cursor-pointer'
                  onClick={() => setShowAddModal(true)}
                >
                  <PlusOutlined />
                  添加
                </button>
              </div>
              <div
                className={`bg-[#F5F5F5] border border-[#EFF1F4] rounded-lg p-1 pr-[6px] min-h-[100px] max-h-[200px] overflow-y-auto ${scrollStyles.scroll}`}
              >
                {fetching || userLoading ? (
                  <div className='flex items-center justify-center h-20'>
                    <Spin indicator={LoadingIcon} />
                  </div>
                ) : (
                  <div className='flex flex-wrap gap-1'>
                    {banList.map((uin) => renderUserTag(uin))}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* 底部按钮 */}
          <div className='flex gap-[6px] justify-end items-center px-5 py-2.5 border-t border-gray-200'>
            <button
              className='px-6 py-2 bg-[#F5F5F5] text-[#0C1F17] rounded-md text-sm cursor-pointer font-medium hover:bg-[#F5F5F5] disabled:cursor-not-allowed'
              onClick={onClose}
              disabled={loading}
            >
              取消
            </button>
            <button
              className={`px-6 py-2 rounded-md text-sm font-medium flex items-center cursor-pointer gap-2 ${
                loading
                  ? 'bg-[#0C99FF] text-[#ffffff] opacity-50 cursor-not-allowed'
                  : 'bg-[#0C99FF] text-[#ffffff] hover:bg-[#0C99FF]'
              }`}
              onClick={handleSave}
              disabled={loading || fetching}
            >
              {loading ? <Spin size='small' /> : null}
              确定
            </button>
          </div>
        </div>
      </Drawer>

      <AddMembersModal
        open={showAddModal}
        onClose={() => setShowAddModal(false)}
        onConfirm={(ids) => {
          setBanList(ids)
          setShowAddModal(false)
        }}
        initialSelectedIds={banList}
        lockedIds={[]}
        disabledIds={userInfo.uinId ? [Number(userInfo.uinId)] : []}
        minSelected={0}
      />
    </>
  )
}
