import { FC, useState } from 'react'
import { Modal, Button, message } from 'antd'
import { useRequest } from 'ahooks'
import { CloseIcon } from 'tdesign-icons-react'
import { createAgentApiKey } from '@/api/apiKey'
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
    <Modal
      title={null}
      open={visible}
      onCancel={onClose}
      width={1000}
      footer={null}
      closeIcon={<CloseIcon />}
      styles={{
        header: { display: 'none' },
      }}
    >
      <div className='p-6'>
        {/* 标题和创建按钮 */}
        <div className='flex items-center justify-between mb-6'>
          <h2 className='text-lg font-semibold text-gray-900'>API Key管理</h2>
          <Button
            type='primary'
            loading={creating}
            onClick={handleCreateApiKey}
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
    </Modal>
  )
}

export default ApiKeyManagementModal
