import { FC, useMemo } from 'react'
import { Checkbox, Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { loadAccountBindingList } from '@/api/accountBindings'

type SupportedProvider = {
  provider: string
  logo: string
}

type ExternalDataProviderSelectProps = {
  /** 已选中的 provider 列表 */
  value?: string[]
  /** 选中项变更回调 */
  onChange?: (providers: string[]) => void
}

/**
 * 外接数据模式 - 外部数据源选择器
 * 展示形式为无子节点的平铺列表，支持全选和单选
 */
export const ExternalDataProviderSelect: FC<
  Style & ExternalDataProviderSelectProps
> = (props) => {
  const { value = [], onChange, className, style } = props

  // 调用 account.Bindings 接口获取数据
  const { data, loading } = useRequest(async () => {
    const res = await loadAccountBindingList()
    return res
  })

  const supported: SupportedProvider[] = useMemo(
    () => data?.supported || [],
    [data],
  )

  // 判断是否全选
  const isSelectAll = useMemo(() => {
    if (!supported.length || !value.length) return false
    return supported.every((item) => value.includes(item.provider))
  }, [supported, value])

  // 全选/取消全选
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      onChange?.(supported.map((item) => item.provider))
    } else {
      onChange?.([])
    }
  }

  // 单个选择/取消
  const handleChange = (provider: string, checked: boolean) => {
    if (checked) {
      onChange?.([...value, provider])
    } else {
      onChange?.(value.filter((p) => p !== provider))
    }
  }

  if (loading) {
    return <Skeleton active className='p-4 w-40' />
  }

  if (!supported.length) {
    return (
      <div className='p-4 text-xs text-[#6E757F]'>暂无可用数据源</div>
    )
  }

  // 格式化 provider 名称（首字母大写）
  const formatProviderName = (provider: string) => {
    return provider
      .split('-')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ')
  }

  return (
    <div
      className={cn(
        'flex flex-col gap-[3px] p-2.5 pb-5 overflow-auto min-w-60',
        className,
      )}
      style={style}
    >
      {/* 头部：全选 + 已选计数 */}
      <div className='font-medium'>
        <div className='flex items-center gap-1 px-[5px] py-2 border-b border-solid border-[#EEEEEE]'>
          <Checkbox
            checked={isSelectAll}
            indeterminate={value.length > 0 && !isSelectAll}
            onChange={(e) => handleSelectAll(e.target.checked)}
          />
          <span>
            已选数据源（{value.length}/{supported.length}）
          </span>
        </div>
      </div>
      {/* 数据源列表 - 平铺无子节点结构 */}
      <div className='flex flex-col'>
        {supported.map((item) => (
          <div
            key={item.provider}
            className={cn(
              'flex items-center gap-2 px-[5px] py-2 rounded-md',
              'cursor-pointer hover:bg-[#F5F5F5] transition-colors',
            )}
            onClick={() => handleChange(item.provider, !value.includes(item.provider))}
            role='option'
            aria-selected={value.includes(item.provider)}
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                handleChange(item.provider, !value.includes(item.provider))
              }
            }}
          >
            <Checkbox checked={value.includes(item.provider)} />
            <img
              src={item.logo}
              alt={item.provider}
              className='w-5 h-5 rounded-sm object-contain'
            />
            <span className='text-sm text-[#333]'>
              {formatProviderName(item.provider)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
