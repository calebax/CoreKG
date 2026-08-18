import { DialogList } from '../..'

/** 从后端获取的对话历史 */
export type QAHistory = {
  /** 原有字段直接在对象中 */
  ID?: number
  session_id?: number
  question?: string
  answer?: string
  reasoning?: string
  image_url_list?: string[]
  query_reference_list?: {
    file_id: number
    filename: string
    forest_id: 41
  }[]
  /** 新增的_source字段包装原有数据 */
  _source?: {
    /** 问题的id */
    req_id: string
    session_id: number
    /** 用户提问 */
    question?: string
    /** ai回答 */
    answer?: string
    /** 深度思考 */
    reasoning?: string
    /** 提问时上传的图片 */
    image_url_list?: string[]
    /** 参考文献 */
    query_reference_list?: {
      file_id: number
      file_name: string
      forest_id: 41
    }[]
    chat_reference_list?: any[]
    status?: string
    created_at?: string
    sub_question?: string[]
    [key: string]: any
  }
  [key: string]: any
}[]

/**
 * 将对话历史转化为对话数据
 * @example
 * ```js
 * const [dialog,setDialog]=useImmer<DialogList>()
 * const res = await listChat({ file_id, forest_id })
 * setDialog(getDialogFromHistory(res.Data))
 * ```
 */
export const getDialogFromHistory = (history: QAHistory) => {
  const historyDialog: DialogList = []
  history.forEach((item) => {
    // 兼容新旧数据结构：优先使用_source中的数据，如果没有则使用原有字段
    const data = item._source || item

    historyDialog.push(
      {
        created_at: (data as any).created_at,
        role: 'question',
        content: data.question ?? '',
        images: data.image_url_list || [],
      },
      {
        role: 'answer',
        thinkingContent: data.reasoning ?? '',
        content: data.answer ?? '',
        status: 'answered',
        reference: data.query_reference_list ?? [],
        graph: {
          graph_chat_reference: data.graph_chat_reference,
          graph_reference: data.graph_reference,
        },
      },
    )
  })
  return historyDialog
}
