import { FC, ReactNode, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { App, Breadcrumb, Button, Empty, Skeleton } from 'antd'
import { getKnowledgeGraph } from '@/api/graph'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'
import EmptyIcon from '@/assets/icons/project/empty.svg'
import Neo4jGraph from '@/components/Neo4jGraph'
import { NodeModal } from '@/pages/graph/components/NodeModal'
import { useGraphInfo, withGraphProvider } from '../GraphProvider'
import {
  Neo4jGraphInstance,
  Neo4jGraphWithFilter,
} from '../components/Neo4jGraphWithFilter'
import { NodeDrawer } from '../components/NodeDrawer'
import { Filter, NodeFilter } from '../components/NodeFilter'
import { SpinWithMask } from '../components/SpinWithMask'
import useKnowledgeGraphData from '../hooks/useKnowledgeGraphData'
import GraphIcon from '../images/graph.svg?react'
import NodeSidebar from './NodeSidebar'

/** 空状态组件 */
const GraphEmptyState: FC = () => {
  return (
    <div className='flex flex-col items-center justify-center flex-1'>
      <img src={EmptyIcon} alt='' className='w-40 h-40 mb-3' />
      <p className='text-sm text-[#616373] font-normal'>图谱生成中，请稍后～</p>
    </div>
  )
}

/** neo4j图表  */
const Detail: FC = withGraphProvider(() => {
  const { data, loading, reloadTag } = useGraphInfo()
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const isEmpty = searchParams.get('empty') === 'true'
  const filterCache = useRef<Filter>()
  const [filter, setFilter] = useState<Filter>()
  const graphRef = useRef<Neo4jGraphInstance | null>(null)
  const [createNodeModalVisible, setCreateNodeModalVisible] = useState(false)
  const [nodeDrawerVisible, setNodeDrawerVisible] = useState(false)
  const [selectedNodeName, setSelectedNodeName] = useState<string>('')
  const [selectedTagName, setSelectedTagName] = useState<string>('')
  const [selectedOriginNode, setSelectedOriginNode] = useState<any>(null)

  // 权限控制：无权限时禁用操作按钮
  const isReadOnly = !data?.is_admin

  const selectedTagId = useMemo(() => {
    if (!selectedTagName) return undefined
    return data?.tags?.find((t) => t.tag_name === selectedTagName)?.tag_id
  }, [data?.tags, selectedTagName])

  const {
    nodes,
    relationships,
    loading: graphLoading,
    key,
  } = useKnowledgeGraphData({
    graph_id: data?.id,
    source: filter,
    twoWay: true,
  })
  // 面包屑导航项配置（必须在所有条件返回之前调用）
  const breadcrumbItems = useMemo(() => {
    const name = data?.name || ''
    return [
      {
        title: (
          <span
            className='flex items-center gap-2 text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
            onClick={() => navigate('/graph')}
          >
            <GraphIcon className='w-4 h-4' />
            <span>知识图谱</span>
          </span>
        ),
      },
      {
        title: (
          <span className='text-sm font-medium text-[#3C4149]'>{name}</span>
        ),
      },
    ]
  }, [data?.name, navigate])

  const withWrapper = (children: ReactNode) => {
    return (
      <div className='w-full h-full p-4 pt-[2px] bg-white flex flex-col'>
        {children}
      </div>
    )
  }
  if (loading) return withWrapper(<Skeleton active />)
  if (!data) return withWrapper(<Empty />)

  const { id: graph_id } = data

  const fetchGraphData = async () => {
    const { knowledge_graph } = await getKnowledgeGraph({
      graph_id,
      src_tag: filter?.tag_name,
      src_name: filter?.search,
      is_two_way: true,
    })
    return {
      nodes: knowledge_graph.nodes ?? [],
      relationships: knowledge_graph.edges ?? [],
    }
  }

  // 如果是空状态，显示空状态组件
  if (isEmpty) {
    return withWrapper(
      <>
        <div className='w-full h-12 bg-white flex items-center pt-[2px] font-medium px-1'>
          <Breadcrumb
            className='[&_span.ant-breadcrumb-separator]:inline-flex [&_span.ant-breadcrumb-separator]:items-center [&_span.ant-breadcrumb-separator]:align-middle'
            separator={<NavigationIcon className='inline-block' />}
            items={breadcrumbItems}
          />
        </div>
        <GraphEmptyState />
      </>,
    )
  }

  return withWrapper(
    <>
      <div className='w-full h-12 bg-white flex items-center pt-[2px] font-medium px-1'>
        <Breadcrumb
          className='[&_span.ant-breadcrumb-separator]:inline-flex [&_span.ant-breadcrumb-separator]:items-center [&_span.ant-breadcrumb-separator]:align-middle'
          separator={<NavigationIcon className='inline-block' />}
          items={breadcrumbItems}
        />
      </div>
      <div className='flex gap-4 items-center mt-4'>
        <NodeFilter
          onChange={(val) => (filterCache.current = val)}
          onClear={setFilter}
        />
        <Button
          type='primary'
          className='ml-2'
          onClick={() => setFilter(filterCache.current)}
        >
          搜索
        </Button>
        <Link
          className='ml-auto'
          to={`/graph/search-relationship?graphId=${graph_id}`}
        >
          <Button>实体找关联</Button>
        </Link>
        <Button
          type='primary'
          disabled={isReadOnly}
          className={isReadOnly ? 'cursor-not-allowed !text-[#C0C4CC]' : ''}
          onClick={() => {
            if (isReadOnly) return
            if (!data?.tags || data.tags.length === 0) {
              message.warning('请先创建实体类型')
              return
            }
            setCreateNodeModalVisible(true)
          }}
        >
          新建实体
        </Button>
        <Link
          to={isReadOnly ? '#' : `/graph/edit?graphId=${data.id}&rule=true`}
          onClick={(e) => {
            if (isReadOnly) {
              e.preventDefault()
            }
          }}
        >
          <Button
            type='primary'
            disabled={isReadOnly}
            className={isReadOnly ? 'cursor-not-allowed !text-[#C0C4CC]' : ''}
          >
            编辑规则
          </Button>
        </Link>
      </div>
      <div className='mt-4 flex-1 relative overflow-hidden'>
        <SpinWithMask show={graphLoading} />
        <Neo4jGraphWithFilter
          ref={graphRef}
          key={key}
          className=' absolute inset-0 right-11'
          graph_id={graph_id}
          nodes={nodes}
          relationships={relationships}
          fetchGraphData={fetchGraphData}
          autoFitOnRefresh
          isReadOnly={isReadOnly}
          onViewNodeDetail={(nodeName, tagName) => {
            setSelectedNodeName(nodeName)
            setSelectedTagName(tagName)
            setSelectedOriginNode(graphRef.current?.getOriginNode?.(nodeName))
            setNodeDrawerVisible(true)
          }}
        />
        <NodeSidebar
          nodes={nodes}
          tags={data.tags}
          onClickNode={(tag_name, node) => {
            setSelectedNodeName(node.name)
            setSelectedTagName(tag_name)
            setSelectedOriginNode(
              graphRef.current?.getOriginNode?.(node.name) ?? null,
            )
            setNodeDrawerVisible(true)
          }}
        />
      </div>
      {/* 创建实体弹窗 */}
      <NodeModal
        open={createNodeModalVisible}
        onCancel={() => {
          setCreateNodeModalVisible(false)
          graphRef.current?.setSelectedNode(undefined)
        }}
        onSuccess={(result) => {
          reloadTag()
          if (result.mode !== 'create') return
          const nodeName = result.node_name
          const tagName = result.tag_name
          if (!nodeName) return
          const graphInstance = graphRef.current?.graph
          graphInstance?.dispatchNodeAction({
            type: 'add',
            nodeData: {
              id: nodeName,
              elementId: nodeName,
              labels: tagName ? [tagName] : ['null'],
              properties: { name: nodeName },
              propertyTypes: { name: 'String' },
            },
          })

          graphRef.current?.upsertOriginNode?.({
            name: nodeName,
            tags: [
              {
                name: tagName,
                properties_values: result.properties_values,
              },
            ],
          })

          const graphModel = graphInstance?.graph
          ;(result.edges ?? [])
            .filter((e) => e?.dst_node_name && e?.edge_name)
            .forEach((e) => {
              const src = e.src_node_name || nodeName
              const dst = e.dst_node_name
              const type = e.edge_name
              if (!graphModel?.findNode(dst)) {
                graphInstance?.dispatchNodeAction({
                  type: 'add',
                  nodeData: {
                    id: dst,
                    elementId: dst,
                    labels: ['null'],
                    properties: { name: dst },
                    propertyTypes: { name: 'String' },
                  },
                })
              }
              const id = `${src}_${dst}_${type}`
              graphInstance?.dispatchEdgeAction({
                type: 'add',
                edgeData: {
                  id,
                  elementId: id,
                  startNodeId: src,
                  endNodeId: dst,
                  type,
                  properties: {},
                  propertyTypes: {},
                },
              })
            })

          graphInstance?.zoomToFit()
        }}
        graphId={graph_id!}
      />
      {/* 实体详情 Drawer */}
      {selectedNodeName && selectedTagId ? (
        <NodeDrawer
          open={nodeDrawerVisible}
          onCancel={() => {
            setNodeDrawerVisible(false)
            setSelectedNodeName('')
            setSelectedTagName('')
            setSelectedOriginNode(null)
            graphRef.current?.setSelectedNode(undefined)
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
              graphRef.current?.renameOriginNode?.(oldName, newName, originNode)
            } else {
              graphRef.current?.upsertOriginNode?.(originNode)
            }

            const graphInstance = graphRef.current?.graph
            const graphModel = graphInstance?.graph
            if (!graphModel) {
              void graphRef.current?.refresh?.()
              return
            }

            const currentNodeModel = graphModel.findNode(oldName)
            const connected = currentNodeModel
              ? (graphModel.findAllRelationshipToNode(currentNodeModel) ?? [])
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

            graphInstance.dispatchNodeAction({
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
              graphInstance.dispatchEdgeAction({ type: 'add', edgeData })
            })

            graphInstance.zoomToFit()
          }}
          graphId={graph_id!}
          tag_id={selectedTagId}
          isReadOnly={isReadOnly}
          initialValues={{
            node_name: selectedNodeName,
            tag_id: selectedTagId,
            tag_name: selectedTagName,
            properties_values: [
              {
                tag_id: selectedTagId,
                tag_name: selectedTagName,
                properties_values:
                  selectedOriginNode?.tags?.find(
                    (t: any) => t?.name === selectedTagName,
                  )?.properties_values ?? [],
              },
            ],
          }}
        />
      ) : null}
    </>,
  )
})

export default Detail
