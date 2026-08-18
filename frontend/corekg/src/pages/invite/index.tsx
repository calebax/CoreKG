import React, { useState } from 'react'
import { Tabs } from 'antd'
import type { TabsProps } from 'antd'
import { useMount } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { getLoginConfig } from '@/api/account'
import LoadingCover from '@/components/common/LoadingCover'
import InvitePasswordLogin from './InvitePasswordLogin'
import WechatLogin from './WechatLogin'
import { useInviteCode as useInviteKey } from './useInviteKey'

const LoginPage: React.FC = () => {
  const { t } = useTranslation('pages')
  const { key, type, invitor, companyName, handlePasswordLogin } =
    useInviteKey()
  const [activeTab, setActiveTab] = useState('wechat')
  const [tabItems, setTabItems] = useState<TabsProps['items']>([])
  const [loginConfig, setLoginConfig] = useState<{
    bg: string
    title: string
  }>({
    bg: '',
    title: '',
  })
  const [passwordLoginLoading, setPasswordLoginLoading] = useState(false)

  const [initLoading, setInitLoading] = useState(true)
  const init = async () => {
    if (!key || type !== 'key') return
    try {
      const res = await getLoginConfig({
        domain_name: location.origin,
        // domain_name: 'https://example.com', //为了本地开发登录页面，临时设置的，开发完之后需要注释掉这行代码，恢复上一行的代码
      })
      const loginTypes: TabsProps['items'] = []
      loginTypes.push({
        key: 'wechat',
        label: <span>{t('invite.wechatLogin')}</span>,
        children: <WechatLogin appId={res.wechat.appid} />,
      })
      loginTypes.push({
        key: 'password',
        label: <span>{t('other.accountPasswordLogin')}</span>,
        children: (
          <InvitePasswordLogin
            onLogin={async (username, password) => {
              setPasswordLoginLoading(true)
              try {
                await handlePasswordLogin(username, password)
              } finally {
                setPasswordLoginLoading(false)
              }
            }}
            loading={passwordLoginLoading}
          />
        ),
      })

      setTabItems(loginTypes)
      setActiveTab(loginTypes[0].key)
      setLoginConfig({
        bg: res.image_url,
        title: res.title,
      })
    } finally {
      setInitLoading(false)
    }
  }

  useMount(() => {
    init()
  })

  return (
    <div
      className='w-full h-screen flex bg-gray-50'
      style={{
        background: `url(${loginConfig.bg}) no-repeat center center / cover`,
      }}
    >
      <div className='flex-1 h-full'>&nbsp;</div>
      <div className='flex-1 h-full flex items-center justify-center'>
        <div className='w-100 p-10 bg-white rounded-lg shadow-md'>
          <div className='text-center flex flex-col'>
            <h2
              className={cn('text-base text-gray-900 mb-2', {
                invisible: !invitor || !companyName,
              })}
            >
              {invitor}
              {t('invite.inviteYouToJoinTeam')}
              <span className='ml-2'>{companyName}</span>
              <br />
              {t('invite.pleaseScanQrCodeToLoginAndJoin')}
            </h2>
          </div>
          <div className='w-full h-90 relative'>
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              centered
              items={tabItems}
              className='login-tabs'
            />
            <LoadingCover loading={initLoading} />
          </div>
        </div>
      </div>
    </div>
  )
}

export default LoginPage
