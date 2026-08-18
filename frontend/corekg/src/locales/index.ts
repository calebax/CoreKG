import i18n from 'i18next'
// import LanguageDetector from 'i18next-browser-languagedetector'
import { initReactI18next } from 'react-i18next'
import { SupportedLangs } from './types'

window.__LANG = window.__LANG || 'zh-Hans'
document.documentElement.lang = window.__LANG

// 动态导入语言包 - 实现按需加载，减少初始包体积
const loadResource = async (lang: string, ns: string) => {
  const resource = await import(`./${languageMap[lang]}/${ns}.json`)
  return resource.default
}

// 语言映射表 - 处理浏览器语言码到应用语言码的转换
const languageMap: Record<string, SupportedLangs> = {
  zh: 'zh-CN',
  'zh-Hans': 'zh-CN',
  'zh-CN': 'zh-CN',
  'zh-Hant': 'zh-TW',
  'zh-TW': 'zh-TW',
  'zh-HK': 'zh-TW', // 香港使用繁体
  en: 'en-US',
  'en-US': 'en-US',
  'en-GB': 'en-US', // 英国英语映射到美国英语
  fr: 'fr-FR',
  'fr-FR': 'fr-FR',
}

const preloadResources = async () => {
  const defaultLang = window.__LANG
  const namespaces = ['common', 'pages', 'messages']

  for (const ns of namespaces) {
    const resource = await loadResource(defaultLang, ns)
    i18n.addResourceBundle(defaultLang, ns, resource)
  }
}

async function initializeI18n() {
  await i18n
    // .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      // 默认语言 - 找不到翻译时的后备语言
      fallbackLng: window.__LANG,

      // 默认命名空间 - 不指定命名空间时使用
      defaultNS: 'common',

      // 声明所有命名空间
      ns: ['common', 'pages', 'messages'],
      interpolation: {
        escapeValue: false, // React已处理XSS，无需再转义
        skipOnVariables: false, // 确保占位符未提供时仍尝试渲染
      },
      // detection: {
      //   // 检测顺序：先查localStorage，后查浏览器设置
      //   order: ['localStorage', 'navigator'],

      //   // 缓存用户选择到localStorage
      //   caches: ['localStorage'],

      //   // localStorage的key名
      //   lookupLocalStorage: 'i18n_lang',

      //   // 语言码转换函数
      //   convertDetectedLanguage: (lng: string) => {
      //     return languageMap[lng] || 'zh-CN'
      //   },
      // },

      // React配置
      react: {
        useSuspense: true, // 使用Suspense进行加载状态处理
      },
    })
  await preloadResources()
}

// 预加载默认语言资源 - 避免首次渲染闪烁

i18n.on('languageChanged', async (lng: SupportedLangs) => {
  window.__LANG = lng
  document.documentElement.lang = lng
  // 加载当前语言的所有命名空间资源
  const namespaces = ['common', 'pages', 'messages']

  for (const ns of namespaces) {
    try {
      // 检查该语言的命名空间是否已加载，避免重复加载
      if (!i18n.hasResourceBundle(lng, ns)) {
        const resource = await loadResource(lng, ns)
        i18n.addResourceBundle(lng, ns, resource) // 新增语言包
      }
    } catch (error) {
      console.error(`Failed to load ${lng}/${ns}.json:`, error)
    }
  }
})

export { initializeI18n }

export default i18n
