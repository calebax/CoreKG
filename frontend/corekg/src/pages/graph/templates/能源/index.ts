import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '变电站',
    properties: [
      { name: '站点编号', comment: 'substation_id', type: 'string' },
      { name: '站点名称', comment: 'substation_name', type: 'string' },
      { name: '电压等级(kV)', comment: 'voltage_kv', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
      { name: '是否在役', comment: 'in_service', type: 'string' },
    ],
  },
  {
    tag_name: '输电线路',
    properties: [
      { name: '线路编号', comment: 'line_id', type: 'string' },
      { name: '线路名称', comment: 'line_name', type: 'string' },
      { name: '额定电压(kV)', comment: 'rated_voltage_kv', type: 'string' },
      { name: '长度(km)', comment: 'length_km', type: 'string' },
    ],
  },
  {
    tag_name: '设备',
    properties: [
      { name: '设备编号', comment: 'asset_id', type: 'string' },
      { name: '设备名称', comment: 'asset_name', type: 'string' },
      { name: '设备类型', comment: 'asset_type', type: 'string' },
      { name: '投运日期', comment: 'commission_date', type: 'string' },
      { name: '健康评分', comment: 'health_score', type: 'string' },
    ],
  },
  {
    tag_name: '测点',
    properties: [
      { name: '测点编号', comment: 'point_id', type: 'string' },
      { name: '测点名称', comment: 'point_name', type: 'string' },
      { name: '量测类型', comment: 'measurement_type', type: 'string' },
      { name: '单位', comment: 'unit', type: 'string' },
    ],
  },
  {
    tag_name: '告警',
    properties: [
      { name: '告警ID', comment: 'alarm_id', type: 'string' },
      { name: '级别', comment: 'severity', type: 'string' },
      { name: '产生时间', comment: 'start_time', type: 'string' },
      { name: '恢复时间', comment: 'end_time', type: 'string' },
      { name: '是否确认', comment: 'acknowledged', type: 'string' },
    ],
  },
  {
    tag_name: '工单',
    properties: [
      { name: '工单编号', comment: 'work_order_id', type: 'string' },
      { name: '工单类型', comment: 'work_order_type', type: 'string' },
      { name: '状态', comment: 'status', type: 'string' },
      { name: '创建时间', comment: 'created_at', type: 'string' },
      { name: '完成时间', comment: 'completed_at', type: 'string' },
    ],
  },
  {
    tag_name: '用户',
    properties: [
      { name: '用户编号', comment: 'customer_id', type: 'string' },
      { name: '用户名称', comment: 'customer_name', type: 'string' },
      { name: '用电性质', comment: 'usage_type', type: 'string' },
      { name: '所在城市', comment: 'city', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '连接', source: '输电线路', target: '变电站' },
  { name: '安装于', source: '设备', target: '变电站' },
  { name: '监测', source: '测点', target: '设备' },
  { name: '触发', source: '设备', target: '告警' },
  { name: '检修', source: '工单', target: '设备' },
  { name: '供电到', source: '变电站', target: '用户' },
]
const Template: GraphTemplate = {
  name: '能源电力资产-调度模板',
  description:
    '围绕“电力资产（电站/线路/变电）、负荷与发电计划、调度指令、告警事件”的关系，清晰展示电网资产、负荷调度与运行状态。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
