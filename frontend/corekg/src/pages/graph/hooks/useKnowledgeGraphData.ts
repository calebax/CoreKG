import { useRef } from 'react'
import { App } from 'antd'
import { useRequest } from 'ahooks'
import { getKnowledgeGraph } from '@/api/graph'
import { type Filter } from '../components/NodeFilter'

export type UseKnowledgeGraphDataConfig = {
  graph_id?: number
  source?: Filter
  target?: Filter
  twoWay?: boolean
  disableFilter?: boolean
}

const useKnowledgeGraphData = (config: UseKnowledgeGraphDataConfig) => {
  const { message } = App.useApp()
  const { graph_id, source, target, twoWay, disableFilter } = config
  const keyRef = useRef(0)

  const stripSurroundingDoubleQuotes = (input: unknown): unknown => {
    if (typeof input !== 'string') return input
    if (input.length < 2) return input
    if (!input.startsWith('"') || !input.endsWith('"')) return input
    return input.slice(1, -1)
  }

  const normalizeNodes = (nodes: any[] | undefined) => {
    if (!nodes || nodes.length === 0) return nodes ?? []
    return nodes.map((n) => {
      const tags = n?.tags
      if (!Array.isArray(tags) || tags.length === 0) return n
      return {
        ...n,
        tags: tags.map((t: any) => {
          const pvs = t?.properties_values
          if (!Array.isArray(pvs) || pvs.length === 0) return t
          return {
            ...t,
            properties_values: pvs.map((pv: any) => ({
              ...pv,
              value: stripSurroundingDoubleQuotes(pv?.value),
            })),
          }
        }),
      }
    })
  }

  const { data, loading } = useRequest(
    async () => {
      if (!graph_id) return
      const { knowledge_graph } = await getKnowledgeGraph({
        graph_id,
        src_tag: source?.tag_name,
        src_name: source?.search,
        dst_tag: target?.tag_name,
        dst_name: target?.search,
        is_two_way: twoWay,
      })
      const { nodes, edges: relationships } = knowledge_graph
      const normalizedNodes = normalizeNodes(nodes)
      if (!nodes || nodes.length === 0) {
        message.info('没有找到相关节点')
      }
      ++keyRef.current
      return {
        nodes: normalizedNodes,
        relationships,
      }
    },
    {
      refreshDeps: [graph_id, source, target],
      ready: !disableFilter,
    },
  )

  return {
    nodes: data?.nodes ?? [],
    relationships: data?.relationships ?? [],
    loading: loading || !graph_id,
    key: keyRef.current,
  }
}

export default useKnowledgeGraphData
