import { useState, useEffect, useCallback, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { message, Modal } from 'antd'
import { Checkbox } from 'antd'
import TableWithPagination from '@/components/common/TableWithPagination'
// 导入自定义hooks
import {
  // 文件上传
  useFileUpload,
  // 轮询
  usePolling,
  // 表格存储
  useTableStorage,
  // 存储清理
  useStorageCleanup,
  // 文件编辑
  useFileEdit,
  // 文件列表
  useFileList,
  // 文件删除
  useFileDelete,
  // 文件导航
  useFileNavigation,
  // 行选择
  useRowSelection,
} from '../hooks'
// 导入提取的工具和组件

// 导入类型
import { FileItem, SortItem, STORAGE_KEY } from '../types'
// 导入表格列
import { getColumns } from '../utils/tableColumns'
// 导入操作按钮
import ActionButtons from './ActionButtons'
// 导入面包屑导航
import BreadcrumbNav from './BreadcrumbNav'
// 导入删除确认对话框
import DeleteConfirmModal from './DeleteConfirmModal'
// 导入移动到对话框
import MoveToModal from './MoveToModal'

// 定义轮询间隔时间（毫秒）
const POLLING_INTERVAL = 5000 // 5秒钟

interface PathItem {
  id: number
  name: string
  level: number
}

interface FileExplorerProps {
  isRootLevel: boolean
}

export default function FileExplorer({ isRootLevel }: FileExplorerProps) {
  const { id, folderId } = useParams<{ id: string; folderId: string }>()
  const navigate = useNavigate()

  // 解析路径数据 - 仅在非根级别时使用
  const [folderInfo, setFolderInfo] = useState<{
    name: string
    level: number
    path: PathItem[] // 路径记录
  }>({ name: '', level: 0, path: [] })

  const [searchKeyword, setSearchKeyword] = useState<string>('')

  // 移动文件对话框状态
  const [moveModalVisible, setMoveModalVisible] = useState(false)
  const [itemToMove, setItemToMove] = useState<FileItem | null>(null)

  const [featureModalVisible, setFeatureModalVisible] = useState(false)

  // 本地存储键（区分不同层级）
  const getStorageKey = () => {
    if (isRootLevel) {
      return 'knowledgeDetailTable'
    } else {
      return `folderDetail-${folderId}`
    }
  }

  // 读取存储参数的辅助函数
  const getStorageParams = (storageKey: string) => {
    try {
      const localStorageData = localStorage.getItem(STORAGE_KEY)
      const data = localStorageData ? JSON.parse(localStorageData) : {}

      return (
        data[storageKey] || {
          currentPage: 1,
          pageSize: 10,
          sorts: [],
          selectedRowKeys: [],
        }
      )
    } catch (error) {
      console.error('读取存储参数失败:', error)
      return {
        currentPage: 1,
        pageSize: 10,
        sorts: [],
        selectedRowKeys: [],
      }
    }
  }

  // 使用本地存储hook
  const storageKey = getStorageKey()
  const { getInitialTableParams, saveTableParamsToStorage } = useTableStorage({
    storageKey,
  })
  const initialParams = getInitialTableParams()

  // 使用存储清理hook
  const { clearAllKnowledgeBaseStorage } = useStorageCleanup()

  const [currentPage, setCurrentPage] = useState<number>(
    initialParams.currentPage,
  )
  const [pageSize, setPageSize] = useState<number>(initialParams.pageSize)
  const [sortItems, setSortItems] = useState<SortItem[]>(
    initialParams.sorts || [],
  )

  const pageSizeOptions = [10, 20, 50, 100]

  // 使用文件导航hook
  const { getPathFromQuery, handleFolderClick, handleFileClick } =
    useFileNavigation({
      knowledgeBaseId: id!,
      folderId,
      isRootLevel,
      folderInfo,
      knowledgeBaseName: '',
    })

  // 使用文件列表hook
  const {
    isAdmin,
    isSystemKnowledgeBase,
    files,
    setFiles,
    total,
    loading,
    filters,
    setFilters,
    knowledgeBaseName,
    fetchFileList,
    fetchKnowledgeBaseInfo,
    fetchFolderDetail,
    handleSearch: handleListSearch,
  } = useFileList({
    knowledgeBaseId: Number(id),
    parentId: isRootLevel ? 0 : Number(folderId),
    isRootLevel,
    currentPage,
    pageSize,
    sortItems,
    getPathFromQuery,
    folderInfo,
    setFolderInfo,
  })

  // 优化选择变化回调函数，避免不必要的重新创建
  const handleSelectionChange = useCallback(
    (keys: React.Key[]) => {
      // 立即更新本地存储，确保选中状态被正确保存
      saveTableParamsToStorage({
        currentPage,
        pageSize,
        sorts: sortItems,
        selectedRowKeys: keys,
      })
    },
    [currentPage, pageSize, sortItems, saveTableParamsToStorage],
  )

  // 使用行选择hook
  const {
    selectedRowKeys,
    globalSelectedKeys,
    setSelectedRowKeys,
    rowSelection: baseRowSelection,
  } = useRowSelection({
    files,
    initialSelectedKeys: initialParams.selectedRowKeys || [],
    onSelectionChange: handleSelectionChange,
  })

  // 创建带有自定义表头的行选择配置
  const rowSelection = useMemo(
    () => ({
      ...baseRowSelection,
      // 控制表头全选框的状态
      columnTitle: () => {
        // 计算当前页面数据
        const currentPageIds = files.map((item) => item.id)

        // 检查当前页是否有被选中的行
        const hasSelected = globalSelectedKeys.some((key: React.Key) =>
          currentPageIds.includes(Number(key)),
        )

        // 跨页选择：有选中项且不全是当前页的
        const hasCrossPageSelection =
          globalSelectedKeys.length > 0 &&
          globalSelectedKeys.some(
            (key: React.Key) => !currentPageIds.includes(Number(key)),
          )

        // 当前页面是否全部选中
        const allCurrentPageSelected =
          currentPageIds.length > 0 &&
          currentPageIds.every((id) => globalSelectedKeys.includes(id))

        // 部分选中状态 - 有跨页选择或当前页部分选中
        const indeterminate =
          (hasSelected && !allCurrentPageSelected) || hasCrossPageSelection

        // 处理全选框的点击事件
        const handleSelectAllToggle = (e: any) => {
          const checked = e.target.checked
          let newSelectedKeys = [...globalSelectedKeys]

          if (checked) {
            // 全选：将当前页的所有ID添加到选中列表中
            currentPageIds.forEach((id) => {
              if (!newSelectedKeys.includes(id)) {
                newSelectedKeys.push(id)
              }
            })
          } else {
            // 取消选择：从选中列表中移除当前页的所有ID
            newSelectedKeys = newSelectedKeys.filter(
              (key) => !currentPageIds.includes(Number(key)),
            )
          }

          // 更新选中状态
          setSelectedRowKeys(newSelectedKeys)

          // 更新本地存储
          saveTableParamsToStorage({
            currentPage,
            pageSize,
            sorts: sortItems,
            selectedRowKeys: newSelectedKeys,
          })
        }

        return (
          <div className='relative'>
            <Checkbox
              indeterminate={indeterminate}
              checked={allCurrentPageSelected}
              onChange={handleSelectAllToggle}
            />
          </div>
        )
      },
    }),
    [
      baseRowSelection,
      files,
      globalSelectedKeys,
      currentPage,
      pageSize,
      sortItems,
      saveTableParamsToStorage,
      setSelectedRowKeys,
    ],
  )

  // 路径变化时重新读取存储参数和清空选中状态
  useEffect(() => {
    // 当路径变化时，重新获取对应路径的存储参数
    const newStorageKey = getStorageKey()
    const newParams = getStorageParams(newStorageKey)

    // 更新状态为新路径对应的存储参数
    setCurrentPage(newParams.currentPage)
    setPageSize(newParams.pageSize)
    setSortItems(newParams.sorts || [])

    // 清空选中状态
    setSelectedRowKeys([])
  }, [id, folderId, setSelectedRowKeys])

  // 初始加载数据
  useEffect(() => {
    // 获取知识库信息
    fetchKnowledgeBaseInfo()

    // 清空搜索和过滤状态
    setSearchKeyword('')
    setFilters([])

    if (isRootLevel) {
      // 根级别不需要获取文件夹详情
      if (id) {
        fetchFileList(1, pageSize)
      }
    } else {
      // 文件夹级别需要获取文件夹详情
      fetchFolderDetail()
      if (id) {
        fetchFileList(1, pageSize)
      }
    }
  }, [id, folderId, isRootLevel])

  // 使用文件编辑hook
  const {
    editingId,
    editInputRef,
    startEditing,
    saveEditing,
    createNewFolder,
  } = useFileEdit({
    files,
    setFiles,
    knowledgeBaseId: Number(id),
    parentId: isRootLevel ? 0 : Number(folderId),
    currentLevel: isRootLevel ? 0 : folderInfo.level,
    onSuccess: () => fetchFileList(currentPage, pageSize, filters),
  })

  // 使用文件上传hook
  const {
    uploadLoading,
    uploadRef,
    allowedFileTypes,
    handleFileSelect,
    handleUpload,
  } = useFileUpload({
    knowledgeBaseId: Number(id),
    parentId: isRootLevel ? 0 : Number(folderId),
    onSuccess: () => fetchFileList(currentPage, pageSize, filters),
  })

  // 使用文件删除hook
  const {
    deleteModalVisible,
    deleteModalConfig,
    setDeleteModalVisible,
    handleDelete,
    handleBatchDelete,
    handleConfirmSingleDelete,
    handleConfirmBatchDelete,
  } = useFileDelete({
    files,
    setFiles,
    selectedRowKeys: globalSelectedKeys,
    setSelectedRowKeys,
    onRefresh: (page?: number) =>
      fetchFileList(page || currentPage, pageSize, filters),
    saveTableParams: saveTableParamsToStorage,
    currentTableParams: {
      currentPage,
      pageSize,
      sorts: sortItems,
      selectedRowKeys: globalSelectedKeys,
    },
    setCurrentPage,
    total,
  })

  // 使用轮询hook
  const { setPollingEnabled } = usePolling({
    callback: () =>
      fetchFileList(currentPage, pageSize, filters, sortItems, true),
    interval: POLLING_INTERVAL,
    enabled: true,
  })

  // 在进入编辑模式或上传文件时暂停轮询，避免干扰用户操作
  useEffect(() => {
    if (editingId !== null || uploadLoading) {
      setPollingEnabled(false)
    } else {
      setPollingEnabled(true)
    }
  }, [editingId, uploadLoading, setPollingEnabled])

  // 处理分页变化
  const handlePageChange = (page: number, size?: number) => {
    const newPageSize = size || pageSize

    // 先更新状态
    setCurrentPage(page)
    if (size && size !== pageSize) {
      setPageSize(size)
    }

    // 立即保存更新后的表格参数到本地存储
    const paramsToSave = {
      currentPage: page,
      pageSize: newPageSize,
      sorts: sortItems,
      selectedRowKeys: globalSelectedKeys, // 保持当前选中状态
    }

    saveTableParamsToStorage(paramsToSave)

    // 获取新页面数据
    fetchFileList(page, newPageSize, filters)
  }

  // 处理表格变化
  const handleTableChange = (sorter: any) => {
    // 处理表格排序
    let newSortItems: SortItem[] = []

    if (sorter) {
      // 处理单列排序
      if (
        sorter.field ||
        sorter.columnKey ||
        (sorter.column && sorter.column.dataIndex)
      ) {
        const field =
          sorter.field ||
          sorter.columnKey ||
          (sorter.column && sorter.column.dataIndex)

        if (sorter.order) {
          newSortItems = [
            {
              field,
              order: sorter.order as 'ascend' | 'descend',
            },
          ]
        }
      }
    }

    // 更新状态
    setSortItems(newSortItems)

    // 立即保存更新后的表格参数到本地存储
    saveTableParamsToStorage({
      currentPage,
      pageSize,
      sorts: newSortItems,
      selectedRowKeys: globalSelectedKeys, // 保持当前选中状态
    })

    // 使用新的排序参数获取数据
    fetchFileList(currentPage, pageSize, undefined, newSortItems)
  }

  // 处理移动文件/文件夹
  const handleMove = (fileId: number) => {
    // 查找要移动的项目
    const item = files.find((item) => item.id === fileId)
    if (!item) {
      message.warning('未找到要移动的项目')
      return
    }

    // 设置要移动的项目并显示对话框
    setItemToMove(item)
    setMoveModalVisible(true)
  }

  // 移动成功后的回调
  const handleMoveSuccess = () => {
    // 刷新文件列表
    fetchFileList(currentPage, pageSize, filters)
  }

  // 使用提取的getColumns函数获取表格列
  const columns = getColumns({
    disabled: !isAdmin,
    editingId,
    editInputRef,
    sortItems,
    handleDelete,
    startEditing,
    saveEditing,
    handleFolderClick,
    currentLevel: isRootLevel ? 0 : folderInfo.level,
    handleMove,
    handleFileClick,
  })

  // 处理搜索
  const handleSearch = (values: { value: string; image_url?: string }) => {
    const { value, image_url } = values
    setSearchKeyword(value)

    handleListSearch(value, (newFilters) => {
      setFilters(newFilters)

      // 重置到第一页
      setCurrentPage(1)

      // 立即保存更新后的表格参数到本地存储
      saveTableParamsToStorage({
        currentPage: 1,
        pageSize,
        sorts: sortItems,
        selectedRowKeys: globalSelectedKeys, // 保持当前选中状态
      })

      // 使用新的过滤条件获取数据
      fetchFileList(1, pageSize, newFilters, undefined, undefined, image_url)
    })
  }

  return (
    <div className='w-full h-full p-4'>
      {/* 整个内容区域使用白色背景 */}
      <div className='bg-white rounded-lg p-4'>
        {/* 面包屑导航 */}
        <BreadcrumbNav
          isRootLevel={isRootLevel}
          knowledgeBaseId={id!}
          knowledgeBaseName={knowledgeBaseName}
          folderInfo={folderInfo}
        />

        {/* 操作区域 */}
        <ActionButtons
          forest_id={Number(id)}
          parent_id={isRootLevel ? 0 : Number(folderId)}
          refreshTable={() => fetchFileList(currentPage, pageSize, filters)}
          disabled={!isAdmin}
          uploadLoading={uploadLoading}
          onUpload={handleUpload}
          onCreateFolder={createNewFolder}
          onBatchDelete={handleBatchDelete}
          onSearch={handleSearch}
          searchKeyword={searchKeyword}
          setSearchKeyword={setSearchKeyword}
          uploadRef={uploadRef}
          allowedFileTypes={allowedFileTypes}
          onFileSelect={handleFileSelect}
          onWordCloud={() => navigate(`/docs/${id}/wordcloud`)}
          onKnowledgeGraph={() => navigate(`/docs/${id}/knowledge-graph`)}
        />

        {/* 表格与分页 */}
        <div>
          <TableWithPagination<FileItem>
            columns={columns}
            dataSource={files}
            rowKey='id'
            current={currentPage}
            pageSize={pageSize}
            total={total}
            pageSizeOptions={pageSizeOptions}
            onPageChange={handlePageChange}
            onTableChange={handleTableChange}
            loading={loading}
            className={
              isRootLevel ? 'knowledge-detail-table' : 'folder-detail-table'
            }
            rowSelection={rowSelection}
            rowClassName={(record) => {
              const baseClass = 'h-10'
              const hoverClass = 'hover:bg-[#f7f9fc]'
              const editingClass = editingId === record.id ? 'bg-[#f7f9fc]' : ''
              return `${baseClass} ${hoverClass} ${editingClass}`.trim()
            }}
            tableHeight={{
              default: 'h-[calc(100vh-224px)]',
              sm: 'sm:h-[calc(100vh-224px)]',
              lg: 'lg:h-[calc(100vh-224px)]',
              '2xl': '2xl:h-[calc(100vh-224px)]',
            }}
          />
        </div>
      </div>

      {/* 删除确认对话框 */}
      <DeleteConfirmModal
        visible={deleteModalVisible}
        isFolder={deleteModalConfig.isFolder}
        isMultiple={deleteModalConfig.isMultiple}
        onCancel={() => setDeleteModalVisible(false)}
        onConfirm={
          deleteModalConfig.isMultiple
            ? handleConfirmBatchDelete
            : handleConfirmSingleDelete
        }
      />

      {/* 移动到对话框 */}
      <MoveToModal
        visible={moveModalVisible}
        onCancel={() => {
          setMoveModalVisible(false)
          setItemToMove(null)
        }}
        knowledgeBaseId={id}
        selectedItem={itemToMove!}
        onSuccess={handleMoveSuccess}
      />
    </div>
  )
}
