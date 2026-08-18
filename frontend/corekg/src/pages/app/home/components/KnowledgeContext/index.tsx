import { createContext, FC, PropsWithChildren, useContext } from 'react'
import { useRequest } from 'ahooks'
import { getFileList, getKnowledgeBaseList } from '@/api/knowledge'

export type Forest = { id: number; name: string }

export type FileItem = {
  id: number
  name: string
  parent_id: number | null
  is_dir: boolean
}

export type Knowledge = {
  forestList?: Forest[]
  fileList?: FileItem[]
}

const KnowledgeContext = createContext<Knowledge>({})
export const KnowledgeProvider: FC<PropsWithChildren> = (props) => {
  const { data: forestList } = useRequest(async () => {
    const forests: any[] =
      (
        await getKnowledgeBaseList({
          limit: 9999,
          offset: 0,
          filters: [
            {
              field: 'forest_type',
              value: ['file', 'qa', 'cad'],
              exactMatch: true,
            },
          ],
        })
      ).Data ?? []
    const forestList = forests.map((item) => {
      const { ID: id, name } = item
      return { id, name } as Forest
    })

    return forestList
  })

  const { data: fileList } = useRequest(
    async () => {
      if (!forestList) return undefined
      const fileList: FileItem[] = forestList.map((item) => {
        const { id, name } = item
        return { id, name, parent_id: null, is_dir: true }
      })

      await Promise.all(
        forestList.map(async (item) => {
          const { data: files } = await getFileList({
            limit: 9999,
            offset: 0,
            forest_id: item.id,
          })
          files?.forEach((fileItem: any) => {
            const {
              ID: id,
              name,
              parent_id: _parent_id,
              is_dir,
              forest_id,
              parse_status,
            } = fileItem
            if (!is_dir && parse_status !== 'success') {
              // 过滤没解析的文件
              return
            }
            const parent_id = _parent_id === 0 ? forest_id : _parent_id
            fileList.push({
              id,
              name,
              parent_id,
              is_dir,
            })
          })
        }),
      )

      return fileList
    },
    { refreshDeps: [forestList] },
  )

  return (
    <KnowledgeContext.Provider value={{ forestList, fileList }}>
      {props.children}
    </KnowledgeContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useKnowledge = () => useContext(KnowledgeContext)
