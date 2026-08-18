import { FC, useMemo } from 'react'
import dayjs from 'dayjs'
import { cn } from '@/utils'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import type { UserDialog } from '@/components/dialog'
import { AttachmentList } from '@/components/dialog'
import useLocalStore from '@/stores/local'

export const ProjectUserDialog: FC<
  Style & { value: UserDialog; isCompact?: boolean }
> = (props) => {
  const { value, className, style, isCompact } = props
  const { content, created_at, attachments } = value
  const { userInfo, uinList } = useLocalStore()
  const { avatar } = userInfo
  const name = useMemo(() => {
    const current = uinList.find(
      (item) => String(item.id) === String(userInfo.uinId),
    )
    return current?.uinName || userInfo.name
  }, [uinList, userInfo.uinId, userInfo.name])
  const time = useMemo(() => {
    return dayjs(created_at).format('YYYY/MM/DD HH:mm')
  }, [created_at])

  // 当 isCompact 为 true 时，使用左对齐布局（一列前后展示）
  // 否则使用右对齐布局（一左一右展示）
  return (
    <div
      className={cn(
        'flex gap-4',
        {
          'flex-row-reverse': !isCompact, // 非紧凑模式：右对齐
        },
        className,
      )}
      style={style}
    >
      <img src={avatar} className='w-9 h-9 rounded-full' />
      <div className='flex flex-col gap-2.5'>
        <span
          className={cn('flex items-center', {
            'justify-end': !isCompact, // 非紧凑模式：右对齐
          })}
        >
          <span className='font-medium'>{name}</span>
          <span className='text-[#919497] text-xs ml-1'>{time}</span>
        </span>
        {/* 附件展示 */}
        <AttachmentList attachments={attachments || []} isCompact={isCompact} />
        {content ? (
          <MarkdownPreview content={content} className='p-0! bg-white!' />
        ) : null}
      </div>
    </div>
  )
}
