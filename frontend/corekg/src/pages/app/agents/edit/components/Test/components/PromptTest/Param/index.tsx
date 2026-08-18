import { FC } from 'react'
import { Input, Select } from 'antd'
import { Agent } from 'Agent'

export const Param: FC<
  Style &
    ValueController<string> & {
      param: Agent['params'][number]
    }
> = (props) => {
  const { value, onChange, param } = props
  const { input_type, input_array } = param

  return input_type === 'text' ? (
    <Input value={value} onChange={(e) => onChange?.(e.target.value)} />
  ) : (
    <Select
      value={value}
      onChange={onChange}
      showSearch
      options={input_array?.map((v) => ({ label: v, value: v }))}
    />
  )
}
