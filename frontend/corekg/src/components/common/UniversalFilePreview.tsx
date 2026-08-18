import { lazy, FC, useMemo } from 'react'
import { Empty, Spin } from 'antd'
import { withSuspense } from '@/utils/withSuspense'

const MediaViewer = withSuspense(lazy(() => import('@/components/MediaViewer')))
const PDFViewer = withSuspense(lazy(() => import('@/components/PDFViewer')))
const SheetViewer = withSuspense(lazy(() => import('@/components/SheetViewer')))
const MarkdownPreview = withSuspense(
  lazy(() => import('@/components/common/MarkdownPreview')),
)

interface UniversalFilePreviewProps {
  id?: string
  url: string
  fileName: string
  fileType?: string
  loading?: boolean
  error?: string | null
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

export const UniversalFilePreview: FC<UniversalFilePreviewProps> = ({
  id,
  url,
  fileName,
  fileType,
  loading = false,
  error = null,
}) => {
  const detectedFileType = useMemo(() => {
    let finalFileType = fileType || ''

    if (!finalFileType && fileName) {
      const extension = fileName.split('.').pop()?.toLowerCase()
      if (extension) {
        finalFileType = extension
      }
    }

    return (
      FILE_TYPES[finalFileType.toLowerCase() as keyof typeof FILE_TYPES] ||
      finalFileType
    )
  }, [fileType, fileName])

  // 根据文件类型选择适当的预览组件
  const renderFilePreview = () => {
    if (!url) return null
    switch (detectedFileType) {
      case 'pdf':
        return <PDFViewer file={url} name={fileName} id={id} />
      case 'spreadsheet':
        return <SheetViewer file={url} />
      case 'image':
        return <MediaViewer file={url} />
      case 'video':
        return <MediaViewer file={url} isVideo={true} />
      case 'text':
        return <MarkdownPreview file={url} />
      default:
        return (
          <div className='h-full flex items-center justify-center'>
            <div className='text-center'>
              <p className='text-gray-500 mb-2 text-base'>不支持预览该文件类型</p>
              <p className='text-sm text-gray-400'>
                当前支持PDF、Excel表格、图片、视频和文本格式
              </p>
              <a
                href={url}
                target='_blank'
                rel='noreferrer'
                className='text-blue-500 mt-4 inline-block'
              >
                直接下载文件
              </a>
            </div>
          </div>
        )
    }
  }

  return (
    <div className='h-full flex flex-col overflow-hidden bg-white'>
      <div className='h-full overflow-auto custom-preview-scroll'>
        {loading ? (
          <div className='h-full flex items-center justify-center'>
            <Spin size='large' />
          </div>
        ) : error ? (
          <div className='h-full flex items-center justify-center'>
            <div className='text-center'>
              <p className='text-red-500 mb-2 text-base'>加载文件失败</p>
              <p className='text-sm text-gray-400'>{error}</p>
            </div>
          </div>
        ) : url ? (
          <div className='h-full overflow-auto'>{renderFilePreview()}</div>
        ) : (
          <div className='h-full flex items-center justify-center'>
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description='暂无可预览的内容'
            />
          </div>
        )}
      </div>
    </div>
  )
}

