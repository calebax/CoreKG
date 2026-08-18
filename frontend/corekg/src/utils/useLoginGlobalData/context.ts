import { createContext, useContext } from 'react'
import type { GlobalContextValue } from '.'

export const LoginGlobalContext = createContext<GlobalContextValue | null>(null)

/** 获取登陆后可用的全局数据 */
export const useLoginGlobalData = () => {
  const contextValue = useContext(LoginGlobalContext)
  if (!contextValue) throw new Error('LoginGlobalContext')
  return contextValue
}
