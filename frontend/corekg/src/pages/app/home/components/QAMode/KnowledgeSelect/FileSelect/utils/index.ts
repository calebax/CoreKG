import { FileItem } from '../../../../KnowledgeContext'

export type TreeNode = {
  id: number
  key: string
  name: string
  title: string
  parent_id: number | null
  children: TreeNode[]
  checkable: boolean
  is_dir: boolean
}
/**
 * 将节点数组转化树结构
 * @param list 节点数组
 * @param filter 筛选字符串
 */
export const arrayToTree = (list: FileItem[], filter?: string) => {
  const treeMap = new Map<string, TreeNode>()
  list.forEach((item) => {
    const { id, name, parent_id, is_dir } = item
    /** 知识库和文件id可能相同 用不同的前缀做区分 */
    const treeId = is_dir ? `dir-${id}` : `file-${id}`
    treeMap.set(treeId, {
      parent_id,
      id,
      name,
      key: treeId,
      title: name,
      children: [],
      checkable: false,
      is_dir,
    })
  })
  ;[...treeMap.values()].forEach((item) => {
    const { parent_id } = item
    if (!parent_id) return
    const parent_node = treeMap.get(`dir-${parent_id}`)
    if (!parent_node) return
    parent_node.children.push(item)
  })
  const tree = [...treeMap.values()].filter((item) => !item.parent_id)
  const filteredTree = filterTree(tree, filter)
  /**
   * 统一设置各节点的checkable属性
   * @return 本节点或子元素是否包含有效文件
   */
  const setCheckable = (node: TreeNode) => {
    if (isFileNode(node)) {
      node.checkable = true
      return true
    }
    const hasFile = node.children.map(setCheckable).some(Boolean) as boolean
    node.checkable = hasFile
    return hasFile
  }
  filteredTree.forEach(setCheckable)
  return filteredTree
}
/** 按名称筛选树结构 */
const filterTree = (tree: TreeNode[], search?: string) => {
  if (!search) return tree
  const _filter = (node: TreeNode): TreeNode | null => {
    if (node.name.includes(search)) return node
    const filteredChildren = node.children
      .map((child) => _filter(child))
      .filter(Boolean)
    if (filteredChildren.length === 0) return null
    return { ...node, children: filteredChildren as TreeNode[] }
  }
  return tree.map((node) => _filter(node)).filter(Boolean) as TreeNode[]
}
/** 当前节点是否是文件节点 */
export const isFileNode = (node: Pick<TreeNode, 'parent_id' | 'is_dir'>) => {
  return Boolean(!node.is_dir && node.parent_id)
}
