import { FC, useMemo, useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ConfigProvider, message, Spin, Tooltip } from 'antd'
import { RightOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { match, P } from 'ts-pattern'
import { cn, scrollToEnd } from '@/utils'
import DocIcon from '@/assets/icons/docs/doc.svg?react'
import DownIcon from '@/assets/icons/down.svg?react'
import UpIcon from '@/assets/icons/up.svg?react'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import { getFileIcon } from '@/components/common/ReferencePreviewModal/utils'
import type { AIDialog } from '@/components/dialog'
import type { AgentStage } from '@/components/dialog/AIDialog'
import { useProject } from '@/pages/project'
import { copyToClipboard } from '@/utils/copy'
import { markdownToText } from '@/utils/markdownToText'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useSessionInfo } from '../../..'
import { EDrawerType } from '../../../hooks'
import { useAiDialog } from './hooks'
import icon from './icon.svg'
import CopyIcon from './images/copy.svg?react'
import DeepSeek from './images/deepseek.svg'
import DropLeftIcon from './images/drop-left.svg?react'
import DropIcon from './images/fold-up.svg?react'
import ThinkingIcon from './images/thinking.svg?react'
import ToolIcon from './images/tools.svg?react'
import styles from './index.module.scss'

export const ProjectAIDialog: FC<Style & { value: AIDialog; index: number }> = (
  props,
) => {
  const { index, value, className, style } = props
  const { status, thinkingContent, reference, graph } = value
  const content = useMemo(() => {
    const originContent = value.content
    return originContent?.replace(
      /\{Reference\s+§(?:file_)?(\d+)\[[^\]]+\]\}/g,
      (match, number) => {
        const num = parseInt(number)
        const hasReference = reference?.some((item) => item.file_id === num)
        return hasReference ? match : ''
      },
    )
  }, [reference, value.content])
  const referencedChunks = useMemo(() => {
    if (!content || !reference?.length) return []

    const annotationMap = new Map<
      string,
      {
        annotationIndex: number
        file_id: number
        forest_id: number
        file_name: string
        sequence: number
      }
    >()
    const referencePattern = /§(?:file_)?(\d+)\[([^\]]+)\]/g
    let match: RegExpExecArray | null
    let nextAnnotationIndex = 1

    while ((match = referencePattern.exec(content)) !== null) {
      const fileId = Number(match[1])
      const file = reference.find((item) => item.file_id === fileId)
      if (!file) continue

      const chunks = match[2]
        .split(',')
        .map((chunk) => chunk.trim())
        .filter((chunk) => chunk.length > 0 && !chunk.startsWith('chunkid:'))

      chunks.forEach((chunk) => {
        const sequence = Number(chunk)
        if (Number.isNaN(sequence)) return

        const uniqueKey = `${fileId}-${sequence}`
        if (annotationMap.has(uniqueKey)) return

        annotationMap.set(uniqueKey, {
          annotationIndex: nextAnnotationIndex++,
          file_id: file.file_id,
          forest_id: file.forest_id,
          file_name: file.file_name,
          sequence,
        })
      })
    }

    return Array.from(annotationMap.values()).sort(
      (a, b) => a.annotationIndex - b.annotationIndex,
    )
  }, [content, reference])
  const referencedFiles = useMemo(() => {
    if (!referencedChunks.length) return []

    const fileMap = new Map<
      number,
      {
        file_id: number
        forest_id: number
        file_name: string
        annotations: number[]
      }
    >()

    referencedChunks.forEach((chunk) => {
      const existingFile = fileMap.get(chunk.file_id)
      if (existingFile) {
        existingFile.annotations.push(chunk.annotationIndex)
        return
      }

      fileMap.set(chunk.file_id, {
        file_id: chunk.file_id,
        forest_id: chunk.forest_id,
        file_name: chunk.file_name,
        annotations: [chunk.annotationIndex],
      })
    })

    return Array.from(fileMap.values())
  }, [referencedChunks])
  const deployConfig = useDeployConfig() as any
  const { version } = deployConfig
  const { t } = useTranslation('common')
  const { t: tC } = useTranslation('pages')
  const { models, isOtherPage, session_id } = useProject()
  const {
    sessionConfig,
    handleOpenReference,
    handleOpenGraph,
    handleOpenChart,
    handleCloseDrawer,
    drawerVisible,
    drawerType,
    dialogIndex,
  } = useSessionInfo()

  const {
    thinkingContentValue,
    referenceText,
    showThinking,
    thinkingVisible,
    setThinkingVisible,
  } = useAiDialog(props)

  // 引用资料展开/收起状态
  const [referenceExpanded, setReferenceExpanded] = useState(true)

  /**
   * 旧的思考过程容器的自动滚动控制
   * 用于兼容旧的 thinkingContent 字段（单阶段思考过程）
   */
  const thinkingContainerRef = useRef<HTMLDivElement>(null)
  const shouldScrollThinking = useRef(true)

  /**
   * Agent 阶段的展开状态管理
   * key: messageId（消息唯一ID，用于唯一标识每个阶段）
   * value: boolean（是否展开，undefined 表示未初始化，默认展开）
   */
  const [agentStageExpanded, setAgentStageExpanded] = useState<
    Record<string, boolean>
  >({})

  /**
   * Agent 阶段的滚动控制
   * 每个阶段都有独立的滚动容器和滚动控制标志
   * - agentStageScrollRefs: 存储每个阶段的 DOM 引用（使用 messageId 作为 key）
   * - agentStageShouldScroll: 控制每个阶段是否应该自动滚动（用户手动滚动后设为 false）
   */
  const agentStageScrollRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const agentStageShouldScroll = useRef<Record<string, boolean>>({})

  const currentModel = useMemo(() => {
    return models.find((item) => item.id === sessionConfig?.model_id)
  }, [models, sessionConfig?.model_id])
  // 当 AI 尚未开始输出内容且对话未结束时，展示一个简洁的旋转 loading
  // 思考过程中（有 thinkingContent 或 agentStages）不应该显示 loading
  // 一旦开始输出内容（有 content），就应该隐藏 loading
  const hasThinkingContent =
    !!thinkingContent || (!!value.agentStages && value.agentStages.length > 0)
  const showLoading = !hasThinkingContent && !content && status !== 'answered'

  // 监听思考内容变化，自动滚动到底部（复用问答部分的滚动逻辑）
  useEffect(() => {
    const dom = thinkingContainerRef.current
    if (!dom || !showThinking || !shouldScrollThinking.current) return
    scrollToEnd(dom)
  }, [thinkingContentValue, showThinking])

  const renderThinkingContent = () => {
    // 判断思考是否完成（有content表示思考完成）
    const isThinkingComplete = !!content

    // 收起状态下的文本：思考完成显示"已完成思考"，未完成显示"思考和行动过程"
    const collapsedText = isThinkingComplete ? '已完成思考' : '思考和行动过程'

    return (
      <div
        className={cn(
          'border border-[#EFF1F4] rounded-[6px]',
          // 展开时：使用 w-fit 让宽度根据内容自适应，最大宽度由父容器限制
          // 收起时：固定宽度 171px
          showThinking ? 'w-fit' : 'w-[171px]',
          styles.thinkingContainer,
        )}
      >
        <div
          onClick={() => setThinkingVisible(!showThinking)}
          className={cn(
            'flex items-center justify-between cursor-pointer',
            showThinking ? 'px-[12px] py-[10px]' : 'px-2.5 py-2',
            styles.thinkDropBtn,
          )}
        >
          <span className='text-[13px] text-[#6E757F] font-medium whitespace-nowrap'>
            {showThinking ? '思考和行动过程' : collapsedText}
          </span>
          <DropIcon
            style={{
              transform: `rotate(${showThinking ? '0deg' : '180deg'})`,
            }}
            className={cn('flex-shrink-0', styles.thinkDropBtnIcon)}
          />
        </div>
        {showThinking && (
          <div
            ref={thinkingContainerRef}
            onWheel={(e) => {
              // 用户向上滚动时，停止自动滚动（复用问答部分的滚动逻辑）
              if (e.deltaY < 0) {
                shouldScrollThinking.current = false
              }
            }}
            className={cn(
              'px-[12px] pb-[12px]',
              'text-[#374151] leading-[1.5em] font-[400] text-[14px] whitespace-pre-wrap',
              'max-h-[200px] overflow-y-auto',
              styles.thinkingContent,
            )}
          >
            {thinkingContentValue}
          </div>
        )}
      </div>
    )
  }

  const handleReferenceClick = () => {
    if (
      drawerVisible &&
      EDrawerType.REFERENCE === drawerType &&
      dialogIndex === index
    ) {
      handleCloseDrawer()
      return
    }
    handleOpenReference(props.index)
  }

  const renderReferenceText = () => {
    return (
      <div className='flex'>
        <div
          onClick={handleReferenceClick}
          className={cn(
            'flex gap-[4px] items-center h-[28px] cursor-pointer',
            'rounded-[6px] bg-[#0000000D] font-[500]  text-[14px] text-[#374151] px-[6px]',
          )}
        >
          {referenceText}
          <DropLeftIcon />
        </div>
      </div>
    )
  }

  const handleCopy = async () => {
    try {
      await copyToClipboard(
        markdownToText(content.replace(/\{Reference §(\d+)\[[^\]]+\]}/g, '')),
      )
      message.success('复制成功')
    } catch (e) {
      console.error(e)
      return
    }
  }

  // 获取调用工具的显示文本
  const getToolText = (messageType: AgentStage['messageType']) => {
    switch (messageType) {
      case 'file':
        return '读取文件内容'
      case 'code':
        return '正在执行代码'
      case 'chart':
        return '正在生成图表'
      case '':
        return '资源处理中'
      default:
        return ''
    }
  }

  /**
   * 判断阶段是否展开
   * expanded !== false 表示展开或未初始化（未初始化时默认展开）
   * 使用 messageId 作为唯一标识
   */
  const isStageExpanded = (stage: AgentStage) => {
    const stageId = (stage as any).messageId || `order-${stage.messageOrder}`
    return agentStageExpanded[stageId] !== false
  }

  /**
   * 记录已自动收起的阶段的唯一标识
   * 使用 `${index}-${messageOrder}` 作为唯一标识,避免不同对话之间的冲突
   * 防止用户手动展开后，因为数据更新导致再次被自动收起
   */
  const autoCollapsedStages = useRef<Set<string>>(new Set())

  // 切换会话时，重置自动收起记录和展开状态
  useEffect(() => {
    autoCollapsedStages.current.clear()
    setAgentStageExpanded({})
  }, [session_id])

  /**
   * 初始化 Agent 阶段的展开状态
   * 规则：
   * 1. 如果思考过程已完成（isFinal=true），且未自动收起过，则自动收起
   * 2. 如果还未初始化（undefined），默认展开并启用自动滚动
   * 3. 只处理 task_thought 类型的阶段（思考过程阶段）
   * 4. 使用 messageId 作为唯一标识，避免不同阶段的展开状态冲突
   */
  useEffect(() => {
    if (value.agentStages) {
      value.agentStages.forEach((stage) => {
        if (stage.messageType === 'task_thought') {
          const messageId = (stage as any).messageId
          const stageId = messageId || `order-${stage.messageOrder}`
          const stageKey = `${index}-${stageId}`

          // 如果思考过程已完成，且尚未执行过自动收起，则自动收起
          if (stage.isFinal) {
            if (!autoCollapsedStages.current.has(stageKey)) {
              setAgentStageExpanded((prev) => ({
                ...prev,
                [stageId]: false,
              }))
              autoCollapsedStages.current.add(stageKey)
            }
          } else if (agentStageExpanded[stageId] === undefined) {
            // 如果还未初始化，默认展开并启用自动滚动
            setAgentStageExpanded((prev) => ({
              ...prev,
              [stageId]: true,
            }))
            agentStageShouldScroll.current[stageId] = true
          }
        }
      })
    }
  }, [value.agentStages, index])

  /**
   * 监听思考内容变化，自动滚动到底部
   * 当思考内容更新时，如果阶段是展开状态且允许自动滚动，则滚动到底部
   * 只处理 task_thought 类型的阶段
   */
  useEffect(() => {
    if (!value.agentStages) return

    value.agentStages.forEach((stage) => {
      if (stage.messageType === 'task_thought') {
        const messageId = (stage as any).messageId
        const stageId = messageId || `order-${stage.messageOrder}`
        const dom = agentStageScrollRefs.current[stageId]
        const expanded = isStageExpanded(stage)
        const shouldScroll = agentStageShouldScroll.current[stageId]

        // 只有在展开状态且允许自动滚动时才滚动
        if (dom && expanded && shouldScroll) {
          scrollToEnd(dom)
        }
      }
    })
  }, [value.agentStages, agentStageExpanded])

  /**
   * 渲染单个思考过程阶段（task_thought 类型）
   * 支持展开/收起功能，展开时显示思考内容，收起时只显示标题
   * 思考内容支持自动滚动，用户手动向上滚动后停止自动滚动
   */
  const renderThinkingStage = (stage: AgentStage) => {
    const messageId = (stage as any).messageId
    const stageId = messageId || `order-${stage.messageOrder}`
    const expanded = isStageExpanded(stage)

    return (
      <div
        key={`thinking-${stageId}`}
        className={cn(
          'border border-[#EFF1F4] rounded-[6px]',
          // 展开时：使用 w-fit 让宽度根据内容自适应，最大宽度由父容器限制
          // 收起时：固定宽度 171px
          expanded ? 'w-fit' : 'w-[171px]',
          styles.thinkingContainer,
        )}
      >
        {/* 展开/收起按钮 */}
        <div
          onClick={() => {
            setAgentStageExpanded((prev) => ({
              ...prev,
              [stageId]: !expanded,
            }))
          }}
          className={cn(
            'flex items-center justify-between cursor-pointer',
            expanded ? 'px-[12px] py-[10px]' : 'px-2.5 py-2',
            styles.thinkDropBtn,
          )}
        >
          <div className='flex gap-[4px] items-center'>
            <ThinkingIcon className='w-[14px] h-[14px] flex-shrink-0' />
            <span className='text-[14px] text-[#6E757F] font-medium whitespace-nowrap'>
              {expanded
                ? '思考和行动过程'
                : stage.isFinal
                  ? '已完成思考'
                  : '思考和行动过程'}
            </span>
          </div>
          <DropIcon
            style={{
              transform: `rotate(${expanded ? '0deg' : '180deg'})`,
            }}
            className={cn('flex-shrink-0', styles.thinkDropBtnIcon)}
          />
        </div>
        {/* 思考内容区域：展开时显示，支持滚动 */}
        {expanded && (
          <div
            ref={(el) => {
              // 保存 DOM 引用，用于自动滚动
              agentStageScrollRefs.current[stageId] = el
            }}
            onWheel={(e) => {
              // 用户向上滚动时（deltaY < 0），停止自动滚动
              // 这样用户查看历史内容时不会被自动滚动打断
              if (e.deltaY < 0) {
                agentStageShouldScroll.current[stageId] = false
              }
            }}
            className={cn(
              'px-[12px] pb-[12px]',
              'text-[#6E757F] leading-[1.5em] font-[400] text-[13px] whitespace-pre-wrap',
              'max-h-[100px] overflow-y-auto',
              styles.thinkingContent,
            )}
          >
            {stage.taskThought}
          </div>
        )}
      </div>
    )
  }

  /**
   * 渲染调用工具阶段（file、code、chart 等类型）
   * 显示工具调用的状态信息，不支持展开/收起
   */
  const renderToolStage = (stage: AgentStage) => {
    const toolText = getToolText(stage.messageType)
    if (!toolText) return null

    return (
      <div
        key={`tool-${stage.messageOrder}`}
        className='border border-[#EFF1F4] rounded-[6px] px-[10px] py-[8px] self-start'
      >
        <div className='flex gap-[4px] items-center whitespace-nowrap'>
          <ToolIcon className='w-[14px] h-[14px] flex-shrink-0' />
          <span className='text-[14px] text-[#6E757F] font-medium'>
            调用工具
          </span>
          <span className='text-[12px] text-[#6E757F] font-medium'>
            {toolText}
          </span>
        </div>
      </div>
    )
  }

  /**
   * 渲染问题改写阶段 (question_rewrite)
   * 仅在 本地环境、测试环境、custom 版本的 cimc 模式，且为文档问答时展示
   */
  const renderQuestionRewriteStage = (
    stage: AgentStage,
    deployConfig: any,
    sessionConfig: any,
  ) => {
    const isCimcMode =
      deployConfig.mode === 'cimc' ||
      import.meta.env.DEV ||
      import.meta.env.MODE === 'test'
    const isKnowledgeMode =
      sessionConfig?.mode === 'knowledge' || sessionConfig?.mode === undefined

    // 仅在本地/测试环境，或 cimc 部署环境下的文档问答模式展示
    if (!isCimcMode || !isKnowledgeMode) return null

    return (
      <div
        key={`rewrite-${stage.messageOrder}`}
        className='text-[13px] text-[#6E757F] leading-[1.5em] flex gap-1'
      >
        <span className='font-medium flex-shrink-0'>改写后问题：</span>
        <span className='break-all'>{stage.taskThought}</span>
      </div>
    )
  }

  /**
   * 渲染所有 Agent 阶段
   * 根据 messageType 判断阶段类型：
   * - task_thought: 思考过程阶段，调用 renderThinkingStage
   * - question_rewrite: 问题改写阶段，展示纯文本
   * - 其他（file、code、chart 等）: 工具调用阶段，调用 renderToolStage
   */
  const renderAgentStages = () => {
    if (!value.agentStages || value.agentStages.length === 0) return null

    return (
      <div className='flex flex-col gap-[8px]'>
        {value.agentStages.map((stage) => {
          if (stage.messageType === 'task_thought') {
            return renderThinkingStage(stage)
          } else if (stage.messageType === 'question_rewrite') {
            return renderQuestionRewriteStage(
              stage,
              deployConfig,
              sessionConfig,
            )
          } else {
            return renderToolStage(stage)
          }
        })}
      </div>
    )
  }

  // 渲染图表按钮
  const renderChartButton = () => {
    if (status !== 'answered' || !value.hasCharts) return null

    // 判断仪表盘侧边栏是否已打开
    const isChartDrawerOpen = drawerVisible && drawerType === EDrawerType.CHART

    const buttonContent = (
      <div
        className={cn(
          'bg-[#FBE9FF] rounded-[6px] px-[6px] py-[4px] cursor-pointer self-start',
          // 如果仪表盘已打开，按钮置灰；未打开时高亮显示
          isChartDrawerOpen ? 'opacity-[0.64]' : '',
        )}
        onClick={() => handleOpenChart?.()}
      >
        <span className='text-[14px] text-[#CC5DE8] font-normal whitespace-nowrap'>
          图表绘制成功，前往「仪表盘」查看
        </span>
      </div>
    )

    // 如果仪表盘已打开，显示 tooltip
    if (isChartDrawerOpen) {
      return (
        <Tooltip title='已打开仪表盘' placement='top'>
          {buttonContent}
        </Tooltip>
      )
    }

    return buttonContent
  }

  /**
   * 渲染引用文件列表
   * 从 content 中提取引用标记（格式：{Reference §文件ID[...]}），
   * 然后从 reference 数组中过滤出真正被引用的文件并显示
   */
  const renderReferenceFiles = () => {
    // 检查 content 中是否包含引用标记，如果没有则不显示引用资料
    if (!content?.includes('{Reference')) return null
    if (referencedChunks.length === 0) return null

    return (
      <div className='mt-4 inline-block'>
        <div
          className='inline-flex items-center rounded-md bg-[#0000000D] gap-2 px-[6px] py-[4px] mb-3 cursor-pointer'
          onClick={() => setReferenceExpanded(!referenceExpanded)}
        >
          <span className='text-sm font-medium text-[#374151]'>
            引用{referencedFiles.length}篇资源
          </span>
          {referenceExpanded ? (
            <UpIcon className='w-4 h-4 text-[#6B7280]' />
          ) : (
            <DownIcon className='w-4 h-4 text-[#6B7280]' />
          )}
        </div>
        {referenceExpanded && (
          <div className='space-y-1'>
            {referencedFiles.map((item) => (
              <div
                key={`${item.file_id}`}
                className='flex items-center gap-2 py-1'
              >
                <Link
                  to={`/docs/detail/${item.forest_id}/file/${item.file_id}`}
                  target='_blank'
                  className='inline-flex items-center gap-2 cursor-pointer'
                >
                  <span className='inline-flex items-center gap-1'>
                    {item.annotations.map((annotation) => (
                      <span
                        key={`${item.file_id}-${annotation}`}
                        className='inline-flex items-center justify-center !w-[16px] !h-[16px] text-[#ffffff] text-[12px] font-normal bg-[#bbb] relative rounded-[50%] leading-none'
                      >
                        {annotation}
                      </span>
                    ))}
                  </span>
                  {getFileIcon(item.file_name) || (
                    <DocIcon className='w-4 h-4 text-[#6B7280] flex-shrink-0' />
                  )}
                  <span className='text-sm text-[#374151] font-medium truncate hover:text-[#1A79FF]'>
                    {item.file_name}
                  </span>
                </Link>
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className={cn('flex gap-4', className)} style={style}>
      <img src={icon} className='w-9 h-9' />
      <div className='flex flex-col gap-2.5 overflow-hidden max-w-[calc(100%-102px)]'>
        <span className='flex items-center'>
          <span className='font-medium'>
            {match(version)
              .with('saas', () => 'CoreKG AI')
              .with('custom', () => 'AI')
              .with('international', () => tC('project.opoAI'))
              .exhaustive()}
          </span>
          {sessionConfig?.mode === 'h3c-test' ? null : (
            <>
              <img
                src={currentModel?.avatar || DeepSeek}
                className='w-4 h-4 ml-2.5'
              />
              <span className='text-[#919497] text-xs ml-1 mr-2 font-medium'>
                {currentModel?.name}
              </span>
            </>
          )}

          {match({ mode: sessionConfig?.mode, status: value.status, graph })
            .with(
              {
                mode: P.not(P.union('graph', 'graph_search')),
              },
              () => null,
            )
            .with(
              {
                status: P.not('answered'),
              },
              () => (
                <div
                  className={cn(
                    'bg-[#FBE9FF] text-[#CC5DE8] text-xs p-1.5 rounded',
                    'flex gap-1 items-center',
                  )}
                >
                  <ConfigProvider
                    theme={{
                      components: { Spin: { colorPrimary: '#CC5DE8' } },
                    }}
                  >
                    <Spin size='small' />
                  </ConfigProvider>
                  知识图谱洞察中
                </div>
              ),
            )
            .with({ graph: { graph_reference: P.nonNullable } }, () => (
              <div
                className={cn(
                  'bg-[#FBE9FF] text-[#CC5DE8] text-xs p-1.5 rounded cursor-pointer',
                  'flex gap-1 items-center',
                )}
                onClick={() => {
                  if (
                    drawerVisible &&
                    EDrawerType.GRAPH === drawerType &&
                    dialogIndex === index
                  ) {
                    handleCloseDrawer()
                    return
                  }
                  handleOpenGraph(index)
                }}
              >
                知识图谱洞察完成
                <RightOutlined />
              </div>
            ))
            .otherwise(() => (
              <div
                className={cn(
                  'bg-[#0000000D] font-[500] text-[14px] text-[#374151] p-1.5 rounded',
                )}
              >
                暂无洞察结果
              </div>
            ))}
        </span>
        {showLoading ? (
          <div
            className='flex items-center gap-2 text-xs text-[#919497]'
            aria-live='polite'
          >
            <span
              className='inline-block w-3.5 h-3.5 rounded-full border-2 border-[#E5E7EB] border-t-[#B97CFF] animate-spin'
              aria-label={t('status.loading')}
            />
            <span>{t('status.loading')}</span>
          </div>
        ) : null}
        {/* 
          兼容新旧两种思考内容显示方式：
          1. 新方式：如果有 agentStages，显示多阶段思考过程和调用工具（支持多个思考阶段和工具调用）
          2. 旧方式：如果没有 agentStages 但有 thinkingContent，显示旧的单阶段思考过程
        */}
        {value.agentStages && value.agentStages.length > 0
          ? renderAgentStages()
          : thinkingContent
            ? renderThinkingContent()
            : null}
        {referenceText ? renderReferenceText() : null}
        {content ? (
          <MarkdownPreview
            references={reference}
            // 单文件问答页也要展示正文里的数字引用标注，并复用统一的点击预览/跳转逻辑
            disableReference={!reference?.length}
            content={content}
            className='p-0! bg-white!'
          />
        ) : null}
        {/* 在正文后显示图表按钮 */}
        {renderChartButton()}
        {status === 'answered' && renderReferenceFiles()}
        {status === 'answered' && (
          <div className='dialog-footer'>
            <Tooltip title='复制'>
              <CopyIcon className='cursor-pointer' onClick={handleCopy} />
            </Tooltip>
          </div>
        )}
      </div>
    </div>
  )
}
