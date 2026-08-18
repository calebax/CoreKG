import {
  useEffect,
  useRef,
  useState,
  forwardRef,
  useImperativeHandle,
} from 'react'
import { Spin } from 'antd'
import { Graph as G6Graph } from '@antv/g6'
import { getKnowledgeBaseGraph, getNodesByID } from '@/api/knowledge'
import ControlPanel from './components/ControlPanel'
import SearchPanel from './components/SearchPanel'
import {
  Node,
  Edge,
  GraphData,
  KnowledgeGraphProps,
  KnowledgeGraphRef,
  calculateNodeDegrees,
  determineNodeClusters,
  applyInitialPositioning,
} from './graphData'
import {
  getClusterColors,
  groupNodesByCluster,
  createHullPlugins,
  getGraphOptions,
  getLayoutStages,
} from './graphOptions'

const KnowledgeGraphComponent = forwardRef<
  KnowledgeGraphRef,
  KnowledgeGraphProps
>(({ knowledgeBaseId, selectedNodeId }, ref) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const graphRef = useRef<G6Graph | null>(null)
  const [graphData, setGraphData] = useState<GraphData>({
    nodes: [],
    edges: [],
  })
  const [processedNodes, setProcessedNodes] = useState<Node[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [highlightedNodeId, setHighlightedNodeId] = useState<string | null>(
    null,
  )
  const [showLegend, setShowLegend] = useState<boolean>(false)
  const [clusterNames, setClusterNames] = useState<string[]>([])
  const [activeClusterFilters, setActiveClusterFilters] = useState<string[]>([])

  useImperativeHandle(ref, () => ({
    resetView: () => {
      if (graphRef.current) {
        // 重置视图：先重置缩放到1.1，然后适应视图并居中
        graphRef.current.zoomTo(1.1, {
          duration: 300,
          easing: 'ease-in-out',
        })

        // 延迟执行fitView和fitCenter，确保缩放完成后再调整位置
        setTimeout(() => {
          if (graphRef.current) {
            graphRef.current.fitView(40)
            graphRef.current.fitCenter()
          }
        }, 350)
      }
    },
  }))

  useEffect(() => {
    const fetchGraphData = async () => {
      if (!knowledgeBaseId) return

      setIsLoading(true)
      try {
        let response
        if (selectedNodeId) {
          // 如果有选中的节点ID，调用getNodesByID接口
          console.log('获取特定节点的图谱数据:', selectedNodeId)
          response = await getNodesByID({
            knowledgeBaseId: Number(knowledgeBaseId),
            nodeId: selectedNodeId,
          })
          console.log('节点图谱 API 响应:', response)
        } else {
          // 否则获取完整的知识图谱
          response = await getKnowledgeBaseGraph(Number(knowledgeBaseId))
          console.log('完整知识图谱 API 响应:', response)
        }

        if (
          response.graph &&
          response.graph.nodes &&
          response.graph.edges &&
          Array.isArray(response.graph.nodes) &&
          Array.isArray(response.graph.edges)
        ) {
          setGraphData({
            nodes: response.graph.nodes.map((node: any) => ({
              id: node.id || node.ID,
              label: node.id,
              ...node,
            })),
            edges: response.graph.edges.map((edge: any, index: number) => ({
              source: edge.source || edge.from,
              target: edge.target || edge.to,
              id:
                edge.id ||
                `edge-${edge.source || edge.from}-${edge.target || edge.to}-${index}-${Date.now()}`,
              ...edge,
            })),
          })
        } else {
          console.log('接口返回数据为空，使用模拟数据')
        }
      } catch (error) {
        console.error('获取知识图谱数据失败:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchGraphData()
  }, [knowledgeBaseId, selectedNodeId])

  // TODO:
  // 处理节点搜索

  // 处理集群过滤
  const handleClusterFilterChange = (cluster: string) => {
    if (!graphRef.current) return

    let newFilters: string[]

    if (activeClusterFilters.includes(cluster)) {
      // 如果已经激活，则删除
      newFilters = activeClusterFilters.filter((c) => c !== cluster)
    } else {
      // 否则添加到激活列表
      newFilters = [...activeClusterFilters, cluster]
    }

    setActiveClusterFilters(newFilters)

    // 应用过滤
    if (newFilters.length === 0) {
      // 如果没有过滤器，显示所有节点
      processedNodes.forEach((node) => {
        graphRef.current!.setElementState(node.id, 'active')
      })
    } else {
      // 否则只显示选中集群的节点
      processedNodes.forEach((node) => {
        const nodeCluster = node.cluster || 'default'
        if (newFilters.includes(nodeCluster)) {
          graphRef.current!.setElementState(node.id, 'selected')
        } else {
          graphRef.current!.setElementState(node.id, 'inactive')
        }
      })
    }
  }

  useEffect(() => {
    if (!graphData.nodes.length && !graphData.edges.length) return

    // 添加集群信息到节点
    const nodesWithClusters = determineNodeClusters(
      graphData.nodes,
      graphData.edges,
    )

    // 计算节点度数
    const nodesWithDegreesAndClusters = calculateNodeDegrees(
      nodesWithClusters,
      graphData.edges,
    )

    // 更新处理后的节点
    setProcessedNodes(nodesWithDegreesAndClusters)

    // 提取集群名称用于图例
    const clusters = new Set<string>()
    nodesWithDegreesAndClusters.forEach((node) => {
      if (node.cluster) {
        clusters.add(node.cluster)
      }
    })
    setClusterNames(Array.from(clusters))

    // 按集群分组节点
    const groupedNodesByCluster = groupNodesByCluster(
      nodesWithDegreesAndClusters,
    )

    // 获取集群颜色映射
    const clusterColors = getClusterColors(groupedNodesByCluster)

    // 创建Hull插件配置
    const hullPlugins = createHullPlugins(groupedNodesByCluster, clusterColors)

    // 初始化节点位置
    const positionedNodes = applyInitialPositioning(
      nodesWithDegreesAndClusters,
      containerRef.current?.clientWidth || 0,
      containerRef.current?.clientHeight || 0,
    )

    // 获取图形配置
    const graphOptions = getGraphOptions(
      containerRef.current?.clientWidth || 0,
      containerRef.current?.clientHeight || 0,
      positionedNodes,
      graphData.edges,
      clusterColors,
      hullPlugins,
    )

    // 设置容器
    graphOptions.container = containerRef.current!

    const graph = new G6Graph(graphOptions)

    // 添加流畅的滚轮缩放事件
    const handleWheel = (event: WheelEvent) => {
      event.preventDefault()

      if (!graph) return

      // 计算缩放因子，使用更小的步长让缩放更平滑
      const delta = event.deltaY
      const scaleFactor = delta > 0 ? 0.95 : 1.05 // 更小的缩放步长，更平滑

      // 获取当前缩放比例
      const currentZoom = graph.getZoom()
      const newZoom = Math.min(Math.max(currentZoom * scaleFactor, 0.2), 5)

      // 如果缩放比例没有变化，则不执行缩放
      if (Math.abs(newZoom - currentZoom) < 0.001) return

      // 获取鼠标相对于画布的位置
      const point = {
        x: event.offsetX,
        y: event.offsetY,
      }

      // 以鼠标位置为中心进行缩放
      graph.zoomTo(newZoom, {
        x: point.x,
        y: point.y,
        duration: 100,
        easing: 'ease-out',
      })
    }

    // 添加滚轮事件监听
    const container = containerRef.current
    if (container) {
      container.addEventListener('wheel', handleWheel, { passive: false })
    }

    graphRef.current = graph
    graph.render()

    // 获取分阶段布局配置
    const layoutStages = getLayoutStages()

    // 分阶段布局优化
    setTimeout(() => {
      // 第一阶段
      graph.setLayout(layoutStages[0])

      // 第二阶段：平衡布局
      setTimeout(() => {
        graph.setLayout(layoutStages[1])

        setTimeout(() => {
          graph.fitView(40)
          graph.fitCenter()
        }, 350)
      }, 350)
    }, 100)

    return () => {
      if (container) {
        container.removeEventListener('wheel', handleWheel)
      }
      graph.destroy()
    }
  }, [graphData])

  if (isLoading) {
    return (
      <div className='flex h-full w-full items-center justify-center'>
        <Spin size='large' />
      </div>
    )
  }

  return (
    <div className='relative w-full h-full'>
      <div ref={containerRef} className='w-full h-full' />

      {/* 搜索面板 */}
      {/* <SearchPanel onSearch={handleSearch} nodes={processedNodes} /> */}

      {/* 控制面板 */}
      <ControlPanel
        onFitView={() => {
          if (graphRef.current) {
            graphRef.current.fitView(50)
            graphRef.current.fitCenter()
          }
        }}
        onClearHighlight={() => {
          setHighlightedNodeId(null)
        }}
        highlightedNodeId={highlightedNodeId}
        onToggleLegend={() => setShowLegend(!showLegend)}
        showLegend={showLegend}
        clusterNames={clusterNames}
        onClusterFilterChange={handleClusterFilterChange}
        activeClusterFilters={activeClusterFilters}
      />
    </div>
  )
})

export default KnowledgeGraphComponent
