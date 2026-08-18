import { FC, PropsWithChildren } from 'react'
import { Empty, Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { globalSearch } from '@/api/knowledge'
import { ResultContent } from './ResultContent'

export type ResultType =
  | 'forest'
  | 'doc'
  | 'table'
  /** 外部数据源 */
  | 'connect_app'
  | 'agent'
  | 'pic'
  | 'video'

export type SearchResult = Style & {
  search?: string
}
export const SearchResult: FC<SearchResult> = (props) => {
  const { className, style, search } = props
  const { data, loading } = useRequest(
    async () => {
      if (!search?.trim()) return null
      const {
        forest_search_result,
        doc_search_result,
        agent_search_result,
        video_search_result,
        image_search_result,
        external_search_result,
      } = (await globalSearch({ text: search! })) as Record<
        string,
        any | undefined
      >
      const _result: { type: ResultType; values: any[] }[] = []
      if (forest_search_result) {
        _result.push({
          type: 'forest',
          values: forest_search_result,
        })
      }
      if (doc_search_result) {
        _result.push({
          type: 'doc',
          values: doc_search_result,
        })
      }
      // 不同类型的外部数据源
      const {
        gmail_search,
        gmail_drive_search,
        confluence_search,
        slack_search,
      } = external_search_result ?? {}
      const externalData: { external_type: string; value: any }[] = []
      if (Array.isArray(gmail_search?.items)) {
        gmail_search.items.forEach((value: any) => {
          externalData.push({
            external_type: 'gmail',
            value,
          })
        })
      }

      if (Array.isArray(gmail_drive_search?.files)) {
        gmail_drive_search.files.forEach((value: any) => {
          externalData.push({
            external_type: 'google_drive',
            value,
          })
        })
      }

      if (Array.isArray(confluence_search?.results)) {
        confluence_search.results.forEach((value: any) => {
          externalData.push({
            external_type: 'confluence',
            value,
          })
        })
      }

      if (Array.isArray(slack_search?.files?.files)) {
        slack_search.files.files.forEach((value: any) => {
          externalData.push({
            external_type: 'slack',
            value,
          })
        })
      }

      if (externalData.length > 0) {
        _result.push({
          type: 'connect_app',
          values: externalData,
        })
      }

      if (agent_search_result) {
        _result.push({
          type: 'agent',
          values: agent_search_result,
        })
      }
      if (video_search_result) {
        _result.push({
          type: 'video',
          values: video_search_result,
        })
      }
      if (image_search_result) {
        _result.push({
          type: 'pic',
          values: image_search_result,
        })
      }
      // console.log(_result)

      return _result
    },
    {
      refreshDeps: [search],
      debounceWait: 1000,
      // ready: Boolean(search?.trim()),
    },
  )

  if (loading && Boolean(search?.trim())) {
    return (
      <SearchResultWrapper className={className} style={style}>
        <Skeleton className='p-4' active />
      </SearchResultWrapper>
    )
  }
  if (!data) return null
  if (!data.length) {
    return (
      <SearchResultWrapper className={className} style={style}>
        <Empty className='p-4' />
      </SearchResultWrapper>
    )
  }
  return (
    <SearchResultWrapper className={className} style={style}>
      <ResultContent result={data} />
    </SearchResultWrapper>
  )
}

const SearchResultWrapper: FC<PropsWithChildren & Style> = (props) => {
  const { children, className, style } = props
  return (
    <div
      className={cn('bg-white rounded-xl overflow-hidden', className)}
      style={style}
    >
      {children}
    </div>
  )
}
