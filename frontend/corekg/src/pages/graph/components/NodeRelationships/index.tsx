import { FC, useMemo, useState } from 'react'
import { App, Button, Empty, Form, Modal, Select } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import { ChevronRightIcon, Delete1Icon, EditIcon } from 'tdesign-icons-react'
import { AddCircleIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { listGraphNode } from '@/api/graph'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import EmptyPic from '../../edit/Step2/GraphTagDetails/EmptyPic.svg?react'
import { RelationshipNameSelect } from '../RelationshipNameSelect'
import styles from './styles.module.scss'

export type NodeEdge = {
  dst_node_name: string
  edge_name: string
  src_node_name: string
  /** 目标实体类型 */
  dst_tag_id: number
  /** 起点实体类型 */
  src_tag_id: number
}

type NodeRelationshipsProps = Style & {
  value: NodeEdge[]
  onChange: (value: NodeEdge[]) => void
  srcNodeName: string // 当前实体的名称
  srcTagId?: number // 当前实体的类型（tag_id）
  /** 源实体可选类型（必须是当前源节点自身的全部 tag） */
  srcTagOptions?: { label: string; value: number }[]
  /** 只读模式：隐藏新建/编辑/删除等操作入口 */
  disabled?: boolean
}

export const NodeRelationships: FC<NodeRelationshipsProps> = (props) => {
  const {
    value,
    onChange,
    srcNodeName,
    srcTagId,
    srcTagOptions,
    disabled,
    className,
    style,
  } = props
  const { data } = useGraphInfo()
  const { edgeOptions = [] } = data ?? {}

  return (
    <div className={cn('flex-1 flex flex-col gap-3', className)} style={style}>
      {/* 右上角按钮（与 Step2 同款位置） */}
      {disabled ? null : (
        <div className='flex justify-end'>
          <CreateEdgeButton
            srcNodeName={srcNodeName}
            srcTagId={srcTagId}
            srcTagOptions={srcTagOptions}
            value={value}
            onChange={onChange}
            edgeOptions={edgeOptions}
          />
        </div>
      )}

      <div className='flex-1 overflow-auto flex flex-col gap-2'>
        {value.length === 0 ? (
          <Empty
            image={<EmptyPic className='mx-auto' />}
            description='该实体暂无关系'
          />
        ) : (
          value.map((edge, index) => (
            <EdgeItem
              key={`${edge.src_node_name}_${edge.dst_node_name}_${edge.edge_name}_${index}`}
              edge={edge}
              index={index}
              srcNodeName={srcNodeName}
              value={value}
              onChange={onChange}
              edgeOptions={edgeOptions}
              srcTagOptions={srcTagOptions}
              disabled={disabled}
            />
          ))
        )}
      </div>
    </div>
  )
}

const EdgeItem: FC<{
  edge: NodeEdge
  index: number
  srcNodeName: string
  value: NodeEdge[]
  onChange: (value: NodeEdge[]) => void
  edgeOptions: string[]
  srcTagOptions?: { label: string; value: number }[]
  disabled?: boolean
}> = (props) => {
  const { modal } = App.useApp()
  const {
    edge,
    index,
    srcNodeName,
    value,
    onChange,
    edgeOptions,
    srcTagOptions,
    disabled,
  } = props
  const { dst_node_name, edge_name } = edge
  const [open, { toggle }] = useBoolean()

  const handleDelete = () => {
    modal.confirm({
      title: '确定删除？',
      onOk: () => {
        const newValue = value.filter((_, i) => i !== index)
        onChange(newValue)
      },
    })
  }

  const handleEdit = (newEdge: NodeEdge) => {
    const newEdges = value.map((e, i) => (i === index ? newEdge : e))
    onChange(newEdges)
  }

  return (
    <>
      <div
        className={cn(
          'h-11 relative flex gap-3 items-center justify-center',
          styles.relationship,
        )}
      >
        <span className='text-[#7445E0] font-medium'>{srcNodeName}</span>
        <span className='text-[#A895FC]'>
          ------
          <ChevronRightIcon />
        </span>
        <span className='bg-[#DDF1FF] text-[#0C99FF] rounded px-1.5 py-0.5'>
          {edge_name}
        </span>
        <span className='text-[#A895FC]'>
          ------
          <ChevronRightIcon />
        </span>
        <span className='text-[#7445E0] font-medium'>{dst_node_name}</span>
        <div
          className={cn(
            'absolute right-1 top-1/2 -translate-y-1/2',
            styles.operator,
          )}
        >
          {disabled ? null : (
            <>
              <EditIcon
                onClick={toggle}
                className='cursor-pointer mr-2.5 text-[#616373]'
              />
              <Delete1Icon
                onClick={handleDelete}
                className='cursor-pointer text-[#616373]'
              />
            </>
          )}
        </div>
      </div>
      {open && !disabled ? (
        <EdgeModal
          title='编辑关系'
          onCancel={toggle}
          srcNodeName={srcNodeName}
          initialValues={edge}
          edgeOptions={edgeOptions}
          srcTagOptions={srcTagOptions}
          onSubmit={(val) => {
            handleEdit(val)
            toggle()
          }}
        />
      ) : null}
    </>
  )
}

const CreateEdgeButton: FC<{
  srcNodeName: string
  srcTagId?: number
  srcTagOptions?: { label: string; value: number }[]
  value: NodeEdge[]
  onChange: (value: NodeEdge[]) => void
  edgeOptions: string[]
}> = (props) => {
  const { srcNodeName, srcTagId, srcTagOptions, value, onChange, edgeOptions } =
    props
  const [open, { toggle }] = useBoolean()

  return (
    <>
      <Button onClick={toggle}>
        <AddCircleIcon />
        新建关系
      </Button>
      {open ? (
        <EdgeModal
          title='新建关系'
          onCancel={toggle}
          srcNodeName={srcNodeName}
          srcTagId={srcTagId}
          edgeOptions={edgeOptions}
          srcTagOptions={srcTagOptions}
          onSubmit={(val) => {
            onChange([...value, val])
            toggle()
          }}
        />
      ) : null}
    </>
  )
}

type EdgeFormValues = {
  src_tag_id: number
  src_node_name: string
  dst_tag_id: number
  dst_node_name: string
  edge_name: string
}

const EdgeModal: FC<{
  title: string
  onCancel: () => void
  srcNodeName: string
  srcTagId?: number
  initialValues?: NodeEdge
  edgeOptions: string[]
  srcTagOptions?: { label: string; value: number }[]
  onSubmit: (val: NodeEdge) => void
}> = (props) => {
  const {
    title,
    onCancel,
    srcNodeName,
    srcTagId,
    initialValues,
    edgeOptions,
    srcTagOptions,
    onSubmit,
  } = props
  const { data } = useGraphInfo()
  const { id: graph_id, tags = [] } = data ?? {}
  const FormItem = Form.Item<EdgeFormValues>
  const [form] = Form.useForm<EdgeFormValues>()

  const isEdit = !!initialValues

  // 搜索目标实体
  const [dstNodeSearch, setDstNodeSearch] = useState('')
  const dstTagId = Form.useWatch('dst_tag_id', form)
  const { data: nodeOptions } = useRequest(
    async () => {
      if (!graph_id || !dstTagId || !dstNodeSearch) return []
      const { list } = await listGraphNode({
        graph_id,
        graph_tag_id: dstTagId,
        graph_node_name: dstNodeSearch,
      })
      return (list as any[]).map((item) => ({
        label: item.graph_node_name as string,
        value: item.graph_node_name as string,
      }))
    },
    { refreshDeps: [dstNodeSearch, dstTagId], debounceWait: 500 },
  )

  const handleSubmit = () => {
    form.validateFields().then((formValue) => {
      onSubmit({
        src_node_name: formValue.src_node_name,
        dst_node_name: formValue.dst_node_name,
        edge_name: formValue.edge_name,
        dst_tag_id: formValue.dst_tag_id,
        src_tag_id: formValue.src_tag_id,
      })
    })
  }

  const edgeNameOptions = useMemo(() => {
    return edgeOptions.map((opt) => ({ label: opt, value: opt }))
  }, [edgeOptions])

  const srcTagSelectOptions = useMemo(() => {
    if (srcTagOptions && srcTagOptions.length > 0) return srcTagOptions
    return tags.map((t) => ({ label: t.tag_name, value: t.tag_id }))
  }, [srcTagOptions, tags])

  const dstNodeNameOptions = useMemo(() => {
    if (!isEdit) return []
    const dstName = initialValues?.dst_node_name
    if (!dstName) return []
    return [{ label: dstName, value: dstName }]
  }, [initialValues?.dst_node_name, isEdit])

  return (
    <Modal title={title} open onCancel={onCancel} onOk={handleSubmit}>
      <Form
        form={form}
        initialValues={{
          ...initialValues,
          src_tag_id:
            initialValues?.src_tag_id ??
            srcTagId ??
            srcTagSelectOptions[0]?.value,
          src_node_name: srcNodeName,
          dst_tag_id: initialValues?.dst_tag_id,
        }}
        layout='vertical'
      >
        <FormItem
          name='src_tag_id'
          label='源实体类型'
          rules={[{ required: true, message: '请选择源实体类型' }]}
        >
          <Select
            placeholder='请选择源实体类型'
            options={srcTagSelectOptions}
            showSearch
            optionFilterProp='label'
            disabled={true}
          />
        </FormItem>
        <FormItem
          name='src_node_name'
          label='源实体'
          rules={[{ required: true, message: '请选择源实体' }]}
        >
          <Select
            disabled
            placeholder='源实体'
            options={[{ label: srcNodeName, value: srcNodeName }]}
          />
        </FormItem>
        <FormItem
          name='dst_tag_id'
          label='目标实体类型'
          rules={[{ required: true, message: '请选择目标实体类型' }]}
        >
          <Select
            placeholder='请选择目标实体类型'
            options={tags.map((t) => ({ label: t.tag_name, value: t.tag_id }))}
            showSearch
            optionFilterProp='label'
            disabled={isEdit}
          />
        </FormItem>
        <FormItem
          name='dst_node_name'
          label='目标实体'
          rules={[{ required: true, message: '请选择或输入目标实体' }]}
        >
          <Select
            disabled={isEdit}
            showSearch={!isEdit}
            placeholder={isEdit ? '目标实体' : '搜索并选择目标实体'}
            options={isEdit ? dstNodeNameOptions : nodeOptions}
            onSearch={isEdit ? undefined : setDstNodeSearch}
            filterOption={isEdit ? true : false}
            notFoundContent={
              isEdit
                ? undefined
                : !dstTagId
                  ? '请先选择目标实体类型'
                  : dstNodeSearch
                    ? '没有找到实体，请继续输入'
                    : '请输入实体名称搜索'
            }
          />
        </FormItem>
        <FormItem
          name='edge_name'
          label='关系名称'
          rules={[{ required: true, message: '请选择关系名称' }]}
        >
          <RelationshipNameSelect
            options={edgeNameOptions}
            placeholder='请选择关系名称'
            maxLength={20}
          />
        </FormItem>
      </Form>
    </Modal>
  )
}
