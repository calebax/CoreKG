import { FC } from 'react'
import { Form, Radio } from 'antd'
import { Agent } from 'Agent'
import { cn } from '@/utils'
import { FormItem } from '../..'

/** public_scope 和 scope_type */
export const PublicScope: FC<Style> = (props) => {
  return (
    <FormItem
      name='public_scope'
      rules={[
        {
          validator: async (_, val: Agent['public_scope']) => {
            switch (val) {
              case 'private':
              case 'public':
                throw new Error('请选择公开类型')
            }
          },
        },
      ]}
    >
      <InnerPublicScope {...props} />
    </FormItem>
  )
}

const InnerPublicScope: FC<Style & ValueController<Agent['public_scope']>> = (
  props,
) => {
  const { value, onChange } = props
  const form = Form.useFormInstance()

  return (
    <div
      className={cn('flex flex-col gap-1', props.className)}
      style={props.style}
    >
      <span className='text-base text-title font-medium'>公开类型</span>
      <Radio.Group
        value={value}
        onChange={(e) => {
          const public_scope = e.target.value
          onChange?.(public_scope)
          if (public_scope === 'company') {
            form.setFieldValue('scope_type', 'company')
          } else if (public_scope === 'custom') {
            form.setFieldValue('scope_type', 'user')
          }
        }}
        options={[
          {
            value: 'company',
            label: '公司',
          },
          {
            value: 'custom',
            label: '自定义',
          },
        ]}
      />
    </div>
  )
}
