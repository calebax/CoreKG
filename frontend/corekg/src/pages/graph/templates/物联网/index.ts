import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '设备',
    properties: [
      { name: '设备ID', comment: 'device_id', type: 'string' },
      { name: '设备名称', comment: 'device_name', type: 'string' },
      { name: '设备类型', comment: 'device_type', type: 'string' },
      { name: '固件版本', comment: 'firmware_version', type: 'string' },
      { name: '是否在线', comment: 'is_online', type: 'string' },
    ],
  },
  {
    tag_name: '网关',
    properties: [
      { name: '网关ID', comment: 'gateway_id', type: 'string' },
      { name: '网关名称', comment: 'gateway_name', type: 'string' },
      { name: '接入协议', comment: 'protocol', type: 'string' },
      { name: 'IP地址', comment: 'ip_address', type: 'string' },
    ],
  },
  {
    tag_name: '传感器',
    properties: [
      { name: '传感器ID', comment: 'sensor_id', type: 'string' },
      { name: '传感器类型', comment: 'sensor_type', type: 'string' },
      { name: '采样频率(Hz)', comment: 'sample_rate_hz', type: 'string' },
      { name: '精度', comment: 'precision', type: 'string' },
    ],
  },
  {
    tag_name: '遥测点',
    properties: [
      { name: '点位ID', comment: 'telemetry_id', type: 'string' },
      { name: '指标名称', comment: 'metric_name', type: 'string' },
      { name: '单位', comment: 'unit', type: 'string' },
      { name: '最近值', comment: 'last_value', type: 'string' },
    ],
  },
  {
    tag_name: 'SIM卡',
    properties: [
      { name: 'ICCID', comment: 'iccid', type: 'string' },
      { name: '运营商', comment: 'carrier', type: 'string' },
      { name: '套餐名称', comment: 'plan_name', type: 'string' },
      { name: '是否可用', comment: 'is_active', type: 'string' },
    ],
  },
  {
    tag_name: '地点',
    properties: [
      { name: '地点名称', comment: 'location_name', type: 'string' },
      { name: '城市', comment: 'city', type: 'string' },
      { name: '经度', comment: 'longitude', type: 'string' },
      { name: '纬度', comment: 'latitude', type: 'string' },
    ],
  },
  {
    tag_name: '告警',
    properties: [
      { name: '告警ID', comment: 'alarm_id', type: 'string' },
      { name: '级别', comment: 'severity', type: 'string' },
      { name: '产生时间', comment: 'start_time', type: 'string' },
      { name: '是否确认', comment: 'acknowledged', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '接入', source: '设备', target: '网关' },
  { name: '管理', source: '网关', target: '传感器' },
  { name: '产生', source: '传感器', target: '遥测点' },
  { name: '安装SIM', source: '设备', target: 'SIM卡' },
  { name: '安装于', source: '设备', target: '地点' },
  { name: '触发', source: '设备', target: '告警' },
  { name: '关联指标', source: '遥测点', target: '设备' },
]
const Template: GraphTemplate = {
  name: '智慧城市物联网模板',
  description:
    '围绕“城市感知设备、数据采集点、事件告警、管理部门”的关系，清晰展示物联网设备台账、实时数据与告警联动。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
