import { FC, useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useBoolean, useMemoizedFn, useMount, useRequest } from 'ahooks'
import globalConfig from '@/config'
import { cn } from '@/utils'
import { createSession, createStream } from '@/api/agent'
import {
  getFileList,
  getKnowledgeBaseList,
  listForestDB,
  listForestTable,
} from '@/api/knowledge'
import Excel from '@/assets/icons/docs/Excel.svg?react'
import MySQL from '@/assets/icons/docs/MySQL.svg?react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { SendButton } from '../SendButton'
import { DataInput } from './DataInput'
import { KnowledgeSelect } from './KnowledgeSelect'

export type KnowledgeNode = {
  /** 当前元素id或数据库、表的name */
  id: number | string
  /** 父元素的id */
  parent_id?: number | string
  key: string
  title: string
  /** excel知识库、文件、表 mysql知识库、数据库、表 */
  type: 'excel' | 'file' | 'sheet' | 'mysql' | 'db' | 'table'
  /** 当前元素所属的知识库 */
  forest_id: number
  // 树行为
  icon?: any
  children?: KnowledgeNode[]
  isLeaf?: boolean
}

export const DataMode: FC<{ hidden?: boolean }> = (props) => {
  const { hidden } = props
  const navigate = useNavigate()

  const { version, mode: deployMode } = useDeployConfig()
  const showTagTab =
    globalConfig.apiEnv === 'test' ||
    (version === 'custom' &&
      (deployMode === 'cimc' || deployMode === 'h3c'))

  const [search, setSearch] = useState<string | undefined>()
  // 当前选择模式：知识库或标签
  const [mode, setMode] = useState<'knowledge' | 'tag'>('knowledge')
  // 已选中的知识
  const [selectedKeys, setKeys] = useState<string[]>()
  // 已选中的标签
  const [selectedTags, setTags] = useState<number[]>([])
  // 标签对应的文件数量
  const [tagFileCount, setTagFileCount] = useState<number>(0)

  const { knowledge, treeData, loadData } = useTreeData()

  // 当选择标签时，获取对应的文件数量
  useRequest(
    async () => {
      // 过滤出真正的标签 ID（正数或0）
      const actualTagIds = selectedTags.filter((id) => id >= 0)
      if (mode !== 'tag' || actualTagIds.length === 0) {
        setTagFileCount(0)
        return
      }
      const res = await getFileList({
        forest_id: 0, // 全局搜索
        filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
        limit: 1,
      })
      setTagFileCount(res.total ?? 0)
    },
    {
      refreshDeps: [selectedTags, mode],
    },
  )

  const { loading, refresh: ask } = useRequest(
    async () => {
      let session_id: number
      const actualTagIds = selectedTags.filter((id) => id >= 0)

      if (mode === 'knowledge') {
        // 如果只选择了标签，没有选择知识库资源
        if (
          (!selectedKeys || selectedKeys.length === 0) &&
          actualTagIds.length > 0
        ) {
          // 先获取标签对应的文件列表
          const fileListRes = await getFileList({
            forest_id: 0, // 全局搜索
            filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
            limit: 9999, // 获取所有文件
          })
          const fileIds = (fileListRes.data || []).map((file: any) => file.ID)

          const body: any = {
            base_type: 'standard',
            resource_type: 'file_list',
            ids: fileIds,
            model_id: 1,
            tag_ids: actualTagIds,
            tag_resourse_type: 'file',
          }
          const { ID } = await createSession(body)
          session_id = ID
        } else {
          // 选择了知识库资源（可能同时选择了标签）
          const nodes = selectedKeys!.map((k) => {
            return knowledge.find((item) => item.key === k)!
          })
          session_id = await createNodesSession(nodes, selectedTags)
        }
      } else {
        // 标签模式下的会话创建逻辑
        // 先获取标签对应的文件列表
        const fileListRes = await getFileList({
          forest_id: 0, // 全局搜索
          filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
          limit: 9999, // 获取所有文件
        })
        const fileIds = (fileListRes.data || []).map((file: any) => file.ID)

        const body: any = {
          base_type: 'standard',
          resource_type: 'file_list',
          ids: fileIds,
          model_id: 1,
          tag_ids: actualTagIds,
          tag_resourse_type: 'file',
        }
        const { ID } = await createSession(body)
        session_id = ID
      }

      const { question_id } = await createStream({
        session_id,
        question: search,
      })
      const searchParams = new URLSearchParams()
      searchParams.append('session_id', `${session_id}`)
      searchParams.append('question_id', question_id)
      navigate(`/QA?${searchParams.toString()}`)
    },
    { manual: true },
  )

  const isBtnActive = useMemo(() => {
    // 知识库模式：需要选择知识库资源或标签
    // 标签模式：需要选择标签
    const actualTagIds = selectedTags.filter((id) => id >= 0)
    const hasResource =
      mode === 'knowledge'
        ? (selectedKeys && selectedKeys.length > 0) || actualTagIds.length > 0
        : actualTagIds.length > 0
    return Boolean(hasResource && search?.trim() && !loading)
  }, [mode, selectedKeys, selectedTags, search, loading])

  return (
    <div className={cn('w-[50vw]', { hidden })}>
      <div className='bg-[rgba(255,255,255,0.42)] relative rounded-[20px] w-full border border-[#e6e8f0] border-solid shadow-[0px_4px_10px_0px_rgba(0,0,0,0.1)] min-h-[128px]  p-4 pb-14'>
        <DataInput value={search} onChange={setSearch} />
        <div className='absolute bottom-0 left-0 right-0 flex items-center px-4 py-3'>
          <KnowledgeSelect
            mode={mode}
            onModeChange={(m) => {
              setMode(m)
              // 切换tab时清空另一侧的选择
              if (m === 'knowledge') {
                setTags([])
              } else {
                setKeys([])
              }
            }}
            knowledge={knowledge}
            value={selectedKeys}
            onChange={setKeys}
            selectedTags={selectedTags}
            onTagsChange={setTags}
            tagFileCount={tagFileCount}
            treeData={treeData}
            loadData={loadData}
          />
          <SendButton active={isBtnActive} className='ml-auto' onClick={ask} />
        </div>
      </div>
    </div>
  )
}

const useTreeData = () => {
  const [loading, { toggle }] = useBoolean(true)
  const [knowledge, setKnowledge] = useState<KnowledgeNode[]>([])
  useMount(async () => {
    const forests: any[] =
      (
        await getKnowledgeBaseList({
          limit: 9999,
          offset: 0,
          filters: [
            {
              field: 'forest_type',
              value: ['data'],
              exactMatch: true,
            },
          ],
        })
      ).Data ?? []
    setKnowledge(
      forests.map((item) => {
        const { ID: forest_id } = item
        const type = item.data_source_subtype
        const icon = (() => {
          switch (type) {
            case 'excel':
              return <Excel className='h-full w-full mr-2' />
            case 'mysql':
              return <MySQL className='h-full w-full mr-2' />
            default:
              return null
          }
        })()
        const key = `forest-${forest_id}`
        return {
          id: forest_id,
          key,
          forest_id,
          title: item.name,
          type,
          icon,
        }
      }),
    )
    toggle()
  })
  const loadData = useMemoizedFn(async (node: KnowledgeNode) => {
    const { type, key, id, forest_id } = node
    // 将新的node加入树结构及扁平化的数组
    const addNodes = (nodes: KnowledgeNode[]) => {
      setKnowledge((prev) => {
        const newNodes = [...prev]
        newNodes.push(...nodes)
        const target = newNodes.find((item) => item.key === key)!
        target.children = nodes
        return newNodes
      })
    }
    // 按照被点击节点的类型区分
    switch (type) {
      case 'excel': {
        // excel知识库
        const data: any[] =
          (
            await getFileList({
              forest_id: id as number,
              filters: [
                { field: 'parse_status', value: ['success'], exactMatch: true },
                { field: 'is_dir', value: ['-1'] },
              ],
            })
          ).data ?? []
        const files: KnowledgeNode[] = data.map((item) => {
          const { ID, name } = item
          return {
            id: ID,
            key: `${key}-${ID}`,
            title: name,
            type: 'file',
            forest_id,
            // 后端不再返回sheet层级，文件即为叶子节点
            isLeaf: true,
          }
        })
        addNodes(files)
        break
      }
      case 'file': {
        // excel文件 - 后端不再返回sheet层级，文件即为叶子节点
        // 将file节点标记为叶子节点，不再加载子节点
        setKnowledge((prev) => {
          const newNodes = [...prev]
          const target = newNodes.find((item) => item.key === key)!
          target.isLeaf = true
          return newNodes
        })
        break
      }
      case 'mysql': {
        const res = await listForestDB({
          forest_id: id as number,
        })
        const db_list: any[] = res.db_list ?? []
        addNodes(
          db_list.map((item) => {
            const { db_name } = item
            return {
              id: db_name,
              key: `${key}-${db_name}`,
              type: 'db',
              title: db_name,
              forest_id,
            }
          }),
        )
        break
      }
      case 'db': {
        const res = await listForestTable({
          forest_id,
          forest_db_name: id as string,
        })
        const table_list: any[] = res.table_list
        addNodes(
          table_list.map((item) => {
            const { forest_table_name } = item
            return {
              id: forest_table_name,
              parent_id: id,
              key: `${key}-${forest_table_name}`,
              type: 'table',
              title: forest_table_name,
              isLeaf: true,
              forest_id,
            }
          }),
        )
        break
      }
      default:
        // 没有子节点
        break
    }
  })
  const treeData = useMemo(() => {
    if (loading) return undefined
    return knowledge.filter((item) => {
      switch (item.type) {
        case 'excel':
        case 'mysql':
          return true
        default:
          return false
      }
    })
  }, [knowledge, loading])
  return { knowledge, treeData, loadData }
}

const createNodesSession = async (
  nodes: KnowledgeNode[],
  selectedTags?: number[],
): Promise<number> => {
  // 确保选中的是同一类型 同一知识库的资源
  const { type, forest_id } = nodes[0]
  // id或name
  const ids = nodes.map((item) => item.id)
  const body: any = {
    resource_id: forest_id,
    model_id: 1,
  }
  switch (type) {
    case 'excel': {
      body.base_type = 'excel'
      body.ids = ids
      body.resource_type = 'forest'
      break
    }
    case 'file': {
      body.base_type = 'excel'
      body.ids = ids
      body.resource_type = 'excel_list'
      break
    }
    case 'sheet': {
      body.base_type = 'react_excel'
      body.ids = ids
      body.resource_type = 'react_excel_list'
      break
    }
    case 'mysql': {
      body.base_type = 'mysql'
      body.ids = ids
      body.resource_type = 'forest'
      break
    }
    case 'db': {
      body.base_type = 'mysql'
      body.names = ids
      body.resource_type = 'db_list'
      delete body.ids
      break
    }
    case 'table': {
      body.base_type = 'mysql'
      body.names = ids
      body.resource_type = 'db_table_list'
      break
    }
  }
  // 如果选择了标签，添加标签相关参数
  if (selectedTags && selectedTags.length > 0) {
    const actualTagIds = selectedTags.filter((id) => id >= 0)
    if (actualTagIds.length > 0) {
      body.tag_ids = actualTagIds
      body.tag_resourse_type = 'file'
    }
  }
  const { ID } = await createSession(body)
  return ID
}
