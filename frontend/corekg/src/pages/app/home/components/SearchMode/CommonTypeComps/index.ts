import { FC } from 'react'
import { LegalSearchType } from '../searchType'

export type LegalSearchTypeComp = FC<{ value: any }>

const Components: Record<string, any> = import.meta.glob('./*/index.tsx', {
  eager: true,
})
/** 获取类型不为'doc'的渲染组件 */
export const getCompByType = (
  type: Exclude<LegalSearchType, 'doc'>,
): LegalSearchTypeComp => {
  const entries = Object.entries(Components)
  for (const [key, value] of entries) {
    const res = /^\.\/([a-zA-Z0-9]*)\/index\..*$/g.exec(key)
    if (res?.[1] === type) {
      return value.default
    }
  }
  throw new Error(`type:${type}没有对应的组件`)
}
