import { FC } from 'react'
import { Select } from 'antd'
import { cn } from '@/utils'
import { useModelData } from '@/hooks/data'

type SelectModel = {
  className?: string
  style?: React.CSSProperties
  value?: number
  onChange: (val: number) => void
}
const SelectModel: FC<SelectModel> = (props) => {
  const { value, onChange, className, style } = props
  const { modelList } = useModelData()
  if (!modelList.length) return null
  return (
    <Select
      className={cn('w-full h-10', className)}
      style={style}
      placeholder='请选择模型'
      value={value}
      onChange={onChange}
      options={modelList}
      showSearch
      filterOption={(val, option) => {
        return Boolean((option?.label as string).includes(val))
      }}
    />
  )
}
export default SelectModel
