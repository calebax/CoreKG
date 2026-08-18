import { GraphTag, GraphTagRelationship } from 'Graph'

export type GraphEditValue = {
  id: number
  /** 已选择的模板 */
  type?: string
  /** 编辑实体类型 */
  graphTags?: {
    tags: GraphTag[]
    relationships: GraphTagRelationship[]
  }
  /** 已选中的文件id */
  files?: number[]
}
