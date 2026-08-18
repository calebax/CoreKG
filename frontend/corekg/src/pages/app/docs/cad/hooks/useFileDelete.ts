import { useState, useCallback } from 'react'
import { message } from 'antd'
import { deleteFile } from '@/api/knowledge'
import { FileItem } from '../types'

interface UseFileDeleteProps {
  files: FileItem[]
  setFiles: (files: FileItem[]) => void
  selectedRowKeys: React.Key[]
  setSelectedRowKeys: (keys: React.Key[]) => void
  onRefresh: (page?: number) => void
  saveTableParams: (params: any) => void
  currentTableParams: any
  setCurrentPage: (page: number) => void
  total: number
}

interface UseFileDeleteReturn {
  deleteModalVisible: boolean
  deleteModalConfig: {
    isFolder: boolean
    isMultiple: boolean
    itemId: number
  }
  setDeleteModalVisible: (visible: boolean) => void
  handleDelete: (id: number) => void
  handleBatchDelete: () => void
  handleConfirmSingleDelete: () => Promise<void>
  handleConfirmBatchDelete: () => Promise<void>
}

export const useFileDelete = ({
  files,
  setFiles,
  selectedRowKeys,
  setSelectedRowKeys,
  onRefresh,
  saveTableParams,
  currentTableParams,
  setCurrentPage,
  total,
}: UseFileDeleteProps): UseFileDeleteReturn => {
  // 删除对话框状态
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [deleteModalConfig, setDeleteModalConfig] = useState({
    isFolder: false,
    isMultiple: false,
    itemId: 0,
  })

  // 处理单个删除
  const handleDelete = useCallback(
    (id: number) => {
      // 查找要删除的项目
      const itemToDelete = files.find((item) => item.id === id)
      const isFolder = itemToDelete?.isFolder

      // 设置删除对话框配置
      setDeleteModalConfig({
        isFolder: !!isFolder,
        isMultiple: false,
        itemId: id,
      })

      // 显示删除对话框
      setDeleteModalVisible(true)
    },
    [files],
  )

  // 执行单个删除
  const handleConfirmSingleDelete = useCallback(async () => {
    const id = deleteModalConfig.itemId

    try {
      // 调用删除文件的接口
      await deleteFile({ file_id: [Number(id)] })

      // 更新文件列表
      const newFiles = files.filter((item) => item.id !== id)
      setFiles(newFiles)

      // 同时从选中项中移除
      const newSelectedKeys = selectedRowKeys.filter((key) => key !== id)
      setSelectedRowKeys(newSelectedKeys)

      // 判断删除后应该跳转到哪一页
      const { currentPage, pageSize } = currentTableParams
      const deletedCount = 1 // 单个删除
      const newTotal = total - deletedCount

      // 计算删除后的最大页数
      const maxPage = Math.max(1, Math.ceil(newTotal / pageSize))

      // 如果当前页超出了最大页数，则跳转到最大页
      let targetPage = currentPage
      if (currentPage > maxPage) {
        targetPage = maxPage
        // 同时更新组件状态和本地存储
        setCurrentPage(targetPage)
        saveTableParams({
          ...currentTableParams,
          currentPage: targetPage,
          selectedRowKeys: newSelectedKeys,
        })
      } else {
        // 否则保持当前页码，只更新选中状态
        saveTableParams({
          ...currentTableParams,
          selectedRowKeys: newSelectedKeys,
        })
      }

      message.success(
        deleteModalConfig.isFolder ? '删除文件夹成功' : '删除文件成功',
      )

      // 关闭对话框
      setDeleteModalVisible(false)

      // 刷新数据列表，使用正确的页码
      onRefresh(targetPage)
    } catch (error) {
      console.error('删除失败:', error)
      message.error('删除失败，请重试')
    }
  }, [
    deleteModalConfig,
    files,
    setFiles,
    selectedRowKeys,
    setSelectedRowKeys,
    saveTableParams,
    currentTableParams,
    onRefresh,
    setCurrentPage,
    total,
  ])

  // 处理批量删除
  const handleBatchDelete = useCallback(() => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要删除的项目')
      return
    }

    // 检查选中的项目中是否包含文件夹
    const selectedItems = files.filter((item) =>
      selectedRowKeys.includes(item.id),
    )
    const hasFolders = selectedItems.some((item) => item.isFolder)
    const hasFiles = selectedItems.some((item) => !item.isFolder)

    // 设置删除对话框配置
    setDeleteModalConfig({
      isFolder: hasFolders && !hasFiles,
      isMultiple: true,
      itemId: 0,
    })

    // 显示删除对话框
    setDeleteModalVisible(true)
  }, [selectedRowKeys, files])

  // 执行批量删除
  const handleConfirmBatchDelete = useCallback(async () => {
    try {
      // 将selectedRowKeys转换为number数组
      const fileIds = selectedRowKeys.map((key) => Number(key))

      // 调用批量删除文件的接口
      await deleteFile({ file_id: fileIds })

      // 更新文件列表
      const newFiles = files.filter(
        (item) => !selectedRowKeys.includes(item.id),
      )
      setFiles(newFiles)

      // 判断删除后应该跳转到哪一页
      const { currentPage, pageSize } = currentTableParams
      const deletedCount = selectedRowKeys.length // 批量删除的数量
      const newTotal = total - deletedCount

      // 计算删除后的最大页数
      const maxPage = Math.max(1, Math.ceil(newTotal / pageSize))

      // 如果当前页超出了最大页数，则跳转到最大页
      let targetPage = currentPage
      if (currentPage > maxPage) {
        targetPage = maxPage
        // 同时更新组件状态和本地存储
        setCurrentPage(targetPage)
        saveTableParams({
          ...currentTableParams,
          currentPage: targetPage,
          selectedRowKeys: [],
        })
      } else {
        // 否则保持当前页码，清空选中状态
        saveTableParams({
          ...currentTableParams,
          selectedRowKeys: [],
        })
      }

      // 构造成功消息
      let successMessage = '删除成功'
      if (deleteModalConfig.isFolder) {
        successMessage = '删除文件夹成功'
      } else if (!deleteModalConfig.isFolder && !deleteModalConfig.isMultiple) {
        successMessage = '删除文件成功'
      }

      message.success(successMessage)

      // 清空选中状态
      setSelectedRowKeys([])

      // 关闭对话框
      setDeleteModalVisible(false)

      // 刷新数据列表，使用正确的页码
      onRefresh(targetPage)
    } catch (error) {
      console.error('批量删除失败:', error)
      message.error('删除失败，请重试')
    }
  }, [
    selectedRowKeys,
    files,
    setFiles,
    deleteModalConfig,
    setSelectedRowKeys,
    saveTableParams,
    currentTableParams,
    onRefresh,
    setCurrentPage,
    total,
  ])

  return {
    deleteModalVisible,
    deleteModalConfig,
    setDeleteModalVisible,
    handleDelete,
    handleBatchDelete,
    handleConfirmSingleDelete,
    handleConfirmBatchDelete,
  }
}
