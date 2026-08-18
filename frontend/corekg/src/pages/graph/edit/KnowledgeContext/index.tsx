import { createContext, FC, PropsWithChildren } from 'react'
import { useRequest } from 'ahooks'
import { getKnowledgeBaseList } from '@/api/knowledge'

export type Forest = {
  id: number
  name: string
  graph_status: 'uncreated' | (string & {})
  CreatedAt?: string
  UpdatedAt?: string
}
export type FileItem = {
  id: number
  name: string
  parent_id: number | null
  is_dir: boolean
}

type ContextValue = {
  data?: Forest[]
  loading: boolean
  loadData: () => void
}

const KnowledgeContext = createContext<ContextValue | null>(null)
export const KnowledgeProvider: FC<PropsWithChildren> = (props) => {
  // 首次进入知识选择界面 主动点击获取知识
  const {
    data,
    loading,
    run: loadData,
  } = useRequest(
    async () => {
      const forests: any[] =
        (
          await getKnowledgeBaseList({
            limit: 9999,
            offset: 0,
            filters: [
              {
                field: 'forest_type',
                value: ['file'],
                exactMatch: true,
              },
            ],
          })
        ).Data ?? []
      const forestList = forests
        .filter((item) => item.is_admin)
        .map((item) => {
          return { id: item.ID, ...item } as Forest
        })

      return forestList
    },
    { manual: true },
  )

  return (
    <KnowledgeContext.Provider
      value={{
        data,
        loading,
        loadData,
      }}
    >
      {props.children}
    </KnowledgeContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useKnowledge = () => {
  const contextValue = useContext(KnowledgeContext)
  if (!contextValue) throw new Error('必须被FileListContext包裹')
  return contextValue
}
