import { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from 'antd'
import { useBoolean } from 'ahooks'
import { createAgent } from '@/api'
import { cn } from '@/utils'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { useVersion } from '@/utils/useVersion'
import { getAgentUrl } from '../../utils/getAgentUrl'
import { useAgentList } from '../AgentContext'
import { AgentModal } from '../AgentModal'
import AddIcon from '../images/add.svg?react'
import superStyles from '../styles.module.scss'

export type AddAgentBtn = {
  className?: string
}
export const AddAgentBtn: FC<AddAgentBtn> = (props) => {
  const { className } = props
  const [open, { toggle }] = useBoolean(false)
  const navigate = useNavigate()
  const agentList = useAgentList()
  const { version, refresh: refreshVersion } = useVersion()
  const { show: showQuotaLimitModal } = useQuotaLimitModal()
  const isQuotaLimited = version && version.agent.used >= version.agent.quota

  return (
    <>
      <Button
        onClick={() => {
          if (isQuotaLimited) {
            showQuotaLimitModal({ type: 'agent' })
            return
          }
          toggle()
        }}
        icon={<AddIcon className={superStyles.createBtnIcon} />}
        className={cn(className, 'bg-white', superStyles.createBtn, {
          'opacity-50': isQuotaLimited,
        })}
      >
        新建智能体
      </Button>
      <AgentModal
        title='创建智能体'
        open={open}
        onCancel={toggle}
        onOk={async (val) => {
          const { agent_type } = val
          const { id } = (await createAgent(val)) as any
          agentList.refresh()
          refreshVersion()
          navigate(getAgentUrl(id, agent_type, true))
        }}
      />
    </>
  )
}
