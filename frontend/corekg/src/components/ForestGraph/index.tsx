import { useEffect, useRef } from 'react'
import { Circle } from '@antv/g'
import {
  Graph,
  treeToGraphData,
  register,
  CubicHorizontal,
  ExtensionCategory,
  subStyleProps,
  type NodeData,
} from '@antv/g6'
import clsx from 'clsx'

interface ForestGraphProps {
  renderData: any
}

interface NodeType {
  uuid: string
  id: string
  children?: NodeType[]
  [key: string]: any
}

const ForestGraph: React.FC<ForestGraphProps> = ({ renderData }) => {
  const mindMapRef = useRef<HTMLDivElement | null>(null)
  const graphRef = useRef<Graph | null>(null)

  // id去重
  const addUniqueKey = (node: NodeType, parentId = ''): NodeType => {
    const uniqueId = parentId ? `${parentId}-${node.uuid}` : node.uuid
    const newNode = { ...node, id: uniqueId, title: node.id }
    if (node.children && node.children.length) {
      newNode.children = node.children.map((child: NodeType) =>
        addUniqueKey(child, uniqueId),
      )
    }
    return newNode
  }

  // 飞行标记
  class FlyMarkerCubic extends CubicHorizontal {
    getMarkerStyle(attributes: any) {
      return {
        r: 5,
        fill: '#c3d5f9',
        offsetPath: this.shapeMap.key,
        ...subStyleProps(attributes, 'marker'),
      }
    }

    onCreate() {
      const marker = this.upsert(
        'marker',
        Circle,
        this.getMarkerStyle(this.attributes),
        this,
      )
      if (marker) {
        marker.animate([{ offsetDistance: 0 }, { offsetDistance: 1 }], {
          duration: 3000,
          iterations: Infinity,
        })
      }
    }
  }

  useEffect(() => {
    // 注册自定义边类型
    register(ExtensionCategory.EDGE, 'fly-marker-cubic', FlyMarkerCubic)
  }, [])

  useEffect(() => {
    if (mindMapRef.current && renderData && renderData.id) {
      const container = mindMapRef.current
      const Data = addUniqueKey(renderData)

      // 清理之前的图表
      if (graphRef.current) {
        graphRef.current.destroy()
      }

      const graph = new Graph({
        container: container,
        autoFit: 'view',
        data: treeToGraphData(Data),
        zoomRange: [0.2, 3],
        layout: {
          type: 'compact-box',
          direction: 'LR',
          getHeight: () => 32,
          getWidth: () => 32,
          getVGap: () => 10,
          getHGap: () => 100,
          preventOverlap: true,
        },
        node: {
          style: {
            labelText: (data: NodeData) => (data.title as string) || '',
            labelPlacement: 'right',
            labelMaxWidth: 200,
            ports: [{ placement: 'right' }, { placement: 'left' }],
          },
          animation: {
            enter: false,
          },
          state: {
            highlight: {
              fill: '#67C23A',
              halo: true,
              lineWidth: 0,
            },
            dim: {
              fill: '#99ADD1',
            },
          },
        },
        edge: {
          type: 'fly-marker-cubic',
          animation: {
            enter: false,
          },
          state: {
            highlight: {
              stroke: '#67C23A',
            },
          },
        },
        behaviors: [
          'drag-canvas',
          'click-select',
          'drag-element',
          {
            type: 'hover-activate',
            enable: (event: any) => event.targetType === 'node',
            degree: 1,
            state: 'highlight',
            inactiveState: 'dim',
            onHover: (event: any) => {
              event.view.setCursor('pointer')
            },
            onHoverEnd: (event: any) => {
              event.view.setCursor('default')
            },
          },
        ],
        plugins: [
          'grid',
          'contextmenu',
          {
            type: 'tooltip',
            shouldBegin: (e: any) => {
              return e.targetType === 'node'
            },
            getContent: (e: any, items: any[]) => {
              if (!items || !items.length) return ''

              let result = ''
              items.forEach((item: any) => {
                if (item.title) {
                  result += `<p>${item.title}</p>`
                }
              })
              return result
            },
            style: {
              className: 'custom-tooltip',
              background: '#fff',
              boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
              borderRadius: '6px',
              border: '1px solid #e8e8e8',
              padding: '0',
              maxWidth: '250px',
            },
          },
        ],
      } as any)

      graph.render()
      graphRef.current = graph

      // 添加自定义滚轮缩放事件，实现相对中心位置的缩放
      const handleWheel = (event: WheelEvent) => {
        event.preventDefault()

        if (!graph) return

        // 计算缩放因子，使用更小的步长让缩放更平滑
        const delta = event.deltaY
        const scaleFactor = delta > 0 ? 0.9 : 1.1 // 更小的缩放步长

        // 获取当前缩放比例
        const currentZoom = graph.getZoom()
        const newZoom = Math.min(Math.max(currentZoom * scaleFactor, 0.2), 3)

        // 如果缩放比例没有变化，则不执行缩放
        if (Math.abs(newZoom - currentZoom) < 0.001) return

        // 使用G6的zoomTo方法，它会自动处理动画，相对于中心点缩放
        graph.zoomTo(newZoom, {
          duration: 10, // 短动画时间
          easing: 'ease-out',
        })
      }

      // 添加滚轮事件监听
      const mindMapContainer = mindMapRef.current
      if (mindMapContainer) {
        mindMapContainer.addEventListener('wheel', handleWheel, {
          passive: false,
        })
      }

      // 添加重置视图的方法到图表实例上，以便外部调用
      ;(graph as any).resetView = () => {
        graph.fitView()
      }

      return () => {
        if (mindMapContainer) {
          mindMapContainer.removeEventListener('wheel', handleWheel)
        }
        if (graphRef.current) {
          graphRef.current.destroy()
          graphRef.current = null
        }
      }
    }
  }, [renderData])

  useEffect(() => {
    return () => {
      if (graphRef.current) {
        graphRef.current.destroy()
      }
    }
  }, [])

  return (
    <div
      ref={mindMapRef}
      className={clsx('relative h-full w-full')}
      style={{ minHeight: '400px' }}
    ></div>
  )
}

export default ForestGraph
