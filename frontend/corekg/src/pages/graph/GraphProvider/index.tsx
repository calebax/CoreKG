import {
  createContext,
  FC,
  PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
} from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { App } from 'antd'
import { GraphBaseInfo, GraphTag } from 'Graph'
import { useMemoizedFn, useRequest } from 'ahooks'
import { getGraphEdges, getGraphInfo, listGraphTag } from '@/api/graph'

export type GraphTagWithId = GraphTag & { tag_id: number }

type BaseInfo = Pick<
  GraphBaseInfo,
  | 'id'
  | 'description'
  | 'name'
  | 'status'
  | 'parse_mode'
  | 'forest_id'
  | 'is_admin'
>
type ContextValue = {
  data?: BaseInfo & {
    tags: GraphTagWithId[]
    isEditingRules?: boolean
    edgeOptions: string[]
  }
  loading: boolean
  /** 在已经获取到基础信息后可用 进行增量更新 */
  updateBaseInfo: (val: Partial<BaseInfo>) => void
  reloadTag: () => void
  mutateTags: (tags: GraphTagWithId[]) => void
}
const GraphInfoContext = createContext<ContextValue | null>(null)
/**
 * 提供图谱基础信息的上下文.\
 * 图谱可能尚未创建 id可能不存在
 */
// eslint-disable-next-line react-refresh/only-export-components
const GraphInfoProvider: FC<PropsWithChildren & { strictly: boolean }> = (
  props,
) => {
  const { strictly } = props
  const [searchParams] = useSearchParams()
  /** 是否正在编辑规则 */
  const isEditingRules = searchParams.get('rule') === 'true'
  const graphId = useGraphId()
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const {
    data: baseInfo,
    loading: loading1,
    error: error1,
    mutate: mutateBaseInfo,
  } = useRequest(
    async () => {
      if (graphId) {
        const res = await getGraphInfo({
          graph_id: graphId,
        })
        const {
          ID,
          description,
          name,
          status,
          parse_mode,
          forest_id,
          is_admin,
        } = res
        const info: BaseInfo = {
          is_admin,
          id: ID,
          description,
          name,
          status,
          parse_mode,
          forest_id,
        }
        // 仅允许编辑规则
        if (
          !isEditingRules &&
          ['edit'].some((s) => pathname.includes(s)) &&
          info.status !== 'draft'
        ) {
          message.warning('本图谱已经开始解析，不能编辑')
          throw new Error('本图谱已经开始解析，不能编辑')
        }
        return info
      } else if (!strictly) {
        const defaultInfo: BaseInfo = {
          id: 0,
          is_admin: true,
          description: '',
          name: '',
          status: 'draft',
          parse_mode: 'auto',
          forest_id: 0,
        }
        return defaultInfo
      } else {
        throw new Error('graphId错误')
      }
    },
    { refreshDeps: [graphId, pathname] },
  )
  const updateBaseInfo = useMemoizedFn((val: Partial<BaseInfo>) => {
    if (!baseInfo) return
    mutateBaseInfo({ ...baseInfo, ...val })
  })
  const {
    data: tags,
    loading: loading2,
    error: error2,
    run: reloadTag,
    mutate: mutateTags,
  } = useRequest(
    async () => {
      if (graphId) {
        const res = await listGraphTag({
          graph_id: graphId,
          filters: [{ field: 'tag_type', value: ['TAG'] }],
        })
        const data: any[] = res.Data ?? []
        const _tags = data.map((item): GraphTagWithId => {
          const { tag_name, description, properties, ID } = item
          return {
            tag_id: ID,
            tag_name,
            description,
            properties,
          }
        })
        return _tags
      } else if (!strictly) {
        return []
      } else {
        throw new Error('graphId错误')
      }
    },
    { refreshDeps: [graphId] },
  )

  const {
    data: edgeOptions,
    loading: loading3,
    error: error3,
  } = useRequest(
    async () => {
      if (graphId) {
        const res = await getGraphEdges({
          graph_id: graphId,
        })
        const edges = res.edges ?? []
        // 提取所有唯一的 edge_name 作为备选项
        const uniqueEdgeNames = Array.from(
          new Set(edges.map((edge) => edge.edge_name)),
        )
        return uniqueEdgeNames
      } else if (!strictly) {
        return []
      } else {
        return []
      }
    },
    { refreshDeps: [graphId] },
  )
  useEffect(() => {
    if (error1 || error2 || error3) {
      navigate('/graph')
    }
  }, [error1, error2, error3, navigate])
  const data = useMemo(() => {
    if (baseInfo && tags && edgeOptions !== undefined) {
      return {
        ...baseInfo,
        tags,
        isEditingRules,
        edgeOptions: edgeOptions ?? [],
      }
    }
    return undefined
  }, [baseInfo, isEditingRules, tags, edgeOptions])
  const loading = loading1 || loading2 || loading3
  const contextValue = useMemo((): ContextValue => {
    return {
      data,
      loading,
      updateBaseInfo,
      reloadTag,
      mutateTags,
    }
  }, [data, loading, mutateTags, reloadTag, updateBaseInfo])
  return (
    <GraphInfoContext.Provider value={contextValue}>
      {props.children}
    </GraphInfoContext.Provider>
  )
}

/**
 *
 * @param _Comp
 * @param strictly 强制要求graph有数据
 */
export function withGraphProvider<T>(
  _Comp: FC<T>,
  strictly: boolean = true,
): FC<T> {
  const Comp = _Comp as any
  return (props) => {
    return (
      <GraphInfoProvider strictly={strictly}>
        <Comp {...props} />
      </GraphInfoProvider>
    )
  }
}

export const useGraphInfo = () => {
  const value = useContext(GraphInfoContext)
  if (!value) throw new Error('graph 不处于正确的context内')
  return value
}

/** 从查询参数中获取graphId */
export const useGraphId = () => {
  const [searchParams] = useSearchParams()
  const graphId = useMemo(() => {
    const _id = parseInt(searchParams.get('graphId')!)
    if (Number.isNaN(_id)) return null
    return _id
  }, [searchParams])
  return graphId
}
