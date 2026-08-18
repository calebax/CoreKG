import { useState, useCallback } from 'react'
import dayjs from 'dayjs'
import { getKnowledgeBaseDetail, getFileList } from '@/api/knowledge'
import { FileItem, FilterItem, SortItem } from '../types'
import { buildOrderByParams, formatFileSize } from '../utils/fileUtils'

interface PathItem {
  id: number
  name: string
  level: number
}

interface UseFileListProps {
  knowledgeBaseId: number
  parentId: number
  isRootLevel: boolean
  currentPage: number
  pageSize: number
  sortItems: SortItem[]
  getPathFromQuery: () => PathItem[]
  folderInfo: {
    name: string
    level: number
    path: PathItem[]
  }
  setFolderInfo: (info: any) => void
}

interface UseFileListReturn {
  isAdmin: boolean
  isSystemKnowledgeBase: boolean
  files: FileItem[]
  setFiles: (files: FileItem[]) => void
  total: number
  loading: boolean
  filters: FilterItem[]
  setFilters: (filters: FilterItem[]) => void
  knowledgeBaseName: string
  fetchFileList: (
    page: number,
    size: number,
    customFilters?: FilterItem[],
    customSortItems?: SortItem[],
    isPolling?: boolean,
    image_url?: string,
  ) => Promise<void>
  fetchKnowledgeBaseInfo: () => Promise<void>
  fetchFolderDetail: () => Promise<void>
  handleSearch: (
    value: string,
    onSearch: (filters: FilterItem[]) => void,
  ) => void
}

// 处理文件数据的工具函数
const processFileData = (item: any): FileItem => {
  return {
    ...item,
    id: item.ID,
    name: item.name,
    size: formatFileSize(item.size),
    updatedAt: dayjs(item.CreatedAt).format('YYYY-MM-DD HH:mm:ss'), // 显示用的日期格式
    createdAtFull: item.CreatedAt, // 完整的时间戳，用于排序
    isFolder: item.is_dir === true,
    fileType: item.ext ? item.ext.replace(/^\./, '') : '-', // 移除可能的前导点
  }
}

// 默认时间排序函数
const sortFilesByTime = (files: FileItem[], sortsToUse: SortItem[]) => {
  return [...files].sort((a, b) => {
    // 如果没有用户自定义排序，则按时间降序排列（最新的在前）
    if (!sortsToUse || sortsToUse.length === 0) {
      // 使用完整时间戳进行排序，如果没有则使用显示时间作为备选
      const timeA = new Date(a.createdAtFull || a.updatedAt).getTime()
      const timeB = new Date(b.createdAtFull || b.updatedAt).getTime()
      return timeB - timeA // 降序：新的在前
    }
    return 0 // 有用户排序时不进行额外排序
  })
}

export const useFileList = ({
  knowledgeBaseId,
  parentId,
  isRootLevel,
  currentPage,
  pageSize,
  sortItems,
  getPathFromQuery,
  folderInfo,
  setFolderInfo,
}: UseFileListProps): UseFileListReturn => {
  const [isAdmin, setIsAdmin] = useState(false)
  const [files, setFiles] = useState<FileItem[]>([])
  const [total, setTotal] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(false)
  const [filters, setFilters] = useState<FilterItem[]>([])
  const [knowledgeBaseName, setKnowledgeBaseName] = useState<string>('')
  // 是否是预置知识库
  const [isSystemKnowledgeBase, setIsSystemKnowledgeBase] =
    useState<boolean>(false)

  // 获取知识库详情
  const fetchKnowledgeBaseInfo = useCallback(async () => {
    try {
      if (knowledgeBaseId) {
        const res = await getKnowledgeBaseDetail({ id: knowledgeBaseId })
        if (res && res.data) {
          setKnowledgeBaseName(res.data.name)
          setIsAdmin(res.data.is_admin)
          setIsSystemKnowledgeBase(
            res.data.knowledge_status === 'system' ? true : false,
          )
          // 如果是根级别，更新folderInfo的name
          if (isRootLevel) {
            setFolderInfo((prev: any) => ({ ...prev, name: res.data.name }))
          }
        }
      }
    } catch (error) {
      console.error('获取知识库详情失败:', error)
      // 失败时保持空字符串，不显示错误的默认名称
      setKnowledgeBaseName('')
      if (isRootLevel) {
        setFolderInfo((prev: any) => ({ ...prev, name: '' }))
      }
    }
  }, [knowledgeBaseId, isRootLevel, setFolderInfo])

  // 获取文件夹详情 - 仅在非根级别使用
  const fetchFolderDetail = useCallback(async () => {
    if (isRootLevel || !parentId) return

    // 从URL获取路径数据
    const pathFromUrl = getPathFromQuery()

    try {
      // 如果路径不为空，说明有父级文件夹，通过父级获取当前文件夹信息
      if (pathFromUrl.length > 0) {
        const parentFolderId = pathFromUrl[pathFromUrl.length - 1].id
        const res = await getFileList({
          forest_id: knowledgeBaseId,
          limit: 1000, // 获取所有子项来查找当前文件夹
          offset: 0,
          filters: [{ field: 'parent_id', value: [parentFolderId.toString()] }],
        })

        if (res && res.data && Array.isArray(res.data)) {
          // 查找当前文件夹
          const currentFolder = res.data.find(
            (item: any) => item.ID === parentId && item.is_dir,
          )
          if (currentFolder) {
            setFolderInfo({
              name: currentFolder.name,
              level: pathFromUrl.length,
              path: pathFromUrl,
            })
            return
          }
        }
      } else {
        // 如果没有父级路径，说明是从根级别直接进入的文件夹，从根级别获取
        const res = await getFileList({
          forest_id: knowledgeBaseId,
          limit: 1000, // 获取所有子项来查找当前文件夹
          offset: 0,
          filters: [{ field: 'parent_id', value: ['0'] }], // 根级别
        })

        if (res && res.data && Array.isArray(res.data)) {
          // 查找当前文件夹
          const currentFolder = res.data.find(
            (item: any) => item.ID === parentId && item.is_dir,
          )
          if (currentFolder) {
            setFolderInfo({
              name: currentFolder.name,
              level: pathFromUrl.length,
              path: pathFromUrl,
            })
            return
          }
        }
      }

      // 如果找不到文件夹信息，使用默认名称
      setFolderInfo({
        name: `文件夹${parentId}`,
        level: pathFromUrl.length,
        path: pathFromUrl,
      })
    } catch (error) {
      console.error('获取文件夹详情失败:', error)
      // 发生错误时使用默认名称
      setFolderInfo({
        name: `文件夹${parentId}`,
        level: pathFromUrl.length,
        path: pathFromUrl,
      })
    }
  }, [isRootLevel, parentId, getPathFromQuery, knowledgeBaseId, setFolderInfo])

  // 获取文件列表数据
  const fetchFileList = useCallback(
    async (
      page: number,
      size: number,
      customFilters?: FilterItem[],
      customSortItems?: SortItem[],
      isPolling: boolean = false,
      image_url?: string,
    ) => {
      // 如果是轮询模式且正在加载，则跳过此次轮询
      if (isPolling && loading) {
        return
      }

      // 如果不是轮询模式，则设置加载状态
      if (!isPolling) {
        setLoading(true)
      }

      const start = (page - 1) * size

      // 使用自定义排序项或当前状态中的排序项
      const sortsToUse =
        customSortItems !== undefined ? customSortItems : sortItems

      // 使用自定义过滤条件或当前状态中的过滤条件
      const filtersToUse = customFilters !== undefined ? customFilters : filters

      // 构建排序参数
      const orderBy = buildOrderByParams(sortsToUse)

      try {
        // 检查是否有搜索关键词
        const hasSearchKeyword = filtersToUse.some(
          (filter) =>
            filter.field === 'name' &&
            filter.value.length > 0 &&
            filter.value[0].trim() !== '',
        )

        // 构建过滤条件
        let allFilters: FilterItem[]
        if (hasSearchKeyword) {
          // 如果有搜索关键词，进行全局搜索，不限制parent_id
          allFilters = [...filtersToUse]
        } else {
          // 如果没有搜索关键词，添加parent_id限制，只显示当前文件夹下的内容
          allFilters = [
            {
              field: 'parent_id',
              value: [isRootLevel ? '0' : parentId.toString()],
            },
            ...filtersToUse,
          ]
        }

        // 构建接口参数
        const requestParams = {
          forest_id: knowledgeBaseId,
          limit: size,
          offset: start,
          orderBy: orderBy,
          filters: allFilters,
          image_url,
        }

        const res = await getFileList(requestParams)

        if (res && res.data && Array.isArray(res.data)) {
          // 处理返回的文件列表数据
          const processedFiles = res.data.map(processFileData)

          // 前端默认按时间排序（由近及远），不影响表头排序状态
          const sortedFiles = sortFilesByTime(processedFiles, sortsToUse)

          setFiles(sortedFiles)
          setTotal(res.total || 0)

          // 如果返回的数据为空，设置空状态
          if (sortedFiles.length === 0) {
            setFiles([])
          }
        } else {
          // 接口返回空数据或格式不正确时设置空状态
          setFiles([])
          setTotal(0)
        }
      } catch (error) {
        console.error('获取文件列表失败:', error)
        // 发生错误时清空数据
        setFiles([])
        setTotal(0)
      } finally {
        // 非轮询请求才更新加载状态
        if (!isPolling) {
          setLoading(false)
        }
      }
    },
    [knowledgeBaseId, parentId, isRootLevel, loading, sortItems, filters],
  )

  // 处理搜索
  const handleSearch = useCallback(
    (value: string, onSearch: (filters: FilterItem[]) => void) => {
      // 构建符合要求的过滤条件
      const newFilters: FilterItem[] = []

      if (value) {
        newFilters.push({
          field: 'name',
          value: [value],
          exactMatch: false, // 模糊匹配
        })
      }

      // 更新过滤条件状态
      setFilters(newFilters)

      // 调用外部回调
      onSearch(newFilters)
    },
    [],
  )

  return {
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
    handleSearch,
  }
}
