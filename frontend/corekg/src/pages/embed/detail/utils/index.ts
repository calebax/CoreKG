import { useParams } from 'react-router-dom'

export const useAgentInfo = () => {
  const { id } = useParams<{ id: string }>()

  return {
    agentId: id ? Number(id) : null,
    // 在iframe模式下，agentDetail通过props传递，不需要在这里获取git
    agentDetail: null,
  }
}
