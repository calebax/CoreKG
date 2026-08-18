import { forwardRef, PropsWithChildren, ReactNode } from 'react'
import { Dropdown, DropDownProps, Typography } from 'antd'
import { cn } from '@/utils'
import Op from './op.svg'
import './styles.css'

export type ItemCard = Style &
  PropsWithChildren & {
    avatar?: string
    title?: ReactNode
    desc?: ReactNode
    operators?: DropDownProps['menu']
    extra?: ReactNode
    hiddenOperator?: boolean
    onClick?: () => void
  }
/** 一级页面的展示卡片 */
export const ItemCard = forwardRef<HTMLDivElement, ItemCard>((props, ref) => {
  const {
    avatar,
    title,
    desc,
    operators,
    extra,
    hiddenOperator,
    onClick,
    children,
    className,
    style,
  } = props
  return (
    <div
      className={cn(
        'global-itemcard',
        'overflow-visible',
        {
          'cursor-pointer': onClick,
        },
        className,
      )}
      onClick={onClick}
      style={style}
      ref={ref}
    >
      {children}
      <div className={cn('cardinner', 'flex flex-col overflow-visible')}>
        <div className='flex-1 overflow-visible flex gap-3'>
          <div
            className={'avatar flex-none'}
            style={{ backgroundImage: `url(${avatar})` }}
          />
          <div className='pt-[10px] flex-1 overflow-visible flex flex-col gap-2'>
            {typeof title === 'string' ? (
              <Typography.Paragraph
                className={cn('title', 'm-0 w-40 max-w-40 overflow-hidden')}
                style={{ lineHeight: '26px' }}
                ellipsis={{ rows: 1 }}
              >
                {title}
              </Typography.Paragraph>
            ) : (
              title
            )}
            <div
              className={cn(
                'extra',
                'flex-1 overflow-visible relative whitespace-pre',
              )}
            >
              {extra}
            </div>
          </div>
        </div>
        {typeof desc === 'string' ? (
          <Typography.Paragraph
            className={cn('desc', 'm-0 overflow-hidden')}
            // 列表卡片仅需两行省略，不需要悬停展示全文
            ellipsis={{ rows: 2, tooltip: false }}
          >
            {desc}
          </Typography.Paragraph>
        ) : (
          desc
        )}
        <Dropdown placement='bottomRight' menu={operators}>
          <div className='op rounded'>
            <img
              src={Op}
              className={cn({
                hidden: hiddenOperator || !operators,
              })}
              onClick={(e) => {
                e.stopPropagation()
                e.preventDefault()
              }}
            />
          </div>
        </Dropdown>
      </div>
    </div>
  )
})
