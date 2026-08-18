export {}
declare global {
  type Style = {
    className?: string
    style?: React.CSSProperties
  }
  type ValueController<T> = {
    value?: T
    onChange?: (value?: T) => void
  }
  /**
   * 通用接口参数\
   * 分页 筛选项 起止时间 排序等
   */
  type CommonArgs = {
    limit?: number
    offset?: number
    filters?: {
      field: string
      value: string[]
      exactMatch?: boolean
    }[]
    beginTime?: string
    endTime?: string
    orderBy?: string[]
    listAll?: boolean
  }

  /**
   * 私有化套餐的key
   */
  type CustomPackage = 'agent' | 'all'

  /**
   * 私有化模块的key
   */
  type CustomModule = 'chat' | 'forest' | 'agent' | 'graph'
}
