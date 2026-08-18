/** 单次提问所需的数据 */
export type QAData = {
  text: string
  images?: string[]
}

/** 会话数据(不包括id) */
export type SessionInfo = {
  /** 选中的知识库、文件等 */
  knowledge: { id: number; name: string }[]
  /** 选中项的类型 */
  type:
    | 'file_list'
    | 'forest'
    | 'excel_list'
    | 'react_excel_list'
    | 'db_list'
    | 'db_table_list'
  /** 大模型id */
  model: number
  /** 大模型名称 */
  modelName?: string
}

/** 初始化会话所需的数据 */
export type DialogInitData = QAData & SessionInfo
