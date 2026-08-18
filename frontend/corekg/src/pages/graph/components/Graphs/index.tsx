import { FC, PropsWithChildren } from 'react'
import { Empty } from 'antd'
import { GraphBaseInfo } from 'Graph'
import { cn } from '@/utils'
import { GraphCard } from '../GraphCard'
import EmptyIcon from './images/empty.svg?react'

export type Graphs = PropsWithChildren &
  Style & {
    data: GraphBaseInfo[]
    reload: () => void
  }
export const Graphs: FC<Graphs> = (props) => {
  const { data, reload, children, className, style = {} } = props
  if (data.length === 0) {
    return (
      <>
        {children}
        <div
          className='w-full h-60 flex items-center flex-col justify-center text-[#919497]'
          style={{ gridColumn: '1/-1' }}
        >
          <EmptyIcon />
          暂无知识图谱数据
        </div>
      </>
    )
  }
  return (
    <>
      {children}
      {data.map((item) => (
        <GraphCard key={item.id} reload={reload} value={item} />
      ))}
    </>
  )
}
