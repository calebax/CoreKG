import { FC, ReactNode } from 'react'
import { Mentions } from 'antd'
import { MentionsRef } from 'antd/es/mentions'
import { useControllableValue } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { useQaInputMaxLength } from '@/utils/useQaInputMaxLength'
import styles from './styles.module.scss'

export type DataInput = ValueController<string> & {
  // 输入at后会出现的选项
  atItems?: { label: ReactNode; value: string; key: number }[]
  // 选中一个选项
  onSelect?: (key: number) => void
  // 按下回车键的回调
  onEnter?: () => void
}
/**
 * 可以`@`数据知识库\
 * 具有一个前缀\
 * 无样式
 */
export const DataInput: FC<DataInput> = (props) => {
  const { atItems, onSelect, onEnter } = props
  const [value, onChange] = useControllableValue(props)
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const qaInputMaxLength = useQaInputMaxLength()
  const options = useMemo(() => {
    return atItems?.map((item) => {
      return {
        label: item.label,
        value: String(item.value),
      }
    })
  }, [atItems])

  const [inputHeight, setInputHeight] = useState(72)
  const [textAreaHeight, setTextAreaHeight] = useState(0)
  const textArea = useRef<MentionsRef>(null)
  const setSize = () => {
    if (!textArea.current) return
    const wrapper = textArea.current.nativeElement
    const textarea = wrapper.querySelector('textarea')
    if (!textarea) return
    const newHeight = textarea.scrollHeight
    if (Math.abs(textAreaHeight - newHeight) < 1) return
    setTextAreaHeight(newHeight)
    // 每行24px
    // 最少3行 最多5行
    if (newHeight > 72) {
      setInputHeight(5 * 24)
    } else {
      setInputHeight(72)
    }
  }
  useLayoutEffect(setSize)

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      if (e.shiftKey) {
        // shift+回车：换行（默认行为）
        return
      } else {
        // 回车：发送消息
        e.preventDefault()
        onEnter?.()
      }
    }
  }

  return (
    <div
      className={cn(
        'relative overflow-auto transition-all duration-300 ease-in-out ',
      )}
      style={{
        height: inputHeight,
      }}
    >
      <Prefix text={tC('button.analyse', { target: tC('resource.data') })} />
      <Mentions
        value={value}
        onChange={(s) => {
          onChange(s)
          setSize()
        }}
        options={options}
        onSelect={(option) => {
          const { key } = option
          onSelect?.(key as any)
          setSize()
        }}
        className={cn('text-base', styles.dataInput)}
        maxLength={qaInputMaxLength}
        style={{
          height: `max(${textAreaHeight}px,100% )`,
        }}
        // 没options不让@
        prefix={options ? '@' : '\u200B'}
        placeholder={t('app.home.actionAskMessage', {
          target: tC('button.search'),
        })}
        onKeyDown={handleKeyDown}
      />
      <div className='w-full h-0 overflow-hidden'>
        <Mentions
          ref={textArea}
          value={value}
          className={cn('text-base min-h-full', styles.dataInput)}
        />
      </div>
    </div>
  )
}

const Prefix = (props: { text: string }) => (
  <div
    className='absolute left-0 top-0 pointer-events-none text-[#7445e0] text-[16px] font-medium z-10'
    style={{
      fontFamily: 'Inter , sans-serif',
      lineHeight: '24px',
    }}
  >
    {props.text}
  </div>
)
