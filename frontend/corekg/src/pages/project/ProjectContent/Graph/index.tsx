import { FC } from 'react'
import { Button } from 'antd'
import { SyncOutlined } from '@ant-design/icons'
import { useCounter } from 'ahooks'
import { cn } from '@/utils'
import Neo4jGraph from '@/components/Neo4jGraph'
import { GraphStyleModel } from '@/components/Neo4jGraph/models/GraphStyle'
import { BasicNode, BasicRelationship } from '@/components/Neo4jGraph/types'
import { AIDialog } from '@/components/dialog'
import { useSessionInfo } from '..'
import { NodeDetail } from './NodeDetail'
import NoDataIcon from './empty.svg?react'

type GraphData = {
  edges: { src: string; dst: string; name: string }[]
  nodes: { name: string; tags: { name: string; properties_values: any[] }[] }[]
}
export const Graph: FC = () => {
  const { dialog, dialogIndex } = useSessionInfo()
  const originGraph = useMemo(() => {
    const current = (dialog[dialogIndex] ?? dialog.at(-1)) as
      | AIDialog
      | undefined
    return current?.graph as
      | undefined
      | {
          /** 高亮 */
          graph_chat_reference: GraphData | null
          /** 全部节点 */
          graph_reference: GraphData | null
        }
  }, [dialog, dialogIndex])
  const [count, { inc }] = useCounter(0)
  const graph = useMemo(() => {
    if (!originGraph || !originGraph.graph_reference) return null
    const { graph_reference, graph_chat_reference } = originGraph
    const nodes = getNeo4jNodes(graph_reference.nodes)
    nodes.forEach((n) => {
      if (graph_chat_reference?.nodes.some((item) => item.name === n.id)) {
        n.labels = ['highlight']
      } else {
        n.labels = ['common']
      }
    })
    const edges: BasicRelationship[] = []
    getNeo4jEdges(graph_reference.edges).map((item) => {
      const start = nodes.find((n) => n.id === item.startNodeId)
      const end = nodes.find((n) => n.id === item.endNodeId)
      if (!start || !end) return
      item.properties = { name: item.type }
      item.propertyTypes = { name: 'String' }
      item.type = 'common'
      if ([start.labels[0], end.labels[0]].includes('highlight')) {
        item.type = 'highlight'
      }
      edges.push(item)
    })
    return {
      nodes,
      edges,
    }
  }, [originGraph])
  const [selectedItem, setSelectedItem] = useState<any>()
  useEffect(() => {
    setSelectedItem(null)
  }, [dialogIndex])
  const graphStyle = useMemo(() => {
    const _style = new GraphStyleModel()
    _style.loadRules({
      node: {
        diameter: '50px',
        color: '#A5ABB6',
        'border-width': '0px',
        'text-color-internal': '#FFFFFF',
        'font-size': '10px',
      },
      'node.common': {
        color: '#AAA',
      },
      'node.highlight': {
        color: '#0C99FF',
      },
      relationship: {
        color: '#A5ABB6',
        'shaft-width': '1px',
        'font-size': '8px',
        padding: '3px',
        'text-color-external': '#000000',
        'text-color-internal': '#FFFFFF',
        caption: '{name}',
      },
      'relationship.highlight': {
        color: '#0C99FF',
      },
    })
    return _style
  }, [])
  if (!graph) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-[30px]'>
        <div>
          <NoDataIcon />
        </div>
        <div className='text-[#616373] font-[400] text-[14px] leading-[1]'>
          暂无相关图谱～
        </div>
      </div>
    )
  }
  const { nodes, edges } = graph
  return (
    <div className={'h-full p-2 relative flex flex-col'}>
      <Button
        className={cn(
          'my-2 self-start',
          ' bg-[#F8F9FD] text-[#616373] hover:text-[#CC5DE8] hover:border-[#CC5DE8]',
        )}
        icon={<SyncOutlined />}
        onClick={() => inc()}
      >
        重置视角
      </Button>
      <div className='flex-1 overflow-auto'>
        <Neo4jGraph
          key={`${dialogIndex}-${count}`}
          isFullscreen={false}
          nodes={nodes}
          relationships={edges.filter(
            (e) =>
              nodes.some((n) => n.id === e.startNodeId) &&
              nodes.some((n) => n.id === e.endNodeId),
          )}
          graphStyle={graphStyle}
          onItemSelect={(item) => {
            if (item.type === 'node') {
              const node: any = item.item
              const name = node.propertyList[0].value
              setSelectedItem(
                originGraph?.graph_reference?.nodes.find(
                  (item) => item.name === name,
                ),
              )
            }
          }}
          getNodeNeighbours={async (node: any, _, cb) => {
            cb({
              nodes: [],
              relationships: [],
            })
          }}
        />
      </div>

      {selectedItem ? (
        <NodeDetail
          className='absolute left-2 bottom-2 z-10'
          item={selectedItem}
          clearItem={() => setSelectedItem(undefined)}
          key={selectedItem.name}
        />
      ) : null}
    </div>
  )
}

const getNeo4jNodes = (
  nodes: {
    name: string
    tags?:
      | {
          name: string
          properties_values?:
            | {
                name: string
                value: string
              }[]
            | null
        }[]
      | null
  }[],
): BasicNode[] => {
  return nodes.map((n) => {
    const { name, tags } = n
    return {
      id: name,
      elementId: name,
      labels: tags ? tags.map((t) => t.name) : ['null'],
      properties: { name },
      propertyTypes: { name: 'String' },
    }
  })
}
const getNeo4jEdges = (
  edges: { src: string; dst: string; name: string }[],
): BasicRelationship[] => {
  return edges.map((e) => {
    const { src, dst, name } = e
    const id = `${src}_${dst}_${name}`
    return {
      id,
      elementId: id,
      startNodeId: src,
      endNodeId: dst,
      type: name,
      properties: {},
      propertyTypes: {},
    }
  })
}
