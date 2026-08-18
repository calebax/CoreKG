import { FC } from 'react'
import { Select } from 'antd'
import { Agent } from 'Agent'
import { cn } from '@/utils'
import { FormItem, useEditContext } from '../..'
import styles from './styles.module.scss'

/** 大模型选择 */
export const ModelSelect: FC<Style> = (props) => {
  return (
    <FormItem
      name='chat_models'
      rules={[{ required: true, message: '请选择大模型' }]}
    >
      <InnerModelSelect {...props} />
    </FormItem>
  )
}

export const InnerModelSelect: FC<
  Style & ValueController<Agent['chat_models']>
> = (props) => {
  const { value, onChange } = props
  const { models } = useEditContext()
  const options = useMemo(() => {
    return models.map((item) => {
      return {
        value: item.id,
        label: item.name,
        description: item.description,
      }
    })
  }, [models])
  if (!options) return null
  return (
    <div
      className={cn('flex flex-col gap-1', props.className)}
      style={props.style}
    >
      <span className='text-base text-title font-medium'>选择大模型</span>
      <Select
        placeholder='请选择大模型'
        options={options}
        value={value?.[0].id}
        onSelect={(id) => {
          onChange?.(models.filter((item) => item.id === id))
        }}
        showSearch
        filterOption={(search, option) => {
          return Boolean(option?.label.includes(search))
        }}
        popupMatchSelectWidth={230}
        className={styles.modelSelect}
        classNames={{ popup: { root: styles.scroll } }}
      />
    </div>
  )
}
