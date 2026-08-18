declare module 'Agent' {
  /** 智能体全部信息 */
  export type Agent = BasicAgentInfo & AgentAdminDetail

  /** 智能体基础信息 */
  export type BasicAgentInfo = {
    id: number
    avatar: string
    title: string
    description: string
    isAdmin: boolean
    /** 收藏 */
    favorite: boolean
    /** 指令型 角色型 知识库型 工作流型 */
    type: 'prompt' | 'role_play' | 'knowledge' | 'workflow'

    /** 系统应用 用户应用 */
    source: 'system' | 'custom'
    /** 仅系统应用 热门/推荐 */
    tag: 'popular' | 'recommend'
    /** 仅用户应用 已发布/草稿 */
    status: 'published' | 'draft'
    /** 是否已同步至Coze */
    is_synced?: boolean
    coze_workflow_id?: any
    coze_space_id?: any
    CreatedAt: string
    UpdatedAt: string
  }

  /** 用于编辑和管理的信息 只对管理员开放 */
  export type AgentAdminDetail = {
    /** 管理员 */
    manager_ids: number[]
    /** 大模型 */
    chat_models: {
      id: number
      name: string
      description: string
    }[]
    /** 提示词 */
    prompt_template: string
    /** 大模型温度 */
    temperature: number

    /** 仅指令型应用 参数 */
    params: {
      /** 中文名称 */
      name: string
      description: string
      is_required: boolean
      /** 实际输入 */
      input: string
      input_type: 'text' | 'select'
      /** input_type为select时 下拉数组 */
      input_array: string[]
    }[]

    /** 角色型 知识库型 问候语 */
    greeting_message: string

    /** 仅知识库型 挂载的知识库 */
    forests: {
      id: number
      name: string
      forest_type: 'file' | 'qa' | 'cad' | 'data'
    }[]
    /** 公开范围：公司 自定义 其余与前端无关 */
    public_scope: 'company' | 'custom' | 'private' | 'public'
    /** 权限系统的冗余 与公开范围的关系 company-company user-custom */
    scope_type: 'company' | 'user'
    /** public_scope为自定义时 需要选公开的员工 */
    scope_ids: number[]
  }
}
