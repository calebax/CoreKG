import { produce } from 'immer'
import { create } from 'zustand'
import { getGlobalInfo } from '@/api/organization'

const defaultDeployConfig: DeployConfig = {
  version: 'custom',
  mode: 'h3c',
  logo: '',
  title: '',
  appName: '',
  qaInputMaxLength: 10000,
  favicon: {
    light: '',
    dark: '',
  },
  coze_url: '',
}

export const useDeployConfig = create<
  DeployConfig & {
    setConfig: (config: Partial<DeployConfig>) => void
  }
>((set, get) => {
  const initialConfig = window.__DEPLOYCONFIG ?? defaultDeployConfig
  if (initialConfig.version === 'custom') {
    getGlobalInfo()
      .then((globalInfo) => {
        const { website_name, website_logo } = globalInfo.website_info ?? {}
        get().setConfig(
          produce(get(), (draft) => {
            if (website_name) {
              draft.appName = website_name
            }
            if (website_logo) {
              draft.favicon = {
                light: website_logo,
                dark: website_logo,
              }
            }
          }),
        )
      })
      .catch((error) => {
        console.error('Failed to fetch global info:', error)
      })
  }
  return {
    ...initialConfig,
    setConfig: (partialConfig) => {
      const currentConfig = get()
      const updatedConfig: DeployConfig = {
        ...currentConfig,
        ...partialConfig,
        ...(partialConfig.favicon && {
          favicon: {
            ...currentConfig.favicon,
            ...partialConfig.favicon,
          },
        }),
      }
      set(updatedConfig)
      window.__DEPLOYCONFIG = updatedConfig
    },
  }
})

export const H3CConfig: DeployConfig = {
  version: 'custom',
  coze_url: '/coze',
  mode: 'h3c',
  logo: '/icons/h3c/logo.png',
  title: '/icons/h3c/title.svg',
  appName: 'TuringQuery',
  qaInputMaxLength: 10000,
  favicon: {
    light: '/icons/h3c/favicon-light.ico',
    dark: '/icons/h3c/favicon-dark.ico',
  },
}

export const SAASConfig: DeployConfig = {
  version: 'saas',
  coze_url: 'https://coze.corekg.com',
  mode: 'h3c',
  logo: '/icons/saas/logo.svg',
  title: '/icons/saas/title.svg',
  appName: 'CoreKG AI',
  qaInputMaxLength: 500,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
}

export const resolveDeployUrl = (
  baseUrl: string,
  targetUrl: string,
  fallback: string,
) => {
  if (!targetUrl) {
    return fallback
  }

  if (/^https?:\/\//.test(targetUrl)) {
    return targetUrl
  }

  if (targetUrl.startsWith(':')) {
    return `${baseUrl}${targetUrl}`
  }

  if (targetUrl.startsWith('/')) {
    return `${baseUrl}${targetUrl}`
  }

  return `${baseUrl}/${targetUrl}`
}
