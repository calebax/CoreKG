import { create } from 'zustand'

// 定义数据类型
interface UserListItem {
  value: number
  label: string
}

interface DataStore {
  // 知识库列表
  knowledgeBaseList: any[]
  setKnowledgeBaseList: (list: any[]) => void

  // 用户列表
  userList: UserListItem[]
  setUserList: (list: UserListItem[]) => void

  // 模型列表
  modelList: any[]
  setModelList: (list: any[]) => void
}

const useDataStore = create<DataStore>((set) => ({
  // 知识库列表
  knowledgeBaseList: [],
  setKnowledgeBaseList: (list) => set({ knowledgeBaseList: list }),

  // 用户列表
  userList: [],
  setUserList: (list) => set({ userList: list }),

  // 模型列表
  modelList: [],
  setModelList: (list) => set({ modelList: list }),
}))

export default useDataStore
