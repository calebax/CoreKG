import { useBoolean, useMemoizedFn, useMount } from 'ahooks'
import { useImmer } from 'use-immer'
import { listChat, startChat } from '@/api/knowledge'
import { DialogList } from '@/components/dialog'
import { useAddLimitedAnswer } from '@/components/dialog/utils'
import { getChunkFromStream } from '@/components/dialog/utils/getChunkFromStream'
import { getDialogFromHistory } from '@/components/dialog/utils/getDialogFromHistory'
import { updateDialog } from '@/components/dialog/utils/updateDialog'

/** 单次提问所需的数据 */
export type QAData = {
  text: string
  images?: string[]
}

/** 新提问所需的数据 */
export type InitData = QAData & {
  /** 选中的知识库 */
  docs: number[]
}

/** 单文件问答对话 */
export const useFileDialog = (file_id: number, forest_id: number) => {
  const [dialog, setDialog] = useImmer<DialogList>([])
  const addLimitedAnswer = useAddLimitedAnswer(setDialog)
  const [isAnswering, setAnswering] = useState(false)

  const [historyLoading, { setFalse: stopLoading }] = useBoolean(true)
  useMount(async () => {
    const res = await listChat({ file_id, forest_id })
    setDialog(getDialogFromHistory(res.Data ?? []))
    stopLoading()
  })

  // 开启问答
  const startQA = useMemoizedFn(async (data: QAData) => {
    setAnswering(true)
    let currentDialogIndex: number
    setDialog((draft) => {
      draft.push(
        { role: 'question', content: data.text, images: data.images },
        {
          role: 'answer',
          content: '',
          thinkingContent: '',
          status: 'thinking',
          reference: [],
        },
      )
      currentDialogIndex = draft.length - 1
    })
    const { body } = (await startChat({
      file_id,
      forest_id,
      question: data.text,
    })) as any

    getChunkFromStream(
      body,
      (type, data) => {
        if (type === 'content') {
          updateDialog(data, setDialog, currentDialogIndex)
        } else if (type === 'limited') {
          addLimitedAnswer(currentDialogIndex)
        } else {
          setDialog((draft) => {
            draft[currentDialogIndex] = {
              role: 'answer',
              content: '',
              thinkingContent: '',
              status: 'thinking',
              reference: [],
            }
          })
        }
      },
      () => {
        setAnswering(false)
        setDialog((draft) => {
          const currentDialog = draft[currentDialogIndex]
          if (currentDialog?.role !== 'answer') return
          currentDialog.status = 'answered'
        })
      },
    )
  })
  return { historyLoading, isAnswering, dialog, setDialog, startQA }
}
