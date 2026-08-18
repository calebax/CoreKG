import { createContext, FC, PropsWithChildren } from 'react'
import { BasicAgentInfo } from 'Agent'
import { useRequest } from 'ahooks'
import { listAgent } from '@/api'

// 通过context共享useRequest提供的数据
type _AgentList = ReturnType<typeof useRequest<BasicAgentInfo[], []>>
const AgentContext = createContext<_AgentList | null>(null)

export const AgentProvider: FC<PropsWithChildren> = (props) => {
  const agentList = useRequest(async () => {
    const res: any = await listAgent({})
    const _agentList: BasicAgentInfo[] = []
    ;(res.Data as any[])
      .filter((item) => item.ID !== 269)
      .forEach((item: any) => {
        _agentList.push({
          ...item,
          id: item.ID,
          avatar: item.avatar_url,
          title: item.show_name,
          description: item.description,
          isAdmin: item.is_admin,
          favorite: item.is_collected,
          type: item.agent_type,
          source: item.created_type === 'system' ? 'system' : 'custom',
          tag: item.tag?.[0],
          status: item.publish_status === 'published' ? 'published' : 'draft',
          is_synced: item.is_synced,
        })
      })
    return _agentList
  })
  return (
    <AgentContext.Provider value={agentList}>
      {props.children}
    </AgentContext.Provider>
  )
}

/** 获取当前的应用列表 */
// eslint-disable-next-line react-refresh/only-export-components
export const useAgentList = () => {
  const agentList = useContext(AgentContext)
  if (!agentList) throw new Error()
  return agentList
}
