import { createContext, useContext, FC, PropsWithChildren } from 'react'
import { useRequest } from 'ahooks'
import { getSessionHistory } from '@/api'

const HistoryContext = createContext<{
  agentId: number
  value?: any[]
  setValue: (val: any[]) => void
  loading: boolean
  refresh: () => void
} | null>(null)

export const HistoryProvider: FC<PropsWithChildren & { agentId: number }> = (
  props,
) => {
  const { agentId } = props
  const { data, mutate, loading, refresh } = useRequest(
    async () => {
      const res = await getSessionHistory({
        limit: 9999,
        offset: 0,
        filters: [
          { field: 'resource_id', value: [String(agentId)] },
          { field: 'resource_type', value: ['agent'] },
        ],
      })
      const data: any[] = res.Data ?? []
      return data
    },
    {
      refreshDeps: [agentId],
    },
  )

  return (
    <HistoryContext.Provider
      value={{ agentId, value: data, setValue: mutate, loading, refresh }}
    >
      {props.children}
    </HistoryContext.Provider>
  )
}

/** 当前agent的历史记录 */
// eslint-disable-next-line react-refresh/only-export-components
export const useHistory = () => {
  const history = useContext(HistoryContext)
  if (!history) {
    // 安全回退：在未被 Provider 包裹时，返回一个稳定的占位对象，避免页面崩溃
    return {
      agentId: 0,
      value: [],
      setValue: () => {},
      loading: true,
      refresh: () => {},
    }
  }
  return history
}
