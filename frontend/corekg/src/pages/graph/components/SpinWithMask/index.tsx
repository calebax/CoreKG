import { FC } from 'react'
import { Spin } from 'antd'
import { cn } from '@/utils'

/**
 * 具有蒙层的spin
 * @example
 * ```ts
 * <div className='relative'>
 *   <SpinWithMask show />
 * <div>
 * ```
 */
export const SpinWithMask: FC<Style & { show?: boolean }> = (props) => {
  const { show, className, style } = props
  return (
    <div
      className={cn(
        ' absolute bg-[#ffffffaa] inset-0 z-20',
        { hidden: !show },
        className,
      )}
      style={style}
    >
      <Spin
        spinning
        className=' absolute top-1/2 left-1/2 translate-x-1/2 translate-y-1/2'
      ></Spin>
    </div>
  )
}
