import { send } from '@/api/request'

export const getAllPackage = () =>
  send('forest.ListPackage', {}) as Promise<{
    list: {
      additional_notes: string[]
      agent_quota: number
      /** 写作空间额度（套餐权益） */
      article_quota: number
      description: string
      disk_quota: number
      employee_quota: number
      is_purchased: boolean
      name: string
      package_id: number
      price: number
      qa_quota: number
      sale_price: number
      edition: 'free_trail' | 'professional'
    }[]
  }>

/**
 * 创建套餐订单，返回订单信息
 */
export const createOrder = (data: { package_id: number }) =>
  send('forest.CreateOrder', data) as Promise<{
    /** 时间戳 */
    expire_time: string
    order_sn: string
    pay_url: string
  }>

/**
 * 查询订单状态
 */
export const queryOrderStatus = (data: { order_sn: string }) =>
  send('forest.QueryOrderStatus', data) as Promise<{
    status: 'pending' | 'success' | 'closed'
  }>

/** 列表查询支付订单记录 */
export const getPaymentRecord = (data: CommonArgs) =>
  send('forest.ListPaymentOrderRecord', data) as Promise<{
    data: {
      id: number
      order_sn: string
      amount: number
      company_id: number
      uin: number
      status: 'pending' | 'success' | 'failed'
      created_at: string
      paid_at: string
    }[]
    limit: number
    offset: number
    total: number
  }>
