import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { message, Modal, Button } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  updateKnowledgeBaseDesc,
  updateKnowledgeBaseName,
  updateResourceEnable,
} from '@/api/knowledge'
import TableWithPagination from '@/components/common/TableWithPagination'
import { useDeployConfig } from '@/utils/useDeployConfig'
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
} from '../hooks'
// 导入提取的工具和组件

// 导入类型
import { FileItem, SortItem, STORAGE_KEY } from '../types'
// 导入表格列
import { getColumns } from '../utils/tableColumns'
// 导入操作按钮
// 导入面包屑导航
import BreadcrumbNav from './BreadcrumbNav'
// 导入删除确认对话框
import DeleteConfirmModal from './DeleteConfirmModal'
// 导入文件权限侧边栏
import FilePermissionSidebar from './FilePermissionSidebar'
// 导入知识库信息组件
import KnowledgeBaseInfo from './KnowledgeBaseInfo'
// 导入移动到对话框
import MoveToModal from './MoveToModal'
// 导入操作区域组件
import OperationBar from './OperationBar'
// 导入标签选择弹窗
import TagSelectModal from './TagSelectModal'

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
  const { t } = useTranslation('pages')
  const { version, mode } = useDeployConfig()

  // 是否显示文件权限功能：本地环境、测试环境、生产环境、或 custom+cimc/h3c 模式
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isProdEnv = import.meta.env.MODE === 'production'
  const showFilePermission =
    isDevEnv ||
    isTestEnv ||
    isProdEnv ||
    (version === 'custom' && (mode === 'cimc' || mode === 'h3c'))

  // 是否显示标签功能：本地环境、测试环境、生产环境、或 custom+cimc/h3c 模式
  const showTagColumn =
    isDevEnv ||
    isTestEnv ||
    isProdEnv ||
    (version === 'custom' && (mode === 'cimc' || mode === 'h3c'))

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

  // 文件权限侧边栏状态
  const [permissionSidebarVisible, setPermissionSidebarVisible] =
    useState(false)
  const [selectedFileForPermission, setSelectedFileForPermission] =
    useState<FileItem | null>(null)

  // 标签选择弹窗状态
  const [tagModalVisible, setTagModalVisible] = useState(false)
  const [selectedFileForTag, setSelectedFileForTag] = useState<FileItem | null>(
    null,
  )

  // 启用状态切换loading状态（记录正在处理的文件ID）
  const [enablingFileId, setEnablingFileId] = useState<number | null>(null)

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
        }
      )
    } catch (error) {
      console.error('读取存储参数失败:', error)
      return {
        currentPage: 1,
        pageSize: 10,
        sorts: [],
      }
    }
  }

  // 使用本地存储hook
  const storageKey = getStorageKey()
  const {
    getInitialTableParams,
    saveTableParamsToStorage,
    clearCurrentStorage,
  } = useTableStorage({
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
  const {
    getPathFromQuery,
    handleFolderClick: navigateToFolder,
    handleFileClick,
    handleFileEdit,
  } = useFileNavigation({
    knowledgeBaseId: id!,
    folderId,
    isRootLevel,
    folderInfo,
    knowledgeBaseName: '',
  })

  const handleFolderClick = useCallback(
    (folder: FileItem) => {
      clearCurrentStorage()
      navigateToFolder(folder)
    },
    [clearCurrentStorage, navigateToFolder],
  )

  const handleBreadcrumbNavigate = useCallback(() => {
    clearCurrentStorage()
  }, [clearCurrentStorage])

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
    setKnowledgeBaseName,
    knowledgeBaseDesc,
    setKnowledgeBaseDesc,
    knowledgeBaseCreatedAt,
    knowledgeBaseType,
    fetchFileList,
    fetchKnowledgeBaseInfo,
    fetchFolderDetail,
    handleSearch: handleListSearch,
    handleParseStatusFilter,
    handleTagFilter,
    knowledgeBaseSize,
    knowledgeBaseFileCount,
    graph_info,
    graph_updatable,
    knowledge_status,
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

  // 路径变化时重新读取存储参数
  useEffect(() => {
    // 当路径变化时，重新获取对应路径的存储参数
    const newStorageKey = getStorageKey()
    const newParams = getStorageParams(newStorageKey)

    // 更新状态为新路径对应的存储参数
    setCurrentPage(newParams.currentPage)
    setPageSize(newParams.pageSize)
    setSortItems(newParams.sorts || [])
  }, [id, folderId])

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
        fetchFileList(currentPage, pageSize)
      }
    } else {
      // 文件夹级别需要获取文件夹详情
      fetchFolderDetail()
      if (id) {
        fetchFileList(currentPage, pageSize)
      }
    }
  }, [id, folderId, isRootLevel, currentPage, pageSize])

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
    selectedRowKeys: [],
    setSelectedRowKeys: () => {},
    onRefresh: (page?: number) =>
      fetchFileList(page || currentPage, pageSize, filters),
    saveTableParams: saveTableParamsToStorage,
    currentTableParams: {
      currentPage,
      pageSize,
      sorts: sortItems,
      selectedRowKeys: [],
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

  // 处理文件权限
  const handlePermission = (file: FileItem) => {
    setSelectedFileForPermission(file)
    setPermissionSidebarVisible(true)
  }

  // 处理标签选择
  const handleTag = (file: FileItem) => {
    setSelectedFileForTag(file)
    setTagModalVisible(true)
  }

  // 标签设置成功后的回调
  const handleTagSuccess = () => {
    fetchFileList(currentPage, pageSize, filters)
  }

  // 启用状态
  const handleEnableChange = async (enable: boolean, file: FileItem) => {
    // 如果正在处理该文件，直接返回
    if (enablingFileId === file.id) {
      return
    }

    // 设置loading状态，禁用开关
    setEnablingFileId(file.id)

    try {
      await updateResourceEnable({
        enable: enable ? 1 : -1,
        forest_id: Number(id),
        resource_ids: [String(file.id)],
        resource_type: knowledgeBaseType === 'excel' ? 'excel' : 'file',
      })

      // 接口成功后再更新UI状态
      setFiles(
        files.map((item) => (item.id === file.id ? { ...item, enable } : item)),
      )
    } catch (error) {
      console.log(error)
    } finally {
      // 清除loading状态
      setEnablingFileId(null)
    }
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
    handlePermission,
    handleFileClick,
    handleFileEdit,
    knowledgeBaseType,
    handleEnableChange,
    showFilePermission,
    showTagColumn,
    handleTag,
    enablingFileId,
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
      })

      // 使用新的过滤条件获取数据
      fetchFileList(1, pageSize, newFilters, undefined, undefined, image_url)
    })
  }

  const handleChangeDesc = async (newDesc: string) => {
    try {
      // if (!newDesc.trim()) {
      //   message.error(t('app.docs.detail.fileEdit.nameRequired'))
      //   return
      // }
      if (newDesc === knowledgeBaseDesc) {
        // message.success(t('app.docs.detail.fileEdit.renameSuccess'))
        return
      }
      if (id !== undefined && id !== null) {
        await updateKnowledgeBaseDesc({
          forest_id: +id,
          description: newDesc,
        })
        setKnowledgeBaseDesc(newDesc)
        message.success(t('app.docs.detail.fileEdit.renameSuccess'))
      }
    } catch (error) {
      console.log(error)
    }
  }

  const handleChangeName = async (newName: string) => {
    try {
      if (!newName.trim()) {
        message.error(t('app.docs.kbNameRequired'))
        return
      }
      if (newName === knowledgeBaseName) {
        // message.success(t('app.docs.detail.fileEdit.renameSuccess'))
        return
      }
      if (id !== undefined && id !== null) {
        await updateKnowledgeBaseName({ id: +id, name: newName })
        setKnowledgeBaseName(newName)
        message.success(t('app.docs.detail.fileEdit.renameSuccessTip'))
      }
    } catch (error) {
      console.log(error)
    }
  }

  return (
    <div className='w-full h-full bg-white flex flex-col'>
      {/* 面包屑导航 */}
      <BreadcrumbNav
        isRootLevel={isRootLevel}
        knowledgeBaseId={id!}
        knowledgeBaseName={knowledgeBaseName}
        folderInfo={folderInfo}
        onNavigateAway={handleBreadcrumbNavigate}
      />

      {/* 主内容容器 - 统一设置内边距 */}
      <div className='flex-1 flex flex-col px-[48px] py-[10px] overflow-hidden gap-[20px]'>
        {/* 知识库信息区域 - 在所有层级显示 */}
        <KnowledgeBaseInfo
          knowledgeBaseName={knowledgeBaseName}
          knowledgeBaseId={Number(id)}
          knowledgeBaseSize={knowledgeBaseSize}
          knowledgeBaseFileCount={knowledgeBaseFileCount}
          createdAt={knowledgeBaseCreatedAt}
          isAdmin={isAdmin}
          knowledgeBaseType={knowledgeBaseType}
          parent_id={isRootLevel ? 0 : Number(folderId)}
          refreshTable={() => fetchFileList(currentPage, pageSize, filters)}
          disabled={!isAdmin}
          description={knowledgeBaseDesc}
          onCreateFolder={createNewFolder}
          onChangeDesc={handleChangeDesc}
          onChangeName={handleChangeName}
          graph_info={graph_info}
          graph_updatable={graph_updatable}
          is_admin={isAdmin}
          knowledge_status={knowledge_status}
        />

        {/* 操作区域 - 新增的筛选和搜索区域 */}
        <OperationBar
          forest_id={Number(id)}
          parent_id={isRootLevel ? 0 : Number(folderId)}
          refreshTable={() => fetchFileList(currentPage, pageSize, filters)}
          disabled={!isAdmin}
          knowledgeBaseType={knowledgeBaseType}
          onCreateFolder={createNewFolder}
          onSearch={(keyword) => {
            handleListSearch(keyword, (newFilters) => {
              setFilters(newFilters)
              setCurrentPage(1)
              saveTableParamsToStorage({
                currentPage: 1,
                pageSize,
                sorts: sortItems,
              })
              fetchFileList(1, pageSize, newFilters)
            })
          }}
          onFilterChange={(parseStatus) => {
            handleParseStatusFilter(parseStatus, (newFilters) => {
              setFilters(newFilters)
              setCurrentPage(1)
              saveTableParamsToStorage({
                currentPage: 1,
                pageSize,
                sorts: sortItems,
              })
              fetchFileList(1, pageSize, newFilters)
            })
          }}
          onTagFilterChange={(tagId) => {
            handleTagFilter(tagId, (newFilters) => {
              setFilters(newFilters)
              setCurrentPage(1)
              saveTableParamsToStorage({
                currentPage: 1,
                pageSize,
                sorts: sortItems,
              })
              fetchFileList(1, pageSize, newFilters)
            })
          }}
        />

        {/* 表格与分页 */}
        <div className='flex-1 overflow-hidden'>
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
            rowClassName={(record) => {
              const baseClass = 'h-10'
              const hoverClass = 'hover:bg-[#f7f9fc]'
              const editingClass = editingId === record.id ? 'bg-[#f7f9fc]' : ''
              return `${baseClass} ${hoverClass} ${editingClass}`.trim()
            }}
            onRow={(record) => ({
              onClick: (e) => {
                // 如果点击的是操作列，不触发行点击
                const target = e.target as HTMLElement
                const isActionColumn = target.closest('td[key="actions"]')

                if (isActionColumn) return

                // 如果是编辑状态，不触发行点击
                if (editingId === record.id) return

                // 文件夹点击
                if (record.isFolder && handleFolderClick) {
                  handleFolderClick(record)
                  return
                }

                // 文件点击
                if (!record.isFolder && handleFileClick) {
                  handleFileClick(record)
                }
              },
              style: { cursor: 'pointer' },
            })}
            tableHeight={{
              default: 'h-full',
              sm: 'sm:h-full',
              lg: 'lg:h-full',
              '2xl': '2xl:h-full',
            }}
            paginationProps={{
              className: 'pagination-compact',
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
        knowledgeBaseName={knowledgeBaseName}
        knowledgeBaseCreatedAt={knowledgeBaseCreatedAt}
        knowledgeBaseType={knowledgeBaseType}
        selectedItem={itemToMove!}
        onSuccess={handleMoveSuccess}
      />

      {/* 文件权限侧边栏 */}
      <FilePermissionSidebar
        open={permissionSidebarVisible}
        onClose={() => {
          setPermissionSidebarVisible(false)
          setSelectedFileForPermission(null)
        }}
        fileId={selectedFileForPermission?.id || 0}
      />

      {/* 标签选择弹窗 */}
      <TagSelectModal
        open={tagModalVisible}
        file={selectedFileForTag}
        onCancel={() => {
          setTagModalVisible(false)
          setSelectedFileForTag(null)
        }}
        onSuccess={handleTagSuccess}
      />
    </div>
  )
}
