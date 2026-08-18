import config from '@/config'
import { chat } from './chat.js'
import { send, upload } from './request'

// 获取知识库列表
// 参数：limit: 每页条数, offset: 页码
export const getKnowledgeBaseList = (
  data: {
    limit?: number
    offset?: number
    [key: string]: any
  } & CommonArgs,
) => send('forest.ListForest', data)

// 创建知识库
// 参数：img: 图片地址, name: 知识库名称, description: 知识库描述
export const createKnowledgeBase = (data: {
  name: string
  description: string
  forest_type: string
  public_scope: 'company' | 'private' | 'custom'
  data_source_type: 'standard' | 'excel' | 'db'
  data_source_subtype: 'standard' | 'excel' | 'mysql'
}) => send('forest.CreateForest', data)

/** 全量更新知识库信息 新版本 包含scope_ids */
export const updateForestWithPerm = (data: {
  id: number
  name: string
  description: string
  manager_ids: number[]
  public_scope: string
  scope_ids?: number[]
  data_source_type: 'standard' | 'excel' | 'db'
  data_source_subtype: 'standard' | 'excel' | 'mysql'
}) => send('forest.UpdateForestWithPerm', data)

// 更新知识库名称
export const updateKnowledgeBaseName = (data: { id: number; name: string }) =>
  send('forest.RenameForest', data)

// 获取知识库详情
// 参数：id: 知识库ID
export const getKnowledgeBaseDetail = (data: { id: number }) =>
  send('forest.GetForest', data)

// 删除知识库
// 参数：id: 知识库ID
export const deleteKnowledgeBase = (data: { id: number }) =>
  send('forest.DeleteForest', data)

// 同步知识库至coze
export const syncKnowledgeBaseToCoze = (data: {
  detail_id: number
  name: string
  forest_type: string
}) => send('forest.CreateCozeKnowledge', data)

// ============= 文件列表相关 API =============

// 获取文件列表
// 参数：forest_id: 知识库ID, limit: 每页条数, offset: 页码, filters: 过滤条件(包含parent_id过滤), orderBy: 排序条件, beginTime: 开始时间, endTime: 结束时间
export const getFileList = (data: {
  forest_id: number
  limit?: number
  offset?: number
  filters?: Array<{ field: string; value: string[]; exactMatch?: boolean }>
  orderBy?: string[]
  beginTime?: string
  endTime?: string
  listAll?: boolean
  image_url?: string
}) => {
  // 固定按更新时间倒序排序，确保新内容优先展示
  const finalData = {
    ...data,
    orderBy: ['created_at desc'],
  }
  return send('forest.ListFile', finalData)
}

// 创建文件夹
// 参数：forest_id: 知识库ID, name: 文件夹名称, parent_id: 父文件夹ID
export const createFolder = (data: {
  forest_id: number
  name: string
  parent_id: number
}) => send('forest.CreateDir', data)

// 上传文件
// 参数：forest_id: 知识库ID, parent_id: 父文件夹ID, file: 文件对象
export const uploadFile = (
  data: { forest_id: number; parent_id: number; file: File },
  config?: any,
) => upload('forest.UploadFile', data, config)

// 删除文件夹/文件
export const deleteFile = (data: { file_id: Array<number> }) =>
  send('forest.DeletePath', data)

// 重命名文件夹/文件
export const renameFile = (data: { file_id: number; new_name: string }) =>
  send('forest.RenamePath', data)

// 移动文件夹/文件
export const moveFile = (data: { file_id: number; dst_parent_id: number }) =>
  send('forest.MovePath', data)

// ============= 文件详情相关 API =============

// 获取文件预览 URL
export const getPreviewFileURL = (data: {
  file_id: number
  is_download?: boolean
}) => send('forest.PreviewFileByURL', data)

// 获取文件信息
export const getFileInfo = (data: { file_id: number }) =>
  send('forest.GetFileInfo', data)

/** 获取文件的上级文件夹 */
export const getFilePath = (data: { forest_id: number; file_id: number }) =>
  send('forest.GetFilePath', data)

// 获取文档解析内容
export const getDocContent = (data: { file_id: number }) =>
  send('forest.GetContent', data)

// 获取智能分析内容
export const getIntelContent = (data: { file_id: number }) =>
  send('forest.GetAnalysis', data)

// 获取思维导图内容
export const getMindMapContent = (data: { file_id: number }) =>
  send('forest.GetMindMap', data)

// 保存文档解析
export const saveDocParse = (data: {
  file_id: number
  forest: string
  content: string
}) => send('forest.SaveParsedFileContent', data)

// 保存智能分析
export const saveIntelParse = (data: {
  file_id: number
  forest: string
  analysis: string
}) => send('forest.SaveParsedFileAnalysis', data)

// 导出文档解析
export const exportDocParse = (data: { file_id: number; forest_id: number }) =>
  send('forest.ExportParsedFile', data)

// 导出智能分析
export const exportIntelParse = (data: {
  file_id: number
  forest_id: number
}) => send('forest.ExportAnalysisFile', data)

// ============= 智慧问答相关 API =============

// 准备知识库聊天
export const prepareChat = (data: { file_id: number; forest_id: number }) =>
  send('forest.FileQAPrepare', data)

// 获取问答对话记录列表
export const listChat = (data: { file_id: number; forest_id: number }) =>
  send('chat.ListFileQA', data)

// 开始问答对话
export const startChat = (data: {
  file_id: number
  forest_id: number
  question: string
}) => chat('chat.FileChat', data)

// 获取预制问题
export const releaseChat = (data: { file_id: number; forest_id: number }) =>
  send('forest.FileQARelease', data)

// 清除对话记忆
export const deleteFileQA = (data: { file_id: number; forest_id: number }) =>
  send('chat.DeleteFileQA', data)

export const RecentlyForest = (body: any) => send('forest.RecentlyForest', body)

// 获取知识库词云数据
export const getKnowledgeBaseWordCloud = (knowledgeBaseId: number) => {
  return send('forest.GetForestWordCloud', {
    forest_id: knowledgeBaseId,
  })
}

// 获取知识库知识图谱数据
export const getKnowledgeBaseGraph = (knowledgeBaseId: number) => {
  return send('forest.GetForestWordCloudGraph', {
    forest_id: knowledgeBaseId,
  })
}

// 获取节点相关的图谱数据
export const getNodesByID = (params: {
  knowledgeBaseId: number
  nodeId: string
}) => {
  return send('forest.GetNodesByID', {
    forest_id: params.knowledgeBaseId,
    node_id: params.nodeId,
  })
}

/**
 * 知识库搜索:全部类型
 * is_semantics:false或不传表示联想
 * 为true表示语义搜索(用户回车搜索)
 */
export const forestSearch = (data: {
  forest_id: number
  is_semantics?: boolean
  text: string
  image_url?: string
}) => send('kesearch.ForestSearch', data)

/** 知识库搜索:文档类型 */
export const forestSearchDoc = (data: {
  forest_id: number
  is_semantics?: boolean
  text: string
  image_url?: string
}) => send('kesearch.ForestSearchDoc', data)

/** 知识库搜索:图片类型 */
export const forestSearchImage = (data: {
  forest_id: number
  is_semantics?: boolean
  text: string
  image_url?: string
}) => send('kesearch.ForestSearchImage', data)

/** 知识库搜索:视频类型 */
export const forestSearchVideo = (data: {
  forest_id: number
  is_semantics?: boolean
  text: string
  image_url?: string
}) => send('kesearch.ForestSearchVideo', data)

/**
 * 全局搜索
 */
export const globalSearch = (data: {
  text: string
  is_semantics?: boolean
  image_url?: string
  forest_ids?: number[]
}) => send('kesearch.GlobalSearch', data)

/**
 * 全局搜索智能体类型
 */
export const globalSearchAgent = (data: {
  text: string
  is_semantics?: boolean
  image_url?: string
}) => send('kesearch.GlobalSearchAgent', data)

/**
 * 全局搜索文档类型
 */
export const globalSearchDoc = (data: {
  text: string
  is_semantics?: boolean
  image_url?: string
}) => send('kesearch.GlobalSearchDoc', data)

/**
 * 全局搜索知识库
 */
export const globalSearchForest = (data: {
  text: string
  is_semantics?: boolean
  image_url?: string
}) => send('kesearch.GlobalSearchForest', data)

/**
 * 全局搜索图片类型
 */
export const globalSearchImage = (data: {
  text: string
  is_semantics?: boolean
  image_url?: string
}) => send('kesearch.GlobalSearchImage', data)

/**
 * 全局搜索视频类型
 */
export const globalSearchVideo = (data: {
  text: string
  is_semantics?: boolean
  image_url?: string
}) => send('kesearch.GlobalSearchVideo', data)

/**
 * 点赞/取消点赞文档
 */
export const toggleDocLike = (data: {
  resource_id: number
  resource_type: string
  enable: boolean
}) => send('forest.MarkResourceLike', data)

/**
 * 收藏/取消收藏文档
 */
export const toggleDocCollect = (data: {
  resource_id: number
  resource_type: string
  enable: boolean
}) => send('forest.MarkResourceCollection', data)

/**
 * 获取点赞列表
 */
export const getLikeList = (data: { limit?: number; offset?: number }) =>
  send('forest.ListLikes', {
    ...data,
    filters: [{ field: 'resource_type', value: ['forest_file'] }],
  })

/**
 * 获取收藏列表
 */
export const getCollectionList = (data: { limit?: number; offset?: number }) =>
  send('forest.ListCollection', {
    ...data,
    filters: [{ field: 'resource_type', value: ['forest_file'] }],
  })

/**
 * 创建问答对
 */
export const createQAPair = (data: {
  forest_id: number
  question: string
  answer: string
  sub_question?: string[]
}) => send('forest.CreateQAPair', data)

/**
 * 删除问答对
 */
export const deleteQAPair = (data: {
  forest_id: number
  question_ids: string[]
}) => send('forest.DeleteQAPair', data)

/**
 * 修改问答对
 */
export const modifyQAPair = (data: {
  main: {
    id: string
    forest_id: number
    type: string
    qa_question: string
    qa_answer: string
  }
  child: Array<{
    id?: string
    question: string
    is_deleted: boolean
    created_at?: string
  }>
}) => send('forest.ModifyQAPair', data)

/**
 * 获取问答对列表
 */
export const listQAPair = (data: {
  forest_id: number
  limit: number
  offset: number
  filters?: any[]
  orderBy?: string[]
  listAll?: boolean
}) => send('forest.ListQAPair', data)

/**
 * 预签名上传文件 获取临时的url
 */
export const getPresignedUrl = (
  data: {
    forest_id: number
    parent_id: number
    files: {
      filename: string
      hash: string
      size: number
      content_type: string
      split_config: {
        split_mode: 'auto' | 'rule'
        chunk_size?: number
        split_mark?: Array<string>
        split_overlap?: number
        preprocessing_rules?: {
          remove_empty_line: boolean
          remove_URL: boolean
          remove_email: boolean
        }
      }
    }[]
  },
  config?: any,
) => send('forest.PreUploadFile', data, config)

/** 预签名上传结束后的回调接口 */
export const PresignedUploadFileCallBack = (data: {
  forest_id: number
  upload_id: any
  filename: string
  hash: string
  /** 分片上传各片的信息 */
  parts?: { part_number: number; etag: string }[]
}) => send('forest.UploadFileCallBack', data)

/** 取消一个上传 */
export const abortUpload = (data: { upload_id: any; hash: string }) =>
  send('forest.AbortUpload', data)
/** 上传url失效后获取新的url */
export const getNewUploadUrl = (data: {
  upload_id: any
  hash: string
  /** 分片上传时 提供已上传的片 */
  completed_parts?: number[]
  /** 需要的片 */
  expired_parts?: number[]
}) => send('forest.RenewUploadUrl', data)

/** 单文档问答推荐问题 */
export const getRecommendQuestions = (data: { file_id: number }) =>
  send('forest.GetRecommendQuestions', data)

/** 上传问答对excel */
export const uploadQAPair = (data: { file: File; forest_id: number }) =>
  upload('forest.UploadQAPair', data, {
    timeout: 0,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
/** 提交经过解析和用户确认后的问答对 */
export const commitQAPair = (data: {
  forest_id: number
  qa_list: {
    answer: string
    question: string
  }[]
}) => send('forest.CommitQAPair', data)

/** 获取excel文件的sheet */
export const getExcelSheet = (data: { forest_file_ids: number[] }) =>
  send('forest.ListExcelSheet', data)

// ============= 文件分段(Chunk)管理相关 API =============

/**
 * 获取文件分段列表
 * @param data - 包含 file_id
 * @returns 分段列表数据
 */
export const getFileSegments = (data: { file_id: number; forest_id: number }) =>
  send('kesearch.ListFileChunk', data)

/**
 * 更新分段内容
 * @param params - 包含 segment_id 和 content
 * @returns 更新结果
 */
export const updateFileSegment = (data: {
  chunk_id: string
  description?: string
  table?: string
  file_id: number
}) => send('kesearch.UpdateChunk', data)

/**
 * 删除分段
 * @param params - 包含 segment_id
 * @returns 删除结果
 */
export const deleteFileSegment = (data: {
  chunk_id: string
  file_id: number
}) => send('kesearch.DeleteChunk', data)

/** 重试解析 */
export const retryParse = (id: number) => send('forest.RetryParse', { id })
/** 获取5天内的解析历史 */
export const listParseHistory = (data: CommonArgs) =>
  send('forest.ListParseHistory', data)

/**
 * 修改文件分段规则
 * @param data - 包含文件ID和新的分段规则配置
 * @returns 修改结果
 */
export const modifyFileSegmentRule = (data: {
  file_id: number
  forest_id: number
  split_config: {
    split_mode: 'auto' | 'rule'
    chunk_size?: number
    split_mark?: Array<string>
    split_overlap?: number
    preprocessing_rules?: {
      remove_empty_line: boolean
      remove_url: boolean
      remove_email: boolean
    }
  }
}) => send('forest.ResplitChunk', data)

/** 获取知识库、数据库或者表的名称 */
export const getNameByModuleIDs = (data: {
  module_id_list: {
    ids: number[]
    module: 'forest' | 'database' | 'table'
  }[]
}) => send('forest.GetNameByModuleIDs', data)

/** 数据库基本信息 */
export type DatabaseInfo = {
  /** 数据库名称 */
  database: string
  host: string
  password: string
  port: number
  username: string
}
/** 创建数据库实例 */
export const createForestDBInstance = (
  data: DatabaseInfo & {
    forest_id: number
  },
) => send('forest.CreateForestDBInstance', data)
/** 获取数据库实例详情 */
export const getForestDBInstance = (data: { forest_id: number }) =>
  send('forest.GetForestDBInstance', data)
/** 修改数据库实例 */
export const modifyForestDBInstance = (
  data: DatabaseInfo & {
    forest_id: number
  },
) => send('forest.ModifyForestDBInstance', data)
/** 测试数据库连接 */
export const testForestDBInstanceConnection = (
  data: DatabaseInfo & {
    forest_id: number
  },
) => send('forest.TestForestDBInstanceConnection', data)
/** 获取数据库列表 */
export const listForestDB = (data: { forest_id: number } & CommonArgs) =>
  send('forest.ListForestDB', data)
/** 获取数据库表列表 */
export const listForestTable = (
  data: { forest_id: number; forest_db_name: string } & CommonArgs,
) => send('forest.ListForestTable', data)
/** 获取数据库表头 */
export const getForestTableHeader = (data: {
  forest_id: number
  forest_db_name: string
  forest_table_name: string
}) => send('forest.GetForestTableMetadata', data)

/** 全量获取所有知识库的所有文件、sheet、table等 */
export const getAllForestData = () => send('forest.GetResourceBaseInfo', {})

/**
 * 禁用/启用文件问答功能
 * @param data - 包含 file_id 和 is_disable 参数
 * @returns 操作结果
 */
export const disableFileChunk = (data: {
  file_id: number
  is_disable: boolean
}) => send('kesearch.DisableFileChunk', data)

// 设置资源启用状态
export const updateResourceEnable = (data: {
  enable: number
  forest_id: number
  resource_ids: string[]
  resource_type: string
}) => {
  return send('forest.SetResourceEnable', data)
}

// 更新知识库描述
export const updateKnowledgeBaseDesc = (data: {
  forest_id: number
  description: string
}) => send('forest.UpdateForestDescription', data)

// 获取资源权限
export const getResourcePerm = (data: {
  resource_id: number
  resource_type: string
}) =>
  send('forest.GetResourcePerm', data) as Promise<{
    access_result: {
      BanList: number[]
      ManagerList: number[]
      ScopeType: string
      ViewerList: number[]
    }
  }>

// 设置资源权限
export const setResourcePerm = (data: {
  resource_id: number
  resource_type: string
  perm_option: {
    ban_list: number[]
  }
}) => send('forest.SetResourcePerm', data)

// ============= 标签管理相关 API =============

/** 获取标签分类列表 */
export const listTagGroup = (data: { limit: number; offset: number }) =>
  send('forest.ListTagGroup', data)

/** 创建标签分类 */
export const createTagGroup = (data: { name: string }) =>
  send('forest.CreateTagGroup', data)

/** 修改标签分类 */
export const modifyTagGroup = (data: { name: string; tag_group_id: number }) =>
  send('forest.ModifyTagGroup', data)

/** 删除标签分类 */
export const deleteTagGroup = (data: { tag_group_id: number }) =>
  send('forest.DeleteTagGroup', data)

/** 获取标签列表 */
export const listResourceTag = (data: {
  limit: number
  offset: number
  name?: string
}) => send('forest.ListResourceTag', data)

/** 创建标签 */
export const createResourceTag = (data: {
  name: string
  tag_group_id: number
}) => send('forest.CreateResourceTag', data)

/** 修改标签 */
export const modifyResourceTag = (data: {
  name: string
  tag_group_id: number
  tag_id: number
}) => send('forest.ModifyResourceTag', data)

/** 删除标签 */
export const deleteResourceTag = (data: { tag_id: number }) =>
  send('forest.DeleteResourceTag', data)

/** 获取标签树 */
export const getTagTree = () => send('forest.GetTagTree', {})

// ============= 同义词管理相关 API =============

/** 获取同义词列表 */
export const listSynonymKeywords = (data: {
  limit: number
  offset: number
  word?: string
  filters?: any[]
}) => {
  return send('forest.ListSynonymKeywords', data)
}

/** 创建同义词 */
export const createSynonymKeyword = (data: {
  word: string
  child_words: string[]
}) => send('forest.CreateSynonymKeyword', data)

/** 修改同义词 */
export const updateSynonymKeyword = (data: {
  id: number
  word: string
  child_words: string[]
}) => send('forest.UpdateSynonymKeyword', data)

/** 获取同义词详情 */
export const getSynonymKeyword = (data: { id: number }) =>
  send('forest.GetSynonymKeyword', data)

/** 删除同义词 */
export const deleteSynonymKeyword = (data: { id: number }) =>
  send('forest.DeleteSynonymKeyword', data)

// ============= 行业名词管理相关 API =============

/** 获取行业名词列表 */
export const listIndustryTerms = (data: {
  limit: number
  offset: number
  word?: string
  filters?: Array<{ field: string; value: string[]; exactMatch?: boolean }>
}) => {
  return send('forest.ListMajorKeywords', data)
}

/** 创建行业名词 */
export const createIndustryTerm = (data: {
  word: string
  description: string
}) => send('forest.CreateMajorKeyword', data)

/** 修改行业名词 */
export const updateIndustryTerm = (data: {
  id: number
  word: string
  description: string
}) => send('forest.UpdateMajorKeyword', data)

/** 获取行业名词详情 */
export const getIndustryTerm = (data: { id: number }) => {
  return send('forest.GetMajorKeyword', data)
}

/** 删除行业术语 */
export const deleteIndustryTerm = (data: { id: number }) =>
  send('forest.DeleteMajorKeyword', data)

/** 设置资源标签 */
export const setResourceTag = (data: {
  resource_id: number
  resource_type: string
  tag_ids: number[]
}) => send('forest.SetResourceTag', data)

// 获取知识库问答会话id
export const getFileSession = (file_id: number) =>
  send('chat.GetFileSession', {
    file_id,
  })

// 获取知识库问答project_id
export const getFileQaProject = (file_id?: number) => {
  return send('forest.GetDefaultProject', file_id ? { file_id } : {})
}

// 获取热词
export const getHotWords = () => send('forest.GetHotWords', {})

/**
 * 智能润色
 */
export const expansionQuestion = (data: {
  file_ids: number[]
  question: string
  session_id: number
}) => send('chat.ExpansionQuestion', data)
