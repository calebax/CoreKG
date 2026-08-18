import { useState, useRef } from 'react'
import { message } from 'antd'
import type { InputRef } from 'antd/es/input'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
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
  const { t } = useTranslation('pages')
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
  const saveEditing = async (
    fileId: number,
    newName: string,
    ext: string = '',
  ) => {
    if (!newName.trim()) {
      message.error(t('app.docs.detail.fileEdit.nameRequired'))
      return
    }

    // 检查文件名是否已存在（排除当前编辑的文件）
    const hasSameNameFile = files.some(
      (file) => file.name === newName && file.id !== fileId,
    )
    if (hasSameNameFile) {
      message.error(t('app.docs.detail.fileEdit.nameExists'))
      return
    }
    const { name } = spliteFileName(newName)
    if (name.length > 50) {
      message.error(t('app.docs.detail.fileEdit.nameTooLong', { count: 50 }))
      return
    }

    // 查找当前编辑的文件并保存原始名称
    const currentFileToUpdate = files.find((file) => file.id === fileId)
    const isNewFolderAction = currentFileToUpdate?.isNewFolder
    const originalName = currentFileToUpdate?.name || ''

    // 如果名称未变化且不是新建文件夹，直接退出编辑态，不调用后端
    if (!isNewFolderAction && newName.trim() === originalName.trim()) {
      setEditingId(null)
      return
    }

    setEditingId(null)

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
          name: newName + ext,
          parent_id: parentId,
        })
        message.success(t('app.docs.detail.fileEdit.createFolderSuccess'))

        // 执行成功回调
        onSuccess?.()
      } catch (error) {
        console.error('新建文件夹失败:', error)

        // 创建失败时，从列表中移除临时添加的文件夹
        setFiles(files.filter((file) => file.id !== fileId))
        message.error(t('app.docs.detail.fileEdit.createFolderFail'))
        return
      }
    } else {
      // 调用重命名文件/文件夹的接口
      try {
        await renameFile({ file_id: fileId, new_name: newName + ext })
        message.success(t('app.docs.detail.fileEdit.renameSuccess'))
      } catch (error) {
        console.error('重命名失败:', error)

        // 重命名失败时，恢复原始文件名
        setFiles(
          files.map((file) =>
            file.id === fileId ? { ...file, name: originalName } : file,
          ),
        )
        message.error(t('app.docs.detail.fileEdit.renameFail'))
        return
      }
    }
  }

  // 创建新文件夹
  const createNewFolder = () => {
    const timestamp = dayjs().format('YYYY-MM-DD HH:mm:ss')
    const newFolderName = t('app.docs.detail.fileEdit.newFolderName', {
      timestamp,
    })

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
      file_status: 'pending',
      file_progress: '0%',
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
