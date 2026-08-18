import { FC } from 'react'
import { Form as AntdForm, Input, Modal } from 'antd'
import { useBoolean } from 'ahooks'
import { useTranslation } from 'react-i18next'

export type EmployeeInfo = {
  user_name: string
  email: string
  phone?: string
  password?: string
}

const Form = AntdForm<EmployeeInfo>
const FormItem = Form.Item<EmployeeInfo>

export type EmployeeInfoModal = {
  value?: EmployeeInfo
  open: boolean
  onClose: () => void
  onOk: (val: EmployeeInfo) => any
  title: string
  requirePassword?: boolean
}
export const EmployeeInfoModal: FC<EmployeeInfoModal> = (props) => {
  const { t } = useTranslation('pages')
  const { value, open, onClose, onOk, title, requirePassword } = props
  const [form] = Form.useForm<EmployeeInfo>()
  const [loading, { toggle }] = useBoolean()
  return (
    <Modal
      open={open}
      onCancel={onClose}
      title={title}
      destroyOnHidden
      closable={!loading}
      maskClosable={false}
      keyboard={false}
      okButtonProps={{ loading }}
      onOk={async () => {
        const formValue = await form.validateFields()
        toggle()
        try {
          await onOk(formValue)
          onClose()
        } finally {
          toggle()
        }
      }}
    >
      <Form form={form} preserve={false} initialValues={value}>
        <FormItem
          name='user_name'
          label={t('settings.userName')}
          rules={[
            {
              required: true,
              message: t('settings.inputContent', {
                target: t('settings.userName'),
              }),
            },
          ]}
        >
          <Input />
        </FormItem>
        <FormItem
          name='email'
          label={t('settings.email')}
          rules={[
            {
              required: true,
              message: t('settings.inputContent', {
                target: t('settings.email'),
              }),
            },
            { type: 'email', message: t('settings.inputValidEmail') },
          ]}
        >
          <Input />
        </FormItem>
        <FormItem
          name='phone'
          label={t('settings.phone')}
          rules={[
            {
              pattern: /^1[3-9]\d{9}$/,
              message: t('settings.enterValidMobile'),
            },
          ]}
        >
          <Input />
        </FormItem>
        <FormItem
          name='password'
          label={t('settings.password')}
          rules={[
            {
              required: requirePassword,
              message: t('settings.inputContent', {
                target: t('settings.password'),
              }),
            },
            {
              min: 8,
              message: t('settings.passwordLengthMin', { count: 8 }),
            },
            {
              max: 36,
              message: t('settings.passwordLengthMax', { count: 36 }),
            },
          ]}
        >
          <Input.Password />
        </FormItem>
      </Form>
    </Modal>
  )
}
