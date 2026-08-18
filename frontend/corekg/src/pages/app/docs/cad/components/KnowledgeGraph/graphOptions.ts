import { Node, Edge } from './graphData'

// 从字符串生成颜色
export const generateColorFromString = (str: string) => {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }

  // 只变化色相，以生成不同但视觉上平衡的颜色
  const h = ((hash % 360) + 360) % 360 // 0-359范围的色相
  const s = 65 + (hash % 20) // 饱和度在65%-85%之间
  const l = 55 + (hash % 15) // 亮度在55%-70%之间

  return `hsl(${h}, ${s}%, ${l}%)`
}

// 预定义的集群颜色
export const predefinedClusterColors: Record<string, string> = {
  'cluster-a': '#1783FF', // Blue
  'cluster-b': '#00C9C9', // Teal
  'cluster-c': '#F08F56', // Orange
  'cluster-d': '#D580FF', // Purple
  'cluster-e': '#4CAF50', // Green
  'cluster-f': '#FF5252', // Red
  'cluster-g': '#FFC107', // Amber
  'cluster-h': '#607D8B', // Blue Grey
}

// 创建Hull样式
export const createHullStyle = (baseColor: string) => ({
  fill: baseColor,
  fillOpacity: 0,
  stroke: baseColor,
  strokeOpacity: 0,
  strokeWidth: 0,
  padding: 35,
  // labelFill: '#333',
  // labelPadding: 5,
  // labelBackgroundFill: baseColor,
  // labelBackgroundOpacity: 0.7,
  // labelBackgroundRadius: 4,
})

// 获取集群颜色映射
export const getClusterColors = (
  groupedNodesByCluster: Record<string, string[]>,
): Record<string, string> => {
  // 存储所有集群的颜色映射
  const clusterColors: Record<string, string> = {}

  // 处理所有集群，为未预定义的集群生成颜色
  Object.keys(groupedNodesByCluster).forEach((clusterKey) => {
    if (predefinedClusterColors[clusterKey]) {
      clusterColors[clusterKey] = predefinedClusterColors[clusterKey]
    } else {
      // 为新的集群生成一个基于其名称的唯一颜色
      clusterColors[clusterKey] = generateColorFromString(clusterKey)
    }
  })

  return clusterColors
}

// 创建Hull插件配置
export const createHullPlugins = (
  groupedNodesByCluster: Record<string, string[]>,
  clusterColors: Record<string, string>,
) => {
  return Object.entries(groupedNodesByCluster).map(([clusterKey, members]) => ({
    key: `hull-${clusterKey}`,
    type: 'hull',
    members,
    // labelText: clusterKey,
    ...createHullStyle(clusterColors[clusterKey] || '#B4B4B4'),
  }))
}

// 按集群分组节点
export const groupNodesByCluster = (
  nodes: Node[],
): Record<string, string[]> => {
  return nodes.reduce(
    (acc, node) => {
      const cluster = node.cluster || 'default'
      acc[cluster] = acc[cluster] || []
      acc[cluster].push(node.id)
      return acc
    },
    {} as Record<string, string[]>,
  )
}

// 获取图形配置选项
export const getGraphOptions = (
  containerWidth: number,
  containerHeight: number,
  nodes: Node[],
  edges: Edge[],
  clusterColors: Record<string, string>,
  hullPlugins: any[],
) => {
  return {
    container: null, // 在使用时设置
    autoFit: 'view',
    padding: 50,
    data: {
      nodes,
      edges,
    },
    zoomRange: [0.2, 5],
    // 使用behaviors配置交互
    behaviors: [
      'drag-canvas',
      'zoom-canvas',
      'drag-node',
      'drag-element',
      {
        type: 'hover-activate',
        degree: 1,
        state: 'highlight',
        // inactiveState: 'dim',
      },
    ],
    layout: {
      type: 'force-atlas2',
      preventOverlap: true,
      center: [
        containerWidth ? containerWidth / 2 : 400,
        containerHeight ? containerHeight / 2 : 300,
      ],
      width: containerWidth || 800,
      height: containerHeight || 600,
      kr: 25,
      kg: 10,
      ks: 0.3,
      preventOverlapPadding: 50,
      maxIteration: 1500, // 增加迭代次数，布局更稳定
      getCenter: () => {
        return [containerWidth / 2, containerHeight / 2]
      },
      workerEnabled: false,
    },
    node: {
      style: {
        // 使用预先计算的度数来设置节点大小
        size: (node: any) => {
          return Math.min(45 + (node.degree || 0) * 2, 80)
        },
        stroke: '#fff',
        lineWidth: 2,
        fill: (node: any) => {
          const cluster = node.cluster || 'default'
          return clusterColors[cluster] || '#B4B4B4'
        },
        labelText: (node: any) => node.label || node.id,
        labelFontSize: 14,
        labelFontWeight: 500,
        labelPlacement: 'bottom',
        labelOffset: 12,
        labelBackground: true,
        labelBackgroundFill: '#ffffff',
        labelBackgroundOpacity: 0.7,
        labelBackgroundPadding: [2, 4, 2, 4],
      },
      state: {
        highlight: {
          lineWidth: 3,
          shadowColor: 'rgba(0,0,0,0.3)',
          shadowBlur: 10,
        },
        dim: {
          opacity: 0.2,
        },
      },
    },
    edge: {
      style: {
        stroke: '#ddd',
        lineWidth: 1.5,
        opacity: 0.7,
        endArrow: true,
        // 为不同集群间的边设置不同样式
        strokeOpacity: (edge: any) => {
          return 0.7
        },
      },
      state: {
        highlight: {
          stroke: '#1890ff',
          lineWidth: 2.5,
          opacity: 1,
        },
        dim: {
          opacity: 0.1,
        },
      },
    },
    plugins: [
      {
        type: 'tooltip',
        getContent: (e: any, items: any[]) => {
          return `
            <div style="color: rgba(0,0,0,0.8); background: white; border-radius: 4px; font-size: 12px;">
              <h4>节点详情</h4>
              ${items
                .map(
                  (item: any) =>
                    `<p style="margin: 0; line-height: 1.4;">
                  <strong>${item.id}</strong>: ${item?.tag?.[0]?.description || '无描述'}
                </p>`,
                )
                .join('')}
            </div>
          `
        },
      },
      ...hullPlugins, // 添加hull插件
    ],
  }
}

// 获取分阶段布局配置
export const getLayoutStages = () => {
  return [
    {
      // 第一阶段
      type: 'force-atlas2',
      preventOverlap: true,
      kr: 15, // 较低斥力
      kg: 15, // 较高重力
      ks: 0.4, // 较强边引力
      preventOverlapPadding: 50,
      maxIteration: 300,
    },
    {
      // 第二阶段：平衡布局
      type: 'force-atlas2',
      preventOverlap: true,
      kr: 25,
      kg: 10,
      ks: 0.3,
      preventOverlapPadding: 50,
      maxIteration: 300,
    },
  ]
}
