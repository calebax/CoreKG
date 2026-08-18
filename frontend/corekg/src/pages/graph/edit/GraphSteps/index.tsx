import { FC, Fragment } from 'react'
import { Steps } from 'antd'
import { cn } from '@/utils'
import Arrow from './arrow.svg?react'

export type GraphSteps = Style & {
  value: number
  onChange?: (val: number) => void
  steps: string[]
}
export const GraphSteps: FC<GraphSteps> = (props) => {
  const { value, onChange, steps, className, style } = props
  return (
    <div
      className={cn('flex items-center justify-between', className)}
      style={style}
    >
      {steps.map((s, i) => {
        return (
          <Fragment key={`${s}-${i}`}>
            <Steps
              className='w-auto'
              current={value}
              items={[{ title: s }]}
              initial={i}
              onChange={onChange}
            />
            {i !== steps.length - 1 ? <Arrow /> : null}
          </Fragment>
        )
      })}
    </div>
  )
}
