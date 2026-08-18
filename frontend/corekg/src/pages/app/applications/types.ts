export type AppTemplateType =
  | 'website'
  | 'product'
  | 'aftersales'
  | 'training'
  | 'policy'

export type AppStatus = 'online' | 'draft' | 'paused'

export type SyncStatus = 'success' | 'failed' | 'syncing'

export interface ApplicationStats {
  knowledgeCount: number
  faqCount: number
  syncStatus: SyncStatus
}

export interface Application {
  id: string
  name: string
  type: AppTemplateType
  status: AppStatus
  description: string
  icon?: string
  color: string
  stats: ApplicationStats
  lastSyncAt?: string
  lastPublishAt?: string
  updatedAt: string
  config: Record<string, unknown>
}

export interface AppTemplate {
  type: AppTemplateType
  name: string
  description: string
  emoji: string
  color: string
  configFields: ConfigField[]
}

export interface ConfigField {
  key: string
  label: string
  type: 'input' | 'toggle'
  defaultValue?: string | boolean
}

export interface AIStatus {
  model: string
  promptStatus: 'ok' | 'missing' | 'error'
  workflowStatus: 'ok' | 'missing' | 'error'
  embeddingModel: string
  rerankEnabled: boolean
  knowledgeCount: number
}

export type TabKey =
  | 'overview'
  | 'data'
  | 'ai'
  | 'publish'
  | 'analytics'
  | 'settings'

export const TAB_LABELS: Record<TabKey, string> = {
  overview: '概览',
  data: '数据',
  ai: 'AI',
  publish: '发布',
  analytics: '运营',
  settings: '设置',
}
