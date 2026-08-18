import { useState, useEffect, useRef, useCallback } from 'react'
import { Popover, Spin } from 'antd'
import { getProjectList } from '@/api/project'
import { getFileQaProject } from '@/api/knowledge'
import styles from './index.module.scss'

interface ProjectItem {
  id: number
  name: string
}

interface ProjectListPopoverProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (projectId: number) => void
  children: React.ReactNode
  excludeProjectId?: number | null
}

const PAGE_SIZE = 10

const ProjectListPopover = ({
  open,
  onOpenChange,
  onSelect,
  children,
  excludeProjectId,
}: ProjectListPopoverProps) => {
  const [displayedList, setDisplayedList] = useState<ProjectItem[]>([])
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [rawOffset, setRawOffset] = useState(0) // 记录原始数据的 offset
  const scrollContainerRef = useRef<HTMLDivElement>(null)

  const loadProjectList = useCallback(
    async (isLoadMore = false) => {
      if (isLoadMore) {
        if (loadingMore || loading || !hasMore) return
        setLoadingMore(true)
      } else {
        if (loading) return
        setLoading(true)
        setRawOffset(0) // 重置 offset
      }

      try {
        // 使用原始数据的 offset，而不是筛选后的数据长度
        const currentOffset = isLoadMore ? rawOffset : 0

        const res = await getProjectList({
          orderBy: ['updated_at desc'],
          limit: PAGE_SIZE,
          offset: currentOffset,
        })

        const rawList = (res && res.data) || []
        
        // 筛选逻辑：剔除当前会话分组和智能体问答分组
        const filtered: ProjectItem[] = rawList
          .filter((item: any) => {
            // 剔除当前会话分组（严格比较，确保类型一致）
            if (excludeProjectId != null && excludeProjectId !== undefined) {
              const itemId = Number(item.id)
              const excludeId = Number(excludeProjectId)
              if (!isNaN(itemId) && !isNaN(excludeId) && itemId === excludeId) {
                return false
              }
            }
            const projectType = String(item.project_type || '').toLowerCase().trim()
            if (projectType === 'forest_qa' || projectType === 'agent_qa') {
              return false
            }
            return true
          })
          .map((item: any) => ({
            id: item.id,
            name: item.name,
          }))
          .filter((i: ProjectItem) => i && typeof i.id === 'number')

        // 更新原始数据的 offset
        setRawOffset(currentOffset + rawList.length)

        if (isLoadMore) {
          setDisplayedList((prev) => [...prev, ...filtered])
        } else {
          setDisplayedList(filtered)
        }

        // 判断是否还有更多数据：基于 API 返回的原始数据长度
        // 如果原始数据长度等于 PAGE_SIZE，说明可能还有更多数据
        setHasMore(rawList.length === PAGE_SIZE)
      } catch (error) {
        console.log('获取会话分组列表失败:', error)
        if (!isLoadMore) {
          setDisplayedList([])
          setRawOffset(0)
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
    [loadingMore, loading, hasMore, excludeProjectId, rawOffset],
  )

  useEffect(() => {
    if (open) {
      loadProjectList(false)
    } else {
      // 关闭时重置状态
      setDisplayedList([])
      setHasMore(true)
      setRawOffset(0)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const target = e.currentTarget
      const { scrollTop, scrollHeight, clientHeight } = target

      // 滚动到底部附近时加载更多（距离底部50px时触发）
      const distanceToBottom = scrollHeight - scrollTop - clientHeight
      if (distanceToBottom < 50) {
        if (hasMore && !loadingMore && !loading) {
          loadProjectList(true)
        }
      }
    },
    [hasMore, loadingMore, loading, loadProjectList],
  )

  const handleProjectClick = (project: ProjectItem) => {
    onSelect(project.id)
    onOpenChange(false)
  }

  const renderContent = () => {
    if (loading) {
      return (
        <div className={styles.loadingContainer}>
          <Spin size='small' />
        </div>
      )
    }

    if (displayedList.length === 0) {
      return (
        <div className={styles.emptyContainer}>
          <span className={styles.emptyText}>还没有会话分组，去创建一个吧~</span>
        </div>
      )
    }

    return (
      <div
        ref={scrollContainerRef}
        className={styles.projectList}
        onScroll={handleScroll}
      >
        {displayedList.map((project) => (
          <div
            key={project.id}
            className={styles.projectItem}
            onClick={() => handleProjectClick(project)}
          >
            <span className={styles.projectName}>{project.name}</span>
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
      placement='rightTop'
      arrow={false}
      overlayClassName={styles.popoverOverlay}
      trigger={['click']}
      destroyTooltipOnHide
    >
      {children}
    </Popover>
  )
}

export default ProjectListPopover

