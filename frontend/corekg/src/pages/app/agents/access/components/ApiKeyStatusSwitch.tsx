import { FC } from 'react'
import { Switch, message } from 'antd'
import { useRequest } from 'ahooks'
import { setAgentApiKeyStatus } from '@/api/apiKey'

interface ApiKeyStatusSwitchProps {
  agentId: number
  apiKeyId: number
  currentStatus: string
  onStatusChange?: (newStatus: string) => void
}

const ApiKeyStatusSwitch: FC<ApiKeyStatusSwitchProps> = ({
  agentId,
  apiKeyId,
  currentStatus,
  onStatusChange,
}) => {
  const isActive = currentStatus === 'normal'

  const { loading, run: updateStatus } = useRequest(
    (status: string) =>
      setAgentApiKeyStatus({
        agent_id: agentId,
        apikey_id: apiKeyId,
        status,
      }),
    {
      manual: true,
      onSuccess: (_, [status]) => {
        message.success('状态更新成功')
        onStatusChange?.(status)
      },
      onError: (error) => {
        message.error(error.message || '状态更新失败')
      },
    },
  )

  const handleChange = (checked: boolean) => {
    const newStatus = checked ? 'normal' : 'disabled'
    updateStatus(newStatus)
  }

  return (
    <Switch
      checked={isActive}
      onChange={handleChange}
      loading={loading}
      size='small'
    />
  )
}

export default ApiKeyStatusSwitch
