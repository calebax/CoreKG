import {
  globalSearch,
  globalSearchAgent,
  globalSearchDoc,
  globalSearchForest,
  globalSearchImage,
  globalSearchVideo,
} from '@/api/knowledge'

/** 联想和搜索结果的类型 agent:智能体 forest:知识库*/
export type SearchType = 'all' | 'doc' | 'image' | 'video' | 'agent' | 'forest'

/** 从后端获取数据 单个文件 */
export type FileInSearchResult = {
  id: number
  forest_id: number
  /** 根据类型不同 字段名有所差别 */
  highlighted_file_name?: string
  highlighted_forest_name?: string
  highlighted_agent_name?: string
  /** 同一文件的多个搜索结果 */
  highlights?: {
    highlighted_description: string
    image_url?: string
    location: number[]
  }[]
  /** forest知识库 agent应用 只有一个结果 没有highlights 转而提供以下两项 */
  image_url?: string
  highlighted_description?: string
}

/** 不包括all的SearchType */
export type LegalSearchType = Exclude<SearchType, 'all'>

/** 搜索结果展示顺序 */
export const SearchTypeOrder: LegalSearchType[] = [
  'agent',
  'forest',
  'doc',
  'image',
  'video',
]

/** 各项搜索结果及其标题 */
export const SearchType_TitleMap: Record<LegalSearchType, string> = {
  agent: '智能体',
  forest: '知识库',
  doc: '文档',
  image: '图片',
  video: '视频',
}

/** 各项搜索结果在后端返回里的key */
export const SearchType_ResultKeyMap: Record<LegalSearchType, string> = {
  agent: 'agent_search_result',
  forest: 'forest_search_result',
  doc: 'doc_search_result',
  image: 'image_search_result',
  video: 'video_search_result',
}

/** 各类型搜索项对象的数据获取函数 */
export const SearchType_FnMap: Record<
  SearchType,
  (data: {
    text: string
    is_semantics?: boolean
    image_url?: string
  }) => Promise<any>
> = {
  all: globalSearch,
  agent: globalSearchAgent,
  forest: globalSearchForest,
  doc: globalSearchDoc,
  image: globalSearchImage,
  video: globalSearchVideo,
}
