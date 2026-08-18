import { send } from './request'

/** 获取当前组织成员 */
export { listEmployee as getEmployee } from '@/api'
/** 删除组织成员 */
export const deleteEmployee = (data: {
  delete_reason?: string
  employee_id: number
}) => send('account.DeleteEmployeePrivate', data)
/** 删除组织成员(私有化) */
export const deleteEmployeePrivate = (data: {
  delete_reason?: string
  employee_id: number
}) => send('account.DeleteEmployeePrivate', data)
/** 单个知识库的权限信息 */
export type ForestPermItem = {
  forest: {
    ID: number
    public_scope?: 'public' | 'private' | 'company' | 'custom'
  }
  manage_perm: boolean
  use_perm: boolean
}
/** 单个智能体的权限信息 */
export type AgentPermItem = {
  agent: {
    ID: number
    public_scope?: 'public' | 'private' | 'company' | 'custom'
  }
  manage_perm: boolean
  use_perm: boolean
}

/** 获取知识库权限 */
export const getForestPermSet = (data: {
  uin?: number
}): Promise<{
  perm_set?: (ForestPermItem & {
    forest: { name: string }
  })[]
}> => send('forest.GetForestPermSet', data) as any
/** 修改知识库权限 */
export const modifyForestPermSet = (data: {
  uin: number
  perm_set: (ForestPermItem & {
    act_option: 'update'
  })[]
}): Promise<unknown> => send('forest.ModifyForestPermSet', data)

/** 获取智能体权限 */
export const getAgentPermSet = (data: {
  uin?: number
}): Promise<{
  perm_set?: (AgentPermItem & {
    agent: { show_name: string }
  })[]
}> => send('chat.GetAgentPermSet', data) as any
/** 修改智能体权限 */
export const modifyChatPermSet = (data: {
  uin: number
  perm_set: (AgentPermItem & {
    act_option: 'update'
  })[]
}): Promise<unknown> => send('chat.ModifyChatPermSet', data)

/** 根据权限获取key 用于邀请链接 */
export const getBindCompanyKeyWithPermSet = (data: {
  perm_set: {
    chatPs: (AgentPermItem & {
      act_option: 'update'
    })[]
    forestPs: (ForestPermItem & {
      act_option: 'update'
    })[]
  }
  count: 1
  invitation_role: 'sys_employee' | 'sys_admin'
  department_ids?: number[]
  issuer: 'yygu'
}): Promise<{
  key: string
}> => send('account.GetBindCompanyKeyWithPermSet', data) as any
/** 根据key 将当前用户添加进组织并给予权限 */
export const bindCompanyWithPermSet = (data: {
  key: string
  domain_name: string
  way: string
  code?: string // 微信登录时需要
  username?: string // 账号密码登录时需要
  password?: string // 账号密码登录时需要
}): Promise<any> => send('account.BindCompanyWithPermSet', data)

export const getInviteInfo = (data: { key: string }): Promise<any> =>
  send('account.GetInviteInfo', data)
