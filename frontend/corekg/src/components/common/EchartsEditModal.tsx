import React, { useState, useEffect, useMemo } from 'react'
import { Modal, Form, Button, Space } from 'antd'
import ReactECharts from 'echarts-for-react'
import scrollStyles from '../../styles/scroll/styles.module.scss'
import { BarChartForm, LineChartForm, PieChartForm } from './EchartsEditForms'

interface EchartsEditModalProps {
  visible: boolean
  onClose: () => void
  onSave: (config: any) => void
  initialConfig: any
}

type ChartType = 'bar' | 'line' | 'pie'

// 判断图表类型
const getChartType = (config: any): ChartType => {
  if (!config?.series || !Array.isArray(config.series)) {
    return 'bar'
  }

  const firstSeriesType = config.series[0]?.type
  if (['bar', 'line', 'pie'].includes(firstSeriesType)) {
    return firstSeriesType as ChartType
  }

  return 'bar'
}

// 转换图表类型
const convertChartType = (config: any, targetType: ChartType): any => {
  if (!config || !config.series) return config

  const newConfig = JSON.parse(JSON.stringify(config))

  newConfig.series = newConfig.series.map((series: any) => ({
    ...series,
    type: targetType,
    // 根据图表类型调整特定配置
    ...(targetType === 'pie' && {
      radius: '50%',
      center: ['50%', '50%'],
    }),
    ...(targetType === 'line' && {
      smooth: true,
    }),
  }))

  // 饼图不需要坐标轴
  if (targetType === 'pie') {
    delete newConfig.xAxis
    delete newConfig.yAxis
    delete newConfig.grid
  } else if (!newConfig.xAxis && config.series[0]?.data) {
    // 如果是从饼图转换过来，需要添加坐标轴
    const dataLength = config.series[0].data.length
    newConfig.xAxis = {
      type: 'category',
      data: Array.from({ length: dataLength }, (_, i) => `类目${i + 1}`),
    }
    newConfig.yAxis = {
      type: 'value',
    }
  }

  return newConfig
}

export const EchartsEditModal: React.FC<EchartsEditModalProps> = ({
  visible,
  onClose,
  onSave,
  initialConfig,
}) => {
  const [form] = Form.useForm()
  const [currentConfig, setCurrentConfig] = useState<any>(initialConfig || {})
  const [chartType, setChartType] = useState<ChartType>('bar')
  const [theme, setTheme] = useState<'light' | 'dark'>('light')
  const [chartReady, setChartReady] = useState(false)

  useEffect(() => {
    if (initialConfig) {
      setCurrentConfig(initialConfig)
      setChartType(getChartType(initialConfig))
    }
  }, [initialConfig])

  // 当Modal打开时，延迟一帧让容器尺寸计算完成
  useEffect(() => {
    if (visible) {
      setChartReady(false)
      const timer = setTimeout(() => {
        setChartReady(true)
      }, 100)
      return () => clearTimeout(timer)
    } else {
      setChartReady(false)
    }
  }, [visible])

  // 处理配置更新
  const handleConfigChange = (newConfig: any) => {
    setCurrentConfig(newConfig)
  }

  // 处理图表类型切换
  const handleChartTypeChange = (newType: ChartType) => {
    const convertedConfig = convertChartType(currentConfig, newType)
    setChartType(newType)
    setCurrentConfig(convertedConfig)
  }

  // 处理主题切换
  const handleThemeChange = (newTheme: 'light' | 'dark') => {
    setTheme(newTheme)
    const updatedConfig = {
      ...currentConfig,
      backgroundColor: newTheme === 'dark' ? '#1e1e1e' : 'transparent',
      textStyle: {
        color: newTheme === 'dark' ? '#ffffff' : '#333333',
      },
    }
    setCurrentConfig(updatedConfig)
  }

  // 处理保存
  const handleSave = () => {
    onSave(currentConfig)
    onClose()
  }

  // 预览配置（应用主题并确保正确的布局间距）
  const previewConfig = useMemo(() => {
    const config = { ...currentConfig }

    // 确保标题有正确的位置和样式
    if (config.title) {
      config.title = {
        ...config.title,
        top: '2%',
        textStyle: {
          fontSize: 16,
          ...config.title.textStyle,
        },
        subtextStyle: {
          fontSize: 12,
          ...config.title.subtextStyle,
        },
      }
    }

    // 确保图例有正确的位置
    if (config.legend) {
      config.legend = {
        ...config.legend,
        top: '13%',
      }
    }

    // 确保grid有正确的间距
    if (config.grid && chartType !== 'pie') {
      config.grid = {
        left: '10%',
        right: '10%',
        top: '25%',
        bottom: '15%',
        containLabel: true,
        ...config.grid,
      }
    }

    if (chartType === 'pie') {
      // 饼图：调整center和radius给标题和图例留空间
      config.series = config.series.map((series: any) => ({
        ...series,
        center: ['50%', '55%'], // 向下移动，给标题和图例留空间
        radius: ['30%', '70%'], // 调整半径，给周围留出空间
      }))
    }

    // 应用主题
    if (theme === 'dark') {
      config.backgroundColor = '#1e1e1e'
      config.textStyle = {
        ...config.textStyle,
        color: '#ffffff',
      }
      // 更新坐标轴颜色
      if (config.xAxis) {
        config.xAxis = {
          ...config.xAxis,
          axisLine: { lineStyle: { color: '#ffffff' } },
          axisLabel: { color: '#ffffff' },
        }
      }
      if (config.yAxis) {
        config.yAxis = {
          ...config.yAxis,
          axisLine: { lineStyle: { color: '#ffffff' } },
          axisLabel: { color: '#ffffff' },
        }
      }
    }

    return config
  }, [currentConfig, theme, chartType])

  const renderForm = () => {
    const commonProps = {
      config: currentConfig,
      onChange: handleConfigChange,
      chartType,
      onChartTypeChange: handleChartTypeChange,
      theme,
      onThemeChange: handleThemeChange,
    }

    switch (chartType) {
      case 'bar':
        return <BarChartForm {...commonProps} />
      case 'line':
        return <LineChartForm {...commonProps} />
      case 'pie':
        return <PieChartForm {...commonProps} />
      default:
        return <BarChartForm {...commonProps} />
    }
  }

  return (
    <Modal
      title='编辑图表'
      open={visible}
      onCancel={onClose}
      width={1200}
      style={{ top: 20 }}
      bodyStyle={{ height: '80vh', overflow: 'hidden' }}
      footer={
        <Space>
          <Button onClick={onClose}>取消</Button>
          <Button type='primary' onClick={handleSave}>
            保存
          </Button>
        </Space>
      }
    >
      <div className='flex h-full'>
        {/* 左侧预览区 */}
        <div className='flex-1 border-r border-gray-200 pr-4'>
          <div className='h-full flex flex-col'>
            <h3 className='text-lg font-semibold mb-4'>图表预览</h3>
            <div className='flex-1 border border-gray-200 rounded-lg p-3'>
              {chartReady ? (
                <ReactECharts
                  option={previewConfig}
                  style={{
                    width: '100%',
                    height: '100%',
                  }}
                  theme={theme}
                  opts={{
                    renderer: 'canvas',
                  }}
                  key={`chart-${visible}-${Date.now()}`}
                />
              ) : (
                <div className='flex items-center justify-center h-full'>
                  <div className='text-gray-500'>图表加载中...</div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* 右侧配置区 */}
        <div className='w-96 pl-4'>
          <div className={`h-full overflow-auto pr-3 ${scrollStyles.scroll}`}>
            <h3 className='text-lg font-semibold mb-4'>图表配置</h3>
            <Form form={form} layout='vertical' size='small'>
              {renderForm()}
            </Form>
          </div>
        </div>
      </div>
    </Modal>
  )
}

export default EchartsEditModal
