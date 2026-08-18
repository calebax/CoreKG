/*
 * eg: format('date', new Date())
 */
import dayjs from 'dayjs'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'

dayjs.extend(utc)
dayjs.extend(timezone)

export default {
  /** *************** time *************** **/
  // 获取UTC时间字符串
  utcTimeString(t: Date | string | number) {
    return dayjs(t).utc().format('YYYY/MM/DD HH:mm')
  },
  // 获取系统当前时间
  localTimeString(
    utcTime: Date | string | number | null,
    format = 'YYYY/MM/DD HH:mm',
  ) {
    if (!utcTime) return ''
    return dayjs(utcTime).utc().format(format)
  },
  // 获取系统当前时间
  localTimeString1(
    utcTime: Date | string | number | null,
    format = 'YYYY/MM/DD HH:mm',
  ) {
    if (!utcTime) return ''
    const systemTime = dayjs.utc(utcTime).local()
    return systemTime.format(format)
  },
  // 获取系统当前时间戳
  localTimeValueOf(utcTime: Date | string | number | null) {
    if (!utcTime) return ''
    const systemTime = dayjs.utc(utcTime).local()
    return systemTime.valueOf()
  },

  month: (t: Date | string | number | null) => {
    if (!t) return ''
    return dayjs(t).format('YYYY/MM')
  },
  date: (t: Date | string | number | null) => {
    if (!t) return ''
    return dayjs(t).format('YYYY/MM/DD')
  },
  minute: (t: Date | string | number | null) => {
    if (!t) return ''
    return dayjs(t).format('YYYY/MM/DD HH:mm')
  },
  second: (t: Date | string | number | null) => {
    if (!t) return ''
    return dayjs(t).format('YYYY/MM/DD HH:mm:ss')
  },
  timestamp: (t: Date | string | number | null) => {
    if (!t) return ''
    return new Date(t as string | number | Date).getTime()
  },
  time: (t: Date | string | number | null) => {
    if (!t) return ''
    return dayjs(t).format('MM/DD HH:mm:ss')
  },
}
