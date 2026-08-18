import {
  FC,
  ReactNode,
  forwardRef,
  useRef,
  useState,
  useCallback,
  useEffect,
} from 'react'
import { Input } from 'antd'
import { TextAreaProps } from 'antd/es/input'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { useQaInputMaxLength } from '@/utils/useQaInputMaxLength'

export type CustomDialogInputProps = Omit<
  TextAreaProps,
  'onPressEnter' | 'prefix' | 'onSubmit' | 'onChange'
> & {
  prefix?: ReactNode
  onSubmit?: (val: string) => void
  mode?: 'QA' | 'search'
  children?: ReactNode
  onChange?: (val: string) => void
}

export const CustomDialogInput = forwardRef<any, CustomDialogInputProps>(
  (props, ref) => {
    const {
      prefix,
      children,
      onSubmit,
      maxLength,
      value,
      className,
      style,
      mode = 'QA',
      placeholder,
      onChange,
      ...rest
    } = props
    const { t } = useTranslation('pages')
    const { t: tC } = useTranslation('common')
    const qaInputMaxLength = useQaInputMaxLength()
    // 未显式传入时，问答类首页输入框使用统一的环境上限配置。
    const inputMaxLength = maxLength ?? qaInputMaxLength
    const textareaRef = useRef<any>(null)
    const [scrollTop, setScrollTop] = useState(0)
    const [textareaHeight, setTextareaHeight] = useState(72) // 默认3行高度

    const handleScroll = useCallback((e: any) => {
      const target = e.target
      setScrollTop(target.scrollTop)
    }, [])

    // 计算考虑textIndent的实际行数
    const calculateRows = useCallback((text: string, width: number) => {
      if (!text) return 3 // 最小3行

      const lines = text.split('\n')
      let totalRows = 0

      lines.forEach((line, index) => {
        if (line.length === 0) {
          totalRows += 1
          return
        }

        // 第一行需要考虑textIndent，减少可用宽度
        const availableWidth = index === 0 ? width - 55 : width
        const charsPerRow = Math.floor(availableWidth / 16) // 假设每个字符约16px
        const rowsForThisLine =
          Math.ceil((line.length * 16) / availableWidth) || 1
        totalRows += rowsForThisLine
      })

      return Math.max(3, Math.min(5, totalRows))
    }, [])

    // 动态计算高度
    useEffect(() => {
      if (textareaRef.current && value) {
        const textarea = textareaRef.current.resizableTextArea?.textArea
        if (textarea) {
          const style = window.getComputedStyle(textarea)
          const width = parseFloat(style.width)
          const rows = calculateRows(String(value), width)
          const newHeight = rows * 24 // 每行24px
          setTextareaHeight(newHeight)
        }
      } else {
        setTextareaHeight(72) // 默认3行
      }
    }, [value, calculateRows])

    return (
      <div className={cn('w-[50vw]', className)} style={style}>
        <div
          ref={ref}
          className='bg-[rgba(255,255,255,0.42)] relative rounded-[20px] w-full border border-[#e6e8f0] border-solid shadow-[0px_4px_10px_0px_rgba(0,0,0,0.1)] transition-all duration-300 ease-in-out min-h-[128px]'
        >
          <div className='relative w-full p-4 pb-14 pt-2.5'>
            {/* prefix区域 */}
            {prefix && <div className='mb-3'>{prefix}</div>}

            {/* 文本输入区域 */}
            <div className='relative overflow-hidden'>
              {' '}
              {/* 添加overflow-hidden来裁剪超出部分 */}
              <Input.TextArea
                {...rest}
                ref={textareaRef}
                value={value}
                onChange={(e) => onChange?.(e.target.value)}
                onScroll={handleScroll}
                maxLength={inputMaxLength}
                autoSize={false}
                bordered={false}
                className='w-full text-base bg-transparent resize-none'
                style={{
                  resize: 'none',
                  background: 'transparent',
                  fontFamily: 'Inter , sans-serif',
                  fontSize: '16px',
                  lineHeight: '24px',
                  border: 'none',
                  boxShadow: 'none',
                  outline: 'none',
                  padding: '0',
                  textIndent: '55px', // 恢复第一行缩进
                  height: `${textareaHeight}px`, // 使用计算出的高度
                  overflowY: textareaHeight >= 120 ? 'auto' : 'hidden', // 5行时显示滚动条
                }}
                placeholder={t('app.home.actionAskMessage', {
                  target: tC('button.search'),
                })}
                onPressEnter={(e) => {
                  const target = e.target as HTMLTextAreaElement
                  if (!e.ctrlKey && !e.shiftKey) {
                    e.preventDefault()
                    onSubmit?.(target.value)
                  }
                }}
              />
              {/* AI问答/AI搜索标签 - 跟随滚动一起移动 */}
              <div
                className='absolute left-0 top-0 pointer-events-none text-[#7445e0] text-[16px] font-medium z-10'
                style={{
                  fontFamily: 'Inter , sans-serif',
                  lineHeight: '24px',
                  transform: `translateY(-${scrollTop}px)`, // 跟随滚动移动
                }}
              >
                {mode === 'QA' ? t('app.home.aiQa') : t('app.home.aiQaService')}
              </div>
            </div>
          </div>

          {/* 底部控制区域 - 减少间距 */}
          <div className='absolute bottom-0 left-0 right-0 flex flex-row items-center justify-between px-4 py-3'>
            {children}
          </div>
        </div>
      </div>
    )
  },
)

CustomDialogInput.displayName = 'CustomDialogInput'
