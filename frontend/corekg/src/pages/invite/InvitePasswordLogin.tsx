import { useState } from 'react'
import { Button, Input, Form } from 'antd'
import { useTranslation } from 'react-i18next'
import { Agreement } from '@/components/agreement'
import ForgotPasswordModal from '@/pages/other/login/components/ForgotPasswordModal'
import { useDeployConfig } from '@/utils/useDeployConfig'

interface InvitePasswordLoginProps {
  onLogin: (username: string, password: string) => Promise<void>
  loading?: boolean
}

interface LoginValues {
  username: string
  password: string
}

const InvitePasswordLogin: React.FC<InvitePasswordLoginProps> = ({
  onLogin,
  loading = false,
}) => {
  const { t } = useTranslation('pages')
  const { version } = useDeployConfig()
  const [form] = Form.useForm()
  const [showForgotModal, setShowForgotModal] = useState(false)

  // custom 环境下不展示忘记密码功能
  const showForgotPassword = version !== 'custom'

  const handlePasswordLogin = async (values: LoginValues) => {
    try {
      const { username, password } = values
      await onLogin(username, password)
    } catch (error) {
      console.log('登录失败:', error)
    }
  }

  return (
    <div className='w-full min-h-[15rem] flex flex-col'>
      <Form
        className='w-full'
        form={form}
        name='invite-password-login'
        onFinish={handlePasswordLogin}
        autoComplete='off'
        layout='vertical'
        initialValues={{
          username: '',
          password: '',
        }}
      >
        <Form.Item
          label={t('other.phoneNumberOrEmail')}
          name='username'
          rules={[
            {
              required: true,
              message: t('other.pleaseEnterPhoneNumberOrEmail'),
            },
          ]}
        >
          <Input
            placeholder={t('other.pleaseEnterPhoneNumberOrEmail')}
            size='large'
          />
        </Form.Item>

        <Form.Item
          label={t('other.password')}
          name='password'
          rules={[
            { required: true, message: t('other.pleaseEnterPassword') },
            { min: 8, message: t('other.passwordLengthMin', { count: 8 }) },
            { max: 36, message: t('other.passwordLengthMax', { count: 36 }) },
          ]}
        >
          <Input.Password
            placeholder={t('other.pleaseEnterPassword')}
            size='large'
          />
        </Form.Item>

        {showForgotPassword && (
          <div className='flex justify-end'>
            <span
              className='text-sm font-medium text-[#0C99FF] cursor-pointer hover:opacity-80'
              onClick={() => setShowForgotModal(true)}
            >
              忘记密码
            </span>
          </div>
        )}

        <Form.Item className='mb-2'>
          <Button
            type='primary'
            htmlType='submit'
            size='large'
            loading={loading}
            className='w-full mt-4'
          >
            {t('other.login')}
          </Button>
        </Form.Item>
      </Form>
      <Agreement className='mx-auto' />
      {/* 提示无账号的用户可扫码快速注册 */}
      <div
        className='mt-2 text-sm text-gray-500 text-center'
        aria-live='polite'
      >
        {t('other.noAccountScanQrCodeToRegisterImmediately')}
      </div>

      {/* 忘记密码弹窗：custom 环境下不展示 */}
      {showForgotPassword && (
        <ForgotPasswordModal
          visible={showForgotModal}
          onCancel={() => setShowForgotModal(false)}
        />
      )}
    </div>
  )
}

export default InvitePasswordLogin
