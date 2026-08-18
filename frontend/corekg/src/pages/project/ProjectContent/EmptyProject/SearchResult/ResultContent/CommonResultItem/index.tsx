import { FC, PropsWithChildren, ReactNode, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Avatar, Typography } from 'antd'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { ResultType } from '../..'
import { ResultIcon } from '../../ResultIcon'

export type CommonResultItem = Style &
  PropsWithChildren & {
    type?: ResultType
    icon?: ReactNode
    title_html: string
    titleExtra?: ReactNode
    creator: string
    creatorAvatar: string
    time: string
    to: string
  }
/** 不同搜索结果可复用的组件 */
export const CommonResultItem: FC<CommonResultItem> = (props) => {
  const {
    children,
    icon,
    type,
    title_html,
    titleExtra,
    creator,
    creatorAvatar,
    time,
    to,
    className,
    style,
  } = props
  const { t } = useTranslation('common')
  const timeLabel = useMemo(() => {
    const now = dayjs()
    const target = dayjs(time)
    const diffSec = now.diff(target, 'second')
    const diffMin = now.diff(target, 'minute')
    const diffHour = now.diff(target, 'hour')
    const diffDay = now.diff(target, 'day')

    if (diffSec < 60) {
      return t('time.secondsAgo', { count: diffSec })
    }
    if (diffMin < 60) {
      return t('time.minutesAgo', { count: diffMin })
    }
    if (diffHour < 24) {
      return t('time.hoursAgo', { count: diffHour })
    }
    if (diffDay <= 2) {
      return t('time.daysAgo', { count: diffDay })
    }

    return target.format('YYYY.MM.DD HH:mm:ss')
  }, [t, time])
  return (
    <Link
      to={to}
      target='_blank'
      className={cn('flex gap-3', className)}
      style={style}
    >
      {type ? <ResultIcon type={type} className='w-10 h-10' /> : icon}
      <div className='flex-1 flex flex-col gap-2.5'>
        <div className='flex items-center gap-2'>
          <div
            className='text-black text-base line-clamp-1 whitespace-nowrap'
            dangerouslySetInnerHTML={{ __html: title_html }}
          ></div>
          {titleExtra}
        </div>
        <div className='flex items-center text-[#919497] text-xs overflow-hidden'>
          {creatorAvatar ? (
            <Avatar src={creatorAvatar} className='w-4.5 h-4.5 mr-1' />
          ) : null}
          {creator && (
            <span className='whitespace-nowrap max-w-[60%] overflow-hidden text-ellipsis mr-2.5'>
              {creator}
            </span>
          )}
          {creator && <span className=' mr-2.5'>·</span>}
          {timeLabel}
        </div>
        <Typography.Paragraph ellipsis={{ rows: 3 }}>
          {children}
        </Typography.Paragraph>
      </div>
    </Link>
  )
}
