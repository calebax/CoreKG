import { useMemoizedFn, useMount } from 'ahooks'
import { useImmer } from 'use-immer'
import {
  getIframeHistory,
  createIframeStream,
  sendIframeStream,
} from '@/api/iframe'
import { DialogList } from '@/components/dialog'
import {
  getChunkFromStream,
  getDialogFromHistory,
  updateDialog,
} from '@/components/dialog/utils'

/**
 * iframe模式下的对话控制
 */
export const useIframeDialog = (
  agentDetail: any,
  setAnswering?: (val: boolean) => void,
) => {
  const [dialog, setDialog] = useImmer<DialogList>([
    {
      role: 'answer',
      content:
        agentDetail.greet_message ||
        '你好！我是您的AI助手，有什么可以帮助您的吗？',
      thinkingContent: '',
      status: 'answered',
      reference: [],
    },
  ])

  const [historyLoading, setHistoryLoading] = useState(true)

  // 加载历史对话
  useMount(async () => {
    try {
      const res = await getIframeHistory()
      if (res.messages.data && res.messages.data.length > 0) {
        const historyDialog = getDialogFromHistory(res.messages.data)
        setDialog(historyDialog)
      }
    } catch (err) {
      console.error('加载历史对话失败:', err)
      // 使用默认的问候消息
    } finally {
      setHistoryLoading(false)
    }
  })

  const startQA = useMemoizedFn(async (text: string, images?: string[]) => {
    // 添加用户问题和AI回答占位符
    let currentAnswerIndex: number
    setDialog((draft) => {
      draft.push(
        {
          role: 'question',
          content: text,
          images,
        },
        {
          role: 'answer',
          content: '',
          thinkingContent: '',
          status: 'thinking',
          reference: [],
        },
      )
      currentAnswerIndex = draft.length - 1
    })

    setAnswering?.(true)

    try {
      // 创建流
      const streamResponse = await createIframeStream(text)
      const streamKey = streamResponse.stream_key

      // 发送流并接收响应
      const response = await sendIframeStream(streamKey, text)

      if (response && 'body' in response && response.body) {
        await getChunkFromStream(
          response.body,
          (type, data) => {
            if (type === 'content') {
              updateDialog(data, setDialog, currentAnswerIndex)
            }
          },
          () => {
            // 回答结束
            setAnswering?.(false)
            setDialog((draft) => {
              const currentDialog = draft[currentAnswerIndex]
              if (currentDialog?.role !== 'answer') return
              currentDialog.status = 'answered'
            })
          },
        )
      } else {
        throw new Error('流响应失败')
      }
    } catch (err) {
      console.error('发送消息失败:', err)
      setAnswering?.(false)

      // 显示错误消息
      setDialog((draft) => {
        const currentDialog = draft[currentAnswerIndex]
        if (currentDialog?.role !== 'answer') return
        currentDialog.content = '抱歉，发送消息时出现错误，请稍后重试。'
        currentDialog.status = 'answered'
      })
    }
  })

  return {
    dialog,
    startQA,
    historyLoading,
    setDialog,
  }
}
