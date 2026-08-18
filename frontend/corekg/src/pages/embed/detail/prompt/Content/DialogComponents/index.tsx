import { FC, useRef, useEffect, useState } from 'react'
import { Skeleton } from 'antd'
import { cn, scrollToEnd } from '@/utils'
import { AIDialog, DialogList, UserDialog } from '@/components/dialog'
import { DialogInput as GlobalDialogInput } from '@/components/dialog/DialogInput'
import SendIcon from '@/pages/app/home/images/send.svg'

export const Dialog: FC<{ dialog: DialogList; loading: boolean }> = (props) => {
  const { dialog, loading } = props
  const shouldScroll = useRef(true)
  const container = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const dom = container.current
    if (!dom || !shouldScroll.current) return
    scrollToEnd(dom)
  }, [dialog])

  useEffect(() => {
    shouldScroll.current = true
  }, [dialog.length])
  if (loading) {
    return (
      <Skeleton active paragraph={{ rows: 15 }} className='overflow-hidden' />
    )
  }
  return (
    <div
      className={cn('flex-1 overflow-auto p-2')}
      ref={container}
      onWheel={(e) => {
        if (e.deltaY < 0) {
          shouldScroll.current = false
        }
      }}
    >
      <div className={cn('w-[60vw] mx-auto', 'flex flex-col gap-2')}>
        {dialog.map((item, i) => {
          switch (item.role) {
            case 'question':
              return <UserDialog key={i} value={item} />
            case 'answer':
              return <AIDialog key={i} value={item} showConcatus />
          }
        })}
      </div>
    </div>
  )
}

export const DialogInput: FC<{
  onSend: (val: string) => void
  disabled: boolean
}> = (props) => {
  const { onSend: _onSend, disabled } = props
  const [search, setSearch] = useState('')
  const onSend = () => {
    _onSend(search)
    setSearch('')
  }
  return (
    <GlobalDialogInput
      value={search}
      onChange={setSearch}
      onSubmit={onSend}
      className='w-[60vw] mx-auto my-2'
    >
      {search.trim() && !disabled ? (
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
    </GlobalDialogInput>
  )
}
