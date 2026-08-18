import { FC, useRef, useEffect, useState } from 'react'
import { Skeleton } from 'antd'
import { cn, scrollToEnd } from '@/utils'
import { AIDialog, UserDialog } from '@/components/dialog'
import { DialogInput } from '@/components/dialog/DialogInput'
import SendIcon from '@/pages/app/home/images/send.svg'
import { useIframeDialog } from './useIframeDialog'

export type Content = {
  agentDetail: any
  answering: ValueController<boolean>
  client_id?: string
}

export const Content: FC<Style & Content> = (props) => {
  const { agentDetail, answering, client_id } = props
  const { dialog, startQA, historyLoading } = useIframeDialog(
    agentDetail,
    answering.onChange,
  )
  const [search, setSearch] = useState('')
  const shouldScroll = useRef(true)
  const container = useRef<HTMLDivElement>(null)
  const isSystemRobot = useMemo(() => {
    const robots = [
      // 官网机器人
      'B7C5Tks',
      // 应用内部机器人
      'HjW3zoe',
    ]
    return robots.includes(client_id!)
  }, [client_id])
  const onSend = async () => {
    if (search.trim()) {
      startQA(search)
      setSearch('')
      shouldScroll.current = true
    }
  }

  useEffect(() => {
    const dom = container.current
    if (!dom || !shouldScroll.current) return
    scrollToEnd(dom)
  }, [dialog])

  // 监听对话长度变化，当新增对话时，重置 shouldScroll 为 true
  useEffect(() => {
    shouldScroll.current = true
  }, [dialog.length])

  if (historyLoading) {
    return (
      <Skeleton active className={props.className} paragraph={{ rows: 15 }} />
    )
  }

  return (
    <div className={cn('overflow-hidden flex flex-col', props.className)}>
      <div
        className={cn('flex-1 overflow-auto p-2')}
        ref={container}
        onWheel={(e) => {
          if (e.deltaY < 0) {
            shouldScroll.current = false
          }
        }}
      >
        <div className={cn('w-[80vw] mx-auto', 'flex flex-col gap-2')}>
          {dialog.map((item, index) => {
            console.log(item)
            switch (item.role) {
              case 'question':
                return <UserDialog key={index} value={item} />
              case 'answer':
                return (
                  <AIDialog
                    key={index}
                    value={item}
                    showConcatus={Boolean(
                      isSystemRobot && !item.reference?.length && index !== 0,
                    )}
                  />
                )
              default:
                return null
            }
          })}
        </div>
      </div>

      <DialogInput
        value={search}
        onChange={setSearch}
        onSubmit={onSend}
        className='w-[80vw] mx-auto my-2'
      >
        {search.trim() && !answering.value ? (
          <div
            className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#1e1f28] cursor-pointer transition-colors hover:bg-[#2a2b36] ml-auto'
            onClick={onSend}
          >
            <div className='relative w-4 h-4 flex items-center justify-center'>
              <img src={SendIcon} alt='send' />
            </div>
          </div>
        ) : (
          <div className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#dfe0eb] cursor-not-allowed ml-auto'>
            <div className='relative w-4 h-4 flex items-center justify-center'>
              <img src={SendIcon} alt='send' />
            </div>
          </div>
        )}
      </DialogInput>
    </div>
  )
}
