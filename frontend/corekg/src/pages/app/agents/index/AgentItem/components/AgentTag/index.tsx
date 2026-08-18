import { FC, PropsWithChildren } from 'react'
import type { Agent } from 'Agent'
import { cn } from '@/utils'

type Style = {
  className?: string
  style?: React.CSSProperties
}
export type AgentTag = Style & {
  tag: Agent['tag']
}
export const AgentTag: FC<AgentTag> = (props) => {
  const { className, style, tag: type } = props
  const tagClassName =
    'z-10 absolute top-0 left-0 -translate-y-1/2 -translate-x-[1px]'
  switch (type) {
    case 'popular':
      return <Popular className={cn(tagClassName, className)} style={style} />
    case 'recommend':
      return <Recommend className={cn(tagClassName, className)} style={style} />
    default:
      return null
  }
}

const BaseTag: FC<PropsWithChildren<Style>> = (props) => {
  const { className, style, children } = props
  return (
    <div
      className={cn(
        'py-0.5 px-1.5',
        'rounded-tl-xl rounded-br-xl',
        'text-white text-[12px]',
        className,
      )}
      style={style}
    >
      {children}
    </div>
  )
}

const Recommend: FC<Style> = (props) => {
  const { className, style = {} } = props
  return (
    <BaseTag
      style={{
        background: 'linear-gradient(303.17deg, #F83600 7.51%, #F9D423 93.58%)',
        ...style,
      }}
      className={className}
    >
      最新推荐
    </BaseTag>
  )
}

const Popular: FC<Style> = (props) => {
  const { className, style = {} } = props
  return (
    <BaseTag
      style={{
        background: 'linear-gradient(90deg, #00C6FB 0%, #005BEA 100%)',
        ...style,
      }}
      className={className}
    >
      热门
    </BaseTag>
  )
}
