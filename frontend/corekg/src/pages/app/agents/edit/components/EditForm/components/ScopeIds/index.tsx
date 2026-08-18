import { FC } from 'react'
import { Form, Select, Tag } from 'antd'
import { Agent } from 'Agent'
import { uniqueArray } from '@/utils'
import { useEditContext } from '../..'

/** 自定义公开用户scope_ids */
export const ScopeIds: FC<Style> = (props) => {
  const scope_type = Form.useWatch('public_scope', { preserve: true })
  if (scope_type !== 'custom') return null
  return (
    <Form.Item
      name='scope_ids'
      rules={[{ required: true, message: '请选择公开范围' }]}
    >
      <InnerScopeIds {...props} />
    </Form.Item>
  )
}

const InnerScopeIds: FC<Style & ValueController<Agent['scope_ids']>> = (
  props,
) => {
  const { value, onChange } = props
  const manager_ids: number[] = Form.useWatch('manager_ids') ?? []

  const { managers } = useEditContext()
  const options = useMemo(() => {
    return managers.map((item) => {
      return {
        value: item.uin,
        label: item.name,
      }
    })
  }, [managers])
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
