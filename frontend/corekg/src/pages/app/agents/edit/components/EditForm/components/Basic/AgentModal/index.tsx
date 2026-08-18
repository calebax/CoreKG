import { FC } from 'react'
import { Form, Input, Modal } from 'antd'
import { Agent } from 'Agent'
import { FormItem } from '../../..'
import { AgentAvator } from './AgentAvator'

type AgentConfig = Pick<Agent, 'avatar' | 'title' | 'description'>

export type AgentModal = {
  value: AgentConfig
  onSubmit: (val: AgentConfig) => void
  open: boolean
  onClose: () => void
}
export const AgentModal: FC<AgentModal> = (props) => {
  const { value, onSubmit, open, onClose } = props
  const [form] = Form.useForm<AgentConfig>()
  return (
    <Modal
      open={open}
      onCancel={onClose}
      onOk={async () => {
        const formValue = await form.validateFields()
        onSubmit(formValue)
        onClose()
      }}
      keyboard={false}
      maskClosable={false}
      destroyOnHidden
    >
      <Form
        form={form}
        layout='vertical'
        preserve={false}
        initialValues={value}
      >
        <FormItem name='avatar'>
          <AgentAvator className='block mx-auto' />
        </FormItem>
        <FormItem
          name='title'
          label='智能体名称'
          required
          rules={[{ required: true, message: '请输入智能体名称' }]}
        >
          <Input maxLength={50} showCount />
        </FormItem>
        <FormItem name='description' label='智能体描述'>
          <Input.TextArea maxLength={256} showCount />
        </FormItem>
      </Form>
    </Modal>
  )
}
