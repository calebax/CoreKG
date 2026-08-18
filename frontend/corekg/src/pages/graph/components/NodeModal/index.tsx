import { FC, useEffect, useMemo, useState } from 'react'
import { App, Form, Input, Modal, Select, Tabs, Tooltip } from 'antd'
import { QuestionCircleOutlined } from '@ant-design/icons'
import type { Property } from 'Graph'
import { useRequest } from 'ahooks'
import { GRAPH_NODE_NAME_INVALID_TOOLTIP, validateGraphNodeName } from '@/utils'
import { createNode, editNode, getNodeEdges, renameNode } from '@/api/graph'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import type { NodePropertiesValue, NodePropertyValue } from '../NodeProperties'
import { NodeProperties } from '../NodeProperties'
import type { NodeEdge } from '../NodeRelationships'
import { NodeRelationships } from '../NodeRelationships'

export type NodeModalSuccessResult = {
  mode: 'create' | 'edit'
  node_name: string
  tag_id: number
  tag_name: string
  old_node_name?: string
  properties_values: NodePropertyValue[]
  edges: NodeEdge[]
  node_tags: {
    tag_id: number
    tag_name: string
    properties_values: NodePropertyValue[]
  }[]
}

type NodeTagProperties = {
  tag_id: number
  tag_name: string
  properties?: Property[]
  properties_values?: NodePropertyValue[]
}

type NodeModalProps = {
  open: boolean
  onCancel: () => void
  onSuccess?: (result: NodeModalSuccessResult) => void
  graphId: number
  initialValues?: {
    node_name: string
    tag_id?: number
    tag_name?: string
    /** 多 tag 的属性集合（对齐 NodeDetail：切换 tag 展示对应 properties_values） */
    properties_values?: NodeTagProperties[]
    edges?: NodeEdge[]
    available_tags?: { tag_id: number; tag_name: string }[]
  }
}

type FormValues = {
  node_name: string
  tag_id: number
}

export const NodeModal: FC<NodeModalProps> = (props) => {
  const { open, onCancel, onSuccess, graphId, initialValues } = props
  const { data } = useGraphInfo()
  const { tags = [] } = data ?? {}
  const { message } = App.useApp()
  const [form] = Form.useForm<FormValues>()
  const tag_id = Form.useWatch('tag_id', form)
  const node_name = Form.useWatch('node_name', form)
  const [activeTab, setActiveTab] = useState('info')
  const [tagPropertiesList, setTagPropertiesList] = useState<
    NodeTagProperties[]
  >([])
  const [edges, setEdges] = useState<NodeEdge[]>([])

  const isEdit = !!initialValues

  const { runAsync: fetchNodeEdges } = useRequest(
    async (nodeName: string, tagId: number) => {
      if (!graphId || !nodeName || !tagId) return []
      const { edges: fetchedEdges } = await getNodeEdges({
        graph_id: graphId,
        node_name: nodeName,
        tag_id: tagId,
      })
      return (fetchedEdges ?? []).map((e) => ({
        src_node_name: e.src_node_name,
        dst_node_name: e.dst_node_name,
        edge_name: e.edge_name,
        src_tag_id: e.src_tag_id,
        dst_tag_id: e.dst_tag_id,
      })) as NodeEdge[]
    },
    { manual: true },
  )

  // 初始化表单和状态
  useEffect(() => {
    if (!open) return
    if (initialValues) {
      const nodeTags = initialValues.properties_values ?? []
      const initialTagId = initialValues.tag_id ?? nodeTags[0]?.tag_id
      // 编辑模式（对齐 NodeDetail：支持切换 tag 展示对应 properties）
      form.setFieldsValue({
        node_name: initialValues.node_name,
        tag_id: initialTagId,
      })

      setTagPropertiesList(
        nodeTags.map((t) => {
          const tagDef = tags.find((d) => d.tag_id === t.tag_id)
          return {
            ...t,
            tag_name: t.tag_name ?? tagDef?.tag_name ?? '',
            properties: t.properties ?? tagDef?.properties ?? [],
            properties_values: t.properties_values ?? [],
          }
        }),
      )
      // 边以 getNodeEdges 为准（会取到该节点所有 tag 的边）
      setEdges([])
    } else {
      // 新建模式
      form.resetFields()
      setTagPropertiesList([])
      setEdges([])
      setActiveTab('info')
    }
  }, [form, initialValues, open, tags])

  // 编辑模式：打开弹窗后拉取该节点（作为起点）所有 tag 的边
  useEffect(() => {
    if (!open) return
    if (!isEdit) return
    const originNodeName = initialValues?.node_name
    if (!originNodeName) return
    if (!tag_id) return
    fetchNodeEdges(originNodeName, tag_id)
      .then((nextEdges) => {
        setEdges(nextEdges)
      })
      .catch(() => {
        message.error('获取实体关系失败')
      })
  }, [fetchNodeEdges, initialValues?.node_name, isEdit, message, open, tag_id])

  // 节点名称变化后同步更新本地 edges（当前节点始终作为起点）
  useEffect(() => {
    if (!open) return
    if (!node_name) return
    setEdges((prev) =>
      prev.map((e) =>
        e.src_node_name === node_name ? e : { ...e, src_node_name: node_name },
      ),
    )
  }, [node_name, open])

  const currentNodePropertiesValue = useMemo<NodePropertiesValue>(() => {
    if (!tag_id) {
      return { properties: [], properties_values: [] }
    }
    const existed = tagPropertiesList.find((t) => t.tag_id === tag_id)
    if (existed) {
      return {
        properties: existed.properties ?? [],
        properties_values: existed.properties_values ?? [],
      }
    }
    const selectedTag = tags.find((t) => t.tag_id === tag_id)
    return { properties: selectedTag?.properties ?? [], properties_values: [] }
  }, [tagPropertiesList, tag_id, tags])

  const srcTagOptions = useMemo(() => {
    if (isEdit) {
      return (tagPropertiesList ?? []).map((t) => ({
        label: t.tag_name,
        value: t.tag_id,
      }))
    }
    if (!tag_id) return []
    const tag = tags.find((t) => t.tag_id === tag_id)
    if (!tag) return []
    return [{ label: tag.tag_name, value: tag.tag_id }]
  }, [isEdit, tagPropertiesList, tag_id, tags])

  const { run: submit, loading } = useRequest(
    async () => {
      const formValues = await form.validateFields()

      const selectedTag = tags.find((t) => t.tag_id === tag_id)
      if (!selectedTag) {
        message.error('实体类型不存在')
        return
      }

      if (!tag_id) {
        message.error('请选择实体类型')
        return
      }

      // 构建 tags 数据（编辑模式：全量提交该节点已有的 tags；新建模式：仅提交当前 tag）
      const currentTagState = tagPropertiesList.find((t) => t.tag_id === tag_id)
      const currentTagDef = tags.find((t) => t.tag_id === tag_id)
      if (!currentTagDef) {
        message.error('实体类型不存在')
        return
      }

      const currentTagPropertiesValues =
        currentTagState?.properties_values ??
        currentNodePropertiesValue.properties_values

      const nodeTags = [
        {
          tag_id: currentTagDef.tag_id,
          tag_name: currentTagDef.tag_name,
          properties: currentNodePropertiesValue.properties ?? [],
          properties_values: currentTagPropertiesValues.map((pv) => ({
            name: pv.name,
            value: pv.value,
          })),
        },
      ]

      const edgesToSubmit = edges

      if (edgesToSubmit.some((e) => !e.src_tag_id)) {
        message.error('请为每条关系选择源实体类型')
        return
      }

      if (edgesToSubmit.some((e) => !e.dst_tag_id)) {
        message.error('请为每条关系选择目标实体类型')
        return
      }

      if (isEdit && initialValues) {
        // 编辑实体
        if (formValues.node_name !== initialValues.node_name) {
          await renameNode({
            graph_id: graphId,
            node_name: formValues.node_name,
            old_node_name: initialValues.node_name,
            tag_id,
          })
        }

        await editNode({
          graph_id: graphId,
          // 不负责重命名 需要传新名字
          old_node_name: formValues.node_name,
          tags: nodeTags,
          edges:
            edgesToSubmit.length > 0
              ? edgesToSubmit.map((e) => ({
                  src_node_name: e.src_node_name,
                  dst_node_name: e.dst_node_name,
                  edge_name: e.edge_name,
                  src_tag_id: e.src_tag_id,
                  dst_tag_id: e.dst_tag_id!,
                }))
              : undefined,
        })
        message.success('编辑实体成功')
      } else {
        // 创建实体
        await createNode({
          graph_id: graphId,
          node_name: formValues.node_name,
          tags: nodeTags,
          edges:
            edgesToSubmit.length > 0
              ? edgesToSubmit.map((e) => ({
                  src_node_name: e.src_node_name,
                  dst_node_name: e.dst_node_name,
                  edge_name: e.edge_name,
                  src_tag_id: e.src_tag_id,
                  dst_tag_id: e.dst_tag_id!,
                }))
              : undefined,
        })
        message.success('创建实体成功')
      }

      onSuccess?.({
        mode: isEdit ? 'edit' : 'create',
        node_name: formValues.node_name,
        tag_id: selectedTag.tag_id,
        tag_name: selectedTag.tag_name,
        old_node_name: initialValues?.node_name,
        properties_values: currentNodePropertiesValue.properties_values,
        edges: edgesToSubmit,
        node_tags:
          isEdit && tagPropertiesList.length
            ? tagPropertiesList.map((t) => ({
                tag_id: t.tag_id,
                tag_name: t.tag_name,
                properties_values: t.properties_values ?? [],
              }))
            : [
                {
                  tag_id: selectedTag.tag_id,
                  tag_name: selectedTag.tag_name,
                  properties_values:
                    currentNodePropertiesValue.properties_values,
                },
              ],
      })
      onCancel()
    },
    { manual: true },
  )

  const tagOptions = useMemo(() => {
    if (isEdit) {
      const nodeTags = initialValues?.properties_values ?? []
      if (nodeTags.length > 0) {
        return nodeTags.map((tag) => ({
          label: tag.tag_name,
          value: tag.tag_id,
        }))
      }
      return (initialValues?.available_tags ?? []).map((tag) => ({
        label: tag.tag_name,
        value: tag.tag_id,
      }))
    }
    return tags.map((tag) => ({
      label: tag.tag_name,
      value: tag.tag_id,
    }))
  }, [
    initialValues?.available_tags,
    initialValues?.properties_values,
    isEdit,
    tags,
  ])

  return (
    <Modal
      title={isEdit ? '编辑实体' : '新建实体'}
      open={open}
      onCancel={onCancel}
      onOk={submit}
      okButtonProps={{ loading }}
      cancelButtonProps={{ disabled: loading }}
      width={720}
      destroyOnClose
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'info',
            label: '实体信息',
            children: (
              <Form form={form} layout='vertical' className='mt-4'>
                <Form.Item
                  name='tag_id'
                  label='实体类型'
                  rules={[{ required: true, message: '请选择实体类型' }]}
                >
                  <Select placeholder='通用名' options={tagOptions} />
                </Form.Item>
                <Form.Item
                  name='node_name'
                  label={
                    <div className='flex items-center gap-2'>
                      <span>实体名称</span>
                      <Tooltip title={GRAPH_NODE_NAME_INVALID_TOOLTIP}>
                        <QuestionCircleOutlined className='text-[#9AA0A6]' />
                      </Tooltip>
                    </div>
                  }
                  rules={[
                    { required: true, message: '请输入实体名称' },
                    { validator: validateGraphNodeName },
                  ]}
                >
                  <Input
                    placeholder='请输入实体名称'
                    maxLength={50}
                    showCount
                  />
                </Form.Item>
              </Form>
            ),
          },
          {
            key: 'properties',
            label: '实体属性',
            disabled: !tag_id,
            children: (
              <div className='mt-4'>
                <NodeProperties
                  value={currentNodePropertiesValue}
                  onChange={(next) => {
                    setTagPropertiesList((prev) => {
                      if (!tag_id) return prev
                      const exists = prev.some((t) => t.tag_id === tag_id)
                      if (!exists) {
                        const tagName =
                          prev.find((t) => t.tag_id === tag_id)?.tag_name ??
                          tags.find((t) => t.tag_id === tag_id)?.tag_name ??
                          ''
                        return [
                          ...prev,
                          {
                            tag_id,
                            tag_name: tagName,
                            properties: next.properties,
                            properties_values: next.properties_values,
                          },
                        ]
                      }
                      return prev.map((t) =>
                        t.tag_id === tag_id
                          ? {
                              ...t,
                              properties: next.properties,
                              properties_values: next.properties_values,
                            }
                          : t,
                      )
                    })
                  }}
                  tagId={tag_id}
                />
              </div>
            ),
          },
          {
            key: 'relationships',
            label: '实体关系',
            disabled: !node_name || !tag_id,
            children: (
              <div className='mt-4'>
                <NodeRelationships
                  value={edges}
                  onChange={setEdges}
                  srcNodeName={node_name}
                  srcTagId={tag_id}
                  srcTagOptions={srcTagOptions}
                />
              </div>
            ),
          },
        ]}
      />
    </Modal>
  )
}
