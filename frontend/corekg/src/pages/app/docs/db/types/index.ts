// 文件项接口
export interface FileItem {
  id: number
  name: string
  size: string
  updatedAt: string // 显示用的日期格式（如：2025-05-28）
  createdAtFull?: string // 完整的时间戳，用于排序（如：2025-05-28T09:39:16.94+08:00）
  isFolder: boolean
  fileType?: string
  isEditing?: boolean
  isNewFolder?: boolean
  /** 不支持 未开始 已完成 解析中 失败*/
  file_status: 'unsupported' | 'pending' | 'success' | 'running' | 'fail'
  /** 如果状态为解析中 展示这个 */
  file_progress: string
}

// 排序项接口
export interface SortItem {
  field: string
  order: 'ascend' | 'descend'
}

// 过滤条件接口
export interface FilterItem {
  field: string
  value: string[]
  exactMatch?: boolean
}

// 表格参数接口
export interface TableParams {
  currentPage: number
  pageSize: number
  sorts: SortItem[]
  selectedRowKeys?: React.Key[] // 所有跨页选中的行键（全局选中状态）
}

// 本地存储键名
export const STORAGE_KEY = 'ai-yygu-local-storage'
