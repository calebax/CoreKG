import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Input, Form, message } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  chooseUin,
  loginByPassword,
  loginByPasswordPrivate,
  setPasswordChangeReminder,
} from '@/api/account'
import { Agreement } from '@/components/agreement'
import useLocalStore from '@/stores/local'
import { encryptPassword } from '@/utils/crypto'
import { useDeployConfig } from '@/utils/useDeployConfig'
import ForgotPasswordModal from './components/ForgotPasswordModal'
import PasswordChangeReminderModal from './components/PasswordChangeReminderModal'
import PasswordModal from './components/PasswordModalForLogin'

interface PasswordLoginProps {
  onChooseUin: (uin: {
    refreshToken: string
    userInfo: any
    uinList: any[]
    username?: string
    password?: string
    companyQuota?: number
    companyUserId?: number
  }) => void
  onLogin?: () => void
}

interface LoginValues {
  username: string
  password: string
}

const PasswordLogin: React.FC<PasswordLoginProps> = ({
  onChooseUin,
  onLogin,
}) => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const localStore = useLocalStore()

  // 密码修改提醒相关状态
  const [showReminderModal, setShowReminderModal] = useState(false)
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  const [showForgotModal, setShowForgotModal] = useState(false)
  const [reminderLoading, setReminderLoading] = useState(false)
  const [loginData, setLoginData] = useState<any>(null)
  const [currentPassword, setCurrentPassword] = useState('')

  // custom 环境下不展示忘记密码功能
  const showForgotPassword = version !== 'custom'

  // 处理密码修改提醒弹窗的跳过操作
  const handleSkipReminder = async (dontShowAgain?: boolean) => {
    if (dontShowAgain && loginData) {
      try {
        setReminderLoading(true)
        await setPasswordChangeReminder({
          always_ignore: true,
          user_id: loginData.userInfo.id,
          refresh_token: loginData.refreshToken,
        })
      } catch (error) {
        console.log('error', error)
        return
      } finally {
        setReminderLoading(false)
      }
    }
    setShowReminderModal(false)
    proceedToChooseUin()
  }

  // 处理去修改密码操作
  const handleModifyPassword = () => {
    setShowReminderModal(false)
    setShowPasswordModal(true)
  }

  // 处理密码修改成功
  const handlePasswordChangeSuccess = () => {
    setShowPasswordModal(false)
    setShowReminderModal(false)
    form.resetFields()
    setCurrentPassword('')
    setLoginData(null) // 清空登录数据
    // message.success('密码修改成功，请重新登录')
  }

  // 处理密码修改弹窗取消
  const handlePasswordModalCancel = () => {
    setShowPasswordModal(false)
    // 如果用户取消修改密码，回到密码修改提醒弹窗
    setShowReminderModal(true)
  }

  // 继续执行选择团队的逻辑
  const proceedToChooseUin = async () => {
    if (!loginData) return

    const {
      userInfo,
      uinList,
      refreshToken,
      jwtToken,
      username,
      password,
      companyQuota,
      companyUserId,
    } = loginData

    if (uinList.length > 0) {
      // 私有化环境下无需用户手动选组织，直接自动选第一个
      if (version === 'custom') {
        try {
          setLoading(true)
          const res = await chooseUin({
            login_way: userInfo.loginWay,
            refresh_token: refreshToken,
            uin_id: uinList[0].id,
            user_id: userInfo.id,
          })
          localStore.setLogin({
            token: res.jwt_token,
            userInfo: { ...userInfo, uinId: uinList[0].id },
            uinList,
          })
          message.success(tM('loginSuccess'))
          onLogin?.()
        } catch (error) {
          console.error('auto chooseUin error', error)
        } finally {
          setLoading(false)
        }
        return
      }
      onChooseUin({
        refreshToken: refreshToken,
        userInfo: userInfo,
        uinList: uinList,
        username: username,
        password: password,
        companyQuota: companyQuota,
        companyUserId: companyUserId,
      })
    } else {
      // 如果没有组织，直接登录
      localStore.setLogin({
        token: jwtToken || refreshToken,
        userInfo: userInfo,
        uinList: uinList,
      })
      message.success(tM('loginSuccess'))
      onLogin?.()
    }
  }

  const handlePasswordLogin = async (values: LoginValues) => {
    setLoading(true)
    try {
      const { username, password } = values
      // 对密码进行加密
      const encryptedPassword = encryptPassword(password)
      // 保存当前密码用于修改密码弹窗
      setCurrentPassword(password)
      const body = {
        username: username,
        password: encryptedPassword,
        domain_name: location.origin,
        // domain_name: 'https://example.com', //为了本地开发登录页面，临时设置的，开发完之后需要注释掉这行代码，恢复上一行的代码
      }
      const res =
        version === 'custom'
          ? await loginByPasswordPrivate(body)
          : await loginByPassword(body)
      if (res.login_status === 'failed') {
        message.error(tM('loginFailedPleaseRelogin'))
        navigate('/')
        return
      }
      const userInfo = {
        id: res.user_id,
        avatar: res.user_info.avatar_url,
        name: res.user_info.name,
        uinId: res.user_info.uin,
        loginWay: res.login_way,
      }
      const uinList = res.uin.map((x: any) => {
        return {
          id: x.uin.ID,
          role: x.role,
          uinName: x.uin.Name,
          companyName: x.company_name,
          companyStatus: x.company_status, // 组织状态
          uinStatus: x.uin.UinStatus, // 身份状态
          subjectType: x.uin.SubjectType, // company 表示企业、individual 表示个人
          subjectId: x.uin.SubjectID, // 企业 id，个人为 0
          logo: x.company_logo, // 组织 logo
          companyQuota: x.company_quota, // 组织配额
          companyUserId: x.company_user_id, // 组织用户id
        }
      })

      // 检查 password_changed 字段
      if (res.user_info.password_changed === false) {
        // 保存登录数据，用于后续处理
        setLoginData({
          userInfo,
          uinList,
          refreshToken: res.refresh_token,
          jwtToken: res.jwt_token, // 保存 jwt_token
          username: username, // 保存用户名
          password: password, // 保存密码
          companyQuota: res.user_info.company_quota, // 保存组织配额
          companyUserId: res.user_info.company_user_id, // 保存组织用户id
        })
        // 显示密码修改提醒弹窗
        setShowReminderModal(true)
        return
      }

      // password_changed 为 true，正常执行原有逻辑
      // 私有化环境下直接自动选第一个 uin，无需用户手动选组织
      if (uinList.length > 0) {
        if (version === 'custom') {
          try {
            const chooseRes = await chooseUin({
              login_way: userInfo.loginWay,
              refresh_token: res.refresh_token,
              uin_id: uinList[0].id,
              user_id: userInfo.id,
            })
            localStore.setLogin({
              token: chooseRes.jwt_token,
              userInfo: { ...userInfo, uinId: uinList[0].id },
              uinList,
            })
            message.success(tM('loginSuccess'))
            onLogin?.()
          } catch (error) {
            console.error('auto chooseUin error', error)
          }
          return
        }
        onChooseUin({
          refreshToken: res.refresh_token,
          userInfo: userInfo,
          uinList: uinList,
          username: username,
          password: password,
          companyQuota: res.user_info.company_quota,
          companyUserId: res.user_info.company_user_id,
        })
      } else {
        // 如果没有组织，直接登录
        localStore.setLogin({
          token: res.jwt_token,
          userInfo: userInfo,
          uinList: uinList,
        })
        message.success(tM('loginSuccess'))
        onLogin?.()
      }
    } catch (error) {
      console.log('error', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <div className='w-full min-h-[15rem] flex flex-col'>
        <Form
          className='w-full'
          form={form}
          name='password-login'
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
              // prefix={<UserOutlined />}
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
              // prefix={<LockOutlined />}
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
        {/* 文案提示用户可通过扫码快速注册 */}
        {/* <div
          className='mt-2 text-sm text-gray-500 text-center'
          aria-live='polite'
        >
          {t('other.noAccountScanQrCodeToRegisterImmediately')}
        </div> */}
      </div>

      {/* 密码修改提醒弹窗 */}
      <PasswordChangeReminderModal
        visible={showReminderModal}
        onCancel={() => setShowReminderModal(false)}
        onSkip={handleSkipReminder}
        onModifyPassword={handleModifyPassword}
        loading={reminderLoading}
      />

      {/* 修改密码弹窗 */}
      <PasswordModal
        visible={showPasswordModal}
        onCancel={handlePasswordModalCancel}
        onSuccess={handlePasswordChangeSuccess}
        initialPassword={currentPassword}
        userId={loginData?.userInfo?.id || ''}
        refreshToken={loginData?.refreshToken || ''}
      />

      {/* 忘记密码弹窗：custom 环境下不展示 */}
      {showForgotPassword && (
        <ForgotPasswordModal
          visible={showForgotModal}
          onCancel={() => setShowForgotModal(false)}
        />
      )}
    </>
  )
}

export default PasswordLogin
