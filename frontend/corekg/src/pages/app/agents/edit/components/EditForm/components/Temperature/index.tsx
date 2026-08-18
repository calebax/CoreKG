import { FC } from 'react'
import { Slider, Tooltip } from 'antd'
import { QuestionCircleOutlined } from '@ant-design/icons'
import { Agent } from 'Agent'
import { cn } from '@/utils'
import { FormItem } from '../..'
import styles from './styles.module.scss'

/** 温度 */
export const Temperature: FC<Style> = (props) => {
  return (
    <FormItem
      name={'temperature'}
      rules={[{ required: true, message: '请选择温度' }]}
      className='m-0'
    >
      <InnerTemperature {...props} />
    </FormItem>
  )
}

export const InnerTemperature: FC<
  Style & ValueController<Agent['temperature']>
> = (props) => {
  const { value, onChange } = props
  return (
    <div
      className={cn('flex items-center', props.className)}
      style={props.style}
    >
      <span className='text-title text-base'>温度</span>
      <Tooltip
        title='较高的数值会使输出更加随机，而更低的数值会使其更加集中和确定'
        placement='right'
      >
        <QuestionCircleOutlined className='mx-1 mr-2.5 text-[#C9CDD4]' />
      </Tooltip>
      <Slider
        className={cn('flex-1', styles.slider)}
        value={value}
        onChange={onChange}
        max={2}
        min={0.01}
        step={0.01}
      />
    </div>
  )
}
