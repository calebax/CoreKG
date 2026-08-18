import { FC, PropsWithChildren } from 'react'
import { BasicAgentInfo } from 'Agent'
import { cn } from '@/utils'

type Style = {
  className?: string
  style?: React.CSSProperties
}
export type AgentType = Style & {
  type: BasicAgentInfo['type']
}

export const AgentType: FC<AgentType> = (props) => {
  const { className, style, type } = props
  switch (type) {
    case 'role_play':
    case 'knowledge':
      // 角色型绑定知识库就是knowledge
      return <RoleType className={className} style={style} />
    case 'prompt':
      return <PromptType className={className} style={style} />
    default:
      return null
  }
}
const BaseType: FC<PropsWithChildren<Style>> = (props) => {
  const { className, style, children } = props
  return (
    <div
      className={cn(
        'w-15 h-5 px-1 py-[1.5px]',
        'flex items-center justify-center',
        'border-[0.5px] border-[#DFD8FF] rounded-sm',
        'bg-[#F9F8FF] text-[#653EC4] text-[12px]',
        className,
      )}
      style={style}
    >
      {children}
    </div>
  )
}

const RoleType: FC<Style> = (props) => {
  const { className, style } = props
  return (
    <div className={className} style={style}>
      简单应用
    </div>
  )
}

const PromptType: FC<Style> = (props) => {
  const { className, style } = props
  return (
    <BaseType className={className} style={style}>
      高级编排
    </BaseType>
  )
}
