import { forwardRef, HTMLAttributes, ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDownIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import KnowledgeIcon from './images/KnowledgeIcon.svg?react'

export type KnowledgeStatus = Style & {
  title?: ReactNode
  active?: boolean
  defaultTitle?: ReactNode
} & HTMLAttributes<HTMLDivElement>
/**
 * 展示选择知识的状态 可以配合popover弹出知识选择器等
 */
export const KnowledgeStatus = forwardRef<HTMLDivElement, KnowledgeStatus>(
  (props, ref) => {
    const { title, active, defaultTitle, className, style, ...rest } = props
    const { t: tC } = useTranslation('common')

    return (
      <div
        ref={ref}
        {...rest}
        className={cn(
          ' cursor-pointer border border-[#eff1f4] rounded-full bg-[#f7f7f7]',
          'text-[13px] text-[#6e757f]',
          'py-1 px-3 flex items-center gap-1 font-[500]',
          {
            'bg-[#FBE9FF] text-[#CC5DE8] font-[500] border-[#CC5DE833]': active,
          },
          className,
        )}
        style={style}
      >
        <KnowledgeIcon />
        {active ? title : defaultTitle || tC('model.selectKnowledge')}
        <ChevronDownIcon />
      </div>
    )
  },
)
