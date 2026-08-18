import { useMemoizedFn } from 'ahooks'
import { useImmer } from 'use-immer'
import { testRoleTypeAgent, testForestTypeAgent } from '@/api'
import { DialogList } from '@/components/dialog'
import { useAddLimitedAnswer } from '@/components/dialog/utils'
import { getChunkFromStream } from '@/components/dialog/utils/getChunkFromStream'
import { updateDialog } from '@/components/dialog/utils/updateDialog'

export const useDialog = () => {
  const [dialog, setDialog] = useImmer<DialogList<{ withKnowledge?: boolean }>>(
    [],
  )
  const addLimitedAnswer = useAddLimitedAnswer(setDialog)
  const [isAnswering, setAnswering] = useState(false)

  /** 从后端流式接受回答 并更新至对话中 */
  const accpetAnswer = useCallback(
    (stream: ReadableStream, index: number, onEnd?: () => void) => {
      getChunkFromStream(
        stream,
        (type, data) => {
          if (type === 'content') {
            updateDialog(data, setDialog, index)
          } else if (type === 'limited') {
            addLimitedAnswer(index)
          } else {
            setDialog((draft) => {
              draft[index] = {
                role: 'answer',
                content: '',
                thinkingContent: '',
                status: 'thinking',
                reference: [],
              }
            })
          }
        },
        async () => {
          setAnswering(false)
          setDialog((draft) => {
            const currentDialog = draft[index]
            if (currentDialog?.role !== 'answer') return
            currentDialog.status = 'answered'
          })
          onEnd?.()
        },
      )
    },
    [addLimitedAnswer, setDialog],
  )

  const startQA = useMemoizedFn(
    async (
      data: {
        chat_model_ids: number[]
        greeting_message: string
        forest_ids?: number[]
        prompt_template: string
        temperature: number
        question: string
      },
      options?: {
        onEnd?: () => void
      },
    ) => {
      const { onEnd } = options ?? {}
      setAnswering(true)
      let currentDialogIndex: number
      setDialog((draft) => {
        draft.push(
          { role: 'question', content: data.question },
          {
            role: 'answer',
            content: '',
            thinkingContent: '',
            status: 'thinking',
            reference: [],
            withKnowledge: data.forest_ids && data.forest_ids.length > 0,
          },
        )
        currentDialogIndex = draft.length - 1
      })
      const fn = data.forest_ids ? testForestTypeAgent : testRoleTypeAgent
      const { body } = (await fn(data)) as any
      accpetAnswer(body, currentDialogIndex!, onEnd)
    },
  )

  return {
    isAnswering,
    dialog,
    setDialog,
    startQA,
  }
}
