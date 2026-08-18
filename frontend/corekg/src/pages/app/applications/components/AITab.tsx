import { Card, Tag, Switch } from 'antd'
import { CheckCircleFilled, CloseCircleFilled } from '@ant-design/icons'
import { MOCK_AI_STATUS } from '../mock'
import styles from './AITab.module.scss'

const STATUS_ICON: Record<string, React.ReactNode> = {
  ok: <CheckCircleFilled style={{ color: '#52c41a' }} />,
  missing: <CloseCircleFilled style={{ color: '#faad14' }} />,
  error: <CloseCircleFilled style={{ color: '#ff4d4f' }} />,
}

const STATUS_TEXT: Record<string, string> = {
  ok: 'OK',
  missing: '未配置',
  error: '异常',
}

export default function AITab() {
  const ai = MOCK_AI_STATUS

  return (
    <div className={styles.container}>
      <h3 className={styles.title}>AI 状态</h3>
      <div className={styles.grid}>
        <Card className={styles.item} size='small'>
          <span className={styles.label}>模型</span>
          <span className={styles.value}>{ai.model}</span>
        </Card>
        <Card className={styles.item} size='small'>
          <span className={styles.label}>Prompt</span>
          <span className={styles.valueRow}>
            {STATUS_ICON[ai.promptStatus]}
            <Tag
              color={ai.promptStatus === 'ok' ? 'success' : ai.promptStatus === 'missing' ? 'warning' : 'error'}
            >
              {STATUS_TEXT[ai.promptStatus]}
            </Tag>
          </span>
        </Card>
        <Card className={styles.item} size='small'>
          <span className={styles.label}>Workflow</span>
          <span className={styles.valueRow}>
            {STATUS_ICON[ai.workflowStatus]}
            <Tag
              color={ai.workflowStatus === 'ok' ? 'success' : ai.workflowStatus === 'missing' ? 'warning' : 'error'}
            >
              {STATUS_TEXT[ai.workflowStatus]}
            </Tag>
          </span>
        </Card>
        <Card className={styles.item} size='small'>
          <span className={styles.label}>Embedding</span>
          <span className={styles.value}>{ai.embeddingModel}</span>
        </Card>
        <Card className={styles.item} size='small'>
          <span className={styles.label}>Rerank</span>
          <span className={styles.valueRow}>
            <Switch checked={ai.rerankEnabled} size='small' disabled />
          </span>
        </Card>
        <Card className={styles.item} size='small'>
          <span className={styles.label}>知识量</span>
          <span className={styles.value}>{ai.knowledgeCount}</span>
        </Card>
      </div>
    </div>
  )
}
