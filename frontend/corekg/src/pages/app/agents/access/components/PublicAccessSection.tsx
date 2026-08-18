import { FC, useState, useEffect } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { Switch, message } from 'antd'
import { getExternalStatus, setExternalStatus } from '@/api/request'
import useAccessStore from '@/stores/access'

const PublicAccessSection: FC = () => {
  const { botId: storeBotId, spaceId: storeSpaceId } = useAccessStore()
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()

  const clientId = storeBotId || searchParams.get('client_id') || id
  const currentSpaceId = storeSpaceId || searchParams.get('space_id') || ''

  const [status, setStatus] = useState<string>('disabled')
  const [loading, setLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)

  const isPublicAccess = status === 'normal'

  useEffect(() => {
    const fetchInitialStatus = async () => {
      if (!clientId || !currentSpaceId) {
        setInitialLoading(false)
        return
      }

      try {
        const response = await getExternalStatus({
          space_id: currentSpaceId,
          bot_id: clientId,
        })

        const statusValue = response?.data?.status

        setStatus(statusValue === 'normal' ? 'normal' : 'disabled')
      } catch (error) {
        console.error('获取公开访问状态失败:', error)
        setStatus('disabled')
      } finally {
        setInitialLoading(false)
      }
    }

    fetchInitialStatus()
  }, [clientId, currentSpaceId])

  const handleSwitchChange = async (checked: boolean) => {
    if (!clientId || !currentSpaceId) {
      message.error('缺失必要参数：Client ID 或 Space ID')
      return
    }

    const newStatus = checked ? 'normal' : 'disabled'
    const previousStatus = status

    setLoading(true)
    try {
      setStatus(newStatus)

      await setExternalStatus({
        space_id: currentSpaceId,
        bot_id: clientId,
        status: newStatus,
      })

      message.success(checked ? '已开启公开访问' : '已关闭公开访问')
    } catch (error) {
      console.error('设置公开访问状态失败:', error)
      setStatus(previousStatus)
      message.error('设置失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  if (!clientId) {
    return (
      <div className='flex items-center gap-3 text-gray-500'>
        <span>无法获取客户端ID</span>
      </div>
    )
  }

  return (
    <div className='flex items-center gap-3'>
      <div className='flex items-center gap-2'>
        <div className='h-4 w-1 rounded-full bg-[#165DFF]' />
        <span className='text-base font-medium text-gray-900'>公开访问</span>
      </div>
      <Switch
        checked={isPublicAccess}
        onChange={handleSwitchChange}
        loading={loading || initialLoading}
        size='small'
      />
    </div>
  )
}

export default PublicAccessSection
