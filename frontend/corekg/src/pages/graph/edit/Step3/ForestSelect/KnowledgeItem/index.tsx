import { FC, ReactNode } from 'react'
import { Checkbox, Typography } from 'antd'
import { cn } from '@/utils'
import Icon from './icon.png'
import styles from './styles.module.scss'

export type CheckStatus = 'checked' | 'half-checked' | 'unchecked'

export type KnowledgeItem = Style & {
  name?: string
  title?: ReactNode
  status?: CheckStatus
  onCheck?: () => void
  onUncheck?: () => void
  disabled?: boolean
}
export const KnowledgeItem: FC<KnowledgeItem> = (props) => {
  const {
    name,
    title,
    status,
    onCheck,
    onUncheck,
    disabled,
    className,
    style,
  } = props
  console.log(title)
  return (
    <div
      className={cn(
        'rounded cursor-pointer',
        'px-2.5 py-4 w-36 h-36 relative',
        'flex flex-col gap-4 items-center',
        styles.item,
        className,
      )}
      style={style}
      onClick={() => {
        switch (status) {
          // 仅在全选的情况下点击 才会取消选择
          case 'checked':
            return onUncheck?.()
          default:
            return onCheck?.()
        }
      }}
    >
      <Checkbox
        disabled={disabled}
        className={cn('z-10 absolute top-2.5 left-2.5')}
        checked={disabled || status !== 'unchecked'}
        indeterminate={status === 'half-checked'}
      />
      <img src={Icon} className='w-20 h-20' />
      <Typography.Paragraph
        className='break-all whitespace-pre max-w-full m-0'
        ellipsis={{ rows: 1, tooltip: name }}
      >
        {title}
      </Typography.Paragraph>
    </div>
  )
}
