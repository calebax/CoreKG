import { Tooltip as AntdTooltip } from 'antd'
import { FilterOutlined, InfoCircleOutlined } from '@ant-design/icons'

const ControlPanel = ({
  onFitView,
  onClearHighlight,
  highlightedNodeId,
  onToggleLegend,
  showLegend,
  clusterNames,
  onClusterFilterChange,
  activeClusterFilters,
}: {
  onFitView: () => void
  onClearHighlight: () => void
  highlightedNodeId: string | null
  onToggleLegend: () => void
  showLegend: boolean
  clusterNames: string[]
  onClusterFilterChange: (cluster: string) => void
  activeClusterFilters: string[]
}) => {
  return (
    <div className='absolute top-0 right-28 z-10 flex flex-col gap-2'>
      <div className='flex justify-end gap-2'>
        {highlightedNodeId && (
          <button
            onClick={onClearHighlight}
            className='bg-white hover:bg-gray-100 text-gray-800 font-semibold py-1 px-3 border border-gray-200 rounded shadow text-sm'
          >
            清除高亮
          </button>
        )}
        <button
          onClick={onToggleLegend}
          className={`${
            showLegend ? 'bg-blue-100' : 'bg-white'
          } hover:bg-gray-100 text-gray-800 font-semibold py-1 px-3 border border-gray-200 rounded shadow text-sm flex items-center`}
        >
          <InfoCircleOutlined style={{ marginRight: 4 }} />
          图例
        </button>
      </div>

      {showLegend && (
        <div className='bg-white p-3 border border-gray-200 rounded shadow-md mt-1 w-[250px]'>
          <div className='text-sm font-bold mb-2 h-6 flex items-center'>
            <FilterOutlined className='flex-shrink-0 mr-1' />
            <span className='flex-grow'>图例过滤</span>
            <AntdTooltip title='点击图例切换显示/隐藏'>
              <InfoCircleOutlined className='flex-shrink-0 text-[12px] text-gray-400' />
            </AntdTooltip>
          </div>
          <div className='flex flex-wrap gap-2 max-h-[300px] overflow-y-auto pr-1'>
            {clusterNames.map((cluster) => {
              const isActive = activeClusterFilters.includes(cluster)
              return (
                <AntdTooltip
                  key={cluster}
                  title={cluster}
                  placement='top'
                  mouseEnterDelay={0.5}
                >
                  <button
                    onClick={() => onClusterFilterChange(cluster)}
                    className={`text-xs py-1 px-2 rounded-full border max-w-[160px] whitespace-nowrap overflow-hidden text-ellipsis ${
                      isActive
                        ? 'bg-blue-50 border-blue-300 text-blue-700'
                        : 'bg-gray-50 border-gray-300 text-gray-500'
                    }`}
                  >
                    {cluster}
                  </button>
                </AntdTooltip>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

export default ControlPanel
