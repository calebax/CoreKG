import { FC, ReactNode } from 'react'
import { Tooltip } from 'antd'
import { cn } from '@/utils'

export type Item = {
  className?: string
  collapsed: boolean
  active: boolean
  icon: ReactNode
  iconActive?: ReactNode
  title: string
}
export const Item: FC<Item> = (props) => {
  const { collapsed, active, icon, iconActive, title, className } = props
  if (collapsed) {
    return (
      <Tooltip title={title} placement='right' arrow color='#1E1F28'>
        <div
          className={cn(
            'h-10 w-10 rounded',
            'flex items-center justify-center',
            'bg-transparent text-[#1E1F28]',
            {
              'hover:bg-[#F0F2F7]': !active,
            },
            'transition-colors duration-200',
            {
              'bg-[#E6E8F0]': active,
            },
            className,
          )}
        >
          {active && iconActive ? iconActive : icon}
        </div>
      </Tooltip>
    )
  }
  return (
    <div
      className={cn(
        'w-50 h-10 rounded pl-3',
        'flex items-center gap-3',
        'bg-transparent text-[#1E1F28]',
        {
          'hover:bg-[#fcfcfe] hover:shadow-[0_1px_3px_rgba(29,33,41,0.1)]':
            !active,
        },
        'font-medium',
        {
          'bg-[#E6E8F0]': active,
        },
      )}
    >
      {active && iconActive ? iconActive : icon}
      {title}
    </div>
  )
}
