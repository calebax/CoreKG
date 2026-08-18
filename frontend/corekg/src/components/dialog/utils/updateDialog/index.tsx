import { Updater } from 'use-immer'
import { DialogList } from '../..'
import { StreamChunk } from '../getChunkFromStream'

/**
 * 将流的数据更新至dialog中
 * @param setDialog 来自useImmer
 * @param index 更新的索引
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
export const updateDialog = (
  chunk: StreamChunk,
  setDialog: Updater<DialogList>,
  index: number,
  handleOpenReference?: (index: number) => void,
) => {
  setDialog((draft) => {
    const currentDialog = draft[index]
    if (currentDialog?.role !== 'answer') return
    const { flag } = chunk
    if (flag === undefined) {
      currentDialog.status = 'answered'
      return
    }
    const { content, reference, reasoning_content } = chunk
    switch (flag) {
      case 'search':
        currentDialog.status = 'search'
        break
      case 'found':
        if (reference.file_id) {
          currentDialog.status = 'found'
          currentDialog.reference.push({
            ...reference,
            file_name: reference.file_name,
          } as any)

          handleOpenReference?.(index)
        }
        break
      case 'answering':
        currentDialog.status = 'answering'
        break
      case '':
        if (reasoning_content) {
          currentDialog.thinkingContent += reasoning_content
        } else if (content) {
          currentDialog.content += content
        }
        break
      case 'echarts':
        // 处理echarts图表数据
        if (content) {
          currentDialog.content += content
        }
        break
      case 'sql':
        // 处理sql代码块数据
        if (content) {
          currentDialog.content += content
        }
        break
      case 'thinking':
      case 'answered':
      case undefined:
      default:
        return
    }
  })
}
