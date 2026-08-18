import config from '@/config'
import { send } from './request'

// 获取apiKey数据列表

export const getApiKeyList = (data: {
  limit: number
  offset: number
  orderBy: Array<string>
}) => send('account.ListAPIKey', data)

// 创建apiKey
export const createApiKey = (data: { name: string; expired_at: string }) =>
  send('account.CreateAPIKey', data)

// 删除apiKey
export const deleteApiKey = (data: { id: number }) =>
  send('account.DeleteAPIKey', data)

// agent
export const createAgentApiKey = (data: { agent_id: number }) =>
  send('account.CreateAgentApiKey', data)
export const deleteAgentApiKey = (data: {
  agent_id: number
  apikey_id: number
}) => send('account.DeleteAgentApikey', data)
export const getAgentApiKeyList = (data: {
  limit: number
  offset: number
  orderBy: Array<string>
  filters?: Array<{ field: string; value: string[]; exactMatch?: boolean }>
}) => send('account.ListAgentAPIKey', data)
export const setAgentApiKeyStatus = (data: {
  agent_id: number
  apikey_id: number
  status: string | 'normal'
}) => send('account.SetAgentApiKeyStatus', data)

/** 私有化 创建成员 */
export const createEmployee = (data: {
  email: string
  password: string
  phone?: string
  user_name: string
}) => send('account.CreateEmployee', data)

/** 私有化 编辑成员 */
export const editEmployee = (data: {
  email: string
  password?: string
  phone?: string
  user_name: string
  uin: number
}) => send('account.EditEmployee', data)
