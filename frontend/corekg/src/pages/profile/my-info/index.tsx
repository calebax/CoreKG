import { useState, useEffect, useRef, FC } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Form, Input, Button, Spin, App, InputRef } from 'antd'
import { useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { produce } from 'immer'
import { useTranslation } from 'react-i18next'
import {
  changeWX_account,
  DetailPersonalCenter,
  getLoginConfig,
  UpdatePersonalInfo,
  uploadAvatarImg,
} from '@/api/account'
import BackIcon from '@/assets/icons/backIcon.svg'
import UserIconDefault from '@/assets/icons/userIcon-default.svg'
import UserIconUpdate from '@/assets/icons/userIcon-update.svg'
import { LanguageSelector } from '@/components/languageSelector'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import backIcon from '../../../assets/icons/backIcon.svg'
import myInformationBg from '../../../assets/icons/myInfomation-bg.svg'
import userIconDefault from '../../../assets/icons/userIcon-default.svg'
import userIconUpdate from '../../../assets/icons/userIcon-update.svg'
import PasswordModal from './components/PasswordModal'
import PhoneModal from './components/PhoneModal'
import { UsageManagement } from './components/UsageManagement'
import { WeChat } from './components/Wechat'
import EditIcon from './images/edit.svg?react'
import styles from './styles.module.scss'

interface UserInfo {
  wechat_union_id: string
  wechat_web_open_id: string
  identify: string
  avatar_url: string
  bio: string
  email: string
  phone: string
  name: string
  real_name: string
  id: number
  created_at: string
  uin: number
  /** 0没密码 */
  has_password: number
}

interface CompanyInfo {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: null | string
  name: string
  alias: string
  description: string
  logo: string
  address: string
  tel: string
  email: string
  website: string
  company_status: string
}

interface Position {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: null | string
  company_id: number
  name: string
  description: string
}

interface EmployeeDetail {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: null | string
  CompanyID: number
  UserID: number
  Uin: number
  sys_role: string
  user_name: string
  real_name: string
  phone: string
  email: string
  positions: Position[]
  action_paths: null | any
}

interface PersonalCenterData {
  user_info: UserInfo
  company_info: CompanyInfo
  employee_detail: EmployeeDetail
}
/**
 * 高阶组件\
 * 存在code和state时会绑定微信
 */
const withBindWX = (Comp: FC) => {
  const CompWithBindWX: FC = () => {
    const { t } = useTranslation('pages')
    const { t: tM } = useTranslation('messages')
    const { message } = App.useApp()
    const navigate = useNavigate()
    const [searchParams] = useSearchParams()
    const code = searchParams.get('code')
    const state = searchParams.get('state')
    useEffect(() => {
      if (code && state) {
        changeWX_account({ code })
          .then(() => {
            message.success(tM('changeBindingSuccess'))
          })
          .finally(() => {
            const newSearchParams = new URLSearchParams(searchParams)
            newSearchParams.delete('code')
            newSearchParams.delete('state')
            navigate(
              {
                search: newSearchParams.toString(),
              },
              { replace: true },
            )
          })
      }
    }, [code, message, navigate, searchParams, state])
    if (code && state)
      return <Spin tip={t('profile.changingWechatBinding')} fullscreen />
    return <Comp />
  }
  return CompWithBindWX
}

const MyInfo = withBindWX(() => {
  const { version } = useDeployConfig()
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [personalData, setPersonalData] = useState<PersonalCenterData | null>(
    null,
  )
  const hasPassword = personalData?.user_info.has_password !== 0
  // custom 环境下不展示修改密码入口
  const showPasswordEdit = version !== 'custom'
  const [isLargeScreen, setIsLargeScreen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [passwordModalVisible, setPasswordModalVisible] = useState(false)
  const [phoneModalVisible, setPhoneModalVisible] = useState(false)
  const [wechatModalVisible, setWechatModalVisible] = useState(false)
  const [avatarLoading, setAvatarLoading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const userInfoStore = useLocalStore() // 获取本地存储实例
  // 微信appId
  const { data: appId } = useRequest(async () => {
    if (version === 'custom') return
    const res = await getLoginConfig({
      domain: location.origin,
    })
    return res.wechat.appid as string
  })
  useEffect(() => {
    fetchPersonalData()

    // 添加媒体查询检测大屏幕
    const mediaQuery = window.matchMedia('(min-width: 1920px)')

    // 初始化状态
    setIsLargeScreen(mediaQuery.matches)

    // 添加监听器
    const handleMediaQueryChange: (event: MediaQueryListEvent) => void = (
      event,
    ) => {
      setIsLargeScreen(event.matches)
    }

    mediaQuery.addEventListener('change', handleMediaQueryChange)

    // 清除监听器
    return () => {
      mediaQuery.removeEventListener('change', handleMediaQueryChange)
    }
  }, [])

  const fetchPersonalData = async () => {
    try {
      const data = await DetailPersonalCenter()
      // console.log('data', data)
      setPersonalData(data)
      form.setFieldsValue({
        identify: data?.user_info?.name,
        email: data?.user_info?.email,
      })
      userInfoStore.setUserInfo({
        ...userInfoStore.userInfo,
        name: data?.user_info?.name,
        avatar: data?.user_info?.avatar_url,
      })
    } catch (error) {
      console.error('获取个人信息失败', error)
      message.error(tM('getPersonalInfoFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      const newName = form.getFieldValue('identify')?.trim() || ''

      await UpdatePersonalInfo({
        avatar_url: personalData?.user_info?.avatar_url || '',
        name: newName,
        email: values.email,
      })

      // 更新本地存储：仅更新当前组织的 uinName
      const updatedUinList = userInfoStore.uinList.map((item) => {
        if (String(item.id) !== String(userInfoStore.userInfo.uinId))
          return item
        return { ...item, uinName: newName }
      })
      userInfoStore.setLogin({
        token: userInfoStore.token,
        userInfo: userInfoStore.userInfo,
        uinList: updatedUinList,
      })
      message.success(tM('saveSuccess'))
      fetchPersonalData()
    } finally {
      setLoading(false)
    }
  }

  const handleAvatarClick = () => {
    // 如果正在上传，不允许点击
    if (avatarLoading) return

    // 直接打开文件选择器
    fileInputRef.current?.click()
  }

  const handleAvatarChange = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const files = event.target.files
    // console.log('files', files)
    if (!files || !files.length) return

    const file = files[0]

    // 验证文件类型
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif']
    if (!allowedTypes.includes(file.type)) {
      message.error(
        tM('incorrectImageFormatOnlyAllowTargetFormat', {
          target: 'jpg、png、gif',
        }),
      )
      return
    }

    // 验证文件大小 (小于2MB)
    const isLt2M = file.size / 1024 / 1024 < 2
    if (!isLt2M) {
      message.error(
        tM('uploadedImageTooLargeMaxAllowedTarget', { target: '2MB' }),
      )
      return
    }

    try {
      setAvatarLoading(true)

      const data = {
        file,
        purpose: 'cu-image',
      }
      // 调用上传API
      const { url: newAvatarUrl } = await uploadAvatarImg(data, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setPersonalData((prev) => {
        if (!prev) return null
        return {
          ...prev,
          user_info: {
            ...prev.user_info,
            avatar_url: newAvatarUrl,
          },
        }
      })

      // 1. 自动调用保存接口，更新用户信息
      try {
        const values = await form.validateFields()
        await UpdatePersonalInfo({
          avatar_url: newAvatarUrl || '',
          name: values.identify,
          email: values.email,
        })

        // 2. 更新本地存储中的avatar
        userInfoStore.setUserInfo({
          ...userInfoStore.userInfo,
          avatar: newAvatarUrl,
        })

        message.success(tM('avatarUpdateSuccess'))
      } catch (error) {
        console.error('保存用户信息失败', error)
        message.error(tM('avatarUploadedButSaveUserInfoFailed'))
      }

      setAvatarLoading(false)
    } catch (error) {
      console.error('头像上传失败', error)
      message.error(tM('avatarUploadFailed'))
      setAvatarLoading(false)
    }
  }

  // 定义输入框的通用样式
  const inputStyle = {
    width: '100%',
    fontSize: '16px',
    fontWeight: 400,
    color: '#4E5969',
  }

  // 占位符样式
  const placeholderStyle = 'text-[#616373] font-normal'

  // 背景图
  const myInfoBg = {
    backgroundImage: `url(${myInformationBg})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  }
  const nickname = Form.useWatch('identify', { form, preserve: true })
  const nameInputRef = useRef<InputRef>(null)
  const [edit, setEdit] = useState<boolean>(false)

  const handleBlur = (newVal: string) => {
    if (newVal?.trim?.()?.length) {
      form.setFieldValue('identify', newVal.trim())
    }
    setEdit(!edit)
  }

  return (
    <div className='w-full h-[calc(100%-40px)] flex flex-col lg:flex-row gap-4 px-23 py-10 overflow-auto'>
      {/* 左侧：基本信息 */}
      <div className='flex-1 border border-[#E3E6ED] rounded-[10px] p-[16px] md:p-[24px] pb-[60px] md:pb-[70px] relative'>
        {/* 标题 */}
        <div className='text-[16px] md:text-[18px] text-[#000000] font-[500] mb-[12px] md:mb-[16px]'>
          {t('profile.basicInformation')}
        </div>

        {/* 表单页面*/}
        <div>
          <Form form={form} layout='vertical' requiredMark={false}>
            {/* 个人信息 */}
            <div className='flex h-[50px] gap-[8px] md:gap-[12px] mb-[10px]'>
              <div
                className={`w-[50px] h-[50px] flex-shrink-0 cursor-pointer ${styles.myInfoAvatar}`}
                onClick={!avatarLoading ? handleAvatarClick : undefined}
              >
                {avatarLoading ? (
                  <div className='w-full h-full flex items-center justify-center bg-gray-100 rounded-full'>
                    <Spin size='small' />
                  </div>
                ) : personalData?.user_info?.avatar_url ? (
                  <img
                    src={personalData.user_info.avatar_url}
                    alt={t('profile.avatar')}
                    className='w-full h-full object-cover rounded-full'
                  />
                ) : (
                  <div className='w-full h-full flex items-center justify-center rounded-full'>
                    <img
                      src={UserIconDefault}
                      alt={t('profile.avatar')}
                      className='w-full h-full'
                    />
                  </div>
                )}
                <div className={styles.myInfoAvatarMask}>
                  <EditIcon className='text-[#E1E1E1]' />
                </div>
                {/* 隐藏的文件输入框 */}
                <input
                  type='file'
                  ref={fileInputRef}
                  accept='image/png, image/jpeg, image/gif'
                  className='hidden'
                  onChange={handleAvatarChange}
                  disabled={avatarLoading}
                />
              </div>
              <div className='flex-1 h-[100%] flex flex-col py-[4px] min-w-0'>
                <div className='h-[22px] flex items-center gap-[8px] md:gap-[14px] leading-[22px] font-[500] mb-[4px]'>
                  {!edit ? (
                    <div className='max-w-[8em] md:max-w-[10em] whitespace-nowrap overflow-hidden text-ellipsis text-[14px] md:text-[16px]'>
                      {nickname}
                    </div>
                  ) : (
                    <Input
                      ref={nameInputRef}
                      autoFocus
                      className='h-[22px] w-[8em] md:w-[10em] hover:border-[#0C99FF] !shadow-none focus:border-[#0C99FF]'
                      maxLength={20}
                      defaultValue={nickname}
                      onPressEnter={() => nameInputRef.current?.blur?.()}
                      onBlur={(event) => handleBlur(event.target.value)}
                    />
                  )}

                  <EditIcon
                    onClick={() => setEdit(!edit)}
                    className='cursor-pointer text-[#919497] flex-shrink-0'
                  />
                </div>
                <div className='flex flex-col sm:flex-row gap-[4px] sm:gap-[7px] font-[500] text-[#C4C8CC] leading-[14px] sm:leading-[16px] text-[11px] sm:text-[12px]'>
                  <div className='whitespace-nowrap overflow-hidden text-ellipsis'>
                    {t('profile.id')}:
                    {personalData?.user_info?.id
                      ? String(personalData.user_info.id)
                      : '2025-05-05 12:04:07'}
                  </div>
                  <div className='whitespace-nowrap overflow-hidden text-ellipsis'>
                    {t('profile.registrationTime')}
                    {personalData?.user_info?.created_at
                      ? dayjs(personalData.user_info.created_at).format(
                          'YYYY-MM-DD HH:mm:ss',
                        )
                      : '2025-05-05 12:04:07'}
                  </div>
                </div>
              </div>
            </div>
            <Form.Item
              className={`mb-[8px] md:mb-[10px] bg-[#FAFAFA] p-[8px] md:p-[10px] rounded-[6px] ${styles.myInfoFormItem}`}
              label={
                <div className='flex gap-1 items-center overflow-hidden'>
                  <span className='font-[500] text-[#0C1F17] text-[14px] md:text-[16px] leading-[20px] md:leading-[22px] flex-grow'>
                    {t('profile.email')}
                  </span>
                </div>
              }
              name='email'
              rules={[
                {
                  required: version === 'custom',
                  message: '请输入邮箱',
                },
                {
                  type: 'email',
                  message: t('profile.pleaseEnterValidEmail'),
                },
              ]}
            >
              <Input
                placeholder='请输入邮箱'
                className='h-[40px] md:h-[44px] bg-white border !shadow-none border-[#d9d9d9] rounded-[6px] text-[#3C4149] text-[14px] md:text-[16px] leading-6 px-[8px] py-2 hover:border-[#E3E6ED] focus:border-[#0C99FF]'
              />
            </Form.Item>
            <Form.Item
              className={`mb-[8px] md:mb-[10px] bg-[#FAFAFA] p-[8px] md:p-[10px] rounded-[6px] ${styles.myInfoFormItem}`}
              label={
                <div className='flex gap-1 items-center overflow-hidden'>
                  <span className='font-[500] text-[#0C1F17] text-[14px] md:text-[16px] leading-[20px] md:leading-[22px] flex-grow'>
                    {t('profile.password')}
                  </span>
                </div>
              }
            >
              <div className='relative w-full'>
                <Input
                  type='password'
                  value={hasPassword ? '******' : ''}
                  className='h-[40px] md:h-[44px] bg-white border !shadow-none border-[#d9d9d9] rounded-[6px] text-[#3C4149] text-[14px] md:text-[16px] leading-6 px-[8px] py-2 hover:border-[#E3E6ED] focus:border-[#0C99FF] placeholder:text-[#3C4149]'
                  readOnly
                />
                {showPasswordEdit && (
                  <span
                    className='absolute right-[9px] top-1/2 transform -translate-y-1/2 text-[#0C1F17] cursor-pointer text-[14px] md:text-[16px] font-[500] leading-[20px] md:leading-[24px] whitespace-nowrap'
                    onClick={() => setPasswordModalVisible(true)}
                  >
                    {t('profile.editPassword')}
                  </span>
                )}
              </div>
            </Form.Item>

            <Form.Item
              className={`mb-[8px] md:mb-[10px] bg-[#FAFAFA] p-[8px] md:p-[10px] rounded-[6px] ${styles.myInfoFormItem}`}
              label={
                <div className='flex gap-1 items-center overflow-hidden'>
                  <span className='font-[500] text-[#0C1F17] text-[14px] md:text-[16px] leading-[20px] md:leading-[22px] flex-grow'>
                    {t('profile.phoneNumber', { target: ':' })}
                  </span>
                </div>
              }
            >
              <div className='relative w-full'>
                <Input
                  value={personalData?.user_info?.phone || ''}
                  className='h-[40px] md:h-[44px] bg-white !shadow-none border border-[#d9d9d9] rounded-[6px] text-[#3C4149] text-[14px] md:text-[16px] leading-6 px-[8px] py-2 hover:border-[#E3E6ED] focus:border-[#0C99FF] placeholder:text-[#3C4149]'
                  readOnly
                  disabled={version === 'custom'}
                />
                {version !== 'custom' && (
                  <span
                    className='absolute right-[9px] top-1/2 transform -translate-y-1/2 text-[#0C1F17] cursor-pointer text-[14px] md:text-[16px] font-[500] leading-[20px] md:leading-[24px] whitespace-nowrap'
                    onClick={() => {
                      if (!hasPassword) {
                        message.warning(tM('pleaseSetPasswordFirst'))
                        return
                      }
                      setPhoneModalVisible(true)
                    }}
                  >
                    {t('profile.editPhoneNumber')}
                  </span>
                )}
              </div>
            </Form.Item>
            {version === 'custom' ? null : (
              <Form.Item>
                <WeChat
                  appId={appId}
                  name={personalData?.user_info.name}
                  hasPassword={hasPassword}
                />
              </Form.Item>
            )}
            <div className='absolute bottom-[16px] md:bottom-[24px] right-[16px] md:right-[24px]'>
              <Button
                onClick={handleSave}
                loading={loading}
                className='h-[32px] bg-[#0C99FF] !border-0 hover:border-[#0C99FF] items-center !text-[#ffffff] font-[500] text-[13px] md:text-[14px] leading-[1] px-[12px] md:px-[15px] rounded-[6px]'
                style={{
                  lineHeight: '24px',
                }}
              >
                {t('profile.save')}
              </Button>
            </div>
          </Form>
        </div>
        {/* 修改密码弹窗：custom 环境下不展示 */}
        {showPasswordEdit && (
          <PasswordModal
            hasPassword={hasPassword}
            visible={passwordModalVisible}
            onCancel={() => setPasswordModalVisible(false)}
            onSuccess={fetchPersonalData}
          />
        )}
        <PhoneModal
          visible={phoneModalVisible}
          onCancel={() => setPhoneModalVisible(false)}
          currentPhone={personalData?.user_info?.phone}
          onSuccess={() => {
            fetchPersonalData()
            setPhoneModalVisible(false)
          }}
        />
      </div>

      {/* 右侧：用量管理 */}
      <div className='flex-1'>
        <UsageManagement />
      </div>
    </div>
  )
})

export default MyInfo
