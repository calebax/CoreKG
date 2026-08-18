import { AIDialog } from '../../AIDialog'

/** 从流中每次获取的对话数据 */
export type StreamChunk =
  | {
      /** 回答结束 不携带数据 */
      flag: undefined
    }
  | {
      content: string
      /** 标识本条回复的类型 可以表示 正在思考、正在搜索、找到的文件、普通文本、echarts图表、sql代码 */
      flag: Exclude<AIDialog['status'], undefined> | 'echarts' | 'sql'
      /** 找到文件的信息 */
      reference: {
        forest_id: number
        file_id: number
        file_name: string
      }
      // 特殊模型的思考过程
      reasoning_seconds: number
      reasoning_content: string
      code?: number
    }

/** 允许onLoadChunk函数接受不同类型的数据 */
export type onLoadChunkArgs =
  | ['content', StreamChunk]
  | ['history', undefined]
  | ['limited', undefined] // 额度受限

/**
 * 从流中获取回答
 * @param onLoadChunk 接受不同类型的chunk
 * @example
 * ```js
 * const [dialog,setDialog]=useImmer<DialogList>([])
 * const [isAnswering,setAnswering]=useState(false)
 * const sendMsg = (msg:string)=>{
 *    setAnswering(false)
 *    let currentIndex:number
 *    setDialog(draft=>{
 *      draft.pusn({role:'question',xx},{role:'answer',xxx})
 *      currentIndex = draft.length-1
 *    })
 *    const last = dialog
 *    const { body } = await sendStream(xxx)
 *    getChunkFromStream(body,
 *    chunk=>{
 *      updateDialog(chunk,setDialog,currentIndex)
 *    },
 *    ()=>{
 *      setAnswering(false)
 *    }
 *   )
 * }
 * ```
 */
export const getChunkFromStream = async (
  stream: ReadableStream,
  onLoadChunk: (...args: onLoadChunkArgs) => void,
  onEnd?: () => void,
) => {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  while (true) {
    const { done, value } = await reader.read()
    const text = decoder.decode(value, { stream: true })
    const messages = text.split('\n').filter(Boolean)
    messages.forEach((msg) => {
      try {
        // 如果是重新回答的历史记录
        if (msg.startsWith('history')) {
          // 使用单独的'history' 标记下方是历史记录
          // 会将历史记录和新的chunk以普通chunk的方式传输
          onLoadChunk('history', undefined)
          return
        }
        const value = JSON.parse(msg)
        if (value.code === 400) {
          onLoadChunk('limited', undefined)
        }
        onLoadChunk('content', value)
      } catch {
        return
      }
    })
    if (done) {
      onEnd?.()
      return
    }
  }
}
