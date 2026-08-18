import { useState, useMemo } from 'react'
import type { AIDialog } from '@/components/dialog'

export function useAiDialog(props: { value: AIDialog }) {
  const { thinkingContent, reference, content } = props.value

  const [thinkingVisible, setThinkingVisible] = useState<boolean | null>(null)
  const showThinking = useMemo(() => {
    if (thinkingVisible === null) return thinkingContent && !content
    return thinkingVisible
  }, [thinkingContent, content, thinkingVisible])

  // 深度思考
  const thinkingContentValue = useMemo(() => {
    return thinkingContent.replace(/\{Reference.*?\}/g, '')
  }, [thinkingContent])

  // 搜索到的引用标题
  const referenceText = useMemo(() => {
    if (!reference?.length) return ''
    const fileSet = new Set<number>()
    const forestSet = new Set<number>()
    reference.forEach((item) => {
      fileSet.add(item.file_id)
      forestSet.add(item.forest_id)
    })

    return (
      (forestSet.size ? '搜索到' + forestSet.size + '个知识库' : '') +
      (fileSet.size ? fileSet.size + '篇资源' : '')
    )
  }, [reference])

  return {
    referenceText,
    thinkingContentValue,
    showThinking,
    thinkingVisible,
    setThinkingVisible,
  }
}
