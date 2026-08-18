import type { AxiosRequestConfig } from 'axios'
import { upload, send } from './request'

export const uploadImage = (
  body: {
    file: File
    purpose?: string
  },
  config?: AxiosRequestConfig,
) =>
  upload('chat.UploadImage', body, config) as Promise<{
    url?: string
  }>

/** 大模型模式上传附件接口 */
export const uploadAttachment = (
  body: {
    file: File
    purpose?: string
  },
  config?: AxiosRequestConfig,
) =>
  upload('chat.UploadAttachment', body, config) as Promise<{
    url?: string
    id?: string
    md_url?: string
  }>

/** 获取公司配额等通用信息 需要登录 */
export const getCommonInfo = (config: AxiosRequestConfig) =>
  send('forest.GetCommonInfo', {}, config) as Promise<{
    company_quota: {
      agent_quota: number
      agent_quota_used: number
      /** 写作空间额度 */
      article_quota: number
      article_quota_used: number
      disk_quota: number
      disk_quota_used: number
      employee_quota: number
      employee_quota_used: number
      qa_quota: number
      qa_quota_used: number
      is_purchased: boolean
      company_quota: number
    }
  }>

/** 获取消息通知数量 */
export const getMessageCount = (
  filters?: Array<{ field: string; value: string[] }>,
) => {
  return send('forest.GetMessageCount', { filters }) as Promise<{
    count: number
  }>
}
