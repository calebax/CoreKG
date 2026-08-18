import { useState, useRef } from 'react'
import { message } from 'antd'
import type { InputRef } from 'antd/es/input'
import dayjs from 'dayjs'
import { spliteFileName } from '@/utils'
import { createFolder, renameFile } from '@/api/knowledge'
import { FileItem } from '../types'

interface UseFileEditProps {
  files: FileItem[]
  setFiles: (files: FileItem[]) => void
  knowledgeBaseId: number
  parentId: number
  currentLevel: number
  onSuccess?: () => void
}

export const useFileEdit = ({
  files,
  setFiles,
  knowledgeBaseId,
  parentId,
  currentLevel,
  onSuccess,
}: UseFileEditProps) => {
  const [editingId, setEditingId] = useState<number | null>(null)
  const editInputRef = useRef<InputRef>(null)

  // 开始编辑文件名
  const startEditing = (id: number) => {
    setEditingId(id)
    // 下一帧聚焦到输入框
    setTimeout(() => {
      editInputRef.current?.focus()
      editInputRef.current?.select()
    }, 0)
  }

  // 保存编辑后的文件名
  const saveEditing = async (fileId: number, newName: string) => {
    if (!newName.trim()) {
      message.error('文件名不能为空')
      return
    }

    // 检查文件名是否已存在（排除当前编辑的文件）
    const hasSameNameFile = files.some(
      (file) => file.name === newName && file.id !== fileId,
    )
    if (hasSameNameFile) {
      message.error('文件名已存在')
      return
    }
    const { name } = spliteFileName(newName)
    if (name.length > 50) {
      message.error('文件名不能超过50个字符')
      return
    }

    setEditingId(null)

    // 查找当前编辑的文件并保存原始名称
    const currentFileToUpdate = files.find((file) => file.id === fileId)
    const isNewFolderAction = currentFileToUpdate?.isNewFolder
    const originalName = currentFileToUpdate?.name || ''

    // 先更新本地状态
    setFiles(
      files.map((file) =>
        file.id === fileId
          ? { ...file, name: newName, isNewFolder: false }
          : file,
      ),
    )

    // 调用重命名或新建文件夹接口
    if (isNewFolderAction) {
      try {
        await createFolder({
          forest_id: knowledgeBaseId,
          name: newName,
          parent_id: parentId,
        })
        message.success('新建文件夹成功')

        // 执行成功回调
        onSuccess?.()
      } catch (error) {
        console.error('新建文件夹失败:', error)

        // 创建失败时，从列表中移除临时添加的文件夹
        setFiles(files.filter((file) => file.id !== fileId))
        return
      }
    } else {
      // 调用重命名文件/文件夹的接口
      try {
        await renameFile({ file_id: fileId, new_name: newName })
        message.success('重命名成功')
      } catch (error) {
        console.error('重命名失败:', error)

        // 重命名失败时，恢复原始文件名
        setFiles(
          files.map((file) =>
            file.id === fileId ? { ...file, name: originalName } : file,
          ),
        )
        return
      }
    }
  }

  // 创建新文件夹
  const createNewFolder = () => {
    const timestamp = dayjs().format('YYYY-MM-DD HH:mm:ss')
    const newFolderName = `新建文件夹-${timestamp}`

    // 创建新文件夹项
    const newFolder: FileItem = {
      id: Date.now(), // 临时ID，后端返回后会替换
      name: newFolderName,
      size: '-',
      updatedAt: dayjs().format('YYYY-MM-DD'),
      createdAtFull: dayjs().toISOString(), // 完整的时间戳，用于排序
      isFolder: true,
      isEditing: true,
      isNewFolder: true, // 标记为新创建的文件夹
    }

    // 将新文件夹添加到列表顶部
    setFiles([newFolder, ...files])

    // 设置为编辑状态
    setEditingId(newFolder.id)

    // 下一帧聚焦到输入框
    setTimeout(() => {
      editInputRef.current?.focus()
      editInputRef.current?.select()
    }, 0)
  }

  return {
    editingId,
    editInputRef,
    startEditing,
    saveEditing,
    createNewFolder,
  }
}
