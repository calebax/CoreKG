import { FC, PropsWithChildren, ReactNode } from 'react'
import { Input } from 'antd'
import { TextAreaProps } from 'antd/es/input'
import { cn } from '@/utils'
import styles from './styles.module.scss'

const maxLength = 150

type HomeInput = Omit<CtrlInput, 'onPressEnter' | 'prefix' | 'onSubmit'> & {
  prefix?: ReactNode
  className?: string
  onSubmit?: (val: string) => void
}
export const HomeInput: FC<PropsWithChildren<HomeInput>> = (props) => {
  const { prefix, children, onSubmit, className, value, ...rest } = props
  return (
    <div
      className={cn('rounded-xl w-[50vw]', styles.ctrlInputWrapper, className)}
    >
      <div
        className={cn(
          'w-full min-h-[120px] p-4 pb-2 bg-white flex flex-col gap-2 rounded-xl',
        )}
      >
        {prefix}
        <CtrlInput onPressEnter={onSubmit} value={value} {...rest} />
        <div className='h-10 flex items-center gap-2'>
          <div className='flex-1 overflow-hidden relative'>{children}</div>
          <span className='opacity-50'>
            {value?.length ?? 0}/{maxLength}
          </span>
        </div>
      </div>
    </div>
  )
}

type _CtrlInput = {
  value?: string
  onChange?: (val: string) => void
  onPressEnter?: (val: string) => void
}
type CtrlInput = _CtrlInput & Omit<TextAreaProps, keyof _CtrlInput>
/** input组件 enter不换行 ctrl+enter 换行 */
const CtrlInput: FC<CtrlInput> = (props) => {
  const { value, onChange, onPressEnter, className, ...rest } = props
  const isCtrlPressed = useRef(false)

  return (
    <Input.TextArea
      {...rest}
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
      maxLength={maxLength}
      className={cn(
        'flex-1 mb-6 text-base',
        'border-none shadow-none',
        className,
      )}
      style={{ resize: 'none' }}
      placeholder='请输入你的问题，按Enter发送，按Ctrl+Enter换行'
      onKeyDown={(e) => {
        if (e.key === 'Control') {
          isCtrlPressed.current = true
        }
      }}
      onKeyUp={(e) => {
        if (e.key === 'Control') {
          isCtrlPressed.current = false
        }
      }}
      onPressEnter={(e) => {
        const target = e.target as HTMLTextAreaElement
        // 只按住回车 搜索
        if (!isCtrlPressed.current) {
          e.preventDefault()
          onPressEnter?.(target.value)
          return
        }
        // 同时按住ctrl 换行
        const { selectionStart, selectionEnd, value: inputValue } = target
        const newText =
          inputValue.slice(0, selectionStart) +
          '\n' +
          inputValue.slice(selectionEnd)
        onChange?.(newText)
        setTimeout(() => {
          target.selectionStart = target.selectionEnd = selectionStart + 1
        }, 0)
        return
      }}
    ></Input.TextArea>
  )
}
