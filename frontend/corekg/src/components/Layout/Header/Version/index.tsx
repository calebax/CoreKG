import { FC, ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Button, Popover } from 'antd'
import { cn } from '@/utils'
import { useVersion } from '@/utils/useVersion'
import styles from './styles.module.scss'

interface Style {
  className?: string
  style?: React.CSSProperties
}

export const VersionBtn: FC<Style & { children?: ReactNode }> = (props) => {
  const { version } = useVersion()
  if (!version) return props.children
  const { qa, agent, disk: knowledge, employee: team } = version
  return (
    <Popover
      arrow={false}
      trigger='hover'
      placement='top'
      content={
        <div
          className={cn(
            styles.popover,
            'p-4 w-54 rounded-[4px] flex flex-col gap-6',
          )}
        >
          <div className='flex flex-col gap-1 leading-[19.759px]'>
            <span className='text-[#1e1f28] text-sm'>
              今日问答次数{qa.quota}次
            </span>
            <span className='text-[#165dff] text-base font-medium'>
              {version.name}
            </span>
          </div>
          <div className='flex flex-col gap-4'>
            <Item
              title='今日问答次数'
              current={qa.used}
              limit={qa.quota + '次'}
              rate={qa.used / qa.quota}
            />
            <Item
              title='智能体数量'
              current={agent.used}
              limit={agent.quota + '个'}
              rate={agent.used / agent.quota}
            />
            <Item
              title='知识库容量'
              limit={knowledge.quota}
              rate={Number(knowledge.use_ratio)}
            />
            <Item
              title='团队成员数量'
              current={team.used}
              limit={team.quota}
              rate={team.used / team.quota}
            />
          </div>
          <Link to={'/version?type=upgrade'} target='_blank'>
            <Button
              block
              className='font-medium text-white border-0 text-base bg-[#165dff] h-auto py-1 px-0 rounded-[2px]'
            >
              升级版本
            </Button>
          </Link>
        </div>
      }
    >
      {props.children}
    </Popover>
  )
}

const Item: FC<
  Style & {
    title: string
    current?: string | number
    limit: string | number
    rate: number
  }
> = (props) => {
  const { title, current, limit, rate } = props
  const percentage = Math.min(rate * 100, 100)
  return (
    <div
      className={cn('flex flex-col gap-[6px] w-full', props.className)}
      style={props.style}
    >
      <div className='flex items-center justify-between w-full'>
        <span className='text-base text-[#4e5969] leading-[19.759px]'>
          {title}
        </span>
        <span className='text-xs leading-[19.759px]'>
          {current !== undefined ? (
            <>
              <span className='text-[#4e5969]'>{current}</span>
              <span className='text-[#86909c]'>/</span>
              <span className='text-[#86909c]'>{limit}</span>
            </>
          ) : (
            <span className='text-[#4e5969]'>{limit}</span>
          )}
        </span>
      </div>
      <div className='relative bg-[#eaf2ff] rounded-[50px] h-1.5 w-full'>
        <div
          className='absolute left-0 top-0 bottom-0 rounded-[50px] bg-[#3d7fff]'
          style={{ width: `${percentage}%` }}
        ></div>
      </div>
    </div>
  )
}
