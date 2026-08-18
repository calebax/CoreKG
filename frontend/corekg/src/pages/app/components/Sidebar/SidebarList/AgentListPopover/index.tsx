import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Popover, Spin } from 'antd'
import { getAgentDetail } from '@/api'
import { getLatestAgents } from '@/api/agent'
import { getDefaultAvatar } from '@/pages/app/agents/index/getDefaultAvatar'
import { getAgentUrl } from '@/pages/app/agents/utils/getAgentUrl'
import styles from './index.module.scss'

interface AgentItem {
  id: number
  name: string
  avatar?: string
}

interface AgentListPopoverProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  children: React.ReactNode
}

const PAGE_SIZE = 10

const AgentListPopover = ({
  open,
  onOpenChange,
  children,
}: AgentListPopoverProps) => {
  const navigate = useNavigate()
  const [displayedList, setDisplayedList] = useState<AgentItem[]>([])
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const scrollContainerRef = useRef<HTMLDivElement>(null)

  const loadAgentList = useCallback(
    async (isLoadMore = false) => {
      if (isLoadMore) {
        if (loadingMore || loading || !hasMore) return
        setLoadingMore(true)
      } else {
        if (loading) return
        setLoading(true)
      }

      try {
        // 直接使用当前列表长度计算 offset，避免异步状态导致重复请求
        const currentOffset = isLoadMore ? displayedList.length : 0

        const res = await getLatestAgents({
          limit: PAGE_SIZE,
          offset: currentOffset,
        })
        const rawList = (res && res.data) || []
        const normalized: AgentItem[] = rawList.map((item: any) => ({
          id: item.id,
          name: item.show_name,
          avatar: item.avatar_url,
        }))
        const filtered = normalized.filter((i) => i && typeof i.id === 'number')

        if (isLoadMore) {
          setDisplayedList((prev) => [...prev, ...filtered])
        } else {
          setDisplayedList(filtered)
        }

        // 如果返回的数据少于请求的数量，说明没有更多数据了
        setHasMore(filtered.length === PAGE_SIZE)
      } catch (error) {
        console.log('获取最近使用的智能体失败:', error)
        if (!isLoadMore) {
          setDisplayedList([])
        }
        setHasMore(false)
      } finally {
        if (isLoadMore) {
          setLoadingMore(false)
        } else {
          setLoading(false)
        }
      }
    },
    [displayedList.length, loadingMore, loading, hasMore],
  )

  useEffect(() => {
    if (open) {
      loadAgentList(false)
    } else {
      // 关闭时重置状态
      setDisplayedList([])
      setHasMore(true)
    }
  }, [open])

  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const target = e.currentTarget
      const { scrollTop, scrollHeight, clientHeight } = target

      // 滚动到底部附近时加载更多（距离底部50px时触发）
      if (scrollHeight - scrollTop - clientHeight < 50) {
        if (hasMore && !loadingMore && !loading) {
          loadAgentList(true)
        }
      }
    },
    [hasMore, loadingMore, loading, loadAgentList],
  )

  const handleAgentClick = async (agent: AgentItem) => {
    try {
      // 获取智能体详情以获取 type 信息
      const detail: any = await getAgentDetail(agent.id)
      const agentType = detail?.agent_type
      const url = getAgentUrl(agent.id, agentType)
      navigate(url)
      onOpenChange(false)
    } catch (error) {
      console.log('获取智能体详情失败:', error)
      // 如果获取详情失败，使用默认类型跳转
      const url = getAgentUrl(agent.id, 'role_play')
      navigate(url)
      onOpenChange(false)
    }
  }

  const getAgentAvatar = (agent: AgentItem) => {
    return agent.avatar === 'default'
      ? getDefaultAvatar('role_play')
      : (agent.avatar as string)
  }

  const renderContent = () => {
    if (loading) {
      return (
        <div className={styles.loadingContainer}>
          <Spin size='small' />
        </div>
      )
    }

    // 如果没有数据，显示提示信息
    if (displayedList.length === 0) {
      return (
        <div className={styles.emptyContainer}>
          <span className={styles.emptyText}>
            还没有使用记录，<Link to='/agents'>去体验一下吧~</Link>
          </span>
        </div>
      )
    }

    return (
      <div
        ref={scrollContainerRef}
        className={styles.agentList}
        onScroll={handleScroll}
      >
        {displayedList.map((agent) => (
          <div
            key={agent.id}
            className={styles.agentItem}
            onClick={() => handleAgentClick(agent)}
          >
            <div className={styles.agentAvatar}>
              <img
                src={getAgentAvatar(agent)}
                alt={agent.name}
                className={styles.avatarImage}
              />
            </div>
            <span className={styles.agentName}>{agent.name}</span>
          </div>
        ))}
        {loadingMore && (
          <div className={styles.loadingMoreContainer}>
            <Spin size='small' />
          </div>
        )}
      </div>
    )
  }

  return (
    <Popover
      content={renderContent()}
      open={open}
      onOpenChange={onOpenChange}
      placement='bottomLeft'
      arrow={false}
      overlayClassName={styles.popoverOverlay}
      trigger={['click']}
      destroyTooltipOnHide
    >
      {children}
    </Popover>
  )
}

export default AgentListPopover
