import { FC, useRef, useEffect, useState } from 'react'
import { Skeleton } from 'antd'
import { cn, scrollToEnd } from '@/utils'
import { AIDialog, UserDialog } from '@/components/dialog'
import { DialogInput } from '@/components/dialog/DialogInput'
import SendIcon from '@/pages/app/home/images/send.svg'
import { scroll } from '@/styles/scroll'
import { useDialog } from './useDialog'

export type Content = {
  agentDetail: any
  session_id?: number
  answering: ValueController<boolean>
  /** 下次切换session_id本组件不会被重置 */
  setSessionId: (id: number) => void
}
export const Content: FC<Style & Content> = (props) => {
  const { agentDetail, session_id, answering, setSessionId } = props
  const { dialog, startQA, historyLoading } = useDialog(
    agentDetail,
    session_id,
    setSessionId,
    answering.onChange,
  )
  const [search, setSearch] = useState('')
  const shouldScroll = useRef(true)
  const container = useRef<HTMLDivElement>(null)

  const onSend = async () => {
    startQA(search)
    setSearch('')
    shouldScroll.current = true
  }

  useEffect(() => {
    const dom = container.current
    if (!dom || !shouldScroll.current) return
    scrollToEnd(dom)
  }, [dialog])

  if (historyLoading) {
    return (
      <Skeleton
        active
        className={props.className}
        paragraph={{ rows: 15 }}
      ></Skeleton>
    )
  }
  const showReference = true
  return (
    <div className={cn(' overflow-hidden flex flex-col', props.className)}>
      {historyLoading ? (
        <Skeleton active paragraph={{ rows: 15 }} className='overflow-hidden' />
      ) : null}
      <div
        className={cn('flex-1 overflow-auto p-2', scroll)}
        ref={container}
        onWheel={(e) => {
          if (e.deltaY < 0) {
            shouldScroll.current = false
          }
        }}
      >
        <div className={cn('w-[60vw] mx-auto', 'flex flex-col gap-2')}>
          {dialog.map((item, index) => {
            switch (item.role) {
              case 'question':
                return <UserDialog key={index} value={item} />
              case 'answer':
                return (
                  <AIDialog key={index} showReference={showReference} value={item} />
                )
            }
          })}
        </div>
      </div>

      <DialogInput
        value={search}
        onChange={setSearch}
        onSubmit={onSend}
        className='w-[60vw] mx-auto my-2'
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
