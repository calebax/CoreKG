import { FC, memo, useEffect, useMemo, useRef } from 'react'
import {
  AutoComplete,
  App,
  Button,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Tag,
} from 'antd'
import type { Property } from 'Graph'
import { PropertyLabelMap } from 'Graph'
import { useBoolean } from 'ahooks'
import { Delete1Icon, EditIcon } from 'tdesign-icons-react'
import { AddCircleIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import EmptyPic from '../../edit/Step2/GraphTagDetails/EmptyPic.svg?react'
import styles from './styles.module.scss'

export type NodePropertyValue = {
  name: string
  value: null | string | number | boolean
  /** 自定义属性时用于 UI 展示与输入控件选择（提交到后端前会剥离） */
  type?: Property['type']
}

export type NodePropertiesValue = {
  /** 属性定义（名称 + 类型） */
  properties: Property[]
  /** 属性值（名称 + 值） */
  properties_values: NodePropertyValue[]
}

type NodePropertiesProps = Style & {
  value: NodePropertiesValue
  onChange: (value: NodePropertiesValue) => void
  tagId?: number
  tagName?: string
  /** 只读模式：隐藏新建/编辑/删除等操作入口 */
  disabled?: boolean
}

const syncNodePropertiesValue = (
  input: NodePropertiesValue,
): NodePropertiesValue => {
  const normalizedProperties = (input.properties ?? [])
    .map((p) => ({ ...p, name: (p.name ?? '').trim() }))
    .filter((p) => !!p.name)
    .reduce<Property[]>((acc, cur) => {
      const existedIndex = acc.findIndex((p) => p.name === cur.name)
      if (existedIndex === -1) return [...acc, cur]
      const next = [...acc]
      next[existedIndex] = cur
      return next
    }, [])

  const normalizedValues = (input.properties_values ?? [])
    .map((pv) => ({ ...pv, name: (pv.name ?? '').trim() }))
    .filter((pv) => !!pv.name)
    .reduce<NodePropertyValue[]>((acc, cur) => {
      const existedIndex = acc.findIndex((pv) => pv.name === cur.name)
      if (existedIndex === -1) return [...acc, cur]
      const next = [...acc]
      next[existedIndex] = cur
      return next
    }, [])

  const mergedProperties: Property[] = [...normalizedProperties]
  normalizedValues.forEach((pv) => {
    if (mergedProperties.some((p) => p.name === pv.name)) return
    mergedProperties.push({
      name: pv.name,
      type: pv.type ?? 'string',
    })
  })

  const nextValues: NodePropertyValue[] = mergedProperties.map((p) => {
    const existed = normalizedValues.find((pv) => pv.name === p.name)
    if (existed) {
      return {
        ...existed,
        type: existed.type ?? p.type,
      }
    }
    return { name: p.name, value: null, type: p.type }
  })

  const sortedProperties = [...mergedProperties].sort((a, b) =>
    a.name.localeCompare(b.name, 'zh-Hans-CN'),
  )
  const sortedValues = [...nextValues].sort((a, b) =>
    a.name.localeCompare(b.name, 'zh-Hans-CN'),
  )

  return {
    properties: sortedProperties,
    properties_values: sortedValues,
  }
}

const isSameNodePropertiesValue = (
  a: NodePropertiesValue,
  b: NodePropertiesValue,
): boolean => {
  if (a.properties.length !== b.properties.length) return false
  if (a.properties_values.length !== b.properties_values.length) return false

  for (let i = 0; i < a.properties.length; i += 1) {
    const ap = a.properties[i]
    const bp = b.properties[i]
    if (ap.name !== bp.name || ap.type !== bp.type) return false
  }

  for (let i = 0; i < a.properties_values.length; i += 1) {
    const av = a.properties_values[i]
    const bv = b.properties_values[i]
    if (av.name !== bv.name) return false
    if (av.value !== bv.value) return false
    if ((av.type ?? null) !== (bv.type ?? null)) return false
  }

  return true
}

export const NodeProperties: FC<NodePropertiesProps> = (props) => {
  const { value, onChange, tagId, tagName, disabled, className, style } = props
  const { data } = useGraphInfo()
  const { tags } = data ?? {}

  // 根据 tagId 或 tagName 找到对应的实体类型
  const currentTag = useMemo(() => {
    if (!tags) return undefined
    if (tagId !== undefined) {
      return tags.find((t) => t.tag_id === tagId)
    }
    if (tagName !== undefined) {
      return tags.find((t) => t.tag_name === tagName)
    }
    return undefined
  }, [tags, tagId, tagName])
  const tagProperties = useMemo(() => {
    return currentTag?.properties ?? []
  }, [currentTag?.properties])

  const mergedValue = useMemo(() => {
    const baseProperties =
      value.properties?.length > 0 ? value.properties : tagProperties
    return syncNodePropertiesValue({
      properties: baseProperties,
      properties_values: value.properties_values ?? [],
    })
  }, [tagProperties, value.properties, value.properties_values])

  useEffect(() => {
    if (isSameNodePropertiesValue(value, mergedValue)) return
    onChange(mergedValue)
  }, [mergedValue, onChange, value])

  const properties = mergedValue.properties
  const propertiesValues = mergedValue.properties_values
  const suggestedPropertyNameOptions = useMemo(() => {
    return properties.map((p) => ({
      label: `${p.name}`,
      value: p.name,
    }))
  }, [properties])

  const displayItems = useMemo(() => {
    return propertiesValues.filter((item) => {
      return ['string', 'number', 'boolean'].includes(typeof item.value)
    })
  }, [propertiesValues])

  const handleChange = (next: NodePropertiesValue) => {
    onChange(syncNodePropertiesValue(next))
  }

  return (
    <div className={cn('flex-1 flex flex-col gap-3', className)} style={style}>
      {/* 右上角按钮（与 Step2 同款位置） */}
      {disabled ? null : (
        <div className='flex justify-end'>
          <CreatePropertyButton value={mergedValue} onChange={handleChange} />
        </div>
      )}

      <div className='flex-1 overflow-auto flex flex-col gap-7'>
        {displayItems.length === 0 ? (
          <Empty
            image={<EmptyPic className='mx-auto' />}
            description='该实体暂无属性'
          />
        ) : (
          displayItems.map((item, index) => (
            <PropertyItem
              key={item.name}
              propertyName={item.name}
              propertyType={
                properties.find((p) => p.name === item.name)?.type ?? item.type
              }
              propertyValue={item}
              allPropertyNames={properties.map((p) => p.name)}
              suggestedPropertyNameOptions={suggestedPropertyNameOptions}
              index={index}
              value={mergedValue}
              onChange={handleChange}
              disabled={disabled}
            />
          ))
        )}
      </div>
    </div>
  )
}

const PropertyItem: FC<{
  propertyName: string
  propertyType?: Property['type']
  propertyValue?: NodePropertyValue
  allPropertyNames: string[]
  suggestedPropertyNameOptions: { label: string; value: string }[]
  index: number
  value: NodePropertiesValue
  onChange: (value: NodePropertiesValue) => void
  disabled?: boolean
}> = (props) => {
  const { modal } = App.useApp()
  const {
    propertyName,
    propertyType,
    propertyValue,
    allPropertyNames,
    suggestedPropertyNameOptions,
    value,
    onChange,
    disabled,
  } = props
  const name = propertyName
  const type = propertyType
  const [open, { toggle }] = useBoolean()

  const withWrapper = (children: any) => {
    return (
      <div className='flex-1 flex items-center justify-center'>{children}</div>
    )
  }

  const handleDelete = () => {
    modal.confirm({
      title: '确定删除？',
      onOk: () => {
        const next: NodePropertiesValue = {
          properties: value.properties.filter((p) => p.name !== name),
          properties_values: value.properties_values.filter(
            (pv) => pv.name !== name,
          ),
        }
        onChange(next)
      },
    })
  }

  const handleEdit = (newValue: NodePropertyValue) => {
    const nextName = newValue.name
    if (!nextName) return
    const isRenaming = nextName !== name
    if (isRenaming && allPropertyNames.includes(nextName)) {
      const existed = value.properties_values.find((pv) => pv.name === nextName)
      // 仅当目标属性已存在且它已经有 string/number/boolean 值时才提示重名
      if (
        existed &&
        ['string', 'number', 'boolean'].includes(typeof existed.value)
      ) {
        modal.warning({ title: '该属性名称已存在' })
        return
      }
    }
    const nextType =
      newValue.type ??
      value.properties.find((p) => p.name === name)?.type ??
      type ??
      'string'

    const nextProperties: Property[] = [
      ...value.properties.filter((p) => p.name !== name),
      {
        ...(value.properties.find((p) => p.name === name) ?? {
          name: nextName,
          type: nextType,
        }),
        name: nextName,
        type: nextType,
      },
    ]

    const nextValues: NodePropertyValue[] = [
      ...value.properties_values.filter((pv) => pv.name !== name),
      {
        ...newValue,
        name: nextName,
        type: nextType,
      },
    ]

    onChange({
      properties: nextProperties,
      properties_values: nextValues,
    })
  }

  const formattedValue = useMemo(() => {
    const v = propertyValue?.value
    if (typeof v === 'boolean') return v ? 'true' : 'false'
    if (v === null || v === undefined) return null
    return String(v)
  }, [propertyValue?.value])

  return (
    <>
      <div
        className={cn(
          'flex items-center justify-between rounded px-4 py-3',
          styles.property,
        )}
      >
        {withWrapper(
          <span className='text-title font-medium mr-auto'>{name}</span>,
        )}
        {withWrapper(<PropertyType type={type} />)}
        {withWrapper(
          <span className='mx-auto'>
            {formattedValue !== null
              ? `属性值：${formattedValue}`
              : '属性值：无'}
          </span>,
        )}
        {withWrapper(
          <span
            className={cn(
              'ml-auto flex gap-2 items-center text-[#616373]',
              styles.operator,
            )}
          >
            {disabled ? null : (
              <>
                <EditIcon onClick={toggle} className='cursor-pointer' />
                <Delete1Icon
                  onClick={handleDelete}
                  className='cursor-pointer'
                />
              </>
            )}
          </span>,
        )}
      </div>
      {open && !disabled ? (
        <PropertyValueModal
          title='编辑属性'
          onCancel={toggle}
          originalName={name}
          suggestedPropertyNameOptions={suggestedPropertyNameOptions}
          initialValue={propertyValue}
          onSubmit={(val) => {
            handleEdit(val)
            toggle()
          }}
        />
      ) : null}
    </>
  )
}

const PropertyType: FC<Style & { type?: Property['type'] }> = memo((props) => {
  const { type, className, style = {} } = props
  const [color, backgroundColor] = useMemo(() => {
    if (!type) return ['#616373', '#EFF1F4']
    switch (type) {
      case 'string':
        return ['#1990FF', '#DDF1FF']
      case 'int64':
        return ['#FF6600', '#FFEFD2']
      case 'double':
        return ['#D24BD2', '#FFECFF']
      case 'bool':
        return ['#266EFF', '#EBF2FF']
    }
  }, [type])
  return (
    <Tag
      className={cn('border-none rounded-full', className)}
      style={{
        color,
        backgroundColor,
        ...style,
      }}
    >
      {type ? PropertyLabelMap[type] : '自定义'}
    </Tag>
  )
})

const PropertyValueModal: FC<{
  title: string
  onCancel: () => void
  originalName: string
  suggestedPropertyNameOptions: { label: string; value: string }[]
  initialValue?: NodePropertyValue
  onSubmit: (val: NodePropertyValue) => void
}> = (props) => {
  const {
    title,
    onCancel,
    originalName,
    suggestedPropertyNameOptions,
    initialValue,
    onSubmit,
  } = props
  type EditFormValues = {
    propertyName: string
    propertyType: Property['type']
    value: null | string | number | boolean
  }
  const [form] = Form.useForm<EditFormValues>()
  const propertyType = Form.useWatch('propertyType', form) ?? 'string'
  const prevPropertyTypeRef = useRef<Property['type'] | null>(null)

  useEffect(() => {
    if (
      prevPropertyTypeRef.current &&
      prevPropertyTypeRef.current !== propertyType
    ) {
      form.setFieldValue('value', null)
    }
    prevPropertyTypeRef.current = propertyType
  }, [form, propertyType])

  const valueInput = useMemo(() => {
    switch (propertyType) {
      case 'string':
        return <Input />
      case 'int64':
        return <IntegerInput />
      case 'double':
        return <InputNumber className='w-full' controls={false} />
      case 'bool':
        return (
          <Select
            options={[
              { value: true, label: 'true' },
              { value: false, label: 'false' },
            ]}
          />
        )
      default:
        return <Input />
    }
  }, [propertyType])

  const handleSubmit = () => {
    form.validateFields().then((formValue) => {
      onSubmit({
        name: formValue.propertyName,
        value: formValue.value,
        type: formValue.propertyType,
      })
    })
  }

  return (
    <Modal title={title} open onCancel={onCancel} onOk={handleSubmit}>
      <Form
        form={form}
        initialValues={{
          propertyName: initialValue?.name ?? originalName,
          propertyType: initialValue?.type ?? 'string',
          value: initialValue?.value ?? null,
        }}
        layout='vertical'
      >
        <Form.Item
          name='propertyName'
          label='属性名称'
          rules={[{ required: true, message: '请输入属性名称' }]}
        >
          <AutoComplete
            options={suggestedPropertyNameOptions}
            placeholder='输入属性名称'
            filterOption={(input, option) =>
              (option?.value ?? '').toLowerCase().includes(input.toLowerCase())
            }
          >
            <Input maxLength={50} showCount />
          </AutoComplete>
        </Form.Item>
        <Form.Item
          name='propertyType'
          label='属性类型'
          rules={[{ required: true, message: '请选择属性类型' }]}
        >
          <Select
            options={[
              { label: PropertyLabelMap.string, value: 'string' },
              { label: PropertyLabelMap.int64, value: 'int64' },
              { label: PropertyLabelMap.double, value: 'double' },
              { label: PropertyLabelMap.bool, value: 'bool' },
            ]}
          />
        </Form.Item>
        <Form.Item
          name='value'
          label={`属性值 (${PropertyLabelMap[propertyType]})`}
          rules={[
            { required: true, message: '请输入属性值' },
            {
              validator: async (_, v: unknown) => {
                if (v === null || v === undefined) return Promise.resolve()
                if (propertyType === 'int64') {
                  if (typeof v !== 'number' || !Number.isInteger(v)) {
                    return Promise.reject(new Error('请输入整数'))
                  }
                }
                if (propertyType === 'double') {
                  if (typeof v !== 'number' || Number.isNaN(v)) {
                    return Promise.reject(new Error('请输入数字'))
                  }
                }
                if (propertyType === 'bool') {
                  if (typeof v !== 'boolean') {
                    return Promise.reject(new Error('请选择 true/false'))
                  }
                }
                if (propertyType === 'string') {
                  if (typeof v !== 'string') {
                    return Promise.reject(new Error('请输入字符串'))
                  }
                }
                return Promise.resolve()
              },
            },
          ]}
        >
          {valueInput}
        </Form.Item>
      </Form>
    </Modal>
  )
}

const CreatePropertyButton: FC<{
  value: NodePropertiesValue
  onChange: (value: NodePropertiesValue) => void
}> = (props) => {
  const { value, onChange } = props
  const [open, { toggle }] = useBoolean()

  return (
    <>
      <Button onClick={toggle}>
        <AddCircleIcon />
        新建属性
      </Button>
      {open ? (
        <CreatePropertyModal
          onCancel={toggle}
          properties={value.properties}
          existingValues={value.properties_values}
          onSubmit={(propertyValue) => {
            const nextProperties = value.properties.some(
              (p) => p.name === propertyValue.name,
            )
              ? value.properties.map((p) =>
                  p.name === propertyValue.name
                    ? { ...p, type: propertyValue.type ?? p.type ?? 'string' }
                    : p,
                )
              : [
                  ...value.properties,
                  {
                    name: propertyValue.name,
                    type: propertyValue.type ?? 'string',
                  },
                ]

            const nextPropertyValues = value.properties_values.some(
              (pv) => pv.name === propertyValue.name,
            )
              ? value.properties_values.map((pv) =>
                  pv.name === propertyValue.name ? propertyValue : pv,
                )
              : [...value.properties_values, propertyValue]

            onChange({
              properties: nextProperties,
              properties_values: nextPropertyValues,
            })
            toggle()
          }}
        />
      ) : null}
    </>
  )
}

const CreatePropertyModal: FC<{
  onCancel: () => void
  properties: Property[]
  existingValues: NodePropertyValue[]
  onSubmit: (propertyValue: NodePropertyValue) => void
}> = (props) => {
  const { message } = App.useApp()
  const { onCancel, properties, existingValues, onSubmit } = props
  type CreateFormValues = {
    propertyName: string
    propertyType: Property['type']
    value: null | string | number | boolean
  }
  const [form] = Form.useForm<CreateFormValues>()
  const options = useMemo(() => {
    return properties.map((p) => ({
      label: `${p.name}`,
      value: p.name,
    }))
  }, [properties])
  const propertyType = Form.useWatch('propertyType', form) ?? 'string'
  const prevPropertyTypeRef = useRef<Property['type'] | null>(null)

  useEffect(() => {
    if (
      prevPropertyTypeRef.current &&
      prevPropertyTypeRef.current !== propertyType
    ) {
      form.setFieldValue('value', null)
    }
    prevPropertyTypeRef.current = propertyType
  }, [form, propertyType])

  const valueInput = useMemo(() => {
    switch (propertyType) {
      case 'string':
        return <Input />
      case 'int64':
        return <IntegerInput />
      case 'double':
        return <InputNumber className='w-full' controls={false} />
      case 'bool':
        return (
          <Select
            options={[
              { value: true, label: 'true' },
              { value: false, label: 'false' },
            ]}
          />
        )
      default:
        return <Input />
    }
  }, [propertyType])

  const handleSubmit = () => {
    form.validateFields().then((formValue) => {
      const name = (formValue.propertyName ?? '').trim()
      if (!name) return
      const existed = existingValues.find((v) => v.name === name)
      // 仅当已存在且当前属性已经有 string/number/boolean 值时才提示重复
      if (
        existed &&
        (typeof existed.value === 'string' ||
          typeof existed.value === 'number' ||
          typeof existed.value === 'boolean')
      ) {
        message.warning('该属性已存在')
        return
      }
      onSubmit({
        name,
        value: formValue.value,
        type: formValue.propertyType,
      })
    })
  }

  return (
    <Modal title='新建属性' open onCancel={onCancel} onOk={handleSubmit}>
      <Form form={form} layout='vertical'>
        <Form.Item
          name='propertyName'
          label='属性名称'
          rules={[{ required: true, message: '请输入属性名称' }]}
        >
          <AutoComplete
            options={options}
            placeholder='输入属性名称'
            filterOption={(input, option) =>
              (option?.value ?? '').toLowerCase().includes(input.toLowerCase())
            }
          >
            <Input maxLength={50} showCount />
          </AutoComplete>
        </Form.Item>
        <Form.Item
          name='propertyType'
          label='属性类型'
          rules={[{ required: true, message: '请选择属性类型' }]}
          initialValue='string'
        >
          <Select
            options={[
              { label: PropertyLabelMap.string, value: 'string' },
              { label: PropertyLabelMap.int64, value: 'int64' },
              { label: PropertyLabelMap.double, value: 'double' },
              { label: PropertyLabelMap.bool, value: 'bool' },
            ]}
          />
        </Form.Item>
        <Form.Item
          name='value'
          label='属性值'
          rules={[{ required: true, message: '请输入属性值' }]}
        >
          {valueInput}
        </Form.Item>
      </Form>
    </Modal>
  )
}

/** 整数输入框 */
const IntegerInput: FC<ValueController<number>> = (props) => {
  const { value, onChange } = props
  return (
    <InputNumber
      controls={false}
      className='w-full'
      value={value}
      onChange={(val) => {
        if (val === null) {
          onChange?.(undefined)
          return
        }
        if (Number.isInteger(val)) onChange?.(val)
      }}
    />
  )
}
