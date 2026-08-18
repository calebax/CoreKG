import { GraphTemplate } from 'Graph'
import 专利 from './专利'
import 仅测试 from './仅测试'
import 供应链 from './供应链'
import 保险 from './保险'
import 医药 from './医药'
import 媒体 from './媒体'
import 招聘 from './招聘'
import 教育 from './教育'
import 无线 from './无线设备'
import 汽车维修 from './汽车维修'
import 汽车维修故障诊断模板2 from './汽车维修故障诊断模板2'
import 生产 from './生产'
import 能源 from './能源'
import 通用装备 from './通用装备'
import 金融 from './金融'
import 零售 from './零售'

export { default as EmptyTemplate } from './空白'

const Templates: GraphTemplate[] = [
  医药,
  专利,
  无线,
  招聘,
  通用装备,
  生产,
  能源,
  媒体,
  零售,
  金融,
  教育,
  供应链,
  保险,
  汽车维修,
  汽车维修故障诊断模板2,
]

if (
  // 用环境变量判断
  (import.meta.env.MODE === 'test' || import.meta.env.DEV) &&
  // 用host判断
  ['127.0.0.1', 'localhost', 'example.com'].includes(
    window.location.hostname,
  )
) {
  Templates.unshift(仅测试)
}

export { Templates }
