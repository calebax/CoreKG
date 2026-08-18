import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'
import avatar from './avatar.png'

const graphTags: GraphTag[] = [
  {
    tag_name: '设备',
    properties: [
      { name: '设备名称', type: 'string' },
      { name: '型号', type: 'string' },
      { name: '设备类型', type: 'string' }, // 路由器/接入点/AP/中继/网卡/网关/Mesh节点
      { name: '形态', type: 'string' }, // 桌面/壁挂/吸顶/户外/USB/PCIe
      { name: '处理器', type: 'string' }, // SoC/主频/架构
      { name: '内存', type: 'string' }, // RAM
      { name: '存储', type: 'string' }, // Flash/eMMC
      { name: '以太网端口', type: 'string' }, // 数量/速率/LAN/WAN/PoE
      { name: '天线配置', type: 'string' }, // 数量/增益/dBi/内外置
      { name: 'MIMO与流数', type: 'string' }, // 2x2/4x4/8x8/Spatial Streams
      { name: 'Mesh支持', type: 'string' },
      { name: '最高速率', type: 'string' }, // 汇聚速率 Gbps
      { name: '功耗', type: 'string' },
      { name: '尺寸与重量', type: 'string' },
      { name: '工作温度范围', type: 'string' },
      { name: '电源规格', type: 'string' }, // 适配器/PoE/电压电流
      { name: '固件', type: 'string' }, // 官方/第三方(OpenWrt/RouterOS/UniFi)
      { name: '安全特性', type: 'string' }, // 防火墙/ACL/WPA3/802.1X
      { name: '上市时间', type: 'string' },
      { name: '认证', type: 'string' }, // CE/FCC/SRRC/RoHS
    ],
  },
  {
    tag_name: '无线标准',
    properties: [
      { name: '标准名', type: 'string' }, // IEEE 802.11a/b/g/n/ac/ax/be
      { name: '频段', type: 'string' }, // 2.4GHz/5GHz/6GHz
      { name: '信道带宽', type: 'string' }, // 20/40/80/160/320 MHz
      { name: '调制编码', type: 'string' }, // OFDM/256-QAM/1024-QAM/4096-QAM
      { name: '特性', type: 'string' }, // MU-MIMO/OFDMA/TWT/BSS Coloring
    ],
  },
  {
    tag_name: '芯片组',
    properties: [
      { name: '厂商', type: 'string' }, // Qualcomm/Broadcom/MediaTek/Realtek/Intel
      { name: '芯片型号', type: 'string' },
      { name: '制程', type: 'string' },
      { name: 'CPU架构', type: 'string' }, // ARM/x86
      { name: '射频特性', type: 'string' }, // 发射功率/接收灵敏度
      { name: '支持标准', type: 'string' },
    ],
  },
  {
    tag_name: '频段与信道',
    properties: [
      { name: '频段', type: 'string' }, // 2.4GHz/5GHz/6GHz
      { name: '信道范围', type: 'string' }, // 1-13 / 36-165 / 1-233
      { name: '地区限规', type: 'string' }, // CN/US/EU 规则
    ],
  },
  {
    tag_name: '协议与特性',
    properties: [
      { name: '名称', type: 'string' }, // WPA3/802.1X/RADIUS/802.11k/v/r
      { name: '类别', type: 'string' }, // 安全/漫游/QoS/管理
      { name: '描述', type: 'string' },
    ],
  },
  {
    tag_name: '固件与软件',
    properties: [
      { name: '名称', type: 'string' }, // OpenWrt/RouterOS/EdgeOS/UniFi/企业控制器
      { name: '版本', type: 'string' },
      { name: '功能模块', type: 'string' }, // 控制器/云管理/SD‑WAN/VPN/QoS
      { name: '许可证', type: 'string' },
    ],
  },
  {
    tag_name: '企业或机构',
    properties: [
      { name: '名称', type: 'string' },
      { name: '机构类型', type: 'string' }, // 厂商/代工/运营商/系统集成商/认证机构
      { name: '国家', type: 'string' },
      { name: '地址', type: 'string' },
      { name: '产品线', type: 'string' }, // 路由/AP/网关/网卡/控制器
    ],
  },
  {
    tag_name: '认证与法规',
    properties: [
      { name: '认证名称', type: 'string' }, // CE/FCC/SRRC/Anatel/TELEC
      { name: '编号', type: 'string' },
      { name: '适用地区', type: 'string' },
      { name: '发布日期', type: 'string' },
    ],
  },
  {
    tag_name: '测试报告',
    properties: [
      { name: '报告编号', type: 'string' },
      { name: '测试项目', type: 'string' }, // 吞吐/覆盖/时延/并发/稳定性/EMC
      { name: '测试方法', type: 'string' }, // RFC2544/ixChariot/ATT/内外场
      { name: '结果', type: 'string' },
      { name: '测试机构', type: 'string' },
      { name: '日期', type: 'string' },
    ],
  },
  {
    tag_name: '应用场景',
    properties: [
      { name: '名称', type: 'string' }, // 家用/企业/校园/酒店/园区/工业/户外
      { name: '规模', type: 'string' }, // 终端数/AP数/面积
      { name: 'SLA指标', type: 'string' }, // 吞吐/覆盖/时延/抖动/丢包
    ],
  },
  {
    tag_name: '故障与告警',
    properties: [
      { name: '事件名称', type: 'string' }, // 掉线/干扰/过热/端口Down
      { name: '严重程度', type: 'string' },
      { name: '发生率', type: 'string' },
      { name: '影响组件', type: 'string' }, // 无线/交换/电源/回程
      { name: '解决方案', type: 'string' },
    ],
  },
]
const graphTagRelationships: GraphTagRelationship[] = [
  // 设备相关
  { name: '采用芯片组', source: '设备', target: '芯片组' }, // uses_chipset
  { name: '支持', source: '设备', target: '无线标准' }, // supports_standard
  { name: '工作于', source: '设备', target: '频段与信道' }, // operates_on_band_channel
  { name: '集成', source: '设备', target: '协议与特性' }, // integrates_feature
  { name: '运行', source: '设备', target: '固件与软件' }, // runs_firmware
  { name: '由…生产或销售', source: '设备', target: '企业或机构' }, // manufactured_or_sold_by
  { name: '获得', source: '设备', target: '认证与法规' }, // certified_by
  { name: '通过测试', source: '设备', target: '测试报告' }, // validated_by_test
  { name: '面向', source: '设备', target: '应用场景' }, // targeted_for
  { name: '产生', source: '设备', target: '故障与告警' }, // generates_alarm

  // 芯片组相关
  { name: '提供给', source: '芯片组', target: '设备' }, // provided_to
  { name: '支持', source: '芯片组', target: '无线标准' }, // supports
  { name: '符合', source: '芯片组', target: '认证与法规' }, // complies_with

  // 无线标准与频谱
  { name: '定义信道于', source: '无线标准', target: '频段与信道' }, // defines_channels_on
  { name: '适用于', source: '无线标准', target: '应用场景' }, // applicable_to

  // 固件与软件相关
  { name: '管理', source: '固件与软件', target: '设备' }, // manages
  { name: '支持', source: '固件与软件', target: '协议与特性' }, // supports_feature

  // 企业或机构相关
  { name: '生产或销售', source: '企业或机构', target: '设备' }, // manufactures_or_sells
  { name: '提供', source: '企业或机构', target: '固件与软件' }, // provides_software
  { name: '出具或参与', source: '企业或机构', target: '测试报告' }, // issues_or_participates

  // 认证/法规相关
  { name: '认证', source: '认证与法规', target: '设备' }, // certifies
  { name: '约束', source: '认证与法规', target: '频段与信道' }, // constrains

  // 测试与运维相关
  { name: '测试对象', source: '测试报告', target: '设备' }, // tests
  { name: '覆盖场景', source: '测试报告', target: '应用场景' }, // covers_scenario
  { name: '发现', source: '测试报告', target: '故障与告警' }, // discovers_issue

  // 场景与可靠性
  { name: '发生于', source: '故障与告警', target: '设备' }, // occurs_on
  { name: '影响', source: '故障与告警', target: '应用场景' }, // impacts
]
const Template: GraphTemplate = {
  name: '无线通信设备行业模板',
  description:
    '围绕“设备型号、芯片/射频前端、通信协议、供应商/制造商、认证与合规、应用场景”的基本关系，清晰展示无线通信设备信息',
  avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
