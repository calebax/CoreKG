import { useRef, useEffect } from 'react'
import type { ReactGridLayoutProps, Layout } from 'react-grid-layout'
import ReactGridLayout from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { throttle } from '@/utils'
import './index.scss'

export { Layout }

export interface IGridLayoutProps extends Omit<ReactGridLayoutProps, 'layout'> {
  layouts: Layout[]
  renderCustomItem?: (item: Layout) => JSX.Element
}
export default function GridLayout(
  props: IGridLayoutProps & {
    defaultWidth?: number
  },
) {
  // 关键：初始值设为一个安全的默认值（非undefined）
  const [gridWidth] = useState<number>(props.defaultWidth ?? 1200)

  const [activeItem, setActiveItem] = useState<string | null>(null)

  const renderItem = (item: Layout) => {
    const className = ['custom-grid-item']
    if (item.i === activeItem) {
      className.push('custom-grid-item-active')
    }
    return (
      <div
        className={className.join(' ')}
        onClick={() => {
          setActiveItem(item.i)
        }}
        key={item.i}
      >
        {props.renderCustomItem ? props.renderCustomItem(item) : item.i}
      </div>
    )
  }
  if (!props.layouts || !Array.isArray(props.layouts)) {
    return null
  }

  return (
    <ReactGridLayout
      width={gridWidth}
      cols={6}
      rowHeight={100}
      isDraggable={true}
      isResizable={true}
      preventCollision={false}
      onLayoutChange={props.onLayoutChange}
      margin={[10, 10]}
      draggableHandle='.drag-handle' // 指定拖拽手柄
      style={{ width: '100%' }}
      {...props}
      layout={props.layouts}
    >
      {props.layouts!.map(renderItem)}
    </ReactGridLayout>
  )
}
