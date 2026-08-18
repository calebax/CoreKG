import { FC, useMemo } from 'react'
import { Select, AutoComplete, Empty } from 'antd'
import { GraphTag } from 'Graph'
import { useControllableValue, useRequest } from 'ahooks'
import { cn } from '@/utils'
import { listGraphNode } from '@/api/graph'
import { useGraphInfo } from '../../GraphProvider'

export type Filter = {
  tag_name?: GraphTag['tag_name']
  search?: string
}
/** 筛选实体节点 */
export const NodeFilter: FC<
  ValueController<Filter> & {
    title?: string
    inputPlaceholder?: string
    onClear?: (val: Filter, type: 'tag_name' | 'search') => void
  } & Style
> = (props) => {
  const [value, onChange] = useControllableValue<Filter>(props)
  const {
    onClear,
    title = '实体类型',
    inputPlaceholder = '搜索实体',
    className,
    style,
  } = props
  const { tag_name, search } = value ?? {}

  const { data } = useGraphInfo()
  const { tags = [], id } = data ?? {}
  const options = useMemo(() => {
    return [
      { label: '全部', value: '' },
      ...tags.map((t) => {
        return {
          label: t.tag_name,
          value: t.tag_name,
        }
      }),
    ]
  }, [tags])

  const { data: nodeOptions } = useRequest(
    async () => {
      // 类型和名称至少有一个
      // tag_name 为空字符串表示"全部"，等同于 undefined
      const actualTagName = tag_name === '' ? undefined : tag_name
      if (!actualTagName && !search) return

      const { list } = await listGraphNode({
        graph_id: id!,
        graph_tag_id: actualTagName
          ? tags.find((t) => t.tag_name === actualTagName)?.tag_id
          : undefined,
        graph_node_name: search,
      })
      return (list as any[]).map((item) => {
        const name = item.graph_node_name as string
        return { label: name, value: name }
      })
    },
    { refreshDeps: [tag_name, search], debounceWait: 1000 },
  )
  return (
    <div className={cn('flex items-center gap-4', className)} style={style}>
      <span className='text-[#666666]'>{title}</span>
      <Select
        options={options}
        showSearch
        allowClear
        value={tag_name ?? ''}
        onChange={(v) => {
          // 空字符串表示"全部"，转换为 undefined
          const actualTagName = v === '' ? undefined : v
          onChange?.({
            tag_name: actualTagName,
            search,
          })
        }}
        onClear={() => {
          const newFilter = {
            search,
          }
          onChange?.(newFilter)
          onClear?.(newFilter, 'tag_name')
        }}
        className='w-40'
        placeholder='选择实体类型'
      />
      <AutoComplete
        value={search}
        onChange={(v) => {
          onChange?.({
            tag_name,
            search: v,
          })
        }}
        allowClear
        onClear={() => {
          const newFilter = {
            tag_name,
          }
          onChange?.(newFilter)
          onClear?.(newFilter, 'search')
        }}
        className='w-50'
        placeholder={inputPlaceholder}
        options={nodeOptions}
        notFoundContent={
          // 两个全为空时 不搜索 而不是没有找到
          // tag_name 为空字符串表示"全部"，等同于 undefined
          (tag_name === '' ? undefined : tag_name) || search ? (
            <Empty description='没有找到实体' />
          ) : null
        }
      />
    </div>
  )
}
