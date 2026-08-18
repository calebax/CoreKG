import { FC, PropsWithChildren } from 'react'
import { Tooltip } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { cn } from '@/utils'

export type AdvancedItem = {
  title: string
  tooltip?: string
}
export const AdvancedItem: FC<Style & PropsWithChildren & AdvancedItem> = (
  props,
) => {
  const { title, tooltip } = props
  return (
    <div
      className={cn('flex flex-col gap-6', props.className)}
      style={props.style}
    >
      <span
        className={cn(
          'flex items-center gap-2 text-base font-medium',
          props.className,
        )}
        style={props.style}
      >
        <div className='bg-[#165DFF] rounded-full h-3 w-[3px]'></div>
        {title}
        {tooltip ? (
          <Tooltip title={tooltip}>
            <InfoCircleOutlined className='text-[#C9CDD4]' />
          </Tooltip>
        ) : null}
      </span>
      {props.children}
    </div>
  )
}
