/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  FC,
  ReactNode,
  useMemo,
  useEffect,
  useRef,
  useState,
} from 'react'
import { App, Tooltip } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { match, P } from 'ts-pattern'
import { getSessionHistory } from '@/api'
import { cn } from '@/utils'
import { type AIDialog, DialogList } from '@/components/dialog'
import { useProject } from '..'
import type { Knowledge } from '..'
import { Charts } from './Charts'
import { EmptyProject } from './EmptyProject'
import { Graph } from './Graph'
import { GraphUpdateMessage } from './GraphUpdateMessage'
import { useKnowledgeData, withKnowledgeDataProvider } from './Knowledge'
import References from './References'
import { SessionChatHeader } from './SessionChat'
import { Dialog } from './SessionChat/Dialog'
import { SessionHistory } from './SessionChat/SessionHistory'
import { EDrawerType, useProjectSection } from './hooks'
import CanvasIcon from './images/canvas.svg?react'
import CloseIcon from './images/close.svg?react'
import HistoryIcon from './images/history.svg?react'
import OpoIcon from './images/opo.svg?react'
import styles from './index.module.scss'
import { useSession } from './useSession'

/** 单次提问所需数据 */
export type QAData = {
  content: string
  images?: string[]
  options?: {
    enable_web_search?: boolean
  }
  input?: {
    attachments?: Array<{
      id?: string
      url?: string
      md_url?: string
      type: string
      name: string
      mime_type?: string
    }>
  }
}

export { type Knowledge }

/**
 * 创建session的配置\
 * 外部数据源和知识库是冲突的\
 */
export type SessionConfig = {
  /** 提问模式 */
  mode: 'model' | 'knowledge' | 'graph' | 'table' | 'database' | 'h3c-test' | 'graph_search' | 'external_data'
  /** 模型id */
  model_id: number
  /** 选中的知识 */
  knowledge?: Knowledge[]
  /** 外部数据源id 后端自动区分类型 */
  externalIds?: number[]
  /** 已选中的具有图谱的知识库 */
  graphKnowledgeBase?: Knowledge[]
  /** 已选中的表格知识库 */
  tableKnowledgeBase?: Knowledge[]
  /** 已选中的数据库知识库 */
  databaseKnowledgeBase?: Knowledge[]
  /** 问答模式提示词key */
  prompt_key?: string
  /** 是否启用联网搜索 */
  enable_web_search?: boolean
  /** 选中的标签id */
  tag_ids?: number[]
  /** 外接数据模式 - 已选中的外部数据源provider列表 */
  externalDataProviders?: string[]
}
export type SessionInfo = {
  /** 等待首次提问 已经提问正在创建 已创建 不展示对话组件 */
  sessionStatus: 'new' | 'creating' | 'created' | 'none'
  setSessionStatus: (val: SessionInfo['sessionStatus']) => void
  dialog: DialogList
  dialogStatus: 'loading' | 'ready' | 'asking' | 'answering'
  sessionConfig?: Partial<SessionConfig>
  /** 增量更新配置.传入的函数会通过immer方式进行更新 */
  setSessionConfig: (
    val: Partial<SessionConfig> | ((draft: Partial<SessionConfig>) => void),
  ) => void
  startQA: (data: QAData) => void
  stopQA: () => void
  /** 当前活跃的对话索引 */
  dialogIndex: number
  handleOpenReference: (index: number) => void
  handleOpenGraph: (index: number) => void
  handleOpenChart: () => void
  handleCloseDrawer: () => void
  drawerVisible: boolean
  drawerType: EDrawerType
  sessionVisible: boolean
}
export const SessionContext = createContext<SessionInfo | null>(null)

export const useSessionInfo = () => {
  const context = useContext(SessionContext)
  if (!context) {
    throw new Error('useSessionInfo 必须在 SessionContext.Provider 内部使用')
  }
  return context
}

export const ProjectContent: FC = withKnowledgeDataProvider(() => {
  const { t } = useTranslation('pages')
  const { message } = App.useApp()
  const { defaultKnowBase, isOtherPage, isUngroupedSession } = useProject()
  const {
    project_id,
    session_id,
    setSessionId,
    data: { sessions, charts },
    type,
  } = useProject()
  const { knowledgeList, loading } = useKnowledgeData()

  // 初始化左右section的拖拽功能hooks
  const {
    sessionWidth,
    sessionResizing,
    sessionVisible,
    handleOpenSession,
    handleCloseSession,
    drawerWidth,
    drawerResizing,
    drawerVisible,
    drawerType,
    handleOpenDrawer,
    handleCloseDrawer,
    dialogIndex,
    setDialogIndex,
    handleDrawerWidthChange,
    handleSessionWidthChange,
  } = useProjectSection()

  const handleOpenReference = (index: number) => {
    handleOpenDrawer(EDrawerType.REFERENCE)
    setDialogIndex(index)
  }

  const handleOpenGraph = (index: number) => {
    handleOpenDrawer(EDrawerType.GRAPH)
    setDialogIndex(index)
  }

  // 当切换会话时，关闭引用资源侧边栏
  const prevSessionIdRef = useRef(session_id)
  useEffect(() => {
    // 新会话改变url
    if (prevSessionIdRef.current) {
      handleCloseDrawer()
    }
    prevSessionIdRef.current = session_id
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session_id])

  // 对于未分组会话，如果有 session_id，不应该认为 projectEmpty 为 true
  // 这样可以避免在跳转到会话页面时显示空状态
  const projectEmpty = useMemo(
    () =>
      isUngroupedSession || session_id
        ? false
        : sessions.length === 0 && charts.length === 0,
    [isUngroupedSession, session_id, sessions.length, charts.length],
  )

  const handleOpenChart = () => {
    handleOpenDrawer(EDrawerType.CHART)
  }

  const {
    sessionStatus,
    setSessionStatus,
    dialog,
    dialogStatus,
    sessionConfig,
    setSessionConfig,
    startQA,
    stopQA,
    resetSession,
  } = useSession(
    projectEmpty,
    handleOpenReference,
    handleOpenChart,
    handleOpenGraph,
  )

  // 单文档问答场景：持久化 session_id 到 sessionStorage
  useEffect(() => {
    if (isOtherPage && defaultKnowBase && session_id && session_id !== 0) {
      const storageKey = `file_session_${defaultKnowBase}`
      sessionStorage.setItem(storageKey, String(session_id))
    }
  }, [session_id, isOtherPage, defaultKnowBase])

  useEffect(() => {
    if (loading === false && knowledgeList.length && defaultKnowBase) {
      const val = knowledgeList.find(
        (item) => item.id === defaultKnowBase && item.knowledgeType === 'file',
      )
      if (val) {
        setSessionConfig({ knowledge: [val] })
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, knowledgeList, defaultKnowBase])

  // 用于触发未分组会话信息重新获取的 key
  const [ungroupedRefreshKey, setUngroupedRefreshKey] = useState(0)

  // 监听未分组会话列表刷新事件，触发会话信息重新获取
  useEffect(() => {
    const handleRefresh = () => {
      // 使用 queueMicrotask 确保在下一个微任务中执行，避免阻塞当前事件循环
      queueMicrotask(() => {
        setUngroupedRefreshKey((prev) => prev + 1)
      })
    }
    window.addEventListener('refreshUngroupedSessions', handleRefresh)
    return () => {
      window.removeEventListener('refreshUngroupedSessions', handleRefresh)
    }
  }, [])

  // 对于未分组会话，需要通过 API 获取会话名称
  const { data: ungroupedSessionInfo } = useRequest(
    async () => {
      if (!isUngroupedSession || !session_id || sessionStatus !== 'created') {
        return null
      }
      try {
        // 使用 getSessionHistory 获取未分组会话列表，然后查找当前会话
        const res = await getSessionHistory({
          limit: 9999,
          offset: 0,
          project_id: -1,
        })
        const session = (res.Data || []).find(
          (item: any) => item.ID === session_id,
        )
        return session?.name || null
      } catch {
        return null
      }
    },
    {
      refreshDeps: [
        isUngroupedSession,
        session_id,
        sessionStatus,
        ungroupedRefreshKey,
      ],
    },
  )

  const sessionLabel = useMemo(() => {
    switch (sessionStatus) {
      case 'none':
        return null
      case 'new':
      case 'creating':
        return t('project.newChat')
      case 'created':
        // 对于未分组会话，使用 API 获取的名称
        if (isUngroupedSession) {
          return ungroupedSessionInfo || undefined
        }
        // 对于分组会话，从 sessions 数组中查找
        return sessions.find((item) => item.session_id === session_id)?.name
    }
  }, [
    sessionStatus,
    session_id,
    sessions,
    t,
    isUngroupedSession,
    ungroupedSessionInfo,
  ])

  const drawerName = useMemo(() => {
    if (drawerType === EDrawerType.CHART) return '仪表盘'
    if (drawerType === EDrawerType.GRAPH) return '图谱洞察'
    if (drawerType === EDrawerType.REFERENCE) {
      const dialogInfo = dialog[dialogIndex] as AIDialog
      const reference = dialogInfo?.reference

      if (!reference?.length) return ''
      const fileSet = new Set<number>()
      const forestSet = new Set<number>()
      reference.forEach((item) => {
        fileSet.add(item.file_id)
        forestSet.add(item.forest_id)
      })

      return (
        (forestSet.size ? '搜索到' + forestSet.size + '个知识库' : '') +
        (fileSet.size ? fileSet.size + '篇资源' : '')
      )
    }
    return ''
  }, [drawerType, dialog, dialogIndex])

  useReset({
    session_id,
    sessionStatus,
    reset: resetSession,
    setSessionStatus,
    dialog,
    type,
  })

  // 当切换项目后再回来时，如果有历史会话记录，自动进入第一个会话
  // 这样避免了"新建会话"状态在切换项目后继续保留的问题
  useEffect(() => {
    if (
      type !== 'global' &&
      sessionStatus === 'none' &&
      !session_id &&
      sessions.length > 0
    ) {
      // 先重置状态，确保 resetKey 更新，然后设置 session_id
      // 这样 useAsyncEffect 能够正确加载会话数据
      resetSession()
      setSessionId(sessions[0].session_id)
    }
  }, [sessionStatus, session_id, sessions, setSessionId, resetSession, type])

  const withProvider = (children: ReactNode) => (
    <SessionContext.Provider
      value={{
        sessionStatus,
        setSessionStatus,
        dialog,
        dialogStatus,
        sessionConfig,
        setSessionConfig,
        startQA,
        stopQA,
        dialogIndex,
        handleOpenReference,
        handleOpenGraph,
        handleOpenChart,
        handleCloseDrawer,
        drawerVisible,
        drawerType,
        sessionVisible,
      }}
    >
      {children}
    </SessionContext.Provider>
  )

  // 顶部的新建会话打开历史记录按钮
  const renderHistoryOperator = () => {
    const isDisabled = sessionStatus !== 'created'

    const handleNewSession = () => {
      if (isDisabled) {
        message.info(t('project.alreadyLatestConversation'))
        return
      }
      setSessionStatus('new')
      setSessionId()
    }

    // 未分组会话模式下不显示历史记录和新建会话按钮
    if (isUngroupedSession) {
      return null
    }

    return (
      <div
        className={cn(
          'flex items-center gap-[8px] pr-[6px] mr-[6px]',
          styles.historyOperator,
        )}
      >
        <Tooltip title='打开历史记录'>
          <HistoryIcon className='cursor-pointer' onClick={handleOpenSession} />
        </Tooltip>
        {/* 左侧按钮 */}
        <div
          className={cn(
            'rounded-[6px]',
            'h-[30px] flex items-center justify-center',
            'gap-[5px] px-[10px]',
            styles.opo_btn,
            isDisabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
          )}
          onClick={handleNewSession}
        >
          <OpoIcon />
          <span className={`${styles.text} whitespace-nowrap`}>
            {/* {match(version)
            .with('saas', () => 'CoreKG AI')
            .with('custom', () => 'AI')
            .with('international', () => t('project.opoAI'))
            .exhaustive()} */}
            {t('project.newSession')}
          </span>
        </div>
      </div>
    )
  }
  // 问答详情页头部
  const renderHeader = () => {
    if (type === 'single-file') return null
    // 知识库问答中的数据库场景也需要展示仪表盘
    const shouldShowDashboard =
      sessionConfig.mode === 'table' ||
      sessionConfig.mode === 'database' ||
      (sessionConfig.mode === 'knowledge' &&
        sessionConfig.knowledge?.some(
          (item) => item.knowledgeType === 'mysql_table',
        ))
    return (
      <div className='relative flex px-4 items-center h-[50px] border-b border-b-[#EFF1F4]'>
        {!sessionVisible && renderHistoryOperator()}
        <div className='flex items-center gap-2'>
          {sessionLabel && (
            <>
              <div className='text-base text-[#000000] font-[500] whitespace-nowrap leading-none'>
                {sessionLabel}
              </div>
              <span
                className={cn('inline-block flex-shrink-0', {
                  hidden: sessionStatus === 'new',
                })}
                style={{
                  width: '1px',
                  height: '10px',
                  backgroundColor: '#919497',
                }}
              />
            </>
          )}
          <span
            className={cn('text-[#0C99FF] text-sm font-[400] leading-none', {
              hidden: sessionStatus === 'new',
            })}
          >
            {match(sessionConfig.mode)
              .with('model', () => '大模型问答模式')
              .with('graph', () => '图谱洞察模式')
              .with('graph_search', () => '图搜模式')
              .with('knowledge', () => '文档模式')
              .with('table', () => '表格模式')
              .with('database', () => '数据库模式')
              .with('h3c-test', () => '智能体模式')
              .with('external_data', () => '外接数据模式')
              .otherwise(() => null)}
          </span>
        </div>
        {shouldShowDashboard && (
          <div
            className={cn(
              'absolute cursor-pointer flex items-center gap-[8px]',
              'right-[15px] font-semibold top-[50%] translate-y-[-50%]',
              ' text-[#919497] hover:text-[#CC5DE8CC]',
              {
                'text-[#CC5DE8CC]':
                  drawerVisible && drawerType === EDrawerType.CHART,
              },
              {
                hidden: sessionStatus === 'new' && !project_id,
              },
            )}
            onClick={() => handleOpenDrawer(EDrawerType.CHART)}
          >
            <CanvasIcon />
            仪表盘
          </div>
        )}
        {(sessionConfig.mode === 'graph' ||
          sessionConfig.mode === 'graph_search') && (
          <div
            className={cn(
              'absolute cursor-pointer flex items-center gap-[8px]',
              'right-[15px] font-semibold top-[50%] translate-y-[-50%]',
              ' text-[#919497] hover:text-[#CC5DE8CC]',
              {
                'text-[#CC5DE8CC]':
                  drawerVisible && drawerType === EDrawerType.GRAPH,
              },
              {
                hidden: sessionStatus === 'new' && !project_id,
              },
            )}
            onClick={() => handleOpenDrawer(EDrawerType.GRAPH)}
          >
            <CanvasIcon />
            图谱洞察
          </div>
        )}
      </div>
    )
  }

  // 历史记录侧边栏
  const renderSessionHistoryWrap = () => {
    return (
      <div
        className={cn(
          `relative h-full bg-[#fff] overflow-hidden white-space-nowrap`,
          styles.section,
        )}
        style={{ width: sessionVisible ? sessionWidth : 0 }}
      >
        <SessionChatHeader onClose={handleCloseSession} />
        <SessionHistory />
        <div
          className={cn(styles.separatorWrapper)}
          onMouseDown={handleSessionWidthChange}
        >
          <div
            className={cn(
              'absolute col-resize cursor-[col-resize] bg-[#eff1f4] right-[0px] top-[0] w-[0.5px] hover:w-[0.5px] hover:bg-[rgba(12, 153, 255, 0.8)] h-full',
              styles.separator,
              {
                [styles.resizing]: sessionResizing,
              },
            )}
            style={{ right: '6px' }}
          />
        </div>
      </div>
    )
  }

  // 问答详情
  const renderProjectQa = () => {
    // 如果是created状态（有具体会话），始终显示header
    // 如果是new/none状态且侧边栏收起，也要显示header
    const shouldShowHeader = sessionStatus === 'created' || !sessionVisible

    return (
      <>
        <div
          className={cn(
            'flex flex-col flex-1 h-full bg-[#fff] overflow-hidden',
            // {
            //   '!bg-[#F7F7F780]': isOtherPage,
            // },
          )}
        >
          {shouldShowHeader && renderHeader()}
          <GraphUpdateMessage
            sessionStatus={sessionStatus}
            sessionConfig={sessionConfig}
          />
          <div className='flex-1 flex overflow-hidden'>
            {sessionStatus === 'new' || sessionStatus === 'none'
              ? withProvider(<EmptyProject />)
              : withProvider(<Dialog />)}
          </div>
        </div>
      </>
    )
  }

  // 右边抽屉头部
  const renderDrawerHeader = () => {
    return (
      <div className='flex h-[50px] bg-[#fff] '>
        <div className='flex h-[100%] flex-1 items-center justify-between border-b border-b-[#EFF1F4] px-[20px]'>
          <div className='h-full text-sm border-b-[2px] border-b-[#0C1F17] leading-[50px] text-[#0C1F17] font-[500]'>
            {drawerName}
          </div>
          <CloseIcon className='cursor-pointer' onClick={handleCloseDrawer} />
        </div>
      </div>
    )
  }

  // 右边抽屉
  const renderDrawer = () => {
    return (
      <div
        className={cn(
          ` flex flex-col relative h-full bg-[#fff]`,
          styles.section,
        )}
        style={{ width: drawerVisible ? drawerWidth : 0 }}
      >
        {renderDrawerHeader()}
        <div className='flex-1 overflow-auto'>
          {EDrawerType.CHART === drawerType &&
            ['table', 'database'].includes(sessionConfig.mode!) && <Charts />}
          {EDrawerType.REFERENCE === drawerType && <References />}
          {EDrawerType.GRAPH === drawerType && <Graph />}
        </div>
        <div
          className={cn(styles.separatorWrapper, styles.separatorWrapperLeft)}
          onMouseDown={handleDrawerWidthChange}
        >
          <div
            className={cn(
              'absolute bg-[#eff1f4] cursor-[col-resize] left-[0px] top-[0] w-[0.5px] hover:w-1 hover:bg-[rgba(12, 153, 255, 0.8)] h-full',
              styles.separator,
              {
                [styles.resizing]: drawerResizing,
              },
            )}
            style={{ left: '6px' }}
          />
        </div>
      </div>
    )
  }

  // 整个布局
  const renderProjectContent = () => {
    return (
      <div className='flex w-full h-full flex-col overflow-hidden'>
        <div className='flex w-full h-full overflow-hidden'>
          <div className='flex w-full h-full overflow-hidden'>
            {/* 未分组会话和其他页面模式下不显示历史记录侧边栏 */}
            {!isOtherPage && !isUngroupedSession && renderSessionHistoryWrap()}
            {renderProjectQa()}
          </div>
          {renderDrawer()}
        </div>
      </div>
    )
  }

  return withProvider(renderProjectContent())
})

/**
 * 组件的重置时机:\
 * 1.老会话切换.session_id存在且变为另一个值\
 * 2.历史记录->新/老会话.sessionStatus从none变其他\
 * 3.老新会话切换.sessionStatus从created变new 或者反过来\
 * 注:session_id不由useState创建 与sessionStatus的更新时机可能不一致
 */
const useReset = (config: {
  session_id: any
  sessionStatus: SessionInfo['sessionStatus']
  reset: () => void
  setSessionStatus: (val: SessionInfo['sessionStatus']) => void
  dialog: DialogList
  type: string | undefined
}) => {
  const { session_id, sessionStatus, reset, setSessionStatus, dialog, type } =
    config
  const prevSessionIdRef = useRef(session_id)
  const prevSessionStatusRef = useRef(sessionStatus)
  const prevTypeRef = useRef(type)
  useEffect(() => {
    const prevSessionId = prevSessionIdRef.current
    const prevSessionStatus = prevSessionStatusRef.current
    const prevType = prevTypeRef.current
    prevSessionIdRef.current = session_id
    prevSessionStatusRef.current = sessionStatus
    prevTypeRef.current = type
    if (prevSessionId === session_id && prevSessionStatus === sessionStatus) {
      return
    }
    // 类型变化需要重置
    if (prevType !== type) {
      reset()
      return
    }
    // 从 undefined 变为有值：自动进入会话时，设置状态为 'created'
    if (!prevSessionId && session_id) {
      setSessionStatus('created')
      // 只有当 dialog 为空时才重置（说明是点击历史记录进入，而不是新建会话提问进入）
      // 如果 dialog 不为空，说明是新建会话正在提问中，不需要重置，避免清空正在流式输出的内容
      if (dialog.length === 0) {
        reset()
      }
      return
    }
    if (prevSessionId && prevSessionId !== session_id) {
      // 路由从具体会话跳回项目根路径时，需要切换为"历史列表"视图
      // 仅当当前仍处于已创建会话状态且当前没有session_id时设置为 none，
      // 避免覆盖"新建会话(+)"场景中已切到 new 的状态。
      if (!session_id) {
        // 如果 session_id 被清空，且当前状态是 'created'，说明是从会话跳回，设置为 'none' 显示历史列表
        // 如果当前状态已经是 'new'，说明是新建会话，保持 'new' 状态
        if (sessionStatus === 'created') {
          setSessionStatus('none')
        } else if (sessionStatus === 'new') {
          // 保持 'new' 状态，确保知识库选择器显示
        }
      }
      reset()
      return
    }
    if (prevSessionStatus === 'none' && prevSessionStatus !== sessionStatus) {
      reset()
      return
    }
    if (prevSessionStatus === 'created' && sessionStatus === 'new') {
      reset()
      return
    }
    // 如果从 'new' 变为 'created' 且有对话内容，说明是首次问答完成，不重置
    // 这是关键逻辑：当用户首次提问后，会话状态从 'new' 变为 'created'，
    // 此时如果已经有对话内容，说明是正常的问答流程，不应该重置对话记录
    if (prevSessionStatus === 'new' && sessionStatus === 'created') {
      // 只要有对话内容（即使只有一条），就不重置，保持对话记录不刷新
      // 对于 global 类型，即使没有对话内容，也不应该重置（因为可能是首次问答刚完成）
      if (dialog.length > 0 || type === 'global') {
        return
      }
      reset()
      return
    }
  })
}
