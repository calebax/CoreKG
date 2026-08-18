import { FC } from 'react'
import { Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { forestSearch } from '@/api/knowledge'
import { FileInSearchResult, ResultType, useSearchInputContext } from '..'
import { LinkItem } from './LinkItem'

export type LinkResProps = {
  hidden?: boolean
  value?: string
  /** 点击查看更多 会直接进行搜索 */
  startSearch: (type: ResultType) => void
}
/** 联想结果 */
export const LinkRes: FC<LinkResProps> = (props) => {
  const { hidden, value, startSearch } = props
  const { forest_id } = useSearchInputContext()
  const result = useRequest(
    async () => {
      if (!value || hidden || !forest_id) return null
      const res = await forestSearch({ forest_id, text: value })
      return res
    },
    {
      refreshDeps: [value, hidden, forest_id],
      debounceWait: 300,
    },
  )
  if (hidden) return null
  if (result.loading) {
    return (
      <div className='p-4'>
        <Skeleton active />
      </div>
    )
  }
  if (!result.data) return null
  const { doc_search_result, image_search_result, video_search_result } =
    result.data as Record<string, FileInSearchResult[]>
  return (
    <div className='flex flex-col max-h-[50vh] overflow-auto'>
      <LinkItem
        value={doc_search_result}
        type='doc'
        onMoreItem={() => startSearch('doc')}
      />
      <LinkItem
        value={image_search_result}
        type='image'
        onMoreItem={() => startSearch('image')}
      />
      <LinkItem
        value={video_search_result}
        type='video'
        onMoreItem={() => startSearch('video')}
      />
    </div>
  )
}
