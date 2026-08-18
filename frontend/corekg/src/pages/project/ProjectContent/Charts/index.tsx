import { FC } from 'react'
import { Empty } from 'antd'
import EChartsReact from 'echarts-for-react'
import { ErrorBoundary } from 'react-error-boundary'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import NoDataIcon from '@/assets/icons/project/empty.svg?react'
import GridLayout from '@/components/grid-layout'
import { useProject } from '../..'
import DeleteIcon from './images/delete.svg?react'
import DragMoveIcon from './images/drag.svg?react'

export const Charts: FC<Style> = (props) => {
  const { className, style } = props
  const { t } = useTranslation('pages')
  const {
    data: { charts },
  } = useProject()
  if (charts.length === 0) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-[30px]'>
        <div>
          <NoDataIcon />
        </div>
        <div className='text-[#616373] font-[400] text-[14px] leading-[1]'>
          暂无相关图表～
        </div>
      </div>
    )
  }
  return (
    <div
      className={cn(
        'w-[1200px] min-w-[1200px] min-h-full flex flex-col',
        className,
      )}
      style={style}
    >
      {/* <div className='w-full h-[150px] flex flex-col items-center bg-white text-center'>
        <span className='text-lg font-bold mt-5'>
          {t('project.graph.title')}
        </span>
        <div className='text-xs text-[#919497] mt-2.5 w-[555px]'>
          {t('project.graph.desc')}
        </div>
      </div> */}
      <div className='w-full flex-1'>
        <ChartsInner />
      </div>
    </div>
  )
}

const ChartsInner: FC = () => {
  const {
    chartsOperators,
    data: { charts },
  } = useProject()
  const layouts = useMemo(() => {
    return charts.map((item) => item.layout)
  }, [charts])
  return (
    <GridLayout
      defaultWidth={1200}
      layouts={layouts}
      renderCustomItem={(val) => {
        const i = val.i
        const { option, id } = charts.find((item) => item.layout.i === i) ?? {}
        return (
          <ErrorBoundary fallback={null}>
            <div className='w-full h-full bg-white rounded p-2 flex flex-col'>
              <span className='flex  justify-between items-center gap-2'>
                <div
                  className='text-[12px] 
                font-medium text-[#3c4149] 
                whitespace-nowrap overflow-hidden text-ellipsis'
                >
                  {option?.title?.text}
                </div>
                {/* 抓取用的class */}
                <div className='flex items-center gap-2'>
                  <DragMoveIcon className='drag-handle cursor-grab' />
                  <DeleteIcon
                    className='cursor-pointer'
                    onClick={() => chartsOperators.del(id!)}
                  />
                </div>
              </span>
              <EChartsReact className='flex-1' option={option} />
            </div>
          </ErrorBoundary>
        )
      }}
      onLayoutChange={chartsOperators.setLayouts}
    />
  )
}
