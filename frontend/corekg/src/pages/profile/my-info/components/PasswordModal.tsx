import { useState, useEffect } from 'react'
import { Modal, Form, Input, Button, message } from 'antd'
import {
  UserOutlined,
  LockOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { UpdateAccountPassword } from '@/api/account'
import { encryptPassword } from '@/utils/crypto'
import styles from '../styles.module.scss'

interface PasswordModalProps {
  visible: boolean
  onCancel: () => void
  hasPassword: boolean
  onSuccess: () => void
  initialPassword?: string // 新增：预设的原密码
}

const PasswordModal = ({
  visible,
  onCancel,
  hasPassword,
  onSuccess,
  initialPassword,
}: PasswordModalProps) => {
  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')
  const { t } = useTranslation('pages')
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)

  // 当弹窗显示时，如果有预设密码则设置到表单中
  useEffect(() => {
    if (visible && initialPassword && hasPassword) {
      form.setFieldsValue({
        oldPassword: initialPassword,
      })
    }
  }, [visible, initialPassword, form, hasPassword])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      // 对密码进行加密
      const encryptedOldPassword = encryptPassword(values.oldPassword)
      const encryptedNewPassword = encryptPassword(values.newPassword)

      setLoading(true)
      await UpdateAccountPassword({
        old_password: encryptedOldPassword,
        new_password: encryptedNewPassword,
      })
      message.success(tM('modifySuccess', { target: t('profile.password') }))
      onSuccess()
      form.resetFields()
      onCancel()
    } catch (error) {
      console.error(
        tM('modifyFailure', { target: t('profile.password') }),
        error,
      )
    } finally {
      setLoading(false)
    }
  }

  // 自定义图标样式，使其为灰色
  const iconStyle = { color: '#9CA3AF', fontSize: '16px', marginRight: '3px' }

  return (
    <Modal
      title={
        <div className='text-lg font-[500] text-[#1D2129] mb-6'>
          {hasPassword ? t('profile.editPassword') : t('profile.setPassword')}
        </div>
      }
      open={visible}
      onCancel={onCancel}
      footer={null}
      destroyOnHidden
      width={460}
      className={`password-modal ${styles.passwordModal}`}
      centered
    >
      <Form form={form} layout='vertical' preserve={false} requiredMark={false}>
        {hasPassword ? (
          <Form.Item
            name='oldPassword'
            rules={[
              {
                required: true,
                message: t('profile.pleaseEnterOriginalPassword'),
              },
            ]}
          >
            <Input.Password
              placeholder={t('profile.pleaseEnterOriginalPassword')}
              className='h-[32px] border !shadow-none border-[#E3E6ED]  hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC]'
              readOnly={!!initialPassword} // 如果有预设密码则为只读
              iconRender={(visible) =>
                visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>
        ) : null}
        <Form.Item
          name='newPassword'
          rules={[
            { required: true, message: t('profile.pleaseEnterNewPassword') },
            {
              min: 8,
              message: t('profile.passwordLengthMin', { target: 8 }),
            },
            {
              max: 36,
              message: t('profile.passwordLengthMax', { count: 36 }),
            },
          ]}
        >
          <Input.Password
            placeholder={t('profile.pleaseEnterNewPassword')}
            className='h-[32px] border !shadow-none border-[#E3E6ED]  hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC]'
            iconRender={(visible) =>
              visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
            }
          />
        </Form.Item>
        <Form.Item
          name='confirmPassword'
          dependencies={['newPassword']}
          rules={[
            { required: true, message: t('profile.pleaseConfirmNewPassword') },
            {
              min: 8,
              message: t('profile.passwordLengthMin', { target: 8 }),
            },
            {
              max: 36,
              message: t('profile.passwordLengthMax', { count: 36 }),
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('newPassword') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(
                  new Error(t('profile.passwordsEnteredTwiceDoNotMatch')),
                )
              },
            }),
          ]}
        >
          <Input.Password
            placeholder={t('profile.pleaseConfirmNewPassword')}
            className='h-[32px] border !shadow-none border-[#E3E6ED]  hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC]'
            iconRender={(visible) =>
              visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
            }
          />
        </Form.Item>
        <Form.Item className='!mb-0 flex justify-end pt-2'>
          <Button
            className='mr-[6px] h-[30px] !py-[12px] !px-[15px] !border-0 bg-[#F5F5F5] text-[#0C1F17] hover:text-[#0C1F17]  !text-[14px]'
            onClick={onCancel}
          >
            {tC('button.cancel')}
          </Button>
          <Button
            type='primary'
            onClick={handleOk}
            loading={loading}
            className='h-[30px] !py-[12px] !px-[15px] text-[#ffffff] bg-[#0C99FF] !border-0 !text-[14px]'
          >
            {tC('button.confirm')}
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default PasswordModal
