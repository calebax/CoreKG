import { CSSProperties } from 'react'
import { Graph, GraphOptions } from '@antv/g6'

export type ReactG6Props = {
  className?: string
  style?: CSSProperties
  data?: GraphOptions['data']
  layout?: GraphOptions['layout']
  options?: Omit<GraphOptions, 'data' | 'layout'>
}
export const ReactG6 = forwardRef<Graph | undefined, ReactG6Props>(
  (props, ref) => {
    const { className, style, data, layout, options } = props

    const containerRef = useRef<HTMLDivElement>(null)
    const [graph, setGrgph] = useState<Graph>()

    useImperativeHandle(ref, () => graph, [graph])

    useEffect(() => {
      const container = containerRef.current
      if (!container) return

      const newGraph = new Graph({ container })
      setGrgph(newGraph)

      const onResize = () => {
        newGraph.resize()
      }
      container.addEventListener('resize', onResize)

      const onWheel = (e: WheelEvent) => {
        e.preventDefault()
        scaleGraph(e.deltaY, newGraph)
      }
      container.addEventListener('wheel', onWheel, { passive: false })

      return () => {
        newGraph.destroy()
        container.removeEventListener('resize', onResize)
        container.removeEventListener('wheel', onWheel)
      }
    }, [])

    // 注意顺序 setOptions render layout
    useEffect(() => {
      if (!graph || !options) return
      // setOptions会自动重绘
      graph.setOptions({ ...options, data: undefined, layout: undefined })
    }, [graph, options])
    useEffect(() => {
      if (!graph || !data) return
      // setData后需要手动计算和重绘
      graph.setData(data)
      graph.render()
    }, [data, graph])
    useEffect(() => {
      if (!graph || !layout) return
      // 设置layout后 重新计算布局 但需要等待上一次render完成
      graph.setLayout(layout)
      if (graph.rendered) {
        graph.layout()
      }
    }, [graph, layout])

    return <div className={className} style={style} ref={containerRef}></div>
  },
)

const scaleGraph = (delta: number, graph: Graph) => {
  const scaleFactor = delta > 0 ? 0.9 : 1.1 // 更小的缩放步长
  const currentZoom = graph.getZoom()
  const newZoom = Math.min(Math.max(currentZoom * scaleFactor, 0.3), 3)
  if (Math.abs(newZoom - currentZoom) < 0.001) return
  graph.zoomTo(newZoom, {
    duration: 10,
    easing: 'ease-out',
  })
}
