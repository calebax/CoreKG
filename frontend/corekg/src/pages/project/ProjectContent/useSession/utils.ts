import { DialogList } from '@/components/dialog'
import { QAHistory, StreamChunk } from '@/components/dialog/utils'

type OnLoadChunkArgs =
  | ['content', StreamChunk]
  | [
      'echarts',
      {
        chart_id: number
        chart_content: any
        extra?: any
      },
    ]
  | [
      'agent',
      {
        messageType:
          | 'task_thought'
          | 'file'
          | 'code'
          | 'chart'
          | 'question_rewrite'
          | ''
        taskThought?: string
        isFinal: boolean
        messageOrder: number
        messageId?: string
        chartConfig?: any
        chartId?: number // 添加 chartId 字段
      },
    ]
  | ['history', undefined]
  | ['limited', undefined]

type OnLoadChunk = (...args: OnLoadChunkArgs) => void

/**
 * 从 echarts 代码块中提取图表配置
 * @param content - 包含 ```echarts 代码块的内容
 * @returns 解析后的图表配置，如果解析失败返回 null
 */
const extractEchartsConfig = (content: string) => {
  const echartsMatch = content.match(/```echarts([\s\S]*?)```/)
  if (!echartsMatch) return null

  try {
    const {
      chart_id,
      chart_content: _content,
      ...extra
    } = JSON.parse(echartsMatch[1])
    // chart_content 可能是字符串（需要解析）或已经是对象
    const chart_content =
      typeof _content === 'string' ? JSON.parse(_content) : _content
    return { chart_id, chart_content, extra }
  } catch {
    return null
  }
}

/**
 * 从流式响应中解析数据块
 * 支持多种消息类型：content（普通内容）、echarts（图表）、agent（智能体阶段）、history（历史记录）、limited（限流）
 *
 * @param stream - 可读流对象
 * @param onLoadChunk - 每解析到一个数据块时的回调函数
 * @param onEnd - 流读取结束时的回调函数
 */
export const getChunkFromStream = async (
  stream: ReadableStream,
  onLoadChunk: OnLoadChunk,
  onEnd?: () => void,
) => {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  // 用于处理消息过长被分割的情况，累积错误消息后重新解析
  let errorMsg: string[] = []

  /**
   * 处理单条消息的回调函数
   * 根据消息的 flag 字段判断消息类型，并调用相应的处理逻辑
   */
  const cb = (msg: string) => {
    // 处理历史记录标记：当消息以 'history' 开头时，表示后续内容是历史记录
    if (msg.startsWith('history')) {
      // 使用单独的'history' 标记下方是历史记录
      // 会将历史记录和新的chunk以普通chunk的方式传输
      onLoadChunk('history', undefined)
      return
    }
    const value = JSON.parse(msg)

    // 处理限流错误：code=10013 表示请求被限流
    if (value.code === 10013) {
      onLoadChunk('limited', undefined)
      return
    }

    /**
     * 处理 Agent 智能体阶段消息（flag='agent' 或 flag='customize'）
     * Agent 消息包含思考过程、工具调用、图表生成、问题改写等不同阶段的信息
     */
    if (
      (value.flag === 'agent' || value.flag === 'customize') &&
      value.content
    ) {
      try {
        // content 是对象格式
        const agentContent = value.content
        const messageType = agentContent.message_type || ''
        const task_thought = agentContent.task_thought || ''
        const is_final = agentContent.is_final || false
        const message_order = agentContent.message_order || 0
        const task_order = agentContent.task_order || 0
        const message_id = agentContent.message_id
        const result_map = agentContent.result_map

        /**
         * 处理图表配置：当 messageType='chart' 时，图表配置存储在 result_map.chart_content 中
         * chart_content 的内容就是ECharts的配置
         */
        let chartConfig: any = undefined
        let chartId: number | undefined = undefined
        if (messageType === 'chart' && result_map?.chart_content) {
          try {
            chartConfig = result_map.chart_content
            chartId = result_map.chart_id // 获取后端返回的 chart_id
          } catch (error) {
            console.error('图表配置解析失败:', error)
          }
        }

        /**
         * 排序逻辑：使用 task_order 作为主要排序依据（任务级别的顺序），
         * message_order 作为次要排序依据（消息级别的顺序）
         * 这样可以保证多个任务的消息按照正确的顺序显示
         */
        const order = task_order || message_order || 0

        onLoadChunk('agent', {
          messageType: messageType || '',
          taskThought: task_thought || '',
          isFinal: is_final || false,
          messageOrder: order,
          messageId: message_id,
          chartConfig,
          chartId, // 传递 chartId
        })
        return
      } catch {
        // 如果解析失败，作为普通 content 处理
      }
    }

    /**
     * 处理 echarts 格式（从 content 中的 ```echarts 代码块解析）
     * 支持两种情况：
     * 1. flag === 'echarts'：明确的 echarts 消息，提取后直接返回
     * 2. flag === '' 或 undefined：普通 content 消息中包含 ```echarts 代码块，提取后继续处理 content
     */
    if (value.content) {
      const echartsConfig = extractEchartsConfig(value.content)
      if (echartsConfig) {
        onLoadChunk('echarts', echartsConfig)
        // flag === 'echarts' 时直接返回，否则继续处理 content
        if (value.flag === 'echarts') {
          return
        }
        // flag === '' 或 undefined 时，继续处理 content（echarts 代码块会在 MarkdownPreview 中渲染）
      }
    }

    // 处理普通内容消息
    onLoadChunk('content', value)
  }

  /**
   * 流式读取主循环
   * 持续从流中读取数据，按行分割后逐条处理
   * 如果单条消息解析失败（可能是消息过长被分割），会累积到 errorMsg 中重新解析
   */
  while (true) {
    const { done, value } = await reader.read()

    // 将二进制数据解码为文本，stream: true 表示可能还有后续数据
    const text = decoder.decode(value, { stream: true })
    // 按换行符分割消息，过滤空行
    const messages = text.split('\n').filter(Boolean)
    messages.forEach((msg) => {
      try {
        cb(msg)
      } catch {
        /**
         * 错误处理：如果消息解析失败，可能是消息过长被分割成多段
         * 将失败的消息累积到 errorMsg 中，尝试合并后重新解析
         */
        errorMsg.push(msg)
        try {
          const mergeMsg = errorMsg.join('')
          cb(mergeMsg)
          errorMsg = []
        } catch {
          // 如果合并后仍然解析失败，放弃这条消息
          return
        }
        return
      }
    })
    if (done) {
      // 流读取完成，调用结束回调
      onEnd?.()
      return
    }
  }
}

/**
 * 从历史记录中构建对话列表
 * 将历史记录格式转换为前端使用的 DialogList 格式
 *
 * @param history - 历史记录数组，可能包含 _source 字段（ES 搜索结果格式）或直接是数据对象
 * @returns 转换后的对话列表，每个历史记录对应一个问题和答案对
 */
export const getDialogFromHistory = (history: QAHistory) => {
  const historyDialog: DialogList = []
  history.forEach((item) => {
    /**
     * 兼容新旧数据结构：
     * - 新格式：数据在 _source 字段中（ES 搜索结果格式）
     * - 旧格式：数据直接在 item 中
     */
    const data = item._source || item

    // 处理 msg 字段（深度思考内容）
    // 注意：msg 字段在 item 上，不在 _source 中
    const msg = (item as any).msg || []

    // 过滤并转换 msg 为 agentStages
    const agentStages = msg
      .filter(
        (msgItem: any) =>
          (msgItem.flag === 'agent' || msgItem.flag === 'customize') &&
          msgItem.content &&
          typeof msgItem.content === 'object' &&
          msgItem.content.message_type,
      )
      .map((msgItem: any) => {
        const { content } = msgItem
        const {
          message_type,
          task_thought,
          is_final,
          message_id,
          message_time,
          result_map,
          message_order,
          task_order,
        } = content

        const stage: any = {
          messageType: message_type,
          taskThought: task_thought || '',
          isFinal: is_final !== undefined ? is_final : true,
          messageTime: message_time || 0,
          messageOrder: task_order || message_order || 0,
        }

        if (message_id) stage.messageId = message_id

        // 处理图表配置
        if (message_type === 'chart' && result_map?.chart_content) {
          try {
            const parsed = JSON.parse(result_map.chart_content)
            stage.chartConfig = parsed?.structuredContent
            if (result_map.chart_id) stage.chartId = result_map.chart_id
          } catch {
            // 图表配置解析失败，静默忽略
          }
        }

        return stage
      })
      // 去重：使用 messageId 和 messageType 组合进行去重,避免同一工具的开始调用和调用成功消息重复显示
      .reduce((unique: any[], stage: any) => {
        const existingIndex = unique.findIndex(
          (s) =>
            s.messageType === stage.messageType &&
            s.messageId &&
            s.messageId === stage.messageId,
        )
        if (existingIndex === -1) {
          // 不存在,添加到结果
          unique.push(stage)
        } else {
          // 已存在,保留 is_final 为 true 的版本(调用成功的消息)
          if (stage.isFinal) {
            unique[existingIndex] = stage
          }
        }
        return unique
      }, [])
      .sort((a: any, b: any) => {
        if (a.messageOrder !== b.messageOrder) {
          return a.messageOrder - b.messageOrder
        }
        return (a.messageTime || 0) - (b.messageTime || 0)
      })

    // 检查是否有图表
    const hasCharts = agentStages.some(
      (stage: any) => stage.messageType === 'chart',
    )

    // 每个历史记录转换为一个问题和答案对
    historyDialog.push(
      {
        created_at: (data as any).created_at,
        role: 'question',
        content: data.question ?? '',
        images: data.image_url_list || [],
        attachments: data.extra?.input?.attachments || [],
      },
      {
        role: 'answer',
        thinkingContent: data.reasoning ?? '',
        // 移除答案中的 echarts 代码块（图表已在仪表盘中显示，不需要在文本中重复）
        content: (data.answer ?? '').replace(/```echarts[\s\S]*?```/g, ''),
        status: 'answered',
        /**
         * 兼容新旧数据格式：引用文件列表
         * - 新格式使用 file_name 字段
         * - 旧格式使用 filename 字段
         */
        reference: (data.query_reference_list ?? []).map((ref: any) => ({
          ...ref,
          file_name: ref.file_name || ref.filename || '',
        })),
        sub_question: data?.sub_question ?? [],
        graph: {
          graph_chat_reference: data.graph_chat_reference,
          graph_reference: data.graph_reference,
        },
        // 如果有 agentStages，设置 agentStages 和 hasCharts
        ...(agentStages.length > 0
          ? {
              agentStages,
              hasCharts,
            }
          : {}),
      },
    )
  })
  return historyDialog
}
