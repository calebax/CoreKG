import { useState, useEffect, useCallback, useMemo } from 'react'
import { FileItem } from '../types'

interface UseRowSelectionProps {
  files: FileItem[]
  initialSelectedKeys?: React.Key[]
  onSelectionChange?: (keys: React.Key[]) => void
}

interface UseRowSelectionReturn {
  selectedRowKeys: React.Key[]
  globalSelectedKeys: React.Key[]
  setSelectedRowKeys: (keys: React.Key[]) => void
  rowSelection: any
}

export const useRowSelection = ({
  files,
  initialSelectedKeys = [],
  onSelectionChange,
}: UseRowSelectionProps): UseRowSelectionReturn => {
  // 全局选中状态（跨页面）
  const [globalSelectedKeys, setGlobalSelectedKeys] =
    useState<React.Key[]>(initialSelectedKeys)

  // 当前页面的选中状态（仅用于表格显示）
  const selectedRowKeys = useMemo(() => {
    const currentPageIds = files.map((item) => item.id)
    return globalSelectedKeys.filter((key) =>
      currentPageIds.includes(Number(key)),
    )
  }, [files, globalSelectedKeys])

  // 当初始选中状态变化时更新全局状态
  useEffect(() => {
    const currentKeysStr = JSON.stringify([...globalSelectedKeys].sort())
    const initialKeysStr = JSON.stringify([...initialSelectedKeys].sort())

    if (currentKeysStr !== initialKeysStr) {
      setGlobalSelectedKeys(initialSelectedKeys)
    }
  }, [initialSelectedKeys, globalSelectedKeys])

  // 处理选中状态变化
  const handleSelectionChange = useCallback(
    (selectedKeys: React.Key[]) => {
      // 获取当前页面的所有文件ID
      const currentPageIds = files.map((item) => item.id)

      // 从全局选中状态中移除当前页面的所有ID，然后添加新选中的ID
      let newGlobalSelectedKeys = globalSelectedKeys.filter(
        (key) => !currentPageIds.includes(Number(key)),
      )

      // 添加当前页新选中的ID
      newGlobalSelectedKeys = [...newGlobalSelectedKeys, ...selectedKeys]

      // 保存新的全局选中状态
      setGlobalSelectedKeys(newGlobalSelectedKeys)

      // 调用外部回调
      onSelectionChange?.(newGlobalSelectedKeys)
    },
    [files, globalSelectedKeys, onSelectionChange],
  )

  // 行选择配置
  const rowSelection = useMemo(
    () => ({
      selectedRowKeys, // 只包含当前页面的选中项
      onChange: handleSelectionChange,
      checkStrictly: false,
      getCheckboxProps: (record: FileItem) => ({
        name: record.name,
      }),
      columnWidth: 60,
    }),
    [selectedRowKeys, handleSelectionChange],
  )

  // 提供设置全局选中状态的方法
  const setSelectedRowKeysFunc = useCallback((keys: React.Key[]) => {
    setGlobalSelectedKeys(keys)
  }, [])

  return {
    selectedRowKeys, // 当前页面的选中项（用于表格显示）
    globalSelectedKeys, // 全局选中项（用于跨页面状态管理）
    setSelectedRowKeys: setSelectedRowKeysFunc,
    rowSelection,
  }
}
