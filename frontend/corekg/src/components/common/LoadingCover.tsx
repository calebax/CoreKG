import { Spin } from 'antd'
import { cn } from '@/utils'

export default function LoadingCover({
  loading,
  className,
}: {
  loading: boolean
  className?: string
}) {
  if (!loading) return null
  return (
    <div
      className={cn(
        'absolute inset-0 flex items-center justify-center bg-white/50 z-10',
        className,
      )}
    >
      <Spin />
    </div>
  )
}
