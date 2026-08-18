import { FC } from 'react'
import { App, Form, Input, Modal } from 'antd'
import { useBoolean } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { checkPassword } from '@/api/account'
import EyeOutIcon from '../../images/eye-out.svg?react'
import EyeIcon from '../../images/eye.svg?react'
import weChatIcon from '../../images/wx.svg'
import styles from '../../styles.module.scss'
import WechatModal from '../WechatModal'

export type WeChat = {
  appId?: string
  hasPassword?: boolean
  name?: string
}
export const WeChat: FC<WeChat> = (props) => {
  const { t: tM } = useTranslation('messages')
  const { t } = useTranslation('pages')
  const { appId, hasPassword, name } = props
  const { message } = App.useApp()
  const [form] = Form.useForm()
  const [checkPasswordOpen, { toggle: togglePSW }] = useBoolean()
  const [checking, { toggle: toggleChecking }] = useBoolean()
  const [changeWXOpen, { toggle: toggleWX }] = useBoolean()
  const changeWX = () => {
    if (!hasPassword) {
      message.warning(tM('pleaseSetPasswordFirst'))
      return
    }
    togglePSW()
  }
  // if (!appId) return null
  return (
    <>
      <div className='flex items-center justify-between py-2 md:py-2.5 bg-[#FAFAFA] px-[8px] md:px-[10px] rounded-[6px]'>
        <div className='flex gap-[8px] md:gap-[10px] flex-col'>
          <span className='text-[#0C1F17] font-[500] text-[14px] md:text-[16px] leading-[20px] md:leading-[22px]'>
            {t('profile.wechatName')}
          </span>
          <div className='w-[32px] h-[32px]'>
            <img
              src={weChatIcon}
              alt={t('profile.wechat')}
              className='w-full h-full cursor-pointer'
              onClick={changeWX}
            />
          </div>
        </div>
        <div className='flex gap-2.5 h-[60px] md:h-[64px] items-left justify-center pt-[32px] md:pt-[36px] px-[8px] md:px-[10px] rounded'>
          <span
            onClick={changeWX}
            className='cursor-pointer text-[#0C1F17] font-[500] text-[14px] md:text-[16px] leading-[20px] md:leading-[24px] whitespace-nowrap'
          >
            {t('profile.changeWechatBinding')}
          </span>
        </div>
      </div>
      <Modal
        title={
          <div className='text-[22px] font-[500] text-[#1D2129]'>
            {t('profile.validPassword')}
          </div>
        }
        open={checkPasswordOpen}
        onCancel={togglePSW}
        onOk={async () => {
          toggleChecking()
          try {
            await form.validateFields()
          } finally {
            toggleChecking()
          }
          message.success(tM('verificationSuccess'))
          togglePSW()
          toggleWX()
        }}
        okButtonProps={{ loading: checking }}
        className={styles.weChatModal}
        destroyOnHidden
      >
        <Form form={form} layout='vertical' preserve={false}>
          <Form.Item
            name='password'
            label={t('profile.originalPassword')}
            validateFirst
            rules={[
              {
                required: true,
                message: t('profile.pleaseEnterOriginalPassword'),
              },
              { min: 8, message: t('profile.passwordLengthMin', { count: 8 }) },
              {
                max: 36,
                message: t('profile.passwordLengthMax', { count: 36 }),
              },
              {
                validator: async (_, password: string) =>
                  checkPassword({ password }),
                validateTrigger: 'onSubmit',
              },
            ]}
          >
            <Input.Password
              placeholder={t('profile.pleaseEnterOriginalPassword')}
              className='h-[32px] border  border-[#E3E6ED]  hover:border-[#0C99FF] focus:border-[#0C99FF] !rounded-[4px] placeholder:text-[#0C1F17]'
              // prefix={<UserOutlined style={iconStyle} />}
              iconRender={(visible) => (visible ? <EyeIcon /> : <EyeOutIcon />)}
            />
          </Form.Item>
        </Form>
      </Modal>
      <WechatModal appId={appId} visible={changeWXOpen} onCancel={toggleWX} />
    </>
  )
}
