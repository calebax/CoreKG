import { createContext, FC } from 'react'
import { useRequest } from 'ahooks'
import { getAllForestData } from '@/api/knowledge'
import { Knowledge } from '../..'

const KnowledgeDataContext = createContext<{
  knowledgeList: Knowledge[]
  loading: boolean
  loadData: () => void
  /** 判断数据是否进行了第一次加载 */
  hasCalledLoadFn: boolean
} | null>(null)

export const useKnowledgeData = () => {
  const val = useContext(KnowledgeDataContext)
  if (!val) throw new Error('必须被KnowledgeDataContext包裹')
  return val
}
/** 配合KnowledgeSelect组件 */
export function withKnowledgeDataProvider<T>(Comp: FC<T>): FC<T>
export function withKnowledgeDataProvider<T>(Comp: any): FC<T> {
  const CompWithProvider: FC<T> = (props) => {
    const called = useRef(false)
    const {
      data = [],
      loading,
      run,
    } = useRequest(
      async () => {
        called.current = true
        const { tree } = await getAllForestData()
        const knowledgeList: Knowledge[] = []
        // 添加 parentKey 参数，用于保留父子关系以构建树形结构
        const insertNode = (node: any, parentKey?: string) => {
          if (node.forest_type === 'cad') return
          const knowledgeType: Knowledge['knowledgeType'] = (() => {
            // qa数据库是原子的
            if (node.node_type === 'forest' && node.forest_type === 'qa')
              return 'qa'
            if (['file', 'excel_sheet', 'mysql_table'].includes(node.node_type))
              return node.node_type
            // 表格模式下的excel文件节点（后端不再返回sheet层级，文件即为原子节点）
            if (
              node.node_type === 'excel' &&
              node.forest_type === 'data' &&
              node.forest_data_source_type === 'excel'
            )
              return 'excel_sheet'
            return 'other'
          })()
          const nodeData = {
            ...node,
            knowledgeType,
            parentKey,
            children: undefined,
          }
          // 确保 forest 节点有 forest_id
          if (node.node_type === 'forest') nodeData.forest_id = node.id
          knowledgeList.push(nodeData)
          // 递归处理子节点，传递当前节点的 key 作为子节点的 parentKey
          node.children?.forEach((child: any) => insertNode(child, node.key))
        }
        tree?.forEach((node: any) => insertNode(node))
        return knowledgeList
      },
      { manual: true },
    )
    return (
      <KnowledgeDataContext.Provider
        value={{
          knowledgeList: data,
          loadData: run,
          loading,
          hasCalledLoadFn: called.current,
        }}
      >
        <Comp {...props} />
      </KnowledgeDataContext.Provider>
    )
  }
  return CompWithProvider
}
