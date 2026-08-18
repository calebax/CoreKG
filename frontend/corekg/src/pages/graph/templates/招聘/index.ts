import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '候选人',
    properties: [
      { name: '候选人ID', comment: 'candidate_id', type: 'string' },
      { name: '姓名', comment: 'name', type: 'string' },
      { name: '手机号', comment: 'phone', type: 'string' },
      { name: '邮箱', comment: 'email', type: 'string' },
      { name: '是否在职', comment: 'is_employed', type: 'string' },
    ],
  },
  {
    tag_name: '简历',
    properties: [
      { name: '简历ID', comment: 'resume_id', type: 'string' },
      { name: '期望职位', comment: 'expected_title', type: 'string' },
      { name: '期望薪资', comment: 'expected_salary', type: 'string' },
      { name: '工作年限', comment: 'experience_years', type: 'string' },
      { name: '更新时间', comment: 'updated_at', type: 'string' },
    ],
  },
  {
    tag_name: '职位',
    properties: [
      { name: '职位ID', comment: 'job_id', type: 'string' },
      { name: '职位名称', comment: 'job_title', type: 'string' },
      { name: '部门', comment: 'department', type: 'string' },
      { name: '工作地点', comment: 'location', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '招聘需求',
    properties: [
      { name: '需求ID', comment: 'requisition_id', type: 'string' },
      { name: '人数', comment: 'headcount', type: 'string' },
      { name: '预算', comment: 'budget', type: 'string' },
      { name: '优先级', comment: 'priority', type: 'string' },
    ],
  },
  {
    tag_name: '面试',
    properties: [
      { name: '面试ID', comment: 'interview_id', type: 'string' },
      { name: '面试时间', comment: 'interview_time', type: 'string' },
      { name: '面试轮次', comment: 'round', type: 'string' },
      { name: '结果', comment: 'result', type: 'string' },
    ],
  },
  {
    tag_name: 'Offer',
    properties: [
      { name: 'OfferID', comment: 'offer_id', type: 'string' },
      { name: '薪资', comment: 'salary', type: 'string' },
      { name: '入职日期', comment: 'onboard_date', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '员工',
    properties: [
      { name: '员工ID', comment: 'employee_id', type: 'string' },
      { name: '姓名', comment: 'name', type: 'string' },
      { name: '入职日期', comment: 'hire_date', type: 'string' },
      { name: '在职状态', comment: 'is_active', type: 'string' },
    ],
  },
  {
    tag_name: '部门',
    properties: [
      { name: '部门ID', comment: 'dept_id', type: 'string' },
      { name: '部门名称', comment: 'dept_name', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '提交', source: '候选人', target: '简历' },
  { name: '投递', source: '候选人', target: '职位' },
  { name: '对口需求', source: '职位', target: '招聘需求' },
  { name: '安排面试', source: '职位', target: '面试' },
  { name: '参加', source: '候选人', target: '面试' },
  { name: '发放', source: '职位', target: 'Offer' },
  { name: '接受', source: '候选人', target: 'Offer' },
  { name: '转为', source: '候选人', target: '员工' },
  { name: '隶属', source: '员工', target: '部门' },
  { name: '所属', source: '职位', target: '部门' },
]
const Template: GraphTemplate = {
  name: '招聘人力资源模板',
  description:
    '围绕“职位、候选人、面试流程、用人部门”的关系，清晰展示招聘需求、候选人进度与录用决策。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
