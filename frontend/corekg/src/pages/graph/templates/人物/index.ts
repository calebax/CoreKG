import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '人物',
    properties: [
      { name: '人物ID', comment: 'person_id', type: 'string' },
      { name: '姓名', comment: 'name', type: 'string' },
      { name: '别名', comment: 'alias', type: 'string' },
      { name: '性别', comment: 'gender', type: 'string' },
      { name: '出生年份', comment: 'birth_year', type: 'string' },
    ],
  },
  {
    tag_name: '联系方式',
    properties: [
      { name: '联系方式ID', comment: 'contact_id', type: 'string' },
      { name: '类型', comment: 'type', type: 'string' }, // phone/email/wechat
      { name: '值', comment: 'value', type: 'string' },
      { name: '是否验证', comment: 'is_verified', type: 'string' },
    ],
  },
  {
    tag_name: '社交账号',
    properties: [
      { name: '账号ID', comment: 'account_id', type: 'string' },
      { name: '平台', comment: 'platform', type: 'string' }, // wechat/weibo/twitter etc.
      { name: '用户名', comment: 'username', type: 'string' },
      { name: '是否公开', comment: 'is_public', type: 'string' },
    ],
  },
  {
    tag_name: '地址',
    properties: [
      { name: '地址ID', comment: 'address_id', type: 'string' },
      { name: '国家', comment: 'country', type: 'string' },
      { name: '省州', comment: 'state_province', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
    ],
  },
  {
    tag_name: '工作单位',
    properties: [
      { name: '单位ID', comment: 'org_id', type: 'string' },
      { name: '单位名称', comment: 'org_name', type: 'string' },
      { name: '行业', comment: 'industry', type: 'string' },
    ],
  },
  {
    tag_name: '职位',
    properties: [
      { name: '职位ID', comment: 'position_id', type: 'string' },
      { name: '职务名称', comment: 'title', type: 'string' },
      { name: '开始日期', comment: 'start_date', type: 'string' },
      { name: '结束日期', comment: 'end_date', type: 'string' },
      { name: '是否在任', comment: 'is_current', type: 'string' },
    ],
  },
  {
    tag_name: '证件',
    properties: [
      { name: '证件ID', comment: 'document_id', type: 'string' },
      { name: '证件类型', comment: 'doc_type', type: 'string' }, // id_card/passport/student_id
      { name: '证件号码', comment: 'doc_number', type: 'string' },
      { name: '是否有效', comment: 'is_valid', type: 'string' },
    ],
  },
  {
    tag_name: '教育经历',
    properties: [
      { name: '经历ID', comment: 'edu_id', type: 'string' },
      { name: '学校名称', comment: 'school_name', type: 'string' },
      { name: '学历', comment: 'degree', type: 'string' },
      { name: '入学年份', comment: 'start_year', type: 'string' },
      { name: '毕业年份', comment: 'end_year', type: 'string' },
    ],
  },
  {
    tag_name: '关系线索',
    properties: [
      { name: '线索ID', comment: 'clue_id', type: 'string' },
      { name: '线索来源', comment: 'source', type: 'string' }, // tip/call/online
      { name: '可信度(0-1)', comment: 'confidence', type: 'string' },
      { name: '备注', comment: 'note', type: 'string' },
    ],
  },
  {
    tag_name: '事件',
    properties: [
      { name: '事件ID', comment: 'event_id', type: 'string' },
      { name: '事件类型', comment: 'event_type', type: 'string' }, // meeting/call/transaction
      { name: '发生时间', comment: 'occur_time', type: 'string' },
      { name: '地点', comment: 'location', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '使用联系方式', source: '人物', target: '联系方式' },
  { name: '使用社交', source: '人物', target: '社交账号' },
  { name: '居住于', source: '人物', target: '地址' },
  { name: '供职于', source: '人物', target: '工作单位' },
  { name: '担任', source: '人物', target: '职位' },
  { name: '持有', source: '人物', target: '证件' },
  { name: '就读', source: '人物', target: '教育经历' },
  { name: '关联线索', source: '人物', target: '关系线索' },
  { name: '共同参与', source: '人物', target: '事件' },
  // 人与人的直接关系（用关系边表达）
  { name: '家庭关系', source: '人物', target: '人物' }, // spouse/parent/child as edge attribute in your system
  { name: '同事', source: '人物', target: '人物' },
  { name: '同学', source: '人物', target: '人物' },
  { name: '社交关注', source: '人物', target: '人物' },
]
const Template: GraphTemplate = {
  name: '人物关系模板',
  description:
    '围绕“人物实体、亲属/社交/业务关联、事件时间线”的关系，清晰展示人物画像与关系网络。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
