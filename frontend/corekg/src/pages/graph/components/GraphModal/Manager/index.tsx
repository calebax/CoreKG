import { FC } from 'react'
import { Select, Tag, Form } from 'antd'
import { cn, uniqueArray } from '@/utils'
import { useAdmin } from '@/utils/useAdmin'
import { GraphMetaValues } from '..'
import { useEmployee } from '../../EmployeeProvider'

export const Manager: FC<
  Style & ValueController<GraphMetaValues['manager_ids']>
> = (props) => {
  const { value, onChange } = props
  const form = Form.useFormInstance()
  const scope_ids: number[] = Form.useWatch('scope_ids') ?? []

  const { adminIds } = useAdmin()

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
    <div
      className={cn('flex flex-col gap-1', props.className)}
      style={props.style}
    >
      <span className='text-base text-title font-medium'>管理员</span>
      <Select
        mode='multiple'
        placeholder='请选择管理员'
        options={options}
        value={value}
        onChange={(val) => {
          onChange?.(uniqueArray(adminIds, val))
          if (!value || val.length > value.length) {
            // 如果之前没有值 或 新值长度增加 则是选择管理员
            // 此时将其同步至公开范围
            const newScopedIds = uniqueArray(adminIds, scope_ids, val)
            form.setFieldValue('scope_ids', newScopedIds)
          }
        }}
        optionFilterProp='label'
        tagRender={(props) => {
          return (
            <Tag
              {...props}
              closable={!adminIds.includes(props.value)}
              className='text-base my-0.5 mr-1 pl-1 pr-1'
              style={{ backgroundColor: 'rgba(0, 0, 0, 0.06)' }}
              onMouseDown={(e) => e.stopPropagation()}
            >
              {props.label}
            </Tag>
          )
        }}
      />
    </div>
  )
}
