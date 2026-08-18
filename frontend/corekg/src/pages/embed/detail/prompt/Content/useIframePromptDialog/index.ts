import { useState } from 'react'
import { useMemoizedFn, useMount } from 'ahooks'
import { useImmer } from 'use-immer'
import {
  getIframeHistory,
  createIframeStream,
  createIframeStreamWithInput,
  sendIframeStream,
} from '@/api/iframe'
import { DialogList } from '@/components/dialog'
import {
  getChunkFromStream,
  getDialogFromHistory,
  updateDialog,
} from '@/components/dialog/utils'
import { type InputParams } from '../index'

export const useIframePromptDialog = (
  agentDetail: any,
  setAnswering?: (val: boolean) => void,
  workflow?: boolean,
) => {
  const [dialog, setDialog] = useImmer<DialogList>([])
  const [historyLoading, setHistoryLoading] = useState(true)
  const [hasStarted, setHasStarted] = useState(false) // 标记是否已开始对话

  // 加载历史对话
  useMount(async () => {
    try {
      if (workflow) return
      const res = await getIframeHistory()
      if (res.messages.data && res.messages.data.length > 0) {
        const historyDialog = getDialogFromHistory(res.messages.data)
        // prompt模式下，如果有历史记录，说明已经开始对话
        if (historyDialog.length > 0) {
          setDialog(historyDialog)
          setHasStarted(true)
        }
      }
    } catch (err) {
      console.error('加载历史对话失败:', err)
    } finally {
      setHistoryLoading(false)
    }
  })

  // 第一次问答，带参数
  const startQAFirst = useMemoizedFn(async (input: InputParams) => {
    setAnswering?.(true)
    setHasStarted(true)

    // 设置初始的AI回答
    setDialog([
      {
        role: 'answer',
        content: '',
        thinkingContent: '',
        status: 'thinking',
        reference: [],
      },
    ])

    try {
      // 创建带参数的流
      const validInput = input
        .filter((item) => item.value)
        .map((item) => ({
          title: item.title,
          value: item.value!,
          name: item.name,
        }))
      const streamResponse = await createIframeStreamWithInput(validInput)
      const streamKey = streamResponse.stream_key

      // 发送流并接收响应
      // 对于带参数的流，将参数转换为文本进行检测
      const inputText = input
        .map((item) => `${item.title}: ${item.value || ''}`)
        .join(', ')
      const response = await sendIframeStream(streamKey, inputText)

      if (
        response &&
        'ok' in response &&
        response.ok &&
        'body' in response &&
        response.body
      ) {
        await getChunkFromStream(
          response.body,
          (type, data) => {
            if (type === 'content') {
              updateDialog(data, setDialog, 0)
            }
          },
          () => {
            // 回答结束
            setAnswering?.(false)
            setDialog((draft) => {
              const currentDialog = draft[0]
              if (currentDialog?.role !== 'answer') return
              currentDialog.status = 'answered'
            })
          },
        )
      } else {
        throw new Error('流响应失败')
      }
    } catch (err) {
      console.error('首次问答失败:', err)
      setAnswering?.(false)

      // 显示错误消息
      setDialog((draft) => {
        const currentDialog = draft[0]
        if (currentDialog?.role !== 'answer') return
        currentDialog.content = '抱歉，处理参数时出现错误，请稍后重试。'
        currentDialog.status = 'answered'
      })
    }
  })

  // 后续问答
  const startQA = useMemoizedFn(async (text: string, images?: string[]) => {
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
      // 创建普通流
      const streamResponse = await createIframeStream(text)
      const streamKey = streamResponse.stream_key

      // 发送流并接收响应
      const response = await sendIframeStream(streamKey, text)

      if (
        response &&
        'ok' in response &&
        response.ok &&
        'body' in response &&
        response.body
      ) {
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
    startQAFirst,
    startQA,
    historyLoading,
    hasStarted,
    setDialog,
  }
}
