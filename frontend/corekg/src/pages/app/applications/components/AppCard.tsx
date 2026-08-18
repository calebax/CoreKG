import { Badge, Typography } from 'antd'
import dayjs from 'dayjs'
import { cn } from '@/utils'
import type { Application } from '../types'
import styles from './AppCard.module.scss'

interface AppCardProps {
  app: Application
  onClick: () => void
}

const STATUS_MAP: Record<string, { text: string; color: string }> = {
  online: { text: '在线', color: '#52c41a' },
  draft: { text: '草稿', color: '#faad14' },
  paused: { text: '暂停', color: '#ff4d4f' },
}

const SYNC_MAP: Record<string, { text: string; color: string }> = {
  success: { text: '成功', color: '#52c41a' },
  failed: { text: '失败', color: '#ff4d4f' },
  syncing: { text: '同步中', color: '#0C99FF' },
}

export default function AppCard({ app, onClick }: AppCardProps) {
  const status = STATUS_MAP[app.status] || STATUS_MAP.draft
  const sync = SYNC_MAP[app.stats.syncStatus] || SYNC_MAP.syncing

  return (
    <div className={styles.card} onClick={onClick}>
      <div className={styles.icon} style={{ backgroundColor: app.color }}>
        {app.name.charAt(0)}
      </div>
      <div className={styles.content}>
        <div className={styles.header}>
          <Typography.Paragraph
            className={styles.name}
            ellipsis={{ rows: 1 }}
          >
            {app.name}
          </Typography.Paragraph>
          <Badge
            status='default'
            text={status.text}
            className={cn(styles.badge)}
            color={status.color}
          />
        </div>
        <Typography.Paragraph
          className={styles.desc}
          type='secondary'
          ellipsis={{ rows: 1 }}
        >
          {app.description}
        </Typography.Paragraph>
        <div className={styles.stats}>
          <span>知识 {app.stats.knowledgeCount}</span>
          <span className={styles.divider}>|</span>
          <span>FAQ {app.stats.faqCount}</span>
          <span className={styles.divider}>|</span>
          <span style={{ color: sync.color }}>同步 {sync.text}</span>
        </div>
        <div className={styles.time}>
          更新于 {dayjs(app.updatedAt).format('MM-DD HH:mm')}
        </div>
      </div>
    </div>
  )
}
