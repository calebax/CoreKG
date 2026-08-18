import { GraphTag, GraphTemplate } from 'Graph'
import { send } from './request'

/** 获取图谱列表 */
export const listForestGraph = (data: CommonArgs) =>
  send('forest.ListForestGraph', data)
/** 创建图谱 */
export const createGraph = (data: {
  description?: string
  name: string
  manager_ids: number[]
  public_scope: 'company' | 'custom'
  scope_ids?: number[]
  avatar_url: string
}) => send('forest.CreateGraph', data)
/** 删除图谱 */
export const deleteGraph = (data: { graph_id: number }) =>
  send('forest.DeleteGraph', data)
/** 获取图谱信息 */
export const getGraphInfo = (data: { graph_id: number }) =>
  send('forest.GetGraphInfo', data)
/** 获取所有tag */
export const listGraphTag = (data: { graph_id: number } & CommonArgs) =>
  send('forest.ListGraphTag', data)
/** 选择模板 */
export const submitTemplate = (data: {
  graph_id: number
  template: Pick<GraphTemplate, 'edges' | 'tags'>
}) => send('forest.SubmitTemplate', data)
/** 创建实体类型 */
export const createTag = (data: { graph_id: number } & GraphTag) =>
  send('forest.CreateTag', data)
/** 删除实体类型 */
export const deleteTag = (data: { graph_id: number; tag_id: number }) =>
  send('forest.DeleteTag', data)
/** 更新tag */
export const updateTag = (
  data: {
    graph_id: number
    tag_id: number
  } & GraphTag,
) => send('forest.UpdateTag', data)
/** 获取一个类型的关系 */
export const getTagEdge = (data: { graph_id: number; tag_id: number }) =>
  send('forest.GetTagEdge', data)
/** 创建Tag间的关系 */
export const createEdge = (data: {
  graph_id: number
  src_tag_id: number
  dst_tag_id: number
  /** 后端写错 */
  egde_name: string
}) => send('forest.CreateEdge', data)
/** 更新tag间的关系.不能改名字 */
export const updateEdge = (data: {
  dst_tag_id: number
  edge_id: number
  graph_id: number
  src_tag_id: number
}) => send('forest.UpdateEdge', data)
/** 删除Tag间的关系 */
export const deleteEdge = (data: { graph_id: number; edge_id: number }) =>
  send('forest.DeleteEdge', data)

/** 对节点进行模糊搜索 */
export const listGraphNode = (data: {
  graph_id: number
  // 以下两项至少有一个
  graph_tag_id?: number
  graph_node_name?: string
}) => send('forest.ListGraphNode', data)
/** 获取节点和关系 */
export const getKnowledgeGraph = (data: {
  graph_id: number
  limit?: number
  // 起点
  src_name?: string
  src_tag?: string
  // 终点
  dst_name?: string
  dst_tag?: string
  /** 是否查双向边 */
  is_two_way?: boolean
}) => send('forest.GetKnowledgeGraph', data)
/** 增量更新图谱信息 */
export const updateGraph = (data: {
  graph_id: number
  name?: string
  description?: string
  file_id_list?: number[]
  parse_mode?: 'auto' | 'rule'
  manager_ids?: number[]
  public_scope?: 'company' | 'custom'
  scope_ids?: number[]
}) => send('forest.UpdateGraph', data)
/** 开始解析图谱 */
export const parseGraph = (data: { graph_id: number }) =>
  send('forest.ParseGraph', data)

/** 为知识库创建graph */
export const createForestGraph = (data: {
  forest_id: number
  avatar_url?: string
}) => send('forest.CreateForestGraph', data)

/** 全量更新图谱 */
export const restockGraph = (data: { graph_id: number }) =>
  send('forest.RestockGraph', data)

// ============= 图节点相关 API =============

/** 创建实体。入参：实体类型/实体名称/实体属性/实体关系，返回值：节点id */
export const createNode = (data: {
  graph_id: number
  node_name: string
  tags: {
    tag_name: string
    properties: {
      comment?: string
      defaults?: null | string | number | boolean
      name: string
      type: string
    }[]
    properties_values: {
      name: string
      value: null | string | number | boolean
    }[]
  }[]
  edges?: {
    src_tag_id: number
    dst_tag_id: number
    dst_node_name: string
    edge_name: string
    src_node_name: string
  }[]
}) =>
  send('forest.CreateNode', data) as Promise<{
    node_id: number
  }>

/** 编辑实体之间的关系。入参：名称/起点id/终点id */
export const createNodeEdge = (data: {
  graph_id: number
  edge: {
    edge_name: string
    src_node_name?: string
    dst_node_name?: string
    src_tag_id: number
    dst_tag_id: number
  }
}) =>
  send('forest.CreateNodeEdge', data) as Promise<{
    edge_id: number
  }>

/** 删除实体。入参：id */
export const deleteNode = (data: { graph_id: number; node_name: string }) =>
  send('forest.DeleteNode', data) as Promise<void>

/** 编辑实体。入参：id以及创建所需参数（全量） */
export const editNode = (data: {
  graph_id: number
  old_node_name: string
  tags: {
    tag_name: string
    properties: {
      comment?: string
      defaults?: null | string | number | boolean
      name: string
      type: string
    }[]
    properties_values: {
      name: string
      value: null | string | number | boolean
    }[]
  }[]
  edges?: {
    src_tag_id: number
    dst_tag_id: number
    dst_node_name: string
    edge_name: string
    src_node_name: string
  }[]
}) => send('forest.EditNode', data) as Promise<void>

/** 重命名实体 */
export const renameNode = (data: {
  graph_id: number
  node_name: string
  old_node_name: string
  tag_id: number
}) => send('forest.RenameNode', data) as Promise<void>

/** 获取实体的边 */
export const getNodeEdges = (data: {
  graph_id: number
  node_name: string
  tag_id: number
}) =>
  send('forest.GetNodeEdges', data) as Promise<{
    edges: {
      chunk_id_list: string
      dst_node_id: number
      dst_node_name: string
      dst_tag_id: number
      edge_id: number
      edge_name: string
      file_id_list: string
      properties: {
        comment: string
        defaults: null | string | number | boolean
        name: string
        type: string
      }[]
      properties_values: {
        chunk_ids: string
        name: string
        value: null | string | number | boolean
      }[]
      src_node_id: number
      src_node_name: string
      src_tag_id: number
    }[]
  }>

/** 获取图谱的所有边 */
export const getGraphEdges = (data: { graph_id: number }) =>
  send('forest.GetGraphEdges', data) as Promise<{
    edges: {
      chunk_id_list: string
      dst_node_id: number
      dst_node_name: string
      dst_tag_id: number
      edge_id: number
      edge_name: string
      file_id_list: string
      properties: {
        comment: string
        defaults: null | string | number | boolean
        name: string
        type: string
      }[]
      properties_values: {
        chunk_ids: string
        name: string
        value: null | string | number | boolean
      }[]
      src_node_id: number
      src_node_name: string
      src_tag_id: number
    }[]
  }>

/** 获取包含节点的文件以及相关chunk */
export const getNodeReference = (data: {
  graph_id: number
  node_name: string
  tag_id: number
}) =>
  send('forest.GetNodeReference', data) as Promise<{
    node_name: string
    tags: {
      tag_id: number
      tag_name: string
      properties: {
        comment: string
        defaults: null | string | number | boolean
        name: string
        type: string
      }[]
      properties_values: {
        chunk_ids: string
        name: string
        value: null | string | number | boolean
      }[]
    }[]
    edges: {
      chunk_id_list: string
      dst_node_id: number
      dst_node_name: string
      dst_tag_id: number
      edge_id: number
      edge_name: string
      file_id_list: string
      properties: {
        comment: string
        defaults: null | string | number | boolean
        name: string
        type: string
      }[]
      properties_values: {
        chunk_ids: string
        name: string
        value: null | string | number | boolean
      }[]
      src_node_id: number
      src_node_name: string
      src_tag_id: number
    }[]
    files: {
      file_id: number
      file_name: string
      file_url: string
      chunks: {
        abstract: string
        chunk_size: number
        company_id: number
        content: string
        content_source: string
        content_target: string
        created_at: string
        description: string
        description_hash: string
        embedding: number[]
        file_id: number
        file_name: string
        forest_id: number
        formula: string
        image_embedding: number[]
        image_url: string
        level: number
        location: number[]
        mind_map: string
        qa_answer: string
        qa_answer_id: string
        qa_lable: string[]
        qa_main_id: string
        qa_question: string
        questions: string[]
        references: {
          chunk_id: string
          description: string
          file_id: number
          relationship_id: string
        }[]
        sequence: number
        source_from: string
        title_level: string[]
        title_level_ids: string[]
        tokens: number
        type: string
        uin: number
        updated_at: string
        version: string
        yg_location: string
      }[]
    }[]
  }>
