import { FC } from 'react'
import { App, Empty, Form, Input, Modal, Select, Skeleton } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import { ChevronRightIcon, Delete1Icon, EditIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { deleteEdge, getTagEdge, updateEdge } from '@/api/graph'
import { GraphTagWithId, useGraphInfo } from '@/pages/graph/GraphProvider'
import { useActiveTag } from '../..'
import EmptyPic from '../EmptyPic.svg?react'
import styles from './styles.module.scss'

// 来自后端的边结构
type BackendEdge = {
  id: number
  src_tag_id: number
  dst_tag_id: number
  edge_name: string
}

export const Relationships: FC = () => {
  const { data } = useGraphInfo()
  const { activeTag } = useActiveTag()
  const {
    data: edges,
    loading,
    run,
  } = useRequest(
    async () => {
      if (!activeTag) return
      const res = await getTagEdge({
        graph_id: data!.id,
        tag_id: activeTag!.tag_id,
      })
      return res.data as BackendEdge[]
    },
    { refreshDeps: [activeTag] },
  )
  if (loading || !edges) return <Skeleton active />
  return (
    <div className='flex-1 mt-2 overflow-auto flex flex-col gap-2'>
      {edges.length === 0 ? (
        <Empty
          image={<EmptyPic className='mx-auto' />}
          description='该实体类型暂无关系定义'
        />
      ) : (
        edges.map((e) => <Edge key={e.id} edge={e} reload={run} />)
      )}
    </div>
  )
}

const Edge: FC<{
  edge: BackendEdge
  reload: () => void
}> = (props) => {
  const { data } = useGraphInfo()
  const { id: graph_id, tags } = data!
  const { modal, message } = App.useApp()
  const { edge, reload } = props
  const { id: edge_id, src_tag_id, dst_tag_id, edge_name: name } = edge
  const [open, { toggle }] = useBoolean()
  return (
    <>
      <div
        className={cn(
          'h-11 relative flex gap-3 items-center justify-center',
          styles.relationship,
        )}
      >
        <span className='text-[#7445E0] font-medium'>
          {getTagNameById(tags, src_tag_id)}
        </span>
        <span className='text-[#A895FC]'>
          ------
          <ChevronRightIcon />
        </span>
        <span className='bg-[#DDF1FF] text-[#0C99FF] rounded px-1.5 py-0.5'>
          {name}
        </span>
        <span className='text-[#A895FC]'>
          ------
          <ChevronRightIcon />
        </span>
        <span className='text-[#7445E0] font-medium'>
          {getTagNameById(tags, dst_tag_id)}
        </span>
        <div
          className={cn(
            'absolute right-1 top-1/2 -translate-y-1/2',
            styles.operator,
          )}
        >
          <EditIcon
            onClick={toggle}
            className='cursor-pointer mr-2.5 text-[#616373]'
          />
          <Delete1Icon
            onClick={() => {
              modal.confirm({
                title: '确定删除？',
                onOk: async () => {
                  await deleteEdge({
                    graph_id,
                    edge_id,
                  })
                  reload()
                  message.success('操作成功')
                },
              })
            }}
            className='cursor-pointer  text-[#616373]'
          />
        </div>
      </div>
      {open ? (
        <RelationshipModal
          title='编辑关系'
          onCancel={toggle}
          onSubmit={async (val) => {
            await updateEdge({
              graph_id,
              edge_id,
              ...val,
            })
            message.success('修改成功')
            reload()
          }}
          initialValues={edge}
        />
      ) : null}
    </>
  )
}

type EditEdgeType = Pick<BackendEdge, 'dst_tag_id' | 'src_tag_id' | 'edge_name'>
export const RelationshipModal: FC<{
  title: string
  onCancel: () => void
  onSubmit: (val: EditEdgeType) => any
  initialValues?: BackendEdge
}> = (props) => {
  const { title, onCancel, onSubmit, initialValues } = props
  const FormItem = Form.Item<EditEdgeType>
  const [form] = Form.useForm<EditEdgeType>()
  const { run: submit, loading } = useRequest(
    async () => {
      const formValue = await form.validateFields()
      await onSubmit(formValue)
      onCancel()
    },
    { manual: true },
  )
  const { data } = useGraphInfo()
  const { tags } = data!
  const TagSelector = useMemo(() => {
    const options = tags.map((item) => {
      return {
        value: item.tag_id,
        label: item.tag_name,
      }
    })
    return (
      <Select
        options={options}
        showSearch
        allowClear
        optionFilterProp='value'
      />
    )
  }, [tags])
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
      <Form form={form} initialValues={initialValues} layout='vertical'>
        <FormItem
          name='src_tag_id'
          label='源实体类型'
          rules={[{ required: true, message: '请选择源实体类型' }]}
        >
          {TagSelector}
        </FormItem>
        <FormItem
          name='dst_tag_id'
          label='目标实体类型'
          rules={[{ required: true, message: '请选择目标实体类型' }]}
        >
          {TagSelector}
        </FormItem>
        <FormItem
          name='edge_name'
          label='关系名称'
          rules={[{ required: true, message: '请填写关系名称' }]}
        >
          <Input maxLength={20} showCount disabled={Boolean(initialValues)} />
        </FormItem>
      </Form>
    </Modal>
  )
}

const getTagNameById = (tags: GraphTagWithId[], id: number) => {
  return tags.find((item) => item.tag_id === id)?.tag_name
}
