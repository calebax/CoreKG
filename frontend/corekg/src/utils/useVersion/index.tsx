import { useLoginGlobalData } from '../useLoginGlobalData'

export type ContextValue = {
  version?: {
    is_purchased: boolean
    name: string
    qa: { used: number; quota: number }
    agent: { used: number; quota: number }
    // 知识库
    disk: { used: string; quota: string; use_ratio: number }
    // 团队
    employee: { used: number; quota: number }
  }
  refresh: () => void
}

/** 版本信息(试用 专业 企业) */
export const useVersion = () => {
  return useLoginGlobalData().version
}
