import { GraphBaseInfo } from 'Graph'

export type GraphMetaValues = Pick<GraphBaseInfo, 'name' | 'description'> & {
  /** 权限信息 */
  manager_ids: number[]
  public_scope: 'company' | 'custom'
  scope_ids?: number[]
}