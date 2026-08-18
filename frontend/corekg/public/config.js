// 环境配置，需与 src/utils/useDeployConfig/type.d.ts 保持一致。
// version: saas | international | custom
// mode: 仅 custom 有效，可选 h3c | cimc | default | jiefang
//
// 下列仅列出“不同环境表现不一样”的功能，其它功能各环境保持一致。
// 注：“本地/测试/生产”指 Vite 构建环境 development/test/production，可与任意 version 组合。
//
// 智能体：有 saas、custom；无 international
// 知识图谱：有（需开通 graph 权限）saas、custom；无 international
// 数据库 / 问答对知识库：有 saas、custom；无 international
// 订单管理 / 用量 / 套餐：有 saas；无 international、custom
// 忘记密码 / 用户协议：有 saas、international；无 custom
// 联系我们：有 saas、international；无 custom
// 消息与通知：有 saas、international、custom(cimc)；无 custom(其他 mode)
// 帮助与支持：有 saas、custom(h3c/default/jiefang)；无 international、custom(cimc)
// 同义词 / 行业名词：有 本地/测试构建、custom(cimc/h3c)；无 其余生产构建
// 外部数据源：有 本地/测试构建、custom(cimc) 生产构建；无 其余生产构建
// AI问答输入框字数上限：本地/测试/custom 环境最低 10000，可继续通过 qaInputMaxLength 往上调；saas/international 默认 500

/* 新华三
window.__DEPLOYCONFIG = {
  version: 'custom',
  coze_url: '/coze',
  mode: 'h3c',
  logo: '/icons/h3c/logo.png',
  title: '/icons/h3c/title.svg',
  appName: 'TuringQuery',
  qaInputMaxLength: 10000,
  favicon: {
    light: '/icons/h3c/favicon-light.ico',
    dark: '/icons/h3c/favicon-dark.ico',
  },
}
*/

/* saas
window.__DEPLOYCONFIG = {
  version: 'saas',
  coze_url: 'https://coze.corekg.com',
  mode: '',
  logo: '/icons/saas/logo.svg',
  title: '/icons/saas/title.svg',
  appName: 'CoreKG AI',
  qaInputMaxLength: 500,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
}
*/

/* 海外版
window.__DEPLOYCONFIG = {
  version: 'international',
  coze_url: '/coze',
  mode: '',
  logo: '/icons/international/logo.svg',
  title: '/icons/international/title.svg',
  appName: 'OpenPO',
  qaInputMaxLength: 500,
  favicon: {
    light: '/icons/international/favicon-light.ico',
    dark: '/icons/international/favicon-light.ico',
  },
}
*/

/* 中集来福士
window.__DEPLOYCONFIG = {
  version: 'custom',
  coze_url: '/coze',
  mode: 'cimc',
  logo: '',
  title: '',
  appName: 'CIMC RAFFLES',
  qaInputMaxLength: 10000,
  favicon: {
    light: '/icons/cicm/favicon.ico',
    dark: '/icons/cicm/favicon.ico',
  },
  globalGreeting: '一起开启问答新体验',
}
*/

/* 默认私有化
window.__DEPLOYCONFIG = {
  version: 'custom',
  coze_url: '/coze',
  mode: 'default',
  logo: '',
  title: '',
  appName: 'CoreKG AI',
  qaInputMaxLength: 10000,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
  globalGreeting: '一起开启问答新体验',
}
*/

window.__DEPLOYCONFIG = {
  version: 'saas',
  coze_url: 'https://coze.corekg.com',
  // coze_url: 'https://example.com', // 临时修改 coze 地址
  mode: '',
  logo: '/icons/saas/logo.svg',
  title: '/icons/saas/title.svg',
  appName: 'CoreKG AI',
  qaInputMaxLength: 500,
  favicon: {
    light: '/icons/saas/favicon-light.ico',
    dark: '/icons/saas/favicon-dark.ico',
  },
}
