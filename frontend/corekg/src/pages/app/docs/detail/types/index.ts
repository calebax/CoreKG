// 标签项接口
export interface TagItem {
  id: number
  resource_id: number
  resource_type: string
  tag_id: number
  tag_name: string
  tag_group_name: string
}

// 文件项接口
export interface FileItem {
  id: number
  name: string
  size: string
  updatedAt: string // 显示用的日期格式（如：2025-05-28）
  createdAtFull?: string // 完整的创建时间戳，用于排序（如：2025-05-28T09:39:16.94+08:00）
  updatedAtFull?: string // 完整的更新时间戳，用于排序（如：2025-05-28T09:39:16.94+08:00）
  isFolder: boolean
  fileType?: string
  isEditing?: boolean
  isNewFolder?: boolean
  /** 不支持 未开始 已完成 解析中 失败 */
  file_status: 'unsupported' | 'pending' | 'success' | 'running' | 'fail'
  /** 如果状态为解析中 展示这个 */
  file_progress: string
  ext?: string
  // 启用问答
  enable: boolean
  // 标签列表
  tag_list?: TagItem[] | null
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
export const STORAGE_KEY = 'ai-yygu-table-storage'
