import config from '@/config'
import { chat, chatFile } from './chat'
import { send } from './request'

// 模型列表
export { listCustomModel as listModel }
// 获取模型详情
export const getModelDetail = (sessionId) =>
  send('chat.GetChatModelDetail', { id: sessionId })

// agent
export const listAgent = (body) => send('chat.ListChatAgent', body)
export const createAgent = (body) => send('chat.CreateAgentApp', body)
export const updateAgent = (body) => send('chat.UpdateAgentApp', body)
export const deleteAgent = (id) => send('chat.DeleteChatAgent', { id })
export const updateChatAgent = (body) => send('chat.UpdateChatAgent', body)
/** 根据session_id获取信息 */
export const getSessionInfo = (body) => send('chat.GetSessionInfo', body)
/** 最近智能体 */
export const getLatestAgent = (body) => send('chat.GetLatestAgent', body)

export const updateRoleAgent = (body) => send('chat.UpdateRolePlayAgent', body)
export const updatePromptAgent = (body) => send('chat.UpdatePromptAgent', body)
export const updateAgentPublicScope = (body) =>
  send('chat.UpdateAgentPublicScope', body)
export const listAgentPublicScope = (body) =>
  send('chat.ListAgentPublicScope', body)
export const listEmployee = (body) => send('account.ListEmployeeNickID', body)

export const getRoleAgentDetail = (id) =>
  send('chat.GetRolePlayAgentDetail', { id })
export const getPromptAgentDetail = (id) =>
  send('chat.GetPromptAgentDetail', { id })

export const removeChatSession = (id) => send('chat.RemoveChatSession', { id })
export const renameChatSession = (body) => send('chat.RenameChatSession', body)
export const moveChatSession = (body) => send('chat.MoveSession', body)

export const testRoleAgent = (body) => chat('chat.TestRolePlayAgent', body)
export const testPromptAgent = (body) => chat('chat.TestPromptAgent', body)
export const testWorkflow = (body) => send('chat.WorkflowTestRun', body)
// 收藏
export const collectAgent = (id) => send('chat.CollectApp', { id })
export const listCollectApp = (body) => send('chat.ListCollectApp', body)
// chat
export const createSession = (body) => send('chat.NewChatSession', body)
export const getSessionHistory = (body) => send('chat.ListChatSession', body)
export const getSessionChats = (body) => send('chat.ListSessionChats', body)
export const createStream = (body) => send('chat.NewChatQuestionStream', body)
export const sendStream = (body) => chat('chat.ChatQuestionStream', body)
export const setTopChatSession = (id) => send('chat.SetTopChatSession', { id })

// 智能判题
export const reviewQuestion = (body) =>
  chatFile('agent_cj.ReviewQuestion', body)
export const getReviewQuestion = (body) =>
  send('agent_cj.ReviewQuestionList', body)

/** 获取此agent 不受权限控制 信息少 */
export const getAgentInfo = (id) => send('chat.GetAgentInfo', { id })
/** 受权限控制 信息多 */
export const getAgentDetail = (id) => send('chat.GetAgentWithPerm', { id })
/** 发布新应用或更新应用 不区分类型 */
export const publishAgent = (body) => send('chat.UpdateAgentWithPerm', body)
/** 测试指令型应用 */
export const testPromptTypeAgent = (body) => chat('chat.TestAgent', body)
/** 测试知识库应用 */
export const testForestTypeAgent = (body) => chat('chat.TestAgent', body)
/** 测试角色型应用 */
export const testRoleTypeAgent = (body) => chat('chat.TestAgent', body)

// iframe 是否打开
export const getExternalStatus = (agent_id) =>
  send('chat.GetExternalStatus', { agent_id })

export const setExternalStatus = (agent_id, status) =>
  send('chat.SetExternalStatus', { agent_id, status })

// 配置模型
/** 创建模型 */
export const createCustomModel = (body) => send('chat.CreateModel', body)
/** 删除模型 */
export const deleteCustomModel = (body) => send('chat.DeleteModel', body)
/** 模型列表 */
export const listCustomModel = (body) => send('chat.ListModel', body)
/** 修改模型 */
export const updateCustomModel = (body) => send('chat.UpdateModel', body)
/** 测试模型 */
export const testCustomModel = (body) => send('chat.ModelTest', body)
/** 获取模型详细信息用于编辑 */
export const getCustomModelDetail = (body) => send('chat.GetModelDetail', body)

// coze
export const getCoze = (body) => send('account.LoginCoze', body)

/** 同步coze */
export const syncCoze = (body) => send('chat.CreateCozePlugin', body)
