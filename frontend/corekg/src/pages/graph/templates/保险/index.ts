import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '客户',
    properties: [
      { name: '客户编号', comment: 'customer_id', type: 'string' },
      { name: '姓名/名称', comment: 'customer_name', type: 'string' },
      { name: '证件类型', comment: 'id_type', type: 'string' },
      { name: '证件号码', comment: 'id_number', type: 'string' },
      { name: '手机号', comment: 'phone', type: 'string' },
    ],
  },
  {
    tag_name: '保单',
    properties: [
      { name: '保单号', comment: 'policy_id', type: 'string' },
      { name: '生效日期', comment: 'effective_date', type: 'string' },
      { name: '到期日期', comment: 'expiry_date', type: 'string' },
      { name: '保费', comment: 'premium', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '险种',
    properties: [
      { name: '险种代码', comment: 'product_code', type: 'string' },
      { name: '险种名称', comment: 'product_name', type: 'string' },
      { name: '保障额度', comment: 'coverage_limit', type: 'string' },
      { name: '免赔额', comment: 'deductible', type: 'string' },
    ],
  },
  {
    tag_name: '受益人',
    properties: [
      { name: '受益人ID', comment: 'beneficiary_id', type: 'string' },
      { name: '姓名', comment: 'name', type: 'string' },
      { name: '关系', comment: 'relationship', type: 'string' },
      { name: '比例(%)', comment: 'percentage', type: 'string' },
    ],
  },
  {
    tag_name: '出险事件',
    properties: [
      { name: '事件ID', comment: 'claim_event_id', type: 'string' },
      { name: '发生时间', comment: 'event_time', type: 'string' },
      { name: '事件类型', comment: 'event_type', type: 'string' },
      { name: '描述', comment: 'description', type: 'string' },
    ],
  },
  {
    tag_name: '理赔',
    properties: [
      { name: '理赔编号', comment: 'claim_id', type: 'string' },
      { name: '申请日期', comment: 'apply_date', type: 'string' },
      { name: '理赔金额', comment: 'claim_amount', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '代理人',
    properties: [
      { name: '代理人ID', comment: 'agent_id', type: 'string' },
      { name: '姓名', comment: 'name', type: 'string' },
      { name: '等级', comment: 'level', type: 'string' },
      { name: '是否在职', comment: 'is_active', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '投保', source: '客户', target: '保单' },
  { name: '包含', source: '保单', target: '险种' },
  { name: '指定受益人', source: '保单', target: '受益人' },
  { name: '发生', source: '保单', target: '出险事件' },
  { name: '对应理赔', source: '出险事件', target: '理赔' },
  { name: '签单代理', source: '代理人', target: '保单' },
]

const Template: GraphTemplate = {
  name: '保险行业模板',
  description:
    '围绕“投保人/被保人、保单、险种、理赔案件”的关系，清晰展示保单全生命周期与理赔进展。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
