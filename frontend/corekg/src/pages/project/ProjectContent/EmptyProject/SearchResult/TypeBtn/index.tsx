import { FC } from 'react'
import { Typography } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { ResultType } from '..'
import { ResultIcon } from '../ResultIcon'

export const TypeBtn: FC<{
  onClick: () => void
  type: ResultType | 'all'
  active: boolean
  total: number
}> = (props) => {
  const { onClick, type, active, total } = props
  const { t } = useTranslation('pages')
  const label = t(`project.searchResultType.${type}`)
  return (
    <div
      className={cn(
        'border border-[#d5d8db] rounded-2xl cursor-pointer overflow-hidden',
        'flex-shrink-0 p-2.5 w-33 flex flex-col items-start',
        { 'bg-[#F7F7F7]': active },
      )}
      onClick={onClick}
    >
      <ResultIcon type={type} />
      <Typography.Paragraph
        className='m-0 mt-2.5 mb-1 w-full'
        ellipsis={{ rows: 1, tooltip: label }}
      >
        {label}
      </Typography.Paragraph>
      <span className='text-[#919497]'>{total}</span>
    </div>
  )
}
