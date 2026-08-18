import { forwardRef, useImperativeHandle, useRef, useState } from 'react'
import { App, Button, Form, Modal, Select } from 'antd'
import { SyncOutlined } from '@ant-design/icons'
import { useCounter, useCreation } from 'ahooks'
import { cn } from '@/utils'
import { createNodeEdge, deleteNode, getKnowledgeGraph } from '@/api/graph'
import Neo4jGraph from '@/components/Neo4jGraph'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import { NodeModal } from '../NodeModal'
import type { NodePropertyValue } from '../NodeProperties'
import type { NodeEdge } from '../NodeRelationships'
import { RelationshipNameSelect } from '../RelationshipNameSelect'
import { NodeDetail } from './NodeDetail'
import { getNeo4jEdges, getNeo4jNodes, graphStyle } from './common'
import useOriginNodes from './useOriginNodes'

//// 在properties里使用key `Neo4jTitleKey`用于展示其名称 展示node卡片的时候需要过滤掉

export const Neo4jGraphWithFilter = forwardRef<
  Neo4jGraphInstance,
  Style & {
    graph_id: number
    nodes: any[]
    relationships: any[]
    onViewNodeDetail?: (nodeName: string, tagName: string) => void
    fetchGraphData?: () => Promise<{ nodes: any[]; relationships: any[] }>
    autoFitOnRefresh?: boolean
    isReadOnly?: boolean
  }
>((props, ref) => {
  const {
    graph_id,
    nodes,
    relationships,
    className,
    style,
    fetchGraphData,
    autoFitOnRefresh,
    isReadOnly,
  } = props
  const [key, { inc }] = useCounter()
  return (
    <div className={cn('relative', className)} style={style}>
      <Button
        className=' absolute left-2 top-2 z-10'
        icon={<SyncOutlined />}
        onClick={() => {
          if (!nodes || nodes.length === 0) return
          inc()
        }}
      >
        重置视角
      </Button>
      <Neo4jGraphInner
        key={key}
        ref={ref}
        className='w-full h-full'
        graph_id={graph_id}
        defaultNodes={nodes}
        defaultRelationships={relationships}
        onViewNodeDetail={props.onViewNodeDetail}
        fetchGraphData={fetchGraphData}
        autoFitOnRefresh={autoFitOnRefresh}
        isReadOnly={isReadOnly}
      />
    </div>
  )
})

type Neo4jGraphInnerProps = Style & {
  graph_id: number
  defaultNodes: any[]
  defaultRelationships: any[]
  onViewNodeDetail?: (nodeName: string, tagName: string) => void
  fetchGraphData?: () => Promise<{ nodes: any[]; relationships: any[] }>
  autoFitOnRefresh?: boolean
  isReadOnly?: boolean
}

type NodeModalInitialValues = {
  node_name: string
  tag_id?: number
  tag_name?: string
  /** 多 tag 的属性集合（对齐 NodeDetail：切换 tag 展示对应 properties_values） */
  properties_values?: {
    tag_id: number
    tag_name: string
    properties_values?: NodePropertyValue[]
  }[]
  edges?: NodeEdge[]
  available_tags?: { tag_id: number; tag_name: string }[]
}

export type Neo4jGraphInstance = {
  graph?: Neo4jGraph | null
  refresh?: () => Promise<void>
  upsertOriginNode?: (node: any) => void
  renameOriginNode?: (oldName: string, newName: string, node?: any) => void
  getOriginNode?: (nodeName: string) => any
  setSelectedNode: (val: any) => void
}

const Neo4jGraphInner = forwardRef<Neo4jGraphInstance, Neo4jGraphInnerProps>(
  (props, ref) => {
    const {
      graph_id,
      defaultNodes,
      defaultRelationships,
      className,
      style,
      onViewNodeDetail,
      fetchGraphData,
      autoFitOnRefresh,
      isReadOnly,
    } = props
    const { message, modal } = App.useApp()
    const { data, reloadTag } = useGraphInfo()
    const { tags: graphTags = [], edgeOptions = [] } = data ?? {}
    const [selectedItem, setSelectedItem] = useState<any>()
    const [createEdgeModalVisible, setCreateEdgeModalVisible] = useState(false)
    const [createEdgeSubmitting, setCreateEdgeSubmitting] = useState(false)
    type CreateEdgeFormValues = {
      src_tag_id: number
      src_node_name: string
      dst_tag_id: number
      dst_node_name: string
      edge_name: string
    }
    const [createEdgeForm] = Form.useForm<CreateEdgeFormValues>()
    const [createEdgeSrcTagOptions, setCreateEdgeSrcTagOptions] = useState<
      { label: string; value: number }[]
    >([])
    const [createEdgeDstTagOptions, setCreateEdgeDstTagOptions] = useState<
      { label: string; value: number }[]
    >([])
    const [nodeModalOpen, setNodeModalOpen] = useState(false)
    const [nodeModalInitialValues, setNodeModalInitialValues] =
      useState<NodeModalInitialValues>()
    const nodeTagIdCacheRef = useRef<Map<string, number>>(new Map())
    const { addNodes, upsertNode, upsertNodes, renameNode, getNode } =
      useOriginNodes(defaultNodes)
    const neo4jGraphRef = useRef<Neo4jGraph>(null)

    const handleRefreshGraph = async () => {
      if (!neo4jGraphRef.current) return
      try {
        const fetchLatestGraphData =
          fetchGraphData ??
          (async () => {
            const { knowledge_graph } = await getKnowledgeGraph({ graph_id })
            return {
              nodes: knowledge_graph.nodes ?? [],
              relationships: knowledge_graph.edges ?? [],
            }
          })

        const { nodes, relationships } = await fetchLatestGraphData()

        upsertNodes(nodes ?? [])
        nodeTagIdCacheRef.current.clear()
        setSelectedItem(undefined)

        neo4jGraphRef.current.resetGraphData(
          getNeo4jNodes(nodes ?? []),
          getNeo4jEdges(relationships ?? []),
          { zoomToFit: autoFitOnRefresh ?? true },
        )
      } catch {
        message.error('刷新图谱失败')
      }
    }

    useImperativeHandle(ref, () => {
      return {
        graph: neo4jGraphRef.current,
        refresh: handleRefreshGraph,
        upsertOriginNode: upsertNode,
        renameOriginNode: renameNode,
        getOriginNode: getNode,
        setSelectedNode: setSelectedItem,
      }
    })
    // 这两项只生效一次
    const defaultNeo4jNodes = useCreation(() => {
      return getNeo4jNodes(defaultNodes)
    }, [])
    const defaultNeo4jRelationships = useCreation(() => {
      return getNeo4jEdges(defaultRelationships)
    }, [])

    const getNodePrimaryTagId = async (nodeName: string) => {
      if (!nodeName) return undefined
      const cached = nodeTagIdCacheRef.current.get(nodeName)
      if (cached !== undefined) return cached
      const node = getNode(nodeName)
      const firstTagName = node?.tags?.[0]?.name as string | undefined
      const tagId = firstTagName
        ? graphTags.find((t) => t.tag_name === firstTagName)?.tag_id
        : undefined
      if (tagId !== undefined) nodeTagIdCacheRef.current.set(nodeName, tagId)
      return tagId
    }

    const edgeNameOptions = useCreation(() => {
      return edgeOptions.map((opt) => ({ label: opt, value: opt }))
    }, [edgeOptions])

    return (
      <div className={cn('relative', className)} style={style}>
        {/* 左下角节点卡片 */}
        {selectedItem ? (
          <NodeDetail
            className='absolute left-2 bottom-2 z-10'
            item={selectedItem}
            clearItem={() => setSelectedItem(undefined)}
            onViewDetail={onViewNodeDetail}
            key={selectedItem.name}
          />
        ) : null}
        <Neo4jGraph
          ref={neo4jGraphRef}
          showNodeOperators={true}
          isReadOnly={isReadOnly}
          nodes={defaultNeo4jNodes}
          relationships={defaultNeo4jRelationships}
          graphStyle={graphStyle}
          onItemSelect={(item) => {
            if (item.type === 'node') {
              const node: any = item.item
              const name = node.propertyList[0].value
              setSelectedItem(getNode(name))
            }
          }}
          getNodeNeighbours={async (node: any, _, cb) => {
            const name = node.propertyList[0].value
            const { knowledge_graph } = await getKnowledgeGraph({
              graph_id,
              src_name: name,
            })
            const newNodes = knowledge_graph.nodes ?? []
            const newRelationships = knowledge_graph.edges ?? []
            addNodes(newNodes)
            cb({
              nodes: getNeo4jNodes(newNodes),
              relationships: getNeo4jEdges(newRelationships),
            })
          }}
          onClickOperators={async (type, node, endNode) => {
            // 无权限时显示提示并阻止操作
            if (isReadOnly) {
              message.warning('您没有权限执行此操作')
              return
            }
            const nodeName = node.propertyList[0].value
            switch (type) {
              case 'createEdge': {
                if (!endNode) {
                  message.error('请选择目标节点')
                  return
                }
                const srcNodeName = node.propertyList[0].value
                const dstNodeName = endNode.propertyList[0].value
                const [srcTagId, dstTagId] = await Promise.all([
                  getNodePrimaryTagId(srcNodeName),
                  getNodePrimaryTagId(dstNodeName),
                ])
                const srcNodeCached = getNode(srcNodeName)
                const dstNodeCached = getNode(dstNodeName)
                const srcTagNames = (srcNodeCached?.tags ?? [])
                  .map((t: any) => t?.name as string | undefined)
                  .filter(Boolean) as string[]
                const dstTagNames = (dstNodeCached?.tags ?? [])
                  .map((t: any) => t?.name as string | undefined)
                  .filter(Boolean) as string[]

                const srcTagOptions = graphTags
                  .filter((t) => srcTagNames.includes(t.tag_name))
                  .map((t) => ({ label: t.tag_name, value: t.tag_id }))
                const dstTagOptions = graphTags
                  .filter((t) => dstTagNames.includes(t.tag_name))
                  .map((t) => ({ label: t.tag_name, value: t.tag_id }))

                setCreateEdgeSrcTagOptions(srcTagOptions)
                setCreateEdgeDstTagOptions(dstTagOptions)

                createEdgeForm.setFieldsValue({
                  src_tag_id:
                    srcTagId ?? srcTagOptions[0]?.value ?? graphTags[0]?.tag_id,
                  src_node_name: srcNodeName,
                  dst_tag_id:
                    dstTagId ?? dstTagOptions[0]?.value ?? graphTags[0]?.tag_id,
                  dst_node_name: dstNodeName,
                  edge_name: '',
                })
                setCreateEdgeModalVisible(true)
                break
              }
              case 'deleteNode': {
                modal.confirm({
                  title: '确认删除？',
                  onOk: async () => {
                    try {
                      await deleteNode({
                        graph_id,
                        node_name: nodeName,
                      })
                      // 从状态中移除节点和相关边
                      neo4jGraphRef.current?.dispatchNodeAction({
                        type: 'delete',
                        nodeId: nodeName,
                      })
                      // 如果删除的节点当前被选中，清除选中状态
                      if (selectedItem?.name === nodeName) {
                        setSelectedItem(undefined)
                      }
                      message.success('删除成功')
                    } catch {
                      message.error('删除失败')
                    }
                  },
                })
                break
              }
              case 'editNode': {
                if (!nodeName) {
                  message.warning('请选择要编辑的实体')
                  return
                }
                try {
                  const cachedNode = getNode(nodeName)
                  if (!cachedNode) {
                    message.error('未找到实体数据，请刷新图谱后重试')
                    return
                  }
                  const firstTagName = cachedNode.tags?.[0]?.name as
                    | string
                    | undefined
                  const firstTagId = firstTagName
                    ? graphTags.find((t) => t.tag_name === firstTagName)?.tag_id
                    : undefined
                  const availableTags = (cachedNode.tags ?? [])
                    .map((t: any) => t?.name as string | undefined)
                    .filter(Boolean)
                    .map((tagName: string) => {
                      const tag = graphTags.find((t) => t.tag_name === tagName)
                      return tag
                        ? { tag_id: tag.tag_id, tag_name: tag.tag_name }
                        : null
                    })
                    .filter(Boolean) as { tag_id: number; tag_name: string }[]
                  const nodeTags = (cachedNode.tags ?? [])
                    .map((t: any) => {
                      const tagName = t?.name as string | undefined
                      if (!tagName) return null
                      const tagDef = graphTags.find(
                        (gt) => gt.tag_name === tagName,
                      )
                      if (!tagDef) return null
                      const pvs: NodePropertyValue[] =
                        (t?.properties_values ?? [])?.map(
                          (pv: { name: string; value: any }) => ({
                            name: pv.name,
                            value: pv.value,
                          }),
                        ) ?? []
                      return {
                        tag_id: tagDef.tag_id,
                        tag_name: tagDef.tag_name,
                        properties_values: pvs,
                      }
                    })
                    .filter(Boolean) as {
                    tag_id: number
                    tag_name: string
                    properties_values?: NodePropertyValue[]
                  }[]

                  setNodeModalInitialValues({
                    node_name: cachedNode.name ?? nodeName,
                    tag_id: firstTagId ?? nodeTags[0]?.tag_id,
                    tag_name: firstTagName,
                    available_tags: availableTags,
                    properties_values: nodeTags,
                  })
                  setNodeModalOpen(true)
                } catch {
                  message.error('获取实体详情失败')
                }
                break
              }
              default:
                break
            }
          }}
        />
        {/* 创建边的弹窗 */}
        <Modal
          title='编辑实体关系'
          open={createEdgeModalVisible}
          onCancel={() => {
            if (createEdgeSubmitting) return
            setCreateEdgeModalVisible(false)
            createEdgeForm.resetFields()
            setCreateEdgeSrcTagOptions([])
            setCreateEdgeDstTagOptions([])
          }}
          okButtonProps={{ loading: createEdgeSubmitting }}
          cancelButtonProps={{ disabled: createEdgeSubmitting }}
          onOk={async () => {
            if (createEdgeSubmitting) return
            setCreateEdgeSubmitting(true)
            try {
              const formValue = await createEdgeForm.validateFields()
              const {
                edge_name,
                src_tag_id,
                dst_tag_id,
                src_node_name,
                dst_node_name,
              } = formValue
              if (!src_tag_id || !dst_tag_id) {
                message.error('请选择源/目标实体类型')
                return
              }
              if (!src_node_name || !dst_node_name) {
                message.error('请选择源/目标实体')
                return
              }
              await createNodeEdge({
                graph_id,
                edge: {
                  edge_name,
                  src_node_name,
                  dst_node_name,
                  src_tag_id,
                  dst_tag_id,
                },
              })

              const graphInstance = neo4jGraphRef.current
              const graphModel = graphInstance?.graph
              if (
                graphInstance &&
                graphModel &&
                !graphModel.findNode(dst_node_name)
              ) {
                const dstCached = getNode(dst_node_name)
                const dstFirstTagName = dstCached?.tags?.[0]?.name as
                  | string
                  | undefined
                graphInstance.dispatchNodeAction({
                  type: 'add',
                  nodeData: {
                    id: dst_node_name,
                    elementId: dst_node_name,
                    labels: dstFirstTagName ? [dstFirstTagName] : ['null'],
                    properties: { name: dst_node_name },
                    propertyTypes: { name: 'String' },
                  },
                })
              }

              const edgeId = `${src_node_name}_${dst_node_name}_${edge_name}`
              neo4jGraphRef.current?.dispatchEdgeAction({
                type: 'add',
                edgeData: {
                  id: edgeId,
                  elementId: edgeId,
                  startNodeId: src_node_name,
                  endNodeId: dst_node_name,
                  type: edge_name,
                  properties: {},
                  propertyTypes: {},
                },
              })
              message.success('创建关系成功')
              setCreateEdgeModalVisible(false)
              createEdgeForm.resetFields()
            } catch {
              message.error('创建关系失败')
            } finally {
              setCreateEdgeSubmitting(false)
            }
          }}
        >
          <Form form={createEdgeForm} layout='vertical'>
            <Form.Item
              name='src_tag_id'
              label='源实体类型'
              rules={[{ required: true, message: '请选择源实体类型' }]}
            >
              <Select
                placeholder='请选择源实体类型'
                options={createEdgeSrcTagOptions}
                showSearch
                optionFilterProp='label'
                disabled={createEdgeSrcTagOptions.length === 0}
              />
            </Form.Item>
            <Form.Item
              name='src_node_name'
              label='源实体'
              rules={[{ required: true, message: '请选择源实体' }]}
            >
              <Select disabled placeholder='源实体' />
            </Form.Item>
            <Form.Item
              name='edge_name'
              label='实体关系'
              rules={[{ required: true, message: '请选择实体关系' }]}
            >
              <RelationshipNameSelect
                options={edgeNameOptions}
                placeholder='请选择实体关系'
                maxLength={20}
              />
            </Form.Item>
            <Form.Item
              name='dst_tag_id'
              label='目标实体类型'
              rules={[{ required: true, message: '请选择目标实体类型' }]}
            >
              <Select
                placeholder='请选择目标实体类型'
                options={createEdgeDstTagOptions}
                showSearch
                optionFilterProp='label'
                disabled={createEdgeDstTagOptions.length === 0}
              />
            </Form.Item>
            <Form.Item
              name='dst_node_name'
              label='目标实体'
              rules={[{ required: true, message: '请选择或输入目标实体' }]}
            >
              <Select disabled placeholder='目标实体' />
            </Form.Item>
          </Form>
        </Modal>
        {/* 编辑实体弹窗 */}
        {nodeModalOpen && (
          <NodeModal
            open={nodeModalOpen}
            onCancel={() => {
              setNodeModalOpen(false)
              setNodeModalInitialValues(undefined)
              setSelectedItem(undefined)
            }}
            onSuccess={(result) => {
              reloadTag()
              if (result.mode !== 'edit') return
              const oldName = result.old_node_name ?? result.node_name
              const newName = result.node_name
              const originNode = {
                name: newName,
                tags: (result.node_tags ?? []).map((t) => ({
                  name: t.tag_name,
                  properties_values: t.properties_values,
                })),
              }
              if (oldName !== newName) {
                renameNode(oldName, newName, originNode)
                nodeTagIdCacheRef.current.delete(oldName)
              } else {
                upsertNode(originNode)
              }
              nodeTagIdCacheRef.current.set(newName, result.tag_id)

              const graphInstance = neo4jGraphRef.current
              const graphModel = graphInstance?.graph
              const currentNodeModel = graphModel?.findNode(oldName)
              const connected = currentNodeModel
                ? (graphModel?.findAllRelationshipToNode(currentNodeModel) ??
                  [])
                : []
              const incomingEdges = connected
                .filter((r) => r.target.id === oldName)
                .map((r) => {
                  const src = r.source.id
                  const dst = newName
                  const type = r.type
                  const id = `${src}_${dst}_${type}`
                  return {
                    id,
                    elementId: id,
                    startNodeId: src,
                    endNodeId: dst,
                    type,
                    properties: {},
                    propertyTypes: {},
                  }
                })

              graphInstance?.dispatchNodeAction({
                type: 'edit',
                nodeId: oldName,
                nodeData: {
                  id: newName,
                  elementId: newName,
                  labels: result.tag_name ? [result.tag_name] : ['null'],
                  properties: { name: newName },
                  propertyTypes: { name: 'String' },
                },
              })

              const outgoingEdges = (result.edges ?? [])
                .filter((e) => e?.dst_node_name && e?.edge_name)
                .filter((e) => {
                  const dst = e?.dst_node_name
                  if (!dst) return false
                  return !!graphModel?.findNode(dst)
                })
                .map((e) => {
                  const src = e.src_node_name || newName
                  const dst = e.dst_node_name
                  const type = e.edge_name
                  const id = `${src}_${dst}_${type}`
                  return {
                    id,
                    elementId: id,
                    startNodeId: src,
                    endNodeId: dst,
                    type,
                    properties: {},
                    propertyTypes: {},
                  }
                })

              ;[...incomingEdges, ...outgoingEdges].forEach((edgeData) => {
                graphInstance?.dispatchEdgeAction({ type: 'add', edgeData })
              })

              if (
                selectedItem?.name === oldName ||
                selectedItem?.name === newName
              ) {
                setSelectedItem(getNode(newName) ?? originNode)
              }

              graphInstance?.zoomToFit()
            }}
            graphId={graph_id}
            initialValues={nodeModalInitialValues}
          />
        )}
      </div>
    )
  },
)

Neo4jGraphInner.displayName = 'Neo4jGraphInner'
