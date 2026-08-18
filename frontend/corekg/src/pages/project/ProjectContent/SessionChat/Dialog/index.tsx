import {
  FC,
  MutableRefObject,
  useRef,
  useEffect,
  useState,
  useCallback,
} from 'react'
import { Skeleton } from 'antd'
import { cn, scrollToEnd } from '@/utils'
import { AIDialog } from '@/components/dialog'
import { useSessionInfo } from '../..'
import { useProject } from '../../..'
import { ProjectInput } from '../../ProjectInput'
import { ProjectAIDialog } from './AIDialog'
import { ProjectUserDialog } from './UserDialog'
import ArrowRight from './arrow-right.svg?react'

export const Dialog: FC<Style> = (props) => {
  const { className, style } = props
  const { dialogStatus, startQA, sessionVisible, drawerVisible } =
    useSessionInfo()
  const { type } = useProject()
  const shouldScroll = useRef(true)
  const dialogContentRef = useRef<HTMLDivElement>(null)
  const [inputWidth, setInputWidth] = useState<number | undefined>(undefined)

  const isSingleFilePage = type === 'single-file'
  const isSingleFileWithDrawer = isSingleFilePage && drawerVisible
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)

  // 处理对话内容区域宽度变化
  const handleContentWidthChange = useCallback(
    (width: number) => {
      if (!isSingleFilePage) {
        setInputWidth(undefined)
        setIsPreviewOpen(false)
        return
      }

      // 通过容器宽度判断预览区域是否打开（小于窗口宽度80%认为预览区域打开）
      const previewOpen = width < window.innerWidth * 0.8
      setIsPreviewOpen(previewOpen)

      if (drawerVisible) {
        // 当侧边栏打开时，使用容器宽度
        setInputWidth(width)
      } else if (previewOpen) {
        // 当预览区域打开且侧边栏关闭时，使用消息内容的最大宽度
        setInputWidth(width - 170)
      } else {
        // 当预览区域关闭时，不设置固定宽度
        setInputWidth(undefined)
      }
    },
    [isSingleFilePage, drawerVisible],
  )

  useEffect(() => {
    if (!isSingleFilePage) {
      setInputWidth(undefined)
      setIsPreviewOpen(false)
    }
  }, [isSingleFilePage])

  // 计算输入框的左右间距
  const inputMarginX =
    isSingleFileWithDrawer ||
    (isSingleFilePage && !drawerVisible && isPreviewOpen)
      ? undefined
      : sessionVisible && drawerVisible
        ? 'mx-[82px]'
        : 'mx-[125px]'

  // 当两个侧边栏都打开时，使用紧凑模式（一列前后展示）
  const isCompact = sessionVisible && drawerVisible

  if (dialogStatus === 'loading') {
    return <Skeleton active className='mt-4 mx-auto w-4/5' />
  }
  return (
    <div className='flex-1 flex flex-col overflow-hidden'>
      <div
        ref={dialogContentRef}
        className={cn('flex-1 overflow-auto flex flex-col', className)}
        style={style}
      >
        <Dialogs
          shouldScroll={shouldScroll}
          isCompact={isCompact}
          onContentWidthChange={handleContentWidthChange}
          onSend={(text: string) => {
            startQA({ content: text })
            shouldScroll.current = true
          }}
        />
      </div>
      <ProjectInput
        className={cn(
          'mt-2 flex-sh flex-shrink-0',
          inputMarginX,
          isSingleFilePage ? 'mb-[20px]' : 'mb-10',
        )}
        // 只有当预览文件部分和侧边栏（drawer）都显示时，才显示 mr-3
        showRightMargin={isSingleFilePage && drawerVisible}
        style={{
          ...((isSingleFileWithDrawer ||
            (isSingleFilePage && !drawerVisible && isPreviewOpen)) &&
          inputWidth
            ? {
                width: `${inputWidth}px`,
                // 当预览区域打开且侧边栏关闭时，输入框应该和消息内容的最大宽度对齐
                // 消息内容区域从 84px 开始（32px padding + 36px 头像 + 16px gap）
                // 输入框左边缘与消息内容左边缘对齐
                marginLeft:
                  isSingleFilePage && !drawerVisible && isPreviewOpen
                    ? '84px'
                    : 'auto',
                // 输入框右边缘与消息内容右边缘对齐（通过宽度自动对齐，不需要设置 marginRight）
                marginRight: 'auto',
              }
            : {}),
          ...(isSingleFilePage ? { marginBottom: '20px' } : {}),
        }}
        onAsk={() => {
          shouldScroll.current = true
        }}
      />
    </div>
  )
}

const Dialogs: FC<{
  shouldScroll: MutableRefObject<boolean>
  onSend?: (text: string) => void
  isCompact?: boolean
  onContentWidthChange?: (width: number) => void
}> = (props) => {
  const { dialog } = useSessionInfo()
  const container = useRef<HTMLDivElement>(null)
  const { shouldScroll, isCompact, onContentWidthChange } = props
  useEffect(() => {
    const dom = container.current
    if (!dom || !shouldScroll.current) return
    scrollToEnd(dom)
  }, [dialog, shouldScroll])

  // 监听内容区域宽度变化
  useEffect(() => {
    if (!onContentWidthChange || !container.current) return

    const updateWidth = () => {
      if (container.current) {
        onContentWidthChange(container.current.clientWidth)
      }
    }

    updateWidth()
    const resizeObserver = new ResizeObserver(updateWidth)
    resizeObserver.observe(container.current)

    return () => {
      resizeObserver.disconnect()
    }
  }, [onContentWidthChange])

  const renderQuestion = (text: string) => {
    return (
      <div key={text} className='mb-[6px] flex items-center'>
        <div
          className='min-h-[30px] flex items-center gap-[6px] 
        leading-[30px] text-sm  text-[#0C1F17] 
        rounded-[6px] px-[10px] py-[4px] bg-[#0000000D]
        cursor-pointer whitespace-normal break-words'
          onClick={() => {
            props?.onSend?.(text)
          }}
        >
          <span className='flex-1'>{text}</span>
          <ArrowRight className='flex-shrink-0' />
        </div>
      </div>
    )
  }

  return (
    <div
      className='project-ai-dialog flex-1 overflow-auto mt-7 px-8 flex flex-col gap-2'
      ref={container}
      onWheel={(e) => {
        if (e.deltaY < 0) {
          shouldScroll.current = false
        }
      }}
    >
      {dialog.map((item, i) => {
        return item.role === 'question' ? (
          <ProjectUserDialog value={item} key={i} isCompact={isCompact} />
        ) : (
          <ProjectAIDialog index={i} value={item} key={i} />
        )
      })}
      {dialog.length &&
      dialog[dialog.length - 1]?.role === 'answer' &&
      (dialog[dialog.length - 1] as AIDialog)?.sub_question ? (
        <div className={cn({ 'ml-[calc(36px+1em)]': !isCompact })}>
          {(dialog[dialog.length - 1] as AIDialog).sub_question?.map(
            renderQuestion,
          )}
        </div>
      ) : null}
    </div>
  )
}
