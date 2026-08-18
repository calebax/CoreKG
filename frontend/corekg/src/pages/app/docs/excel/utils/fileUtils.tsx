import React from 'react'
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
import { SortItem } from '../types'

// 获取文件图标
export const getFileIcon = (
  fileType: string | undefined,
  isFolder: boolean,
) => {
  if (isFolder) {
    return <KnowledgeFile className='w-6 h-6 mr-2' />
  }

  // 如果文件类型为空、未定义或者是 '-'，则不显示图标
  if (!fileType || fileType === '-') {
    return null
  }

  // 转换为小写进行比较
  const lowerFileType = fileType.toLowerCase()

  switch (lowerFileType) {
    case 'doc':
      return <KnowledgeDoc className='w-6 h-6 mr-2' />
    case 'docx':
      return <KnowledgeDocx className='w-6 h-6 mr-2' />
    case 'xls':
      return <KnowledgeXls className='w-6 h-6 mr-2' />
    case 'xlsx':
      return <KnowledgeXlsx className='w-6 h-6 mr-2' />
    case 'xsl':
      return <KnowledgeXsl className='w-6 h-6 mr-2' />
    case 'csv':
      return <KnowledgeCsv className='w-6 h-6 mr-2' />
    case 'ppt':
      return <KnowledgePpt className='w-6 h-6 mr-2' />
    case 'pptx':
      return <KnowledgePptx className='w-6 h-6 mr-2' />
    case 'pdf':
      return <KnowledgePdf className='w-6 h-6 mr-2' />
    case 'png':
    case 'ofd':
      return <KnowledgePng className='w-6 h-6 mr-2' />
    case 'jpg':
      return <KnowledgeJpg className='w-6 h-6 mr-2' />
    case 'jpeg':
      return <KnowledgeJpeg className='w-6 h-6 mr-2' />
    case 'svg':
      return <KnowledgeSvg className='w-6 h-6 mr-2' />
    case 'txt':
      return <KnowledgeTxt className='w-6 h-6 mr-2' />
    case 'md':
      return <KnowledgeMd className='w-6 h-6 mr-2' />
    case 'mp4':
      return <KnowledgeMP4 className='w-6 h-6 mr-2' />
    case 'zip':
      return <KnowledgeZip className='w-6 h-6 mr-2' />
    default:
      return <KnowledgeWord className='w-6 h-6 mr-2' />
  }
}

// 字段名映射表
export const fieldMapping: { [key: string]: string } = {
  name: 'name',
  size: 'size',
  fileType: 'file_type',
  updatedAt: 'created_at',
  parseStatus: 'parse_status',
}

// 构建排序参数数组
export const buildOrderByParams = (sorts: SortItem[]): string[] => {
  if (!sorts || sorts.length === 0) return []

  return sorts.map((item) => {
    const mappedField = fieldMapping[item.field] || item.field
    if (item.order === 'ascend') {
      return `${mappedField}`
    } else {
      return `${mappedField} desc`
    }
  })
}

/**
 * 格式化文件大小显示
 * @param bytes 文件大小（字节）
 * @returns 格式化后的文件大小字符串
 */
export const formatFileSize = (bytes: number): string => {
  // 如果为0或负数，返回 '-'
  if (!bytes || bytes <= 0) {
    return '-'
  }

  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const k = 1024
  const dm = 2 // 保留2位小数

  // 计算对应的单位索引
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  // 防止索引超出范围
  const sizeIndex = Math.min(i, sizes.length - 1)

  // 计算实际大小
  const size = parseFloat((bytes / Math.pow(k, sizeIndex)).toFixed(dm))

  // 如果大小是整数，则不显示小数点
  const displaySize = size % 1 === 0 ? size.toString() : size.toFixed(dm)

  return `${displaySize}${sizes[sizeIndex]}`
}
