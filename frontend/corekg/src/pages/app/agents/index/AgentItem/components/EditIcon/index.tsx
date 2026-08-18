import { FC } from 'react'
import { App, Button, Popover } from 'antd'
import { Agent } from 'Agent'
import { useRequest } from 'ahooks'
import { deleteAgent } from '@/api'
import { cn } from '@/utils'
import { getAgentUrl } from '@/pages/app/agents/utils/getAgentUrl'
import { useVersion } from '@/utils/useVersion'
import { useAgentList } from '../../../AgentContext'
import Icon from './images/icon.svg?react'
import styles from './styles.module.scss'

export type EditIcon = {
  className?: string
  style?: React.CSSProperties
  id: number
  type: Agent['type']
}
export const EditIcon: FC<EditIcon> = (props) => {
  const { className, style, id, type } = props
  const { refresh } = useVersion()
  const { message } = App.useApp()
  const agentList = useAgentList()
  const url = getAgentUrl(id, type, true)
  const copy = useRequest(
    async () => {
      // 复制应用
      message.success('操作成功')
      agentList.refresh()
    },
    { manual: true },
  )

  const del = useRequest(
    async () => {
      await deleteAgent(id)
      message.success('操作成功')
      agentList.refresh()
      refresh()
    },
    { manual: true },
  )
  return (
    <Popover
      placement='bottom'
      content={
        <div className={cn('p-1.5', 'flex flex-col gap-0.5')}>
          <Link to={url} target='_blank'>
            <Button type='text' className='w-21 h-7'>
              配置应用
            </Button>
          </Link>
          {/* <Button
            type='text'
            className='w-21 h-7'
            loading={copy.loading}
            onClick={copy.run}
          >
            复制应用
          </Button> */}
          <Button
            type='text'
            className='w-21 h-7 text-[#FF0200]'
            loading={del.loading}
            onClick={del.run}
          >
            删除应用
          </Button>
        </div>
      }
    >
      <div
        className={cn(
          'rounded-full w-6 h-6',
          'flex items-center justify-center',
          styles.icon,
          className,
        )}
        style={style}
        onClick={(e) => {
          e.preventDefault()
        }}
      >
        <Icon />
      </div>
    </Popover>
  )
}
