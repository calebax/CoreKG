import { FC } from 'react'
import { Button } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import {
  ArrowRightIcon,
  ChevronDownIcon,
  ChevronUpIcon,
} from 'tdesign-icons-react'
import { cn } from '@/utils'
import { getRecommendQuestions } from '@/api/knowledge'
import AiLogoIcon from './images/aiLogo.svg?react'

export type RecommendQuestions = Style & {
  hidden?: boolean
  setHidden?: (expanded: boolean) => void
  file_id: number
  onSelectQusetion?: (question: string) => void
}
export const RecommendQuestions: FC<RecommendQuestions> = (props) => {
  const { hidden, setHidden, file_id, onSelectQusetion, className, style } =
    props
  const {
    loading,
    error,
    data: questions,
  } = useRequest(async () => {
    const { recommend_questions } = await getRecommendQuestions({ file_id })
    // 如果没有推荐问题，返回空数组而不是抛出错误
    return (recommend_questions as string[]) || []
  })
  
  // 提示文本
  const placeholderText = '我正在为你准备问题建议，先问我点什么吧'
  // 过滤掉占位文本，获取有效的推荐问题
  const validQuestions = questions?.filter((q) => q !== placeholderText) || []
  // 判断是否有有效的推荐问题：没有错误且存在有效问题
  const hasQuestions = !error && validQuestions.length > 0

  return (
    <div
      className={cn('relative flex flex-col gap-3', className)}
      style={style}
    >
      <div
        className={cn(
          '!cursor-pointer',
          'bg-[transparent] !shadow-none p-[5px] text-[#919497] flex gap-[5px] items-center',
        )}
        style={{ userSelect: 'none' }}
        onClick={() => setHidden?.(!hidden)}
      >
        {hidden ? <ChevronDownIcon /> : <ChevronUpIcon />}
        {props.hidden ? '展开' : '收起'}
      </div>
      <div className='flex flex-wrap gap-[5px]'>
        {hidden
          ? null
          : hasQuestions
            ? validQuestions.map((q) => {
                return (
                  <Button
                    key={q}
                    iconPosition='end'
                    className={cn(
                      '!text-[#0C1F17] !font-[500] text-[12px] pointer-events-auto hover:bg#F5F9FF-[#FFFFFF] rounded-[16px] self-start',
                    )}
                    onClick={() => {
                      onSelectQusetion?.(q)
                    }}
                  >
                    {q}
                    <AiLogoIcon />
                  </Button>
                )
              })
            : // 没有推荐问题时，显示纯文本提示
              <div className='flex items-center gap-[5px] text-[#0C1F17] font-[500] text-[12px] self-start ml-4'>
                <AiLogoIcon />
                <span>{placeholderText}</span>
              </div>}
      </div>
    </div>
  )
}
