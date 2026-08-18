import { FC } from 'react'
import { Form, Select, Tag } from 'antd'
import { uniqueArray } from '@/utils'
import { GraphMetaValues } from '..'
import { useEmployee } from '../../EmployeeProvider'

export const ScopeIds: FC<
  Style & ValueController<GraphMetaValues['scope_ids']>
> = (props) => {
  const { value, onChange } = props
  const manager_ids: number[] =
    Form.useWatch('manager_ids', { preserve: true }) ?? []

  const employee = useEmployee()
  const options = useMemo(() => {
    return employee?.map((item) => {
      return {
        value: item.uin,
        label: item.name,
      }
    })
  }, [employee])
  return (
    <Select
      placeholder='请选择公开范围'
      mode='multiple'
      value={value}
      onChange={(val) => {
        const newScopedIds = uniqueArray(manager_ids, val)
        onChange?.(newScopedIds)
      }}
      options={options}
      optionFilterProp='label'
      className={props.className}
      style={props.style}
      tagRender={(props) => {
        return (
          <Tag
            {...props}
            closable={!manager_ids?.includes(props.value)}
            className='text-base my-0.5 mr-1 pl-1 pr-1'
            style={{ backgroundColor: 'rgba(0, 0, 0, 0.06)' }}
            onMouseDown={(e) => e.stopPropagation()}
          >
            {props.label}
          </Tag>
        )
      }}
    />
  )
}
