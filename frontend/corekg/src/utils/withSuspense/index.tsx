import { ComponentType, forwardRef, ReactNode, Suspense } from 'react'
import { Skeleton } from 'antd'

/** 为一个react组件包裹Suspense */
export function withSuspense<T extends ComponentType<any>>(
  Comp: T,
  fallback?: ReactNode,
): T

export function withSuspense<T extends ComponentType<any>>(
  arg: ReactNode | T,
  fallback: ReactNode = <Skeleton active className='p-4' />,
) {
  const Comp: any = arg
  const CompWithSuspense = forwardRef(function CompWithSuspense(props, ref) {
    return (
      <Suspense fallback={fallback}>
        <Comp {...props} ref={ref} />
      </Suspense>
    )
  })
  return CompWithSuspense as unknown as T
}
