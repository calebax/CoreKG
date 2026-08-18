import { produce } from 'immer'
import { create } from 'zustand'
import { persist, PersistOptions } from 'zustand/middleware'
import config from '@/config'
import { fetchOrganizationProfile } from '@/api/organization'

interface UserInfo {
  id: string
  name: string
  avatar: string
  uinId: string | number
  loginWay?: number
}

export interface UinInfo {
  id: string
  role: string
  uinName: string
  companyName: string
  companyStatus: string
  subjectType: string
  subjectId: string
  uinStatus: string
  logo?: string
  companyUserId?: number
}

interface OrganizationUpdatePayload {
  name?: string
  logo?: string
}

interface LoginPayload {
  token: string
  userInfo: UserInfo
  uinList: UinInfo[]
}

interface LocalStore {
  token: string
  userInfo: UserInfo
  uinList: UinInfo[]
  sidebarCollapsed: boolean
  setToken: (payload: string) => void
  setUserInfo: (payload: UserInfo) => void
  setLogin: (payload: Partial<LoginPayload>) => void
  setLogout: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  updateCurrentOrganization: (payload: OrganizationUpdatePayload) => void
}

type LocalStoreState = Omit<
  LocalStore,
  | 'setToken'
  | 'setUserInfo'
  | 'setLogin'
  | 'setLogout'
  | 'setSidebarCollapsed'
  | 'updateCurrentOrganization'
>

const persistOptions: PersistOptions<LocalStore, LocalStoreState> = {
  name: 'ai-yygu-local-storage',
}

const useLocalStore = create<LocalStore>()(
  persist((set) => {
    return {
      token: '',
      setToken: (payload: string) => {
        set({
          token: payload,
        })
      },
      uinList: [],
      userInfo: {
        id: '',
        name: '',
        avatar: '',
        uinId: '',
      },
      setUserInfo: (payload: UserInfo) => {
        set({
          userInfo: payload,
        })
      },

      sidebarCollapsed: true,
      setSidebarCollapsed: (collapsed: boolean) => {
        set({
          sidebarCollapsed: collapsed,
        })
      },

      setLogin: (payload) => {
        set((prev) => ({
          ...prev,
          ...payload,
        }))
      },
      setLogout: () => {
        set({
          token: '',
          userInfo: {
            id: '',
            name: '',
            avatar: '',
            uinId: '',
          },
          uinList: [],
        })
        if (config.env === 'development') {
          console.log('This is a local environment')
        } else {
          window.location.href = `${location.origin}/login`
        }
      },
      updateCurrentOrganization: (payload: OrganizationUpdatePayload) => {
        set((state) => {
          const updatedUinList = state.uinList.map((uin) => {
            if (String(uin.id) !== String(state.userInfo.uinId)) {
              return uin
            }
            return {
              ...uin,
              companyName: payload.name ?? uin.companyName,
              logo: payload.logo ?? uin.logo,
            }
          })

          return {
            uinList: updatedUinList,
          }
        })
      },
    }
  }, persistOptions),
)

// 导出钩子用于React组件
export default useLocalStore
