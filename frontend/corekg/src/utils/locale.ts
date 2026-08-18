// Ant Design语言包
import enUS from 'antd/locale/en_US'
import frFR from 'antd/locale/fr_FR'
import zhCN from 'antd/locale/zh_CN'
import zhTW from 'antd/locale/zh_TW'
// Day.js语言包
import dayjs from 'dayjs'
import 'dayjs/locale/en'
import 'dayjs/locale/fr'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/zh-tw'
import { SupportedLangs, LanguageConfig } from '@/locales/types'

// Ant Design语言包映射
export const antdLocales = {
  'zh-CN': zhCN,
  'zh-TW': zhTW,
  'en-US': enUS,
  'fr-FR': frFR,
}

// Day.js语言代码映射
export const dayjsLocales = {
  'zh-CN': 'zh-cn',
  'zh-TW': 'zh-tw',
  'en-US': 'en',
  'fr-FR': 'fr',
}

// 完整语言配置
export const languages: LanguageConfig[] = [
  {
    code: 'zh-CN',
    name: '简体中文',
    antd: zhCN,
    dayjs: 'zh-cn',
    flag: '🇨🇳',
  },
  {
    code: 'zh-TW',
    name: '繁體中文',
    antd: zhTW,
    dayjs: 'zh-tw',
    flag: '🇹🇼',
  },
  {
    code: 'en-US',
    name: 'English',
    antd: enUS,
    dayjs: 'en',
    flag: '🇺🇸',
  },
  {
    code: 'fr-FR',
    name: 'Français',
    antd: frFR,
    dayjs: 'fr',
    flag: '🇫🇷',
  },
]

// 设置Day.js语言 - 用于日期格式化
export const setDayjsLocale = (lang: SupportedLangs) => {
  const locale = dayjsLocales[lang]
  dayjs.locale(locale)
}

// 获取当前语言配置
export const getLanguageConfig = (lang: SupportedLangs): LanguageConfig => {
  return languages.find((l) => l.code === lang) || languages[0]
}
