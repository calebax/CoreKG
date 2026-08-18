import { FC } from 'react'
import { Tooltip } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import { collectAgent } from '@/api'
import { cn } from '@/utils'
import StarEmpty from './images/star-empty.svg?react'
import Star from './images/star.svg?react'

export type Favorite = {
  className?: string
  style?: React.CSSProperties
  defaultValue: boolean
  id: number
}
export const Favorite: FC<Favorite> = (props) => {
  const { className, style, defaultValue, id } = props
  const [value, { toggle }] = useBoolean(defaultValue)
  const { run } = useRequest(
    async (id: number) => {
      collectAgent(id)
    },
    { manual: true },
  )
  const childrenProps = {
    style,
    className: cn('cursor-pointer', className),
    onClick: (e: React.MouseEvent<SVGSVGElement, MouseEvent>) => {
      e.preventDefault()
      e.stopPropagation()
      toggle()
      run(id)
    },
  }
  return value ? (
    <Tooltip title='取消'>
      <Star {...childrenProps} />
    </Tooltip>
  ) : (
    <Tooltip title='收藏'>
      <StarEmpty {...childrenProps} />
    </Tooltip>
  )
}
