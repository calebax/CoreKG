import { FC, useMemo } from 'react'
import { Button, Tree } from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { match } from 'ts-pattern'
import { uniqueArray } from '@/utils'
import { type Knowledge } from '../../..'
import { KnowledgeIcon } from '../../KnowledgeIcon'
import { ForestTitleInGraph } from './ForestTitleInGraph'
import styles from './styles.module.css'

export type TreeWithFilter = {
  nodes: Knowledge[]
  value?: Knowledge[]
  onChange: (val: Knowledge[]) => void
  search?: string
  table?: boolean
  database?: boolean
  allowCrossForest?: boolean
  graph?: boolean
  /** 全局搜索场景，只显示知识库节点，不显示文件层级 */
  globalSearch?: boolean
}
export const TreeWithFilter: FC<TreeWithFilter> = (props) => {
  const {
    nodes,
    value = [],
    onChange,
    search,
    table,
    database,
    allowCrossForest,
    graph,
    globalSearch,
  } = props
  const selectedKeys = useMemo(() => {
    return value.map((item) => item.key)
  }, [value])
  /** 用于获取节点基本信息 */
  const originTreeMap = useMemo(() => {
    const _treeMap = new Map<string, Knowledge>()
    nodes.forEach((node) => {
      _treeMap.set(node.key, node)
    })
    return _treeMap
  }, [nodes])

  // 判断是否使用简化模式（只显示知识库节点，不显示文件层级）
  // 只有全局搜索使用简化模式
  const useSimpleMode = globalSearch

  // 和originTreeMap的节点数据不同 是筛选后的结果
  const treeData = useMemo(() => {
    // 简化模式：只显示知识库节点，不显示文件节点
    if (useSimpleMode) {
      return arrayToTreeSimple(nodes, search, graph)
    }
    // 表格、数据库、图谱、文档模式：使用完整的树形结构逻辑
    return arrayToTree(nodes, search, graph)
  }, [graph, nodes, search, useSimpleMode])

  // 简化模式的处理逻辑（仅全局搜索）
  if (useSimpleMode) {
    return (
      <Tree
        showIcon
        className={styles.tree}
        treeData={treeData}
        selectable={false}
        checkable
        checkedKeys={selectedKeys}
        defaultExpandAll={false}
        onCheck={(_, info) => {
          const changedKey = info.node.key as string
          const changedNode = originTreeMap.get(changedKey)
          if (!changedNode) return

          // 只处理知识库节点
          if (changedNode.node_type !== 'forest') return

          let newSelectedKeys: string[] = []
          if (!info.checked) {
            // 取消选中
            newSelectedKeys = selectedKeys.filter((k) => k !== changedKey)
          } else {
            // 选中知识库节点
            // 表格和数据库模式允许多选，全局搜索场景也允许多选
            if (selectedKeys.includes(changedKey)) {
              newSelectedKeys = selectedKeys
            } else {
              newSelectedKeys = [...selectedKeys, changedKey]
            }
          }
          // 过滤掉可能不存在的节点
          onChange(
            newSelectedKeys
              .map((k) => originTreeMap.get(k))
              .filter(Boolean) as Knowledge[],
          )
        }}
      />
    )
  }

  // 构建 treeData 的节点映射（包含 children 信息）
  const treeNodeMap = useMemo(() => {
    const map = new Map<string, any>()
    const traverse = (nodes: any[]) => {
      nodes.forEach((node) => {
        map.set(node.key, node)
        if (node.children) {
          traverse(node.children)
        }
      })
    }
    traverse(treeData)
    return map
  }, [treeData])

  // 完整树形模式的处理逻辑（文档、图谱）
  return (
    <Tree
      showIcon
      className={styles.tree}
      treeData={treeData}
      selectable={false}
      checkable
      checkedKeys={selectedKeys}
      defaultExpandAll={false}
      onCheck={(_, info) => {
        const changedKey = info.node.key as string
        const changedNode = originTreeMap.get(changedKey)
        // 从 treeNodeMap 获取带有 children 信息的节点
        const treeNode = treeNodeMap.get(changedKey)
        if (!changedNode) return

        // 判断节点是否是原子节点（文件/excel_sheet/mysql_table/qa）
        const isAtomNode = (node: any) => node.knowledgeType !== 'other'

        // 获取节点及其所有子孙的原子节点keys（使用 treeNode 的 children）
        const getAtomKeys = (node: any): string[] => {
          if (isAtomNode(node)) return [node.key]
          const result: string[] = []
          node.children?.forEach((child: any) => {
            result.push(...getAtomKeys(child))
          })
          return result
        }

        let newSelectedKeys: string[] = []
        if (!info.checked) {
          // 取消选中：移除该节点及其所有子孙的原子节点
          const keysToRemove = new Set(getAtomKeys(treeNode || changedNode))
          newSelectedKeys = selectedKeys.filter((k) => !keysToRemove.has(k))
        } else {
          // 选中
          if (isAtomNode(changedNode)) {
            // 原子节点：直接添加到选中列表
            // 如果不允许跨知识库，则只保留同一个知识库的文件
            if (allowCrossForest) {
              // 允许多选：直接添加
              newSelectedKeys = uniqueArray([...selectedKeys, changedKey])
            } else {
              // 不允许跨知识库（主要用于表格/数据库单选场景）
              // 但文档模式和图谱模式下这个参数是 false，却需要支持多选
              // 所以这里的逻辑应该是：添加到已选列表，而不是替换
              newSelectedKeys = uniqueArray([...selectedKeys, changedKey])
            }
          } else {
            // 非原子节点：选中所有子孙原子节点
            const atomKeys = getAtomKeys(treeNode || changedNode)
            newSelectedKeys = uniqueArray([...selectedKeys, ...atomKeys])
          }
        }
        onChange(
          newSelectedKeys.map((k) => originTreeMap.get(k)!).filter(Boolean),
        )
      }}
    />
  )
}

/**
 * 简化模式：将节点数组转化树结构（只显示知识库节点）
 * 用于全局搜索、表格、数据库模式
 * @param nodes 节点数组
 * @param search 筛选字符串
 * @param graph 是否图谱模式
 */
const arrayToTreeSimple = (
  nodes: Knowledge[],
  search?: string,
  graph?: boolean,
) => {
  // 全局搜索场景下，只显示知识库节点（node_type === 'forest'）
  const filteredNodes = nodes.filter((node) => node.node_type === 'forest')

  // 应用搜索过滤
  let result = filteredNodes
  if (search) {
    result = filteredNodes.filter((node) => node.name.includes(search))
  }

  // 转换为树节点格式
  return result.map((node) => {
    const { key, name, node_type, forest_type, forest_data_source_type } = node
    const icon = (() => {
      if (node_type !== 'forest') return undefined
      if (forest_type !== 'data')
        return <KnowledgeIcon className='flex-shrink-0' type={forest_type} />
      return (
        <KnowledgeIcon
          className='flex-shrink-0'
          type={forest_data_source_type}
        />
      )
    })()

    const treeNode: any = {
      ...node,
      title: name,
      isLeaf: true,
      children: undefined,
    }
    if (icon) {
      treeNode.icon = icon
    }
    if (graph && node_type === 'forest') {
      treeNode.title = <ForestTitleInGraph id={node.forest_id} title={name} />
    }
    // 应用搜索高亮
    if (search && name.includes(search)) {
      const seqs = name.split(search)
      treeNode.title = (
        <>
          {seqs.map((s, i) => {
            if (i === seqs.length - 1) return s
            return (
              <span key={i}>
                {s}
                <span className='bg-[#0C99FF]/20'>{search}</span>
              </span>
            )
          })}
        </>
      )
    }
    return treeNode
  })
}

/**
 * 完整树形模式：将节点数组转化树结构（保留文件层级）
 * 用于文档模式、图谱模式
 * @param nodes 节点数组
 * @param search 筛选字符串
 * @param graph 是否图谱模式
 */
const arrayToTree = (nodes: Knowledge[], search?: string, graph?: boolean) => {
  // 构建父子关系映射
  const keyToNode = new Map<string, Knowledge>()
  nodes.forEach((node) => keyToNode.set(node.key, node))

  // 构建树结构
  const buildTree = (nodeList: Knowledge[]): any[] => {
    // 找出所有根节点（没有父节点或父节点不在列表中的节点）
    const rootNodes = nodeList.filter(
      (node) => !node.parentKey || !keyToNode.has(node.parentKey),
    )

    // 构建子节点映射
    const childrenMap = new Map<string, Knowledge[]>()
    nodeList.forEach((node) => {
      if (node.parentKey && keyToNode.has(node.parentKey)) {
        const children = childrenMap.get(node.parentKey) || []
        children.push(node)
        childrenMap.set(node.parentKey, children)
      }
    })

    // 递归构建树节点
    const buildTreeNode = (node: Knowledge): any => {
      const {
        key,
        name,
        node_type,
        forest_type,
        forest_data_source_type,
        knowledgeType,
      } = node
      const children = childrenMap.get(key)

      const icon = (() => {
        if (node_type === 'forest') {
          if (forest_type !== 'data')
            return (
              <KnowledgeIcon className='flex-shrink-0' type={forest_type} />
            )
          return (
            <KnowledgeIcon
              className='flex-shrink-0'
              type={forest_data_source_type}
            />
          )
        }
        return undefined
      })()

      let title: React.ReactNode = name
      // 应用搜索高亮
      if (search && name.includes(search)) {
        const seqs = name.split(search)
        title = (
          <>
            {seqs.map((s, i) => {
              if (i === seqs.length - 1) return s
              return (
                <span key={i}>
                  {s}
                  <span className='bg-[#0C99FF]/20'>{search}</span>
                </span>
              )
            })}
          </>
        )
      }
      if (graph && node_type === 'forest') {
        title = <ForestTitleInGraph id={node.forest_id} title={name} />
      }

      const treeNode: any = {
        ...node,
        title,
        isLeaf: knowledgeType !== 'other',
        children: children?.map(buildTreeNode),
      }
      if (icon) {
        treeNode.icon = icon
      }
      return treeNode
    }

    return rootNodes.map(buildTreeNode)
  }

  // 搜索过滤
  let filteredNodes = nodes
  if (search) {
    // 如果有搜索，保留匹配的节点及其祖先
    const matchedKeys = new Set<string>()
    const addAncestors = (node: Knowledge) => {
      matchedKeys.add(node.key)
      if (node.parentKey && keyToNode.has(node.parentKey)) {
        addAncestors(keyToNode.get(node.parentKey)!)
      }
    }
    nodes.forEach((node) => {
      if (node.name.includes(search)) {
        addAncestors(node)
      }
    })
    filteredNodes = nodes.filter((node) => matchedKeys.has(node.key))
  }

  // 过滤空文件夹
  const filterEmptyDir = (treeNodes: any[]): any[] => {
    return treeNodes
      .map((node) => {
        if (node.isLeaf) return node
        const filteredChildren = node.children
          ? filterEmptyDir(node.children)
          : []
        if (filteredChildren.length === 0) return null
        return { ...node, children: filteredChildren }
      })
      .filter(Boolean)
  }

  const tree = buildTree(filteredNodes)
  return filterEmptyDir(tree)
}

// 以下函数已注释，全局搜索场景下不需要这些逻辑
// /** 筛选空文件夹 */
// const filterEmptyDir = (tree: Knowledge[]): Knowledge[] => {
//   const _filter = (node: Knowledge): Knowledge | null => {
//     if (isAtomNode(node)) return node
//     // 非空的子节点
//     const filteredChildren = node.children?.map(_filter).filter(Boolean)
//     if (!filteredChildren?.length) return null
//     node.children = filteredChildren as Knowledge[]
//     return node
//   }
//   return tree.map((node) => _filter(node)).filter(Boolean) as Knowledge[]
// }

// /** 筛选匹配的节点  */
// const filterTreeBySearch = (
//   tree: Knowledge[],
//   search?: string,
// ): Knowledge[] => {
//   if (!search) {
//     return tree
//   }
//   const getFilteredTitle = (name: string) => {
//     const seqs = name.split(search)
//     const newTitle = (
//       <>
//         {seqs.map((s, i) => {
//           if (i === seqs.length - 1) return s
//           return (
//             <>
//               {s}
//               <span className='bg-[#0C99FF]/20'>{search}</span>
//             </>
//           )
//         })}
//       </>
//     )
//     return newTitle
//   }
//   const _filter = (node: Knowledge): Knowledge | null => {
//     // 本节点匹配
//     if (node.name.includes(search)) {
//       return {
//         ...node,
//         title: getFilteredTitle(node.name),
//       }
//     }
//     const filteredChildren = node.children?.map(_filter).filter(Boolean)
//     if (!filteredChildren?.length) return null
//     // 子节点匹配
//     return {
//       ...node,
//       children: filteredChildren as Knowledge[],
//     }
//   }
//   return tree.map((node) => _filter(node)).filter(Boolean) as Knowledge[]
// }

// /** 当前节点是否是文件、sheet、等不可分节点 */
// const isAtomNode = (node: Knowledge) => {
//   return node.knowledgeType !== 'other'
// }
