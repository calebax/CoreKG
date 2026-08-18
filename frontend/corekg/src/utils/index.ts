import { useMemoizedFn } from 'ahooks'
import clsx, { ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'
import * as filterConst from '@/filter/const'

export { withSuspense } from './withSuspense'
export {
  GRAPH_NODE_NAME_ILLEGAL_CHAR_REGEXP,
  GRAPH_NODE_NAME_INVALID_ERROR,
  GRAPH_NODE_NAME_INVALID_TOOLTIP,
  validateGraphNodeName,
} from './validate'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const consoleJson = (obj: unknown) => {
  console.log(JSON.stringify(obj, null, 2))
}
export const isJsonString = (str: string) => {
  try {
    JSON.parse(str)
    return true
  } catch {
    // if (e instanceof Error) {
    //   console.error('Invalid JSON:', e.message)
    // }
    return false
  }
}

/**
 * @param {number} n
 * @return {Promise}
 */
export const sleep = (s: number) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      return resolve('')
    }, s)
  })
}

export const constDataToArray = (
  obj: Record<string, string>,
  firstItem: { value: string | number | boolean; label: string } | null = null,
  type = 'string',
) => {
  const arr: Array<{ value: string | number | boolean; label: string }> = []
  if (type === 'string') {
    Object.keys(obj).forEach((k) => {
      if (k !== 'default') {
        arr.push({ value: k, label: obj[k] })
      }
    })
  } else if (type === 'number') {
    Object.keys(obj).forEach((k) => {
      if (k !== 'default') {
        arr.push({ value: Number(k), label: obj[k] })
      }
    })
  } else if (type === 'boolean') {
    Object.keys(obj).forEach((k) => {
      if (k !== 'default') {
        arr.push({ value: k === 'true', label: obj[k] })
      }
    })
  }
  if (firstItem) {
    arr.unshift(firstItem)
  }
  return arr
}

export const filter = (filterName: string, arg: string | number) => {
  const data = filterConst[`${filterName}Data` as keyof typeof filterConst]
  if (data) {
    return data[arg as keyof typeof data] || data.default
  }
  return ''
}

export const sortJson = <T extends Record<string, unknown>>(
  arr: Array<T>,
  key: keyof T,
  order = 'asc',
) => {
  if (order === 'asc') {
    return arr.sort((a, b) => {
      return a[key] < b[key] ? -1 : a[key] > b[key] ? 1 : 0
    })
  } else {
    return arr.sort((a, b) => {
      return a[key] > b[key] ? -1 : a[key] < b[key] ? 1 : 0
    })
  }
}

/** 将一个dom或其ref滚动至底部 */
export const scrollToEnd = (
  arg: HTMLElement | { current?: HTMLElement | null },
) => {
  const dom = 'scrollTo' in arg ? arg : arg.current
  if (dom) {
    dom.scrollTo({
      top: dom.scrollHeight,
    })
  }
}

/**
 * 拆分文件名和拓展名
 * @returns ext不含'.' 无拓展名则是空字符串
 */
export const spliteFileName = (s: string) => {
  const lastDotIndex = s.lastIndexOf('.')
  if (lastDotIndex === -1 || lastDotIndex === 0) {
    // 没有扩展名 或 以.开头的
    return { name: s, ext: '' }
  }
  const name = s.slice(0, lastDotIndex)
  const ext = s.slice(lastDotIndex + 1) // 不含点
  return { name, ext }
}

/** 合并多个数组并去重 */
export function uniqueArray<T>(...args: T[][]) {
  return [...new Set(args.flat())] as T[]
}

/** 以document.execCommand('copy')的方式复制一段文字 */
export const copyText = (s: string) => {
  const textarea = document.createElement('textarea')
  textarea.value = s
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

/** 转化为正整数 */
export const convertToInteger = (s: any) => {
  const _id = Number(s)
  if (Number.isInteger(_id)) return _id
  return null
}

/**
 * 格式化文件大小显示
 * @param bytes 文件大小（字节）
 * @returns 格式化后的文件大小字符串
 */
export const formatFileSize = (bytes: number): string => {
  // 如果为0或负数，返回 '-'
  if (!bytes || bytes <= 0) {
    return '-'
  }

  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const k = 1024
  const dm = 2 // 保留2位小数

  // 计算对应的单位索引
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  // 防止索引超出范围
  const sizeIndex = Math.min(i, sizes.length - 1)

  // 计算实际大小
  const size = parseFloat((bytes / Math.pow(k, sizeIndex)).toFixed(dm))

  // 如果大小是整数，则不显示小数点
  const displaySize = size % 1 === 0 ? size.toString() : size.toFixed(dm)

  return `${displaySize}${sizes[sizeIndex]}`
}

export function throttle<T extends (...args: any[]) => any>(
  func: T,
  delay: number,
): (...args: Parameters<T>) => ReturnType<T> | undefined {
  // 存储上一次执行的时间戳，初始化为 0
  let previous = 0

  return function (
    this: ThisParameterType<T>,
    ...args: Parameters<T>
  ): ReturnType<T> | undefined {
    // 获取当前时间戳
    const now = Date.now()
    // 若当前时间与上一次执行时间的间隔大于延迟时间
    if (now - previous > delay) {
      // 执行传入的函数，并使用保存的上下文和参数
      const result = func.apply(this, args)
      // 更新上一次执行的时间戳
      previous = now
      return result
    }
    return undefined
  }
}

/** 适用于`需要重置的Modal` 每当open变为true时 都提供一个新的key */
export const useOpen = (defaultValue?: boolean) => {
  const [open, _setOpen] = useState(Boolean(defaultValue))
  const keyRef = useRef<number>(0)
  const setOpen = useMemoizedFn(() => {
    ++keyRef.current
    _setOpen(true)
  })
  const setClose = useMemoizedFn(() => {
    _setOpen(false)
  })
  const toggle = useMemoizedFn(() => {
    _setOpen((v) => {
      if (!v) {
        ++keyRef.current
      }
      return !v
    })
  })
  return [open, { toggle, setOpen, setClose }, keyRef.current] as const
}

/**
 * 检查模块权限
 * @param license license 对象，包含 modules 数组（modules 是允许展示的模块列表）
 * @param module 要检查的模块名称
 * @returns 是否有权限展示该模块（开发环境始终返回 true）
 * @description
 * - modules 数组中的模块是允许展示的模块
 * - 如果 license 不存在或 modules 不存在，返回 true（不影响渲染）
 * - 如果 modules 中存在该模块，返回 true（允许展示）
 * - 如果 modules 中不存在该模块，返回 false（不允许展示）
 */
export const hasModulePermission = (
  license: { modules?: CustomModule[] } | undefined,
  module: CustomModule,
): boolean => {
  // 开发环境不生效，始终返回 true
  if (import.meta.env.MODE === 'development' || import.meta.env.DEV) {
    return true
  }
  // license 不存在时不影响渲染，返回 true
  if (!license || !license.modules) {
    return true
  }
  // 检查模块是否在 modules 数组中（modules 是允许展示的模块列表）
  return license.modules.includes(module)
}
