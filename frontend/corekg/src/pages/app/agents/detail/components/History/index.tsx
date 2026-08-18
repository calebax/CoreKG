import { FC } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button, Empty, Skeleton } from 'antd'
import { produce } from 'immer'
import { removeChatSession, renameChatSession, setTopChatSession } from '@/api'
import { cn } from '@/utils'
import MessageIcon from '@/assets/icons/message.svg?react'
import { scroll } from '@/styles/scroll'
import { useHistory } from '../HistoryContext'
import BaseInfo from './BaseInfo'
import HistoryItem from './HistoryItem'

type Style = {
  className?: string
}

export type History = {
  session_id?: number
  agentDetail: any
  answering: boolean
}
export const History: FC<Style & History> = (props) => {
  const { agentDetail, answering, session_id } = props

  return (
    <div
      className={cn(
        'flex-none w-[210px] h-full overflow-hidden',
        'border border-[#CCE9FF]',
        'flex flex-col gap-2',
        props.className,
      )}
    >
      <div className='flex-none p-4 pb-0 flex flex-col gap-6'>
        <BaseInfo agentDetail={agentDetail} />
        <Link to={'?'}>
          <Button
            className='w-full border border-[#0C99FF] bg-white hover:bg-white text-[#0C99FF]'
            disabled={answering}
          >
            <MessageIcon className='w-6 h-6' />
            <svg
              xmlns='http://www.w3.org/2000/svg'
              width='18'
              height='18'
              viewBox='0 0 18 18'
              fill='none'
            >
              <path
                d='M9.60252 1.76855H8.39716C8.29002 1.76855 8.23645 1.82213 8.23645 1.92927V8.2373H2.25031C2.14317 8.2373 2.0896 8.29088 2.0896 8.39802V9.60338C2.0896 9.71052 2.14317 9.76409 2.25031 9.76409H8.23645V16.0721C8.23645 16.1793 8.29002 16.2328 8.39716 16.2328H9.60252C9.70967 16.2328 9.76324 16.1793 9.76324 16.0721V9.76409H15.7503C15.8575 9.76409 15.911 9.71052 15.911 9.60338V8.39802C15.911 8.29088 15.8575 8.2373 15.7503 8.2373H9.76324V1.92927C9.76324 1.82213 9.70967 1.76855 9.60252 1.76855Z'
                fill='currentColor'
              />
            </svg>
            <span>新建会话</span>
          </Button>
        </Link>
        <div className='flex items-center gap-1 px-2'>
          <svg
            xmlns='http://www.w3.org/2000/svg'
            width='18'
            height='18'
            viewBox='0 0 18 18'
            fill='none'
          >
            <path
              d='M9 15.75C7.275 15.75 5.772 15.1782 4.491 14.0347C3.21 12.8912 2.4755 11.463 2.2875 9.75H3.825C4 11.05 4.57825 12.125 5.55975 12.975C6.54125 13.825 7.688 14.25 9 14.25C10.4625 14.25 11.7032 13.7407 12.7222 12.7222C13.7412 11.7037 14.2505 10.463 14.25 9C14.2495 7.537 13.7402 6.2965 12.7222 5.2785C11.7042 4.2605 10.4635 3.751 9 3.75C8.1375 3.75 7.33125 3.95 6.58125 4.35C5.83125 4.75 5.2 5.3 4.6875 6H6.75V7.5H2.25V3H3.75V4.7625C4.3875 3.9625 5.16575 3.34375 6.08475 2.90625C7.00375 2.46875 7.9755 2.25 9 2.25C9.9375 2.25 10.8158 2.42825 11.6347 2.78475C12.4537 3.14125 13.1663 3.62225 13.7723 4.22775C14.3783 4.83325 14.8595 5.54575 15.216 6.36525C15.5725 7.18475 15.7505 8.063 15.75 9C15.7495 9.937 15.5715 10.8152 15.216 11.6347C14.8605 12.4542 14.3793 13.1668 13.7723 13.7723C13.1653 14.3778 12.4527 14.859 11.6347 15.216C10.8167 15.573 9.9385 15.751 9 15.75ZM11.1 12.15L8.25 9.3V5.25H9.75V8.7L12.15 11.1L11.1 12.15Z'
              fill='#1D2129'
            />
          </svg>
          <span className='text-title font-medium'>历史会话</span>
        </div>
      </div>
      <HistoryInner session_id={session_id} />
    </div>
  )
}

const HistoryInner: FC<{ session_id?: number }> = (props) => {
  const { session_id } = props
  const navigate = useNavigate()
  const { value, setValue, loading, refresh } = useHistory()

  if (!value || loading) {
    return <Skeleton active />
  }
  if (value.length === 0) {
    return <Empty />
  }

  const onDel = async (id: number) => {
    setValue(value.filter((item) => item.ID !== id))
    await removeChatSession(id)
    if (id === session_id) {
      navigate('?')
    }
  }
  const onRename = async (id: number, newName: string) => {
    // 重命名会话标题后，需要重新拉取 chat.ListChatSession 列表，保证左侧历史会话名称与后端数据同步
    await renameChatSession({ id, name: newName })
    await refresh()
  }
  const onToTop = async (id: number) => {
    await setTopChatSession(id)
    refresh()
  }

  return (
    <div
      className={cn('flex-1 overflow-auto', 'flex flex-col gap-2 p-4', scroll)}
    >
      {value.map((item, i) => {
        return (
          <HistoryItem
            key={item.ID}
            name={item.name}
            session_id={item.ID}
            active={session_id === item.ID}
            isTop={i === 0}
            onDel={() => onDel(item.ID)}
            onRename={(val) => onRename(item.ID, val)}
            onToTop={() => onToTop(item.ID)}
          />
        )
      })}
    </div>
  )
}
