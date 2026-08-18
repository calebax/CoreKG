import { FC } from 'react'
import { Button } from 'antd'
import { useRequest } from 'ahooks'
import {
  ArrowRightIcon,
  ChevronDownIcon,
  ChevronUpIcon,
} from 'tdesign-icons-react'
import { cn } from '@/utils'
import { getRecommendQuestions } from '@/api/knowledge'

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
    if (!recommend_questions) throw new Error('我正在为你准备问题建议，先问我点什么吧')
    return recommend_questions as string[]
  })
  if (error) return null

  return (
    <div
      className={cn(
        'pointer-events-none',
        'relative flex flex-col gap-3',
        className,
      )}
      style={style}
    >
      <Button
        className={cn(
          'pointer-events-auto self-start',
          'bg-[#f8f9fd] border-[#e6e8f0]',
        )}
        loading={loading}
        icon={hidden ? <ChevronDownIcon /> : <ChevronUpIcon />}
        iconPosition='end'
        onClick={() => setHidden?.(!hidden)}
      >
        推荐问题
      </Button>
      {hidden
        ? null
        : questions?.map((q, i) => {
            const isDisabled =
              questions?.length === 1 &&
              q === '我正在为你准备问题建议，先问我点什么吧'
            return (
              <Button
                iconPosition='end'
                key={i}
                icon={<ArrowRightIcon />}
                className={cn(
                  'pointer-events-auto hover:bg#F5F9FF-[#F5F9FF] self-start',
                  {
                    '!cursor-not-allowed opacity-60': isDisabled,
                  },
                )}
                onClick={() => {
                  if (!isDisabled) {
                    onSelectQusetion?.(q)
                  }
                }}
              >
                🔍{q}
              </Button>
            )
          })}
    </div>
  )
}
