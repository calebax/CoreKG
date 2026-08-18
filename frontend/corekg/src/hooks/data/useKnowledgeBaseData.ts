import { useMemo } from 'react'
import { useMount } from 'ahooks'
import { getKnowledgeBaseList } from '@/api/knowledge'
import useDataStore from '@/stores/data'

export default function useKnowledgeBaseData() {
  const { knowledgeBaseList, setKnowledgeBaseList } = useDataStore()

  const loadKnowledgeBaseList = async () => {
    try {
      const res = await getKnowledgeBaseList()
      const list = res.Data || []
      setKnowledgeBaseList(
        list.map((x) => ({
          value: x.ID,
          label: x.name,
        })),
      )
      return list
    } catch {
      return []
    }
  }

  const refreshKnowledgeBaseList = async () => {
    await loadKnowledgeBaseList()
  }

  // 使用 useMemo 优化性能
  const getKnowledgeBaseNameById = useMemo(() => {
    return (id: number) => {
      const item = knowledgeBaseList.find((x) => x.value === id)
      return item?.label || ''
    }
  }, [knowledgeBaseList])

  const getKnowledgeBasesByIds = useMemo(() => {
    return (ids: number[]) => {
      return knowledgeBaseList.filter((x) => ids.includes(x.value))
    }
  }, [knowledgeBaseList])

  // 根据当前选中的 IDs 字符串获取对应的知识库列表（常用场景）
  const getKnowledgeBasesByIdsString = useMemo(() => {
    return (idsString: string | null) => {
      if (!idsString) return []
      const ids = idsString.split(',').map(Number).filter(Boolean)
      return getKnowledgeBasesByIds(ids)
    }
  }, [getKnowledgeBasesByIds])

  useMount(() => {
    if (knowledgeBaseList.length === 0) {
      loadKnowledgeBaseList()
    }
  })

  return {
    knowledgeBaseList,
    loadKnowledgeBaseList,
    refreshKnowledgeBaseList,
    getKnowledgeBaseNameById,
    getKnowledgeBasesByIds,
    getKnowledgeBasesByIdsString,
  }
}
