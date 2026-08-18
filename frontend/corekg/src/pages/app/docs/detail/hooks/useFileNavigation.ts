import { useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { FileItem } from '../types'

interface PathItem {
  id: number
  name: string
  level: number
}

interface UseFileNavigationProps {
  knowledgeBaseId: string
  folderId?: string
  isRootLevel: boolean
  folderInfo: {
    name: string
    level: number
    path: PathItem[]
  }
  knowledgeBaseName: string
}

interface UseFileNavigationReturn {
  getPathFromQuery: () => PathItem[]
  handleFolderClick: (folder: FileItem) => void
  handleFileClick: (file: FileItem) => void
  handleFileEdit: (file: FileItem) => void
}

export const useFileNavigation = ({
  knowledgeBaseId,
  folderId,
  isRootLevel,
  folderInfo,
  knowledgeBaseName,
}: UseFileNavigationProps): UseFileNavigationReturn => {
  const navigate = useNavigate()
  const location = useLocation()

  // 获取路径信息 - 仅在非根级别使用
  const getPathFromQuery = useCallback(() => {
    if (isRootLevel) return []

    try {
      // 从URL查询参数中获取路径
      const query = new URLSearchParams(location.search)
      const pathStr = query.get('path')
      if (pathStr) {
        return JSON.parse(decodeURIComponent(pathStr)) as PathItem[]
      }
    } catch (error) {
      console.error('解析路径数据失败', error)
    }
    return []
  }, [isRootLevel, location.search])

  // 处理文件夹点击，导航到文件夹详情页
  const handleFolderClick = useCallback(
    (folder: FileItem) => {
      if (isRootLevel) {
        // 根级别进入文件夹，构建空路径
        const pathParam = encodeURIComponent(JSON.stringify([]))
        // 导航到文件夹详情页
        navigate(
          `/docs/detail/${knowledgeBaseId}/folder/${folder.id}?path=${pathParam}`,
        )
      } else {
        // 非根级别，构建新的路径
        const currentPath = getPathFromQuery()
        const newPath = [
          ...currentPath,
          {
            id: parseInt(folderId || '0'),
            name: folderInfo.name,
            level: folderInfo.level,
          },
        ]
        // 将路径数据编码为URL查询参数
        const pathParam = encodeURIComponent(JSON.stringify(newPath))
        // 导航到下一级文件夹详情页
        navigate(
          `/docs/detail/${knowledgeBaseId}/folder/${folder.id}?path=${pathParam}`,
        )
      }
    },
    [
      isRootLevel,
      knowledgeBaseId,
      folderId,
      folderInfo,
      getPathFromQuery,
      navigate,
    ],
  )

  // 处理文件点击，直接导航到分段管理页面
  const handleFileClick = useCallback(
    (file: FileItem) => {
      // 构建路由参数，传递知识库名称、文件名称和路径信息
      const params = new URLSearchParams()

      // 确保文件名包含后缀，用于预览类型识别
      const ensureExt = (f: FileItem) => {
        const name = f.name || '未知文件'
        // 已有扩展名
        if (name.includes('.') && !name.endsWith('.')) return name
        // 从 fileType 补全扩展名
        const ext = (f as any).fileType
        if (ext && ext !== '-') return `${name}.${ext}`
        return name
      }

      params.set('kbName', knowledgeBaseName || '知识库')
      params.set('fileName', ensureExt(file))

      // 传递当前的路径信息，用于面包屑导航
      if (!isRootLevel) {
        const currentPath = getPathFromQuery()
        const fullPath = [
          ...currentPath,
          {
            id: parseInt(folderId || '0'),
            name: folderInfo.name,
            level: folderInfo.level,
          },
        ]
        params.set('folderPath', JSON.stringify(fullPath))
      }

      // 直接导航到分段管理页面
      navigate(
        `/docs/detail/${knowledgeBaseId}/file/${file.id}?${params.toString()}`,
      )
    },
    [
      knowledgeBaseId,
      knowledgeBaseName,
      navigate,
      isRootLevel,
      folderId,
      folderInfo,
      getPathFromQuery,
    ],
  )

  // 处理文件编辑，导航到文件编辑页面
  const handleFileEdit = useCallback(
    (file: FileItem) => {
      // 构建路由参数，传递知识库名称和文件名称
      const params = new URLSearchParams()
      const ensureExt = (f: FileItem) => {
        const name = f.name || '未知文件'
        if (name.includes('.') && !name.endsWith('.')) return name
        const ext = (f as any).fileType
        if (ext && ext !== '-') return `${name}.${ext}`
        return name
      }
      params.set('kbName', knowledgeBaseName || '知识库')
      params.set('fileName', ensureExt(file))

      navigate(
        `/docs/detail/${knowledgeBaseId}/file/${file.id}/edit?${params.toString()}`,
      )
    },
    [knowledgeBaseId, navigate, knowledgeBaseName],
  )

  return {
    getPathFromQuery,
    handleFolderClick,
    handleFileClick,
    handleFileEdit,
  }
}
