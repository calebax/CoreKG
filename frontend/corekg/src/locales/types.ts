import zhCommon from './zh-CN/common.json'
import zhMessages from './zh-CN/messages.json'
import zhPages from './zh-CN/pages.json'

// 支持的语言类型
export type SupportedLangs = 'zh-CN' | 'zh-TW' | 'en-US' | 'fr-FR'

// 基于中文语言包生成类型 - 确保翻译key的类型安全
export type CommonKeys = keyof typeof zhCommon
export type PageKeys = keyof typeof zhPages
export type MessageKeys = keyof typeof zhMessages

// 命名空间类型
export interface Resources {
  common: typeof zhCommon
  pages: typeof zhPages
  messages: typeof zhMessages
}

// 扩展i18next类型定义 - 提供IDE智能提示和编译时检查
declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common'
    resources: Resources
  }
}

// 语言配置接口
export interface LanguageConfig {
  code: SupportedLangs
  name: string // 显示名称
  antd: any // Ant Design语言包
  dayjs: string // Day.js语言代码
  flag?: string // 可选：国旗emoji
}
