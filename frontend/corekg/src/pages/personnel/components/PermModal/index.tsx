import { FC, useState, useCallback } from 'react'
import { App, Modal, ModalProps, Tabs } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { AgentPermItem, ForestPermItem } from '@/api/perm'
import { Forest } from './Forest'

export type PermModal = {
  open?: boolean
  /** 有uin会自动获取权限内容 */
  uin?: number
  onSubmit: (val: {
    chatPs: AgentPermItem[]
    forestPs: ForestPermItem[]
  }) => Promise<void>
  onCancel: () => void
  /** 提交成功后的回调，如果提供则调用此回调而不是 onCancel */
  onSuccess?: () => void
  cancelText?: string
} & Pick<ModalProps, 'open' | 'okText' | 'title'>
export const PermModal: FC<PermModal> = (props) => {
  const { t } = useTranslation('pages')
  const { message } = App.useApp()
  const {
    uin,
    onSubmit,
    onCancel,
    onSuccess,
    cancelText,
    open,
    title = t('settings.setMemberPermissionScope'),
    ...rest
  } = props

  // 已变更的权限
  const [changedForestPerms, setChangedForestPerms] = useState<
    Record<number, ForestPermItem>
  >({})

  const [type, setType] = useState<'forest' | 'agent'>('forest')

  // 处理 Forest 权限变更
  const handleForestPermChange = useCallback(
    (id: number, perm: ForestPermItem) => {
      setChangedForestPerms((prev) => ({
        ...prev,
        [id]: perm,
      }))
    },
    [],
  )

  const { loading: submitLoading, run: submit } = useRequest(
    async () => {
      // 提交增量数据
      const forestPs: ForestPermItem[] = Object.values(changedForestPerms)
      const chatPs: AgentPermItem[] = [] // 智能体权限已隐藏，始终传递空数组

      await onSubmit({ chatPs, forestPs })
      message.success('操作成功')
      // 如果提供了 onSuccess 回调，则调用它，否则调用 onCancel
      if (onSuccess) {
        onSuccess()
      } else {
        onCancel()
      }
    },
    { manual: true },
  )

  return (
    <Modal
      open={open}
      title={title}
      {...rest}
      onCancel={onCancel}
      cancelText={cancelText}
      cancelButtonProps={{
        disabled: submitLoading,
      }}
      okButtonProps={{
        loading: submitLoading,
      }}
      onOk={submit}
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
            // 智能体tab已隐藏
          ]}
        />

        <div className='flex-1 overflow-auto'>
          <Forest
            uin={uin}
            value={changedForestPerms}
            onChange={handleForestPermChange}
          />
        </div>
      </div>
    </Modal>
  )
}
