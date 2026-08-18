import { useRequest } from 'ahooks'
import { getAgentInfo } from '@/api'

/** 获取agent基本信息 */
export const useAgentInfo = () => {
  const params = useParams()
  const agentId = useMemo(() => {
    const _id = parseInt(params.id!)
    if (Number.isInteger(_id)) return _id
    return undefined
  }, [params.id])

  const { data: agentDetail } = useRequest(
    async () => {
      const res = await getAgentInfo(agentId!)
      return res
    },
    {
      refreshDeps: [agentId],
      ready: Boolean(agentId),
    },
  )

  return {
    agentId,
    agentDetail,
  }
}
