import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '人物',
    properties: [{ name: '人物名称', comment: 'person_name', type: 'string' }],
  },
  {
    tag_name: '阵营',
    properties: [{ name: '所属阵营', comment: 'faction', type: 'string' }],
  },
  {
    tag_name: '头衔',
    properties: [{ name: '真实地位', comment: 'title', type: 'string' }],
  },
  {
    tag_name: '生平事迹',
    properties: [{ name: '事迹总结', comment: 'deed_summary', type: 'string' }],
  },
]
// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '妻子', source: '人物', target: '人物' },
  { name: '嫔妃', source: '人物', target: '人物' },
  { name: '弟弟', source: '人物', target: '人物' },
  { name: '丈夫', source: '人物', target: '人物' },
  { name: '同盟', source: '人物', target: '人物' },
  { name: '姐姐', source: '人物', target: '人物' },
  { name: '敌对', source: '人物', target: '人物' },
  { name: '心腹', source: '人物', target: '人物' },
  { name: '儿子', source: '人物', target: '人物' },
  { name: '养子', source: '人物', target: '人物' },
  { name: '父亲', source: '人物', target: '人物' },
  { name: '昔日姐妹', source: '人物', target: '人物' },
  { name: '妹妹', source: '人物', target: '人物' },
  { name: '昔日恋人', source: '人物', target: '人物' },
  { name: '女儿', source: '人物', target: '人物' },
  { name: '地下情', source: '人物', target: '人物' },
  { name: '养女', source: '人物', target: '人物' },
  { name: '单恋', source: '人物', target: '人物' },
  { name: '哥哥', source: '人物', target: '人物' },
  { name: '主子', source: '人物', target: '人物' },
  { name: '母亲', source: '人物', target: '人物' },
  { name: '养母', source: '人物', target: '人物' },
  { name: '养父', source: '人物', target: '人物' },
  { name: '偷吃', source: '人物', target: '人物' },
  { name: '私生女', source: '人物', target: '人物' },
]
const Template: GraphTemplate = {
  name: '甄嬛传模板',
  description: '甄嬛传模板',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
