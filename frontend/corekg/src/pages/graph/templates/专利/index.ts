import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'
import avatar from './avatar.png'

const graphTags: GraphTag[] = [
  {
    tag_name: '专利',
    properties: [
      { name: '专利名称', type: 'string' },
      { name: '公开号', type: 'string' }, // Publication Number
      { name: '申请号', type: 'string' }, // Application Number
      { name: '优先权号', type: 'string' },
      { name: '法律状态', type: 'string' }, // 有效/失效/审中/授权/驳回/终止
      { name: '专利类型', type: 'string' }, // 发明/实用新型/外观设计
      { name: '技术领域', type: 'string' }, // 通信/生物/化工/软件/AI/半导体
      { name: '摘要', type: 'string' },
      { name: '申请日', type: 'string' },
      { name: '公开日', type: 'string' },
      { name: '授权日', type: 'string' },
      { name: '到期日', type: 'string' },
      { name: '申请国', type: 'string' }, // CN/US/EP/JP/KR/WIPO
      { name: 'PCT进入国家阶段', type: 'string' },
      { name: '页数', type: 'string' },
      { name: '被引次数', type: 'string' },
      { name: '族内成员数', type: 'string' },
      { name: '主权利要求项数', type: 'string' },
      { name: '权利要求总数', type: 'string' },
      { name: '引证数量', type: 'string' }, // backward citations
      { name: '被引证数量', type: 'string' }, // forward citations
      { name: 'IPC主分类号', type: 'string' },
      { name: 'CPC分类号', type: 'string' },
      { name: '同族ID', type: 'string' }, // Family ID
      { name: '说明书字数', type: 'string' },
    ],
  },
  {
    tag_name: '申请人或权利人',
    properties: [
      { name: '名称', type: 'string' },
      { name: '类型', type: 'string' }, // 企业/高校/科研院所/个人
      { name: '国家或地区', type: 'string' },
      { name: '地址', type: 'string' },
      { name: '行业', type: 'string' },
      { name: '统一社会信用代码', type: 'string' },
      { name: '是否上市公司', type: 'string' },
    ],
  },
  {
    tag_name: '发明人',
    properties: [
      { name: '姓名', type: 'string' },
      { name: '国籍', type: 'string' },
      { name: '机构', type: 'string' },
      { name: 'ORCID', type: 'string' },
      { name: '发明人序位', type: 'string' },
    ],
  },
  {
    tag_name: '代理机构',
    properties: [
      { name: '名称', type: 'string' },
      { name: '许可证号', type: 'string' },
      { name: '地址', type: 'string' },
      { name: '联系人', type: 'string' },
    ],
  },
  {
    tag_name: '代理人',
    properties: [
      { name: '姓名', type: 'string' },
      { name: '资格证号', type: 'string' },
      { name: '所属机构', type: 'string' },
      { name: '执业状态', type: 'string' },
    ],
  },
  {
    tag_name: '专利局或审查机构',
    properties: [
      { name: '名称', type: 'string' }, // CNIPA/USPTO/EPO/JPO/KIPO/WIPO
      { name: '国家或地区', type: 'string' },
      { name: '地址', type: 'string' },
    ],
  },
  {
    tag_name: '分类',
    properties: [
      { name: '体系', type: 'string' }, // IPC/CPC/LOC/US Class
      { name: '分类号', type: 'string' },
      { name: '标题', type: 'string' },
      { name: '说明', type: 'string' },
    ],
  },
  {
    tag_name: '引证',
    properties: [
      { name: '被引类型', type: 'string' }, // 专利/非专利文献(NPL)
      { name: '引证来源', type: 'string' }, // 审查员/申请人/第三方
      { name: '引证号或文献标识', type: 'string' },
      { name: '引证日期', type: 'string' },
    ],
  },
  {
    tag_name: '非专利文献',
    properties: [
      { name: '标题', type: 'string' },
      { name: '作者', type: 'string' },
      { name: '期刊或会议', type: 'string' },
      { name: '年份', type: 'string' },
      { name: 'DOI', type: 'string' },
      { name: 'URL', type: 'string' },
    ],
  },
  {
    tag_name: '法律事件',
    properties: [
      { name: '事件类型', type: 'string' }, // 审中/授权/缴费/无效/转让/质押/放弃/复审
      { name: '事件日期', type: 'string' },
      { name: '事件编号', type: 'string' },
      { name: '详情', type: 'string' },
    ],
  },
  {
    tag_name: '费用',
    properties: [
      { name: '费用类型', type: 'string' }, // 年费/申请费/审查费/复审费/转让登记费
      { name: '金额', type: 'string' },
      { name: '币种', type: 'string' },
      { name: '到期日', type: 'string' },
      { name: '是否已缴', type: 'string' },
    ],
  },
  {
    tag_name: '许可与转让',
    properties: [
      { name: '类型', type: 'string' }, // 独占/排他/普通/转让/质押
      { name: '协议编号', type: 'string' },
      { name: '生效日', type: 'string' },
      { name: '到期日', type: 'string' },
      { name: '对价金额', type: 'string' },
      { name: '币种', type: 'string' },
    ],
  },
  {
    tag_name: '诉讼与纠纷',
    properties: [
      { name: '案件编号', type: 'string' },
      { name: '法院或机构', type: 'string' },
      { name: '地区', type: 'string' },
      { name: '案由', type: 'string' }, // 侵权/无效/确认不侵权/行政复议
      { name: '立案日期', type: 'string' },
      { name: '判决结果', type: 'string' },
      { name: '赔偿金额', type: 'string' },
    ],
  },
  {
    tag_name: '技术主题',
    properties: [
      { name: '主题名称', type: 'string' }, // 关键词/主题词/主题聚类
      { name: '说明', type: 'string' },
    ],
  },
  {
    tag_name: '企业或机构',
    properties: [
      { name: '名称', type: 'string' },
      { name: '机构类型', type: 'string' }, // 企业/高校/科研院所/律所/投行/政府
      { name: '国家', type: 'string' },
      { name: '地址', type: 'string' },
      { name: '主营领域', type: 'string' },
    ],
  },
  {
    tag_name: '市场产品或标准',
    properties: [
      { name: '名称', type: 'string' }, // 产品名/标准号
      { name: '类型', type: 'string' }, // 产品/行业标准/国家标准/联盟规范
      { name: '版本或型号', type: 'string' },
      { name: '发布日期', type: 'string' },
    ],
  },
  {
    tag_name: '地理区域',
    properties: [
      { name: '名称', type: 'string' }, // 国家/地区/州/省/城市
      { name: 'ISO代码', type: 'string' },
    ],
  },
]
const graphTagRelationships: GraphTagRelationship[] = [
  // 专利核心关系
  { name: '由…申请', source: '专利', target: '申请人或权利人' }, // applied_by
  { name: '当前权利人', source: '专利', target: '申请人或权利人' }, // owned_by
  { name: '发明人包括', source: '专利', target: '发明人' }, // has_inventor
  { name: '代理于', source: '专利', target: '代理机构' }, // prosecuted_by_agency
  { name: '由…代理', source: '专利', target: '代理人' }, // prosecuted_by_agent
  { name: '受理于', source: '专利', target: '专利局或审查机构' }, // filed_at_office
  { name: '属于分类', source: '专利', target: '分类' }, // classified_as
  { name: '涉及技术主题', source: '专利', target: '技术主题' }, // relates_to_topic
  { name: '同族包含', source: '专利', target: '专利' }, // family_contains (同族内专利互链)

  // 引证关系
  { name: '引证', source: '专利', target: '引证' }, // cites_record
  { name: '被引证', source: '专利', target: '引证' }, // cited_by_record
  { name: '引用文献', source: '引证', target: '专利' }, // cites_patent
  { name: '引用非专利文献', source: '引证', target: '非专利文献' }, // cites_npl

  // 法律与费用
  { name: '发生法律事件', source: '专利', target: '法律事件' }, // has_legal_event
  { name: '产生费用', source: '专利', target: '费用' }, // has_fee
  { name: '因事件变更权利人', source: '法律事件', target: '申请人或权利人' }, // changes_owner_to

  // 许可与转让
  { name: '签订', source: '专利', target: '许可与转让' }, // has_license_or_assignment
  { name: '许可给', source: '许可与转让', target: '企业或机构' }, // licensed_to
  { name: '转让给', source: '许可与转让', target: '企业或机构' }, // assigned_to
  { name: '发生于', source: '许可与转让', target: '地理区域' }, // occurs_in_region

  // 诉讼与纠纷
  { name: '涉诉', source: '专利', target: '诉讼与纠纷' }, // involved_in_case
  { name: '原告', source: '诉讼与纠纷', target: '企业或机构' }, // plaintiff
  { name: '被告', source: '诉讼与纠纷', target: '企业或机构' }, // defendant
  { name: '裁判涉及', source: '诉讼与纠纷', target: '专利' }, // case_about

  // 市场与标准
  { name: '实施于', source: '专利', target: '市场产品或标准' }, // implemented_in
  { name: '符合', source: '专利', target: '市场产品或标准' }, // complies_with_standard

  // 地理关联
  { name: '申请地', source: '申请人或权利人', target: '地理区域' }, // located_in
  { name: '发明人所在', source: '发明人', target: '地理区域' }, // inventor_located_in
  { name: '审查地', source: '专利局或审查机构', target: '地理区域' }, // office_in_region
]
const Template: GraphTemplate = {
  name: '专利行业模板',
  description:
    '围绕“专利、发明人/机构、技术领域、法律状态”的基本关系，清晰展示专利信息。',
  avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
