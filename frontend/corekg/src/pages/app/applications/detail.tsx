import { useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Badge, Button, Tabs } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import AnalyticsTab from './components/AnalyticsTab'
import AITab from './components/AITab'
import DataTab from './components/DataTab'
import OverviewTab from './components/OverviewTab'
import PublishTab from './components/PublishTab'
import SettingsTab from './components/SettingsTab'
import { MOCK_APPLICATIONS } from './mock'
import type { TabKey } from './types'
import { TAB_LABELS } from './types'
import styles from './detail.module.scss'

const STATUS_MAP: Record<string, { text: string; color: string }> = {
  online: { text: '在线', color: '#52c41a' },
  draft: { text: '草稿', color: '#faad14' },
  paused: { text: '暂停', color: '#ff4d4f' },
}

const TAB_ITEMS: { key: TabKey; label: string }[] = [
  { key: 'overview', label: TAB_LABELS.overview },
  { key: 'data', label: TAB_LABELS.data },
  { key: 'ai', label: TAB_LABELS.ai },
  { key: 'publish', label: TAB_LABELS.publish },
  { key: 'analytics', label: TAB_LABELS.analytics },
  { key: 'settings', label: TAB_LABELS.settings },
]

export default function ApplicationDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const app = useMemo(
    () => MOCK_APPLICATIONS.find((a) => a.id === id),
    [id],
  )

  if (!app) {
    return (
      <div className={styles.notFound}>
        应用不存在
        <Button onClick={() => navigate('/apps')} style={{ marginTop: 12 }}>
          返回列表
        </Button>
      </div>
    )
  }

  const status = STATUS_MAP[app.status] || STATUS_MAP.draft

  const renderTab = (key: TabKey) => {
    switch (key) {
      case 'overview':
        return <OverviewTab app={app} />
      case 'data':
        return <DataTab />
      case 'ai':
        return <AITab />
      case 'publish':
        return <PublishTab />
      case 'analytics':
        return <AnalyticsTab />
      case 'settings':
        return <SettingsTab app={app} />
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <Button
          type='text'
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/apps')}
          className={styles.backBtn}
        />
        <div
          className={styles.icon}
          style={{ backgroundColor: app.color }}
        >
          {app.name.charAt(0)}
        </div>
        <div className={styles.headerInfo}>
          <h1 className={styles.name}>{app.name}</h1>
          <Badge
            status='default'
            text={status.text}
            color={status.color}
          />
        </div>
      </div>
      <Tabs
        defaultActiveKey='overview'
        items={TAB_ITEMS.map((tab) => ({
          key: tab.key,
          label: tab.label,
          children: renderTab(tab.key),
        }))}
        className={styles.tabs}
      />
    </div>
  )
}
