import { getKnowledgeGraph } from '@/api/graph'
import { BasicNode, BasicRelationship } from '@/components/Neo4jGraph/types'

export * from './graphStyle'

//// 在properties里使用key `Neo4jTitleKey`用于展示其名称

/** 转化为neo4j的node.会加入用于展示的特殊属性 */
export const getNeo4jNodes = (
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
export const getNeo4jEdges = (
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
