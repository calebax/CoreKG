import { useMemoizedFn, useMount } from 'ahooks'
import { useImmer } from 'use-immer'
import {
  createStream,
  sendStream,
  getSessionChats,
  getSessionInfo,
  continueLastChat,
  stopChat,
} from '@/api/agent'
import { DialogList } from '@/components/dialog'
import { useAddLimitedAnswer } from '@/components/dialog/utils'
import { getChunkFromStream } from '@/components/dialog/utils/getChunkFromStream'
import { getDialogFromHistory } from '@/components/dialog/utils/getDialogFromHistory'
import { updateDialog } from '@/components/dialog/utils/updateDialog'
import { useGlobalSessionHistory } from '@/stores/GlobalSessionHistory'
import { DialogInitData, QAData, SessionInfo } from '../../type'

/**
 * 管理AI对话
 * @param session_id 根据session_id处理对话数据 也可以根据数据创建session
 * @param question_id 如果携带了问题id 视作新会话
 */
export const useDialog = (session_id: number, question_id?: any) => {
  const navigate = useNavigate()
  const { loadData, add } = useGlobalSessionHistory()
  const [dialog, setDialog] = useImmer<DialogList>([])
  const addLimitedAnswer = useAddLimitedAnswer(setDialog)
  // 历史对话加载loading
  const [loading, setLoading] = useState(true)
  const [isAnswering, setAnswering] = useState(false)
  /** 用于停止回答 */
  const lastQuestionId = useRef<number>()
  const [sessionInfo, setSessionInfo] = useState<SessionInfo>()

  /** 从后端流式接受回答 并更新至对话中 */
  const accpetAnswer = useCallback(
    (stream: ReadableStream, index: number) => {
      getChunkFromStream(
        stream,
        (type, data) => {
          if (type === 'content') {
            updateDialog(data, setDialog, index)
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
            // 在完成第一次对话后重新载入历史记录
            if (index === 1) {
              loadData()
            }
          })
        },
      )
    },
    [loadData, setDialog],
  )

  useMount(async () => {
    try {
      const res: any = await getSessionChats({
        id: session_id,
      })
      const sessionInfo = await getSessionInfoWithType(session_id)
      setSessionInfo(sessionInfo)
      const historyDialog = getDialogFromHistory(res.Data ?? [])
      setDialog(historyDialog)
      if (!question_id) {
        // 老会话 继续未完成的回答
        const lastItem = res.Data?.at(-1)
        const _lastQuestionId = lastItem?._id
        if (!_lastQuestionId) {
          setLoading(false)
          setAnswering(false)
          return
        }
        lastQuestionId.current = _lastQuestionId
        const { body } = (await continueLastChat({
          question_id: lastQuestionId.current,
        })) as any
        setLoading(false)
        setAnswering(true)
        accpetAnswer(body, historyDialog.length - 1)
      } else {
        // 新会话 直接提问
        // 从url中去除question_id 避免刷新造成影响
        navigate(`?session_id=${session_id}`, { replace: true })
        lastQuestionId.current = question_id
        const { body } = (await sendStream({
          session_id: session_id,
          question_id,
        })) as any
        add({ id: session_id, name: '', nameLoading: true })
        setLoading(false)
        setAnswering(true)
        accpetAnswer(body, 1)
      }
    } catch (e) {
      console.error(e)
      navigate('/')
    }
  })

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
    const { question_id, code } = (await createStream({
      session_id: session_id,
      question: data.text,
    })) as any
    // 额度不足
    if (code) {
      addLimitedAnswer(currentDialogIndex!)
      setAnswering(false)
      return
    }
    lastQuestionId.current = question_id
    const { body } = (await sendStream({
      session_id: session_id,
      question_id,
    })) as any

    accpetAnswer(body, currentDialogIndex!)
  })

  const stopQA = useMemoizedFn(() => {
    if (!lastQuestionId.current) return
    stopChat({ question_id: lastQuestionId.current })
  })

  return {
    isAnswering,
    /** 历史对话加载中 */
    loading,
    /** 会话数据 */
    sessionInfo,
    dialog,
    setDialog,
    startQA,
    stopQA,
  }
}

const getSessionInfoWithType = async (
  session_id: number,
): Promise<SessionInfo> => {
  type KnowledgeNames<T extends string> = {
    [key in T]: { id: number; name: string }[]
  }
  const { model_id, model_name, resource_type, ...knowledgeNames } =
    (await getSessionInfo({ id: session_id })) as any as {
      model_id: number
      model_name: string
      resource_type: DialogInitData['type']
    } & KnowledgeNames<
      | 'forest_names'
      | 'file_names'
      | 'excel_names'
      | 'excel_sheet_names'
      | 'db_names'
      | 'db_table_names'
    >
  const knowledge =
    (() => {
      switch (resource_type) {
        case 'excel_list':
          return knowledgeNames.excel_names
        case 'file_list':
          return knowledgeNames.file_names
        case 'forest':
          return knowledgeNames.forest_names
        case 'react_excel_list':
          return knowledgeNames.excel_sheet_names
        case 'db_list':
          return knowledgeNames.db_names
        case 'db_table_list':
          return knowledgeNames.db_table_names
      }
    })() || []

  return {
    model: model_id,
    modelName: model_name,
    type: resource_type,
    knowledge,
  }
}
