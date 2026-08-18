import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Checkbox, Input, message, Steps } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import { APP_TEMPLATES, DEFAULT_CAPABILITIES, generateAppId } from './mock'
import type { AppTemplateType } from './types'
import styles from './create.module.scss'

const STEPS = [
  { title: '选择模板' },
  { title: '配置' },
  { title: '确认创建' },
]

const CAPABILITY_LABELS: Record<string, string> = {
  aiAssistant: 'AI 客服',
  search: '搜索',
  faq: 'FAQ',
  widget: 'Widget',
}

export default function ApplicationBuilderWizard() {
  const navigate = useNavigate()
  const [current, setCurrent] = useState(0)
  const [selectedType, setSelectedType] = useState<AppTemplateType | null>(null)
  const [configValues, setConfigValues] = useState<Record<string, string>>({})
  const [capabilities, setCapabilities] =
    useState<Record<string, boolean>>(DEFAULT_CAPABILITIES)
  const [creating, setCreating] = useState(false)

  const selectedTemplate = APP_TEMPLATES.find((t) => t.type === selectedType)

  const handleCreate = () => {
    setCreating(true)
    setTimeout(() => {
      setCreating(false)
      message.success('应用创建成功（仅演示）')
      const newId = generateAppId()
      navigate(`/apps/${newId}`)
    }, 1500)
  }

  const renderStep0 = () => (
    <div className={styles.templateGrid}>
      {APP_TEMPLATES.map((tpl) => (
        <div
          key={tpl.type}
          className={cn(styles.templateCard, {
            [styles.templateCardSelected]: selectedType === tpl.type,
          })}
          onClick={() => setSelectedType(tpl.type)}
        >
          <div className={styles.templateEmoji}>{tpl.emoji}</div>
          <div className={styles.templateName}>{tpl.name}</div>
          <div className={styles.templateDesc}>{tpl.description}</div>
        </div>
      ))}
    </div>
  )

  const renderStep1 = () => {
    if (!selectedTemplate) return null
    return (
      <div className={styles.configForm}>
        {selectedTemplate.configFields.map((field) => (
          <div key={field.key} className={styles.configField}>
            <label className={styles.configLabel}>{field.label}</label>
            <Input
              placeholder={`请输入${field.label}`}
              value={configValues[field.key] || ''}
              onChange={(e) =>
                setConfigValues((prev) => ({
                  ...prev,
                  [field.key]: e.target.value,
                }))
              }
            />
          </div>
        ))}
        <div className={styles.capabilities}>
          <label className={styles.configLabel}>启用能力</label>
          <div className={styles.capabilityList}>
            {Object.entries(CAPABILITY_LABELS).map(([key, label]) => (
              <Checkbox
                key={key}
                checked={capabilities[key]}
                onChange={(e) =>
                  setCapabilities((prev) => ({
                    ...prev,
                    [key]: e.target.checked,
                  }))
                }
              >
                {label}
              </Checkbox>
            ))}
          </div>
        </div>
      </div>
    )
  }

  const renderStep2 = () => {
    if (!selectedTemplate) return null
    return (
      <div className={styles.summary}>
        <div className={styles.summaryRow}>
          <span className={styles.summaryLabel}>模板</span>
          <span className={styles.summaryValue}>
            {selectedTemplate.emoji} {selectedTemplate.name}
          </span>
        </div>
        {selectedTemplate.configFields.map((field) => (
          <div key={field.key} className={styles.summaryRow}>
            <span className={styles.summaryLabel}>{field.label}</span>
            <span className={styles.summaryValue}>
              {configValues[field.key] || '(未填写)'}
            </span>
          </div>
        ))}
        <div className={styles.summaryRow}>
          <span className={styles.summaryLabel}>启用能力</span>
          <span className={styles.summaryValue}>
            {Object.entries(capabilities)
              .filter(([, v]) => v)
              .map(([k]) => CAPABILITY_LABELS[k])
              .join('、') || '无'}
          </span>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <Button
          type='text'
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/apps')}
        >
          返回
        </Button>
        <h1 className={styles.title}>创建应用</h1>
      </div>
      <div className={styles.wizard}>
        <Steps current={current} items={STEPS} className={styles.steps} />
        <div className={styles.stepContent}>
          {current === 0 && renderStep0()}
          {current === 1 && renderStep1()}
          {current === 2 && renderStep2()}
        </div>
        <div className={styles.footer}>
          {current > 0 && (
            <Button onClick={() => setCurrent((c) => c - 1)}>上一步</Button>
          )}
          {current < STEPS.length - 1 && (
            <Button
              type='primary'
              onClick={() => setCurrent((c) => c + 1)}
              disabled={current === 0 && !selectedType}
            >
              下一步
            </Button>
          )}
          {current === STEPS.length - 1 && (
            <Button
              type='primary'
              loading={creating}
              onClick={handleCreate}
            >
              开始创建
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
