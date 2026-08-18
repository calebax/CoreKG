import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '电子控制单元',
    properties: [
      { name: 'ID', type: 'string', comment: 'ECU的唯一标识符' },
      { name: '名称', type: 'string', comment: 'ECU的功能名称' },
      {
        name: '支持协议',
        type: 'string',
        comment: '支持的诊断/通信协议（多值，可用分号分隔）',
      },
      { name: '硬件版本', type: 'string', comment: 'ECU 硬件版本号' },
    ],
  },
  {
    tag_name: '故障码',
    properties: [
      {
        name: '代码',
        type: 'string',
        comment: '符合SAE J2012标准的DTC编码（格式：[PCBU][0-9]{4}）',
      },
      { name: '描述', type: 'string', comment: '故障码对应的自然语言描述' },
      {
        name: '类别',
        type: 'string',
        comment: 'DTC 首字母类别（枚举：P=动力, C=底盘, B=车身, U=网络）',
      },
      {
        name: '严重等级',
        type: 'string',
        comment: '故障严重程度（1=轻微，5=危急）',
      },
    ],
  },
  {
    tag_name: '故障现象',
    properties: [
      { name: '描述', type: 'string', comment: '用户可观察到的异常行为' },
      { name: '严重等级', type: 'string', comment: '严重程度（1–5）' },
      {
        name: 'SOD等级',
        type: 'string',
        comment:
          '按照 SOD（Severity, Occurrence, Detection）标准评定的严重性等级',
      },
    ],
  },
  {
    tag_name: '故障原因',
    properties: [
      { name: '描述', type: 'string', comment: '故障的根本或直接原因' },
      {
        name: '类别',
        type: 'string',
        comment:
          '原因大类（枚举：Hardware, Software, Wiring, Environmental, Aging）',
      },
    ],
  },
  {
    tag_name: '故障部位',
    properties: [
      { name: '名称', type: 'string', comment: '故障发生的物理位置或组件名称' },
      { name: '安装位置', type: 'string', comment: '在整车中的安装路径描述' },
    ],
  },
  {
    tag_name: '故障诊断',
    properties: [
      {
        name: '步骤序号',
        type: 'string',
        comment: '该检查项在流程中的顺序编号',
      },
      { name: '详细步骤', type: 'string', comment: '具体操作指令（可含多步）' },
      { name: '所需工具', type: 'string', comment: '执行该步骤所需的工具' },
    ],
  },
  {
    tag_name: '解决办法路径',
    properties: [
      { name: '类型', type: 'string', comment: '解决方案的动作类型' },
      { name: '路径引用', type: 'string', comment: '在诊断手册中的导航路径' },
      { name: '描述', type: 'string', comment: '解决方案的简要说明' },
    ],
  },
  {
    tag_name: '电路图',
    properties: [
      { name: '图编号', type: 'string', comment: '电路图的唯一编号' },
      { name: '标题', type: 'string', comment: '电路图的标题' },
      { name: '所属ECU', type: 'string', comment: '图中主要涉及的 ECU 名称' },
      {
        name: 'ECU型号',
        type: 'string',
        comment: '对应的 ECU 连接器或硬件型号',
      },
      { name: '所属系统', type: 'string', comment: '该电路图归属的车辆子系统' },
      {
        name: '包含传感器',
        type: 'string',
        comment: '图中包含的传感器列表（多值）',
      },
      {
        name: '包含引脚',
        type: 'string',
        comment: 'ECU 在图中使用的引脚编号（多值）',
      },
    ],
  },
  {
    tag_name: '维修操作',
    properties: [
      { name: '所属系统', type: 'string', comment: '该操作针对的车辆系统' },
      {
        name: '操作类别',
        type: 'string',
        comment: '操作类型（枚举：拆卸, 安装等）',
      },
      { name: '目录序号', type: 'string', comment: '手册中的章节编号' },
      { name: '步骤详情', type: 'string', comment: '详细操作流程' },
      {
        name: '操作提示',
        type: 'string',
        comment: '关键注意事项或技巧（多值）',
      },
      { name: '推荐工具', type: 'string', comment: '建议使用的专用工具' },
    ],
  },
  {
    tag_name: '维修注意',
    properties: [
      { name: '所属系统', type: 'string', comment: '注意事项适用的系统' },
      { name: '注意内容', type: 'string', comment: '一般性安全或操作提醒' },
      {
        name: '警告内容',
        type: 'string',
        comment: '严重风险警告（如高压、爆炸）',
      },
    ],
  },
  {
    tag_name: '维修工具',
    properties: [
      { name: '工具编号', type: 'string', comment: '工具的唯一编号' },
      { name: '工具名称', type: 'string', comment: '工具的标准名称' },
      { name: '工具说明', type: 'string', comment: '工具用途或使用场景' },
      { name: '工具外形图', type: 'string', comment: '工具图片的 URI 地址' },
    ],
  },
  {
    tag_name: '维修参数',
    properties: [
      { name: '紧固件名称', type: 'string', comment: '螺栓/螺母的名称' },
      { name: '规格', type: 'string', comment: '尺寸规格（如 M6×20）' },
      { name: '扭矩要求', type: 'string', comment: '扭矩范围及单位' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '出现故障', source: '电子控制单元', target: '故障码' },
  { name: '对应电路图', source: '电子控制单元', target: '电路图' },
  { name: '对应故障部位', source: '电子控制单元', target: '故障部位' },
  { name: '相关维修步骤', source: '电子控制单元', target: '维修操作' },
  { name: '对应现象', source: '故障码', target: '故障现象' },
  { name: '解决方法', source: '故障码', target: '解决办法路径' },
  { name: '关联ECU', source: '故障码', target: '电子控制单元' },
  { name: '表现为', source: '故障现象', target: '故障原因' },
  { name: '发生于', source: '故障现象', target: '故障部位' },
  { name: '可通过诊断步骤诊断', source: '故障现象', target: '故障诊断' },
  { name: '触发报错', source: '故障原因', target: '故障码' },
  { name: '造成表象', source: '故障原因', target: '故障现象' },
  { name: '位于', source: '故障原因', target: '故障部位' },
  { name: '具有解决办法', source: '故障原因', target: '解决办法路径' },
  { name: '包含步骤', source: '解决办法路径', target: '故障诊断' },
  { name: '检查顺序', source: '故障诊断', target: '故障诊断' },
  { name: '相关电路图', source: '故障诊断', target: '电路图' },
  { name: '相关操作步骤', source: '故障诊断', target: '维修操作' },
  { name: '前置注意事项', source: '维修操作', target: '维修注意' },
  { name: '使用工具', source: '维修操作', target: '维修工具' },
  { name: '相关维修参数', source: '维修操作', target: '维修参数' },
]

const Template: GraphTemplate = {
  name: '汽车维修故障诊断模板',
  description:
    '围绕"电子控制单元、故障码、故障现象、故障原因、故障部位、故障诊断、解决办法路径、电路图、维修操作、维修注意、维修工具、维修参数"的关系，清晰展示汽车故障诊断与维修的完整流程。',
  tags: graphTags,
  edges: graphTagRelationships,
}

export default Template
