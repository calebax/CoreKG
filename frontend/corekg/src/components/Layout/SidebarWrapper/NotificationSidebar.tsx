import React, { useRef, useState, useEffect } from 'react'
import { Badge, Spin, message } from 'antd'
import { useInfiniteScroll, useRequest } from 'ahooks'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'
import { cn } from '@/utils'
import {
  listMessages,
  readMessages,
  deleteMessages,
  type MessageItem,
} from '@/api/announcement'
import DeleteConfirmModal from '@/pages/app/docs/detail/components/DeleteConfirmModal'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'
import styles from './NotificationSidebar.module.scss'
import DeleteIcon from './images/delete.svg?react'
import DownIcon from './images/down.svg?react'
import LeftIcon from './images/left.svg?react'
import MessageIcon from './images/messge.svg?react'
import UpIcon from './images/up.svg?react'

dayjs.extend(relativeTime) // 时间格式化
dayjs.locale('zh-cn') // 时间格式化

// 通知侧边栏组件Props
interface NotificationSidebarProps {
  visible: boolean // 是否可见
  onClose: () => void // 关闭侧边栏
}

const PAGE_SIZE = 10 // 分页大小

// 通知侧边栏组件
export const NotificationSidebar: React.FC<NotificationSidebarProps> = ({
  visible,
  onClose,
}) => {
  const containerRef = useRef<HTMLDivElement>(null) // 容器引用
  const [activeId, setActiveId] = useState<string | null>(null) // 当前选中的卡片ID
  const [expandedIds, setExpandedIds] = useState<string[]>([]) // 展开的卡片ID列表
  const [deleteModalVisible, setDeleteModalVisible] = useState(false) // 删除模态框是否可见
  const [deleteTargetIds, setDeleteTargetIds] = useState<number[]>([]) // 删除目标ID列表
  // 删除模式标记：true 为批量删除，false 为单条删除
  const [isBatchDelete, setIsBatchDelete] = useState(false) // 删除模式标记
  // 获取消息通知计数的刷新方法
  const { messageNotificationCount } = useLoginGlobalData()
  // 获取通知消息列表的分页接口：分页滚动加载
  const { data, loading, loadingMore, reload, mutate } = useInfiniteScroll(
    async (d) => {
      // d.list 是已累积的消息列表，用其长度作为下一页的 offset
      const offset = d?.list?.length || 0
      const res = await listMessages({
        limit: PAGE_SIZE,
        offset,
        OrderBy: ['created_at desc'], // 按照创建时间排序
        Filters: [{ field: 'read_status', value: ['unread'] }],
      })
      return {
        list: res.data, // 列表数据
        total: res.total, // 总条数
        hasMore:
          res.data.length === PAGE_SIZE && offset + res.data.length < res.total, // 是否还有更多数据
      } as any
    },
    {
      target: containerRef, // 滚动容器
      isNoMore: (d) => !d?.hasMore || (d?.data?.length ?? 0) >= (d?.total ?? 0), // 是否还有更多数据
      reloadDeps: [visible], // 依赖更新
      manual: !visible, // 手动加载
    },
  )
  // 关闭侧边栏时重置内部选中与展开状态
  useEffect(() => {
    if (!visible) {
      setActiveId(null) // 清空当前选中的卡片
      setExpandedIds([]) // 清空展开的卡片ID列表
      // 关闭时清空列表数据，保证下次打开从头开始加载
      mutate(undefined) // 清空列表数据
    }
  }, [visible, mutate])

  // 标记为已读
  const { run: runRead } = useRequest(readMessages, {
    manual: true,
    onSuccess: () => {
      // 标记为已读成功后，刷新消息通知计数，更新外部红点
      messageNotificationCount.refresh()
    },
  }) // 标记为已读接口
  // 删除接口
  const { runAsync: runDelete } = useRequest(deleteMessages, {
    manual: true,
    onSuccess: () => {
      // 删除成功后，刷新消息通知计数，更新外部红点
      messageNotificationCount.refresh()
    },
  }) // 删除接口

  const handleCardClick = async (item: MessageItem) => {
    // 记录当前选中的卡片，控制选中样式
    setActiveId(item.id)

    // 未读消息点击后标记为已读
    if (item.read_status === 'unread') {
      // 更新本地已读状态，提升交互体验
      const newList = data?.list.map((i) =>
        i.id === item.id ? { ...i, read_status: 'read' } : i,
      )
      // 更新本地列表数据
      if (newList && data) {
        mutate({ ...data, list: newList as MessageItem[] }) // 更新本地列表数据
      }
      // 标记为已读
      runRead({ message_id: Number(item.id), status: 'read' }) // 标记为已读
    }
  }

  // 展开/收起
  const toggleExpand = (e: React.MouseEvent, id: string) => {
    e.stopPropagation() // 阻止事件冒泡
    setExpandedIds(
      (prev) =>
        prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id], // 展开/收起
    ) // 更新展开的卡片ID列表
  }

  // 发版公告的跳转
  const handleAnnouncementClick = (
    e: React.MouseEvent,
    item: MessageItem,
    path?: string,
  ) => {
    e.stopPropagation() // 阻止事件冒泡

    // 未读消息点击箭头跳转时也标记为已读
    if (item.read_status === 'unread') {
      // 更新本地已读状态，提升交互体验
      const newList = data?.list.map((i) =>
        i.id === item.id ? { ...i, read_status: 'read' } : i,
      )
      // 更新本地列表数据
      if (newList && data) {
        mutate({ ...data, list: newList as MessageItem[] }) // 更新本地列表数据
      }
      // 标记为已读
      runRead({ message_id: Number(item.id), status: 'read' }) // 标记为已读
    }

    if (path) {
      window.location.href = `${window.location.origin}${path}` // 跳转
    }
  }

  // 单条删除
  const handleDeleteClick = (e: React.MouseEvent, ids: string[]) => {
    e.stopPropagation() // 阻止事件冒泡
    // 单条删除模式
    setIsBatchDelete(false)
    // 将字符串ID转换为数字ID
    setDeleteTargetIds(ids.map((id) => Number(id)))
    setDeleteModalVisible(true)
  }

  // 顶部批量删除：当前列表全部删除
  const handleHeaderDeleteClick = () => {
    if (!data?.list?.length) return
    // 批量删除模式：清空当前列表
    setIsBatchDelete(true)
    // 将字符串ID转换为数字ID
    const ids = data.list.map((item) => Number(item.id))
    setDeleteTargetIds(ids) // 更新删除目标ID列表
    setDeleteModalVisible(true) // 显示删除模态框
  }

  // 确认删除
  const confirmDelete = async () => {
    // 批量删除时传递 deleteAll: true，单条删除时传递 message_ids
    const deleteParams = isBatchDelete
      ? { delete_all: true }
      : { message_ids: deleteTargetIds }
    await runDelete(deleteParams) // 删除接口
    setDeleteModalVisible(false) // 隐藏删除模态框
    setDeleteTargetIds([]) // 清空删除目标ID列表
    setActiveId(null) // 清空当前选中的卡片
    message.success('删除成功') // 删除成功提示
    reload() // 刷新列表数据
  }

  // 格式化时间
  const formatTime = (timeStr: string) => {
    const time = dayjs(timeStr) // 时间
    const diffMinutes = dayjs().diff(time, 'minute')
    if (diffMinutes < 1) {
      // 不到一分钟显示"刚刚"
      return '刚刚'
    }
    const diffHours = dayjs().diff(time, 'hour') // 小时
    if (diffHours < 1) {
      // 不到一小时显示"X分钟前"
      return time.fromNow().replace(' ', '')
    }
    return time.format('YYYY年MM月DD日 HH:mm') // 格式化时间
  }

  return (
    <>
      {/* 透明遮罩层：点击遮罩关闭侧边栏 */}
      {visible && <div className={styles.overlay} onClick={onClose} />}

      <div
        className={cn(styles.sidebarContainer, { [styles.visible]: visible })}
      >
        {/* 顶部标题与操作区域 */}
        <div className={styles.header}>
          <div className={styles.title}>通知</div>
          <div className={styles.actions}>
            {/* 顶部删除图标：点击直接触发批量删除二次确认 */}
            <div
              className={styles.iconWrapper}
              onClick={handleHeaderDeleteClick}
              title='批量删除'
            >
              <DeleteIcon className={styles.deleteIcon} />
            </div>
          </div>
        </div>

        {/* 通知列表内容区域 */}
        <div ref={containerRef} className={styles.content}>
          <div className={styles.messageList}>
            {data?.list.map((item) => (
              <div
                key={item.id}
                className={cn(styles.messageCard, {
                  [styles.selected]: activeId === item.id,
                })}
                onClick={() => handleCardClick(item)}
              >
                <div className={styles.cardHeader}>
                  <div className={styles.titleRow}>
                    {/* 警告图标 + 红点*/}
                    <Badge
                      dot={item.read_status === 'unread'}
                      className={styles.messageBadge}
                      offset={[-3, 3]}
                    >
                      <MessageIcon />
                    </Badge>
                    <span className={styles.title}>{item.title}</span>
                  </div>
                  {/* 单条删除图标：hover 或选中时展示 */}
                  <div
                    className={styles.deleteBtn}
                    onClick={(e) => handleDeleteClick(e, [item.id])}
                  >
                    <DeleteIcon className={styles.deleteIconSmall} />
                  </div>
                </div>

                {/* Content区域：包含内容和图标，图标与内容垂直居中 */}
                <div className={styles.contentWrapper}>
                  {/* 内容区域 */}
                  <div
                    className={cn(styles.cardContent, {
                      [styles.clamped]:
                        !item.route_path && !expandedIds.includes(item.id),
                    })}
                  >
                    {/* 内容 */}
                    {item.content}
                  </div>

                  {/* 温馨提示的展开/收起按钮 */}
                  {!item.route_path && (
                    <div
                      className={styles.expandToggle}
                      onClick={(e) => toggleExpand(e, item.id)}
                    >
                      {expandedIds.includes(item.id) ? (
                        <UpIcon className={styles.expandIcon} />
                      ) : (
                        <DownIcon className={styles.expandIcon} />
                      )}
                    </div>
                  )}

                  {/* 发版公告的跳转箭头按钮 */}
                  {item.route_path && (
                    <div
                      className={styles.announcementArrow}
                      onClick={(e) =>
                        handleAnnouncementClick(e, item, item.route_path)
                      }
                    >
                      <LeftIcon className={styles.arrowIcon} />
                    </div>
                  )}
                </div>

                {/* Footer：只显示时间 */}
                <div className={styles.cardFooter}>
                  <span>{formatTime(item.created_at)}</span>
                </div>
              </div>
            ))}
          </div>

          {/* 加载中提示 */}
          {(loading || loadingMore) && (
            <div className={styles.loading}>
              <Spin />
            </div>
          )}

          {/* 暂无通知提示 */}
          {!loading && data?.list.length === 0 && (
            <div className={styles.loading}>暂无通知</div>
          )}
        </div>
      </div>

      {/* 删除确认模态框 */}
      <DeleteConfirmModal
        visible={deleteModalVisible}
        onCancel={() => setDeleteModalVisible(false)}
        onConfirm={confirmDelete}
        customTitle='删除通知'
        customText={isBatchDelete ? '是否清空全部通知？' : '要清除这条通知吗？'}
      />
    </>
  )
}
