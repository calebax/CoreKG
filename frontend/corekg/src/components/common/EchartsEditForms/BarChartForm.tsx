import React from 'react'
import { Form, Input, Select, Space, Divider, Switch } from 'antd'

interface BarChartFormProps {
  config: any
  onChange: (config: any) => void
  chartType: 'bar' | 'line' | 'pie'
  onChartTypeChange: (type: 'bar' | 'line' | 'pie') => void
  theme: 'light' | 'dark'
  onThemeChange: (theme: 'light' | 'dark') => void
}

export const BarChartForm: React.FC<BarChartFormProps> = ({
  config,
  onChange,
  chartType,
  onChartTypeChange,
  theme,
  onThemeChange,
}) => {
  // 更新配置的通用方法
  const updateConfig = (path: string[], value: any) => {
    const newConfig = JSON.parse(JSON.stringify(config))
    let current = newConfig

    for (let i = 0; i < path.length - 1; i++) {
      if (!current[path[i]]) current[path[i]] = {}
      current = current[path[i]]
    }

    current[path[path.length - 1]] = value
    onChange(newConfig)
  }

  // 获取配置值的通用方法
  const getConfigValue = (path: string[], defaultValue: any = '') => {
    let current = config
    for (const key of path) {
      if (!current || typeof current !== 'object') return defaultValue
      current = current[key]
    }
    return current !== undefined ? current : defaultValue
  }

  // 检查是否应该显示图例配置（检查配置中是否已经有legend或者多个系列）
  const hasLegendData = () => {
    const series = config?.series
    if (!series || !Array.isArray(series)) return false

    // 如果配置中已经有legend，那么显示图例配置
    if (config?.legend) return true

    // 多个系列时通常需要图例（但也要检查系列有name）
    if (series.length > 1) {
      return series.some((s: any) => s.name && s.name.trim())
    }

    // 单个系列不显示图例配置（因为通常不需要）
    return false
  }

  return (
    <Space direction='vertical' className='w-full' size='large'>
      {/* 通用配置 */}
      <div>
        <Divider>通用配置</Divider>

        <Form.Item label='图表类型'>
          <Select
            value={chartType}
            onChange={onChartTypeChange}
            options={[
              { label: '柱状图', value: 'bar' },
              { label: '折线图', value: 'line' },
              { label: '饼图', value: 'pie' },
            ]}
          />
        </Form.Item>

        <Form.Item label='主题'>
          <Select
            value={theme}
            onChange={onThemeChange}
            options={[
              { label: '浅色', value: 'light' },
              { label: '深色', value: 'dark' },
            ]}
          />
        </Form.Item>

        <Form.Item label='主标题'>
          <Input
            value={getConfigValue(['title', 'text'])}
            onChange={(e) => updateConfig(['title', 'text'], e.target.value)}
            placeholder='月度销售业绩分析'
          />
        </Form.Item>

        <Form.Item label='副标题'>
          <Input
            value={getConfigValue(['title', 'subtext'])}
            onChange={(e) => updateConfig(['title', 'subtext'], e.target.value)}
            placeholder='各月份销量对比'
          />
        </Form.Item>
      </div>

      {/* 图例配置 - 只有在有图例数据时才显示 */}
      {hasLegendData() && (
        <div>
          <Divider>图例配置</Divider>

          <Form.Item label='显示图例'>
            <Switch
              size='small'
              checked={getConfigValue(['legend', 'show'], true)}
              onChange={(checked) => updateConfig(['legend', 'show'], checked)}
            />
          </Form.Item>

          <Form.Item label='图例位置'>
            <Select
              value={getConfigValue(['legend', 'orient'], 'horizontal')}
              onChange={(value) => updateConfig(['legend', 'orient'], value)}
              options={[
                { label: '水平', value: 'horizontal' },
                { label: '垂直', value: 'vertical' },
              ]}
            />
          </Form.Item>

          <Form.Item label='图例对齐'>
            <Select
              value={getConfigValue(['legend', 'left'], 'center')}
              onChange={(value) => updateConfig(['legend', 'left'], value)}
              options={[
                { label: '左侧', value: 'left' },
                { label: '居中', value: 'center' },
                { label: '右侧', value: 'right' },
              ]}
            />
          </Form.Item>
        </div>
      )}

      {/* 坐标轴配置 */}
      <div>
        <Divider>坐标轴配置</Divider>

        <Form.Item label='X轴位置'>
          <Select
            value={getConfigValue(['xAxis', 'position'], 'bottom')}
            onChange={(value) => updateConfig(['xAxis', 'position'], value)}
            options={[
              { label: '底部', value: 'bottom' },
              { label: '顶部', value: 'top' },
            ]}
          />
        </Form.Item>

        <Form.Item label='Y轴位置'>
          <Select
            value={getConfigValue(['yAxis', 'position'], 'left')}
            onChange={(value) => updateConfig(['yAxis', 'position'], value)}
            options={[
              { label: '左侧', value: 'left' },
              { label: '右侧', value: 'right' },
            ]}
          />
        </Form.Item>
      </div>

      {/* 柱状图专用配置 */}
      <div>
        <Divider>柱状图配置</Divider>

        <Form.Item label='柱子颜色'>
          <Input
            type='color'
            value={
              getConfigValue(
                ['series', '0', 'itemStyle', 'color'],
                '#3b82f6',
              ) as string
            }
            onChange={(e) =>
              updateConfig(
                ['series', '0', 'itemStyle', 'color'],
                e.target.value,
              )
            }
          />
        </Form.Item>
      </div>
    </Space>
  )
}

export default BarChartForm
