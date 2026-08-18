import { FC, ReactNode } from 'react'
import { cn } from '@/utils'
import { Param } from '..'
import SelectIcon from './images/select.svg?react'
import TextIcon from './images/text.svg?react'

type Value = Param['input_type']
export const ParamTypeSelect: FC<Style & ValueController<Value>> = (props) => {
  const { value, onChange } = props
  const options: { label: string; icon: ReactNode; value: Value }[] = [
    { label: '文本', value: 'text', icon: <TextIcon /> },
    { label: '选择', value: 'select', icon: <SelectIcon /> },
  ]
  return (
    <div
      className={cn('flex items-center gap-4', props.className)}
      style={props.style}
    >
      {options.map((option) => {
        const { label, value: optionValue, icon } = option
        return (
          <div
            key={option.value}
            className={cn(
              'flex-1 py-3 cursor-pointer rounded border',
              'flex flex-col items-center gap-2.5',
              'border-[#BEDAFF]',
              {
                'border-primary bg-[#EAF3FF] text-[#165DFF]':
                  value === optionValue,
              },
            )}
            onClick={() => onChange?.(optionValue)}
          >
            {icon}
            {label}
          </div>
        )
      })}
    </div>
  )
}
