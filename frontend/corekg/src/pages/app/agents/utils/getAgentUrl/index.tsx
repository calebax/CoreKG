import { Agent } from 'Agent'
import { match, P } from 'ts-pattern'

/**
 * 获取智能体url
 * @param id 应用id
 * @param type 智能体类型
 * @param edit 传入true会跳转到编辑界面
 */
export const getAgentUrl = (
  id: number,
  type: Agent['type'],
  edit?: boolean,
  cozeConfig?: {
    toCoze?: boolean
    coze_workflow_id: any
    coze_space_id: any
  },
) => {
  const pageType = edit ? 'edit' : 'detail'
  // 知识库型挂在角色型下
  const agentType = match(type)
    .with('prompt', () => 'prompt' as const)
    .with('workflow', () => 'workflow' as const)
    .with(P.union('knowledge', 'role_play'), () => 'role' as const)
    .exhaustive()
  if (pageType === 'edit') return `/agents/edit/${id}`

  const { coze_workflow_id, coze_space_id, toCoze } = cozeConfig ?? {}
  if (agentType === 'workflow' && toCoze) {
    return `/coze/work_flow?workflow_id=${coze_workflow_id}&space_id=${coze_space_id}`
  }
  return `/agents/${pageType}/${agentType}/${id}`
}
