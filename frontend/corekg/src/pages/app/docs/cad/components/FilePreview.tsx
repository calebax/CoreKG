import { useState, useEffect } from 'react'
import { lazy } from 'react'
import { useParams, useLocation } from 'react-router-dom'
import { Card, Empty, Spin } from 'antd'
import { getPreviewFileURL } from '@/api/knowledge'
import { useFileLocation } from './FileDetailView/FileLocationContext'
import { withSuspense } from '@/utils/withSuspense'

const MediaViewer = withSuspense(lazy(() => import('@/components/MediaViewer')))
const PDFViewer = withSuspense(lazy(() => import('@/components/PDFViewer')))
const SheetViewer = withSuspense(
  lazy(() => import('@/components/SheetViewer')),
)
const MarkdownPreview = withSuspense(
  lazy(() => import('@/components/common/MarkdownPreview')),
)

interface FilePreviewProps {
  fileName?: string
  fileType?: string
}

// 文件类型映射
const FILE_TYPES = {
  // 文档类型
  pdf: 'pdf',
  pptx: 'pdf',
  ppt: 'pdf',
  docx: 'pdf',
  doc: 'pdf',

  // 表格类型
  xlsx: 'spreadsheet',
  xls: 'spreadsheet',
  csv: 'spreadsheet',

  // 图片类型
  jpg: 'image',
  jpeg: 'image',
  png: 'image',
  webp: 'image',
  gif: 'image',

  mp4: 'video',

  // 文本类型
  md: 'text',
  txt: 'text',
}

export default function FilePreview({
  fileName = '文档预览',
  fileType = '',
}: FilePreviewProps) {
  const { id, fileId } = useParams<{ id: string; fileId: string }>()
  const location = useLocation()
  const [fileUrl, setFileUrl] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [detectedFileType, setDetectedFileType] = useState<string>('')
  const fileLocation = useFileLocation()
  useEffect(() => {
    const queryParams = new URLSearchParams(location.search)
    const urlFileType = queryParams.get('fileType')

    let finalFileType = urlFileType || fileType || ''

    // 如果有文件名但没有确定文件类型，尝试从文件名推断
    if (!finalFileType && fileName) {
      const extension = fileName.split('.').pop()?.toLowerCase()
      if (extension) {
        finalFileType = extension
      }
    }

    // 将具体文件类型转换为文件类别
    const fileCategory =
      FILE_TYPES[finalFileType.toLowerCase() as keyof typeof FILE_TYPES] ||
      finalFileType

    setDetectedFileType(fileCategory)
  }, [location.search, fileType, fileName])

  const fetchFileUrl = async () => {
    if (!id || !fileId) return

    setLoading(true)
    setError(null)
    try {
      const res = await getPreviewFileURL({
        file_id: Number(fileId),
      })
      console.log('获取到的文件URL:', res.url)
      setFileUrl(res.url)
    } catch (error) {
      console.error('获取文件预览URL失败:', error)
      setError(error instanceof Error ? error.message : '获取文件预览URL失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchFileUrl()
  }, [id, fileId])

  // 根据文件类型选择适当的预览组件
  const renderFilePreview = () => {
    if (!fileUrl) return null

    switch (detectedFileType) {
      case 'pdf':
        return (
          <PDFViewer
            defaultPage={fileLocation ? fileLocation[0] + 1 : undefined}
            file={fileUrl}
            id={fileId!}
            name={fileName}
          />
        )
      case 'spreadsheet':
        return <SheetViewer file={fileUrl} />
      case 'image':
        return <MediaViewer file={fileUrl} />
      case 'video':
        return (
          <MediaViewer file={fileUrl} isVideo={true} location={fileLocation} />
        )
      case 'text':
        return <MarkdownPreview file={fileUrl} />
      default:
        return (
          <div className='h-full flex items-center justify-center'>
            <div className='text-center'>
              <p className='text-gray-500 mb-2 text-base'>不支持的文件类型</p>
              <p className='text-sm text-gray-400'>
                当前仅支持PDF、Excel表格、图片和视频格式
              </p>
            </div>
          </div>
        )
    }
  }

  return (
    <div className='h-full flex flex-col overflow-hidden'>
      {/* 预览区域 */}
      <Card
        className='h-full !rounded shadow-none border-0'
        styles={{
          body: {
            height: '100%',
            padding: '16px',
            overflow: 'hidden',
          },
        }}
      >
        <div className='h-full overflow-auto custom-preview-scroll'>
          {loading ? (
            <div className='h-full flex items-center justify-center'>
              <Spin size='large' />
            </div>
          ) : error ? (
            <div className='h-full flex items-center justify-center'>
              <div className='text-center'>
                <p className='text-red-500 mb-2 text-base'>获取文件失败</p>
                <p className='text-sm text-gray-400'>{error}</p>
                <p className='text-xs text-gray-300 mt-2'>
                  请检查接口参数或网络连接
                </p>
              </div>
            </div>
          ) : fileUrl ? (
            <div className='h-full overflow-auto'>{renderFilePreview()}</div>
          ) : (
            <div className='h-full flex items-center justify-center'>
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={
                  <div className='text-center'>
                    <p className='text-gray-500 mb-2 text-base'>文件预览区域</p>
                    <p className='text-sm text-gray-400'>暂无可预览的文件</p>
                    <p className='text-xs text-gray-300 mt-2'>
                      支持PDF、Excel表格和图片格式
                    </p>
                  </div>
                }
              />
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}
