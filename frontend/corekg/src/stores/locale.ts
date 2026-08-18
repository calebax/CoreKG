import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import i18n from '@/locales'
import { SupportedLangs } from '@/locales/types'

interface LocaleState {
  language: SupportedLangs
  setLanguage: (lang: SupportedLangs) => Promise<void>
  isChanging: boolean // 语言切换中状态
}
console.log(i18n.language)

export const useLocaleStore = create<LocaleState>()(
  immer((set) => {
    return {
      language: (i18n.language || 'zh-CN') as SupportedLangs,
      isChanging: false,

      setLanguage: async (lang: SupportedLangs) => {
        set((state) => {
          state.isChanging = true
        })

        try {
          // 切换i18n语言

          await i18n.changeLanguage(lang)

          set((state) => {
            state.language = lang
            state.isChanging = false
          })

          // 可选：同步到后端保存用户偏好
          // await updateUserLanguagePreference(lang)

          // 可选：刷新页面以确保所有组件更新
          window.location.reload()
        } catch (error) {
          console.error('Language change failed:', error)
          set((state) => {
            state.isChanging = false
          })
        }
      },
    }
  }),
)
