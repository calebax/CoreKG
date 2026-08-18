import { FC, useRef, useEffect } from 'react'
import { Skeleton } from 'antd'
import { cn, scrollToEnd } from '@/utils'
import { AIDialog, UserDialog } from '@/components/dialog'
import { scroll } from '@/styles/scroll'
import { useDialog } from '../utils/useDialog'
import { DialogInput } from './DialogInput'

type QAContent = {
  session_id: number
  question_id?: any
}

export const QAContent: FC<QAContent> = (props) => {
  const { session_id, question_id } = props
  const { isAnswering, loading, sessionInfo, dialog, startQA, stopQA } =
    useDialog(session_id, question_id)
  const shouldScroll = useRef(true)
  const container = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const dom = container.current
    if (!dom || !shouldScroll.current) return
    scrollToEnd(dom)
  }, [dialog])

  // 监听对话长度变化，当新增对话时，重置 shouldScroll 为 true
  useEffect(() => {
    shouldScroll.current = true
  }, [dialog.length])

  const handleSend = (data: Parameters<typeof startQA>[0]) => {
    startQA(data)
    shouldScroll.current = true
  }
  return (
    <div
      className={cn(
        'h-full bg-white relative overflow-hidden',
        'flex flex-col break-all',
      )}
    >
      {/* 顶部渐变遮罩 */}
      <div
        className={cn(
          'absolute left-0 right-0 top-0 h-12 z-10 pointer-events-none',
          'bg-gradient-to-b from-white via-white/80 to-transparent',
        )}
      ></div>
      <div
        className={cn('flex-1 overflow-auto scrollbar-hide', scroll)}
        ref={container}
        onWheel={(e) => {
          if (e.deltaY < 0) {
            shouldScroll.current = false
          }
        }}
      >
        <div className='w-[60vw] mx-auto'>
          {loading ? <Skeleton active paragraph={{ rows: 10 }} /> : null}
          <div className={cn('flex flex-col mt-14 mb-30', { hidden: loading })}>
            {dialog.map((item, i) => {
              switch (item.role) {
                case 'question':
                  return (
                    <UserDialog key={i} value={item} className='self-end' />
                  )
                case 'answer':
                  return (
                    <AIDialog key={i} value={item} className='' showReference />
                  )
              }
            })}
          </div>
        </div>
      </div>

      <DialogInput
        loading={loading}
        isAnswering={isAnswering}
        onSend={handleSend}
        onStop={stopQA}
        sessionInfo={sessionInfo}
        className='w-[60vw] mx-auto mb-6'
      />
    </div>
  )
}
