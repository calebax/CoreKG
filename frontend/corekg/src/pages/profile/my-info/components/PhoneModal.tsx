import { useState } from 'react'
import { Modal, Form, Input, Button, message } from 'antd'
import { UserOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { checkPassword, getVerifyCode, updatePhone } from '@/api/account'
import EyeOutIcon from '../images/eye-out.svg?react'
import EyeIcon from '../images/eye.svg?react'
import styles from '../styles.module.scss'

interface PhoneModalProps {
  visible: boolean
  onCancel: () => void
  currentPhone?: string
  onSuccess: () => void
}

const PhoneModal = ({
  visible,
  onCancel,
  currentPhone,
  onSuccess,
}: PhoneModalProps) => {
  const { t: tM } = useTranslation('messages')
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)

  const handleOk = async () => {
    try {
      setLoading(true)
      const values = await form.validateFields()
      await updatePhone({
        phone: values.newPhone,
        phone_code: values.verifyCode,
      })
      message.success(tM('modifySuccess', { target: t('profile.phoneNumber') }))
      form.resetFields()
      onSuccess()
    } finally {
      setLoading(false)
    }
  }

  const handleSendVerifyCode = async () => {
    // 验证手机号格式
    const phone = form.getFieldValue('newPhone')
    if (!phone || !/^1[3-9]\d{9}$/.test(phone)) {
      message.error(tM('pleaseEnterValidPhoneNumber'))
      return
    }

    // 发送验证码（也是需要调用接口的）
    await getVerifyCode({ phone })
    message.success(tM('verificationCodeSent'))
    let count = 60
    setCountdown(count)
    const timer = setInterval(() => {
      count--
      setCountdown(count)
      if (count <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  }

  return (
    <Modal
      title={
        <div className='text-[22px] font-[500] text-[#1D2129]'>
          {t('profile.editPhoneNumber')}
        </div>
      }
      open={visible}
      onCancel={onCancel}
      footer={null}
      destroyOnHidden
      width={460}
      className={`phone-modal ${styles.phoneModal}`}
      centered
    >
      <Form
        form={form}
        layout='vertical'
        preserve={false}
        initialValues={{ currentPhone }}
        requiredMark={false}
      >
        <Form.Item
          name='password'
          validateFirst
          rules={[
            {
              required: true,
              message: t('profile.pleaseEnterOriginalPassword'),
            },
            {
              min: 8,
              message: t('profile.passwordLengthMin', { target: 8 }),
            },
            {
              max: 36,
              message: t('profile.passwordLengthMax', { count: 36 }),
            },
            {
              validator: async (_, password: string) => {
                await checkPassword({ password })
              },
              validateTrigger: 'onSubmit',
            },
          ]}
        >
          <Input.Password
            placeholder={t('profile.pleaseEnterOriginalPassword')}
            className='h-[32px] border !shadow-none border-[#E3E6ED] hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC]'
            iconRender={(visible) => (visible ? <EyeIcon /> : <EyeOutIcon />)}
          />
        </Form.Item>
        <Form.Item
          name='newPhone'
          rules={[
            { required: true, message: t('profile.pleaseEnterNewPhoneNumber') },
            {
              pattern: /^1[3-9]\d{9}$/,
              message: t('profile.pleaseEnterValidPhoneNumber'),
            },
          ]}
        >
          <Input
            placeholder={t('profile.pleaseEnterNewPhoneNumberPlaceholder')}
            maxLength={11}
            className='h-[32px] border !shadow-none border-[#E3E6ED] hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC]'
          />
        </Form.Item>
        <Form.Item
          name='verifyCode'
          rules={[
            {
              required: true,
              message: t('profile.pleaseEnterVerificationCode'),
            },
          ]}
        >
          <div className='flex items-center'>
            <Input
              placeholder={t('profile.pleaseEnterVerificationCode')}
              className='h-[32px] border !shadow-none border-[#E3E6ED]  hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#C4C8CC]'
              maxLength={6}
            />
            <Button
              type='default'
              className='ml-[4px] min-w-24 hover:text-[#0C99FF] border-[#E3E6ED] hover:border-[#0C99FF] !h-[32px] flex-shrink-0 !rounded !text-[14] !px-[10px] font-[500]'
              disabled={countdown > 0}
              onClick={handleSendVerifyCode}
            >
              {countdown > 0
                ? t('profile.resendAfterTargetSeconds', { target: countdown })
                : t('profile.sendVerificationCode')}
            </Button>
          </div>
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

export default PhoneModal
