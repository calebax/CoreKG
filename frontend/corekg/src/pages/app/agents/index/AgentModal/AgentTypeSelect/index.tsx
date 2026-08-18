import { FC } from 'react'
import { Radio } from 'antd'
import { cn } from '@/utils'
import { AgentConfig } from '..'
import styles from './styles.module.scss'

export type AgentTypeSelect = {
  className?: string
  style?: React.CSSProperties
  value?: AgentConfig['agent_type']
  onChange?: (val: AgentConfig['agent_type']) => void
}
export const AgentTypeSelect: FC<AgentTypeSelect> = (props) => {
  const { value, onChange, className, style } = props
  const types: {
    name: string
    type: AgentConfig['agent_type']
    desc: string
  }[] = [
    {
      name: '简单应用',
      type: 'role_play',
      desc: '适合需要人性化对话的场景，如法律助手',
    },
    {
      name: '高级编排',
      type: 'prompt',
      desc: '适合自动化任务场景，如日报生成器',
    },
    {
      name: '工作流',
      type: 'workflow',
      desc: '设计工作流，创建逻辑灵活的智能体',
    },
  ]
  return (
    <div className={cn('flex  gap-[15px]', className)} style={style}>
      {types.map((item) => {
        const { name, type, desc } = item
        return (
          <div
            className={cn(
              'flex=1 px-[10px] py-1.5 rounded-[8px]',
              'flex flex-col gap-[5px] ',
              'border border-[#E3E6ED]',
            )}
          >
            <Radio
              className={styles.radio}
              checked={type === value}
              onChange={() => onChange?.(type)}
            >
              <span className=' text-base text-[#1D2129]'>{name}</span>
              <br />
              <span className='text-[#6E757F]'>{desc}</span>
            </Radio>
          </div>
        )
      })}
    </div>
  )
}
