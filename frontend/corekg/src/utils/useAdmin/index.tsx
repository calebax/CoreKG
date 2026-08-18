import { createContext } from 'react'
import { useLoginGlobalData } from '../useLoginGlobalData'

export type Admin = {
  employee_id: number
  uin: number
  user_id: number
  user_name: string
}

export type AdminContextValue = {
  admin: Admin[]
  adminIds: number[]
  refresh: () => void
}

export const AdminContext = createContext<AdminContextValue | null>(null)

/** 当前公司管理员 */
// eslint-disable-next-line react-refresh/only-export-components
export const useAdmin = () => {
  return useLoginGlobalData().admin
}
