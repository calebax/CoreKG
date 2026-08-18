import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'
import avatar from './avatar.png'

const graphTags: GraphTag[] = [
  {
    tag_name: '药品',
    properties: [
      { name: '通用名', type: 'string' },
      { name: '商品名', type: 'string' },
      { name: '剂型', type: 'string' },
      { name: '规格', type: 'string' },
      { name: '批准文号', type: 'string' },
      { name: '生产厂家', type: 'string' },
      { name: '适应症', type: 'string' },
      { name: '用法用量', type: 'string' },
      { name: '不良反应概述', type: 'string' },
      { name: '禁忌', type: 'string' },
      { name: '药代动力学', type: 'string' },
      { name: '上市时间', type: 'string' },
    ],
  },
  {
    tag_name: '成分',
    properties: [
      { name: '分子式', type: 'string' },
      { name: '分子量', type: 'string' },
      { name: 'CAS号', type: 'string' },
      { name: '结构式', type: 'string' },
      { name: '作用机制', type: 'string' },
      { name: '主要靶点', type: 'string' },
    ],
  },
  {
    tag_name: '疾病',
    properties: [
      { name: '名称', type: 'string' },
      { name: 'ICD编码', type: 'string' },
      { name: '症状', type: 'string' },
      { name: '发病机制', type: 'string' },
      { name: '流行病学信息', type: 'string' },
      { name: '治疗方案', type: 'string' },
    ],
  },
  {
    tag_name: '靶点',
    properties: [
      { name: '名称', type: 'string' },
      { name: '类别', type: 'string' },
      { name: 'UniProt编号', type: 'string' },
      { name: '功能描述', type: 'string' },
    ],
  },
  {
    tag_name: '基因',
    properties: [
      { name: '基因名', type: 'string' },
      { name: '别名', type: 'string' },
      { name: '基因ID', type: 'string' },
      { name: '染色体位置', type: 'string' },
      { name: '功能', type: 'string' },
    ],
  },
  {
    tag_name: '临床试验',
    properties: [
      { name: '试验编号', type: 'string' },
      { name: '阶段', type: 'string' },
      { name: '适应症', type: 'string' },
      { name: '研究设计', type: 'string' },
      { name: '结果', type: 'string' },
      { name: '状态', type: 'string' },
    ],
  },
  {
    tag_name: '企业或机构',
    properties: [
      { name: '名称', type: 'string' },
      { name: '机构类型', type: 'string' },
      { name: '国家', type: 'string' },
      { name: '地址', type: 'string' },
      { name: '研发管线', type: 'string' },
    ],
  },
  {
    tag_name: '文献或专利',
    properties: [
      { name: '标题', type: 'string' },
      { name: '作者', type: 'string' },
      { name: '期刊或专利号', type: 'string' },
      { name: '发布时间', type: 'string' },
      { name: '摘要', type: 'string' },
    ],
  },
  {
    tag_name: '不良反应',
    properties: [
      { name: '事件名称', type: 'string' },
      { name: '严重程度', type: 'string' },
      { name: '发生率', type: 'string' },
      { name: '系统器官分类', type: 'string' },
    ],
  },
]
const graphTagRelationships: GraphTagRelationship[] = [
  // 药品相关
  { name: '包含', source: '药品', target: '成分' }, // contains
  { name: '治疗', source: '药品', target: '疾病' }, // treats
  { name: '可能导致', source: '药品', target: '不良反应' }, // may_cause
  { name: '验证于', source: '药品', target: '临床试验' }, // validated_in
  { name: '由…研发或生产', source: '药品', target: '企业或机构' }, // developed_or_manufactured_by
  { name: '报道于', source: '药品', target: '文献或专利' }, // reported_in

  // 成分相关
  { name: '被包含于', source: '成分', target: '药品' }, // is_contained_in
  { name: '作用于', source: '成分', target: '靶点' }, // acts_on
  { name: '受专利保护', source: '成分', target: '文献或专利' }, // protected_by_patent

  // 疾病相关
  { name: '被治疗于', source: '疾病', target: '药品' }, // is_treated_by
  { name: '关联靶点', source: '疾病', target: '靶点' }, // associated_with_target
  { name: '关联或致病基因', source: '疾病', target: '基因' }, // associated_or_causative_gene
  { name: '研究于', source: '疾病', target: '临床试验' }, // studied_in
  { name: '报道于', source: '疾病', target: '文献或专利' }, // reported_in

  // 靶点相关
  { name: '被作用于', source: '靶点', target: '成分' }, // is_acted_on_by
  { name: '关联疾病', source: '靶点', target: '疾病' }, // associated_with_disease
  { name: '由…编码', source: '靶点', target: '基因' }, // encoded_by

  // 基因相关
  { name: '编码', source: '基因', target: '靶点' }, // encodes
  { name: '与疾病相关或致病', source: '基因', target: '疾病' }, // susceptible_or_causative_for
  { name: '报道于', source: '基因', target: '文献或专利' }, // reported_in

  // 临床试验相关
  { name: '研究', source: '临床试验', target: '疾病' }, // investigates
  { name: '验证', source: '临床试验', target: '药品' }, // validates
  { name: '由…资助或开展', source: '临床试验', target: '企业或机构' }, // sponsored_by

  // 企业/机构相关
  { name: '研发或生产', source: '企业或机构', target: '药品' }, // develops_or_manufactures
  { name: '发起或开展', source: '企业或机构', target: '临床试验' }, // initiates_or_conducts
  { name: '报道或关联', source: '企业或机构', target: '文献或专利' }, // reports_or_associated_with

  // 文献/专利相关
  { name: '报道', source: '文献或专利', target: '药品' }, // reports
  { name: '报道', source: '文献或专利', target: '疾病' }, // reports
  { name: '报道', source: '文献或专利', target: '基因' }, // reports
  { name: '报道', source: '文献或专利', target: '靶点' }, // reports
  { name: '保护', source: '文献或专利', target: '成分' }, // protects
  { name: '保护', source: '文献或专利', target: '药品' }, // protects

  // 不良反应相关
  { name: '由…引起', source: '不良反应', target: '药品' }, // is_caused_by
  { name: '报道于', source: '不良反应', target: '临床试验' }, // reported_in
  { name: '报道于', source: '不良反应', target: '文献或专利' }, // reported_in
]
const Template: GraphTemplate = {
  name: '医药行业模板',
  description:
    '围绕“患者、疾病、症状、检验/检查、药物与治疗方案、医疗机构/医生”的基本关系，清晰呈现医疗信息',
  avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
