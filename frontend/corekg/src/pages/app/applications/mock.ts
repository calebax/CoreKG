import type {
  Application,
  AppTemplate,
  AIStatus,
} from './types'

export const APP_TEMPLATES: AppTemplate[] = [
  {
    type: 'website',
    name: '网站助手',
    description: '基于官方网站构建智能客服与搜索助手',
    emoji: '🌐',
    color: '#0C99FF',
    configFields: [
      { key: 'url', label: '网址', type: 'input', defaultValue: '' },
    ],
  },
  {
    type: 'product',
    name: '产品助手',
    description: '产品介绍、版本管理与技术文档智能问答',
    emoji: '📖',
    color: '#52c41a',
    configFields: [
      { key: 'productName', label: '产品名称', type: 'input', defaultValue: '' },
    ],
  },
  {
    type: 'aftersales',
    name: '售后助手',
    description: '售后维修、故障排查与客户服务智能支持',
    emoji: '🔧',
    color: '#fa8c16',
    configFields: [
      { key: 'serviceScope', label: '服务范围', type: 'input', defaultValue: '' },
    ],
  },
  {
    type: 'training',
    name: '培训助手',
    description: '培训课程、学习资料与考试题库智能管理',
    emoji: '📚',
    color: '#722ed1',
    configFields: [
      { key: 'department', label: '培训部门', type: 'input', defaultValue: '' },
    ],
  },
  {
    type: 'policy',
    name: '制度助手',
    description: '企业制度、流程规范与合规文档智能查询',
    emoji: '📑',
    color: '#eb2f96',
    configFields: [
      { key: 'orgName', label: '组织名称', type: 'input', defaultValue: '' },
    ],
  },
]

export const DEFAULT_CAPABILITIES = {
  aiAssistant: true,
  search: true,
  faq: true,
  widget: true,
}

export const MOCK_APPLICATIONS: Application[] = [
  {
    id: 'app-001',
    name: '官方网站助手',
    type: 'website',
    status: 'online',
    description: '公司官网智能客服与搜索助手',
    color: '#0C99FF',
    stats: { knowledgeCount: 1234, faqCount: 521, syncStatus: 'success' },
    lastSyncAt: '2026-07-17 14:30',
    lastPublishAt: '2026-07-18 09:00',
    updatedAt: '2026-07-18 09:00',
    config: { url: 'https://company.com', capabilities: DEFAULT_CAPABILITIES },
  },
  {
    id: 'app-002',
    name: '产品助手',
    type: 'product',
    status: 'online',
    description: '产品文档与版本管理智能问答',
    color: '#52c41a',
    stats: { knowledgeCount: 876, faqCount: 0, syncStatus: 'success' },
    lastSyncAt: '2026-07-16 20:00',
    updatedAt: '2026-07-17 10:15',
    config: { productName: 'CoreKG', capabilities: DEFAULT_CAPABILITIES },
  },
  {
    id: 'app-003',
    name: '售后助手',
    type: 'aftersales',
    status: 'draft',
    description: '售后维修与故障排查智能支持',
    color: '#fa8c16',
    stats: { knowledgeCount: 45, faqCount: 0, syncStatus: 'syncing' },
    updatedAt: '2026-07-18 08:30',
    config: { capabilities: DEFAULT_CAPABILITIES },
  },
  {
    id: 'app-004',
    name: '培训助手',
    type: 'training',
    status: 'online',
    description: '培训课程与学习资料智能管理',
    color: '#722ed1',
    stats: { knowledgeCount: 234, faqCount: 87, syncStatus: 'success' },
    lastSyncAt: '2026-07-18 06:00',
    updatedAt: '2026-07-18 06:00',
    config: { department: '技术中心', capabilities: DEFAULT_CAPABILITIES },
  },
  {
    id: 'app-005',
    name: '制度助手',
    type: 'policy',
    status: 'paused',
    description: '企业制度与合规文档查询',
    color: '#eb2f96',
    stats: { knowledgeCount: 123, faqCount: 34, syncStatus: 'failed' },
    lastSyncAt: '2026-07-10 12:00',
    updatedAt: '2026-07-12 16:45',
    config: { orgName: 'CoreKG', capabilities: DEFAULT_CAPABILITIES },
  },
]

export const MOCK_AI_STATUS: AIStatus = {
  model: 'DeepSeekV3',
  promptStatus: 'ok',
  workflowStatus: 'ok',
  embeddingModel: 'BGE-M3',
  rerankEnabled: true,
  knowledgeCount: 542,
}

let nextId = 6

export function generateAppId(): string {
  return `app-${String(nextId++).padStart(3, '0')}`
}
