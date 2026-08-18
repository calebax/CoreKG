import { useState, useEffect, useMemo, useCallback } from 'react'
import { Modal, Form, Input, Button, message } from 'antd'
import {
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import {
  forgotPassword,
  sendVerifyCodeForForgot,
} from '@/api/account'
import { encryptPassword } from '@/utils/crypto'
import CloseIcon from '@/assets/icons/close.svg?react'

interface ForgotPasswordModalProps {
  visible: boolean
  onCancel: () => void
}

// 公共样式常量
const INPUT_CLASS_NAME =
  'h-[32px] border !shadow-none border-[#E3E6ED] hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC] placeholder:text-[14px]'

// 密码校验规则
const getPasswordRules = (t: any) => [
  { required: true, message: '请输入新密码' },
  {
    min: 8,
    message: t('other.passwordLengthMin', { count: 8 }),
  },
  {
    max: 36,
    message: t('other.passwordLengthMax', { count: 36 }),
  },
]

const ForgotPasswordModal = ({
  visible,
  onCancel,
}: ForgotPasswordModalProps) => {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [codeLoading, setCodeLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)

  // 监听表单变化，用于控制按钮状态
  const [submittable, setSubmittable] = useState(false)
  const values = Form.useWatch([], form)

  useEffect(() => {
    // 检查四个输入框是否都有值
    const hasAllFields =
      values?.phone?.trim() &&
      values?.code?.trim() &&
      values?.password?.trim() &&
      values?.confirmPassword?.trim()
    setSubmittable(!!hasAllFields)
  }, [values])

  // 倒计时逻辑
  useEffect(() => {
    if (countdown <= 0) return

    const timer = setTimeout(() => {
      setCountdown((prev) => prev - 1)
    }, 1000)

    return () => clearTimeout(timer)
  }, [countdown])

  // 重置表单和状态
  const resetForm = useCallback(() => {
    form.resetFields()
    setCountdown(0)
  }, [form])

  // 弹窗打开时重置表单
  useEffect(() => {
    if (visible) {
      resetForm()
    }
  }, [visible, resetForm])

  // 关闭弹窗重置状态
  const handleClose = useCallback(() => {
    resetForm()
    onCancel()
  }, [resetForm, onCancel])

  // 发送验证码
  const handleSendCode = useCallback(async () => {
    try {
      await form.validateFields(['phone'])
      const phone = form.getFieldValue('phone')
      
      setCodeLoading(true)
      await sendVerifyCodeForForgot({ phone, key: 'ForgotPassword' })
      
      message.success('验证码发送成功')
      setCountdown(60)
    } catch (error: any) {
      console.log('发送验证码失败', error)
    } finally {
      setCodeLoading(false)
    }
  }, [form])

  const onFinish = useCallback(async (values: any) => {
    // 手动校验两次密码是否一致
    if (values.password !== values.confirmPassword) {
      message.warning('两次输入的密码不一致')
      return
    }

    try {
      setLoading(true)
      const { phone, code, password } = values
      const encryptedPassword = encryptPassword(password)
      
      await forgotPassword({
        phone,
        code,
        password: encryptedPassword,
      })

      message.success('密码修改成功，请使用新密码登录')
      handleClose()
    } catch (error: any) {
      console.log('重置密码失败', error)
      // 接口失败时清空验证码输入框
      form.setFieldValue('code', '')
    } finally {
      setLoading(false)
    }
  }, [handleClose])

  // 发送验证码按钮样式
  const sendCodeButtonClassName = useMemo(() => {
    const hasPhone = values?.phone?.trim()
    const baseClass =
      'h-[32px] !py-[5px] !px-[10px] !border !text-[14px] font-medium !rounded-[4px]'

    // 倒计时中
    if (countdown > 0) {
      return `${baseClass} !border-[#E3E6ED] bg-[#F5F5F5] text-[#9CA3AF] cursor-not-allowed hover:!border-[#E3E6ED] hover:!text-[#9CA3AF] hover:!bg-[#F5F5F5]`
    }

    // 没有输入手机号
    if (!hasPhone) {
      return `${baseClass} !border-[#0C99FF] bg-white text-[#0C99FF] !opacity-50 cursor-not-allowed hover:!border-[#0C99FF] hover:!text-[#0C99FF] hover:!bg-white hover:!opacity-50`
    }

    // 正常状态
    return `${baseClass} !border-[#0C99FF] bg-white text-[#0C99FF] hover:!border-[#0C99FF] hover:!text-[#0C99FF]`
  }, [countdown, values?.phone])

  return (
    <Modal
      open={visible}
      onCancel={handleClose}
      footer={null}
      closable={false}
      width={480}
      centered
      styles={{
        body: {
          padding: 0,
        },
        content: {
          padding: 0,
          borderRadius: '8px',
        },
      }}
    >
      <div className='flex flex-col p-6'>
        {/* 头部 */}
        <div className='flex items-center justify-between pb-2.5 mb-6 border-b border-[#E3E6ED]'>
          <div className='text-[22px] font-medium text-[#0C1F17]'>
            忘记密码
          </div>
          <button
            type='button'
            onClick={handleClose}
            className='flex items-center justify-center w-6 h-6 rounded cursor-pointer hover:bg-[#F5F7FA] transition-colors'
          >
            <CloseIcon className='w-4 h-4' />
          </button>
        </div>

        {/* 表单 */}
        <Form
          form={form}
          layout='vertical'
          onFinish={onFinish}
          requiredMark={false}
          preserve={false}
          autoComplete='off'
        >
          <Form.Item
            name='phone'
            rules={[
              { required: true, message: '请输入手机号' },
              {
                pattern: /^1[3-9]\d{9}$/,
                message: '请输入正确的手机号格式',
              },
            ]}
          >
            <Input
              placeholder='请输入手机号'
              autoComplete='off'
              maxLength={11}
              className={INPUT_CLASS_NAME}
            />
          </Form.Item>

          <Form.Item className='!mb-0'>
            <div className='flex gap-1'>
              <Form.Item
                name='code'
                rules={[{ required: true, message: '请输入验证码' }]}
                className='mb-0 flex-1'
              >
                <Input
                  placeholder='请输入验证码'
                  autoComplete='off'
                  className={INPUT_CLASS_NAME}
                />
              </Form.Item>
              <Button
                onClick={handleSendCode}
                disabled={!values?.phone || countdown > 0}
                loading={codeLoading}
                className={sendCodeButtonClassName}
              >
                {countdown > 0 ? `${countdown}秒后重发` : '发送验证码'}
              </Button>
            </div>
          </Form.Item>

          <Form.Item
            name='password'
            rules={getPasswordRules(t)}
            className='mt-4'
          >
            <Input.Password
              placeholder='请输入新密码'
              autoComplete='new-password'
              className={INPUT_CLASS_NAME}
              iconRender={(visible) =>
                visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item
            name='confirmPassword'
            rules={[
              { required: true, message: '请再次输入新密码' },
              ...getPasswordRules(t).slice(1),
            ]}
          >
            <Input.Password
              placeholder='请再次输入新密码'
              autoComplete='new-password'
              className={INPUT_CLASS_NAME}
              iconRender={(visible) =>
                visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item className='!mb-0 flex justify-end pt-2'>
            <Button
              className='mr-[6px] h-[32px] !py-[9px] !px-[24.5px] !border-0 bg-[#F5F5F5] text-[#0C1F17] hover:text-[#0C1F17] !text-[14px] font-medium rounded-[6px]'
              onClick={handleClose}
            >
              {tC('button.cancel')}
            </Button>
            <Button
              type='primary'
              htmlType='submit'
              loading={loading}
              disabled={!submittable}
              className='h-[32px] !py-[9px] !px-[24.5px] text-[#ffffff] bg-[#0C99FF] !border-0 !text-[14px] font-medium rounded-[6px] disabled:!bg-[#0C99FF] disabled:!text-[#ffffff] disabled:!opacity-50 disabled:cursor-not-allowed'
            >
              {tC('button.confirm')}
            </Button>
          </Form.Item>
        </Form>
      </div>
    </Modal>
  )
}

export default ForgotPasswordModal
