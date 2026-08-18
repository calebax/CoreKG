import { useEffect } from 'react'
import { STORAGE_KEY } from '../types'

interface UseStorageCleanupProps {
  knowledgeBaseId?: string | number
  enabled?: boolean
}

export const useStorageCleanup = (props: UseStorageCleanupProps = {}) => {
  // 清除所有知识库相关的存储数据
  const clearAllKnowledgeBaseStorage = () => {
    try {
      const localStorageData = localStorage.getItem(STORAGE_KEY)
      if (!localStorageData) return

      const data = JSON.parse(localStorageData)
      const updatedData = { ...data }

      // 清除所有知识库相关的存储键
      Object.keys(updatedData).forEach((key) => {
        // if (key === 'knowledgeDetailTable' || key.startsWith('folderDetail-')) {
        //   delete updatedData[key]
        // }
      })

      localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedData))
    } catch (error) {
      console.error('清除知识库存储失败', error)
    }
  }

  // 监听页面卸载事件，在用户关闭浏览器或刷新页面时清理存储
  useEffect(() => {
    if (!props.enabled) return

    const handleBeforeUnload = () => {
      if (props.knowledgeBaseId) {
        clearAllKnowledgeBaseStorage()
      }
    }

    // 监听页面卸载事件
    window.addEventListener('beforeunload', handleBeforeUnload)

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [props.knowledgeBaseId, props.enabled])

  return {
    clearAllKnowledgeBaseStorage,
  }
}
