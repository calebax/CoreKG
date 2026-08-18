import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '用户',
    properties: [
      { name: '用户ID', comment: 'user_id', type: 'string' },
      { name: '昵称', comment: 'nickname', type: 'string' },
      { name: '手机号', comment: 'phone', type: 'string' },
      { name: '会员等级', comment: 'member_level', type: 'string' },
      { name: '是否拉黑', comment: 'is_blocked', type: 'string' },
    ],
  },
  {
    tag_name: '商品',
    properties: [
      { name: '商品ID', comment: 'product_id', type: 'string' },
      { name: '商品名称', comment: 'product_name', type: 'string' },
      { name: '类目', comment: 'category', type: 'string' },
      { name: '品牌', comment: 'brand', type: 'string' },
    ],
  },
  {
    tag_name: 'SKU',
    properties: [
      { name: 'SKU编号', comment: 'sku_id', type: 'string' },
      { name: '条码', comment: 'barcode', type: 'string' },
      { name: '售价', comment: 'price', type: 'string' },
      { name: '重量(kg)', comment: 'weight_kg', type: 'string' },
      { name: '是否上架', comment: 'is_active', type: 'string' },
    ],
  },
  {
    tag_name: '订单',
    properties: [
      { name: '订单号', comment: 'order_id', type: 'string' },
      { name: '下单时间', comment: 'order_time', type: 'string' },
      { name: '订单状态', comment: 'order_status', type: 'string' },
      { name: '应付金额', comment: 'amount_payable', type: 'string' },
      { name: '实付金额', comment: 'amount_paid', type: 'string' },
    ],
  },
  {
    tag_name: '仓库',
    properties: [
      { name: '仓库编号', comment: 'warehouse_id', type: 'string' },
      { name: '仓库名称', comment: 'warehouse_name', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
      { name: '地址', comment: 'address', type: 'string' },
    ],
  },
  {
    tag_name: '物流单',
    properties: [
      { name: '运单号', comment: 'waybill_no', type: 'string' },
      { name: '承运商', comment: 'carrier', type: 'string' },
      { name: '发货时间', comment: 'ship_time', type: 'string' },
      { name: '签收时间', comment: 'delivery_time', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
    ],
  },
  {
    tag_name: '评价',
    properties: [
      { name: '评价ID', comment: 'review_id', type: 'string' },
      { name: '评分', comment: 'rating', type: 'string' },
      { name: '内容', comment: 'content', type: 'string' },
      { name: '创建时间', comment: 'created_at', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '下单', source: '用户', target: '订单' },
  { name: '包含', source: '订单', target: 'SKU' },
  { name: '隶属商品', source: 'SKU', target: '商品' },
  { name: '存放于', source: 'SKU', target: '仓库' },
  { name: '发运', source: '订单', target: '物流单' },
  { name: '撰写评价', source: '用户', target: '评价' },
  { name: '评价对象', source: '评价', target: '商品' },
]
const Template: GraphTemplate = {
  name: '零售电商模板',
  description:
    '围绕“商品、库存、订单、客户”的关系，清晰展示商品信息、交易流程与库存履约。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
