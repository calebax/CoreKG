import { FC } from 'react'
import { Checkbox, Form, Input, Modal } from 'antd'
import { AgentEditValue } from '..'
import { ParamOption } from './ParamOption'
import { ParamTypeSelect } from './ParamTypeSelect'

export type Param = AgentEditValue['params'][number]
const FormItem = Form.Item<Param>

export type ParamModal = {
  title: string
  open: boolean
  onClose: () => void
  value?: Param
  onSubmit: (val: Param) => void
}
export const ParamModal: FC<ParamModal> = (props) => {
  const { title, open, onClose, value, onSubmit } = props
  const [form] = Form.useForm<Param>()
  const type = Form.useWatch('input_type', form)
  return (
    <Modal
      title={title}
      open={open}
      onCancel={onClose}
      destroyOnHidden
      onOk={async () => {
        const formValue = await form.validateFields()
        onSubmit({ ...formValue, key: value?.key ?? String(performance.now()) })
        onClose()
      }}
    >
      <Form
        form={form}
        preserve={false}
        layout='vertical'
        initialValues={value}
      >
        <FormItem
          name='input_type'
          label='变量类型'
          initialValue={'text'}
          rules={[{ required: true, message: '请选择变量类型' }]}
        >
          <ParamTypeSelect />
        </FormItem>
        <FormItem
          name='input'
          label='变量名'
          rules={[{ required: true, message: '请输入变量名' }]}
        >
          <Input className='h-10' placeholder='请输入变量名' />
        </FormItem>
        <FormItem
          name='name'
          label='显示名称'
          rules={[{ required: true, message: '请输入显示名称' }]}
        >
          <Input className='h-10' placeholder='请输入显示名称' />
        </FormItem>

        {type === 'select' ? <ParamOption /> : null}

        <FormItem
          className='mt-6'
          name='is_required'
          valuePropName='checked'
          initialValue={true}
        >
          <Checkbox>必填</Checkbox>
        </FormItem>
      </Form>
    </Modal>
  )
}
