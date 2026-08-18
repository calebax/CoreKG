import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

export const graphTags: GraphTag[] = [
  {
    tag_name: '装备资产',
    properties: [
      { name: '资产编号', comment: 'asset_id', type: 'string' },
      { name: '资产名称', comment: 'asset_name', type: 'string' },
      { name: '资产类别', comment: 'asset_category', type: 'string' },
      { name: '型号/规格', comment: 'model_spec', type: 'string' },
      { name: '序列号', comment: 'serial_number', type: 'string' },
      { name: '投运日期', comment: 'commission_date', type: 'string' },
      { name: '生命周期阶段', comment: 'lifecycle_stage', type: 'string' },
      { name: '在役状态', comment: 'in_service', type: 'string' },
      { name: '资产价值', comment: 'asset_value', type: 'string' },
      { name: '所属组织', comment: 'owner_org', type: 'string' },
    ],
  },
  {
    tag_name: '制造商',
    properties: [
      { name: '厂商名称', comment: 'manufacturer_name', type: 'string' },
      { name: '统一社会信用代码', comment: 'reg_id', type: 'string' },
      { name: '国家/地区', comment: 'country_region', type: 'string' },
      { name: '联系邮箱', comment: 'email', type: 'string' },
      { name: '联系电话', comment: 'phone', type: 'string' },
    ],
  },
  {
    tag_name: '供应商',
    properties: [
      { name: '供应商名称', comment: 'supplier_name', type: 'string' },
      { name: '供应商编号', comment: 'supplier_id', type: 'string' },
      { name: '国家/地区', comment: 'country_region', type: 'string' },
      { name: '联系人', comment: 'contact_name', type: 'string' },
      { name: '联系电话', comment: 'phone', type: 'string' },
    ],
  },
  {
    tag_name: '零部件',
    properties: [
      { name: '部件名称', comment: 'part_name', type: 'string' },
      { name: '部件编号', comment: 'part_number', type: 'string' },
      { name: '版本/修订', comment: 'revision', type: 'string' },
      { name: '材质', comment: 'material', type: 'string' },
      { name: '单价', comment: 'unit_price', type: 'string' },
      { name: '可替代标记', comment: 'is_substitutable', type: 'string' },
    ],
  },
  {
    tag_name: '维护工单',
    properties: [
      { name: '工单编号', comment: 'work_order_id', type: 'string' },
      { name: '工单类型', comment: 'work_order_type', type: 'string' },
      { name: '工单状态', comment: 'work_order_status', type: 'string' },
      { name: '创建时间', comment: 'created_at', type: 'string' },
      { name: '完成时间', comment: 'completed_at', type: 'string' },
      { name: '工时', comment: 'labor_hours', type: 'string' },
      { name: '是否紧急', comment: 'is_urgent', type: 'string' },
    ],
  },
  {
    tag_name: '检验记录',
    properties: [
      { name: '检验编号', comment: 'inspection_id', type: 'string' },
      { name: '检验类型', comment: 'inspection_type', type: 'string' },
      { name: '检验日期', comment: 'inspection_date', type: 'string' },
      { name: '结论', comment: 'result', type: 'string' },
      { name: '是否合格', comment: 'is_passed', type: 'string' },
    ],
  },
  {
    tag_name: '认证证书',
    properties: [
      { name: '证书名称', comment: 'certificate_name', type: 'string' },
      { name: '证书编号', comment: 'certificate_id', type: 'string' },
      { name: '签发机构', comment: 'issuer', type: 'string' },
      { name: '生效日期', comment: 'effective_date', type: 'string' },
      { name: '到期日期', comment: 'expiry_date', type: 'string' },
      { name: '是否有效', comment: 'is_valid', type: 'string' },
    ],
  },
  {
    tag_name: '项目',
    properties: [
      { name: '项目名称', comment: 'project_name', type: 'string' },
      { name: '项目编号', comment: 'project_id', type: 'string' },
      { name: '项目阶段', comment: 'project_phase', type: 'string' },
      { name: '预算', comment: 'budget', type: 'string' },
      { name: '负责人', comment: 'owner', type: 'string' },
    ],
  },
  {
    tag_name: '地点',
    properties: [
      { name: '地点名称', comment: 'location_name', type: 'string' },
      { name: '国家/地区', comment: 'country_region', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
      { name: '经度', comment: 'longitude', type: 'string' },
      { name: '纬度', comment: 'latitude', type: 'string' },
    ],
  },
  {
    tag_name: '组织',
    properties: [
      { name: '组织名称', comment: 'org_name', type: 'string' },
      { name: '组织编码', comment: 'org_code', type: 'string' },
      { name: '类型', comment: 'org_type', type: 'string' },
      { name: '国家/地区', comment: 'country_region', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '由制造', source: '制造商', target: '装备资产' },
  { name: '由供应', source: '供应商', target: '零部件' },
  { name: '包含部件', source: '装备资产', target: '零部件' },
  { name: '更换为', source: '零部件', target: '零部件' },
  { name: '维护于', source: '维护工单', target: '装备资产' },
  { name: '检验自', source: '检验记录', target: '装备资产' },
  { name: '适用于', source: '认证证书', target: '装备资产' },
  { name: '开展于', source: '项目', target: '地点' },
  { name: '归属组织', source: '装备资产', target: '组织' },
  { name: '负责组织', source: '项目', target: '组织' },
  { name: '安装于', source: '装备资产', target: '地点' },
]
const Template: GraphTemplate = {
  name: '通用装备资产模板',
  description:
    '围绕“资产设备、位置/科室、巡检维护、使用人”的关系，清晰展示资产台账、生命周期与维保记录。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
