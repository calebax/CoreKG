import { useRef, useState, useEffect } from 'react'
import { useAsyncEffect, useMemoizedFn } from 'ahooks'
import dayjs from 'dayjs'
import { produce } from 'immer'
import { useImmer } from 'use-immer'
import {
  createSession,
  createStream,
  sendStream,
  getSessionChats,
  getSessionInfo,
  continueLastChat,
  stopChat,
  getQuestionInfo,
} from '@/api/agent'
import { getFileList } from '@/api/knowledge'
import { DialogList } from '@/components/dialog'
import { updateDialog, useLimitedAnswer } from '@/components/dialog/utils'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { SessionConfig, QAData, SessionInfo, Knowledge } from '..'
import { useProject } from '../..'
import { getColorArray } from '../../constant'
import { getChunkFromStream, getDialogFromHistory } from './utils'

const SESSION_DIALOG_CACHE_UPDATED = 'projectSessionDialogCacheUpdated'
const sessionDialogCache = new Map<number, DialogList>()
const sessionDialogStatusCache = new Map<number, SessionInfo['dialogStatus']>()
const activeStreamSessionIds = new Set<number>()
const streamingQuestionIds = new Map<number, any>()

export const useSession = (
  projectEmpty: boolean,
  handleOpenReference?: (index: number) => void,
  handleOpenChart?: () => void,
  handleOpenGraph?: (index: number) => void,
) => {
  const {
    project_id,
    session_id,
    setSessionId,
    setData,
    reloadKnowledge,
    reloadSessions,
    chartsOperators,
    isOtherPage,
    defaultKnowBase,
    type,
  } = useProject()
  /** 自行维护session状态 由外部修改 不重置 */
  const [sessionStatus, setSessionStatus] = useState<
    SessionInfo['sessionStatus']
  >(() => {
    if (projectEmpty) return 'new'
    if (type === 'single-file') return 'new'
    if (type === 'global') return 'new'
    return session_id ? 'created' : isOtherPage ? 'new' : 'none'
  })

  // 当项目为空（新建会话分组）时，确保 sessionStatus 为 'new'
  // 当 project_id 变化时，如果项目为空且没有 session_id，设置为 'new'
  useEffect(() => {
    if (projectEmpty && !session_id) {
      if (sessionStatus !== 'new') {
        setSessionStatus('new')
      }
    }
  }, [projectEmpty, session_id, project_id, sessionStatus])

  const {
    dialog,
    setDialog,
    dialogStatus,
    setDialogStatus,
    sessionConfig,
    setSessionConfig,
    resetKey,
    isExpired: checkExpired,
    reset,
    setDialogDirect,
    setDialogStatusDirect,
  } = useSessionState()

  const limitedAnswer = useLimitedAnswer()
  const activeSessionIdRef = useRef(session_id)
  const dialogRef = useRef<DialogList>(dialog)
  const dialogCacheRef = useRef(sessionDialogCache)
  const dialogStatusCacheRef = useRef(sessionDialogStatusCache)
  const activeStreamSessionIdsRef = useRef(activeStreamSessionIds)
  const streamingQuestionIdsRef = useRef(streamingQuestionIds)

  useEffect(() => {
    activeSessionIdRef.current = session_id
  }, [session_id])

  useEffect(() => {
    dialogRef.current = dialog
  }, [dialog])

  useEffect(() => {
    const handleCacheUpdated = (event: Event) => {
      const targetSessionId = (event as CustomEvent<{ id?: number }>).detail
        ?.id
      if (!targetSessionId || activeSessionIdRef.current !== targetSessionId) {
        return
      }
      const cachedDialog = dialogCacheRef.current.get(targetSessionId)
      if (cachedDialog) {
        dialogRef.current = cachedDialog
        setDialogDirect(cachedDialog)
      }
      const cachedStatus = dialogStatusCacheRef.current.get(targetSessionId)
      if (cachedStatus) {
        setDialogStatusDirect(cachedStatus)
      }
    }
    window.addEventListener(SESSION_DIALOG_CACHE_UPDATED, handleCacheUpdated)
    return () => {
      window.removeEventListener(
        SESSION_DIALOG_CACHE_UPDATED,
        handleCacheUpdated,
      )
    }
  }, [setDialogDirect, setDialogStatusDirect])

  const setSessionDialogStatus = (
    targetSessionId: number | undefined,
    status: SessionInfo['dialogStatus'],
  ) => {
    if (!targetSessionId) {
      setDialogStatus(status)
      return
    }
    dialogStatusCacheRef.current.set(targetSessionId, status)
    if (activeSessionIdRef.current === targetSessionId) {
      setDialogStatusDirect(status)
    }
    window.dispatchEvent(
      new CustomEvent(SESSION_DIALOG_CACHE_UPDATED, {
        detail: { id: targetSessionId },
      }),
    )
  }

  const updateSessionDialog = (
    targetSessionId: number | undefined,
    updater: DialogList | ((draft: DialogList) => void),
  ) => {
    if (!targetSessionId) {
      setDialog(updater as any)
      return
    }

    const current =
      dialogCacheRef.current.get(targetSessionId) ??
      (activeSessionIdRef.current === targetSessionId ? dialogRef.current : [])
    const next =
      typeof updater === 'function'
        ? produce(current, updater as (draft: DialogList) => void)
        : updater

    dialogCacheRef.current.set(targetSessionId, next)
    if (activeSessionIdRef.current === targetSessionId) {
      dialogRef.current = next
      setDialogDirect(next)
    }
    window.dispatchEvent(
      new CustomEvent(SESSION_DIALOG_CACHE_UPDATED, {
        detail: { id: targetSessionId },
      }),
    )
  }

  const getCachedDialogUpdater = (targetSessionId: number | undefined) => {
    return (updater: DialogList | ((draft: DialogList) => void)) => {
      updateSessionDialog(targetSessionId, updater)
    }
  }

  const keepAnswerStreaming = (
    targetSessionId: number | undefined,
    index: number,
  ) => {
    updateSessionDialog(targetSessionId, (draft) => {
      const currentDialog = draft[index]
      if (currentDialog?.role !== 'answer') return
      currentDialog.status = currentDialog.content ? 'answering' : 'thinking'
    })
  }

  /** 从后端流式接受回答 并更新至对话中 */
  const accpetAnswer = (
    stream: ReadableStream,
    index: number,
    question_id?: any,
    onFirstChunk?: () => void,
    onComplete?: () => void,
    chartSessionId?: number,
  ) => {
    const targetSessionId = chartSessionId ?? activeSessionIdRef.current
    const setTargetDialog = getCachedDialogUpdater(targetSessionId)
    if (targetSessionId) {
      activeStreamSessionIdsRef.current.add(targetSessionId)
      if (question_id) {
        streamingQuestionIdsRef.current.set(targetSessionId, question_id)
      }
      setSessionDialogStatus(targetSessionId, 'answering')
    }

    // 只有流结束回调才能把消息标记为 answered；中间 chunk 不得提前显示复制操作。
    keepAnswerStreaming(targetSessionId, index)
    let isFirstChunk = true
    getChunkFromStream(
      stream,
      (type, data) => {
        if (type === 'limited') {
          setTargetDialog((draft) => {
            if (typeof index === 'number') {
              draft[index] = limitedAnswer
            } else {
              draft.push(limitedAnswer)
            }
          })
          return
        } else if (type === 'content') {
          // 收到第一个内容块时，触发回调（用于更新路由）
          if (isFirstChunk) {
            isFirstChunk = false
            onFirstChunk?.()
          }
          updateDialog(data, setTargetDialog as any, index, handleOpenReference)
          keepAnswerStreaming(targetSessionId, index)
        } else if (type === 'agent') {
          // 处理 agent 消息
          const {
            messageType,
            taskThought,
            isFinal,
            messageOrder,
            chartConfig,
            chartId, // 获取后端返回的 chartId
          } = data
          setTargetDialog((draft) => {
            const currentDialog = draft[index]
            if (currentDialog?.role !== 'answer') return

            // 初始化 agentStages 数组
            if (!currentDialog.agentStages) {
              currentDialog.agentStages = []
            }

            if (messageType === 'task_thought') {
              // 查找或创建对应的思考过程阶段
              // 同一个思考过程有相同的 messageId，但 messageOrder 可能不同（流式返回）
              const { messageId } = data as any
              let stage = currentDialog.agentStages.find(
                (s) =>
                  s.messageType === 'task_thought' &&
                  (messageId
                    ? (s as any).messageId === messageId
                    : s.messageOrder === messageOrder),
              )

              if (!stage) {
                // 创建新阶段
                const newStage: any = {
                  messageType: 'task_thought',
                  taskThought: '',
                  isFinal: false,
                  messageOrder,
                }
                if (messageId) {
                  newStage.messageId = messageId
                }
                currentDialog.agentStages.push(newStage)
                stage = newStage
              }

              // 流式更新思考内容
              if (taskThought && stage && !isFinal) {
                stage.taskThought += taskThought
              }
              if (stage) {
                stage.isFinal = isFinal
              }
            } else {
              // 调用工具类型（file, code, chart, ''）
              // 使用 messageId 进行去重,避免同一工具的开始调用和调用成功消息重复显示
              const { messageId } = data as any
              const existingStage = currentDialog.agentStages.find(
                (s) =>
                  s.messageType === messageType &&
                  (messageId
                    ? (s as any).messageId === messageId
                    : s.messageOrder === messageOrder),
              )

              if (!existingStage) {
                // 创建新的阶段
                const newStage: any = {
                  messageType,
                  taskThought: taskThought || '',
                  isFinal,
                  messageOrder,
                }
                if (messageId) {
                  newStage.messageId = messageId
                }

                // 如果是图表类型，保存配置
                if (messageType === 'chart' && chartConfig) {
                  newStage.chartConfig = chartConfig
                  // 标记有图表
                  currentDialog.hasCharts = true

                  // 处理图表配置并添加到图表列表
                  try {
                    const chart_content = { ...chartConfig }

                    // 确保有 title 配置
                    if (!chart_content.title) {
                      chart_content.title = { text: '图表', show: true }
                    } else if (chart_content.title?.left) {
                      chart_content.title.show = false
                    }

                    chart_content.color = getColorArray(
                      chart_content?.series?.length || 0,
                    )
                    chart_content?.series?.forEach?.((serie: any) => {
                      Reflect.deleteProperty(serie, 'itemStyle')
                    })

                    // 使用后端返回的 chart_id（如果有），否则使用时间戳 + messageOrder 生成
                    const chart_id = chartId ?? Date.now() + messageOrder

                    // 添加图表到图表列表并打开仪表盘
                    chartsOperators.add(
                      chart_id,
                      chart_content,
                      {},
                      chartSessionId,
                    )
                    handleOpenChart?.()
                  } catch (error) {
                    // 图表配置处理失败
                    console.error('图表配置处理失败:', error)
                  }
                }

                currentDialog.agentStages.push(newStage)
              } else if (existingStage) {
                // 更新现有阶段
                existingStage.isFinal = isFinal
                if (messageType === 'chart' && chartConfig) {
                  existingStage.chartConfig = chartConfig
                  currentDialog.hasCharts = true

                  // 如果是最终状态且还没有添加图表，则添加图表
                  if (isFinal && !(existingStage as any).chartAdded) {
                    try {
                      const chart_content = { ...chartConfig }

                      // 确保有 title 配置
                      if (!chart_content.title) {
                        chart_content.title = { text: '图表', show: true }
                      } else if (chart_content.title?.left) {
                        chart_content.title.show = false
                      }

                      chart_content.color = getColorArray(
                        chart_content?.series?.length || 0,
                      )
                      chart_content?.series?.forEach?.((serie: any) => {
                        Reflect.deleteProperty(serie, 'itemStyle')
                      })

                      // 使用后端返回的 chart_id（如果有），否则使用时间戳 + messageOrder 生成
                      const chart_id = chartId ?? Date.now() + messageOrder

                      // 添加图表到图表列表并打开仪表盘
                      chartsOperators.add(
                        chart_id,
                        chart_content,
                        {},
                        chartSessionId,
                      )
                      // 标记已添加，避免重复添加
                      ;(existingStage as any).chartAdded = true
                      handleOpenChart?.()
                    } catch (error) {
                      // 图表配置处理失败
                      console.error('图表配置处理失败:', error)
                    }
                  }
                }
              }
            }

            // 按 messageOrder 排序
            currentDialog.agentStages.sort(
              (a, b) => a.messageOrder - b.messageOrder,
            )
          })
        } else if (type === 'echarts') {
          // 处理旧的 echarts 格式（flag='echarts' 或普通 content 中包含 ```echarts 代码块）
          const { chart_id, chart_content, extra } = data
          chart_content.color = getColorArray(
            chart_content?.series?.length || 0,
          )
          chart_content?.series?.forEach?.((serie: any) => {
            Reflect.deleteProperty(serie, 'itemStyle')
          })
          if (chart_content?.title?.left) {
            chart_content.title.show = false
          }

          // 标记有图表（用于显示图表按钮）
          setTargetDialog((draft) => {
            const currentDialog = draft[index]
            if (currentDialog?.role === 'answer') {
              currentDialog.hasCharts = true
            }
          })

          chartsOperators.add(chart_id, chart_content, extra, chartSessionId)
          handleOpenChart?.()
        } else if (type === 'history') {
          setTargetDialog((draft) => {
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
        if (targetSessionId) {
          activeStreamSessionIdsRef.current.delete(targetSessionId)
          streamingQuestionIdsRef.current.delete(targetSessionId)
        }
        setStreaming(false)
        setSessionDialogStatus(targetSessionId, 'ready')

        // 获取推荐问题
        let questions: any = null
        let graph: any = null
        if (question_id) {
          const res = await getQuestionInfo(question_id)
          try {
            const {
              question: {
                _source: {
                  sub_question,
                  graph_chat_reference,
                  graph_reference,
                },
              },
            } = res
            questions = sub_question
            graph = {
              graph_chat_reference,
              graph_reference,
            }
            const isGraphMode =
              sessionConfig.mode === 'graph' ||
              sessionConfig.mode === 'graph_search'
            if (isGraphMode && graph_reference) {
              handleOpenGraph?.(index)
            }
          } catch (error) {
            console.log('获取推荐问题失败:', error)
          }
        }

        // 更新状态为 'answered'，显示复制图标和推荐问题
        setTargetDialog((draft) => {
          const currentDialog = draft[index]
          if (currentDialog?.role !== 'answer') return
          currentDialog.status = 'answered'
          if (questions?.length) {
            currentDialog.sub_question = questions
            currentDialog.graph = graph
          }
        })

        // 首次问答完成后，更新会话状态并刷新未分组列表
        if (index === 1) {
          if (
            (project_id === 0 || isOtherPage || type === 'single-file') &&
            isNewSessionCreatedRef.current
          ) {
            // 先触发未分组会话列表刷新事件
            window.dispatchEvent(new CustomEvent('refreshUngroupedSessions'))
            // 设置标志，触发后续的状态更新
            setShouldRefreshUngrouped(true)
            isNewSessionCreatedRef.current = false
          } else {
            // 非未分组会话，直接更新状态（React 18 自动批处理会处理状态更新）
            setSessionStatus('created')
          }
        }

        // 提问完成后，刷新历史记录列表（确保会话顺序正确）
        // 只在会话分组中刷新，未分组会话不需要刷新
        if (project_id != null && project_id !== 0 && !isOtherPage) {
          setShouldReloadSessions(true)
        }

        onComplete?.()
      },
    )
  }

  /** 用于停止回答 */
  const lastQuestionId = useRef<number>()
  /** 标记是否正在流式传输中，用于防止路由跳转时重置 dialog */
  const isStreamingRef = useRef(false)
  /** 标记是否是用户主动创建的新会话（用于判断是否需要刷新未分组列表） */
  const isNewSessionCreatedRef = useRef(false)
  /** 用于触发未分组会话列表刷新的标志 */
  const [shouldRefreshUngrouped, setShouldRefreshUngrouped] = useState(false)
  /** 用于触发历史记录列表刷新的标志 */
  const [shouldReloadSessions, setShouldReloadSessions] = useState(false)

  /** 流式传输状态管理 */
  const setStreaming = (value: boolean) => {
    isStreamingRef.current = value
  }

  // 处理未分组会话列表刷新后的状态更新
  useEffect(() => {
    if (!shouldRefreshUngrouped) return

    // 使用 queueMicrotask 确保在下一个微任务中执行，给列表刷新一些时间
    queueMicrotask(() => {
      // 再次使用 queueMicrotask 确保列表刷新完成后再更新状态
      queueMicrotask(() => {
        setSessionStatus('created')
        setShouldRefreshUngrouped(false)
      })
    })
  }, [shouldRefreshUngrouped])

  // 处理历史记录列表刷新
  useEffect(() => {
    if (!shouldReloadSessions) return

    // 使用 queueMicrotask 确保在下一个微任务中执行
    queueMicrotask(() => {
      reloadSessions()
      setShouldReloadSessions(false)
    })
  }, [shouldReloadSessions, reloadSessions])

  // 记录上一次的 session_id，用于判断是否切换了会话
  const prevSessionIdRef = useRef(session_id)

  /** 停止回答的函数 */
  const stopQA = useMemoizedFn(() => {
    if (!lastQuestionId.current) return
    const questionIdToStop = lastQuestionId.current
    lastQuestionId.current = undefined
    setStreaming(false)
    // 调用停止接口，确保后端保存已回复的内容
    stopChat({ question_id: questionIdToStop }).catch((error) => {
      console.log('停止回答失败:', error)
    })
  })

  useAsyncEffect(async () => {
    // 保存当前状态，用于过期检查
    const currentSessionId = session_id
    const currentResetKey = resetKey

    // 使用从 useSessionState 返回的 isExpired 函数
    // 该函数直接访问 resetKeyRef.current，可以正确检测到 resetKey 的变化

    const isSessionChanged = prevSessionIdRef.current !== session_id
    prevSessionIdRef.current = session_id

    if (!currentSessionId) {
      setDialogStatus('ready')
      return
    }

    // 切换会话时不要停止旧会话的回答。旧流的回调会因 resetKey 失效而
    // 停止更新当前页面，但后端任务仍需继续执行，用户稍后回到该会话时再续接。

    isNewSessionCreatedRef.current = false

    try {
      // 切换会话时延迟，确保后端数据落库
      if (isSessionChanged) {
        await new Promise((resolve) => setTimeout(resolve, 300))
        if (checkExpired(currentResetKey)) return
      }

      // 加载会话信息和历史记录
      const sessionInfo = await getSessionInfoWithType(currentSessionId)
      setSessionConfig(sessionInfo)
      const _history: any[] = await getSessionChats({
        id: currentSessionId,
      }).then((res) => res.Data ?? [])

      if (checkExpired(currentResetKey)) return

      // 更新对话记录
      lastQuestionId.current = _history.at(-1)?._id
      const dialogFromHistory = getDialogFromHistory(_history)
      const cachedDialog = dialogCacheRef.current.get(currentSessionId)
      const displayDialog =
        cachedDialog && cachedDialog.length >= dialogFromHistory.length
          ? cachedDialog
          : dialogFromHistory
      dialogCacheRef.current.set(currentSessionId, displayDialog)
      dialogRef.current = displayDialog
      setDialogDirect(displayDialog)

      // 从最后一条历史记录中读取 enable_web_search 状态
      const lastHistoryItem = _history.at(-1)?._source
      const historyEnableWebSearch =
        lastHistoryItem?.extra?.agent?.enable_web_search ?? false
      setSessionConfig({ enable_web_search: historyEnableWebSearch })

      // 如果没有历史记录或没有未完成的回答，设置为 ready
      if (activeStreamSessionIdsRef.current.has(currentSessionId)) {
        lastQuestionId.current =
          streamingQuestionIdsRef.current.get(currentSessionId) ??
          lastQuestionId.current
        keepAnswerStreaming(currentSessionId, displayDialog.length - 1)
        setSessionDialogStatus(currentSessionId, 'answering')
        return
      }

      // 如果没有历史记录或没有未完成的回答，设置为 ready
      if (displayDialog.length === 0 || !lastQuestionId.current) {
        setDialogStatus('ready')
        return
      }

      // 继续未完成的回答
      try {
        const { body } = (await continueLastChat({
          question_id: lastQuestionId.current,
        })) as any

        if (checkExpired(currentResetKey)) return

        setDialogStatus('answering')
        accpetAnswer(
          body,
          displayDialog.length - 1,
          undefined,
          undefined,
          undefined,
          currentSessionId,
        )
      } catch (continueError) {
        // continueLastChat 失败时，直接使用历史记录显示
        console.log('继续回答失败，使用历史记录:', continueError)
        setDialogStatus('ready')
      }
    } catch (error) {
      console.log('加载会话失败:', error)
      if (!checkExpired(currentResetKey)) {
        setDialogStatus('ready')
      }
    }
  }, [resetKey])

  const startQA = useMemoizedFn(async (qa_data: QAData) => {
    setDialogStatus('asking')

    const { content, images, input } = qa_data
    const questionDialog: DialogList[number] = {
      role: 'question',
      content,
      images,
      attachments: input?.attachments,
      created_at: dayjs().toString(),
    }
    const answerDialog: DialogList[number] = {
      role: 'answer',
      content: '',
      thinkingContent: '',
      status: 'thinking',
      reference: [],
    }
    const nextDialog = [...dialogRef.current, questionDialog, answerDialog]
    const currentAnswerIndex = nextDialog.length - 1
    dialogRef.current = nextDialog
    setDialog(nextDialog)
    // 获取或创建会话 ID
    let currentSessionId = session_id
    if (!currentSessionId) {
      setSessionStatus('creating')
      if (!sessionConfig.model_id) {
        sessionConfig.model_id = 1
      }

      if (sessionConfig.mode === 'h3c-test') {
        currentSessionId = await createH3CTest()
      } else {
        // 单文档详情页场景下，project_id 应该固定为 0
        const finalProjectId = isOtherPage ? 0 : project_id!
        currentSessionId = await createSessionByKnowledge(
          finalProjectId,
          sessionConfig as SessionConfig,
        )
      }
      // 标记这是用户主动创建的新会话
      isNewSessionCreatedRef.current = true
      if (project_id === 0 || type === 'global') {
        window.dispatchEvent(
          new CustomEvent('pendingUngroupedSession', {
            detail: { id: currentSessionId },
          }),
        )
      }
      setData((draft) => {
        const exists = draft.sessions.some(
          (item) => item.session_id === currentSessionId,
        )
        if (!exists) {
          draft.sessions.unshift({
            session_id: currentSessionId!,
            name: '',
            nameLoading: true,
          })
        }
      })
    }
    if (currentSessionId) {
      dialogCacheRef.current.set(currentSessionId, nextDialog)
      dialogStatusCacheRef.current.set(currentSessionId, 'asking')
      // 会话创建完成到 SSE 连接建立之间也属于进行中，切换页面时不能按历史回答恢复。
      activeStreamSessionIdsRef.current.add(currentSessionId)
    }
    // 更新路由
    // 单文件问答页面和新华三测试不进行路由跳转
    if (!session_id && currentSessionId) {
      setSessionId(
        currentSessionId,
        // 不跳转路由
        sessionConfig.mode === 'h3c-test' || type === 'single-file',
      )
      // 手动更新 ref，避免新建会话时状态不一致
      prevSessionIdRef.current = currentSessionId
    }
    // 创建流式请求
    const requestBody: any = {
      session_id: currentSessionId,
      question: content,
      ...(qa_data.options ? { options: qa_data.options } : {}),
      ...(qa_data.input ? { input: qa_data.input } : {}),
    }

    try {
      const { question_id } = (await createStream(requestBody)) as any
      lastQuestionId.current = question_id

      const { body } = (await sendStream({
        session_id: currentSessionId,
        question_id,
      })) as any

      // 标记开始流式传输，确保在路由更新前设置（避免路由跳转中断流式传输）
      setStreaming(true)
      setDialogStatus('answering')

      // 开始处理流式响应
      accpetAnswer(
        body,
        currentAnswerIndex!,
        question_id,
        undefined,
        undefined,
        currentSessionId,
      )
    } catch (error) {
      if (currentSessionId) {
        activeStreamSessionIdsRef.current.delete(currentSessionId)
        streamingQuestionIdsRef.current.delete(currentSessionId)
        setSessionDialogStatus(currentSessionId, 'ready')
      }
      throw error
    }
  })

  return {
    sessionStatus,
    setSessionStatus,
    dialog,
    dialogStatus,
    sessionConfig,
    setSessionConfig,
    startQA,
    stopQA,
    resetSession: reset,
  }
}

/** 可重置的session状态 */
const useSessionState = () => {
  const { type } = useProject()
  const { isH3CTest } = useDeployConfig()
  const [dialog, setDialog] = useImmer<DialogList>([])
  /** 问答的状态 */
  const [dialogStatus, setDialogStatus] =
    useState<SessionInfo['dialogStatus']>('loading')
  const [sessionConfig, _setSessionConfig] = useState<Partial<SessionConfig>>(
    () => {
      if (type === 'global' && isH3CTest) return { mode: 'h3c-test' }
      if (type === 'single-file') return { mode: 'knowledge' }
      return { mode: 'model' }
    },
  )
  /** 函数或赋值方式 增量更新 */
  const setSessionConfig: SessionInfo['setSessionConfig'] = (arg) => {
    _setSessionConfig((val) => {
      const newVal = typeof arg === 'function' ? produce(val, arg) : arg
      return { ...val, ...newVal }
    })
  }

  /** 如果已经reset过了 则不调用函数 */
  const resetKeyRef = useRef(0)
  const currentKey = resetKeyRef.current
  // 在异步任务中调用setState时 由于闭包的原因 currentKey始终是创建时的key 可以判断是否过期
  // 不能配合useMemoizedFn
  function withReset<Args extends any[], R>(fn: (...args: Args) => R) {
    const fnWithResetKey = (...args: Args) => {
      // 如果 key 过期，静默返回，不执行操作也不抛出错误
      // 这样可以避免在异步操作完成时 key 已过期导致的错误
      if (currentKey !== resetKeyRef.current) {
        return
      }
      return fn(...args)
    }
    return fnWithResetKey
  }
  const reset = () => {
    resetKeyRef.current = performance.now()
    setDialog([])
    setDialogStatus('loading')
    _setSessionConfig({})
  }
  // 检查是否过期的函数：直接访问 resetKeyRef.current
  const isExpired = (capturedKey: number) => {
    return capturedKey !== resetKeyRef.current
  }
  return {
    dialog,
    setDialog: withReset(setDialog),
    setDialogDirect: setDialog,
    dialogStatus,
    setDialogStatus: withReset(setDialogStatus),
    setDialogStatusDirect: setDialogStatus,
    sessionConfig,
    setSessionConfig: withReset(setSessionConfig),
    resetKey: resetKeyRef.current,
    isExpired,
    reset,
  }
}

/** 创建一个新的会话 */
const createSessionByKnowledge = async (
  project_id: number,
  config: SessionConfig,
) => {
  const {
    mode,
    model_id,
    knowledge: _knowledge,
    graphKnowledgeBase,
    tableKnowledgeBase,
    databaseKnowledgeBase,
    externalIds,
    externalDataProviders,
    prompt_key,
    tag_ids,
  } = config

  const body: any = {
    project_id:
      project_id !== undefined && project_id !== null ? project_id : undefined,
    model_id,
    resource_id: model_id,
  }

  // 文档模式（knowledge）时处理参数
  if (mode === 'knowledge') {
    if (prompt_key) {
      body.prompt_key = prompt_key
    }
    const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
    const hasKnowledge = _knowledge?.length || externalIds?.length

    // 如果只选择了标签，没有选择知识库资源
    if (actualTagIds.length > 0 && !hasKnowledge) {
      // 先获取标签对应的文件列表
      const fileListRes = await getFileList({
        forest_id: 0, // 全局搜索
        filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
        limit: 9999, // 获取所有文件
      })
      const fileIds = (fileListRes.data || []).map((file: any) => file.ID)

      body.ids = fileIds
      body.resource_type = 'file_list'
      body.base_type = 'standard'
      body.tag_ids = actualTagIds
      body.tag_resourse_type = 'file'
      const { ID: id } = await createSession(body)
      return id
    }
  }

  switch (mode) {
    case 'model': {
      body.base_type = 'model'
      body.resource_type = 'model'
      break
    }
    case 'graph_search': {
      body.base_type = 'graph_search'
      body.resource_type = 'graph_search'
      break
    }
    case 'external_data': {
      // 外接数据模式
      body.base_type = 'external_data'
      body.resource_type = 'external_data'
      body.providers = externalDataProviders
      break
    }
    case 'graph':
    case 'table':
    case 'database':
    case 'knowledge': {
      // 图谱问答的数据是知识库问答的子集 可以使用相同的逻辑处理
      // 知识库问答时 knowledge(corekg知识库及其内部的项) 和 externalIds(外部数据源) 必定二选一
      const knowledge =
        mode === 'graph'
          ? graphKnowledgeBase
          : mode === 'table'
            ? tableKnowledgeBase
            : mode === 'database'
              ? databaseKnowledgeBase
              : _knowledge
      if (knowledge) {
        const example = knowledge[0]
        body.resource_id = example.forest_id
        // 若选择知识库则必定是原子节点
        switch (example.knowledgeType) {
          case 'qa': {
            // 选择整个知识库的是问答对
            body.base_type = 'standard'
            body.resource_type = 'forest'
            body.ids = knowledge.map((item) => item.forest_id)
            break
          }
          case 'file': {
            // 多模态知识库
            body.base_type = 'standard'
            body.resource_type = 'file_list'
            body.ids = knowledge.map((item) => item.id)
            break
          }
          //excel mysql 所有节点属于同一知识库
          case 'excel_sheet': {
            // excel
            body.base_type = 'react_excel'
            body.resource_type = 'react_excel_list'
            body.ids = knowledge.map((item) => item.id)
            break
          }
          case 'mysql_table': {
            // mysql
            body.base_type = 'mysql'
            body.resource_type = 'db_table_list'
            body.names = knowledge.map((item) => item.name)
            break
          }
          default:
            throw new Error('未知的知识类型')
        }
      } else {
        // 外部数据源
        body.ids = externalIds
        body.resource_type = 'external_data'
      }
      if (mode === 'graph') {
        body.base_type = 'graph_qa'
      }
      // 如果选择了标签，添加标签相关参数
      if (mode === 'knowledge' && tag_ids?.length) {
        const actualTagIds = tag_ids.filter((id) => id >= 0)
        if (actualTagIds.length > 0) {
          body.tag_ids = actualTagIds
          body.tag_resourse_type = 'file'
        }
      }
      break
    }
  }

  const { ID: id } = await createSession(body)
  return id
}

const getSessionInfoWithType = async (
  session_id: number,
): Promise<SessionConfig> => {
  type KnowledgeNames<T extends string> = {
    [key in T]: { id: number; name: string }[]
  }
  const {
    model_id,
    resource_type,
    external_token_id_list,
    mode: base_type,
    prompt_mode,
    ...knowledgeNames
  } = (await getSessionInfo({
    id: session_id,
  })) as any as {
    model_id: number
    model_name: string
    resource_type: string
    mode: string
    prompt_mode?: string
  } & KnowledgeNames<
    | 'forest_names'
    | 'file_names'
    | 'excel_names'
    | 'excel_sheet_names'
    | 'db_names'
    | 'db_table_names'
  > & {
      external_token_id_list: number[]
    }
  // 将后端的 prompt_mode 映射到前端的 prompt_key
  const prompt_key = prompt_mode
  if (external_token_id_list.length !== 0) {
    return {
      mode: 'knowledge',
      model_id,
      externalIds: external_token_id_list,
      ...(prompt_key ? { prompt_key } : {}),
    }
  }
  if (resource_type === 'model') {
    return {
      mode: 'model',
      model_id,
      ...(prompt_key ? { prompt_key } : {}),
    }
  }

  const createKnowledge = (
    knowledgeType: Knowledge['knowledgeType'],
    items: { id: number; name: string }[],
  ): Knowledge[] => {
    return items && items.length > 0
      ? items.map((item) => {
          return {
            ...item,
            forest_id: 0,
            key: item.name,
            parentKey: '',
            type: knowledgeType,
            knowledgeType,
            node_type: '',
          }
        })
      : []
  }

  const knowledge: Knowledge[] = (() => {
    switch (resource_type) {
      case 'file_list': // 多模态
      case 'file': // 多模态
        if (!knowledgeNames.file_names) {
          // 如果后端返回的是 file_id_list（数字数组），需要转换为 name 和 id 的对象数组
          const fileIdList = (knowledgeNames as any).file_id_list as
            | number[]
            | undefined
          if (fileIdList) {
            return createKnowledge(
              'file',
              fileIdList.map((item: number) => ({
                name: String(item),
                id: item,
              })),
            )
          }
          return []
        }
        return createKnowledge('file', knowledgeNames.file_names)

      case 'forest': // qa
        return createKnowledge('qa', knowledgeNames.forest_names)
      case 'react_excel_list': // excel_sheet
        return createKnowledge('excel_sheet', knowledgeNames.excel_names)
      case 'db_table_list': // table
        return createKnowledge('mysql_table', knowledgeNames.db_table_names)
      default:
        return []
    }
  })()
  if (base_type === 'graph_qa') {
    return {
      mode: 'graph',
      model_id,
      graphKnowledgeBase: knowledge,
      ...(prompt_key ? { prompt_key } : {}),
    }
  } else if (base_type === 'graph_search') {
    return {
      mode: 'graph_search',
      model_id,
      ...(prompt_key ? { prompt_key } : {}),
    }
  } else if (base_type === 'external_data') {
    // 外接数据模式
    return {
      mode: 'external_data',
      model_id,
      externalDataProviders: (knowledgeNames as any).providers || [],
      ...(prompt_key ? { prompt_key } : {}),
    }
  } else if (resource_type === 'react_excel_list') {
    return {
      mode: 'table',
      model_id,
      tableKnowledgeBase: knowledge,
      ...(prompt_key ? { prompt_key } : {}),
    }
  } else if (resource_type === 'db_table_list' || base_type === 'mysql') {
    return {
      mode: 'database',
      model_id,
      databaseKnowledgeBase: knowledge,
      ...(prompt_key ? { prompt_key } : {}),
    }
  }
  return {
    mode: 'knowledge',
    model_id,
    knowledge,
    ...(prompt_key ? { prompt_key } : {}),
  }
}
const createH3CTest = async () => {
  const res = await createSession({
    model_id: 1,
    resource_id: 663,
    name: 'new chat',
    resource_type: 'agent',
    source_from: 'agent',
  })
  return res.ID
}
