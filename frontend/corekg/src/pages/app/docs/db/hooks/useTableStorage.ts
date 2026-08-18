import { useCallback } from 'react'
import { TableParams, STORAGE_KEY } from '../types'

interface UseTableStorageProps {
  storageKey: string
}

export const useTableStorage = ({ storageKey }: UseTableStorageProps) => {
  // 从本地存储获取初始表格参数
  const getInitialTableParams = useCallback((): TableParams => {
    try {
      const localStorageData = localStorage.getItem(STORAGE_KEY)
      const data = localStorageData ? JSON.parse(localStorageData) : {}

      if (data[storageKey]) {
        return data[storageKey]
      }

      return {
        currentPage: 1,
        pageSize: 10,
        sorts: [],
        selectedRowKeys: [],
      }
    } catch (error) {
      console.error('读取本地存储失败', error)
      return {
        currentPage: 1,
        pageSize: 10,
        sorts: [],
        selectedRowKeys: [],
      }
    }
  }, [storageKey])

  // 保存表格参数到本地存储
  const saveTableParamsToStorage = useCallback(
    (params: TableParams) => {
      try {
        const localStorageData = localStorage.getItem(STORAGE_KEY)
        const data = localStorageData ? JSON.parse(localStorageData) : {}

        const updatedData = {
          ...data,
          [storageKey]: params,
        }

        localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedData))
      } catch (error) {
        console.error('保存表格参数到本地存储失败', error)
      }
    },
    [storageKey],
  )

  // 清除当前存储键的数据
  const clearCurrentStorage = useCallback(() => {
    try {
      const localStorageData = localStorage.getItem(STORAGE_KEY)
      if (!localStorageData) return

      const data = JSON.parse(localStorageData)
      const updatedData = { ...data }

      if (updatedData[storageKey]) {
        delete updatedData[storageKey]
      }

      localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedData))
    } catch (error) {
      console.error('清除当前存储失败', error)
    }
  }, [storageKey])

  return {
    getInitialTableParams,
    saveTableParamsToStorage,
    clearCurrentStorage,
  }
}
