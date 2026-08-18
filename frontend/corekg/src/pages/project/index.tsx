import { createContext, FC, useRef, useMemo } from 'react'
import { Skeleton } from 'antd'
import { useMemoizedFn, useRequest } from 'ahooks'
import { produce } from 'immer'
import type { Layout } from 'react-grid-layout'
import { getSessionHistory, listCustomModel } from '@/api'
import { cn } from '@/utils'
import { loadAccountBindingList } from '@/api/accountBindings'
import {
  saveChartCanvas,
  deleteProjectCharts,
  getProjectCharts,
  getProjectInfo,
} from '@/api/project'
import { ProjectContent } from './ProjectContent'

export type Knowledge = {
  forest_id: number
  name: string
  node_type:
    | 'file'
    | 'excel_sheet'
    | 'mysql_table'
    | 'qa'
    | 'other'
    | (string & {})
  key: string
  knowledgeType: 'file' | 'excel_sheet' | 'mysql_table' | 'qa' | 'other'
  parentKey?: string
  children?: Knowledge[]
  [key: string]: any
}

export type ExternalDataSourceInfo = {
  bindings: Array<{
    account: string
    boundAt: string
    id: number
    provider: string
    valid: boolean
  }>
  supported: Array<{
    logo: string
    provider: string
  }>
}

export type ExternalDataSourceItem = {
  id?: number
  provider: string
  logo: string
  label: string
}

export type ProjectInfo = {
  project_id: number | undefined
  session_id?: number
  /**
   * 如果需要session_id与sessionStatus同时更改 调用此函数
   */
  setSessionId: (val?: number, stopNavigate?: boolean) => void
  loading: boolean
  models: {
    id: number
    name: string
    desc: string
    avatar: string
    is_last_used: boolean
    last_used_at?: string
    model_group: string
  }[]
  // 所有外部数据源
  externalDataSourceList: ExternalDataSourceItem[]
  data: {
    project_name: string
    sessions: {
      session_id: number
      name: string
      nameLoading?: boolean
    }[]
    charts: {
      id: number
      /** echarts配置 */
      option: any
      /** react-grid的位置 */
      layout: Layout
      /** 传来的额外数据 */
      extra: any
    }[]
    /** 仅展示用 */
    knowledge: Knowledge[]
    /**
     * 选中的外部数据源,
     * 存放的是AccountAppsItem的id
     * */
    checkedExternalDataSourceIds: number[]
  }
  /**
   * 使用一个值或者更新函数 更新本地的对象\
   * 值和更新函数都会进行增量更新 即只更新对象中存在的属性\
   * 更新函数会通过immer进行 直接设置属性即可
   */
  setData: (
    arg: Partial<ProjectInfo['data']> | ((val: ProjectInfo['data']) => void),
  ) => void
  reloadKnowledge: () => void
  reloadSessions: () => void
  chartsOperators: {
    add: (
      id: number,
      option: any,
      extra?: any,
      sessionIdOverride?: number,
    ) => void
    del: (id: number) => void
    setLayouts: (layouts: Layout[]) => void
  }
  isOtherPage?: boolean
  defaultKnowBase?: number
  /** 是否为未分组会话（project_id === 0） */
  isUngroupedSession?: boolean
  type?: 'single-file' | 'global'
}
export interface ProjectProps {
  project_id?: number
  session_id?: number
  defaultKnowBase?: number
  type?: 'single-file' | 'global'
}

export const ProjectContext = createContext<ProjectInfo | null>(null)
const Project: FC<ProjectProps & { useCustomIds?: typeof useIds }> = (
  props,
) => {
  const { useCustomIds = useIds, type } = props
  const { project_id, session_id, setSessionId } = useCustomIds(props)
  const [isOtherPage, setIsOtherPage] = useState<boolean>(
    type === 'single-file',
  )

  // 标记是否为未分组会话模式（project_id === 0）
  const isUngroupedSession = project_id === 0
  const getBaseInfo = async () => {
    // 如果是未分组会话，不需要获取项目信息
    if (isUngroupedSession) {
      return {
        project_name: '未分组会话',
        knowledge: [],
      }
    }

    const { name: project_name, forest } = await getProjectInfo({
      id: project_id!,
    })
    const projectForest: any[] = forest ?? []
    const knowledge: ProjectInfo['data']['knowledge'] = projectForest.map(
      (item) => {
        const { forest_type, name, id } = item
        return {
          forest_id: id,
          key: '' + id,
          parentKey: '',
          name,
          knowledgeType: 'other',
          node_type: 'forest',
          type: forest_type,
        }
      },
    )
    return {
      project_name,
      knowledge,
    }
  }
  const getSessions = async () => {
    // 未分组会话不需要获取历史会话列表
    if (isUngroupedSession) {
      return []
    }

    const res = await getSessionHistory({
      project_id,
    })
    const data: any[] = res.Data ?? []
    const sessions: ProjectInfo['data']['sessions'] = data.map((item: any) => {
      const { ID: id, name } = item
      return {
        session_id: id,
        name,
      }
    })
    return sessions
  }
  const getCharts = useMemoizedFn(async () => {
    if (!session_id) return []
    try {
      const { content } = await getProjectCharts(session_id)
      const charts: any[] = JSON.parse(content)
      charts.forEach((item, i) => {
        if (item.layout) return
        item.layout = { i: '' + i, x: 0, y: 0, w: 1, h: 1 }
      })
      return charts
    } catch {
      return []
    }
  })

  const createExternalDataSourceList = ({
    bindings,
    supported,
  }: ExternalDataSourceInfo) => {
    const bindingInfos: Record<
      string,
      ExternalDataSourceInfo['bindings'][number]
    > = {}
    bindings?.forEach((item: ExternalDataSourceInfo['bindings'][number]) => {
      bindingInfos[item.provider] = item
    })

    return supported.map((item) => {
      const bindInfo = bindingInfos[item.provider]
      const label =
        item.provider.charAt(0).toUpperCase() + item.provider.slice(1)
      if (bindInfo) {
        return {
          ...item,
          id: bindInfo.id,
          label,
        }
      }
      return {
        ...item,
        label,
      }
    })
  }

  // 用于跟踪是否是首次加载（避免在用户清除 session_id 后自动恢复）
  const isInitialLoadRef = useRef(true)
  // 用于跟踪上一次的 session_id，避免重复加载
  const prevSessionIdRef = useRef<number | undefined>(undefined)

  const { data, loading, mutate, error } = useRequest(
    async () => {
      if (project_id === undefined) return
      const [{ project_name, knowledge }, sessions] = await Promise.all([
        getBaseInfo(),
        getSessions(),
      ])
      // 只有在首次加载且没有 session_id 时，才自动设置第一个会话
      // 这样可以避免在用户点击"新建会话"清除 session_id 后，自动恢复
      // 未分组会话不自动设置第一个会话
      if (
        isInitialLoadRef.current &&
        sessions[0] &&
        !session_id &&
        !isOtherPage &&
        !isUngroupedSession
      ) {
        setSessionId(sessions[0].session_id)
      }
      isInitialLoadRef.current = false
      const charts = await getCharts()
      const data: ProjectInfo['data'] = {
        project_name,
        sessions,
        charts,
        knowledge,
        checkedExternalDataSourceIds: [],
      }

      return data
    },
    {
      refreshDeps: [project_id],
      ready: project_id !== undefined,
    },
  )

  const navigate = useNavigate()
  useEffect(() => {
    if (project_id === undefined || project_id === null || error) {
      navigate('/')
    }
  }, [error, navigate, project_id])

  const setData: ProjectInfo['setData'] = useMemoizedFn((arg) => {
    mutate((val) => {
      if (!val) return val
      const newVal = typeof arg === 'function' ? produce(val, arg) : arg
      return {
        ...val,
        ...newVal,
      }
    })
  })

  // 监听 session_id 变化，重新加载图表数据
  useEffect(() => {
    const isSessionChanged = prevSessionIdRef.current !== session_id
    prevSessionIdRef.current = session_id

    if (!session_id) {
      setData({ charts: [] })
      return
    }

    // 只有当 session_id 真正变化时才重新加载图表
    if (!isSessionChanged) {
      return
    }

    // 延迟加载图表，避免与新添加的图表冲突
    const timer = setTimeout(async () => {
      try {
        const charts = await getCharts()
        setData({ charts })
      } catch {
        setData({ charts: [] })
      }
    }, 300) // 延迟300ms，等待后端保存完成

    return () => clearTimeout(timer)
  }, [getCharts, session_id, setData])

  const chartsOperators: ProjectInfo['chartsOperators'] = (() => {
    const add: ProjectInfo['chartsOperators']['add'] = (
      id,
      option,
      extra,
      sessionIdOverride,
    ) => {
      const activeSessionId = sessionIdOverride ?? session_id
      // 如果没有会话 ID，无法保存图表
      if (!activeSessionId) return

      setData((draft) => {
        const { charts } = draft
        // 如果图表已存在，跳过添加（避免重复）
        if (charts.some((item) => item.id === id)) return

        // 添加新图表到数组开头，并设置默认布局
        charts.unshift({
          id,
          option,
          layout: {
            /** 和图表id一样 也是唯一的 */
            i: `${id}`,
            x: 0,
            y: 0,
            w: 2,
            h: 4,
          },
          extra,
        })
        // 保存图表配置到后端
        saveChartCanvas(activeSessionId, JSON.stringify(draft.charts))
      })
    }
    const del: ProjectInfo['chartsOperators']['del'] = (id) => {
      if (!session_id) return
      setData((draft) => {
        draft.charts = draft.charts.filter((item) => item.id !== id)
        deleteProjectCharts([id])
        saveChartCanvas(session_id, JSON.stringify(draft.charts))
      })
    }
    const setLayouts: ProjectInfo['chartsOperators']['setLayouts'] = async (
      layouts,
    ) => {
      if (!session_id) return
      setData((draft) => {
        const { charts } = draft
        charts.forEach((chart) => {
          chart.layout = layouts.find((item) => item.i === chart.layout.i)!
        })
        saveChartCanvas(session_id, JSON.stringify(charts))
      })
    }
    return {
      add,
      del,
      setLayouts,
    }
  })()
  const reloadKnowledge = useMemoizedFn(async () => {
    const { knowledge } = await getBaseInfo()
    setData({ knowledge })
  })
  const reloadSessions = useMemoizedFn(async () => {
    const sessions = await getSessions()
    setData({ sessions })
  })

  const { data: externalDataSourceList, loading: loadingExternal } = useRequest(
    async () => {
      const externalDataSourceInfo = await loadAccountBindingList()
      return createExternalDataSourceList(externalDataSourceInfo)
    },
  )

  const { data: models = [], loading: loadingModels } = useRequest(async () => {
    const { Data: chat_models = [] } = (await listCustomModel()) as any as {
      Data?: any[]
    }

    return chat_models.map((item) => {
      const { ID, show_name, description, head_url } = item
      return {
        ...item,
        id: ID,
        name: show_name,
        desc: description,
        avatar: head_url,
      }
    })
  })

  // 使用 useMemo 优化 contextValue，避免因 loadingModels 和 loadingExternal 变化导致不必要的重新渲染
  // 只保留关键数据作为依赖项，loading 状态的变化不应该导致 context 重新创建
  const contextValue = useMemo(() => {
    if (
      project_id === undefined ||
      project_id === null ||
      !data ||
      !externalDataSourceList
    )
      return null
    const value: ProjectInfo = {
      project_id,
      session_id,
      setSessionId,
      loading: loading || loadingModels || loadingExternal,
      models,
      data,
      setData,
      chartsOperators,
      reloadKnowledge,
      reloadSessions,
      externalDataSourceList,
      isOtherPage,
      defaultKnowBase: props?.defaultKnowBase,
      isUngroupedSession,
      type,
    }
    return value
    // 移除 loadingModels 和 loadingExternal 作为依赖项，只保留关键数据
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    project_id,
    session_id,
    setSessionId,
    models,
    data,
    setData,
    chartsOperators,
    reloadKnowledge,
    reloadSessions,
    externalDataSourceList,
    isOtherPage,
    props?.defaultKnowBase,
    isUngroupedSession,
    // loading 状态仍然需要，但通过 loading 变量传递，而不是 loadingModels 和 loadingExternal
    loading,
  ])

  return (
    <div
      className={cn(
        'bg-[#ffffff] w-full h-full overflow-hidden',
        //   {
        //   '!bg-[#F7F7F780]': isOtherPage,
        // }
      )}
    >
      {!contextValue || contextValue.loading ? (
        <Skeleton active className='m-4' />
      ) : (
        <ProjectContext.Provider value={contextValue}>
          <ProjectContent />
        </ProjectContext.Provider>
      )}
    </div>
  )
}
export default Project

/** 获取已经加载完毕的项目数据 */
// eslint-disable-next-line react-refresh/only-export-components
export const useProject = () => {
  const value = useContext(ProjectContext)
  if (!value) throw new Error('必须被ProjectContext包裹')
  return value
}

const useIds = (props: ProjectProps) => {
  const { id, '*': rest } = useParams()

  const param2 = rest?.split('/')?.[0]

  const project_id = useMemo(() => {
    if (props.project_id !== undefined) return props.project_id
    const _project_id = parseInt(id!)
    // 允许 project_id 为 0（未分组会话）
    return Number.isInteger(_project_id) && _project_id >= 0
      ? _project_id
      : undefined
  }, [id, props.project_id])
  const paramSessionId = useMemo(() => {
    if (project_id === undefined || project_id === null) return undefined
    if (props.session_id !== undefined || props.project_id !== undefined)
      return props.session_id || undefined
    const _session_id = parseInt(param2!)
    return Number.isInteger(_session_id) ? _session_id : undefined
  }, [param2, project_id, props.project_id, props.session_id])
  const [session_id, _setSessionId] = useState(paramSessionId)

  // 用于跟踪用户是否主动清除了 session_id
  const userClearedSessionRef = useRef(false)

  const navigate = useNavigate()
  const setSessionId = useMemoizedFn((val?: number, stopNavigate?: boolean) => {
    // 如果用户主动清除 session_id（传入 undefined），设置标志
    if (val === undefined && session_id) {
      userClearedSessionRef.current = true
    }
    if (!props.project_id && !stopNavigate) {
      navigate(`/project/${project_id}/${val ?? ''}`)
    }
    _setSessionId(val)
  })

  useEffect(() => {
    // 如果用户主动清除了 session_id，不自动从 URL 恢复
    if (userClearedSessionRef.current && !paramSessionId) {
      // URL 已经是空的，用户清除成功，重置标志
      userClearedSessionRef.current = false
      return
    }
    // 如果 URL 参数变化且不是用户主动清除，同步状态
    if (userClearedSessionRef.current && paramSessionId) {
      // URL 又有了新值，说明是新的会话创建，重置标志
      userClearedSessionRef.current = false
    }
    _setSessionId(paramSessionId)
  }, [paramSessionId])

  return {
    project_id,
    session_id,
    setSessionId,
  }
}
