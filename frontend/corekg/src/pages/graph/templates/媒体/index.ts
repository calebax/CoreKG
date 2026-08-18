import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '作品',
    properties: [
      { name: '作品ID', comment: 'work_id', type: 'string' },
      { name: '标题', comment: 'title', type: 'string' },
      { name: '类型', comment: 'genre', type: 'string' },
      { name: '发布时间', comment: 'release_date', type: 'string' },
    ],
  },
  {
    tag_name: '创作者',
    properties: [
      { name: '创作者ID', comment: 'creator_id', type: 'string' },
      { name: '姓名/艺名', comment: 'name', type: 'string' },
      { name: '国籍/地区', comment: 'country_region', type: 'string' },
    ],
  },
  {
    tag_name: '版权',
    properties: [
      { name: '版权ID', comment: 'copyright_id', type: 'string' },
      { name: '版权类型', comment: 'copyright_type', type: 'string' },
      { name: '有效期开始', comment: 'valid_from', type: 'string' },
      { name: '有效期结束', comment: 'valid_to', type: 'string' },
    ],
  },
  {
    tag_name: '授权协议',
    properties: [
      { name: '协议ID', comment: 'license_id', type: 'string' },
      { name: '授权范围', comment: 'scope', type: 'string' },
      { name: '授权费用', comment: 'license_fee', type: 'string' },
      { name: '生效日期', comment: 'effective_date', type: 'string' },
    ],
  },
  {
    tag_name: '平台',
    properties: [
      { name: '平台ID', comment: 'platform_id', type: 'string' },
      { name: '平台名称', comment: 'platform_name', type: 'string' },
      { name: '类型', comment: 'platform_type', type: 'string' },
    ],
  },
  {
    tag_name: '播放记录',
    properties: [
      { name: '记录ID', comment: 'play_id', type: 'string' },
      { name: '开始时间', comment: 'start_time', type: 'string' },
      { name: '时长(秒)', comment: 'duration_sec', type: 'string' },
      { name: '是否完播', comment: 'is_completed', type: 'string' },
    ],
  },
  {
    tag_name: '标签',
    properties: [
      { name: '标签ID', comment: 'tag_id', type: 'string' },
      { name: '标签名称', comment: 'tag_name', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '创作', source: '创作者', target: '作品' },
  { name: '关联版权', source: '作品', target: '版权' },
  { name: '授权', source: '版权', target: '授权协议' },
  { name: '分发到', source: '授权协议', target: '平台' },
  { name: '产生播放', source: '平台', target: '播放记录' },
  { name: '打标签', source: '作品', target: '标签' },
]
const Template: GraphTemplate = {
  name: '内容媒体版权模板',
  description:
    '围绕“作品、版权方/授权方、授权范围与期限、分发渠道”的关系，清晰展示版权归属、授权链路与使用合规。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
