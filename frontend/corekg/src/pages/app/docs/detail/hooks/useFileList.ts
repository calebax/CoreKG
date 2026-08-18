import {
  useState,
  useCallback,
  useRef,
  useEffect,
  Dispatch,
  SetStateAction,
} from 'react'
import dayjs from 'dayjs'
import { getKnowledgeBaseDetail, getFileList } from '@/api/knowledge'
import type { KnowledgeBaseType } from '../components/ActionButtons/UploadButton/Uploader'
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
  knowledgeBaseCreatedAt?: string
  knowledgeBaseType: KnowledgeBaseType
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
  handleParseStatusFilter: (
    parseStatus: string,
    onFilter: (filters: FilterItem[]) => void,
  ) => void
  handleTagFilter: (
    tagId: string | null,
    onFilter: (filters: FilterItem[]) => void,
  ) => void
  knowledgeBaseDesc: string
  setKnowledgeBaseName: Dispatch<SetStateAction<string>>
  setKnowledgeBaseDesc: Dispatch<SetStateAction<string>>
  knowledgeBaseSize: number
  knowledgeBaseFileCount: number
  graph_info?: any
  graph_updatable?: boolean
  knowledge_status?: string
}

// 处理文件数据的工具函数
const processFileData = (item: any): FileItem => {
  return {
    ...item,
    id: item.ID,
    name: item.name,
    size: formatFileSize(item.size),
    updatedAt: dayjs(item.UpdatedAt || item.CreatedAt).format(
      'YYYY-MM-DD HH:mm:ss',
    ), // 显示用的日期格式，优先使用更新时间
    UpdatedAt: item.UpdatedAt, // API返回的更新时间
    createdAtFull: item.CreatedAt, // 完整的时间戳，用于排序
    updatedAtFull: item.UpdatedAt || item.CreatedAt, // 完整的更新时间戳，用于排序
    isFolder: item.is_dir === true,
    fileType: item.ext ? item.ext.replace(/^\./, '') : '-', // 移除可能的前导点
    tag_list: item.tag_list || null, // 保留标签列表
  }
}

// 默认按创建时间降序
const sortFilesByTime = (files: FileItem[], sortsToUse: SortItem[]) => {
  if (!sortsToUse || sortsToUse.length === 0) return files
  return files
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
  const [knowledgeBaseDesc, setKnowledgeBaseDesc] = useState<string>('')
  const [knowledgeBaseCreatedAt, setKnowledgeBaseCreatedAt] = useState<string>()
  const [knowledgeBaseType, setKnowledgeBaseType] =
    useState<KnowledgeBaseType>()
  const [knowledgeBaseSize, setKnowledgeBaseSize] = useState<number>(0)
  const [knowledgeBaseFileCount, setKnowledgeBaseFileCount] =
    useState<number>(0)
  const [knowledge_status, setKnowledge_status] = useState('')

  // 是否是预置知识库
  const [isSystemKnowledgeBase, setIsSystemKnowledgeBase] =
    useState<boolean>(false)

  const sortItemsRef = useRef(sortItems)
  const filtersRef = useRef(filters)

  useEffect(() => {
    sortItemsRef.current = sortItems
  }, [sortItems])

  useEffect(() => {
    filtersRef.current = filters
  }, [filters])
  const [graph_info, setGraphInfo] = useState<any>()
  const [graph_updatable, setGraphUpdatable] = useState<boolean>()
  // 获取知识库详情
  const fetchKnowledgeBaseInfo = useCallback(async () => {
    try {
      if (knowledgeBaseId) {
        const res = await getKnowledgeBaseDetail({ id: knowledgeBaseId })

        if (res && res.data) {
          setKnowledgeBaseName(res.data.name)
          setKnowledgeBaseDesc(res.data.description || '')
          setKnowledgeBaseSize(res.data.total_size)
          setKnowledgeBaseFileCount(res.data.file_count)
          setKnowledge_status(res.data.knowledge_status || '')
          setIsAdmin(res.data.is_admin)
          setIsSystemKnowledgeBase(
            res.data.knowledge_status === 'system' ? true : false,
          )
          setGraphInfo(res.graph_info)
          setGraphUpdatable(res.data.graph_status === 'updatable')
          // 设置创建时间和类型
          setKnowledgeBaseCreatedAt(res.data.CreatedAt)
          // 优先使用 data 类型下的具体子类型来判定（excel/db）
          const forestType = res.data.forest_type
          const dataSourceType = res.data.data_source_type
          let kbType: KnowledgeBaseType
          if (forestType === 'data') {
            kbType = dataSourceType === 'excel' ? 'excel' : 'data'
          } else {
            kbType = normalizeKnowledgeBaseType(forestType)
          }
          setKnowledgeBaseType(kbType)

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
      setKnowledgeBaseDesc('')
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
        customSortItems !== undefined ? customSortItems : sortItemsRef.current

      // 使用自定义过滤条件或当前状态中的过滤条件
      const filtersToUse =
        customFilters !== undefined ? customFilters : filtersRef.current

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
          // 如果有搜索关键词，进行搜索但依然限制parent_id，确保在当前文件夹范围内搜索
          allFilters = [
            {
              field: 'parent_id',
              value: [isRootLevel ? '0' : parentId.toString()],
            },
            ...filtersToUse,
          ]
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

          // 保持服务端排序（若无前端排序指令，不做二次排序）
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
    [knowledgeBaseId, parentId, isRootLevel, loading],
  )

  // 处理搜索
  const handleSearch = useCallback(
    (value: string, onSearch: (filters: FilterItem[]) => void) => {
      // 获取当前的标签筛选条件和资源状态筛选条件（保留tag_ids和file_status）
      const currentTagFilter = filtersRef.current.find(
        (filter) => filter.field === 'tag_ids',
      )
      const currentStatusFilter = filtersRef.current.find(
        (filter) => filter.field === 'file_status',
      )

      // 构建符合要求的过滤条件
      const newFilters: FilterItem[] = []

      if (value) {
        newFilters.push({
          field: 'name',
          value: [value],
          exactMatch: false, // 模糊匹配
        })
      }

      // 保留资源状态筛选条件
      if (currentStatusFilter) {
        newFilters.push(currentStatusFilter)
      }

      // 保留标签筛选条件
      if (currentTagFilter) {
        newFilters.push(currentTagFilter)
      }

      // 更新过滤条件状态
      setFilters(newFilters)

      // 调用外部回调
      onSearch(newFilters)
    },
    [],
  )

  // 处理资源状态筛选
  const handleParseStatusFilter = useCallback(
    (parseStatus: string, onFilter: (filters: FilterItem[]) => void) => {
      // 获取当前的标签筛选条件（保留tag_ids）
      const currentTagFilter = filtersRef.current.find(
        (filter) => filter.field === 'tag_ids',
      )

      // 构建符合要求的过滤条件
      const newFilters: FilterItem[] = []

      if (parseStatus && parseStatus !== 'all') {
        newFilters.push({
          field: 'file_status',
          value: [parseStatus],
        })
      }

      // 保留标签筛选条件
      if (currentTagFilter) {
        newFilters.push(currentTagFilter)
      }

      // 更新过滤条件状态
      setFilters(newFilters)

      // 调用外部回调
      onFilter(newFilters)
    },
    [],
  )

  // 处理标签筛选
  const handleTagFilter = useCallback(
    (tagId: string | null, onFilter: (filters: FilterItem[]) => void) => {
      // 获取当前的其他过滤条件（排除tag_ids）
      const otherFilters = filtersRef.current.filter(
        (filter) => filter.field !== 'tag_ids',
      )

      // 构建新的过滤条件
      const newFilters: FilterItem[] = [...otherFilters]

      if (tagId) {
        newFilters.push({
          field: 'tag_ids',
          value: [tagId],
        })
      }

      // 更新过滤条件状态
      setFilters(newFilters)

      // 调用外部回调
      onFilter(newFilters)
    },
    [],
  )

  // 统一类型到白名单
  const normalizeKnowledgeBaseType = (value: any): KnowledgeBaseType => {
    const allowed: KnowledgeBaseType[] = ['file', 'excel', 'qa', 'data']
    if (allowed.includes(value as KnowledgeBaseType))
      return value as KnowledgeBaseType
    return 'file'
  }

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
    setKnowledgeBaseName,
    knowledgeBaseDesc,
    setKnowledgeBaseDesc,
    knowledgeBaseCreatedAt,
    knowledgeBaseType,
    fetchFileList,
    fetchKnowledgeBaseInfo,
    fetchFolderDetail,
    handleSearch,
    handleParseStatusFilter,
    handleTagFilter,
    knowledgeBaseSize,
    knowledgeBaseFileCount,
    graph_info,
    graph_updatable,
    knowledge_status,
  }
}
