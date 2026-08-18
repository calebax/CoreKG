import { FC, PropsWithChildren } from 'react'
import { Agent } from 'Agent'
import { cn } from '@/utils'

type Style = {
  className?: string
  style?: React.CSSProperties
}
export type AgentStatus = Style & {
  status: Agent['status']
}

export const AgentStatus: FC<AgentStatus> = (props) => {
  const { className, style, status } = props
  switch (status) {
    case 'published':
      return <Published className={className} style={style} />
    case 'draft':
      return <Draft className={className} style={style} />
    default:
      return null
  }
}
const BaseType: FC<PropsWithChildren<Style>> = (props) => {
  const { className, style, children } = props
  return (
    <div
      className={cn(
        'w-14 h-5 px-1 py-[1.5px]',
        'flex items-center justify-center',
        'border-[0.5px] border-[#00000033] rounded-sm',
        'bg-[#F0F2F7] text-[#165DFF] text-[12px]',
        className,
      )}
      style={style}
    >
      {children}
    </div>
  )
}

const Published: FC<Style> = (props) => {
  const { className, style } = props
  return (
    <BaseType
      className={cn(
        'text-[#3473EC] border-[#CCE1FF] bg-[#F5F9FF] font-normal fontFamily-PingFangSC',
        className,
      )}
      style={style}
    >
      已发布
    </BaseType>
  )
}

const Draft: FC<Style> = (props) => {
  const { className, style } = props
  return (
    <BaseType
      className={cn(
        'text-[#7F8295] border-[#E6E8F0] bg-[#FCFCFE] font-normal fontFamily-PingFangSC',
        className,
      )}
      style={style}
    >
      草稿
    </BaseType>
  )
}
