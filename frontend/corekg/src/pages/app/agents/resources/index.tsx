import { FC, useEffect, useRef, useCallback } from 'react'
import { Spin } from 'antd'
import useLocalStore from '@/stores/local'
import { useLocaleStore } from '@/stores/locale'
import { toLogin } from '@/api/request'

const AGENT_HOME_STATUS_KEY = 'agent_resources_is_home'
const AGENT_NAVIGATE_HOME_EVENT = 'agent-resources-navigate-home'
const COZE_STATUS_MESSAGE_TYPE = 'COZE_ROUTE_STATUS'
const COZE_READY_MESSAGE_TYPE = 'I_AM_READY'
const AI_SYNC_MESSAGE_TYPE = 'SYNC_CONTEXT'
const AI_NAVIGATE_HOME_MESSAGE_TYPE = 'AI_NAVIGATE_AGENT_HOME'
const AUTH_ERROR_MESSAGE_TYPE = 'AUTH_ERROR'
const REQUEST_RELOGIN_MESSAGE_TYPE = 'REQUEST_RELOGIN'

const AgentResources: FC = () => {
  const { token, userInfo, uinList } = useLocalStore()
  const language = useLocaleStore((state) => state.language)
  const iframeRef = useRef<HTMLIFrameElement>(null)

  const currentOrg = uinList.find(
    (u) => String(u.id) === String(userInfo.uinId),
  )
  const subjectId = currentOrg?.subjectId ?? ''
  // 本地开发：iframe 直连 Coze 服务，避免经 Vite 代理时静态资源路径错乱
  const cozeOrigin = import.meta.env.DEV
    ? import.meta.env.VITE_WORKFLOW_URL || 'http://localhost:8088'
    : window.location.origin
  const iframeUrl = subjectId
    ? `${cozeOrigin}/coze/space/${subjectId}/develop`
    : 'about:blank'
  const iframeOrigin = new URL(iframeUrl).origin

  // 1. 定义发送消息的逻辑
  const syncDataToChild = useCallback(() => {
    if (!iframeRef.current?.contentWindow) return

    const message = {
      type: AI_SYNC_MESSAGE_TYPE,
      payload: {
        token,
        userInfo: {
          ...userInfo,
          // 适配子应用可能需要的字段映射
          user_id_str: userInfo.id,
          locale: language === 'en-US' ? 'en-US' : 'zh-CN',
        },
        // 如果子应用需要知道当前组织信息，也可以传过去
        currentOrg: currentOrg,
      },
    }

    // 目标源固定为子应用地址，保障安全
    iframeRef.current.contentWindow.postMessage(message, iframeOrigin)
  }, [iframeOrigin, token, userInfo, currentOrg, language])

  const navigateIframeToHome = useCallback(() => {
    if (!iframeRef.current?.contentWindow) return

    // 这里先发指令给 coze 子应用，由子应用自己决定如何无刷新回到首页
    iframeRef.current.contentWindow.postMessage(
      {
        type: AI_NAVIGATE_HOME_MESSAGE_TYPE,
        payload: {
          homePath: subjectId ? `/space/${subjectId}/develop` : '',
          iframeUrl,
        },
      },
      iframeOrigin,
    )
  }, [iframeOrigin, iframeUrl, subjectId])

  // 2. 监听 token 或 userInfo 变化，实时同步给子页面
  useEffect(() => {
    syncDataToChild()
  }, [syncDataToChild])

  useEffect(() => {
    // 外层进入智能体页时先把状态置为 unknown，等待 coze 子应用主动上报
    sessionStorage.setItem(AGENT_HOME_STATUS_KEY, 'unknown')
  }, [iframeUrl])

  // 3. 监听子页面的 READY 信号（处理 Iframe 刚加载完成的情况）
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.origin !== iframeOrigin) return

      if (event.data?.type === COZE_READY_MESSAGE_TYPE) {
        syncDataToChild()
      }

      if (event.data?.type === COZE_STATUS_MESSAGE_TYPE) {
        const isHome = event.data?.payload?.isHome
        if (typeof isHome === 'boolean') {
          sessionStorage.setItem(AGENT_HOME_STATUS_KEY, String(isHome))
        }
      }

      if (
        event.data?.type === AUTH_ERROR_MESSAGE_TYPE ||
        event.data?.type === REQUEST_RELOGIN_MESSAGE_TYPE
      ) {
        console.warn(
          '[AgentResources] iframe reported auth issue:',
          event.data,
        )
        toLogin()
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [iframeOrigin, syncDataToChild])

  useEffect(() => {
    // 左侧菜单点击“智能体”时，由外层发一个自定义事件过来，这里再转发给 coze iframe
    window.addEventListener(AGENT_NAVIGATE_HOME_EVENT, navigateIframeToHome)
    return () =>
      window.removeEventListener(
        AGENT_NAVIGATE_HOME_EVENT,
        navigateIframeToHome,
      )
  }, [navigateIframeToHome])

  return (
    <div className='w-full h-full flex flex-col bg-white'>
      {!token ? (
        <div className='w-full h-full flex items-center justify-center'>
          <Spin tip='正在加载用户信息...' />
        </div>
      ) : (
        <iframe
          ref={iframeRef}
          src={iframeUrl}
          onLoad={syncDataToChild}
          title='Resources Library'
          className='w-full flex-1 border-0'
          sandbox='allow-scripts allow-same-origin allow-forms allow-popups'
        />
      )}
    </div>
  )
}

export default AgentResources
