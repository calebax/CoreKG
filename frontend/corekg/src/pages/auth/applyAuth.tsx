import { useState, useEffect } from 'react'
import { Input, Button, message, App } from 'antd'
import { useMount } from 'ahooks'
import { copyText } from '@/utils'
import { getClusterID, registerLicense } from '@/api/auth'
import { fetchOrganizationProfile } from '@/api/organization'
import Copy from '@/assets/icons/auth-copy.svg?react'
import Notice from '@/assets/icons/auth-notice.svg?react'
import Step from '@/assets/icons/auth-step.svg?react'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { EditableAvatar } from './EditableAvatar'

function RenderStep({
  device,
  activeCode,
  setActiveCode,
  webTabIcon,
  setWebTabIcon,
  webTabName,
  setWebTabName,
  orgLogo,
  setOrgLogo,
  orgName,
  setOrgName,
}: {
  device: string
  activeCode: string
  setActiveCode: (value: string) => void
  webTabIcon: string
  setWebTabIcon: (value: string) => void
  webTabName: string
  setWebTabName: (value: string) => void
  orgLogo: string
  setOrgLogo: (value: string) => void
  orgName: string
  setOrgName: (value: string) => void
}) {
  const { message } = App.useApp()
  const steps = [
    {
      id: 0,
      step: '步骤1：',
      des: '复制设备ID',
      infomation: '请将设备ID发送给客服获取激活码。',
    },
    {
      id: 1,
      step: '步骤2：',
      des: '输入激活码',
      infomation: '收到厂商返回的激活码后在此输入。',
    },
    {
      id: 2,
      step: '步骤3：',
      des: '配置基本信息',
      infomation: '配置网站标签页信息和组织信息（可选）',
    },
  ]

  const stepList = steps.map((stepItem) => (
    <div key={stepItem.id}>
      <div className='flex items-center mt-4'>
        <Step className='w-6 h-6 mr-1' />
        <span className='text-[18px]'>{stepItem.step}</span>
        <span className='text-[16px]'>{stepItem.des}</span>
      </div>
      <p
        className={
          stepItem.id === 2
            ? 'text-[12px] text-[#919497] my-2'
            : 'text-[#616373] my-2'
        }
      >
        {stepItem.infomation}
      </p>
      {stepItem.id === 0 && (
        <div className='flex items-center'>
          <span className='text-[#616373]'>设备ID：</span>
          <span className='text-[#1E1F28]'>{device}</span>
          <Copy
            onClick={() => {
              copyText(device)
              message.success('复制成功')
            }}
            className='w-5 h-5 ml-1.5 cursor-pointer'
          />
        </div>
      )}
      {stepItem.id === 1 && (
        <div className='flex items-center'>
          <Input.TextArea
            className='bg-[#F8F9FD] w-2xl p-3 my-2'
            placeholder='请输入'
            autoSize={{ minRows: 4, maxRows: 4 }}
            value={activeCode}
            onChange={(e) => setActiveCode(e.target.value)}
          />
        </div>
      )}
      {stepItem.id === 2 && (
        <div className='mt-4 space-y-6'>
          {/* 设置网站标签页图标 */}
          <div>
            <div className='text-[16px] font-medium mb-2'>
              设置网站标签页图标
            </div>
            <div className='flex flex-col gap-4'>
              <EditableAvatar
                type='website_logo'
                value={webTabIcon}
                onChange={(v) => {
                  if (v) {
                    setWebTabIcon(v)
                  }
                }}
              />
              <div className='text-[12px] text-[#919497]'>
                支持图片格式: jpg/jpeg/png, 图片大小不超过5M, 为保证显示效果,
                请上传宽高比 1:1 的图片
              </div>
            </div>
          </div>

          {/* 设置网站标签页名称 */}
          <div>
            <div className='text-[16px] font-medium mb-2'>
              设置网站标签页名称
            </div>
            <Input
              className='bg-[#F8F9FD] w-2xl'
              placeholder='请输入'
              value={webTabName}
              onChange={(e) => setWebTabName(e.target.value)}
              maxLength={20}
              showCount
            />
          </div>

          {/* 设置组织/企业logo图标 */}
          <div>
            <div className='text-[16px] font-medium mb-2'>
              设置组织/企业logo图标
            </div>
            <div className='flex flex-col gap-4'>
              <EditableAvatar
                type='company_logo'
                value={orgLogo}
                onChange={(v) => {
                  if (v) setOrgLogo(v)
                }}
              />
              <div className='text-[12px] text-[#919497]'>
                支持图片格式: jpg/jpeg/png, 图片大小不超过5M, 为保证显示效果,
                请上传宽高比 1:1 的图片
              </div>
            </div>
          </div>

          {/* 设置组织/企业名称 */}
          <div>
            <div className='text-[16px] font-medium mb-2'>
              设置组织/企业名称
            </div>
            <Input
              className='bg-[#F8F9FD] w-2xl'
              placeholder='请输入'
              value={orgName}
              onChange={(e) => setOrgName(e.target.value)}
              maxLength={20}
              showCount
            />
          </div>
        </div>
      )}
    </div>
  ))
  return stepList
}

export default function ApplyAuth({
  onAuthSuccess,
}: {
  onAuthSuccess?: () => void
}) {
  const [device, setDevice] = useState('')
  const [activeCode, setActiveCode] = useState('')

  const { setConfig, ...deployConfig } = useDeployConfig()
  const [favicon, setFavicon] = useState('')
  const [appName, setAppName] = useState('')

  useEffect(() => {
    setFavicon(deployConfig.favicon.light)
    setAppName(deployConfig.appName)
  }, [deployConfig.favicon.light, deployConfig.appName])

  const { uinList, userInfo, updateCurrentOrganization } = useLocalStore()
  const [orgLogo, setOrgLogo] = useState('')
  const [orgName, setOrgName] = useState('')

  const getDeviceID = async () => {
    const res = await getClusterID({})
    setDevice(res)
    console.log('res', res)
    return res
  }

  // 获取组织信息用于显示
  const loadOrganizationInfo = async () => {
    const { companyName = '', logo = '' } =
      uinList.find((unit) => String(unit.id) === String(userInfo.uinId)) ?? {}
    try {
      const result = await fetchOrganizationProfile()
      const normalizedName = result?.name?.trim() || companyName
      const normalizedLogo = result?.logo || logo
      setOrgName(normalizedName)
      setOrgLogo(normalizedLogo)

      // 更新本地存储
      updateCurrentOrganization({
        name: normalizedName,
        logo: normalizedLogo,
      })
    } catch {
      setOrgName(companyName)
      setOrgLogo(logo)
    }
  }

  const handleImpower = async () => {
    // 校验组织/企业名称
    const trimmedOrgName = orgName?.trim()
    if (!trimmedOrgName) {
      message.error('请输入组织/企业名称')
      return
    }

    // 校验网站标签页名称
    const trimmedAppName = appName?.trim()
    if (!trimmedAppName) {
      message.error('请输入网站标签页名称')
      return
    }

    try {
      const res = await registerLicense({
        license: activeCode,
        company_name: orgName,
        company_logo: orgLogo,
        website_info: {
          website_logo: favicon,
          website_name: appName,
        },
      })
      const currentConfig = useDeployConfig.getState()
      setConfig({
        ...currentConfig,
        appName,
        favicon: {
          light: favicon,
          dark: favicon,
        },
      })
      updateCurrentOrganization({ name: orgName, logo: orgLogo })

      message.success('授权成功')
      onAuthSuccess?.()
      return res
    } catch (error) {
      console.error('授权失败:', error)
      message.error('授权失败，请检查激活码是否正确')
    }
  }

  useMount(() => {
    getDeviceID()
    loadOrganizationInfo()
  })

  return (
    <div className=' p-6 rounded-lg bg-[#FCFCFE] border-[#D7D9E5] border-1'>
      <div className='flex items-center mb-8 py-3 pl-9 bg-[#165DFF12]'>
        <Notice className='w-4 h-4 mr-1' />
        设备ID仅用于授权，不包含隐私信息。
      </div>
      <p className='font-medium text-base'>授权向导</p>
      <RenderStep
        device={device}
        activeCode={activeCode}
        setActiveCode={setActiveCode}
        webTabIcon={favicon}
        setWebTabIcon={setFavicon}
        webTabName={appName}
        setWebTabName={setAppName}
        orgLogo={orgLogo}
        setOrgLogo={setOrgLogo}
        orgName={orgName}
        setOrgName={setOrgName}
      />
      <Button
        onClick={handleImpower}
        className='w-40 h-9 mt-14 rounded-lg text-[#FFFFFF] font-medium bg-linear-to-r from-[#6598FF] to-[#266EFF]'
      >
        应用授权
      </Button>
    </div>
  )
}
