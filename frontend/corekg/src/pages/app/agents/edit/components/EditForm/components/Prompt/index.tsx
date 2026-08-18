import { FC } from 'react'
import { Input, Tooltip } from 'antd'
import { QuestionCircleOutlined } from '@ant-design/icons'
import { Agent } from 'Agent'
import { cn } from '@/utils'
import { FormItem } from '../..'
import styles from './styles.module.scss'

/** 提示词 */
export const Prompt: FC<Style> = (props) => {
  return (
    <FormItem
      name='prompt_template'
      className='m-0'
      rules={[{ required: true, message: '请输入提示词' }]}
    >
      <InnerPrompt {...props} />
    </FormItem>
  )
}
type Value = Agent['prompt_template']

/** 提示词 */
export const InnerPrompt: FC<Style & ValueController<Value>> = (props) => {
  const { value, onChange } = props

  const placeholder = [
    '#角色名称：描述角色概述和主要职责的一句话',
    '#风格特点：角色说话风格、性格特点',
    '#输出要求：限制角色输出格式、内容字数、输出语言等',
    '#能力限制：描述角色能力范围',
  ].join('\n')
  return (
    <div
      className={cn('flex flex-col gap-1', props.className)}
      style={props.style}
    >
      <div className='flex'>
        <span className='text-base text-title font-medium'>提示词</span>
        <Tooltip
          placement='right'
          title='在提示词中必须使用{{变量名}}语法进行引用，确保变量的定义与使用完全一致。示例：请根据{{文件名称}}，回答{{用户问题}}。'
        >
          <QuestionCircleOutlined className='text-[#C9CDD4] ml-1' />
        </Tooltip>
      </div>
      <div className='overflow-hidden  min-h-[108px]'>
        <Input.TextArea
          placeholder={placeholder}
          value={value}
          onChange={(e) => {
            onChange?.(e.target.value)
          }}
          className={cn(' resize-y h-full min-h-[108px] p-2 ', styles.scroll)}
        ></Input.TextArea>
      </div>
    </div>
  )
}
