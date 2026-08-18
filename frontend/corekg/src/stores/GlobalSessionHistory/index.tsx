import { create } from 'zustand'
import { getSessionHistory } from '@/api/agent'

type HistoryItem = {
  id: number
  name: string
  nameLoading?: boolean
}
type HistoryStore = {
  history: HistoryItem[]
  /** 加载数据 */
  loadData: () => Promise<void>
  /** 设置数据 */
  set: (history: HistoryItem[]) => void
  /** 前方添加 */
  add: (...items: HistoryItem[]) => void
  del: (id: number) => void
  rename: (id: number, newName: string) => void
}
export const useGlobalSessionHistory = create<HistoryStore>((set) => ({
  history: [],
  loadData: async () => {
    const { Data } = (await getSessionHistory({
      limit: 9999,
      offset: 0,
      orderBy: ['created_at desc'],
      filters: [
        {
          field: 'resource_type',
          value: [
            'forest',
            'file_list',
            'dir_list',
            'excel_list',
            'react_excel_list',
          ],
        },
      ],
    })) as any
    if (!Data) return
    const history = Data.map((item: any) => {
      const { ID: id, name } = item
      return { id, name }
    })
    set(() => {
      return { history }
    })
  },
  set: (history) => {
    set(() => {
      return { history }
    })
  },
  add: (...items) => {
    set((state) => {
      return {
        history: [
          ...items,
          ...state.history.filter(
            (item) => !items.some((newItem) => newItem.id === item.id),
          ),
        ],
      }
    })
  },
  del: (id) => {
    set((state) => {
      return {
        history: state.history.filter((item) => item.id !== id),
      }
    })
  },
  rename: (id, newName) => {
    set((state) => {
      return {
        history: state.history.map((item) => {
          if (item.id !== id) return item
          return { id, name: newName }
        }),
      }
    })
  },
}))
