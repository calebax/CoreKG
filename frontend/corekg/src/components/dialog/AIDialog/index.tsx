import { FC, useState, useEffect } from 'react'
import { Spin, Tooltip, Typography } from 'antd'
import { cn } from '@/utils'
import { ConcatUs } from '@/components/ConcatUs'
import EchartsEditModal from '@/components/common/EchartsEditModal'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import { markdownToText } from '@/utils/markdownToText'
import { ReferenceFiles } from './ReferenceFiles'
import { ThinkingContent } from './ThinkingContent'
import Avatar from './avatar.svg?react'
import CopyIcon from './images/copy.svg?react'

export type Chunk = {
  chunk_id: string
  content: string
  image_url: string
  location: number[]
  score: number
  sequence: number
  type: 'image' | 'video' | 'table'
}

/** Agent 消息的阶段类型 */
export type AgentStage = {
  /** 阶段类型：task_thought=思考过程, file=读取文件, code=执行代码, chart=生成图表, question_rewrite=问题改写, ''=资源处理中 */
  messageType:
    | 'task_thought'
    | 'file'
    | 'code'
    | 'chart'
    | 'question_rewrite'
    | ''
  /** 思考内容（task_thought 或 question_rewrite 类型有值，流式更新） */
  taskThought: string
  /** 是否完成当前阶段 */
  isFinal: boolean
  /** 消息序号 */
  messageOrder: number
  /** 图表配置（仅 chart 类型有值） */
  chartConfig?: any
}

export type AIDialog = {
  role: 'answer'
  /** 深度思考内容 */
  thinkingContent: string
  /** 回答内容 */
  content: string
  /**
   * 思考状态
   * @enum
   * - 'thinking'：正在思考
   * - 'search'：正在搜索
   * - 'found'：已经找到参考资料
   * - ''：普通文本
   * - 'answering'：回答开始
   * - undefined：回答结束
   * - 'answered' 回答结束
   */
  status:
    | 'thinking'
    | 'search'
    | 'found'
    | ''
    | 'answering'
    | undefined
    | 'answered'
  /** 参考文献 */
  reference: {
    forest_id: number
    file_id: number
    file_name: string
    chunk_list?: Chunk[]
    user_name?: string
    created_at?: string
  }[]
  sub_question?: string[]
  graph?: any
  /** Agent 消息的阶段列表（flag=agent 时使用） */
  agentStages?: AgentStage[]
  /** 是否有图表（用于显示图表按钮） */
  hasCharts?: boolean
}

type AIDialogProps = {
  className?: string
  value: AIDialog
  /** 展示参考文献 */
  showReference?: boolean
  showConcatus?: boolean
}
/**
 * @example
 * ```js
 * const dialog:(UserDialog|AIDialog)[]
 *
 * dialog.map((item, i)=>{
 *    if(item.role==='question')return <UserDialog key={i} value={item}/>
 *    else return <AIDialog key={i} value={item}/>
 * })
 * ```
 */
export const AIDialog: FC<AIDialogProps> = (props) => {
  const { className, value, showReference, showConcatus } = props
  const { reference, thinkingContent, content, status } = value
  // 回答是否已经结束
  const answered = status === 'answered'
  // 参考文献是否loading
  const referenceLoading = !answered && !thinkingContent && !content
  // 思考过程是否loading
  const thinkingLoading = !answered && !content

  // ECharts编辑状态
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [currentEditConfig, setCurrentEditConfig] = useState<any>(null)
  const [currentEditId, setCurrentEditId] = useState<string>('')
  const [editedContent, setEditedContent] = useState(content)
  const [chartConfigMap, setChartConfigMap] = useState<Map<string, any>>(
    new Map(),
  )

  // 当原始content变化时，更新editedContent
  useEffect(() => {
    setEditedContent(content)
  }, [content])

  // 处理ECharts编辑
  const handleEchartsEdit = (id: string, config: any) => {
    setCurrentEditId(id)
    setCurrentEditConfig(config)
    setEditModalVisible(true)
  }

  // 处理ECharts保存
  const handleEchartsSave = (newConfig: any) => {
    // 保存配置到映射表
    const newChartConfigMap = new Map(chartConfigMap)
    newChartConfigMap.set(currentEditId, newConfig)
    setChartConfigMap(newChartConfigMap)

    setEditModalVisible(false)
    setCurrentEditConfig(null)
    setCurrentEditId('')
  }

  return (
    <>
      <div className={cn('overflow-hidden flex-none flex gap-4 ', className)}>
        <Avatar />
        <div className='mt-2.5 overflow-hidden flex-1 flex flex-col gap-4'>
          {showReference ? (
            <ReferenceFiles loading={referenceLoading} files={reference} />
          ) : null}
          {/* 无参考文献时 改为展示loading */}
          {referenceLoading && !showReference ? (
            <Spin className='self-start' />
          ) : null}
          <ThinkingContent
            loading={thinkingLoading}
            content={thinkingContent}
          />
          {editedContent || content ? (
            <MarkdownPreview
              content={editedContent || content}
              references={reference}
              onEchartsEdit={handleEchartsEdit}
              chartConfigMap={chartConfigMap}
              disableReference={!showReference}
            />
          ) : null}
          <span className='flex items-center gap-4 justify-start'>
            {answered && (editedContent || content) ? (
              <Tooltip title='复制' placement='top'>
                <Typography.Paragraph
                  className='m-0 mt-1 hover:bg-[#f8f9fd]'
                  copyable={{
                    icon: [<CopyIcon />, <CopyIcon />],
                    text: markdownToText(editedContent || content),
                  }}
                ></Typography.Paragraph>
              </Tooltip>
            ) : null}
            {answered && showConcatus ? <ConcatUs /> : null}
          </span>
        </div>
      </div>

      {/* ECharts编辑弹窗 */}
      {currentEditConfig && (
        <EchartsEditModal
          visible={editModalVisible}
          onClose={() => setEditModalVisible(false)}
          onSave={handleEchartsSave}
          initialConfig={currentEditConfig}
        />
      )}
    </>
  )
}
