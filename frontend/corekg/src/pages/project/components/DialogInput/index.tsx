import { FC, PropsWithChildren, useState, useRef, useLayoutEffect } from 'react'
import { Input } from 'antd'
import { useControllableValue } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { Attachment } from '@/components/dialog/UserDialog'
import { AttachmentList } from '@/components/dialog/AttachmentList'
import { useProject } from '../..'
import styles from './styles.module.scss'

export type { Attachment }

export type DialogInput = Style &
  PropsWithChildren &
  ValueController<string> & {
    // 按下回车键的回调
    onEnter?: (val: string) => void
    placeholder?: string
    attachments?: Attachment[]
    onRemoveAttachment?: (index: number) => void
    maxLength?: number
    // 辅助提示文本（如润色中的loading提示）
    helperText?: string
  }

export const DialogInput: FC<DialogInput> = (props) => {
  const {
    onEnter,
    placeholder,
    children,
    className,
    style = {},
    attachments = [],
    onRemoveAttachment,
    maxLength = 500,
    helperText,
  } = props
  const [value, onChange] = useControllableValue<string>(props)
  const { isOtherPage, type } = useProject()
  const { t } = useTranslation('pages')
  const lineHeight = 24
  // 单文档问答页面默认3行，其他页面默认4行
  const defaultRows = type === 'single-file' ? 3 : 4
  const [inputHeight, setInputHeight] = useState(defaultRows * lineHeight)
  const textAreaRef = useRef<HTMLTextAreaElement>()
  const setSize = () => {
    const textarea = textAreaRef.current
    if (!textarea) return
    // 最多5行
    if (textarea.scrollHeight > 4 * lineHeight) {
      setInputHeight(5 * lineHeight)
    } else {
      setInputHeight(defaultRows * lineHeight)
    }
  }
  useLayoutEffect(setSize)

  return (
    <div
      className={cn(
        'bg-white p-3 rounded-xl',
        'overflow-auto transition-all duration-300 ease-in-out ',
        'relative flex flex-col',
        className,
      )}
      style={{
        ...style,
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
          onChange(e.target.value)
        }}
        // 与下方的TextArea同步设置样式
        className={styles.input}
        style={{
          height: inputHeight,
          maxHeight: inputHeight,
          lineHeight: `${lineHeight}px`,
        }}
        placeholder={
          placeholder ??
          (isOtherPage
            ? '基于“数据源&知识库”搜索或提问，shift+enter换行'
            : t('project.placeholder'))
        }
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            if (e.shiftKey || e.ctrlKey || e.altKey) {
              // shift/ctrl/alt+回车: 换行 组件默认行为
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
          className={styles.input}
          style={{
            lineHeight: `${lineHeight}px`,
          }}
        />
      </div>
      {/* 辅助提示文本 */}
      {helperText && (
        <div className='text-[#ABAFB2] text-xs font-medium mt-1'>
          {helperText}
        </div>
      )}
      <div className='w-full relative flex-none mt-2.5 h-[30px] flex justify-between items-center gap-1'>
        {children}
      </div>
    </div>
  )
}

export { default as ActiveBtn } from './images/activeBtn.svg?react'
export { default as DisabledBtn } from './images/disabledBtn.svg?react'
export { default as StopBtn } from './images/stopBtn.svg?react'
