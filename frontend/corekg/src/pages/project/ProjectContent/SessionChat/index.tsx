import { FC, useMemo, useState, useLayoutEffect } from 'react'
import { Link } from 'react-router-dom'
import { useBoolean, useMemoizedFn } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { ChevronLeftDoubleIcon, PlusIcon } from 'tdesign-icons-react'
import { match } from 'ts-pattern'
import { cn } from '@/utils'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useSessionInfo } from '..'
import { useProject } from '../..'
import { Dialog } from './Dialog'
import { SessionHistory } from './SessionHistory'
import AddIcon from './add.svg?react'
import HistoryIcon from './images/history.svg?react'
import Opo from './images/opo.svg?react'
import styles from './index.module.scss'
import { App, Tooltip } from 'antd'

export type SessionChat = Style & {
  hidden: boolean
  onClose: () => void
}
export const SessionChat: FC<SessionChat> = (props) => {
  const { onClose, hidden, className, style } = props
  const { sessionStatus } = useSessionInfo()
  const [isDragging, setDragging] = useState(false)
  // 侧栏宽度 历史记录300 会话最小550 最大720
  const [width, setWidth] = useState(() => {
    return sessionStatus === 'none' ? 300 : 720
  })
  const startDragging = useMemoizedFn((e) => {
    e.preventDefault()
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
    setDragging(true)
    const onDragging = (e: MouseEvent) => {
      setWidth((v) => {
        const newWidth = v + e.movementX
        if (newWidth < 550) return 550
        if (newWidth > 720) return 720
        return newWidth
      })
    }
    const stopDragging = () => {
      setDragging(false)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
      document.removeEventListener('mousemove', onDragging)
      document.removeEventListener('mouseup', stopDragging)
    }
    document.addEventListener('mousemove', onDragging)
    document.addEventListener('mouseup', stopDragging)
  })
  useLayoutEffect(() => {
    switch (sessionStatus) {
      case 'none':
        setWidth(300)
        break
      case 'new':
      case 'creating':
      case 'created':
        setWidth(720)
        break
    }
  }, [sessionStatus])

  return (
    <div className={cn('flex', className)} style={style}>
      <div
        className={cn(
          'bg-white border-t border-[#EFF1F4]',
          'overflow-hidden h-full',
          { 'transition-[300ms]': !isDragging },
          className,
        )}
        style={{
          width: hidden ? 0 : width,
        }}
      >
        <div
          className={cn('h-full overflow-hidden flex flex-col')}
          style={{ width }}
        >
          <SessionChatHeader onClose={onClose} />
          {sessionStatus === 'none' ? <SessionHistory /> : <Dialog />}
        </div>
      </div>
      {/* 拖拽分隔条，仅在未隐藏时展示 */}
      {!hidden && sessionStatus !== 'none' ? (
        <div
          onMouseDown={startDragging}
          className={cn(
            'group relative flex-none w-2 h-full cursor-col-resize select-none',
          )}
        >
          <div
            className={cn(
              'absolute inset-y-0 left-1/2 -translate-x-1/2 w-px',
              'bg-gray-200 group-hover:bg-gray-300',
            )}
          />
        </div>
      ) : null}
    </div>
  )
}

export const SessionChatHeader: FC<{
  onClose: () => void
}> = (props) => {
  const { onClose } = props
  const { t } = useTranslation('pages')
  const { message } = App.useApp()
  const {
    project_id,
    session_id,
    setSessionId,
    data: { sessions },
  } = useProject()
  const { version } = useDeployConfig()
  const { sessionStatus, setSessionStatus } = useSessionInfo()
  const sessionLabel = useMemo(() => {
    switch (sessionStatus) {
      case 'none':
        return null
      case 'new':
      case 'creating':
        return t('project.newChat')
      case 'created':
        return sessions.find((item) => item.session_id === session_id)?.name
    }
  }, [sessionStatus, session_id, sessions, t])

  const isDisabled = sessionStatus !== 'created'

  return (
    <div className='h-[50px] px-[20px] flex items-center text-[#919497] border-b border-[#EFF1F4]'>
      {/* 正在创建的session不允许跳走 */}
      <div className='flex gap-[8px] whitespace-nowrap items-center text-[#0C1F17] font-[500]'>
        <HistoryIcon />
        {t('project.historyRecord')}
      </div>
      {/* 新会话按钮和展开收起图标容器 */}
      <div className='flex gap-[8px] items-center ml-auto'>
        {/* 左侧按钮 */}
        <div
          className={cn(
            'rounded-[6px]',
            'h-[30px] flex items-center justify-center',
            'gap-[5px] px-[10px]',
            styles.opo_btn,
            isDisabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
          )}
          onClick={() => {
            if (isDisabled) {
              message.info(t('project.alreadyLatestConversation'))
              return
            }
            setSessionStatus('new')
            setSessionId()
          }}
        >
          <Opo />
          <span className={`${styles.text} whitespace-nowrap`}>
            {/* {match(version)
              .with('saas', () => 'CoreKG AI')
              .with('custom', () => 'AI')
              .with('international', () => t('project.opoAI'))
              .exhaustive()} */}
            {t('project.newSession')}
          </span>
        </div>
        <Tooltip title='收起历史记录'>
          <ChevronLeftDoubleIcon
            onClick={onClose}
            className='cursor-pointer text-xl'
          />
        </Tooltip>
      </div>
      {/* <AddIcon
        className='cursor-pointer mr-4'
        onClick={() => {
          // 正在创建session时不能开新的
          if (sessionStatus !== 'creating') {
            setSessionStatus('new')
            setSessionId()
          }
        }}
      /> */}
    </div>
  )
}
