import { FC, useState } from 'react'
import { Popover } from 'antd'
import { useControllableValue } from 'ahooks'
import { ChevronDownIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import EasyIcon from '@/assets/icons/home/home-qatype-easy.svg?react'
import FormalIcon from '@/assets/icons/home/home-qatype-formal.svg?react'
import NormalIcon from '@/assets/icons/home/home-qatype-normal.svg?react'
import ReasoningIcon from '@/assets/icons/home/home-qatype-reasoning.svg?react'
import StudyIcon from '@/assets/icons/home/home-qatype-study.svg?react'

export type PromptModeSelect = Style & {
  value?: string
  allowSelect?: boolean
  showArrow?: boolean
  onChange?: (val: string) => void
}

// 问答模式选项配置
const PROMPT_MODE_OPTIONS = [
  { key: 'normal', label: '普通模式', icon: NormalIcon },
  { key: 'concise', label: '简洁高效', icon: EasyIcon },
  { key: 'study', label: '学习研究', icon: StudyIcon },
  { key: 'explanation', label: '解释推理', icon: ReasoningIcon },
  { key: 'formal', label: '正式严谨', icon: FormalIcon },
] as const

export const PromptModeSelect: FC<PromptModeSelect> = (props) => {
  const { allowSelect = true, className, style, showArrow = true } = props
  const [value, onChange] = useControllableValue<string>(props, {
    defaultValue: 'normal',
  })
  const [open, setOpen] = useState(false)

  // 获取当前选中项的标签和图标
  const selectedOption =
    PROMPT_MODE_OPTIONS.find((item) => item.key === value) ||
    PROMPT_MODE_OPTIONS[0]
  const selectedLabel = selectedOption.label
  const SelectedIcon = selectedOption.icon

  const getPromptModeBtn = (
    props: {
      text?: string
      active?: boolean
      icon?: React.ComponentType<{ className?: string }>
    } = {},
  ) => {
    const { text = '普通模式', active, icon: Icon = SelectedIcon } = props
    return (
      <div
        className={cn(
          'cursor-pointer',
          'text-[13px] text-[#6e757f]',
          'py-1 px-3 flex items-center gap-1',
          {
            'text-[#CC5DE8]': active,
          },
          className,
        )}
        style={style}
      >
        <Icon className='w-[15.368px] h-[15.368px] text-current' />
        {text}
        {showArrow && <ChevronDownIcon />}
      </div>
    )
  }

  if (!allowSelect) {
    return getPromptModeBtn({ text: selectedLabel, active: true })
  }

  const popoverContent = (
    <div
      className={cn(
        'bg-white rounded-[10px] shadow-[0px_2px_12px_0px_rgba(0,0,0,0.1)]',
        'flex flex-col gap-[1.921px] p-[9.605px]',
        'min-w-[200px]',
      )}
    >
      {PROMPT_MODE_OPTIONS.map((option) => {
        const isSelected = value === option.key
        const OptionIcon = option.icon
        return (
          <div
            key={option.key}
            className={cn(
              'flex gap-[9.605px] h-[28.814px] items-center',
              'px-[9.605px] py-[8.644px] rounded-[3.842px]',
              'cursor-pointer transition-colors',
              {
                'bg-neutral-100': isSelected,
                'hover:bg-neutral-100': !isSelected,
              },
            )}
            onClick={() => {
              onChange?.(option.key)
              setOpen(false)
            }}
          >
            <div className='flex gap-[7.684px] items-center'>
              <OptionIcon className='w-[15.368px] h-[15.368px] shrink-0 text-current' />
              <p
                className={cn(
                  "font-['Inter:Medium','Noto_Sans_JP:Medium',sans-serif] font-medium leading-none",
                  'text-[13.447px] text-[#0c1f17]',
                )}
              >
                {option.label}
              </p>
            </div>
          </div>
        )
      })}
    </div>
  )

  return (
    <Popover
      trigger={['click']}
      open={open}
      onOpenChange={setOpen}
      placement='topLeft'
      arrow={false}
      content={popoverContent}
    >
      {getPromptModeBtn({ text: selectedLabel, active: !!value })}
    </Popover>
  )
}
