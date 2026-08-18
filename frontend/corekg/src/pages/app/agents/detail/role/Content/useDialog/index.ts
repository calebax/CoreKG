import { useEffect, useState, useRef } from 'react'
import { useMemoizedFn, useMount } from 'ahooks'
import { useImmer } from 'use-immer'
import { createSession, createStream, sendStream, getSessionChats } from '@/api'
import { DialogList } from '@/components/dialog'
import {
  getChunkFromStream,
  getDialogFromHistory,
  updateDialog,
  useAddLimitedAnswer,
} from '@/components/dialog/utils'
import { useHistory } from '../../../components/HistoryContext'

/**
 * 控制角色型智能体问答
 * @param session_id 新会话不传.以第一次传入的数据为准.
 * 已有会话切换时 使用key重置组件
 */
export const useDialog = (
  agentDetail: any,
  session_id: number | undefined,
  setSessionId: (id: number) => void,
  setAnswering?: (val: boolean) => void,
) => {
  const [dialog, setDialog] = useImmer<DialogList>([
    {
      role: 'answer',
      content: agentDetail?.agent_info?.greeting_message || '',
      thinkingContent: '',
      status: 'answered',
      reference: [],
    },
  ])
  const addLimitedAnswer = useAddLimitedAnswer(setDialog)
  const sessionIdRef = useRef<number>()
  const [historyLoading, setHistoryLoading] = useState(true)

  useMount(async () => {
    if (session_id) {
      sessionIdRef.current = session_id
      const res = await getSessionChats({ id: session_id })
      const historyDialog = getDialogFromHistory(res.Data ?? [])
      setDialog(historyDialog.slice(1))
      setHistoryLoading(false)
    } else {
      setHistoryLoading(false)
    }
  })

  const {
    value: history,
    setValue: setHistory,
    refresh: refreshHistory,
  } = useHistory()
  const startQA = useMemoizedFn(async (text: string, images?: string[]) => {
    // 需要更新的回答索引
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
    const agentId = agentDetail.agent_info.ID
    const model_id = agentDetail.agent_info.chat_model_ids[0]
    const resource_type =
      agentDetail.agent_info.agent_type === 'knowledge' ? 'forest' : 'agent'
    setAnswering?.(true)
    // 没有会话时 先创建会话
    if (!sessionIdRef.current) {
      const { ID: id, code } = await createSession({
        model_id,
        resource_id: agentId,
        name: 'new chat',
        resource_type,
        source_from: resource_type,
      })
      // 新会话 额度不足 展示联系售前
      if (code) {
        addLimitedAnswer(currentAnswerIndex!)
        return
      }
      if (history) {
        setHistory([{ ID: id, name: '新建会话' }, ...history])
      }
      setSessionId(id)
      // 会话id始终从ref中取得
      sessionIdRef.current = id
    }
    const { question_id, code } = await createStream({
      session_id: sessionIdRef.current,
      question: text,
      base_agent_id: agentId,
    })
    // 已有会话 额度不足
    if (code) {
      addLimitedAnswer(currentAnswerIndex!)
      setAnswering?.(false)
      return
    }
    const res: any = await sendStream({
      question_id,
    })
    getChunkFromStream(
      res.body,
      (type, data) => {
        if (type === 'content') {
          updateDialog(data, setDialog, currentAnswerIndex)
        }
      },
      () => {
        // 当前回答结束
        setAnswering?.(false)
        setDialog((draft) => {
          const currentDialog = draft[currentAnswerIndex]
          if (currentDialog?.role !== 'answer') return
          currentDialog.status = 'answered'
          if (currentAnswerIndex === 2) {
            // 刷新新会话的第一个回答
            refreshHistory()
          }
        })
      },
    )
  })
  return {
    dialog,
    startQA,
    historyLoading,
    setDialog,
  }
}
