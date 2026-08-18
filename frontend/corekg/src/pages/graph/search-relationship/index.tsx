import { FC, ReactNode } from 'react'
import { Skeleton, Empty, Breadcrumb } from 'antd'
import { useGraphInfo, withGraphProvider } from '../GraphProvider'
import { SearchInner } from './SearchInner'

const SearchRelationship: FC = withGraphProvider(() => {
  const { data, loading } = useGraphInfo()
  const withWrapper = (children: ReactNode) => {
    return (
      <div className='w-full h-full p-4 bg-white flex flex-col'>{children}</div>
    )
  }
  if (loading) return withWrapper(<Skeleton active />)
  if (!data) return withWrapper(<Empty />)
  const { id: graph_id, name } = data
  return withWrapper(
    <>
      <Breadcrumb
        items={[
          {
            title: '知识图谱',
            href: '/graph',
          },
          { title: name, href: `/graph/detail?graphId=${graph_id}` },
          { title: '实体查找' },
        ]}
      />
      <SearchInner graph_id={graph_id} className='flex-1 mt-4' />
    </>,
  )
})

export default SearchRelationship
