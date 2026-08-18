import { useRef } from 'react'
import { useMemoizedFn } from 'ahooks'

/** 各节点的原始数据 */
const useOriginNodes = (defaultNodes: any[]) => {
  const originNodes = useRef<Map<string, any>>(new Map())
  const addNodes = useMemoizedFn((nodes: any[]) => {
    nodes.forEach((n) => {
      const { name } = n
      if (originNodes.current.has(name)) {
        return
      }
      originNodes.current.set(name, n)
    })
  })

  /** 覆盖写入节点（用于创建/更新后同步缓存） */
  const upsertNode = useMemoizedFn((node: any) => {
    if (!node?.name) return
    originNodes.current.set(node.name, node)
  })

  /** 覆盖写入多个节点 */
  const upsertNodes = useMemoizedFn((nodes: any[]) => {
    nodes.forEach((n) => {
      if (!n?.name) return
      originNodes.current.set(n.name, n)
    })
  })

  /** 重命名节点（用于编辑时修改名称） */
  const renameNode = useMemoizedFn(
    (oldName: string, newName: string, node?: any) => {
      if (!oldName || !newName) return
      const current = node ?? originNodes.current.get(oldName)
      if (!current) return
      originNodes.current.delete(oldName)
      originNodes.current.set(newName, { ...current, name: newName })
    },
  )

  const getNode = useMemoizedFn((name: string) => {
    return originNodes.current.get(name)
  })

  const isFirst = useRef(true)
  if (isFirst.current) {
    addNodes(defaultNodes)
    isFirst.current = false
  }
  return { addNodes, upsertNode, upsertNodes, renameNode, getNode }
}

export default useOriginNodes
