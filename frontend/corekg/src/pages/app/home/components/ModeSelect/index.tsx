import { FC } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import QuestionActiveIcon from '../../images/question-active.svg'
import QuestionIcon from '../../images/question.svg'
import SearchActiveIcon from '../../images/search-active.svg'
import SearchIcon from '../../images/search.svg'
import DataIcon from './images/data.svg'
import DataActiveIcon from './images/dataActive.svg'

/** input的三种模式 知识问答 知识库搜索 数据库问答 */
export type Mode = 'QA' | 'search' | 'data'
type ModeSelect = {
  className?: string
  value: Mode
  onChange: (val: Mode) => void
}
export const ModeSelect: FC<ModeSelect> = (props) => {
  const { value, onChange, className } = props
  const { t } = useTranslation('pages')
  const options: {
    value: Mode
    label: string
    normalIcon: string
    activeIcon: string
  }[] = [
    {
      value: 'QA',
      label: t('app.home.knowledgeBaseQa'),
      normalIcon: QuestionIcon,
      activeIcon: QuestionActiveIcon,
    },
    {
      value: 'search',
      label: t('app.home.knowledgeBaseSearch'),
      normalIcon: SearchIcon,
      activeIcon: SearchActiveIcon,
    },
    {
      value: 'data',
      label: t('app.home.tableAndKnowledgeBase'),
      normalIcon: DataIcon,
      activeIcon: DataActiveIcon,
    },
  ]

  return (
    <div className={cn('flex gap-2 ml-[15px]', className)}>
      {options.map((option) => {
        const { value: optionValue, label, normalIcon, activeIcon } = option
        const isActive = value === optionValue
        return (
          <div
            key={optionValue}
            className={cn(
              'flex items-center gap-1 px-2.5 py-[5px] rounded-[26px] border border-solid cursor-pointer transition-colors',
              'text-[14px] font-normal leading-[22px]',
              isActive
                ? 'border-[#a895fc] text-[#653ec4] bg-[#f9f7ff]'
                : 'border-[#e6e8f0] text-[#616373] hover:border-[#a895fc]',
            )}
            style={{ fontFamily: 'Inter , sans-serif' }}
            onClick={() => onChange(optionValue)}
          >
            <img
              src={isActive ? activeIcon : normalIcon}
              alt={label}
              className='w-4 h-4'
            />
            <span>{label}</span>
          </div>
        )
      })}
    </div>
  )
}
