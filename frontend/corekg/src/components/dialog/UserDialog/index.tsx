import { FC } from 'react'
import { Image } from 'antd'
import { cn } from '@/utils'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import { AttachmentList } from '../AttachmentList'

export type Attachment = {
  id?: string
  url?: string
  md_url?: string
  type: string
  name: string
  mime_type?: string
  loading?: boolean
}

export type UserDialog = {
  created_at?: string
  role: 'question'
  content: string
  images?: string[]
  attachments?: Attachment[]
}

type UserDialogProps = {
  markdownClassName?: string
  className?: string
  value: UserDialog
}
/**
 * @example
 * ```js
 * const dialog:(UserDialog|AIDialog)[]
 *
 * dialog.map(item=>{
 *    if(item.role==='question')return <UserDialog value={item}/>
 *    else return <AIDialog value={item}/>
 * })
 * ```
 */
export const UserDialog: FC<UserDialogProps> = (props) => {
  const { value, className } = props
  const { content, images, attachments } = value

  return (
    <div
      className={cn(
        'flex-none flex flex-col gap-2 self-end max-w-[94.6%] mb-6',
      )}
    >
      {/* 附件展示 */}
      <AttachmentList attachments={attachments || []} />
      {content && (
        <MarkdownPreview
          content={content}
          className={cn(
            'bg-[#EAF2FF]! text-[#00000099] break-all rounded-lg',
            className,
          )}
        />
      )}
      {images?.length ? (
        <span className='flex gap-2 self-end'>
          {images.map((url) => (
            <Image key={url} src={url} className='w-8 h-8 mt-2'></Image>
          ))}
        </span>
      ) : null}
    </div>
  )
}
