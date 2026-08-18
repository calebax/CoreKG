import React, { useEffect, useState, useRef } from 'react'
import { Button, Tooltip } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'

interface EchartsBlockProps {
  /** ECharts配置的JSON字符串 */
  config: string
  /** 图表唯一标识，用于编辑功能 */
  id?: string
  /** 是否可编辑（支持柱状图、折线图、饼图） */
  editable?: boolean
  /** 编辑回调函数 */
  onEdit?: (id: string, newConfig: any) => void
}

// 判断图表类型是否支持编辑
const isEditableChart = (option: any): boolean => {
  if (!option?.series || !Array.isArray(option.series)) {
    return false
  }

  const chartTypes = option.series.map((s: any) => s.type)
  const editableTypes = ['bar', 'line', 'pie']

  return chartTypes.every((type: string) => editableTypes.includes(type))
}

export const EchartsBlock: React.FC<EchartsBlockProps> = ({
  config,
  id,
  editable = true,
  onEdit,
}) => {
  const [chartOption, setChartOption] = useState<any>(null)
  const [error, setError] = useState<string | null>(null)
  const chartRef = useRef<any>(null)

  useEffect(() => {
    try {
      const parsedOption = JSON.parse(config)

      // 检查是否是空对象或无效配置
      if (
        !parsedOption ||
        Object.keys(parsedOption).length === 0 ||
        !parsedOption.series ||
        parsedOption.series.length === 0
      ) {
        setChartOption(null)
        setError('ECharts配置为空或无效')
        return
      }

      // 应用优化配置，与编辑弹窗保持一致
      const optimizedOption = { ...parsedOption }
      const chartType = optimizedOption.series?.[0]?.type

      // 所有图表都应用标题配置
      if (optimizedOption.title) {
        optimizedOption.title = {
          ...optimizedOption.title,
          top: chartType === 'pie' ? '5%' : '2%',
          textStyle: {
            fontSize: 16,
            ...optimizedOption.title.textStyle,
          },
          subtextStyle: {
            fontSize: 12,
            ...optimizedOption.title.subtextStyle,
          },
        }
      }

      // 所有图表都应用图例配置
      if (optimizedOption.legend) {
        optimizedOption.legend = {
          ...optimizedOption.legend,
          top: chartType === 'pie' ? '15%' : '13%',
        }
      }

      // 根据图表类型应用不同的间距配置
      if (chartType === 'pie') {
        // 饼图：调整center和radius给标题和图例留空间
        optimizedOption.series = optimizedOption.series.map((series: any) => ({
          ...series,
          center: ['50%', '60%'], // 向下移动，给标题和图例留空间
          radius: ['30%', '70%'], // 调整半径，给周围留出空间
        }))
      } else if (chartType === 'radar') {
        // 雷达图：调整radar配置
        if (optimizedOption.radar) {
          optimizedOption.radar = {
            ...optimizedOption.radar,
            center: ['50%', '60%'], // 向下移动，给标题和图例留空间
            radius: '65%', // 调整半径给周围留空间
          }
        }
      } else {
        // 其他图表类型（柱状图、折线图等）：使用grid
        optimizedOption.grid = {
          left: '10%',
          right: '10%',
          top: '25%',
          bottom: '15%',
          containLabel: true,
          ...optimizedOption.grid,
        }
      }

      setChartOption(optimizedOption)
      setError(null)
    } catch (err) {
      setError('ECharts配置解析失败，请检查JSON格式')
      console.error('ECharts config parse error:', err)
    }
  }, [config])

  // 处理图表resize
  useEffect(() => {
    const handleResize = () => {
      if (chartRef.current) {
        const echartsInstance = chartRef.current.getEchartsInstance()
        if (echartsInstance) {
          // 延迟执行resize，确保DOM已经更新
          setTimeout(() => {
            echartsInstance.resize()
          }, 100)
        }
      }
    }

    window.addEventListener('resize', handleResize)

    // 组件挂载后也触发一次resize
    const timer = setTimeout(handleResize, 200)

    return () => {
      window.removeEventListener('resize', handleResize)
      clearTimeout(timer)
    }
  }, [chartOption])

  const handleEdit = () => {
    if (onEdit && id && chartOption) {
      onEdit(id, chartOption)
    }
  }

  if (error) {
    return (
      <div className='border border-red-300 rounded-lg p-4 bg-red-50'>
        <p className='text-red-600 text-sm m-0'>{error}</p>
        <pre className='mt-2 text-xs bg-red-100 p-2 rounded overflow-auto'>
          {config}
        </pre>
      </div>
    )
  }

  if (!chartOption) {
    return (
      <div className='border border-gray-300 rounded-lg p-4 bg-gray-50'>
        <p className='text-gray-600 text-sm m-0'>加载图表中...</p>
      </div>
    )
  }

  const canEdit = editable && isEditableChart(chartOption) && onEdit

  return (
    <div className='relative border border-gray-200 rounded-lg overflow-hidden bg-white w-full'>
      {/* 编辑按钮 - 常显示 */}
      {canEdit && (
        <div className='absolute top-2 right-2 z-10'>
          <Tooltip title='编辑图表'>
            <Button
              type='primary'
              size='small'
              icon={<EditOutlined />}
              onClick={handleEdit}
              className='shadow-md'
            >
              编辑
            </Button>
          </Tooltip>
        </div>
      )}

      {/* ECharts图表 */}
      <div className='p-4 w-full'>
        <ReactECharts
          ref={chartRef}
          option={chartOption}
          style={{
            width: '100%',
            height: '500px',
            minWidth: '400px',
          }}
          opts={{
            renderer: 'canvas',
            width: 'auto',
            height: 'auto',
          }}
          lazyUpdate={true}
          notMerge={false}
          onEvents={{}}
        />
      </div>
    </div>
  )
}

export default EchartsBlock
