import { useMemo } from 'react'
import { useMemoizedFn } from 'ahooks'
import { Updater } from 'use-immer'
import { AIDialog, DialogList } from '../..'

/** 问答额度受限的回答 */
export const useLimitedAnswer = () => {
  const content = useMemo(() => {
    // Vite 默认 BASE_URL 为 '/'；若写成 `${'/'}/settings/profile` 会得到 '//settings/profile'，
    // 浏览器会当作「协议相对 URL」解析成主机名 settings，错误跳转到 https://settings/profile
    const trimmedBase = import.meta.env.BASE_URL.replace(/\/$/, '')
    const fullPath = trimmedBase
      ? `${trimmedBase}/settings/profile`
      : '/settings/profile'
    return `您的问答额度已用完，如需继续使用，请前往<a href='${fullPath}'>「个人信息」</a>页面查看并升级版本。`
  }, [])
  const limitedAnswer = useMemo(() => {
    const value: AIDialog = {
      role: 'answer',
      thinkingContent: '',
      reference: [],
      status: 'answered',
      content,
    }
    return value
  }, [content])
  return limitedAnswer
}

/** 添加问答额度受限的回答 */
export const useAddLimitedAnswer = (setDialog: Updater<DialogList>) => {
  const limitedAnswer = useLimitedAnswer()
  const addLimitedAnswer = useMemoizedFn((index?: number) => {
    setDialog((prev) => {
      if (typeof index === 'number') {
        prev[index] = limitedAnswer
      } else {
        prev.push(limitedAnswer)
      }
    })
  })
  return addLimitedAnswer
}
