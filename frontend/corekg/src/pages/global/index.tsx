import { FC, useState, useRef, useEffect, useLayoutEffect } from 'react'
import { message, Upload, Tooltip } from 'antd'
import { Input } from 'antd'
import { CloseCircleFilled, LoadingOutlined } from '@ant-design/icons'
import { useMemoizedFn } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { useImmer } from 'use-immer'
import { cn, scrollToEnd } from '@/utils'
import {
  createSession,
  createStream,
  sendStream,
  getQuestionInfo,
} from '@/api/agent'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import { getFileIcon as getKnowledgeFileIcon } from '@/components/common/ReferencePreviewModal/utils'
import {
  AttachmentList,
  type AIDialog,
  type UserDialog,
  type DialogList,
  type Attachment,
} from '@/components/dialog'
import { getChunkFromStream, updateDialog } from '@/components/dialog/utils'
import EmptyProjectIcon from '@/pages/project/ProjectContent/EmptyProject/images/icon.svg?react'
import styles from '@/pages/project/ProjectContent/EmptyProject/styles.module.scss'
import icon from '@/pages/project/ProjectContent/SessionChat/Dialog/AIDialog/icon.svg'
import CopyIcon from '@/pages/project/ProjectContent/SessionChat/Dialog/AIDialog/images/copy.svg?react'
import ArrowRight from '@/pages/project/ProjectContent/SessionChat/Dialog/arrow-right.svg?react'
import { ActiveBtn, DisabledBtn } from '@/pages/project/components/DialogInput'
import inputStyles from '@/pages/project/components/DialogInput/styles.module.scss'
import useLocalStore from '@/stores/local'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import { copyToClipboard } from '@/utils/copy'
import { markdownToText } from '@/utils/markdownToText'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useQaInputMaxLength } from '@/utils/useQaInputMaxLength'

// 临时需求：使用后端下发的 token（由登录流程写入本地 store）。禁止在源码中硬编码 token。
const TEMP_TOKEN = process.env.VITE_TEMP_TOKEN ?? ''

// TODO: 临时需求 - 模拟用户信息，后续需要删除或注释
const MOCK_USER = {
  name: 'user',
  avatar: '/images/user-defaultIcon.svg',
}

// 简化的AI对话组件（不显示引用资源等）
const SimpleAIDialog: FC<{
  value: AIDialog
  onQuestionClick?: (question: string) => void
}> = ({ value, onQuestionClick }) => {
  const { status, sub_question } = value
  const content = value.content?.replace(
    /http:\/\/39.175.132.229:18020\/corekg-bucket\//g,
    'https://pilot.turingcm.com:18020/corekg-bucket/',
  )
  const showLoading = !content && status !== 'answered'

  const handleCopy = useMemoizedFn(async () => {
    try {
      const cleanedContent = content.replace(
        /\{Reference §(\d+)\[([\d,]+)\]\}/g,
        '',
      )
      await copyToClipboard(markdownToText(cleanedContent))
      message.success('复制成功')
    } catch {
      // 复制失败，静默处理
    }
  })

  return (
    <div className='flex gap-2.5'>
      <img src={icon} className='w-9 h-9' />
      <div className='flex flex-col gap-2.5 overflow-hidden flex-1'>
        <span className='font-medium'>Pilot</span>
        {showLoading && (
          <div className='flex items-center gap-2 text-xs text-[#919497]'>
            <span className='inline-block w-3.5 h-3.5 rounded-full border-2 border-[#E5E7EB] border-t-[#B97CFF] animate-spin' />
            <span>加载中...</span>
          </div>
        )}
        {content && (
          <MarkdownPreview
            content={content}
            className='p-0! bg-white!'
            disableReference={true}
          />
        )}
        {status === 'answered' && (
          <div className='dialog-footer'>
            <CopyIcon className='cursor-pointer' onClick={handleCopy} />
          </div>
        )}
        {status === 'answered' && sub_question && sub_question.length > 0 && (
          <div className='mt-4'>
            {sub_question.map((question, index) => (
              <div key={index} className='mb-[6px] flex items-center'>
                <div
                  className='min-h-[30px] flex items-center gap-[6px] leading-[30px] text-sm text-[#0C1F17] rounded-[6px] px-[10px] py-[4px] bg-[#0000000D] cursor-pointer whitespace-normal break-words hover:bg-[#0000001A] transition-colors'
                  onClick={() => {
                    onQuestionClick?.(question)
                  }}
                >
                  <span className='flex-1'>{question}</span>
                  <ArrowRight className='flex-shrink-0' />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// 简化的用户对话组件
const SimpleUserDialog: FC<{ value: UserDialog }> = ({ value }) => {
  const { content, created_at, attachments } = value
  const time = dayjs(created_at).format('YYYY/MM/DD HH:mm')

  return (
    <div className='flex gap-2.5 flex-row-reverse'>
      <img src={MOCK_USER.avatar} className='w-9 h-9 rounded-full' />
      <div className='flex flex-col gap-2.5'>
        <span className='flex items-center justify-end'>
          <span className='font-medium'>{MOCK_USER.name}</span>
          <span className='text-[#919497] text-xs ml-1'>{time}</span>
        </span>
        {/* 附件展示 */}
        <AttachmentList attachments={attachments || []} />
        {content && (
          <MarkdownPreview content={content} className='p-0! bg-white!' />
        )}
      </div>
    </div>
  )
}

// 简化的输入框组件（不依赖 ProjectContext）
const SimpleDialogInput: FC<{
  value?: string
  onChange?: (val: string) => void
  onEnter?: (val: string) => void
  placeholder?: string
  className?: string
  children?: React.ReactNode
  leftActions?: React.ReactNode
  attachments?: Attachment[]
  onRemoveAttachment?: (index: number) => void
  defaultRows?: number
  borderColor?: string
  showShadow?: boolean
  maxLength?: number
}> = ({
  value,
  onChange,
  onEnter,
  placeholder,
  className,
  children,
  leftActions,
  attachments = [],
  onRemoveAttachment,
  defaultRows = 4,
  borderColor = 'rgb(230,178,243)',
  showShadow = true,
  maxLength = 500,
}) => {
  const lineHeight = 24
  const [inputHeight, setInputHeight] = useState(defaultRows * lineHeight)
  const textAreaRef = useRef<HTMLTextAreaElement>()

  // 动态计算输入框高度
  useLayoutEffect(() => {
    const textarea = textAreaRef.current
    if (!textarea) return

    const defaultHeight = defaultRows * lineHeight
    const maxHeight = 5 * lineHeight

    // 如果内容为空，使用默认行数
    if (!value?.trim()) {
      setInputHeight(defaultHeight)
      return
    }

    // 根据内容高度动态调整，但不超过5行
    if (textarea.scrollHeight <= defaultHeight) {
      setInputHeight(defaultHeight)
    } else {
      setInputHeight(Math.min(textarea.scrollHeight, maxHeight))
    }
  }, [value, defaultRows])

  return (
    <div
      className={cn(
        'bg-white p-3 rounded-xl',
        'overflow-auto transition-all duration-300 ease-in-out',
        'relative flex flex-col',
        showShadow
          ? 'shadow-[0_0_10px_rgba(0,0,0,0.1)] focus-within:shadow-[0_0_20px_rgba(0,0,0,0.15)]'
          : '',
        className,
      )}
      style={{
        borderColor,
        borderWidth: '1px',
        borderStyle: 'solid',
      }}
    >
      {/* 附件展示区域 */}
      <AttachmentList
        attachments={attachments}
        onRemove={onRemoveAttachment}
        canRemove
        isCompact
      />

      <Input.TextArea
        maxLength={maxLength}
        value={value}
        onChange={(e) => {
          onChange?.(e.target.value)
        }}
        className={cn(inputStyles.input, 'placeholder:text-[#919497]')}
        style={{
          height: inputHeight,
          maxHeight: inputHeight,
          lineHeight: `${lineHeight}px`,
        }}
        placeholder={placeholder}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            if (e.shiftKey || e.ctrlKey || e.altKey) {
              // shift/ctrl/alt+回车: 换行
              return
            } else {
              // 回车: 发送消息
              e.preventDefault()
              onEnter?.(value ?? '')
            }
          }
        }}
      />
      <div className='w-full h-0 overflow-hidden'>
        {/* 用于测量高度 */}
        <Input.TextArea
          ref={(el) => {
            const textarea = el?.resizableTextArea?.textArea
            if (textarea) {
              textAreaRef.current = textarea
            }
          }}
          value={value}
          className={cn(inputStyles.input, 'placeholder:text-[#919497]')}
          style={{
            lineHeight: `${lineHeight}px`,
          }}
        />
      </div>
      <div className='w-full relative flex-none mt-2.5 h-[30px] flex justify-between items-center gap-1'>
        <div className='flex items-center gap-2'>{leftActions}</div>
        <div className='flex items-center gap-1'>{children}</div>
      </div>
    </div>
  )
}

const GlobalPage: FC = () => {
  const { t } = useTranslation('pages')
  const { version, globalGreeting, mode } = useDeployConfig()
  const qaInputMaxLength = useQaInputMaxLength()
  const [dialog, setDialog] = useImmer<DialogList>([])
  const [dialogStatus, setDialogStatus] = useState<
    'ready' | 'asking' | 'answering'
  >('ready')
  const [sessionId, setSessionId] = useState<number>()
  const dialogContentRef = useRef<HTMLDivElement>(null)
  const lastQuestionId = useRef<number>()

  // 根据部署配置获取应用名称：custom 版本显示 TuringQuery，其他版本显示 CoreKG
  const app_name = version === 'custom' ? '芯模Pilot' : 'CoreKG'

  // 自动滚动到底部
  useEffect(() => {
    if (dialogContentRef.current) {
      scrollToEnd(dialogContentRef.current)
    }
  }, [dialog])

  // 创建H3C测试会话
  const createH3CTest = useMemoizedFn(async () => {
    if (TEMP_TOKEN) {
      useLocalStore.getState().setToken(TEMP_TOKEN)
    }

    const res = await createSession({
      model_id: 1,
      resource_id: 663,
      name: 'new chat',
      resource_type: 'agent',
      source_from: 'agent',
    })
    return res.ID
  })

  // 更新答案状态
  const updateAnswerStatus = useMemoizedFn(
    (index: number, status: AIDialog['status'], content?: string) => {
      setDialog((draft) => {
        const answer = draft[index] as AIDialog
        if (answer?.role === 'answer') {
          answer.status = status
          if (content) {
            answer.content = content
          }
        }
      })
    },
  )

  // 开始问答
  const startQA = useMemoizedFn(async (text: string) => {
    if (!text?.trim() || dialogStatus !== 'ready') return

    // TODO: 临时需求 - 使用临时token，后续需要替换为后端提供的token
    if (TEMP_TOKEN) {
      useLocalStore.getState().setToken(TEMP_TOKEN)
    }

    setDialogStatus('asking')

    // 先计算索引，再更新状态
    let currentAnswerIndex = 0
    setDialog((draft) => {
      const currentLength = draft.length
      draft.push(
        {
          role: 'question',
          content: text,
          created_at: dayjs().toString(),
        },
        {
          role: 'answer',
          content: '',
          thinkingContent: '',
          status: 'thinking',
          reference: [],
        },
      )
      currentAnswerIndex = currentLength + 1
    })

    // 获取或创建会话ID
    const currentSessionId = sessionId || (await createH3CTest())
    if (!sessionId) {
      setSessionId(currentSessionId)
    }

    // 构造请求参数
    const requestBody: any = {
      session_id: currentSessionId,
      question: text,
    }

    try {
      // 创建流式请求
      const { question_id } = (await createStream(requestBody)) as any
      lastQuestionId.current = question_id

      const { body } = (await sendStream({
        session_id: currentSessionId,
        question_id,
      })) as any

      setDialogStatus('answering')

      // 使用工具函数处理流式响应
      await getChunkFromStream(
        body,
        (type, data) => {
          if (type === 'content') {
            updateDialog(data, setDialog, currentAnswerIndex)
            if (data.flag === undefined) {
              updateAnswerStatus(currentAnswerIndex, 'answered')
              setDialogStatus('ready')
            }
          } else if (type === 'limited') {
            updateAnswerStatus(
              currentAnswerIndex,
              'answered',
              '抱歉，您的额度已用完，请联系管理员。',
            )
            setDialogStatus('ready')
          }
        },
        async () => {
          // 流结束，获取推荐问题
          let questions: string[] | null = null
          if (question_id) {
            try {
              const res = await getQuestionInfo(question_id)
              const {
                question: {
                  _source: { sub_question },
                },
              } = res as any
              questions = sub_question || null
            } catch (error) {
              console.log('获取推荐问题失败:', error)
            }
          }

          // 更新状态为 'answered'，显示复制图标和推荐问题
          setDialog((draft) => {
            const answer = draft[currentAnswerIndex] as AIDialog
            if (answer?.role === 'answer') {
              answer.status = 'answered'
              if (questions?.length) {
                answer.sub_question = questions
              }
            }
          })
          setDialogStatus('ready')
        },
      )
    } catch (error) {
      console.log('问答失败:', error)
      updateAnswerStatus(
        currentAnswerIndex,
        'answered',
        '抱歉，回答失败，请稍后重试。',
      )
      setDialogStatus('ready')
    }
  })

  const [inputValue, setInputValue] = useState<string>()
  const isFirstQuestion = dialog.length === 0

  const handleSend = useMemoizedFn(() => {
    if (!inputValue?.trim() || dialogStatus !== 'ready') return
    const text = inputValue.trim()
    setInputValue('')
    startQA(text)
  })

  // 渲染发送按钮
  const renderSendButton = () => {
    if (inputValue?.trim() && dialogStatus === 'ready') {
      return <ActiveBtn className='cursor-pointer' onClick={handleSend} />
    }
    return <DisabledBtn />
  }

  return (
    <div className='w-full h-full flex flex-col bg-[#ffffff] overflow-hidden'>
      {/* 欢迎页面或对话页面 */}
      <div className='flex-1 flex flex-col overflow-hidden p-6 pb-0'>
        {isFirstQuestion ? (
          // 欢迎页面 - 参考 EmptyProject 的样式
          <div className={cn('w-full h-full overflow-hidden flex flex-col')}>
            {/* 顶部空白区域，用于自适应分配上下间距 */}
            <div className='flex-1' />
            <EmptyProjectIcon className='mx-auto' />
            <div
              className={cn(
                styles.title,
                'text-center whitespace-pre-line mx-auto mt-2.5 font-semibold',
              )}
            >
              {globalGreeting ??
                t('project.greeting', { project_name: '', app_name })}
            </div>
            <SimpleDialogInput
              className={cn('w-[850px] mt-18 mx-auto')}
              value={inputValue}
              onChange={setInputValue}
              onEnter={handleSend}
              maxLength={qaInputMaxLength}
              placeholder='输入内容并发送（或按Enter），即可生成回答'
            >
              {renderSendButton()}
            </SimpleDialogInput>
            {/* 底部空白区域，用于自适应分配上下间距 */}
            <div className='flex-1' />
          </div>
        ) : (
          // 对话页面
          <div
            ref={dialogContentRef}
            className={cn(
              'flex-1 overflow-auto flex flex-col px-10 gap-8',
              scrollStyles.scroll,
            )}
          >
            {dialog.map((item, index) => {
              return item.role === 'question' ? (
                <SimpleUserDialog key={index} value={item as UserDialog} />
              ) : (
                <SimpleAIDialog
                  key={index}
                  value={item as AIDialog}
                  onQuestionClick={(question) => {
                    if (dialogStatus === 'ready') {
                      startQA(question)
                    }
                  }}
                />
              )
            })}
          </div>
        )}
      </div>

      {/* 输入框（仅在已有对话时显示） */}
      {!isFirstQuestion && (
        <div className='flex-shrink-0 p-6 pb-8 pt-5'>
          <SimpleDialogInput
            className={cn('mx-30')}
            value={inputValue}
            onChange={setInputValue}
            onEnter={handleSend}
            maxLength={qaInputMaxLength}
            placeholder='输入内容并发送（或按Enter），即可生成回答'
            defaultRows={3}
            borderColor='#E3E6ED'
            showShadow={false}
          >
            {renderSendButton()}
          </SimpleDialogInput>
        </div>
      )}
    </div>
  )
}

export default GlobalPage
