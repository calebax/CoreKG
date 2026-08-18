import { createContext, FC, PropsWithChildren } from 'react'
import { Skeleton } from 'antd'
import { Agent } from 'Agent'
import { useRequest } from 'ahooks'
import { listCustomModel, listEmployee } from '@/api'
import { uniqueArray } from '@/utils'
import { getKnowledgeBaseList } from '@/api/knowledge'
import { useAdmin } from '@/utils/useAdmin'
import useLocalStore from '@/stores/local'
import { AgentEditValue } from '../..'
import { useAgent } from '../../utils/useAgent'

type ContextValue = {
  agent: AgentEditValue
  managers: {
    uin: string
    name: string
  }[]
  forestList: {
    id: number
    name: string
    forest_type: Agent['forests'][number]['forest_type']
  }[]
  models: {
    id: number
    name: string
    description: string
  }[]
}
const EditContext = createContext<ContextValue | null>(null)
export const EditProvider: FC<PropsWithChildren> = (props) => {
  const navigate = useNavigate()
  const { adminIds } = useAdmin()
  const { uinId } = useLocalStore((state) => state.userInfo)
  const { data: agentData, error } = useAgent()
  useEffect(() => {
    if (error) {
      navigate('/agents')
    }
  }, [error, navigate])
  const agent: AgentEditValue | null = useMemo(() => {
    if (!agentData) return null
    const _agent = { ...agentData }
    delete (_agent as any).manager_list
    const _public_scope = _agent.public_scope
    // 默认仅自己为管理员
    const manager_ids: number[] =
      agentData.type === 'workflow'
        ? _agent.manager_ids ?? [uinId as any]
        : uniqueArray([uinId as any], _agent.manager_ids ?? [])
    const _scope_ids: number[] = _agent.scope_ids ?? []
    return {
      ..._agent,
      // 编辑表单时 将管理员id加入scoped_ids
      manager_ids,
      scope_ids: uniqueArray(manager_ids, _scope_ids),
      public_scope: _public_scope === 'custom' ? 'custom' : 'custom',
      params: _agent.params?.map((item, i) => ({ ...item, key: `${i}` })),
    }
  }, [agentData, adminIds, uinId])
  const { data: managers } = useRequest(async () => {
    const res: any = await listEmployee({ listAll: true })
    const data: any[] = res.Data ?? []
    return data.map((item) => {
      return {
        uin: item.uin as string,
        name: item.user_name as string,
      }
    })
  })
  const { data: forestList } = useRequest(async () => {
    const res: any = await getKnowledgeBaseList({
      offset: 0,
      limit: 9999,
      filters: [
        {
          field: 'forest_type',
          value: ['file', 'qa', 'cad'],
        },
      ],
    })
    const data: any[] = res.Data ?? []
    const _forestList: NonNullable<Agent>['forests'] = data.map((item) => {
      return {
        id: item.ID,
        forest_type: item.forest_type,
        name: item.name,
      }
    })
    return _forestList
  })
  const { data: models } = useRequest(async () => {
    const res: any = await listCustomModel()
    const data: any[] = res.Data ?? []
    return data.map((item) => {
      return {
        id: item.ID as number,
        name: item.show_name as string,
        description: item.description as string,
      }
    })
  })

  if (!agent || !managers || !forestList || !models) {
    return (
      <Skeleton
        active
        className='mx-auto mt-4 w-[90%]'
        paragraph={{ rows: 20 }}
      />
    )
  }
  return (
    <EditContext.Provider value={{ agent, managers, forestList, models }}>
      {props.children}
    </EditContext.Provider>
  )
}
/** 获取当前上下文的agent数据 */
// eslint-disable-next-line react-refresh/only-export-components
export const useEditContext = () => {
  const contextValue = useContext(EditContext)
  if (!contextValue) throw new Error('agent上下文数据异常')
  const value = useMemo(() => {
    const { agent, managers, forestList, models } = contextValue

    return {
      agent,
      managers,
      forestList,
      models,
    }
  }, [contextValue])
  return value
}
