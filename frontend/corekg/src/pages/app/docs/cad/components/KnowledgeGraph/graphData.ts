import { Graph as G6Graph } from '@antv/g6'

// 类型定义
export interface Node {
  id: string
  label?: string
  cluster?: string | null
  tag?: Array<{ description?: string; cluster?: string }>
  degree?: number
  [key: string]: any
}

export interface Edge {
  source: string
  target: string
  [key: string]: any
}

export interface GraphData {
  nodes: Node[]
  edges: Edge[]
}

export interface KnowledgeGraphRef {
  resetView: () => void
}

export interface KnowledgeGraphProps {
  knowledgeBaseId?: string
  selectedNodeId?: string // 选中的节点ID，用于显示特定节点的图谱
}

// 计算节点度数（连接数量）
export const calculateNodeDegrees = (nodes: Node[], edges: Edge[]): Node[] => {
  // 创建度数映射
  const degreesMap: Record<string, number> = {}

  // 初始化所有节点的度数为0
  nodes.forEach((node) => {
    degreesMap[node.id] = 0
  })

  // 计算每个节点的连接数
  edges.forEach((edge) => {
    degreesMap[edge.source] = (degreesMap[edge.source] || 0) + 1
    degreesMap[edge.target] = (degreesMap[edge.target] || 0) + 1
  })

  // 将度数信息添加到节点
  return nodes.map((node) => ({
    ...node,
    degree: degreesMap[node.id] || 0,
  }))
}

// 确定节点集群
export const determineNodeClusters = (nodes: Node[], edges: Edge[]): Node[] => {
  // 首先检查节点是否已经通过tag中的cluster字段定义了集群
  const nodesWithClusters = nodes.map((node) => {
    // 尝试从tag数组中获取cluster信息
    const clusterFromTag = node.tag?.find((tag) => tag.cluster)?.cluster

    if (clusterFromTag) {
      return {
        ...node,
        cluster: clusterFromTag,
      }
    }

    return {
      ...node,
      cluster: node.cluster || null,
    }
  })

  // 对于没有从tag中获取到cluster的节点，使用原有的算法分配集群
  const remainingNodes = nodesWithClusters.filter(
    (node) => node.cluster === null,
  )
  const nodeIdsWithoutCluster = remainingNodes.map((node) => node.id)

  // 原有的集群分配逻辑，仅用于没有通过tag定义集群的节点
  if (nodeIdsWithoutCluster.length > 0) {
    // 初始化所有未分配节点为null集群
    const nodeClusters: Record<string, string | null> =
      nodeIdsWithoutCluster.reduce(
        (acc, id) => {
          acc[id] = null
          return acc
        },
        {} as Record<string, string | null>,
      )

    // 简单的集群检测逻辑
    let clusterCounter = 0

    // 查找根节点（出连接多于入连接的节点）
    const incomingConnections: Record<string, number> = {}
    const outgoingConnections: Record<string, number> = {}

    edges.forEach((edge) => {
      incomingConnections[edge.target] =
        (incomingConnections[edge.target] || 0) + 1
      outgoingConnections[edge.source] =
        (outgoingConnections[edge.source] || 0) + 1
    })

    // 从nodeIdsWithoutCluster中识别潜在的根节点
    const rootNodes = remainingNodes.filter(
      (node) =>
        (outgoingConnections[node.id] || 0) >
        (incomingConnections[node.id] || 0),
    )

    // 如果没有明确的根节点，则使用第一个节点作为起点
    const startNodes =
      rootNodes.length > 0
        ? rootNodes
        : remainingNodes.length > 0
          ? [remainingNodes[0]]
          : []

    // 从每个根节点开始分配集群
    startNodes.forEach((startNode) => {
      const clusterName = `cluster-${String.fromCharCode(97 + clusterCounter)}` // a, b, c, ...
      clusterCounter++

      // 为此节点及其直接邻居分配集群
      nodeClusters[startNode.id] = clusterName

      // 查找直接连接
      edges.forEach((edge) => {
        if (
          edge.source === startNode.id &&
          nodeIdsWithoutCluster.includes(edge.target) &&
          !nodeClusters[edge.target]
        ) {
          nodeClusters[edge.target] = clusterName
        }
      })
    })

    // 将剩余的节点分配到它们自己的集群或最近的连接集群
    nodeIdsWithoutCluster.forEach((nodeId) => {
      if (!nodeClusters[nodeId]) {
        // 查找与此节点连接的任何边
        const connectedEdge = edges.find(
          (edge) => edge.source === nodeId || edge.target === nodeId,
        )

        if (connectedEdge) {
          const connectedNodeId =
            connectedEdge.source === nodeId
              ? connectedEdge.target
              : connectedEdge.source

          if (nodeClusters[connectedNodeId]) {
            nodeClusters[nodeId] = nodeClusters[connectedNodeId]
          } else {
            nodeClusters[nodeId] =
              `cluster-${String.fromCharCode(97 + clusterCounter)}`
            clusterCounter++
          }
        } else {
          // 孤立节点获得自己的集群
          nodeClusters[nodeId] =
            `cluster-${String.fromCharCode(97 + clusterCounter)}`
          clusterCounter++
        }
      }
    })

    // 将分配的集群应用到未分配的节点
    return nodesWithClusters.map((node) => {
      if (nodeIdsWithoutCluster.includes(node.id)) {
        return {
          ...node,
          cluster: nodeClusters[node.id] as string,
        }
      }
      return node
    })
  }

  return nodesWithClusters
}

// 初始化节点位置
export const applyInitialPositioning = (
  nodes: Node[],
  containerWidth: number,
  containerHeight: number,
) => {
  const width = containerWidth || 800
  const height = containerHeight || 600
  const centerX = width / 2
  const centerY = height / 2

  // 按集群分组节点
  const clusterGroups: Record<string, Node[]> = {}
  nodes.forEach((node) => {
    const cluster = node.cluster || 'default'
    if (!clusterGroups[cluster]) {
      clusterGroups[cluster] = []
    }
    clusterGroups[cluster].push(node)
  })

  // 为每个集群创建圆形布局
  const numClusters = Object.keys(clusterGroups).length
  const maxRadius = Math.min(width, height) * 0.4

  return nodes.map((node) => {
    const cluster = node.cluster || 'default'
    const clusterIndex = Object.keys(clusterGroups).indexOf(cluster)
    const clusterSize = clusterGroups[cluster].length
    const nodeIndexInCluster = clusterGroups[cluster].indexOf(node)

    // 计算集群位置
    const clusterAngle = (clusterIndex / numClusters) * 2 * Math.PI
    const clusterX = centerX + maxRadius * 0.8 * Math.cos(clusterAngle)
    const clusterY = centerY + maxRadius * 0.8 * Math.sin(clusterAngle)

    // 在集群内部分布节点
    const nodeAngle = (nodeIndexInCluster / clusterSize) * 2 * Math.PI
    const clusterRadius = 30 + Math.sqrt(clusterSize) * 10

    return {
      ...node,
      x: clusterX + clusterRadius * Math.cos(nodeAngle),
      y: clusterY + clusterRadius * Math.sin(nodeAngle),
    }
  })
}
