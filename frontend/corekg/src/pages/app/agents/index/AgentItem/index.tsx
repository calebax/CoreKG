import { FC, useMemo, useState } from 'react'
import { Avatar, Typography, Button, App, Tooltip, Spin } from 'antd'
import {
  SettingOutlined,
  DeleteOutlined,
  LoadingOutlined,
} from '@ant-design/icons'
import { BasicAgentInfo } from 'Agent'
import { useRequest } from 'ahooks'
import { match, P } from 'ts-pattern'
import { deleteAgent, syncCoze } from '@/api'
import { cn } from '@/utils'
import { ItemCard } from '@/components/ItemCard'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useVersion } from '@/utils/useVersion'
import { getAgentUrl } from '../../utils/getAgentUrl'
import { useAgentList } from '../AgentContext'
import { getDefaultAvatar } from '../getDefaultAvatar'
import { AgentStatus } from './components/AgentStatus'
import { AgentTag } from './components/AgentTag'
import { AgentType } from './components/AgentType'
import { Favorite } from './components/Favorite'
import Configuration from './images/configuration.svg?react'
import Delete from './images/delete.svg?react'
import Coze from './images/send.svg?react'
import styles from './styles.module.scss'

export const AgentItem: FC<BasicAgentInfo & Style> = (props) => {
  const { className, style } = props
  const {
    id,
    avatar,
    title,
    description: desc,
    type,
    favorite,
    isAdmin,
    source,
    status,
    tag,
    is_synced,
    coze_workflow_id,
    coze_space_id,
  } = props
  const { version } = useDeployConfig()
  const { message } = App.useApp()
  const { refresh } = useVersion()
  const agentList = useAgentList()
  const [syncLoading, setSyncLoading] = useState(false)

  const url = getAgentUrl(id, type, source === 'custom' && status === 'draft', {
    coze_workflow_id,
    coze_space_id,
  })

  // 删除功能（从EditIcon组件移植）
  const handleDelete = useRequest(
    async () => {
      await deleteAgent(id)
      message.success('操作成功')
      agentList.refresh()
      refresh()
    },
    { manual: true },
  )

  // 配置功能
  const handleConfig = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    const configUrl = getAgentUrl(id, type, true)
    window.open(configUrl, '_blank')
  }

  // 同步coze功能
  const handleSyncCoze = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()

    setSyncLoading(true)
    try {
      // 同步coze接口调用
      const response = await syncCoze({
        agent_id: id,
      })
      if (response.code === 0) {
        message.success('同步coze成功')
        agentList.refresh()
        refresh()
      }
    } catch (error) {
      // message.error('同步coze失败，请稍后重试')
    } finally {
      setSyncLoading(false)
    }
  }

  // 显示操作按钮的条件（基于现有EditIcon逻辑）
  const shouldShowActionButtons = useMemo(() => {
    return isAdmin
  }, [isAdmin])

  return (
    <ItemCard
      className={className}
      style={style}
      onClick={() => window.open(url, '_blank')}
      avatar={avatar === 'default' ? getDefaultAvatar(type) : avatar}
      title={title}
      desc={desc}
      hiddenOperator={!shouldShowActionButtons || version === 'international'}
      extra={
        // 标签
        <div className='flex gap-[10px] items-center'>
          <div className='bg-[#f9f8ff] border border-[#dfd8ff] rounded-full px-1.5 py-0 h-5 flex items-center justify-center'>
            <span className='text-[#653ec4] text-[12px] font-normal'>
              {match(type)
                .with('prompt', () => '指令型-高级编排')
                .with(
                  P.union('role_play', 'knowledge'),
                  () => '指令型-简单应用',
                )
                .with('workflow', () => '工作流')
                .exhaustive()}
            </span>
          </div>

          {match({ source, status })
            .with({ source: 'custom' }, (val) => {
              const { status } = val
              return (
                <div className='bg-[#f5f9ff] border border-[#cce1ff] rounded-full px-1.5 py-0 h-5 flex items-center justify-center'>
                  <span
                    className={cn('text-[12px] font-normal text-[#3473ec]', {
                      'text-[#898ca1]': status === 'draft',
                    })}
                  >
                    {match(status)
                      .with('published', () => '已发布')
                      .with('draft', () => '草稿')
                      .exhaustive()}
                  </span>
                </div>
              )
            })
            .otherwise(() => null)}

          {/* System Agent Tag */}
          {source === 'system' && tag && (
            <div className='bg-[#f5f9ff] border border-[#cce1ff] rounded-[3px] px-1.5 py-0 h-5 flex items-center justify-center'>
              <span className='text-[#3473ec] text-[12px] font-normal'>
                {tag}
              </span>
            </div>
          )}
        </div>
      }
      operators={{
        items: [
          {
            key: 'coze',
            label: (
              <Tooltip title='同步后可直接在 Coze 使用'>
                <div
                  className={cn(
                    'flex gap-1 items-center transition-opacity duration-150',
                    is_synced === true
                      ? 'opacity-50 cursor-not-allowed'
                      : syncLoading
                        ? 'opacity-60 cursor-not-allowed'
                        : 'cursor-pointer hover:opacity-80',
                  )}
                  onClick={
                    is_synced === true || syncLoading
                      ? undefined
                      : handleSyncCoze
                  }
                >
                  {syncLoading ? (
                    <LoadingOutlined
                      className='text-[18px] text-[#919497]'
                      spin
                    />
                  ) : (
                    <Coze
                      className={cn(
                        'text-[18px]',
                        is_synced === true
                          ? 'text-[#C9CDD4]'
                          : 'text-[#2D2D2D]',
                      )}
                    />
                  )}
                  <span
                    className={cn(
                      'text-[12px] leading-[24px] font-[500] whitespace-nowrap overflow-hidden text-ellipsis',
                      is_synced === true
                        ? 'text-[#C9CDD4] cursor-not-allowed'
                        : 'text-[#2D2D2D]',
                    )}
                  >
                    {syncLoading ? '同步中...' : '同步至Coze'}
                  </span>
                </div>
              </Tooltip>
            ),
          },
          {
            key: 'config',
            label: (
              <Tooltip title='修改配置规则'>
                <div
                  className='flex gap-1 items-center cursor-pointer hover:opacity-80'
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    handleConfig(e)
                  }}
                >
                  <Configuration className='text-[18px] text-[#2D2D2D]' />
                  <span className='text-[12px] font-[500] text-[#2D2D2D] leading-[24px] whitespace-nowrap overflow-hidden text-ellipsis'>
                    配置
                  </span>
                </div>
              </Tooltip>
            ),
          },
          {
            key: 'del',
            label: (
              <Tooltip title='删除该智能体'>
                <div
                  className='flex gap-1 items-center cursor-pointer hover:opacity-80'
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    handleDelete.run()
                  }}
                >
                  <Delete className='text-[18px] text-[#2D2D2D]' />
                  <span className='text-[12px] font-[500] text-[#2D2D2D] leading-[24px] whitespace-nowrap overflow-hidden text-ellipsis'>
                    删除
                  </span>
                </div>
              </Tooltip>
            ),
          },
        ],
        onClick: (e) => e.domEvent.stopPropagation(),
      }}
    />
  )
}
