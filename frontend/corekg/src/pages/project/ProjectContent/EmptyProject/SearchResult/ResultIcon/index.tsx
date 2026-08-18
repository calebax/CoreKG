import { FC } from 'react'
import { cn } from '@/utils'
import { ResultType } from '..'

const icons = new Map(
  Object.entries(
    import.meta.glob('./images/*.svg', {
      eager: true,
      import: 'default',
    }),
  ).map(([k, v]) => {
    const type = /^.*\/(.+)\.svg$/.exec(k)![1]
    return [type, v]
  }),
) as Map<ResultType | 'all', string>

export const ResultIcon: FC<Style & { type: ResultType | 'all' }> = (props) => {
  const { type, className, style } = props
  const src = icons.get(type)

  return (
    <img src={src} className={cn('w-6 h-6', className)} style={style}></img>
  )
}
