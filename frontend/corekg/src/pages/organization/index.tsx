import {
  type ChangeEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { useNavigate } from 'react-router-dom'
import { App, Breadcrumb, Spin } from 'antd'
import { useTranslation } from 'react-i18next'
import { updateWebsiteInfo } from '@/api/account'
import { uploadImage } from '@/api/common'
import {
  uploadAvatarImg,
  uploadWebsiteLogo,
  fetchOrganizationProfile,
  updateOrganizationProfile,
} from '@/api/organization'
import UploadIcon from '@/assets/icons/userIcon-update.svg'
import SeparatorIcon from '@/assets/separator.svg?react'
import DefaultOrgLogo from '@/components/Layout/SidebarWrapper/images/defaultLogo.svg'
import useLocalStore from '@/stores/local'
import { loadFile } from '@/utils/loadFile'
import { useDeployConfig } from '@/utils/useDeployConfig'
import styles from './styles.module.scss'

export default function OrganizationManagement() {
  const { t } = useTranslation('pages')
  const { t: tMessages } = useTranslation('messages')
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { title, version } = useDeployConfig()
  const { uinList, userInfo, updateCurrentOrganization } = useLocalStore()
  const { favicon, appName, setConfig } = useDeployConfig()
  // 表单相关状态
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const [loading, setLoading] = useState(true) // 页面初始加载状态
  const [uploading, setUploading] = useState(false) // 头像上传状态
  const [saving, setSaving] = useState(false) // 保存状态

  // 表单数据状态
  const [formState, setFormState] = useState({
    name: '',
    logo: '',
  })
  // 初始数据状态，用于对比是否有变更和重置功能
  const [initialState, setInitialState] = useState({
    name: '',
    logo: '',
  })

  // 获取当前组织信息，用于默认显示
  const currentOrganization = useMemo(() => {
    try {
      const current = uinList.find(
        (unit) => String(unit.id) === String(userInfo.uinId),
      )
      if (current) {
        return {
          name: current.companyName || title,
          logo: current.logo || '',
        }
      }
    } catch (error) {
      console.error('Unable to derive current organization info', error)
    }

    return {
      name: title,
      logo: '',
    }
  }, [title, uinList, userInfo.uinId])

  // 页面初始化：加载组织信息
  useEffect(() => {
    let active = true

    const load = async () => {
      setLoading(true)
      try {
        // 获取组织信息
        const result = await fetchOrganizationProfile()

        if (!active) return

        const normalizedName =
          result?.name?.trim() || currentOrganization.name || ''
        const normalizedLogo = result?.logo || currentOrganization.logo || ''

        // 设置表单数据和初始数据
        setFormState({
          name: normalizedName,
          logo: normalizedLogo,
        })
        setInitialState({
          name: normalizedName,
          logo: normalizedLogo,
        })
        updateCurrentOrganization({
          name: normalizedName,
          logo: normalizedLogo,
        })
      } catch {
        // 获取失败时使用当前组织信息作为fallback
        if (!active) return

        setFormState({
          name: currentOrganization.name || '',
          logo: currentOrganization.logo || '',
        })
        setInitialState({
          name: currentOrganization.name || '',
          logo: currentOrganization.logo || '',
        })
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    load()

    return () => {
      active = false
    }
  }, [
    currentOrganization.logo,
    currentOrganization.name,
    updateCurrentOrganization,
  ])

  // 网页标签页图标上传
  const handleWebsiteLogoUpload = useCallback(() => {
    loadFile(
      async (fileList) => {
        const file = fileList[0]
        if (!file) return

        // 文件格式验证
        const validTypes = ['image/png', 'image/jpeg', 'image/jpg']
        if (!validTypes.includes(file.type)) {
          message.error(
            tMessages('incorrectImageFormatOnlyAllowTargetFormat', {
              target: 'jpg/jpeg/png',
            }),
          )
          return
        }

        // 文件大小验证（5MB限制）
        const isLt5M = file.size / 1024 / 1024 < 5
        if (!isLt5M) {
          message.error(
            tMessages('uploadedImageTooLargeMaxAllowedTarget', {
              target: '5MB',
            }),
          )
          return
        }

        try {
          const public_url = await uploadWebsiteLogo(
            { file, purpose: 'company-logo' },
            {
              headers: { 'Content-Type': 'multipart/form-data' },
            },
          ).then((res: any) => res.public_url)
          await updateWebsiteInfo({
            website_info: {
              website_logo: public_url!,
              website_name: appName,
            },
          })

          // 更新 deployConfig
          setConfig({
            favicon: {
              light: public_url!,
              dark: public_url!,
            },
          })

          message.success(
            tMessages('organizationLogoUpdateSuccess') || '上传成功',
          )
        } catch (error) {
          console.error('Upload website logo failed', error)
          message.error('上传失败，请重试')
        }
      },
      {
        accept: 'image/png,image/jpeg,image/jpg',
      },
    )
  }, [message, tMessages, appName, setConfig])

  // 处理头像上传点击
  const handleAvatarClick = useCallback(() => {
    if (uploading) return
    fileInputRef.current?.click()
  }, [uploading])

  // 处理文件上传
  const handleAvatarChange = useCallback(
    async (event: ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0]

      if (!file) return

      // 文件格式验证
      const validTypes = ['image/png', 'image/jpeg', 'image/jpg']
      if (!validTypes.includes(file.type)) {
        message.error(
          tMessages('incorrectImageFormatOnlyAllowTargetFormat', {
            target: 'jpg/jpeg/png',
          }),
        )
        event.target.value = ''
        return
      }

      // 文件大小验证（5MB限制）
      const isLt5M = file.size / 1024 / 1024 < 5
      if (!isLt5M) {
        message.error(
          tMessages('uploadedImageTooLargeMaxAllowedTarget', {
            target: '5MB',
          }),
        )
        event.target.value = ''
        return
      }

      setUploading(true)

      try {
        // 上传文件到服务器
        const data = {
          file,
          purpose: 'company-logo',
        }

        const newLogo = await uploadAvatarImg(data, {
          headers: { 'Content-Type': 'multipart/form-data' },
        }).then((res: any) => res.public_url)

        // 更新表单中的logo
        setFormState((prev) => ({
          ...prev,
          logo: newLogo,
        }))

        // 更新初始状态（因为头像上传已自动保存到服务器）
        setInitialState((prev) => ({
          ...prev,
          logo: newLogo,
        }))

        // 更新全局状态，同步侧边栏和切换组织弹窗的显示
        updateCurrentOrganization({
          name: formState.name.trim() || currentOrganization.name,
          logo: newLogo,
        })

        message.success(tMessages('organizationLogoUpdateSuccess'))
      } catch {
        console.error('Upload organization logo failed')
      } finally {
        setUploading(false)
        if (event.target) {
          event.target.value = ''
        }
      }
    },
    [
      currentOrganization.name,
      formState.name,
      message,
      tMessages,
      updateCurrentOrganization,
    ],
  )

  // 处理组织名称输入（限制50字符）
  const handleNameChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const value = event.target.value
      // 限制最多50字符
      if (value.length > 50) {
        return
      }
      setFormState((prev) => ({
        ...prev,
        name: value,
      }))
    },
    [],
  )

  // 取消修改：重置表单到初始状态
  const handleCancel = useCallback(() => {
    setFormState(initialState)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [initialState])

  // 保存组织信息
  const handleSave = useCallback(async () => {
    const trimmedName = formState.name.trim()

    if (!trimmedName) {
      message.error(tMessages('organizationNameRequired'))
      return
    }

    // 检查是否有变更
    const hasChanges = trimmedName !== initialState.name

    if (!hasChanges) {
      return
    }

    setSaving(true)

    try {
      // 调用API更新组织名称
      await updateOrganizationProfile({
        name: trimmedName,
      })

      // 更新初始状态和表单状态
      setInitialState({
        name: trimmedName,
        logo: formState.logo,
      })
      setFormState((prev) => ({
        ...prev,
        name: trimmedName,
      }))

      // 更新全局状态，同步侧边栏和切换组织弹窗的显示
      updateCurrentOrganization({
        name: trimmedName,
        logo: formState.logo,
      })

      message.success(tMessages('saveSuccess'))
    } catch (error) {
      console.error('Save organization info failed', error)
    } finally {
      setSaving(false)
    }
  }, [
    formState.logo,
    formState.name,
    initialState,
    message,
    tMessages,
    updateCurrentOrganization,
  ])

  // 计算表单是否已被修改（用于按钮状态控制）
  const isDirty = useMemo(() => {
    const trimmedName = formState.name.trim()
    return trimmedName !== initialState.name
  }, [formState.name, initialState.name])

  // 计算是否可以提交（名称非空、有修改、未在上传中）
  const canSubmit = useMemo(() => {
    return formState.name.trim().length > 0 && isDirty && !uploading
  }, [formState.name, isDirty, uploading])

  return (
    <div className='w-full h-full flex flex-col bg-[#FAFAFA]'>
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] border-b border-[#EFF1F4]'>
        <Breadcrumb
          className={styles.layoutHeader}
          separator={<SeparatorIcon />}
          items={[
            {
              title: (
                <span
                  className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                  onClick={() => {
                    navigate(`/`)
                  }}
                >
                  问答
                </span>
              ),
            },
            {
              title: (
                <span className='cursor-pointer text-sm font-medium text-[#3C4149]'>
                  组织管理
                </span>
              ),
            },
          ]}
        />
      </div>
      <div className='flex-1 overflow-auto p-6 bg-[#FAFAFA]'>
        <div className='flex flex-col h-full justify-between'>
          <div className='flex flex-col gap-4'>
            <h1 className='text-lg font-medium text-[#1A1A1A]'>
              {t('organization.basicInfoTitle')}
            </h1>

            <div className='flex flex-col gap-5'>
              {loading ? (
                <div className='flex justify-center items-center py-20'>
                  <Spin />
                </div>
              ) : (
                <>
                  <div className='bg-[#9194971A] px-2.5 py-6 rounded-[10px] flex flex-col gap-2.5'>
                    <div className='relative'>
                      <div
                        className='w-12 h-12 flex items-center justify-center cursor-pointer'
                        onClick={handleAvatarClick}
                      >
                        {uploading ? (
                          <Spin />
                        ) : formState.logo ? (
                          <img
                            src={formState.logo}
                            alt={t('organization.logoLabel')}
                            className='w-10 h-10 rounded-full'
                          />
                        ) : (
                          <img
                            src={currentOrganization.logo || DefaultOrgLogo}
                            alt={t('organization.logoLabel')}
                            className='w-10 h-10 rounded-full'
                          />
                        )}
                      </div>
                      <div
                        className='absolute bottom-0 right-0 w-4 h-4 rounded-full flex items-center justify-center cursor-pointer'
                        onClick={handleAvatarClick}
                      >
                        <img src={UploadIcon} alt='' className='w-3 h-3' />
                      </div>
                      <input
                        ref={fileInputRef}
                        type='file'
                        accept='image/png,image/jpeg'
                        className='hidden'
                        onChange={handleAvatarChange}
                      />
                    </div>
                    <p className='text-sm font-medium text-[#919497]'>
                      {t('organization.logoHint')}
                    </p>
                    {version === 'custom' ? (
                      <>
                        <div className='relative'>
                          <div
                            className='w-12 h-12 flex items-center justify-center cursor-pointer'
                            onClick={handleWebsiteLogoUpload}
                          >
                            <img
                              src={favicon.light}
                              alt={t('organization.logoLabel')}
                              className='w-10 h-10 rounded-full'
                            />
                          </div>
                          <div
                            className='absolute bottom-0 right-0 w-4 h-4 rounded-full flex items-center justify-center cursor-pointer'
                            onClick={handleWebsiteLogoUpload}
                          >
                            <img src={UploadIcon} alt='' className='w-3 h-3' />
                          </div>
                        </div>
                        <p className='text-sm font-medium text-[#919497]'>
                          网页标签页（支持图片格式：jpg/jpeg/png，图片大小不超过5M，为保证显示效果，请上传宽高比
                          1:1 的图片）
                        </p>
                      </>
                    ) : null}
                  </div>

                  <div className='flex flex-col gap-[6px]'>
                    <label className='text-base font-medium text-[#3C4149]'>
                      {t('organization.nameLabel')}
                    </label>
                    <div className='border border-[#E3E6ED] h-8 rounded-md w-96'>
                      <input
                        id='organization-name'
                        value={formState.name}
                        placeholder={t('organization.namePlaceholder') ?? ''}
                        onChange={handleNameChange}
                        disabled={saving}
                        className='w-full h-full px-2.5 py-[6px] bg-transparent border-none outline-none text-sm text-[3C4149] placeholder:text-[#ABAFB2]'
                      />
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>

          <div className='flex gap-[6px] justify-end items-center'>
            <button
              className='px-6 py-2 bg-[#F5F5F5] text-[#0C1F17] rounded-md text-sm font-medium hover:bg-[#F5F5F5] disabled:cursor-not-allowed'
              onClick={handleCancel}
              disabled={loading || uploading || saving || !isDirty}
            >
              {t('organization.cancel')}
            </button>
            <button
              className={`px-6 py-2 rounded-md text-sm font-medium ${
                canSubmit
                  ? 'bg-[#0C99FF] text-[#ffffff] hover:bg-[#0C99FF]'
                  : 'bg-[#0C99FF] text-[#ffffff]'
              }`}
              onClick={handleSave}
              disabled={!canSubmit || saving}
            >
              {saving ? <Spin size='small' /> : t('organization.save')}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
