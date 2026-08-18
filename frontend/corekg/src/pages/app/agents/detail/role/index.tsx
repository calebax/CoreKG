import { useState, FC } from 'react'
import { Skeleton } from 'antd'
import { cn } from '@/utils'
import { useSessionId } from '@/components/dialog/utils'
import { withAgentStyle } from '../../AgentStyleProvider'
import { History } from '../components/History'
import { HistoryProvider } from '../components/HistoryContext'
import { useAgentInfo } from '../utils'
import { Content } from './Content'

const RoleAgent: FC = withAgentStyle((props) => {
  const navigate = useNavigate()
  const { agentId, agentDetail } = useAgentInfo()
  const [answering, setAnsering] = useState(false)
  const { sessionId, key, setSessionId } = useSessionId()

  if (!agentId) {
    setTimeout(() => {
      navigate('/agents', { replace: true })
    }, 0)
    return null
  }
  if (!agentDetail) {
    return (
      <HistoryProvider agentId={agentId}>
        <Skeleton active paragraph={{ rows: 15 }} className='m-4' />
      </HistoryProvider>
    )
  }

  return (
    <HistoryProvider agentId={agentId}>
      <div
        className={cn(
          'w-full h-full flex overflow-hidden',
          props.extraClassName,
        )}
      >
        <History
          session_id={sessionId}
          agentDetail={agentDetail}
          answering={answering}
        />
        <div
          className={cn(
            'flex-1 overflow-hidden',
            'bg-[#F8FCFF]',
            ' flex flex-col',
          )}
        >
          <h1
            className={cn(
              'flex-0 my-7',
              'text-title font-bold text-[28px] text-center font-alimama-thin',
            )}
          >
            {agentDetail.show_name}
          </h1>
          <Content
            key={`${agentId}-${key}`}
            className='flex-1 overflow-hidden'
            agentDetail={agentDetail}
            session_id={sessionId}
            answering={{
              value: answering,
              onChange: (val) => setAnsering(Boolean(val)),
            }}
            setSessionId={(id) => setSessionId(id, true)}
          />
        </div>
      </div>
    </HistoryProvider>
  )
})
export default RoleAgent
