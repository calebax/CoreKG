import { FC } from 'react'
import { Form, Input, Modal, ModalProps } from 'antd'
import { Agent } from 'Agent'
import { useBoolean } from 'ahooks'
import { AgentAvator } from './AgentAvator'
import { AgentTypeSelect } from './AgentTypeSelect'
import styles from './index.module.scss'

export type AgentConfig = {
  agent_type: Exclude<Agent['type'], 'knowledge'>
  avatar_url: string
  description?: string
  show_name: 'string'
}

export type AgentModal = Pick<
  ModalProps,
  'open' | 'okText' | 'onCancel' | 'title'
> & {
  onOk: (val: AgentConfig) => Promise<void>
}
export const AgentModal: FC<AgentModal> = (props) => {
  const { onOk, ...rest } = props
  const [form] = Form.useForm<AgentConfig>()
  const [loading, { toggle }] = useBoolean(false)
  const agentType = Form.useWatch('agent_type', form)
  return (
    <Modal
      {...rest}
      onOk={async (e) => {
        const formValue = await form.validateFields()
        toggle()
        try {
          await onOk(formValue)
          props.onCancel?.(e)
        } finally {
          toggle()
        }
      }}
      okButtonProps={{ loading }}
      closable={!loading}
      keyboard={false}
      maskClosable={false}
      className={styles.agentModal}
      destroyOnHidden
    >
      <Form form={form} layout='vertical' preserve={false}>
        <Form.Item name='avatar_url' initialValue={'default'}>
          <AgentAvator className='block mx-auto w-15 h-15' type={agentType} />
        </Form.Item>
        <Form.Item
          name='show_name'
          label='智能体名称'
          className={styles.agentModalFormItem}
          required
          rules={[{ required: true, message: '请输入智能体名称' }]}
        >
          <Input
            maxLength={50}
            className={styles.agentModalInput}
            placeholder='请输入智能体名称'
          />
        </Form.Item>
        <Form.Item
          className={styles.agentModalFormItem}
          name='description'
          label='智能体描述'
        >
          <Input.TextArea
            maxLength={256}
            style={{ height: '32px' }}
            className={styles.agentModalInput}
            placeholder='请输入智能体描述'
          />
        </Form.Item>
        <Form.Item
          className={styles.agentModalFormItem}
          name='agent_type'
          label='智能体模式'
          initialValue={'role_play'}
        >
          <AgentTypeSelect />
        </Form.Item>
      </Form>
    </Modal>
  )
}
