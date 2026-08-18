export {}

declare global {
  /** 需要和public/config.js保持一致 */
  type DeployConfig = {
    /** saas 海外版 私有化 */
    version: 'saas' | 'international' | 'custom'
    coze_url: string
    /** 私有化的版本 */
    mode: 'h3c' | 'jiefang' | 'cimc' | 'default'
    /** 侧边栏顶部logo */
    logo: string
    /** 侧边栏顶部 应用名称 */
    title: string
    /** 标签页title */
    appName: string
    favicon: {
      light: string
      dark: string
    }
    isH3CTest?: boolean
    /** 强制指定全局问答的问候语 */
    globalGreeting?: string
    /** AI问答输入框最大字数 */
    qaInputMaxLength?: number
  }
  interface Window {
    /** 需要和public/config.js保持一致 */
    __DEPLOYCONFIG: Readonly<DeployConfig>
  }
}
