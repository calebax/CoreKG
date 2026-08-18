import { Agent } from 'Agent'
import { useRequest } from 'ahooks'
import { getAgentDetail } from '@/api'

/** 获取应用详细信息 */
export const useAgent = () => {
  const navigate = useNavigate()
  const params = useParams()
  const id = useMemo(() => {
    const _id = Number(params.id)
    if (Number.isInteger(_id)) return _id
    return null
  }, [params.id])
  if (!id) {
    navigate('/agents', { replace: true })
  }
  const agent = useRequest(
    async (): Promise<Agent> => {
      const res: any = await getAgentDetail(id)
      return {
        ...res,
        id: res.ID,
        avatar: res.avatar_url,
        title: res.show_name,
        isAdmin: res.is_admin,
        type: res.agent_type,
        source: 'custom',
        status: res.publish_status,
        forests: (res.forests as any[])?.map((item) => ({
          ...item,
          id: item.ID,
        })),
      }
    },
    { refreshDeps: [id], ready: Boolean(id) },
  )
  // 不要把id作为返回值 始终通过agent.data.id获取
  // id和agent.data可能不一致
  return agent
}
