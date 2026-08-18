import { useState, useRef } from 'react'
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
import { type InputParams } from '..'
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
  const [dialog, setDialog] = useImmer<DialogList>([])
  const addLimitedAnswer = useAddLimitedAnswer(setDialog)
  const sessionIdRef = useRef<number>()
  const [historyLoading, setHistoryLoading] = useState(true)

  useMount(async () => {
    if (session_id) {
      sessionIdRef.current = session_id
      const res = await getSessionChats({ id: session_id })
      setDialog(getDialogFromHistory(res.Data ?? []).slice(1))
    }
    setHistoryLoading(false)
  })

  const {
    value: history,
    setValue: setHistory,
    refresh: refreshHistory,
  } = useHistory()
  // 第一次问答 用input提问
  const startQAFirst = useMemoizedFn(async (input: InputParams) => {
    const agentId = agentDetail.agent_info.ID
    const model_id = agentDetail.agent_info.chat_model_ids[0]
    setAnswering?.(true)
    // 没有会话时 先创建会话
    const { ID: id } = await createSession({
      model_id,
      resource_id: agentId,
      name: 'new chat',
      resource_type: 'agent',
      source_from: 'agent',
      input,
    }).catch((e) => {
      setAnswering?.(false)
      throw e
    })
    sessionIdRef.current = id
    if (history) {
      setHistory([{ ID: id, name: '新建会话' }, ...history])
    }
    setSessionId(id)
    setDialog([
      {
        role: 'answer',
        content: '',
        thinkingContent: '',
        status: 'thinking',
        reference: [],
      },
    ])
    const { question_id } = await createStream({
      session_id: id,
      question: '',
      base_agent_id: agentId,
    })
    const res: any = await sendStream({
      question_id,
    })

    getChunkFromStream(
      res.body,
      (type, data) => {
        if (type === 'content') {
          updateDialog(data, setDialog, 0)
        }
      },
      () => {
        // 当前回答结束
        setAnswering?.(false)
        setDialog((draft) => {
          const currentDialog = draft[0]
          if (currentDialog?.role !== 'answer') return
          currentDialog.status = 'answered'
          refreshHistory()
        })
      },
    )
  })

  // 后续问答
  const startQA = useMemoizedFn(async (text: string, images?: string[]) => {
    setAnswering?.(true)
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
    const { question_id, code } = await createStream({
      session_id: sessionIdRef.current,
      question: text,
      base_agent_id: agentId,
    })
    // 额度受限
    if (code) {
      setAnswering?.(false)
      addLimitedAnswer(currentAnswerIndex!)
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
        })
      },
    )
  })

  return {
    dialog,
    startQAFirst,
    startQA,
    historyLoading,
    setDialog,
  }
}
