import { useState, useEffect } from 'react'
import { Modal, Form, Input, Button, message } from 'antd'
import {
  UserOutlined,
  LockOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { updatePasswordWithRefreshToken } from '@/api/account'
import { encryptPassword } from '@/utils/crypto'

interface PasswordModalProps {
  visible: boolean
  onCancel: () => void
  onSuccess: () => void
  initialPassword?: string // 预设的原密码
  userId: number // 用户ID
  refreshToken: string // 刷新令牌
}

const PasswordModal = ({
  visible,
  onCancel,
  onSuccess,
  initialPassword,
  userId,
  refreshToken,
}: PasswordModalProps) => {
  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')
  const { t } = useTranslation('pages')
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)

  // 当弹窗显示时，如果有预设密码则设置到表单中
  useEffect(() => {
    if (visible && initialPassword) {
      form.setFieldsValue({
        oldPassword: initialPassword,
      })
    }
  }, [visible, initialPassword, form])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      // 对密码进行加密
      const encryptedOldPassword = encryptPassword(values.oldPassword)
      const encryptedNewPassword = encryptPassword(values.newPassword)

      setLoading(true)
      await updatePasswordWithRefreshToken({
        old_password: encryptedOldPassword,
        new_password: encryptedNewPassword,
        user_id: userId,
        refresh_token: refreshToken,
      })
      message.success(tM('modifySuccess', { target: t('profile.password') }))
      form.resetFields()
      onCancel()
      onSuccess()
    } catch (error) {
      console.log(
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
          {t('profile.editPassword')}
        </div>
      }
      open={visible}
      onCancel={onCancel}
      footer={null}
      destroyOnHidden
      width={460}
      className='password-modal'
      centered
    >
      <Form form={form} layout='vertical' preserve={false} requiredMark={false}>
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
            className='h-10 border border-gray-300 hover:border-[#0C99FF] focus:border-[#2A69E3] !rounded'
            prefix={<UserOutlined style={iconStyle} />}
            readOnly={!!initialPassword} // 如果有预设密码则为只读
            iconRender={(visible) =>
              visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
            }
          />
        </Form.Item>
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
            className='h-10 border border-gray-300 hover:border-[#0C99FF] focus:border-[#2A69E3] !rounded'
            prefix={<LockOutlined style={iconStyle} />}
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
            className='h-10 border border-gray-300 hover:border-[#0C99FF] focus:border-[#2A69E3] !rounded'
            prefix={<LockOutlined style={iconStyle} />}
            iconRender={(visible) =>
              visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
            }
          />
        </Form.Item>
        <Form.Item className='!mb-0 flex justify-end pt-2'>
          <Button
            className='mr-4 h-10 !py-4 !px-4 border hover:text-[#0C99FF] border-gray-300 !text-base'
            onClick={onCancel}
          >
            {tC('button.cancel')}
          </Button>
          <Button
            type='primary'
            onClick={handleOk}
            loading={loading}
            className='h-10 !py-4 !px-4 bg-[#0C99FF] !text-base'
          >
            {tC('button.confirm')}
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default PasswordModal