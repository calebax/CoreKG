import { FC, useState } from 'react'
import { Modal, Button, message, Drawer } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { CloseIcon, PlusIcon } from 'tdesign-icons-react'
import { createAgentApiKey } from '@/api/apiKey'
import styles from '../../styles.module.scss'
import ApiKeyTable from './ApiKeyTable'

interface ApiKeyManagementModalProps {
  visible: boolean
  onClose: () => void
  agentId: number
}

const ApiKeyManagementModal: FC<ApiKeyManagementModalProps> = ({
  visible,
  onClose,
  agentId,
}) => {
  const [tableRefreshKey, setTableRefreshKey] = useState(0)

  // 创建API Key
  const { loading: creating, run: createApiKey } = useRequest(
    () => createAgentApiKey({ agent_id: agentId }),
    {
      manual: true,
      onSuccess: () => {
        message.success('API Key创建成功')
        setTableRefreshKey((prev) => prev + 1) // 刷新表格
      },
      onError: (error) => {
        message.error(error.message || '创建失败')
      },
    },
  )

  const handleCreateApiKey = () => {
    if (!agentId) {
      message.error('Agent ID不存在')
      return
    }
    createApiKey()
  }

  return (
    <Drawer
      title={'API Key管理'}
      open={visible}
      onClose={onClose}
      width={926}
      closeIcon={<CloseIcon />}
      className={styles.drawer}
    >
      <div className='p-6 pt-0'>
        {/* 标题和创建按钮 */}
        <div className='flex items-center justify-between mb-6'>
          <Button
            loading={creating}
            onClick={handleCreateApiKey}
            className='ml-auto'
            icon={<PlusOutlined />}
          >
            创建API KEY
          </Button>
        </div>

        {/* API Key列表表格 */}
        <ApiKeyTable
          agentId={agentId}
          refreshKey={tableRefreshKey}
          onRefresh={() => setTableRefreshKey((prev) => prev + 1)}
        />
      </div>
    </Drawer>
  )
}

export default ApiKeyManagementModal
