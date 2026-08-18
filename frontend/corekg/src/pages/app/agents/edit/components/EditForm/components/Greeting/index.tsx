import { FC } from 'react'
import { Form, Input } from 'antd'
import { Agent } from 'Agent'
import { cn } from '@/utils'
import { FormItem } from '../..'
import styles from './styles.module.scss'

/** 欢迎语 */
export const Greeting: FC<Style> = (props) => {
  const type = Form.useWatch<Agent['type']>('type', { preserve: true })
  if (type !== 'role_play' && type !== 'knowledge') return null
  return (
    <FormItem
      name={'greeting_message'}
      rules={[{ required: true, message: '请输入欢迎语' }]}
    >
      <InnerGreeting {...props} />
    </FormItem>
  )
}

const InnerGreeting: FC<Style & ValueController<Agent['greeting_message']>> = (
  props,
) => {
  const { value, onChange } = props
  return (
    <div
      className={cn('flex flex-col gap-1', props.className)}
      style={props.style}
    >
      <span className='text-base text-title font-medium'>欢迎语</span>
      <Input.TextArea
        placeholder='请输入欢迎语'
        value={value}
        className={cn(' resize-none h-full min-h-[108px] p-2 ', styles.scroll)}
        onChange={(e) => onChange?.(e.target.value)}
      ></Input.TextArea>
    </div>
  )
}
