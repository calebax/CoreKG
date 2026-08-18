import { Select, Tag } from 'antd'
import type { SelectProps } from 'antd'
import { useUserData } from '@/hooks/data'

interface SelectUserProps {
  className?: string
  placeholder?: string
  mode?: SelectProps['mode']
  value?: any
  onChange?: (value: any) => void
  requiredIds?: number[]
}

export default function SelectUser({
  className,
  placeholder,
  mode,
  value,
  onChange,
  requiredIds,
}: SelectUserProps) {
  const { userList } = useUserData()
  return (
    <Select
      className={className}
      placeholder={placeholder}
      mode={mode}
      options={userList}
      value={value}
      onChange={onChange}
      tagRender={(props) => {
        return (
          <Tag
            {...props}
            closable={!requiredIds?.includes(props.value)}
            className='text-sm my-0.5 mr-2 pl-2 pr-2'
            style={{ backgroundColor: '#FAFAFA', border: '1px solid #D9D9D9' }}
            onMouseDown={(e) => e.stopPropagation()}
          >
            {props.label}
          </Tag>
        )
      }}
    ></Select>
  )
}
