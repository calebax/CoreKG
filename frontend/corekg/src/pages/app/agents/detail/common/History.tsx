import {
  useState,
  forwardRef,
  useImperativeHandle,
  useRef,
  useEffect,
  useCallback,
} from 'react'
import { useParams } from 'react-router-dom'
import { Empty, Skeleton } from 'antd'
import { useMount, useUpdateEffect } from 'ahooks'
import { getSessionHistory } from '@/api'
import { cn } from '@/utils'
import HistoryItem from './HistoryItem'

export interface HistoryRef {
  refresh: () => Promise<void>
}

interface SessionInput {
  [key: string]: unknown
}

interface ApiSessionItem {
  ID: string
  name: string
  resource_type: string
  input?: SessionInput
  is_top?: boolean
}

interface SessionItem {
  id: string
  name: string
  resourceType: string
  input?: SessionInput
  is_top?: boolean
}

const History = forwardRef<
  HistoryRef,
  {
    sendLoading: boolean
    newSessionId: string
    currentSessionId: string
    typePath: boolean
    onClickItem: () => void
  }
>(
  (
    { sendLoading, newSessionId, currentSessionId, typePath, onClickItem },
    ref,
  ) => {
    const { id: agentId } = useParams()
    const [sessionList, setSessionList] = useState<SessionItem[]>([])
    const [loading, setLoading] = useState(false)
    const [loadingMore, setLoadingMore] = useState(false)
    const [hasMore, setHasMore] = useState(true)
    const scrollContainerRef = useRef<HTMLDivElement>(null)

    const LIMIT = 30

    const getData = useCallback(
      async (isLoadMore = false) => {
        if (isLoadMore) {
          setLoadingMore(true)
        } else {
          setLoading(true)
        }

        // 使用当前列表长度作为offset，避免删除后的数据跳跃问题
        const currentOffset = isLoadMore ? sessionList.length : 0

        const res = await getSessionHistory({
          limit: LIMIT,
          offset: currentOffset,
          filters: [
            { field: 'resource_id', value: [agentId] },
            { field: 'resource_type', value: ['agent'] },
          ],
        })

        let list = res.Data || []
        const total = res.total || 0

        list = list.map((item: ApiSessionItem) => {
          return {
            id: item.ID,
            name: item.name,
            resourceType: item.resource_type,
            input: item.input,
            is_top: item.is_top,
          }
        })

        if (isLoadMore) {
          setSessionList((prev) => [...prev, ...list])
        } else {
          setSessionList(list)
        }

        // 使用实际的总数来判断是否还有更多数据
        const currentListLength = isLoadMore
          ? sessionList.length + list.length
          : list.length
        setHasMore(currentListLength < total)

        if (isLoadMore) {
          setLoadingMore(false)
        } else {
          setLoading(false)
        }
      },
      [agentId, sessionList.length],
    )

    const loadMore = useCallback(async () => {
      if (!hasMore || loadingMore) return
      await getData(true)
    }, [hasMore, loadingMore, getData])

    const handleScroll = useCallback(() => {
      const container = scrollContainerRef.current
      if (!container || !hasMore || loadingMore) return

      const { scrollTop, scrollHeight, clientHeight } = container
      // 当滚动到距离底部50px时触发加载更多
      if (scrollTop + clientHeight >= scrollHeight - 50) {
        loadMore()
      }
    }, [hasMore, loadingMore, loadMore])

    useMount(() => {
      getData()
    })

    useUpdateEffect(() => {
      if (newSessionId) {
        setSessionList((prev) => {
          return [
            {
              id: newSessionId,
              name: '新建会话',
              resourceType: 'agent',
              is_top: false,
            },
            ...prev,
          ]
        })
      }
    }, [newSessionId])

    // 滚动到底部加载更多
    useEffect(() => {
      const container = scrollContainerRef.current
      if (container) {
        container.addEventListener('scroll', handleScroll)
        return () => container.removeEventListener('scroll', handleScroll)
      }
    }, [handleScroll])

    const refresh = async () => {
      setHasMore(true)
      await getData()
    }

    useImperativeHandle(ref, () => ({
      refresh: refresh,
    }))

    return (
      <div
        className={cn(
          'flex-grow w-full overflow-hidden relative',
          sendLoading && 'pointer-events-none opacity-80',
        )}
      >
        <div
          ref={scrollContainerRef}
          className='w-full h-full overflow-y-auto flex flex-col gap-2 p-4'
        >
          {sessionList.map((item) => {
            return (
              <HistoryItem
                key={item.id}
                agentId={agentId || ''}
                typePath={typePath}
                item={item}
                currentSessionId={currentSessionId}
                onClick={() => onClickItem()}
                setSessionList={setSessionList}
                refresh={refresh}
              />
            )
          })}

          {/* 底部加载指示器 */}
          {sessionList.length > 0 && loadingMore && (
            <div className='flex justify-center py-4 text-gray-500 text-sm'>
              加载中...
            </div>
          )}

          {sessionList.length > LIMIT && !hasMore && (
            <div className='flex justify-center py-4 text-gray-400 text-sm'>
              没有更多数据了
            </div>
          )}

          {sessionList.length === 0 && !loading && <Empty />}
        </div>

        {loading && (
          <div className='absolute inset-0 p-4'>
            <Skeleton
              className='agent-history-skeleton'
              active
              title={false}
              loading={true}
              paragraph={{ rows: 20 }}
            />
          </div>
        )}
      </div>
    )
  },
)

export default History
