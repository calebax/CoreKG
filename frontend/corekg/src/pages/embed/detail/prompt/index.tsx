import { useState, FC, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Skeleton, Button } from 'antd'
import { Empty } from 'antd'
import { ToolOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import { getIframeHistory } from '@/api/iframe'
import { useIframeAuth } from '@/hooks/useIframeAuth'
import { IframeProvider } from '../components/IframeContext'
import { Content } from './Content'

const PromptAgent: FC<{ workflow?: boolean }> = (props) => {
  const { workflow } = props
  const { id } = useParams<{ id: string }>()
  const {
    loading: authLoading,
    error: authError,
    authenticated,
  } = useIframeAuth()
  const [agentDetail, setAgentDetail] = useState<any>(null)
  const [dataLoading, setDataLoading] = useState(false)
  const [answering, setAnswering] = useState(false)

  // 检查 agentDetail 是否有有效内容
  const hasValidContent = (detail: any): boolean => {
    if (!detail) return false

    // 如果是数组，检查是否非空
    if (Array.isArray(detail)) {
      return detail.length > 0
    }

    // 如果是对象，检查关键字段
    if (typeof detail === 'object') {
      return Object.keys(detail).length > 0
    }

    return true
  }

  const loadAgentData = async () => {
    if (!id) return

    setDataLoading(true)
    try {
      const data = await getIframeHistory()
      setAgentDetail(data.messages)
    } catch (err) {
      console.error('加载agent信息失败:', err)
      setAgentDetail(null)
    } finally {
      setDataLoading(false)
    }
  }

  const handleRetry = () => {
    setAgentDetail(null)
    loadAgentData()
  }

  useEffect(() => {
    if (authenticated && id) {
      loadAgentData()
    }
  }, [authenticated, id])

  if (authLoading || dataLoading) {
    return <Skeleton active paragraph={{ rows: 15 }} className='m-4' />
  }

  if (authError) {
    return (
      <div className='flex items-center justify-center h-full'>
        <div className='text-red-500'>错误: {authError}</div>
      </div>
    )
  }

  // 当 agentDetail 没有内容时显示维护页面
  if (!agentDetail || !hasValidContent(agentDetail)) {
    return (
      <div className='w-full h-full flex flex-col overflow-hidden bg-[#F8FCFF]'>
        <div className='flex items-center justify-center h-full'>
          <Empty
            image={
              <div className='text-6xl text-blue-400 mb-4'>
                <ToolOutlined />
              </div>
            }
            description={
              <div className='text-center'>
                <h3 className='text-lg font-medium text-gray-900 mb-2'>
                  抱歉，无法提供服务。
                </h3>
              </div>
            }
          ></Empty>
        </div>
      </div>
    )
  }

  return (
    <IframeProvider agentDetail={agentDetail}>
      <div className='w-full h-full flex flex-col overflow-hidden bg-[#F8FCFF]'>
        <h1
          className={cn(
            'flex-none my-7',
            'text-title font-bold text-[28px] text-center font-alimama-thin',
          )}
        >
          {agentDetail.session_name || 'AI Assistant'}
        </h1>

        <Content
          workflow={workflow}
          className='flex-1 overflow-hidden'
          agentDetail={agentDetail}
          answering={{
            value: answering,
            onChange: (val) => setAnswering(Boolean(val)),
          }}
        />
      </div>
    </IframeProvider>
  )
}

export default PromptAgent
