import { FC } from 'react'
import { Button } from 'antd'
import GridLayout, { type Layout } from './index'

export default function Demo() {
  const [layouts, setLayouts] = useState<Layout[]>([
    { i: '1', x: 0, y: 0, w: 1, h: 2 }, // 第一列
    { i: '2', x: 1, y: 0, w: 1, h: 2 }, // 第二列
    { i: '3', x: 2, y: 0, w: 1, h: 2 }, // 第三列
    { i: '4', x: 3, y: 0, w: 1, h: 2 }, // 第四列
    { i: '5', x: 0, y: 2, w: 1, h: 2 }, // 第二行第一列（自动换行）
  ])
  //网格数据变化
  const handleLayoutChange = (newLayouts: Layout[]) => {
    console.log(newLayouts)
    // 不需要setLayouts也可以正常使用 但是添加新图的时候会回退到最初的layout
    // setLayouts(
    //   newLayouts.map((item) => ({
    //     x: item.x,
    //     y: item.y,
    //     w: item.w,
    //     h: item.h,
    //     i: item.i,
    //   })),
    // )
  }
  // 自定义网格内容
  const renderCard = (item: any) => {
    return (
      <div className='card' style={{ width: '100%', height: '100%' }}>
        {/* 类名 drag-handle,是被用来拖拽卡片的手柄,不给就不能拖拽*/}
        <div className='drag-handle'>header</div>
        {/* i是key添加元素至layout不会引起组件重载 */}
        <TestComp />
        <div
          className='card-content'
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100%',
          }}
        >
          图表{item.i}
        </div>
      </div>
    )
  }

  return (
    <>
      {/* 这个组件不是完全受控的 但通过这种方式可以添加元素到左上角 */}
      <Button
        onClick={() => {
          setLayouts((layouts) => {
            layouts.unshift({
              i: '' + layouts.length + 2,
              x: 0,
              y: 0,
              w: 1,
              h: 1,
            })
            return [...layouts]
          })
        }}
      ></Button>
      <GridLayout
        layouts={layouts}
        renderCustomItem={renderCard}
        onLayoutChange={handleLayoutChange}
      />
    </>
  )
}

const TestComp: FC = (props) => {
  useEffect(() => {
    console.log(1)
  }, [])
  return null
}
