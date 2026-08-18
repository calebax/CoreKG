import config from '@/config'
import { chat, chatFile } from './chat'
import { send } from './request'

/** 知识库问答
 * session_type: forest, file_list, excel_list, react_excel_list
 */
export const createSession = (body) => send('chat.NewChatSession', body)
export const getSessionHistory = (body) => send('chat.ListChatSession', body)
export const getSessionChats = (body) => send('chat.ListSessionChats', body)

export const createStream = (body) => send('chat.NewChatQuestionStream', body)
export const sendStream = (body) => {
  return chat('chat.ChatQuestionStream', body)
}
export const updateSessionName = (id, name) =>
  send('chat.RenameChatSession', { id: id, name })
export const deleteSession = (id) => send('chat.RemoveChatSession', { id: id })

export const getSessionInfo = (body) => send('chat.GetSessionInfo', body)
/** 继续上一次对话 */
export const continueLastChat = (body) => chat('chat.ChatGetMessage', body)
export const stopChat = (body) => send('chat.StopChat', body)

/** 获取当前用户的版本 */
export const getCompanyQuota = (body) => send('forest.GetCompanyQuota', body)
/** 发送版本升级表单的手机验证码 */
export const sendUpgradeFormCode = (body) =>
  send('forest.VersionUpgradeSendCode', body)
/** 提交版本升级表单 以及联系售前表单 */
export const submitUpgradeForm = (body) =>
  send('forest.VersionUpgradeVerify', body)

/** 获取问题详情 */

export const getQuestionInfo = (question_id) =>
  send('chat.GetQuestionInfo', { question_id })

// 获取最近使用的智能体
export const getLatestAgents = (body) => send('chat.GetLatestAgents', body)
