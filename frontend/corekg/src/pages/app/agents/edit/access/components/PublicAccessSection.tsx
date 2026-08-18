import { FC, useState, useEffect } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { Switch, message } from 'antd'
import { getExternalStatus, setExternalStatus } from '@/api'

const PublicAccessSection: FC = () => {
  const { id } = useParams<{ id: string }>() // 从路由参数获取
  const [searchParams] = useSearchParams()

  const clientId = searchParams.get('client_id') || id

  const [status, setStatus] = useState<string>('disabled')
  const [loading, setLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)

  const isPublicAccess = status === 'normal'

  useEffect(() => {
    const fetchInitialStatus = async () => {
      if (!clientId) {
        setInitialLoading(false)
        return
      }

      try {
        const response = await getExternalStatus(Number(clientId))
        const statusValue =
          typeof response === 'string' ? response : response.status
        setStatus(statusValue === 'normal' ? 'normal' : 'disabled')
      } catch (error) {
        console.error('获取公开访问状态失败:', error)
        message.error('获取公开访问状态失败')
        setStatus('disabled')
      } finally {
        setInitialLoading(false)
      }
    }

    fetchInitialStatus()
  }, [clientId])

  const handleSwitchChange = async (checked: boolean) => {
    if (!clientId) {
      message.error('客户端ID不存在')
      return
    }

    const newStatus = checked ? 'normal' : 'disabled'
    const previousStatus = status

    setLoading(true)
    try {
      setStatus(newStatus)
      await setExternalStatus(Number(clientId), newStatus)
      message.success(checked ? '已开启公开访问' : '已关闭公开访问')
    } catch (error) {
      console.error('设置公开访问状态失败:', error)
      setStatus(previousStatus)
      message.error('设置失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 如果没有clientId，显示提示信息
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
