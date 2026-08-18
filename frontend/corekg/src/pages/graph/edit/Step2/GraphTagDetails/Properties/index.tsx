import { FC, memo, useMemo } from 'react'
import { App, Empty, Form, Input, InputNumber, Modal, Select, Tag } from 'antd'
import type { Property } from 'Graph'
import { PropertyLabelMap } from 'Graph'
import { useBoolean, useRequest } from 'ahooks'
import { produce } from 'immer'
import { Delete1Icon, EditIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { updateTag } from '@/api/graph'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import { useActiveTag } from '../..'
import EmptyPic from '../EmptyPic.svg?react'
import styles from './styles.module.scss'

export const Properties: FC = () => {
  const { activeTag } = useActiveTag()
  if (!activeTag) return null
  const { properties } = activeTag
  return (
    <div className='flex-1 mt-2 overflow-auto flex flex-col gap-7'>
      {!properties || properties.length === 0 ? (
        <Empty
          image={<EmptyPic className='mx-auto' />}
          description='该实体类型暂无属性定义'
        />
      ) : (
        properties.map((p, i) => (
          <Property key={p.name} property={p} index={i} />
        ))
      )}
    </div>
  )
}

const Property: FC<{ property: Property; index: number }> = (props) => {
  const { modal, message } = App.useApp()
  const { property, index } = props
  const { name, type, defaults } = property
  const { data, mutateTags } = useGraphInfo()

  const activeTag = useActiveTag().activeTag!
  const { tag_name } = activeTag
  const [open, { toggle }] = useBoolean()
  const withWrapper = (children: any) => {
    return (
      <div className='flex-1 flex items-center justify-center'>{children}</div>
    )
  }
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
            默认值：
            {defaults !== undefined && defaults !== null ? `${defaults}` : '无'}
          </span>,
        )}
        {withWrapper(
          <span
            className={cn(
              'ml-auto flex gap-2 items-center text-[#616373]',
              styles.operator,
            )}
          >
            <EditIcon onClick={toggle} className='cursor-pointer' />
            <Delete1Icon
              className='cursor-pointer'
              onClick={() => {
                modal.confirm({
                  title: '确定删除？',
                  onOk: async () => {
                    const newProperties = activeTag.properties!.filter(
                      (_, i) => i !== index,
                    )
                    await updateTag({
                      graph_id: data!.id,
                      ...activeTag,
                      properties: newProperties,
                    })
                    message.success('删除成功')
                    mutateTags(
                      produce(data!.tags, (draft) => {
                        const currentTag = draft.find(
                          (item) => item.tag_name === tag_name,
                        )!
                        currentTag.properties = newProperties
                      }),
                    )
                  },
                })
              }}
            />
          </span>,
        )}
      </div>
      {open ? (
        <PropertyModal
          title='编辑属性'
          onCancel={toggle}
          initialValues={property}
          onSubmit={async (val) => {
            const newProperties = activeTag.properties!.map((p, i) =>
              i === index ? val : p,
            )
            await updateTag({
              graph_id: data!.id,
              ...activeTag,
              properties: newProperties,
            })
            message.success('修改成功')
            mutateTags(
              produce(data!.tags, (draft) => {
                const currentTag = draft.find(
                  (item) => item.tag_name === tag_name,
                )!
                currentTag.properties = newProperties
              }),
            )
          }}
        />
      ) : null}
    </>
  )
}

const PropertyType: FC<Style & { type: Property['type'] }> = memo((props) => {
  const { type, className, style = {} } = props
  const [color, backgroundColor] = useMemo(() => {
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
      {PropertyLabelMap[type]}
    </Tag>
  )
})

export const PropertyModal: FC<{
  title: string
  onCancel: () => void
  onSubmit: (val: Property) => any
  initialValues?: Property
}> = (props) => {
  const { title, onCancel, onSubmit, initialValues } = props
  const { activeTag } = useActiveTag()
  const FormItem = Form.Item<Property>
  const [form] = Form.useForm<Property>()
  const { run: submit, loading } = useRequest(
    async () => {
      const formValue = await form.validateFields()
      await onSubmit(formValue)
      onCancel()
    },
    { manual: true },
  )

  const type = Form.useWatch('type', form)
  // 收集默认值
  const defaultsInput = useMemo(() => {
    switch (type) {
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
        return null
    }
  }, [type])

  const selectOptions = useMemo(() => {
    const options = Object.entries(PropertyLabelMap).map(([value, label]) => {
      return {
        label,
        value,
      }
    })
    return options
  }, [])

  return (
    <Modal
      title={title}
      open
      onCancel={onCancel}
      keyboard={!loading}
      maskClosable={!loading}
      closable={!loading}
      cancelButtonProps={{ disabled: loading }}
      okButtonProps={{ loading }}
      onOk={submit}
    >
      <Form
        form={form}
        initialValues={initialValues}
        layout='vertical'
        onValuesChange={(changedFields) => {
          if ('type' in changedFields) {
            form.setFieldValue('defaults', undefined)
          }
        }}
      >
        <FormItem
          name='name'
          label='属性名称'
          rules={[
            { required: true, message: '请填写属性名称' },
            {
              validator: async (_, v) => {
                // 允许名称不变
                if (v === initialValues?.name) return
                // 同一类型 不能和其他属性同名
                if (activeTag?.properties?.some((item) => item.name === v)) {
                  throw new Error('当前实体类型已存在同名属性')
                }
              },
            },
          ]}
        >
          <Input maxLength={20} showCount />
        </FormItem>
        <FormItem
          name='type'
          label='数据类型'
          rules={[{ required: true, message: '请选择数据类型' }]}
        >
          <Select options={selectOptions} />
        </FormItem>
        {type ? (
          <FormItem name='defaults' label='默认值'>
            {defaultsInput}
          </FormItem>
        ) : null}
        <FormItem name='comment' label='描述'>
          <Input.TextArea maxLength={100} showCount />
        </FormItem>
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
