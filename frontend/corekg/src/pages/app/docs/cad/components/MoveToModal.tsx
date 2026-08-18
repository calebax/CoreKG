import { useState, useEffect, useRef } from 'react'
import { Modal, Button, message, Input } from 'antd'
import type { InputRef } from 'antd/es/input'
import dayjs from 'dayjs'
import {
  getFileList,
  createFolder,
  moveFile,
  renameFile,
} from '@/api/knowledge'
import KnowledgeFiles from '@/assets/icons/knowledge-files.svg?react'
import { FileItem } from '../types'

// 文件夹树项接口
interface FolderTreeItem extends FileItem {
  children?: FolderTreeItem[]
  expanded?: boolean
  level: number
}

interface MoveToModalProps {
  visible: boolean
  onCancel: () => void
  knowledgeBaseId: string | undefined
  selectedItem: FileItem | null
  onSuccess: () => void
}

const MoveToModal: React.FC<MoveToModalProps> = ({
  visible,
  onCancel,
  knowledgeBaseId,
  selectedItem,
  onSuccess,
}) => {
  const [folders, setFolders] = useState<FolderTreeItem[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [selectedFolder, setSelectedFolder] = useState<FolderTreeItem | null>(
    null,
  )
  const [editingId, setEditingId] = useState<number | null>(null)
  const editInputRef = useRef<InputRef>(null)
  const [totalFolders, setTotalFolders] = useState<number>(0)

  // 获取文件夹列表
  const fetchFolders = async () => {
    if (!knowledgeBaseId) return

    setLoading(true)
    try {
      // 获取第一层的所有数据
      const res = await getFileList({
        forest_id: Number(knowledgeBaseId),
        limit: 1000,
        offset: 0,
        filters: [
          { field: 'parent_id', value: ['0'] }, // 获取根级别的所有数据
        ],
      })

      if (res && res.data && Array.isArray(res.data)) {
        // 筛选出文件夹（is_dir为true的数据）
        const folderList: FolderTreeItem[] = res.data
          .filter((item: any) => item.is_dir === true)
          .map((item: any) => ({
            id: item.ID,
            name: item.name,
            size: '-',
            updatedAt: item.updated_at || new Date().toISOString(),
            isFolder: true,
            level: 0,
            expanded: false,
          }))
          .reverse()

        setFolders(folderList)
        setTotalFolders(folderList.length)
      } else {
        setFolders([])
        setTotalFolders(0)
      }
    } catch (error) {
      console.error('获取文件夹列表失败', error)
      // message.error('获取文件夹列表失败')
      setFolders([])
      setTotalFolders(0)
    } finally {
      setLoading(false)
    }
  }

  // 获取子文件夹
  const fetchSubFolders = async (
    parentId: number,
    folderIndex: number,
    parentArray: FolderTreeItem[],
  ) => {
    if (!knowledgeBaseId) return

    // 检查是否已达到最大层级（5层）
    if (parentArray[folderIndex].level >= 4) {
      message.warning('已达到最大文件夹嵌套深度（5层）')
      return
    }

    try {
      // 获取指定父文件夹下的所有数据
      const res = await getFileList({
        forest_id: Number(knowledgeBaseId),
        limit: 1000,
        offset: 0,
        filters: [{ field: 'parent_id', value: [parentId.toString()] }],
      })

      if (res && res.data && Array.isArray(res.data)) {
        // 筛选出文件夹（is_dir为true的数据）
        const subFolders: FolderTreeItem[] = res.data
          .filter((item: any) => item.is_dir === true)
          .map((item: any) => ({
            id: item.ID,
            name: item.name,
            size: '-',
            updatedAt: item.updated_at || new Date().toISOString(),
            isFolder: true,
            level: parentArray[folderIndex].level + 1,
            expanded: false,
          }))
          .reverse()

        // 如果没有子文件夹，显示提示
        if (subFolders.length === 0) {
          message.info('该文件夹下没有子文件夹')
          return
        }

        // 更新特定文件夹的子文件夹
        if (parentArray === folders) {
          // 顶层文件夹
          const updatedFolders = [...folders]
          updatedFolders[folderIndex].children = subFolders
          updatedFolders[folderIndex].expanded = true
          setFolders(updatedFolders)
        } else {
          // 嵌套文件夹，使用递归更新
          const updatedFolders = [...folders]
          updateFolderWithChildren(
            updatedFolders,
            parentArray[folderIndex].id,
            subFolders,
          )
          setFolders(updatedFolders)
        }
      } else {
        message.info('该文件夹下没有子文件夹')
      }
    } catch (error) {
      console.error('获取子文件夹失败', error)
      // message.error('获取子文件夹失败')
    }
  }

  // 辅助函数：更新特定文件夹的子文件夹
  const updateFolderWithChildren = (
    folderArray: FolderTreeItem[],
    targetId: number,
    children: FolderTreeItem[],
  ): boolean => {
    for (let i = 0; i < folderArray.length; i++) {
      if (folderArray[i].id === targetId) {
        folderArray[i].children = children
        folderArray[i].expanded = true
        return true
      }

      if (folderArray[i].children) {
        const updated = updateFolderWithChildren(
          folderArray[i].children!,
          targetId,
          children,
        )
        if (updated) return true
      }
    }

    return false
  }

  // 检查文件夹是否可以作为移动目标
  const canMoveToFolder = (folder: FolderTreeItem): boolean => {
    // 如果要移动的不是文件夹，可以移动到任何位置
    if (!selectedItem?.isFolder) {
      return true
    }

    // 如果要移动的是文件夹，检查目标层级
    const targetLevel = folder.level + 1
    return targetLevel <= 4 // 最大层级是4（第5层）
  }

  // 处理文件夹点击
  const handleFolderClick = (
    folder: FolderTreeItem,
    index: number,
    parentArray: FolderTreeItem[],
  ) => {
    // 检查是否可以移动到此文件夹
    if (!canMoveToFolder(folder)) {
      message.warning(
        '无法选择此文件夹：层级过深，移动后将超出最大嵌套深度（5层）',
      )
      return
    }

    // 设置当前选中的文件夹
    setSelectedFolder(folder)

    // 如果已经展开，则折叠
    if (folder.expanded) {
      if (parentArray === folders) {
        // 顶层文件夹
        const updatedFolders = [...folders]
        updatedFolders[index].expanded = false
        setFolders(updatedFolders)
      } else {
        // 嵌套文件夹
        const updatedFolders = [...folders]
        updateFolderTree(updatedFolders, folder.id, { expanded: false })
        setFolders(updatedFolders)
      }
      return
    }

    // 检查是否已达到最大层级（5层）
    if (folder.level >= 4) {
      message.warning('已达到最大文件夹嵌套深度（5层）')
      return
    }

    // 如果还没有子文件夹，则获取
    if (!folder.children) {
      fetchSubFolders(folder.id, index, parentArray)
    } else {
      // 如果已有子文件夹数据
      if (folder.children.length === 0) {
        // 如果子文件夹为空，显示提醒
        message.info('该文件夹下没有子文件夹')
        return
      }

      // 如果有子文件夹，直接展开
      if (parentArray === folders) {
        // 顶层文件夹
        const updatedFolders = [...folders]
        updatedFolders[index].expanded = true
        setFolders(updatedFolders)
      } else {
        // 嵌套文件夹
        const updatedFolders = [...folders]
        updateFolderTree(updatedFolders, folder.id, { expanded: true })
        setFolders(updatedFolders)
      }
    }
  }

  // 递归更新文件夹树
  const updateFolderTree = (
    folderArray: FolderTreeItem[],
    targetId: number,
    update: Partial<FolderTreeItem>,
  ): boolean => {
    for (let i = 0; i < folderArray.length; i++) {
      if (folderArray[i].id === targetId) {
        folderArray[i] = { ...folderArray[i], ...update }
        return true
      }

      if (folderArray[i].children) {
        const updated = updateFolderTree(
          folderArray[i].children!,
          targetId,
          update,
        )
        if (updated) return true
      }
    }

    return false
  }

  // 检查是否可以在当前位置创建文件夹
  const canCreateFolderAtCurrentLocation = (): boolean => {
    const parentLevel = selectedFolder ? selectedFolder.level : -1

    // 基本层级检查：不能超过5层
    if (parentLevel >= 4) {
      return false
    }

    // 如果要移动的是文件夹，还需要检查创建后是否有意义
    if (selectedItem?.isFolder) {
      const newFolderLevel = parentLevel + 1
      // 新建的文件夹层级如果已经是第4层，那移动文件夹进去就会超出限制
      if (newFolderLevel >= 4) {
        return false
      }
    }

    return true
  }

  // 创建新文件夹
  const handleCreateFolder = async () => {
    // 检查是否可以创建文件夹
    if (!canCreateFolderAtCurrentLocation()) {
      if (selectedItem?.isFolder) {
        message.warning(
          '当前位置层级过深，新建文件夹后仍无法移动文件夹到此位置',
        )
      } else {
        message.warning('已达到最大文件夹嵌套深度（5层），无法创建新的文件夹')
      }
      return
    }

    // 确定父文件夹ID：如果没有选择文件夹就是根目录(0)，否则是选择的文件夹ID
    const parentId = selectedFolder ? selectedFolder.id : 0
    const parentLevel = selectedFolder ? selectedFolder.level : -1

    const timestamp = dayjs().format('YYYY-MM-DD HH:mm:ss')
    const newFolderName = `新建文件夹-${timestamp}`

    try {
      // 调用创建文件夹接口
      const createRes = await createFolder({
        forest_id: Number(knowledgeBaseId),
        name: newFolderName,
        parent_id: parentId,
      })

      message.success('文件夹创建成功')

      // 重新获取数据以确保拿到正确的ID
      if (parentId === 0) {
        // 在根目录创建，重新获取根目录数据
        await fetchFolders()

        // 等待状态更新后查找新创建的文件夹
        setTimeout(() => {
          // 直接从更新后的folders中查找
          setFolders((prevFolders) => {
            const newFolder = prevFolders.find((f) => f.name === newFolderName)
            if (newFolder) {
              setSelectedFolder(newFolder)
              setEditingId(newFolder.id)
              // 延迟聚焦到输入框
              setTimeout(() => {
                editInputRef.current?.focus()
                editInputRef.current?.select()
              }, 100)
            }
            return prevFolders
          })
        }, 200)
      } else if (selectedFolder) {
        // 在子文件夹创建，重新获取该子文件夹的数据
        const folderIndex = findFolderIndex(folders, selectedFolder.id)
        if (folderIndex !== -1) {
          await fetchSubFolders(
            selectedFolder.id,
            folderIndex.index,
            folderIndex.parentArray,
          )

          // 等待状态更新后查找新创建的文件夹
          setTimeout(() => {
            setFolders((prevFolders) => {
              const newFolder = findFolderInTree(
                prevFolders,
                newFolderName,
                selectedFolder.id,
              )
              if (newFolder) {
                setSelectedFolder(newFolder)
                setEditingId(newFolder.id)
                // 延迟聚焦到输入框
                setTimeout(() => {
                  editInputRef.current?.focus()
                  editInputRef.current?.select()
                }, 100)
              }
              return prevFolders
            })
          }, 200)
        }
      }
    } catch (error) {
      console.error('新建文件夹失败:', error)
      // message.error('新建文件夹失败')
    }
  }

  // 辅助函数：查找文件夹在树中的位置
  const findFolderIndex = (
    folderArray: FolderTreeItem[],
    targetId: number,
    parentArray?: FolderTreeItem[],
  ): { index: number; parentArray: FolderTreeItem[] } | -1 => {
    for (let i = 0; i < folderArray.length; i++) {
      if (folderArray[i].id === targetId) {
        return { index: i, parentArray: parentArray || folderArray }
      }

      if (folderArray[i].children) {
        const result = findFolderIndex(
          folderArray[i].children!,
          targetId,
          folderArray[i].children!,
        )
        if (result !== -1) return result
      }
    }
    return -1
  }

  // 辅助函数：在树中查找指定名称的文件夹
  const findFolderInTree = (
    folderArray: FolderTreeItem[],
    folderName: string,
    parentId?: number,
  ): FolderTreeItem | null => {
    for (const folder of folderArray) {
      if (folder.name === folderName && (!parentId || folder.level > 0)) {
        return folder
      }

      if (folder.children) {
        const found = findFolderInTree(folder.children, folderName, parentId)
        if (found) return found
      }
    }
    return null
  }

  // 保存编辑的文件夹名称
  const handleSaveEdit = async (folderId: number, newName: string) => {
    if (!newName.trim()) {
      message.warning('文件夹名称不能为空')
      return
    }

    try {
      // 调用重命名接口
      await renameFile({
        file_id: folderId,
        new_name: newName.trim(),
      })

      // 更新本地状态
      const updatedFolders = [...folders]
      updateFolderTree(updatedFolders, folderId, { name: newName.trim() })
      setFolders(updatedFolders)

      setEditingId(null)
      message.success('重命名成功')
    } catch (error) {
      console.error('重命名失败:', error)
      // message.error('重命名失败')
    }
  }

  // 取消编辑
  const handleCancelEdit = () => {
    setEditingId(null)
  }

  // 处理输入框回车和失焦
  const handleInputKeyDown = (
    e: React.KeyboardEvent,
    folderId: number,
    value: string,
  ) => {
    if (e.key === 'Enter') {
      handleSaveEdit(folderId, value)
    } else if (e.key === 'Escape') {
      handleCancelEdit()
    }
  }

  // 移动文件
  const handleMove = async () => {
    if (!selectedItem) {
      message.warning('没有选择要移动的项目')
      return
    }

    // 如果要移动的是文件夹，需要检查层级限制
    if (selectedItem.isFolder) {
      if (selectedFolder) {
        const targetLevel = selectedFolder.level + 1 // 移动后的层级

        // 检查移动后是否会超出最大层级限制（5层，索引0-4）
        if (targetLevel > 4) {
          message.warning(
            '无法移动到此位置：目标文件夹层级过深，移动后将超出最大嵌套深度（5层）',
          )
          return
        }
      }
      // 移动到根目录时，层级为0，始终允许
    }

    setLoading(true)
    try {
      // console.log('移动文件:', selectedItem.id, '到文件夹:', selectedFolder.id)

      // 调用移动文件的接口
      await moveFile({
        file_id: selectedItem.id,
        dst_parent_id: selectedFolder ? selectedFolder.id : 0,
      })

      message.success('移动成功')
      onSuccess()
      onCancel()
    } catch (error) {
      console.error('移动文件失败', error)
      // message.error('移动文件失败')
    } finally {
      setLoading(false)
    }
  }

  // 组件加载时获取文件夹列表
  useEffect(() => {
    if (visible && knowledgeBaseId) {
      setFolders([])
      setSelectedFolder(null)
      setEditingId(null)
      fetchFolders()
    }
  }, [visible, knowledgeBaseId])

  // 扁平化文件夹树为线性列表
  const flattenFolderTree = (
    folderArray: FolderTreeItem[],
  ): FolderTreeItem[] => {
    const result: FolderTreeItem[] = []

    const traverse = (folders: FolderTreeItem[]) => {
      folders.forEach((folder) => {
        result.push(folder)
        if (folder.children && folder.expanded) {
          traverse(folder.children)
        }
      })
    }

    traverse(folderArray)
    return result
  }

  // 渲染文件夹树
  const renderFolderTree = (folderArray: FolderTreeItem[]) => {
    const flatFolders = flattenFolderTree(folderArray)

    return flatFolders.map((folder, index) => {
      const canMove = canMoveToFolder(folder)
      const isSelected = selectedFolder?.id === folder.id

      // 找到原始数组中的位置信息
      const findFolderInOriginalArray = (
        targetId: number,
        searchArray: FolderTreeItem[],
        parentArray?: FolderTreeItem[],
      ): {
        folder: FolderTreeItem
        index: number
        parentArray: FolderTreeItem[]
      } | null => {
        for (let i = 0; i < searchArray.length; i++) {
          if (searchArray[i].id === targetId) {
            return {
              folder: searchArray[i],
              index: i,
              parentArray: parentArray || searchArray,
            }
          }
          if (searchArray[i].children && searchArray[i].expanded) {
            const found = findFolderInOriginalArray(
              targetId,
              searchArray[i].children!,
              searchArray[i].children!,
            )
            if (found) return found
          }
        }
        return null
      }

      const originalInfo = findFolderInOriginalArray(folder.id, folderArray)

      return (
        <div key={`folder-item-${folder.id}`}>
          {/* 文件夹行 */}
          <div
            className={`flex items-center py-4 px-3 rounded-md ${canMove ? `cursor-pointer hover:bg-[#E8F3FF] ${isSelected ? 'bg-[#E8F3FF]' : ''}` : 'cursor-not-allowed opacity-50 bg-gray-50'}`}
            style={{ paddingLeft: `${folder.level * 20 + 12}px` }}
            onClick={() =>
              originalInfo &&
              handleFolderClick(
                originalInfo.folder,
                originalInfo.index,
                originalInfo.parentArray,
              )
            }
          >
            <KnowledgeFiles
              className={`w-5 h-5 mr-2 ${canMove ? 'text-[#165DFF]' : 'text-gray-400'}`}
            />
            {editingId === folder.id ? (
              <Input
                ref={editInputRef}
                defaultValue={folder.name}
                className='flex-1 !border-none !outline-none !shadow-none'
                size='small'
                onBlur={(e) => handleSaveEdit(folder.id, e.target.value)}
                onKeyDown={(e) =>
                  handleInputKeyDown(e, folder.id, e.currentTarget.value)
                }
                onClick={(e) => e.stopPropagation()}
              />
            ) : (
              <span
                className={`flex-1 truncate font-normal text-base ${canMove ? 'text-[#4E5969]' : 'text-gray-400'}`}
              >
                {folder.name}
                {!canMove && selectedItem?.isFolder && (
                  <span className='ml-2 text-xs text-gray-400'>(层级过深)</span>
                )}
              </span>
            )}
          </div>

          {/* 分割线 - 占满整行宽度，不受缩进影响 */}
          {index < flatFolders.length - 1 && (
            <div className='h-[1px] w-full bg-[#E5E6EB] my-2'></div>
          )}
        </div>
      )
    })
  }

  return (
    <Modal
      title={<div className='text-lg font-medium text-[#1D2129]'>移动到</div>}
      open={visible}
      onCancel={onCancel}
      footer={null}
      width={600}
      destroyOnHidden
      className='move-to-modal'
      centered
      closable={false}
    >
      <div>
        {/* 提示信息区域 */}
        <div className='bg-[#F7F8FA] rounded-lg p-4 mt-4 mb-4'>
          <div className='flex items-center gap-2'>
            <div className='w-2 h-2 bg-[#165DFF] rounded-full'></div>
            <span className='text-[#4E5969] text-sm'>
              {selectedFolder ? (
                <>
                  将移动到{' '}
                  <span className='font-medium text-[#1D2129]'>
                    "{selectedFolder.name}"
                  </span>{' '}
                  文件夹
                </>
              ) : (
                <>
                  将移动到{' '}
                  <span className='font-medium text-[#1D2129]'>"根目录"</span>
                  ，请选择目标文件夹或直接移动
                </>
              )}
            </span>
          </div>
          {selectedItem?.isFolder && (
            <div className='flex items-center gap-2 mt-2'>
              <div className='w-2 h-2 bg-[#FF7D00] rounded-full'></div>
              <span className='text-[#4E5969] text-xs'>
                注意：移动文件夹时需确保不超过最大嵌套深度（5层），灰色文件夹表示层级过深无法选择
                {!canCreateFolderAtCurrentLocation() &&
                  '，当前位置无法新建文件夹'}
              </span>
            </div>
          )}
        </div>

        {/* 分割线 */}
        {/* <div className="h-[1px] w-full bg-[#E5E6EB] mb-2"></div> */}

        {/* 自定义滚动条样式 */}
        <style
          dangerouslySetInnerHTML={{
            __html: `
          .move-to-modal .folder-container::-webkit-scrollbar {
            width: 6px;
          }
          .move-to-modal .folder-container::-webkit-scrollbar-track {
            background: #f1f1f1;
            border-radius: 3px;
          }
          .move-to-modal .folder-container::-webkit-scrollbar-thumb {
            background: #ccc;
            border-radius: 3px;
          }
          .move-to-modal .folder-container::-webkit-scrollbar-thumb:hover {
            background: #a5a5a5;
          }
        `,
          }}
        />

        {/* 文件夹列表区域 */}
        <div className='rounded h-[400px] overflow-y-auto mb-6 folder-container pr-3'>
          {loading ? (
            <div className='flex items-center justify-center h-full'>
              <span className='text-gray-400'>加载中...</span>
            </div>
          ) : folders.length > 0 ? (
            <div>{renderFolderTree(folders)}</div>
          ) : (
            <div className='flex items-center justify-center h-full'>
              <span className='text-gray-400'>暂无文件夹</span>
            </div>
          )}
        </div>

        {/* 底部按钮 */}
        <div className='flex justify-between border-t border-[#E5E6EB] pt-4'>
          <Button
            className={`flex items-center gap-1 !px-6 !py-4 !h-13 !border-none !rounded-full !font-medium !text-lg ${canCreateFolderAtCurrentLocation() ? '!bg-[#E8F3FF] hover:!bg-[#E8F3FF] !text-[#165DFF]' : '!bg-gray-100 !text-gray-400 cursor-not-allowed'}`}
            disabled={!canCreateFolderAtCurrentLocation()}
            onClick={handleCreateFolder}
          >
            <span>新建文件夹</span>
          </Button>

          <div className='flex gap-3'>
            {/* 取消 */}
            <Button
              className='!h-13 !px-10 !py-4 !bg-[#E8F3FF] hover:!bg-[#E8F3FF] !border-none !rounded-lg !text-[#165DFF] !font-medium !text-lg'
              onClick={onCancel}
            >
              取消
            </Button>

            {/* 移动至此 */}
            <Button
              type='primary'
              className='!h-13 !px-8 !py-4 !bg-[#165DFF] hover:!bg-[#165DFF] !rounded-lg !min-w-[88px] !text-lg !font-medium !text-white'
              onClick={handleMove}
              loading={loading}
            >
              移动至此
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  )
}

export default MoveToModal
