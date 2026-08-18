import { useState } from 'react'
import { Button, Card, message, Typography } from 'antd'
import {
  GlobalOutlined,
  CodeOutlined,
  ApiOutlined,
  QrcodeOutlined,
  CopyOutlined,
} from '@ant-design/icons'
import styles from './PublishTab.module.scss'

interface Channel {
  key: string
  name: string
  icon: React.ReactNode
  code: string
}

const CHANNELS: Channel[] = [
  {
    key: 'widget',
    name: 'Widget',
    icon: <GlobalOutlined />,
    code: '<script src="https://cdn.corekg.com/widget.js" data-app="app-001"></script>',
  },
  {
    key: 'js',
    name: 'JS SDK',
    icon: <CodeOutlined />,
    code: 'import { CoreKG } from "@corekg/sdk"\nconst app = new CoreKG({ appId: "app-001" })',
  },
  {
    key: 'iframe',
    name: 'Iframe',
    icon: <CodeOutlined />,
    code: '<iframe src="https://api.corekg.com/s/abc123" width="400" height="600"></iframe>',
  },
  {
    key: 'api',
    name: 'API',
    icon: <ApiOutlined />,
    code: 'POST https://api.corekg.com/v1/chat\nAuthorization: Bearer <your-api-key>',
  },
  {
    key: 'qrcode',
    name: '二维码',
    icon: <QrcodeOutlined />,
    code: 'https://api.corekg.com/s/abc123',
  },
]

export default function PublishTab() {
  const [copied, setCopied] = useState<string | null>(null)

  const handleCopy = async (key: string, code: string) => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(key)
      message.success('已复制到剪贴板')
      setTimeout(() => setCopied(null), 2000)
    } catch {
      message.error('复制失败')
    }
  }

  return (
    <div className={styles.container}>
      <h3 className={styles.title}>发布渠道</h3>
      <div className={styles.channels}>
        {CHANNELS.map((ch) => (
          <Card key={ch.key} className={styles.channel} size='small'>
            <div className={styles.channelHeader}>
              <span className={styles.channelIcon}>{ch.icon}</span>
              <span className={styles.channelName}>{ch.name}</span>
            </div>
            <Typography.Paragraph
              className={styles.code}
              copyable={false}
              ellipsis={{ rows: 3, expandable: true, symbol: '展开' }}
            >
              {ch.code}
            </Typography.Paragraph>
            <Button
              size='small'
              icon={<CopyOutlined />}
              type={copied === ch.key ? 'default' : 'primary'}
              ghost={copied !== ch.key}
              onClick={() => handleCopy(ch.key, ch.code)}
            >
              {copied === ch.key ? '已复制' : '复制'}
            </Button>
          </Card>
        ))}
      </div>
    </div>
  )
}
