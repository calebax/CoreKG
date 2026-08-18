import { FC } from 'react'
import { Modal, ModalProps, Tabs } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import {
  getForestPermSet,
  getAgentPermSet,
  AgentPermItem,
  ForestPermItem,
} from '@/api/perm'
import { Agent } from './Agent'
import { Forest } from './Forest'

export type PermModal = {
  /** 有uin会自动获取权限内容 */
  uin?: number
  onSubmit: (val: {
    chatPs: AgentPermItem[]
    forestPs: ForestPermItem[]
  }) => any
  close: () => void
} & Pick<ModalProps, 'open' | 'okText' | 'title'>
export const PermModal: FC<PermModal> = (props) => {
  const { t } = useTranslation('pages')
  const {
    uin,
    onSubmit,
    close,
    open,
    title = t('settings.setMemberPermissionScope'),
    ...rest
  } = props
  const forestPerm = useRequest(
    async () => {
      if (!open) return undefined
      const { perm_set } = await getForestPermSet({ uin })
      return (perm_set ?? []).map((item) => {
        if (item.forest.public_scope !== 'company') return item
        return { ...item, use_perm: true }
      })
    },
    { refreshDeps: [uin, open] },
  )
  const agentPerm = useRequest(
    async () => {
      if (!open) return undefined
      const { perm_set } = await getAgentPermSet({ uin })
      return (perm_set ?? []).map((item) => {
        if (item.agent.public_scope !== 'company') return item
        return { ...item, use_perm: true }
      })
    },
    { refreshDeps: [uin, open] },
  )
  const [type, setType] = useState<'forest' | 'agent'>('forest')
  const [submitLoading, { toggle }] = useBoolean(false)

  return (
    <Modal
      open={open}
      title={title}
      {...rest}
      onCancel={close}
      okButtonProps={{
        loading: submitLoading,
        disabled: !(agentPerm.data && forestPerm.data),
      }}
      onOk={async () => {
        if (!agentPerm.data || !forestPerm.data) {
          return
        }
        toggle()
        try {
          await onSubmit({ chatPs: agentPerm.data, forestPs: forestPerm.data })
          close()
        } finally {
          toggle()
        }
      }}
      width={'80vw'}
      className='top-[5vh]'
      maskClosable={false}
      keyboard={false}
    >
      <div className='w-full max-h-[70vh] overflow-hidden flex flex-col'>
        <Tabs
          activeKey={type}
          onChange={(val) => setType(val as any)}
          items={[
            { label: t('settings.knowledgeBase'), key: 'forest' },
            { label: t('settings.agent'), key: 'agent' },
          ]}
        />

        <div className='flex-1 overflow-auto'>
          <Forest
            value={forestPerm.data}
            onChange={forestPerm.mutate}
            className={cn({ hidden: type !== 'forest' })}
          />
          <Agent
            value={agentPerm.data!}
            onChange={agentPerm.mutate}
            className={cn({ hidden: type !== 'agent' })}
          />
        </div>
      </div>
    </Modal>
  )
}
