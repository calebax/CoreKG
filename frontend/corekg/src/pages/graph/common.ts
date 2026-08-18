//// 图谱

export type GraphBaseInfo = {
  /** 图谱id */
  id: number
  name: string
  description?: string
  updateAt: string
  CreatedAt: string
  avatar_url: string
  status: 'draft' | 'pending' | 'running' | 'success' | 'failed' | 'updatable'
  /** 解析模式 */
  parse_mode: 'auto' | 'rule'
  /** 唯一绑定的知识库 */
  forest_id: number
  /** 草稿状态不会有这些数据 */
  totalNodes: number
  totalRelationships: number
  /** 是否管理员 */
  is_admin: boolean
  /** 任务总数 */
  task_count?: number
  /** 成功任务数 */
  success_task_count?: number
}
/** 图谱模板 */
export type GraphTemplate = {
  avatar?: string
  name: string
  description: string
  tags: GraphTag[]
  edges: GraphTagRelationship[]
}

//// 实体类型

/** 实体类型 */
export type GraphTag = {
  tag_id?: number
  tag_name: string
  description?: string
  properties?: Property[]
}

/** 实体类型的属性 */
export type Property = {
  name: string
  comment?: string
  type: 'int64' | 'bool' | 'string' | 'double'
  /** 默认值类型与type有关 */
  defaults?: any
}

/** 实体类型间的关系 */
export type GraphTagRelationship = {
  name: string
  /** 实体类型的tag_name 可以相同 */
  source: string
  target: string
}

/** 通过类型获取相应的展示名称 */
export const PropertyLabelMap: Record<Property['type'], string> = {
  string: '字符串',
  int64: '整数',
  double: '浮点数',
  bool: '布尔值',
}
//// 实体

/** 实体 */
export type GraphNode = {
  id: string
  name: string
  /** 实体的tag */
  tags: string[]
  properties?: { key: string; value: string }[]
}
/** 实体间的关系 */
export type GraphNodeRelationship = {
  id: string
  name: string
  /** 实体的name tag */
  source: string
  target: string
}
