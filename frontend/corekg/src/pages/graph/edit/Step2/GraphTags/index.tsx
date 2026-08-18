import { FC } from 'react'
import { App, Button, Empty, Form, Input, message, Modal } from 'antd'
import { GraphTag } from 'Graph'
import { useBoolean, useRequest } from 'ahooks'
import { produce } from 'immer'
import { AddCircleIcon, Delete1Icon, EditIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { createTag, deleteTag, updateTag } from '@/api/graph'
import { GraphTagWithId, useGraphInfo } from '@/pages/graph/GraphProvider'
import { useActiveTag } from '..'
import styles from './styles.module.scss'

export const GraphTags: FC<Style> = (props) => {
  const { activeTag, setActiveTagId } = useActiveTag()
  const { className, style } = props
  const { modal, message } = App.useApp()
  const { data, mutateTags } = useGraphInfo()
  const { tags } = data!
  const delTag = async (tag_id: number) => {
    await deleteTag({
      graph_id: data!.id,
      tag_id,
    })
    message.success('删除成功')
    mutateTags(tags.filter((item) => item.tag_id !== tag_id))
  }

  const [open, { toggle }] = useBoolean()
  return (
    <>
      <div
        className={cn(
          'rounded-xl border border-[#D7D9E5] bg-[#FCFCFE]',
          'overflow-hidden p-4 flex flex-col',
          className,
        )}
        style={style}
      >
        <div className='flex justify-between items-center'>
          <span className='text-title font-medium text-lg'>实体类型</span>
          <Button
            onClick={toggle}
            icon={<AddCircleIcon />}
            className='text-[#0C99FF] border-[#0C99FF]'
          >
            新建类型
          </Button>
        </div>
        <div className='flex flex-col gap-2 mt-2 overflow-auto'>
          {tags.length === 0 ? (
            <Empty />
          ) : (
            tags.map((tag) => {
              const { tag_id } = tag
              return (
                <GraphTagItem
                  key={tag_id}
                  tag={tag}
                  active={tag_id === activeTag?.tag_id}
                  onEdit={() => {}}
                  onDel={() => {
                    modal.confirm({
                      title: '确认删除',
                      content:
                        '删除后，该实体类型下包含的实体、属性和关联关系将全部清空，请谨慎操作。',
                      onOk: () => delTag(tag_id),
                    })
                  }}
                  onClick={() => setActiveTagId(tag.tag_id)}
                />
              )
            })
          )}
        </div>
      </div>
      {open ? (
        <GraphTagModal
          onCancel={toggle}
          onSubmit={async (val) => {
            const { ID } = await createTag({ graph_id: data!.id, ...val })
            message.success('添加成功')
            mutateTags(tags.concat({ tag_id: ID, ...val }))
          }}
        />
      ) : null}
    </>
  )
}

const GraphTagItem: FC<{
  tag: GraphTagWithId
  onEdit: () => void
  onDel: () => void
  active?: boolean
  onClick: () => void
}> = (props) => {
  const {
    tag,

    onDel,
    active,
    onClick,
  } = props
  const { data, mutateTags } = useGraphInfo()
  const [open, { toggle }] = useBoolean()
  return (
    <>
      <div
        className={cn(
          'rounded ',
          {
            'bg-[#f8f9fd]': active,
          },
          'px-4 py-3 flex items-baseline',
          'cursor-pointer',
          styles.item,
        )}
        onClick={onClick}
      >
        <div className='bg-[#7990F8] rounded-full w-3 h-3 mr-2'></div>
        <div className='flex flex-col gap-1'>
          <span className='text-base font-medium'>{tag.tag_name}</span>
          <span className='text-description'>{tag.description}</span>
        </div>
        <div
          className={cn(
            'ml-auto text-[#616373] flex items-center gap-2',
            styles.operator,
          )}
        >
          <EditIcon
            className='cursor-pointer'
            onClick={(e) => {
              e.stopPropagation()
              toggle()
            }}
          />
          <Delete1Icon
            className='cursor-pointer'
            onClick={(e) => {
              e.stopPropagation()
              onDel()
            }}
          />
        </div>
      </div>
      {open ? (
        <GraphTagModal
          onCancel={toggle}
          initialValues={tag}
          onSubmit={async (val) => {
            await updateTag({
              graph_id: data!.id,
              ...tag,
              ...val,
            })
            message.success('修改成功')
            mutateTags(
              produce(data!.tags, (draft) => {
                const index = draft.findIndex(
                  (item) => item.tag_id === tag.tag_id,
                )
                draft[index] = {
                  ...draft[index],
                  ...val,
                }
              }),
            )
          }}
        />
      ) : null}
    </>
  )
}

type GraphTagInfo = Pick<GraphTag, 'tag_name' | 'description'>
const FormItem = Form.Item<GraphTagInfo>
const GraphTagModal: FC<{
  onCancel: () => void
  onSubmit: (info: GraphTagInfo) => any
  initialValues?: GraphTagInfo
}> = (props) => {
  const { onCancel, onSubmit, initialValues } = props
  const { data } = useGraphInfo()
  const { tags } = data!
  const [form] = Form.useForm<GraphTag>()
  const { loading, run: submit } = useRequest(
    async () => {
      const formValue = await form.validateFields()
      await onSubmit(formValue)
      onCancel()
    },
    { manual: true },
  )
  return (
    <Modal
      open
      title='新建类型'
      onCancel={onCancel}
      keyboard={!loading}
      maskClosable={!loading}
      closable={!loading}
      cancelButtonProps={{ disabled: loading }}
      okButtonProps={{ loading }}
      onOk={submit}
    >
      <Form form={form} layout='vertical' initialValues={initialValues}>
        <FormItem
          name='tag_name'
          label='类型名称'
          rules={[
            { required: true, message: '请填写类型名称' },
            {
              validator: async (_, v) => {
                if (tags.some((t) => t.tag_name === v)) {
                  throw new Error('实体类型不能重复')
                }
              },
            },
          ]}
        >
          <Input maxLength={20} showCount />
        </FormItem>
        <FormItem name='description' label='类型描述'>
          <Input.TextArea maxLength={100} showCount />
        </FormItem>
      </Form>
    </Modal>
  )
}
