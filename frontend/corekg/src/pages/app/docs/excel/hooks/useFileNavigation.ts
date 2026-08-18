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
          `/docs/excel/${knowledgeBaseId}/folder/${folder.id}?path=${pathParam}`,
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
          `/docs/excel/${knowledgeBaseId}/folder/${folder.id}?path=${pathParam}`,
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
  const [, setSearchParams] = useSearchParams()
  // 处理文件点击，导航到文件详情页
  const handleFileClick = useCallback(
    (file: FileItem) => {
      setSearchParams((prev) => {
        const newParams = new URLSearchParams(prev)
        newParams.append('preview_id', String(file.id))
        return newParams
      })
    },
    [setSearchParams],
  )

  return {
    getPathFromQuery,
    handleFolderClick,
    handleFileClick,
  }
}
