import KnowledgeCsv from '@/assets/icons/knowledge-csv.svg?react'
import KnowledgeDoc from '@/assets/icons/knowledge-doc.svg?react'
import KnowledgeDocx from '@/assets/icons/knowledge-docx.svg?react'
import KnowledgeFile from '@/assets/icons/knowledge-files.svg?react'
import KnowledgeJpeg from '@/assets/icons/knowledge-jpeg.svg?react'
import KnowledgeJpg from '@/assets/icons/knowledge-jpg.svg?react'
import KnowledgeMd from '@/assets/icons/knowledge-md.svg?react'
import KnowledgeMP4 from '@/assets/icons/knowledge-mp4.svg?react'
import KnowledgePdf from '@/assets/icons/knowledge-pdf.svg?react'
import KnowledgePng from '@/assets/icons/knowledge-png.svg?react'
import KnowledgePpt from '@/assets/icons/knowledge-ppt.svg?react'
import KnowledgePptx from '@/assets/icons/knowledge-pptx.svg?react'
import KnowledgeSvg from '@/assets/icons/knowledge-svg.svg?react'
import KnowledgeTxt from '@/assets/icons/knowledge-txt.svg?react'
import KnowledgeWord from '@/assets/icons/knowledge-word.svg?react'
import KnowledgeXls from '@/assets/icons/knowledge-xls.svg?react'
import KnowledgeXlsx from '@/assets/icons/knowledge-xlsx.svg?react'
import KnowledgeXsl from '@/assets/icons/knowledge-xsl.svg?react'
import KnowledgeZip from '@/assets/icons/knowledge-zip.svg?react'

// 获取文件图标
export const getFileIcon = (fileName: string | undefined) => {
  // 如果文件类型为空、未定义或者是 '-'，则不显示图标
  if (!fileName || fileName === '-') {
    return null
  }

  const lastDotIndex = fileName.lastIndexOf('.')
  if (lastDotIndex === -1 || lastDotIndex === fileName.length - 1) {
    return ''
  }
  const fileType = fileName.substring(lastDotIndex + 1)

  // 转换为小写进行比较
  const lowerFileType = fileType.toLowerCase()
  const className = 'w-4 h-4 mr-2'
  switch (lowerFileType) {
    case 'doc':
      return <KnowledgeDoc className={className} />
    case 'docx':
      return <KnowledgeDocx className={className} />
    case 'xls':
      return <KnowledgeXls className={className} />
    case 'xlsx':
      return <KnowledgeXlsx className={className} />
    case 'xsl':
      return <KnowledgeXsl className={className} />
    case 'csv':
      return <KnowledgeCsv className={className} />
    case 'ppt':
      return <KnowledgePpt className={className} />
    case 'pptx':
      return <KnowledgePptx className={className} />
    case 'pdf':
      return <KnowledgePdf className={className} />
    case 'png':
    case 'ofd':
      return <KnowledgePng className={className} />
    case 'jpg':
      return <KnowledgeJpg className={className} />
    case 'jpeg':
      return <KnowledgeJpeg className={className} />
    case 'svg':
      return <KnowledgeSvg className={className} />
    case 'txt':
      return <KnowledgeTxt className={className} />
    case 'md':
      return <KnowledgeMd className={className} />
    case 'mp4':
      return <KnowledgeMP4 className={className} />
    case 'zip':
      return <KnowledgeZip className={className} />
    default:
      return <KnowledgeWord className={className} />
  }
}
