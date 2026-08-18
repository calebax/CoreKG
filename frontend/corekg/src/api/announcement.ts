import { send } from './request'

// 获取发版公告列表
export const listAnnouncement = (body: {
  Limit?: number
  Offset?: number
  OrderBy?: string[]
}) => send('forest.ListAnnouncement', body)

// 消息与通知相关类型
export interface MessageItem {
  id: string
  read_status: 'unread' | 'read'
  template_type: 'announcement' | 'system'
  title: string
  content: string
  created_at: string
  route_path?: string
}

// 获取消息列表
export const listMessages = async (body: {
  limit: number
  offset: number
  OrderBy: string[]
  Filters: Array<{ field: string; value: string[] }>
}): Promise<{ data: MessageItem[]; total: number }> => {
  return send('forest.ListMessage', body)
}

// 标记消息为已读
export const readMessages = async (body: {
  message_id: number
  status: 'read' | 'unread'
}) => {
  return send('forest.SetMessageStatus', body)
}

// 删除消息
export const deleteMessages = async (body: {
  message_ids?: number[]
  delete_all?: boolean
}) => {
  return send('forest.DeleteMessages', body)
}
