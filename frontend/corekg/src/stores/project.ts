import { create } from 'zustand'
import {
  getProjectList,
  createProject,
  deleteProject,
  renameProject,
} from '@/api/project'

interface IProjectItem {
  id: number
  name: string
}

interface IProjectStore {
  projectList: IProjectItem[]
  load: () => Promise<void>
  add: (name?: string) => Promise<number>
  del: (id: number, moveToFree?: boolean) => Promise<void>
  rename: (data: { id: number; name: string }) => Promise<void>
}

const useProjectStore = create<IProjectStore>((set) => ({
  projectList: [],
  load: async () => {
    try {
      const { data } = await getProjectList({
        orderBy: ['updated_at desc'],
      })
      //临时过滤掉project_type的值为forest_qa的那个项目
      const projectList = data
        .filter((item: any) => item.project_type !== 'forest_qa')
        .map(({ id, name }: any) => {
          return {
            id,
            name,
          }
        })
      set({ projectList })
    } catch (error) {
      console.log(error)
    }
  },
  add: async (name?: string) => {
    const response: any = await createProject(name ? { name } : {})
    const id = response.ID
    const projectName = response.name
    set((state) => {
      // 新项目插入到第二个位置（第一个是永久置顶的）
      const firstItem = state.projectList[0]
      const restItems = state.projectList.slice(1)
      return {
        projectList: firstItem
          ? [firstItem, { id, name: projectName }, ...restItems]
          : [{ id, name: projectName }, ...restItems],
      }
    })
    return id
  },
  del: async (id, moveToFree = false) => {
    try {
      await deleteProject({ id, move_to_free: moveToFree })
      set((state) => ({
        projectList: state.projectList.filter((item) => item.id !== id),
      }))
    } catch (error) {
      console.log(error)
    }
  },
  rename: async (data) => {
    try {
      await renameProject(data)
      set((state) => ({
        projectList: state.projectList.map((item) =>
          item.id !== data.id ? item : data,
        ),
      }))
    } catch (error) {
      console.log(error)
    }
  },
}))

export default useProjectStore
