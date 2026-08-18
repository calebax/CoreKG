import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

export const graphTags: GraphTag[] = [
  {
    tag_name: '物料',
    properties: [
      { name: '物料编码', comment: 'material_code', type: 'string' },
      { name: '物料名称', comment: 'material_name', type: 'string' },
      { name: '规格型号', comment: 'specification', type: 'string' },
      { name: '单位', comment: 'uom', type: 'string' },
      { name: '物料属性', comment: 'material_type', type: 'string' },
      { name: '标准成本', comment: 'standard_cost', type: 'string' },
    ],
  },
  {
    tag_name: 'BOM项',
    properties: [
      { name: 'BOM编号', comment: 'bom_id', type: 'string' },
      { name: '父物料编码', comment: 'parent_material_code', type: 'string' },
      { name: '子物料编码', comment: 'child_material_code', type: 'string' },
      { name: '用量', comment: 'quantity', type: 'string' },
      { name: '损耗率', comment: 'scrap_rate', type: 'string' },
      { name: '替代料', comment: 'substitute_material', type: 'string' },
    ],
  },
  {
    tag_name: '销售订单',
    properties: [
      { name: '订单号', comment: 'so_id', type: 'string' },
      { name: '客户编码', comment: 'customer_code', type: 'string' },
      { name: '交货日期', comment: 'delivery_date', type: 'string' },
    ],
  },
  {
    tag_name: '生产工单',
    properties: [
      { name: '工单号', comment: 'work_order_id', type: 'string' },
      { name: '来源单号', comment: 'source_order_id', type: 'string' },
      { name: '产品编码', comment: 'product_code', type: 'string' },
      { name: '计划数量', comment: 'planned_qty', type: 'string' },
      { name: '实际产出', comment: 'actual_qty', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
      { name: '优先级', comment: 'priority', type: 'string' },
      { name: '开工时间', comment: 'start_time', type: 'string' },
      { name: '完工时间', comment: 'end_time', type: 'string' },
    ],
  },
  {
    tag_name: '工艺路线',
    properties: [
      { name: '路线编号', comment: 'routing_id', type: 'string' },
      { name: '产品编码', comment: 'product_code', type: 'string' },
      { name: '版本', comment: 'version', type: 'string' },
      { name: '生效日期', comment: 'effective_date', type: 'string' },
    ],
  },
  {
    tag_name: '工序',
    properties: [
      { name: '工序编号', comment: 'operation_id', type: 'string' },
      { name: '工序名称', comment: 'operation_name', type: 'string' },
      { name: '工序顺序', comment: 'sequence', type: 'string' },
      { name: '标准工时(小时)', comment: 'std_hours', type: 'string' },
      { name: '是否关键工序', comment: 'is_critical', type: 'string' },
    ],
  },
  {
    tag_name: '工艺参数',
    properties: [
      { name: '参数编号', comment: 'param_id', type: 'string' },
      { name: '参数名称', comment: 'param_name', type: 'string' },
      { name: '设定值', comment: 'set_value', type: 'string' },
      { name: '实际值', comment: 'actual_value', type: 'string' },
      { name: '采集时间', comment: 'capture_time', type: 'string' },
    ],
  },
  {
    tag_name: '设备',
    properties: [
      { name: '设备编号', comment: 'equipment_id', type: 'string' },
      { name: '设备名称', comment: 'equipment_name', type: 'string' },
      { name: '设备类型', comment: 'equipment_type', type: 'string' },
      { name: '所在车间', comment: 'workshop', type: 'string' },
      { name: '可用状态', comment: 'is_available', type: 'string' },
    ],
  },
  {
    tag_name: '维保记录',
    properties: [
      { name: '记录单号', comment: 'maintenance_id', type: 'string' },
      { name: '维护类型', comment: 'maint_type', type: 'string' },
      { name: '维护日期', comment: 'maint_date', type: 'string' },
      { name: '维护人员', comment: 'technician', type: 'string' },
    ],
  },
  {
    tag_name: '人员',
    properties: [
      { name: '工号', comment: 'employee_id', type: 'string' },
      { name: '姓名', comment: 'name', type: 'string' },
      { name: '技能等级', comment: 'skill_level', type: 'string' },
      { name: '所属班组', comment: 'team', type: 'string' },
    ],
  },
  {
    tag_name: '质检单',
    properties: [
      { name: '质检单号', comment: 'qc_order_id', type: 'string' },
      { name: '关联单号', comment: 'ref_order_id', type: 'string' },
      { name: '质检类型', comment: 'qc_type', type: 'string' },
      { name: '结论', comment: 'result', type: 'string' },
      { name: '不合格数量', comment: 'defect_qty', type: 'string' },
      { name: '检验员', comment: 'inspector', type: 'string' },
    ],
  },
  {
    tag_name: '缺陷记录',
    properties: [
      { name: '缺陷代码', comment: 'defect_code', type: 'string' },
      { name: '缺陷描述', comment: 'defect_desc', type: 'string' },
      { name: '缺陷等级', comment: 'severity', type: 'string' },
    ],
  },
  {
    tag_name: '批次',
    properties: [
      { name: '批次号', comment: 'lot_no', type: 'string' },
      { name: '物料编码', comment: 'material_code', type: 'string' },
      { name: '生产日期', comment: 'mfg_date', type: 'string' },
      { name: '失效日期', comment: 'expire_date', type: 'string' },
      { name: '数量', comment: 'quantity', type: 'string' },
    ],
  },
  {
    tag_name: '仓库',
    properties: [
      { name: '仓库编号', comment: 'warehouse_id', type: 'string' },
      { name: '仓库名称', comment: 'warehouse_name', type: 'string' },
    ],
  },
  {
    tag_name: '供应商',
    properties: [
      { name: '供应商编码', comment: 'supplier_code', type: 'string' },
      { name: '供应商名称', comment: 'supplier_name', type: 'string' },
      { name: '信用等级', comment: 'credit_level', type: 'string' },
    ],
  },
  {
    tag_name: '客户',
    properties: [
      { name: '客户编码', comment: 'customer_code', type: 'string' },
      { name: '客户名称', comment: 'customer_name', type: 'string' },
    ],
  },
]

export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '下达', source: '客户', target: '销售订单' },
  { name: '生成', source: '销售订单', target: '生产工单' },
  { name: '组成', source: '物料', target: 'BOM项' },
  { name: '消耗', source: 'BOM项', target: '物料' },
  { name: '由路线', source: '生产工单', target: '工艺路线' },
  { name: '包含工序', source: '工艺路线', target: '工序' },
  { name: '生产产品', source: '生产工单', target: '物料' },
  { name: '使用设备', source: '工序', target: '设备' },
  { name: '执行人员', source: '工序', target: '人员' },
  { name: '采集参数', source: '工序', target: '工艺参数' },
  { name: '维护', source: '设备', target: '维保记录' },
  { name: '投料', source: '生产工单', target: '批次' },
  { name: '产出', source: '生产工单', target: '批次' },
  { name: '属于', source: '批次', target: '物料' },
  { name: '存储于', source: '批次', target: '仓库' },
  { name: '供应', source: '供应商', target: '批次' },
  { name: '质检', source: '质检单', target: '生产工单' },
  { name: '检验批次', source: '质检单', target: '批次' },
  { name: '包含缺陷', source: '质检单', target: '缺陷记录' },
]

const Template: GraphTemplate = {
  name: '生产制造执行模板',
  description:
    '扩展版MES模型：覆盖“人机料法环”五要素。包含销售源头、采购供应、生产执行、工艺参数采集、设备维保及全链路质量追溯。',
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
