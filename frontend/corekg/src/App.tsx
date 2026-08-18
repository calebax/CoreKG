import React, { type FC, Fragment, PropsWithChildren, useEffect } from 'react'
import { RouterProvider } from 'react-router-dom'
import {
  ConfigProvider,
  App as AntdApp,
  Radio,
  FloatButton,
  Popover,
  Skeleton,
} from 'antd'
import { StyleProvider } from '@ant-design/cssinjs'
import { RightOutlined, SwapOutlined } from '@ant-design/icons'
import { useFavicon, useRequest, useTheme, useTitle } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { match } from 'ts-pattern'
import { send } from '@/api/request'
import { SupportedLangs } from '@/locales/types'
import router from '@/router/index'
import { getLanguageConfig, setDayjsLocale } from '@/utils/locale'
import { useDeployConfig, SAASConfig, H3CConfig } from '@/utils/useDeployConfig'
import { Agent } from './components/Agent'
import ErrorBoundary from './router/ErrorBoundary'
import ErrorFallback from './router/ErrorFallback'

function App() {
  const { i18n } = useTranslation()

  const currentLang = i18n.language as SupportedLangs
  const langConfig = getLanguageConfig(currentLang)

  useEffect(() => {
    // 同步Day.js语言设置
    setDayjsLocale(currentLang)
  }, [currentLang])
  const content = (
    <>
      <RouterProvider router={router} />
      <Agent />
    </>
  )
  return (
    <StyleProvider layer>
      <ConfigProvider
        locale={langConfig.antd}
        theme={{
          token: {
            colorPrimary: '#0C99FF',
          },
          components: {
            Table: {
              headerBorderRadius: 0,
              headerBg: '#E6E6E6',
              borderColor: '#F0F0F0',
            },
            Input: {
              hoverBorderColor: '#0C99FF',
            },
            Radio: {
              colorPrimaryBorder: '#0C99FF',
              colorPrimaryHover: '#0C99FF',
              colorPrimary: '#0C99FF',
            },
            Tree: {
              colorPrimary: '#0C99FF',
              colorPrimaryHover: '#0C99FF',
              colorPrimaryBorder: '#0C99FF',
              colorWhite: '#fff',
              colorBorder: '#C4C8CC',
              fontSizeLG: 14,
              colorText: '#0C1F17',
              directoryNodeSelectedBg: '#9194971A',
              nodeSelectedBg: '#9194971A',
              nodeHoverBg: '#9194971A',
              nodeHoverColor: '#0C1F17',
              indentSize: 30,
              titleHeight: 30,
              controlInteractiveSize: 14,
              lineHeight: 1,
              marginXS: 0,
              paddingXS: 5,
              borderRadius: 4,
              fontWeightStrong: 500,
            },
            Select: {
              hoverBorderColor: '#0C99FF',
              activeBorderColor: '#0C99FF',
            },
            Switch: {
              colorPrimary: '#0C99FF',
              colorPrimaryHover: '#0C99FF',
            },
            Steps: {
              colorPrimary: '#0C99FF',
            },
            Modal: {
              colorPrimary: '#0C99FF',
            },
            Button: {
              colorPrimary: '#0C99FF',
              colorPrimaryHover: '#0C99FF',
              colorPrimaryActive: '#0C99FF',
              colorPrimaryText: '#0C99FF',
              colorLink: '#0C99FF',
              colorInfo: '#0C99FF',
            },
            Checkbox: {
              colorBgContainerDisabled: '#CEEBFF',
              colorTextDisabled: '#ffffff',
            },
            Spin: {
              colorPrimary: '#0C99FF',
            },
          },
        }}
      >
        <AntdApp
          className='w-full h-full relative m-0 p-0'
          style={{ fontFamily: 'Inter ' }}
        >
          <RightOutlined className='hidden!' />
          <ErrorBoundary fallback={<ErrorFallback />}>
            {
              // 用环境变量判断
              (import.meta.env.MODE === 'test' || import.meta.env.DEV) &&
              // 用host判断
              ['127.0.0.1', 'localhost', 'example.com'].includes(
                window.location.hostname,
              ) ? (
                <VersionSwitcher>{content}</VersionSwitcher>
              ) : (
                content
              )
            }
          </ErrorBoundary>
          <FaviconConfig />
        </AntdApp>
      </ConfigProvider>
    </StyleProvider>
  )
}

export default App

const FaviconConfig: FC = () => {
  const { theme } = useTheme()
  const { favicon, appName } = useDeployConfig()

  // 根据主题选择light或dark favicon
  const faviconUrl = theme === 'dark' ? favicon.dark : favicon.light
  useFavicon(faviconUrl)
  useTitle(appName)

  return null
}

const VersionSwitcher = Fragment
// const VersionSwitcher: FC<PropsWithChildren> = (props) => {
//   const { setConfig, ...currentConfig } = useDeployConfig()
//   const { version, mode } = currentConfig
//   const services = ['forest', 'account', 'kesearch', 'knowledge', 'chat']
//   const { run, loading } = useRequest(
//     async (config: DeployConfig = currentConfig) => {
//       // 后端微服务自动同步
//       const ps = services.map(async (s) => {
//         const { deploy_mode } = await send(`${s}.NowDeployMode`, {})
//         return deploy_mode
//       })
//       const [deploy_mode] = await Promise.all(ps)
//       // 如果当前微服务版本和所需不符 切换其版本
//       if (
//         (deploy_mode && config.version === 'saas') ||
//         (!deploy_mode && config.version === 'custom')
//       ) {
//         await send(`${services[0]}.SwitchPrivateEvn`, {})
//       }
//       setConfig(config)
//     },
//   )
//   const switcher = (
//     <Popover
//       content={
//         <div className='p-4 flex flex-col gap-1'>
//           当前版本:
//           <Radio.Group
//             value={match({ version, mode })
//               .with({ version: 'saas' }, () => 'saas')
//               .with({ version: 'custom', mode: 'h3c' }, () => 'h3c')
//               .otherwise(() => 'saas')}
//             onChange={(e) => {
//               run(
//                 match(e.target.value)
//                   .with('h3c', () => H3CConfig)
//                   .with('saas', () => SAASConfig)
//                   .otherwise(() => SAASConfig),
//               )
//             }}
//           >
//             <Radio value={'saas'}>saas</Radio>
//             <Radio value={'h3c'}>新华三</Radio>
//           </Radio.Group>
//         </div>
//       }
//     >
//       <FloatButton className='left-4' icon={<SwapOutlined />}></FloatButton>
//     </Popover>
//   )
//   if (loading) {
//     return <Skeleton className='p-4' active />
//   }
//   return (
//     <>
//       {switcher}
//       {props.children}
//     </>
//   )
// }
