import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '供应商',
    properties: [
      { name: '供应商编号', comment: 'supplier_id', type: 'string' },
      { name: '供应商名称', comment: 'supplier_name', type: 'string' },
      { name: '国家/地区', comment: 'country_region', type: 'string' },
      { name: '评级', comment: 'rating', type: 'string' },
    ],
  },
  {
    tag_name: '采购单',
    properties: [
      { name: '采购单号', comment: 'po_id', type: 'string' },
      { name: '下单日期', comment: 'order_date', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
      { name: '总金额', comment: 'total_amount', type: 'string' },
    ],
  },
  {
    tag_name: '收货单',
    properties: [
      { name: '收货单号', comment: 'grn_id', type: 'string' },
      { name: '收货日期', comment: 'receive_date', type: 'string' },
      { name: '收货数量', comment: 'received_qty', type: 'string' },
    ],
  },
  {
    tag_name: '仓库',
    properties: [
      { name: '仓库编号', comment: 'warehouse_id', type: 'string' },
      { name: '仓库名称', comment: 'warehouse_name', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
    ],
  },
  {
    tag_name: '库位',
    properties: [
      { name: '库位编码', comment: 'bin_code', type: 'string' },
      { name: '区域', comment: 'zone', type: 'string' },
      { name: '容量', comment: 'capacity', type: 'string' },
    ],
  },
  {
    tag_name: '库存',
    properties: [
      { name: 'SKU', comment: 'sku_id', type: 'string' },
      { name: '批次号', comment: 'batch_no', type: 'string' },
      { name: '数量', comment: 'quantity', type: 'string' },
      { name: '可用数量', comment: 'available_qty', type: 'string' },
    ],
  },
  {
    tag_name: '出库单',
    properties: [
      { name: '出库单号', comment: 'so_id', type: 'string' },
      { name: '出库日期', comment: 'ship_date', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '运单',
    properties: [
      { name: '运单号', comment: 'waybill_no', type: 'string' },
      { name: '承运商', comment: 'carrier', type: 'string' },
      { name: '发运时间', comment: 'ship_time', type: 'string' },
      { name: '签收时间', comment: 'delivery_time', type: 'string' },
      { name: '是否签收', comment: 'is_delivered', type: 'string' },
    ],
  },
  {
    tag_name: '物料',
    properties: [
      { name: '物料编号', comment: 'material_code', type: 'string' },
      { name: '物料名称', comment: 'material_name', type: 'string' },
      { name: '单位', comment: 'uom', type: 'string' },
      { name: '单价', comment: 'unit_price', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '供应', source: '供应商', target: '物料' },
  { name: '采购自', source: '采购单', target: '供应商' },
  { name: '收货自', source: '收货单', target: '采购单' },
  { name: '入库到', source: '收货单', target: '仓库' },
  { name: '存放于', source: '库存', target: '库位' },
  { name: '出库自', source: '出库单', target: '仓库' },
  { name: '发运为', source: '出库单', target: '运单' },
]

const Template: GraphTemplate = {
  name: '供应链仓储物流模板',
  description:
    '围绕“供应商、采购/销售订单、仓储库存、运输节点”的关系，清晰展示供应链流向与库存周转。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
