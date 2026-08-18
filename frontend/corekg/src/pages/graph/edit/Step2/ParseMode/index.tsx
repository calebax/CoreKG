import { FC, PropsWithChildren } from 'react'
import { Radio } from 'antd'
import { cn } from '@/utils'
import { updateGraph } from '@/api/graph'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import Icon from './icon.svg?react'

export const ParseMode: FC<Style> = (props) => {
  const { className, style } = props
  const { data, updateBaseInfo } = useGraphInfo()
  const parse_mode = data?.parse_mode ?? 'auto'

  return (
    <div
      className={cn(
        'px-4 py-2 bg-[#F8F9FD] rounded flex flex-col gap-2',
        className,
      )}
      style={style}
    >
      <span className='flex items-center gap-2'>
        <Icon />
        <span className='text-base font-medium'>模型策略</span>
      </span>
      <div className='flex items-center gap-4'>
        <ParseModeBtn
          active={parse_mode === 'rule'}
          onClick={() => {
            updateBaseInfo({ parse_mode: 'rule' })
            updateGraph({
              graph_id: data!.id,
              parse_mode: 'rule',
            })
          }}
        >
          标准模式
        </ParseModeBtn>
        <ParseModeBtn
          active={parse_mode === 'auto'}
          onClick={() => {
            updateBaseInfo({ parse_mode: 'auto' })
            updateGraph({
              graph_id: data!.id,
              parse_mode: 'auto',
            })
          }}
        >
          深度洞察模式
        </ParseModeBtn>
      </div>

      <span className='text-description'>
        {parse_mode === 'auto'
          ? 'AI自动推断潜在关联，构建更丰富、更深入的知识图谱。'
          : '严格遵循现有规则，不启用模型自动扩展，确保图谱结果高度可靠。'}
      </span>
    </div>
  )
}

const ParseModeBtn: FC<
  PropsWithChildren & { onClick: () => void; active?: boolean }
> = (props) => {
  const { children, onClick, active } = props
  return (
    <div
      className={cn(
        'px-2.5 py-3 bg-[#FFFFFF] border border-[#0C99FF] rounded-xl cursor-pointer',
        active && '[&_.ant-radio-wrapper]:text-[#0C99FF]',
      )}
      onClick={onClick}
    >
      <Radio checked={active}>{children}</Radio>
    </div>
  )
}
