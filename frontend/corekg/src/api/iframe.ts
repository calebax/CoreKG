import { iframeSend, iframeChat } from './iframeRequest'

// iframe认证类
class IframeAuth {
  private token: string = ''
  private clientId: string = ''
  private agentId: string = ''

  // 初始化认证
  async init(agentId: string, clientId: string) {
    try {
      // 调用获取token接口（无需认证）
      const response = await iframeSend('chat.ExternalToken', {
        access_token: agentId,
      })
      this.token = response.token
      this.clientId = clientId // 使用外部传入的 clientId
      this.agentId = agentId
      return { token: this.token, clientId: this.clientId }
    } catch (error) {
      console.error('iframe认证失败:', error)
      throw error
    }
  }

  setClientId(clientId: string) {
    this.clientId = clientId
  }

  getToken(): string {
    return this.token
  }

  getClientId(): string {
    return this.clientId
  }

  clear() {
    this.token = ''
    this.clientId = ''
    this.agentId = ''
  }

  isAuthenticated(): boolean {
    return !!this.token && !!this.clientId
  }
}

export const iframeAuth = new IframeAuth()

export const getIframeAgentInfo = async (client_id: string) => {
  // 此函数暂未使用，保留接口定义
  return iframeSend(
    'chat.GetExternalHistory',
    { client_id },
    iframeAuth.getToken(),
  )
}

export const getIframeHistory = async () => {
  // 智能机器人弹窗中不需要调用历史接口，直接返回空数据
  // 保持数据结构一致：messages 需要包含 data 和 agent 信息
  return {
    messages: {
      data: [], // 历史对话数据为空
      session_name: '客服小助手', // 默认会话名称
      greet_message:
        '👋👋您好，欢迎来到 CoreKG，我是您的AI助手，有什么可以帮助您的吗？我将随时为您解答使用中的问题并提供支持。', // 默认问候语
    },
  }
  // try {
  //   const response = await iframeSend(
  //     'chat.GetExternalHistory',
  //     {
  //       client_id: iframeAuth.getClientId(),
  //     },
  //     iframeAuth.getToken(),
  //   )
  //
  //   return {
  //     messages: response || { data: [] },
  //   }
  // } catch (error) {
  //   console.error('获取历史失败:', error)
  //   return { messages: { data: [] } }
  // }
}

export const createIframeStream = async (question: string) => {
  return iframeSend(
    'chat.NewExternalChatStream',
    {
      client_id: iframeAuth.getClientId(),
      question,
    },
    iframeAuth.getToken(),
  )
}

export const createIframeStreamWithInput = async (
  input: Array<{ title: string; value: string; name: string }>,
) => {
  return iframeSend(
    'chat.NewExternalChatStream',
    {
      client_id: iframeAuth.getClientId(),
      input,
    },
    iframeAuth.getToken(),
  )
}

export const sendIframeStream = async (
  streamKey: string,
  question?: string,
): Promise<Response | { ok: boolean; body: ReadableStream }> => {
  return iframeChat(
    'chat.SubmitExternalChatStream',
    {
      client_id: iframeAuth.getClientId(),
      stream_key: streamKey,
    },
    iframeAuth.getToken(),
  )
}
