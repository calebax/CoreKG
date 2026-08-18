import { FC, useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import {
  Checkbox,
  Divider,
  Empty,
  Input,
  Skeleton,
  Tabs,
  Tag,
  Tree,
  Typography,
} from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { useDebounceFn, useMount, useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { match } from 'ts-pattern'
import globalConfig from '@/config'
import { cn } from '@/utils'
import { getKnowledgeBaseList, getTagTree } from '@/api/knowledge'
import HomeTableSearchIcon from '@/assets/icons/home/home-table-search.svg?react'
import { useSessionInfo } from '@/pages/project/ProjectContent'
import { withCheckboxStyle } from '@/pages/project/components/CheckboxStyleProvider'
import LastModel from '@/pages/project/components/ModelSelect/images/last-model.svg?react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import type { Knowledge } from '../..'
import { useKnowledgeData } from '../KnowledgeDataProvider'
import { TreeWithFilter } from './TreeWithFilter'
import styles from './index.module.css'

/** 知识选择器 用于浮窗 */
export const KnowledgeSelect: FC<
  Style & {
    graph?: boolean
    table?: boolean
    database?: boolean
    tabBarExtraContent?: React.ReactNode
    /** 全局搜索场景，只显示知识库节点，不显示文件层级 */
    globalSearch?: boolean
  }
> = withCheckboxStyle((props) => {
  const { graph, table, database, tabBarExtraContent, globalSearch } = props
  const { sessionConfig, setSessionConfig } = useSessionInfo()
  const { version, mode: deployMode } = useDeployConfig()

  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')

  const [search, setSearch] = useState<string>()
  const { run: deboSetSearch } = useDebounceFn(setSearch, { wait: 500 })

  const {
    knowledgeList: allForestData,
    loadData,
    hasCalledLoadFn,
    loading: allForestDataLoading,
  } = useKnowledgeData()
  useMount(async () => {
    if (!hasCalledLoadFn) {
      loadData()
    }
  })

  // 控制标签 Tab 显示 - 已注释，全局搜索场景下只显示知识库tab
  // const showTagTab =
  //   !graph &&
  //   !table &&
  //   !database &&
  //   (import.meta.env.MODE === 'development' ||
  //     import.meta.env.MODE === 'test' ||
  //     import.meta.env.MODE === 'production' ||
  //     (version === 'custom' && deployMode === 'cimc'))
  const showTagTab = false

  // 内部子 Tab 状态 - 已注释，全局搜索场景下只使用知识库tab
  // const [activeTab, setActiveTab] = useState<'knowledge' | 'tag'>(
  //   sessionConfig?.tag_ids?.length ? 'tag' : 'knowledge',
  // )
  const [activeTab] = useState<'knowledge' | 'tag'>('knowledge')

  // 获取标签树数据 - 已注释，全局搜索场景下不使用标签
  // const { data: tagData, loading: tagLoading } = useRequest(getTagTree, {
  //   ready: showTagTab,
  // })

  // 可以进行图谱问答的知识库（仅 graph_status 为 success 的）
  const { data: graphForests, loading: graphLoading } = useRequest(async () => {
    const res = await getKnowledgeBaseList({ offset: 0, limit: 9999 })
    const data: any[] = res.Data ?? []
    return data.filter((item) => item.graph_status === 'success')
  })

  const loading = allForestDataLoading || (graph ? graphLoading : false)

  // 格式化标签树数据 - 已注释，全局搜索场景下不使用标签
  // const formattedTagTree = useMemo(() => {
  //   if (!tagData?.group_list) return []
  //   return tagData.group_list.map((group: any) => ({
  //     title: group.tag_group_name,
  //     key: `group-${group.tag_group_id}`,
  //     children: (group.tag_list || []).map((tag: any) => ({
  //       title: tag.tag_name,
  //       key: `tag-${tag.tag_id}`,
  //       isLeaf: true,
  //     })),
  //   }))
  // }, [tagData])

  // 映射 ID 到 Key 的辅助逻辑 - 已注释，全局搜索场景下不使用标签
  // const getCheckedKeys = (ids: number[]) => {
  //   const keys: string[] = []
  //   ids.forEach((id) => {
  //     if (id >= 0) {
  //       keys.push(`tag-${id}`)
  //     } else {
  //       keys.push(`group-${-id - 1}`)
  //     }
  //   })
  //   return keys
  // }

  // 更新父节点选中状态的辅助函数 - 已注释，全局搜索场景下不使用标签
  // const updateParentGroupStatus = (
  //   tagIds: number[],
  //   changedTagId: number,
  // ): number[] => {
  //   const tagKeys = tagIds.filter((id) => id >= 0).map((id) => `tag-${id}`)
  //   const parentGroup = tagData?.group_list?.find((g: any) =>
  //     g.tag_list?.some((t: any) => t.tag_id === changedTagId),
  //   )

  //   if (!parentGroup) return tagIds

  //   const groupId = parentGroup.tag_group_id
  //   const childTagIds = (parentGroup.tag_list || []).map((t: any) => t.tag_id)
  //   const childTagKeys = childTagIds.map((id: number) => `tag-${id}`)
  //   const allChildrenSelected = childTagKeys.every((key) =>
  //     tagKeys.includes(key),
  //   )

  //   const encodedGroupId = -groupId - 1
  //   const hasGroupId = tagIds.includes(encodedGroupId)

  //   if (allChildrenSelected && !hasGroupId) {
  //     // 所有子标签都被选中，添加父节点
  //     return [...tagIds, encodedGroupId]
  //   } else if (!allChildrenSelected && hasGroupId) {
  //     // 有子标签未选中，移除父节点
  //     return tagIds.filter((id) => id !== encodedGroupId)
  //   }

  //   return tagIds
  // }

  // 获取所有子节点（最底层的可选项）
  const getAllAtomNodes = (nodes: Knowledge[]): Knowledge[] => {
    // 全局搜索：知识库节点本身就是原子节点
    if (globalSearch) {
      return nodes.filter((node) => node.node_type === 'forest')
    }
    // 表格、数据库、图谱、文档模式：返回文件级别的原子节点
    return nodes.filter((node) => node.knowledgeType !== 'other')
  }

  // 获取节点的知识库 ID（兼容 forest_id、id、ID 三种字段）
  const getForestId = (item: any) =>
    item.forest_id ?? item.id ?? item.ID ?? Number(item.key)

  const knowledgeList = useMemo(() => {
    if (!allForestData) return allForestData
    if (graph && graphForests) {
      // 图谱模式下：筛选已构建图谱的知识库及其文件子节点（需保留完整树结构）
      const graphForestIds = new Set(graphForests.map((item) => item.ID))
      const keyToNode = new Map(allForestData.map((n) => [n.key, n]))
      const getNodeForestId = (item: any): number | null => {
        if (item.node_type === 'forest') return getForestId(item)
        if (item.forest_id != null) return item.forest_id
        // 沿 parentKey 向上查找 forest 节点
        let cursor: any = item
        while (cursor?.parentKey) {
          const parent = keyToNode.get(cursor.parentKey)
          if (!parent) break
          if (parent.node_type === 'forest') return getForestId(parent)
          cursor = parent
        }
        return null
      }
      return allForestData?.filter((item) => {
        const fid = getNodeForestId(item)
        return fid != null && graphForestIds.has(fid)
      })
    }
    if (table) {
      // 表格模式下 显示表格类型的知识库及其文件
      // 筛选条件：forest_type=='data' && forest_data_source_type=='excel'
      return allForestData?.filter(
        (item) =>
          item.forest_type === 'data' &&
          item.forest_data_source_type === 'excel',
      )
    }
    if (database) {
      // 数据库模式下 显示数据库类型的知识库及其表
      // 需要保留完整的层级结构：数据库知识库 → 数据库实例 → 数据库表
      const dbForests = allForestData?.filter(
        (item) =>
          item.node_type === 'forest' &&
          item.forest_type === 'data' &&
          item.forest_data_source_type === 'db',
      )
      const dbForestIds = new Set(dbForests?.map((f) => f.forest_id))
      
      // 筛选：1. 数据库知识库节点  2. 属于这些知识库的所有子节点（包括数据库实例和表）
      return allForestData?.filter(
        (item) =>
          (item.node_type === 'forest' &&
            item.forest_type === 'data' &&
            item.forest_data_source_type === 'db') ||
          dbForestIds.has(item.forest_id),
      )
    }
    // 全局搜索场景下，只显示知识库节点（node_type === 'forest'），不显示文件节点
    // 排除表格类型知识库（forest_type === 'data' && forest_data_source_type === 'excel'）
    if (globalSearch) {
      return allForestData?.filter(
        (item) =>
          item.node_type === 'forest' &&
          !(
            item.forest_type === 'data' &&
            item.forest_data_source_type === 'excel'
          ),
      )
    }
    // 其他场景（文档模式），返回全部数据（排除表格类型知识库和表格文件节点）
    return allForestData?.filter(
      (item) =>
        !['excel_sheet', 'mysql_table'].includes(item.knowledgeType) &&
        !(
          item.forest_type === 'data' &&
          item.forest_data_source_type === 'excel'
        ),
    )
  }, [allForestData, graph, graphForests, table, database, globalSearch])

  // 获取所有子节点
  const allAtomNodes = useMemo(() => {
    if (!knowledgeList) return []
    return getAllAtomNodes(knowledgeList)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [knowledgeList, globalSearch])

  // 获取当前选中的知识库
  const currentKnowledge = useMemo(() => {
    return graph
      ? sessionConfig?.graphKnowledgeBase
      : table
        ? sessionConfig?.tableKnowledgeBase
        : database
          ? sessionConfig?.databaseKnowledgeBase
          : sessionConfig?.knowledge
  }, [
    graph,
    table,
    database,
    sessionConfig?.graphKnowledgeBase,
    sessionConfig?.tableKnowledgeBase,
    sessionConfig?.databaseKnowledgeBase,
    sessionConfig?.knowledge,
  ])

  // 判断是否全选
  const isSelectAll = useMemo(() => {
    if (!currentKnowledge?.length || !allAtomNodes.length) return false
    const selectedKeys = new Set(currentKnowledge.map((item) => item.key))
    return allAtomNodes.every((node) => selectedKeys.has(node.key))
  }, [currentKnowledge, allAtomNodes])

  // 设置知识库配置的辅助函数
  const setKnowledgeConfig = (knowledge: Knowledge[]) => {
    if (graph) {
      setSessionConfig({ graphKnowledgeBase: knowledge })
    } else if (table) {
      setSessionConfig({ tableKnowledgeBase: knowledge })
    } else if (database) {
      setSessionConfig({ databaseKnowledgeBase: knowledge })
    } else {
      setSessionConfig({
        knowledge,
        tag_ids: undefined,
        externalIds: undefined,
      })
    }
  }

  // 渲染标签内容 - 已注释，全局搜索场景下不使用标签
  // const renderTagContent = () => {
  //   const selectedTags = sessionConfig?.tag_ids || []
  //   const recentTags = tagData?.recent_tag_list || []

  //   return (
  //     <div className='flex flex-col gap-1 px-2.5 py-2'>
  //       {/* 最近使用 - 参考 ModelSelect 样式 */}
  //       <div className='flex flex-col gap-2'>
  //         <div className='flex items-center gap-1.5 text-[#616373] text-sm font-medium'>
  //           <LastModel className='w-4 h-4' />
  //           <span>{t('app.home.recentUsed')}</span>
  //         </div>
  //         {recentTags.length > 0 ? (
  //           <div className='flex flex-wrap gap-2'>
  //             {recentTags.map((tag: any) => (
  //               <Typography.Paragraph
  //                 key={tag.tag_id}
  //                 className={cn(
  //                   'cursor-pointer bg-[#0C99FF1A] text-[#0C99FF] rounded-full',
  //                   'px-2.5 py-0.5 m-0 text-xs transition-colors',
  //                   selectedTags.includes(tag.tag_id)
  //                     ? 'bg-[#0C99FF] text-white'
  //                     : 'hover:bg-[#0C99FF33]',
  //                 )}
  //                 onClick={() => {
  //                   const isSelected = selectedTags.includes(tag.tag_id)
  //                   let newTags = isSelected
  //                     ? selectedTags.filter((id: number) => id !== tag.tag_id)
  //                     : [...selectedTags, tag.tag_id]
  //                   // 更新父节点状态
  //                   newTags = updateParentGroupStatus(newTags, tag.tag_id)
  //                   setSessionConfig({
  //                     tag_ids: newTags,
  //                     knowledge: undefined,
  //                     externalIds: undefined,
  //                   })
  //                 }}
  //                 ellipsis={{ rows: 1, tooltip: tag.tag_name }}
  //               >
  //                 {tag.tag_name}
  //               </Typography.Paragraph>
  //             ))}
  //           </div>
  //         ) : (
  //           // 为空时留出位置
  //           <div className='h-6' />
  //         )}
  //         <Divider className='my-1' />
  //       </div>
  //       {/* 标签树 */}
  //       <Tree
  //         checkable
  //         checkStrictly
  //         selectable={false}
  //         className='tag-tree'
  //         treeData={formattedTagTree}
  //         checkedKeys={getCheckedKeys(selectedTags)}
  //         onCheck={(checked, info) => {
  //           const nodeKey = info.node.key as string
  //           const isChecked = info.checked
  //           let newKeys = [...getCheckedKeys(selectedTags)]

  //           if (nodeKey.startsWith('group-')) {
  //             // 选中/取消选中分类及其所有子标签
  //             const groupId = Number(nodeKey.replace('group-', ''))
  //             const encodedGroupId = -groupId - 1
  //             const group = tagData.group_list.find(
  //               (g: any) => g.tag_group_id === groupId,
  //             )
  //             const childTagIds = (group?.tag_list || []).map(
  //               (t: any) => t.tag_id,
  //             )

  //             if (isChecked) {
  //               newKeys = Array.from(
  //                 new Set([
  //                   ...newKeys,
  //                   nodeKey,
  //                   ...childTagIds.map((id: number) => `tag-${id}`),
  //                 ]),
  //               )
  //             } else {
  //               newKeys = newKeys.filter(
  //                 (k) =>
  //                   k !== nodeKey &&
  //                   !childTagIds.some((id: number) => `tag-${id}` === k),
  //               )
  //             }
  //           } else {
  //             // 选中/取消选中单个标签
  //             if (isChecked) {
  //               newKeys.push(nodeKey)
  //             } else {
  //               newKeys = newKeys.filter((k) => k !== nodeKey)
  //             }

  //             // 检查并更新父节点（分类）的选中状态
  //             const tagId = Number(nodeKey.replace('tag-', ''))
  //             const parentGroup = tagData.group_list.find((g: any) =>
  //               g.tag_list?.some((t: any) => t.tag_id === tagId),
  //             )

  //             if (parentGroup) {
  //               const groupKey = `group-${parentGroup.tag_group_id}`
  //               const childTagIds = (parentGroup.tag_list || []).map(
  //                 (t: any) => t.tag_id,
  //               )
  //               const childTagKeys = childTagIds.map(
  //                 (id: number) => `tag-${id}`,
  //               )
  //               // 检查该分类下的所有子标签是否都被选中
  //               const allChildrenSelected = childTagKeys.every((key) =>
  //                 newKeys.includes(key),
  //               )

  //               if (allChildrenSelected) {
  //                 // 所有子标签都被选中，自动选中父节点
  //                 if (!newKeys.includes(groupKey)) {
  //                   newKeys.push(groupKey)
  //                 }
  //               } else {
  //                 // 有子标签未选中，取消选中父节点
  //                 newKeys = newKeys.filter((k) => k !== groupKey)
  //               }
  //             }
  //           }

  //           // 转回 ID
  //           const newIds = newKeys.map((k) => {
  //             if (k.startsWith('tag-')) return Number(k.replace('tag-', ''))
  //             return -Number(k.replace('group-', '')) - 1
  //           })

  //           setSessionConfig({
  //             tag_ids: Array.from(new Set(newIds)),
  //             knowledge: undefined,
  //             externalIds: undefined,
  //           })
  //         }}
  //       />
  //     </div>
  //   )
  // }

  const renderKnowledgeContent = () => {
    return (
      <div className='flex flex-col gap-[3px]'>
        <div className='font-medium'>
          <div className='flex items-center gap-1 px-[5px] py-2 border-b border-solid border-[#EEEEEE]'>
            {/* 文档模式和表格模式显示全选框 */}
            {(table || (!graph && !database)) && (
              <Checkbox
                checked={isSelectAll}
                onChange={(e) => {
                  // 全选或取消全选
                  setKnowledgeConfig(e.target.checked ? allAtomNodes : [])
                }}
              />
            )}
            {table ? (
              <span>
                已选表格（{currentKnowledge?.length || 0}/{allAtomNodes.length}
                ）
              </span>
            ) : database ? (
              <>
                <span>
                  已选数据表（{currentKnowledge?.length || 0}/
                  {allAtomNodes.length}）
                </span>
              </>
            ) : (
              <>
                <span>
                  已选资源（{currentKnowledge?.length || 0}/
                  {allAtomNodes.length}）
                </span>
              </>
            )}
          </div>
        </div>
        <TreeWithFilter
          graph={graph}
          search={search}
          nodes={knowledgeList}
          value={currentKnowledge}
          table={table}
          database={database}
          allowCrossForest={table || database}
          globalSearch={globalSearch}
          onChange={(knowledge) => {
            setKnowledgeConfig(knowledge)
          }}
        />
      </div>
    )
  }

  if (!hasCalledLoadFn || loading) {
    return <Skeleton active className='p-4 w-40' />
  }
  if (!knowledgeList?.length) {
    return (
      <span className='p-4 text-xs text-[#6E757F] leading-6'>
        暂无可用
        {match({ table, database })
          .with({ table: true }, () => '表格')
          .with({ database: true }, () => '数据表')
          .otherwise(() => '资源')}
        ，
        <Link className='text-xs ml-1' to='/docs'>
          立即创建
          <ArrowRightOutlined />
        </Link>
      </span>
    )
  }

  const content = (
    <div
      className={cn(
        'flex flex-col gap-[3px] p-2.5 pb-5 overflow-auto min-w-80',
        props.className,
        props.checkboxClassName,
      )}
      style={props.style}
    >
      {/* 标签tab已注释，全局搜索场景下只显示知识库tab */}
      {/* {showTagTab ? (
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            setActiveTab(key as any)
            // 切换 Tab 时清空另一侧的选择
            if (key === 'knowledge') {
              setSessionConfig({ tag_ids: undefined })
            } else {
              setSessionConfig({ knowledge: undefined, externalIds: undefined })
            }
          }}
          className='resource-tabs flex-1'
          tabBarExtraContent={tabBarExtraContent}
          items={[
            {
              key: 'knowledge',
              label: t('app.home.knowledgeBase'),
              children: (
                <div className='flex flex-col gap-2'>
                  <Input
                    onChange={(e) => deboSetSearch(e.target.value)}
                    allowClear
                    onClear={() => setSearch('')}
                    placeholder={tC('button.search')}
                    prefix={<HomeTableSearchIcon className='w-4 h-4 mr-1.5' />}
                    className={cn('rounded', styles.searchInput)}
                  />
                  {renderKnowledgeContent()}
                </div>
              ),
            },
            {
              key: 'tag',
              label: t('app.home.tag'),
              children: (
                <div className='flex flex-col gap-2'>
                  {tagLoading ? (
                    <Skeleton active className='p-4' />
                  ) : (
                    renderTagContent()
                  )}
                </div>
              ),
            },
          ]}
        />
      ) : ( */}
      <div className='flex flex-col gap-2'>
        {tabBarExtraContent && (
          <div className='flex items-center justify-end px-2 pb-2'>
            {tabBarExtraContent}
          </div>
        )}
        <Input
          onChange={(e) => deboSetSearch(e.target.value)}
          allowClear
          onClear={() => setSearch('')}
          placeholder={tC('button.search')}
          prefix={<HomeTableSearchIcon className='w-4 h-4 mr-1.5' />}
          className={cn('rounded', styles.searchInput)}
        />
        {renderKnowledgeContent()}
      </div>
      {/* )} */}
    </div>
  )

  return content
})
