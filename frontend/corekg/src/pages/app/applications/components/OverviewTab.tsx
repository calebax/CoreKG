import { Button, Card } from 'antd'
import {
  SyncOutlined,
  ExperimentOutlined,
  CopyOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import type { Application } from '../types'
import styles from './OverviewTab.module.scss'

interface OverviewTabProps {
  app: Application
}

const STATUS_TEXT: Record<string, string> = {
  online: '在线',
  draft: '草稿',
  paused: '暂停',
}

export default function OverviewTab({ app }: OverviewTabProps) {
  return (
    <div className={styles.container}>
      <div className={styles.statusRow}>
        <div className={styles.statusCard}>
          <span className={styles.statusLabel}>项目状态</span>
          <span className={styles.statusValue}>{STATUS_TEXT[app.status]}</span>
        </div>
        <div className={styles.statusCard}>
          <span className={styles.statusLabel}>最近同步</span>
          <span className={styles.statusValue}>
            {app.lastSyncAt
              ? dayjs(app.lastSyncAt).format('MM-DD HH:mm')
              : '-'}
          </span>
        </div>
        <div className={styles.statusCard}>
          <span className={styles.statusLabel}>最近发布</span>
          <span className={styles.statusValue}>
            {app.lastPublishAt
              ? dayjs(app.lastPublishAt).format('MM-DD HH:mm')
              : '-'}
          </span>
        </div>
      </div>

      <div className={styles.statsGrid}>
        <Card className={styles.statCard} size='small'>
          <div className={styles.statValue}>{app.stats.knowledgeCount}</div>
          <div className={styles.statLabel}>知识</div>
        </Card>
        <Card className={styles.statCard} size='small'>
          <div className={styles.statValue}>{app.stats.faqCount}</div>
          <div className={styles.statLabel}>FAQ</div>
        </Card>
        <Card className={styles.statCard} size='small'>
          <div className={styles.statValue}>{app.stats.syncStatus === 'success' ? '成功' : app.stats.syncStatus === 'failed' ? '失败' : '同步中'}</div>
          <div className={styles.statLabel}>同步</div>
        </Card>
      </div>

      <div className={styles.actions}>
        <h3 className={styles.actionsTitle}>快捷操作</h3>
        <div className={styles.actionButtons}>
          <Button icon={<SyncOutlined />}>重新同步</Button>
          <Button icon={<ExperimentOutlined />}>测试</Button>
          <Button icon={<CopyOutlined />}>复制 Widget</Button>
          <Button icon={<LinkOutlined />}>查看网站</Button>
        </div>
      </div>
    </div>
  )
}
