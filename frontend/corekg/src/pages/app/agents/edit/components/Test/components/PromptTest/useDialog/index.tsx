import { App } from 'antd'
import { useMemoizedFn } from 'ahooks'
import { match } from 'ts-pattern'
import { useImmer } from 'use-immer'
import { testPromptTypeAgent, testWorkflow } from '@/api'
import { DialogList } from '@/components/dialog'
import { useAddLimitedAnswer } from '@/components/dialog/utils'
import { getChunkFromStream } from '@/components/dialog/utils/getChunkFromStream'
import { updateDialog } from '@/components/dialog/utils/updateDialog'
import { useEditContext } from '../../../../AgentContext'

export const useDialog = () => {
  const { agent } = useEditContext()
  const { message } = App.useApp()
  const [dialog, setDialog] = useImmer<DialogList>([])
  const [isAnswering, setAnswering] = useState(false)
  const addLimitedAnswer = useAddLimitedAnswer(setDialog)
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
          onEnd?.()
          setAnswering(false)
          setDialog((draft) => {
            const currentDialog = draft[index]
            if (currentDialog?.role !== 'answer') return
            currentDialog.status = 'answered'
          })
        },
      )
    },
    [addLimitedAnswer, setDialog],
  )

  const startQA = useMemoizedFn(
    async (
      data: Partial<{
        chat_model_ids: number[]
        input: {
          name: string
          value: string
        }[]
        prompt_template: string
        temperature: number
      }>,
      option?: { onEnd?: () => void },
    ) => {
      const { onEnd } = option ?? {}
      setAnswering(true)
      let currentDialogIndex: number
      setDialog((draft) => {
        draft.push({
          role: 'answer',
          content: '',
          thinkingContent: '',
          status: 'thinking',
          reference: [],
        })
        currentDialogIndex = draft.length - 1
      })
      if (agent.type !== 'workflow') {
        const { body } = (await testPromptTypeAgent(data)) as any
        accpetAnswer(body, currentDialogIndex!, onEnd)
      } else {
        const { coze_workflow_id, coze_space_id } = agent
        try {
          const res = await testWorkflow({
            ...data,
            coze_workflow_id,
            coze_space_id,
          })
          if (res.message) {
            message.warning(res.message)
            return
          } else {
            const { output } = res
            setDialog((draft) => {
              const target = draft[currentDialogIndex]
              if (target.role == 'answer') {
                target.content = output
                target.status = 'answered'
              }
            })
            onEnd?.()
          }
        } finally {
          setDialog((draft) => {
            const target = draft[currentDialogIndex]
            if (target.role == 'answer') {
              target.status = 'answered'
            }
          })
          setAnswering(false)
        }
      }
    },
  )

  return {
    isAnswering,
    dialog,
    setDialog,
    startQA,
  }
}
