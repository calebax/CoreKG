import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '电路图子结点',
    properties: [
      {
        name: '连接ECU',
        type: 'string',
        comment: '记录子结点连接的主元件',
      },
      {
        name: '线和引脚',
        type: 'string',
        comment: '在维修手册中的引脚和线缆信息',
      },
    ],
  },
  {
    tag_name: '故障可能原因',
    properties: [{ name: '具体描述', type: 'string', comment: '' }],
  },
  {
    tag_name: 'ECU',
    description: 'ECU是汽车中用于控制发动机、变速箱等系统的核心计算机。',
    properties: [{ name: '具体描述', type: 'string', comment: '' }],
  },
  {
    tag_name: 'DTC',
    description: 'DTC是车辆ECU检测到故障时生成的代码，为7位字母和数字的组合。',
    properties: [
      {
        name: '解决办法路径',
        type: 'string',
        comment: '故障码的解决办法路径（参考）',
      },
    ],
  },
  {
    tag_name: '故障诊断步骤',
    description: '单个步骤。',
    properties: [
      {
        name: '步骤序号',
        type: 'string',
        comment: '该检查项在流程中的顺序编号',
      },
      {
        name: '涉及的ECU型号和针脚',
        type: 'string',
        comment: '采集诊断步骤中对应的ECU电路图型号与需检查的针脚',
      },
      {
        name: '详细步骤',
        type: 'string',
        comment: '具体操作指令',
      },
      {
        name: '所需工具',
        type: 'string',
        comment: '执行该步骤所需的工具',
      },
    ],
  },
  {
    tag_name: '电路图',
    properties: [
      {
        name: '所属系统',
        type: 'string',
        comment: '该电路图归属的车辆子系统',
      },
      {
        name: 'ECU名称',
        type: 'string',
        comment: '图中主要的ECU名称',
      },
      {
        name: '接口名词及接口号',
        type: 'string',
        comment: '图中包含的传感器字典列表字符词',
      },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '出现故障', source: 'ECU', target: 'DTC' },
  { name: '触发报错', source: '故障可能原因', target: 'DTC' },
  { name: '对应电路图', source: 'ECU', target: '电路图' },
  {
    name: '故障诊断方法',
    source: 'DTC',
    target: '故障诊断步骤',
  },
  {
    name: '相关电路图',
    source: '故障诊断步骤',
    target: '电路图',
  },
  { name: '连接', source: '电路图', target: '电路图子结点' },
]
const Template: GraphTemplate = {
  name: '汽车维修故障诊断模板2',
  description:
    '围绕"电子控制单元、故障码、故障现象、故障原因、故障部位、故障诊断、解决办法路径、电路图、维修操作、维修注意、维修工具、维修参数"的关系，清晰展示汽车故障诊断与维修的完整流程。',
  tags: graphTags,
  edges: graphTagRelationships,
}

export default Template
