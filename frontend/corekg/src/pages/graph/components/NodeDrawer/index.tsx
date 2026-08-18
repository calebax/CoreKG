import { FC, useEffect, useMemo, useRef, useState } from 'react'
import {
  App,
  Button,
  Drawer,
  Form,
  Input,
  Select,
  Skeleton,
  Tabs,
  Tooltip,
} from 'antd'
import { QuestionCircleOutlined } from '@ant-design/icons'
import type { Property } from 'Graph'
import { useRequest } from 'ahooks'
import {
  GRAPH_NODE_NAME_INVALID_TOOLTIP,
  cn,
  validateGraphNodeName,
} from '@/utils'
import {
  editNode,
  getNodeEdges,
  getNodeReference,
  renameNode,
} from '@/api/graph'
import { getDocContent } from '@/api/knowledge'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import type { NodeModalSuccessResult } from '../NodeModal'
import type { NodePropertiesValue, NodePropertyValue } from '../NodeProperties'
import { NodeProperties } from '../NodeProperties'
import type { NodeEdge } from '../NodeRelationships'
import { NodeRelationships } from '../NodeRelationships'

type NodeTagProperties = {
  tag_id: number
  tag_name: string
  properties?: Property[]
  properties_values?: NodePropertyValue[]
}

type NodeDrawerInitialValues = {
  node_name: string
  tag_id?: number
  tag_name?: string
  /** 多 tag 的属性集合（对齐 NodeDetail：切换 tag 展示对应 properties_values） */
  properties_values?: NodeTagProperties[]
  edges?: NodeEdge[]
  available_tags?: { tag_id: number; tag_name: string }[]
}

type NodeDrawerProps = {
  open: boolean
  onCancel: () => void
  onSuccess?: (result: NodeModalSuccessResult) => void
  graphId: number
  /** 固定实体类型（不允许用户选择） */
  tag_id: number
  /** 仅编辑模式：必须传初始值，至少包含 node_name */
  initialValues: NodeDrawerInitialValues
  /** 只读模式：禁用编辑操作 */
  isReadOnly?: boolean
}

type NodeReferenceFile = {
  file_id: number
  file_name: string
}

type NodeReferenceResponse = {
  files: NodeReferenceFile[]
}

type FormValues = {
  node_name: string
  tag_id: number
}

export const NodeDrawer: FC<NodeDrawerProps> = (props) => {
  const { open, onCancel, graphId, tag_id, initialValues, isReadOnly } = props
  const [editing, setEditing] = useState(false)
  const { message } = App.useApp()
  const { data } = useGraphInfo()
  const { tags = [] } = data ?? {}
  const [tagPropertiesList, setTagPropertiesList] = useState<
    NodeTagProperties[]
  >([])

  // 初始化表单和状态
  useEffect(() => {
    if (!open) return
    if (initialValues) {
      const nodeTags = initialValues.properties_values ?? []

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
    } else {
      setTagPropertiesList([])
    }
  }, [initialValues, open, tags])

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

  // ============= 文件预览 =============
  const [referenceNodeName, setReferenceNodeName] = useState('')
  const [selectedFileId, setSelectedFileId] = useState<number | undefined>()
  const markdownCacheRef = useRef<Map<number, string>>(new Map())

  const {
    data: nodeReference,
    loading: loadingReference,
    run: loadNodeReference,
  } = useRequest(
    async () => {
      if (!graphId || !referenceNodeName || !tag_id) return null
      return (await getNodeReference({
        graph_id: graphId,
        node_name: referenceNodeName,
        tag_id,
      })) as NodeReferenceResponse
    },
    { manual: true },
  )

  const fileOptions = useMemo(() => {
    return (nodeReference?.files ?? []).map((f) => ({
      label: f.file_name,
      value: f.file_id,
    }))
  }, [nodeReference?.files])

  useEffect(() => {
    if (!open) return
    setReferenceNodeName(initialValues.node_name)
  }, [initialValues.node_name, open])

  useEffect(() => {
    if (!open) return
    if (!graphId || !referenceNodeName || !tag_id) return
    loadNodeReference()
  }, [graphId, loadNodeReference, open, referenceNodeName, tag_id])

  useEffect(() => {
    if (!open) return
    const nextFileIds = (nodeReference?.files ?? []).map((f) => f.file_id)
    if (nextFileIds.length === 0) {
      setSelectedFileId(undefined)
      return
    }
    setSelectedFileId((prev) => {
      if (prev && nextFileIds.includes(prev)) return prev
      return nextFileIds[0]
    })
  }, [nodeReference?.files, open])

  const {
    data: markdownData,
    loading: markdownLoading,
    error: markdownError,
  } = useRequest(
    async () => {
      if (!selectedFileId) return ''
      const cached = markdownCacheRef.current.get(selectedFileId)
      if (cached !== undefined) return cached
      const res: any = await getDocContent({ file_id: selectedFileId })
      const content = String(res?.content ?? '')
        .replace(/```markdown\s*|\s*```/g, '')
        .replace(/<!--yg_pos[\s\S]*?yg_pos-->/g, '')

      let nextContent = content

      const nodeNameText = String(initialValues.node_name ?? '')
      if (nodeNameText) {
        nextContent = nextContent
          .split(nodeNameText)
          .join(
            `<span style="background-color:#0C99FF80;">${nodeNameText}</span>`,
          )
      }

      ;(currentNodePropertiesValue?.properties_values ?? []).forEach((p) => {
        const valueText = String(p?.value ?? '')
        if (!valueText) return
        nextContent = nextContent
          .split(valueText)
          .join(`<span style="background-color:#CC5DE880;">${valueText}</span>`)
      })

      markdownCacheRef.current.set(selectedFileId, nextContent)
      return nextContent
    },
    {
      ready: open && !!selectedFileId,
      refreshDeps: [selectedFileId, open],
    },
  )

  useEffect(() => {
    if (!open) return
    if (!markdownError) return
    message.error('获取文件内容失败')
  }, [markdownError, message, open])

  return (
    <Drawer
      title='原文对照'
      open={open}
      onClose={onCancel}
      width={1100}
      destroyOnClose
    >
      {loadingReference ? (
        <Skeleton active className='p-4' />
      ) : (
        <div className='flex h-[calc(100vh-160px)] gap-4'>
          {/* 左侧：Markdown 预览 */}
          <div className='w-[520px] flex flex-col gap-3 border-r border-[#E5E7EB] pr-4'>
            <div className='flex items-center justify-between gap-3'>
              <div className='text-[#1E1F28] font-medium truncate'>
                实体类型-{tags.find((item) => item.tag_id === tag_id)?.tag_name}{' '}
                原文对照（共{fileOptions.length}篇）
              </div>
              <Select
                className='min-w-[220px]'
                placeholder='选择文件'
                options={fileOptions}
                value={selectedFileId}
                onChange={setSelectedFileId}
                showSearch
                optionFilterProp='label'
                disabled={fileOptions.length === 0}
              />
            </div>
            <div className='flex-1 overflow-auto'>
              {markdownLoading ? (
                <Skeleton active />
              ) : (
                <MarkdownPreview
                  content={markdownData ?? ''}
                  disableReference
                />
              )}
            </div>
          </div>

          {/* 右侧：对齐 NodeModal（固定 tag） */}
          <NodeContent
            key={String(editing)}
            editing={editing}
            setEditing={setEditing}
            {...props}
            isReadOnly={isReadOnly}
          />
        </div>
      )}
    </Drawer>
  )
}

const NodeContent: FC<
  NodeDrawerProps & {
    editing: boolean
    setEditing: (val: boolean) => void
  }
> = (props) => {
  const {
    open,
    onCancel,
    onSuccess,
    graphId,
    tag_id,
    initialValues,
    editing,
    setEditing,
    isReadOnly,
  } = props
  const { data } = useGraphInfo()
  const { tags = [] } = data ?? {}
  const { message } = App.useApp()
  const [form] = Form.useForm<FormValues>()
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
    <div className='flex-1 flex flex-col min-w-0 pl-2 relative'>
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
                  <Select placeholder='通用名' options={tagOptions} disabled />
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
                    disabled={!editing}
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
                  disabled={!editing || isReadOnly}
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
                  disabled={!editing || isReadOnly}
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
      {!editing ? (
        <Button
          className={`absolute right-0 top-0 border-[#0C99FF] text-[#0C99FF] ${
            isReadOnly ? 'cursor-not-allowed !text-[#C0C4CC] !border-[#C0C4CC]' : ''
          }`}
          disabled={isReadOnly}
          onClick={() => {
            if (isReadOnly) return
            setEditing(true)
          }}
        >
          编辑信息
        </Button>
      ) : null}
      <div
        className={cn(
          'mt-auto flex justify-end gap-2 pt-4 border-t border-[#E5E7EB]',
          {
            hidden: !editing,
          },
        )}
      >
        <Button onClick={() => setEditing(false)} disabled={loading}>
          取消
        </Button>
        <Button type='primary' onClick={submit} loading={loading}>
          确定
        </Button>
      </div>
    </div>
  )
}
