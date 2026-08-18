// 第三方库声明
declare module 'cross-storage'
declare module 'prop-types'
declare module 'react-copy-to-clipboard'

// 资源类型声明
declare module '*.svg?react' {
  import React from 'react'
  const SVGComponent: React.FC<React.SVGProps<SVGSVGElement>>
  export default SVGComponent
}

// 工具和过滤器相关声明
declare module '@/filter/const' {
  const filterConst: Record<string, unknown>
  export default filterConst
}

declare module 'github-markdown-css' {
  const content: string
  export default content
}
