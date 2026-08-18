import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '客户',
    properties: [
      { name: '客户编号', comment: 'customer_id', type: 'string' },
      { name: '客户名称', comment: 'customer_name', type: 'string' },
      { name: '证件类型', comment: 'id_type', type: 'string' },
      { name: '证件号码', comment: 'id_number', type: 'string' },
      { name: '手机号', comment: 'phone', type: 'string' },
      { name: '是否黑名单', comment: 'is_blacklisted', type: 'string' },
    ],
  },
  {
    tag_name: '账户',
    properties: [
      { name: '账户号', comment: 'account_number', type: 'string' },
      { name: '账户类型', comment: 'account_type', type: 'string' },
      { name: '币种', comment: 'currency', type: 'string' },
      { name: '余额', comment: 'balance', type: 'string', defaults: 0 },
      { name: '是否冻结', comment: 'is_frozen', type: 'string' },
    ],
  },
  {
    tag_name: '交易',
    properties: [
      { name: '交易编号', comment: 'transaction_id', type: 'string' },
      { name: '交易时间', comment: 'transaction_time', type: 'string' },
      { name: '交易类型', comment: 'transaction_type', type: 'string' },
      { name: '交易金额', comment: 'amount', type: 'string' },
      { name: '对手方账户', comment: 'counterparty_account', type: 'string' },
    ],
  },
  {
    tag_name: '贷款',
    properties: [
      { name: '贷款编号', comment: 'loan_id', type: 'string' },
      { name: '贷款产品', comment: 'loan_product', type: 'string' },
      { name: '贷款金额', comment: 'loan_amount', type: 'string' },
      { name: '年化利率', comment: 'annual_rate', type: 'string' },
      { name: '贷款期限(月)', comment: 'term_months', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '担保物',
    properties: [
      { name: '担保物编号', comment: 'collateral_id', type: 'string' },
      { name: '担保物类型', comment: 'collateral_type', type: 'string' },
      { name: '评估价值', comment: 'appraised_value', type: 'string' },
      { name: '所在地', comment: 'location', type: 'string' },
    ],
  },
  {
    tag_name: '产品',
    properties: [
      { name: '产品编号', comment: 'product_id', type: 'string' },
      { name: '产品名称', comment: 'product_name', type: 'string' },
      { name: '产品类型', comment: 'product_type', type: 'string' },
      { name: '基准利率', comment: 'base_rate', type: 'string' },
    ],
  },
  {
    tag_name: '网点',
    properties: [
      { name: '网点编号', comment: 'branch_id', type: 'string' },
      { name: '网点名称', comment: 'branch_name', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
      { name: '地址', comment: 'address', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '持有', source: '客户', target: '账户' },
  { name: '发生交易', source: '账户', target: '交易' },
  { name: '申请', source: '客户', target: '贷款' },
  { name: '使用产品', source: '贷款', target: '产品' },
  { name: '以担保', source: '贷款', target: '担保物' },
  { name: '开户于', source: '账户', target: '网点' },
]
const Template: GraphTemplate = {
  name: '金融银行信贷模板',
  description:
    '围绕“客户/企业、授信产品、额度与还款、风险评估”的关系，清晰展示授信审批、放贷与风控监测。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
