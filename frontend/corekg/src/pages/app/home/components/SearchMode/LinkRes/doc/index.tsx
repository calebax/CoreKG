import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { Divider } from 'antd'
// 导入各种文件类型图标
import KnowledgeCsv from '@/assets/icons/knowledge-csv.svg'
import KnowledgeDoc from '@/assets/icons/knowledge-doc.svg'
import KnowledgeDocx from '@/assets/icons/knowledge-docx.svg'
import KnowledgeJpeg from '@/assets/icons/knowledge-jpeg.svg'
import KnowledgeJpg from '@/assets/icons/knowledge-jpg.svg'
import KnowledgeMd from '@/assets/icons/knowledge-md.svg'
import KnowledgeMP4 from '@/assets/icons/knowledge-mp4.svg'
import KnowledgePdf from '@/assets/icons/knowledge-pdf.svg'
import KnowledgePng from '@/assets/icons/knowledge-png.svg'
import KnowledgePpt from '@/assets/icons/knowledge-ppt.svg'
import KnowledgePptx from '@/assets/icons/knowledge-pptx.svg'
import KnowledgeSvg from '@/assets/icons/knowledge-svg.svg'
import KnowledgeTxt from '@/assets/icons/knowledge-txt.svg'
import KnowledgeWord from '@/assets/icons/knowledge-word.svg'
import KnowledgeXls from '@/assets/icons/knowledge-xls.svg'
import KnowledgeXlsx from '@/assets/icons/knowledge-xlsx.svg'
import KnowledgeXsl from '@/assets/icons/knowledge-xsl.svg'
import KnowledgeZip from '@/assets/icons/knowledge-zip.svg'
import { getFileURL } from '@/utils/Forest'
import { Title } from '../../Title'
import { FileInSearchResult } from '../../searchType'

type DocLinkItem = {
  value: FileInSearchResult[]
}

// 从文件名提取文件扩展名
const getFileExtension = (fileName: string): string => {
  const lastDotIndex = fileName.lastIndexOf('.')
  if (lastDotIndex === -1) return ''
  return fileName.substring(lastDotIndex + 1).toLowerCase()
}

// 根据文件类型获取图标URL
const getFileIconByType = (fileName: string): string => {
  const fileType = getFileExtension(fileName)

  switch (fileType) {
    case 'docx':
      return KnowledgeDocx
    case 'doc':
      return KnowledgeDoc
    case 'xlsx':
      return KnowledgeXlsx
    case 'xls':
      return KnowledgeXls
    case 'xsl':
      return KnowledgeXsl
    case 'csv':
      return KnowledgeCsv
    case 'pptx':
      return KnowledgePptx
    case 'ppt':
      return KnowledgePpt
    case 'pdf':
      return KnowledgePdf
    case 'ofd':
    case 'png':
      return KnowledgePng
    case 'jpg':
      return KnowledgeJpg
    case 'jpeg':
      return KnowledgeJpeg
    case 'svg':
      return KnowledgeSvg
    case 'txt':
      return KnowledgeTxt
    case 'md':
      return KnowledgeMd
    case 'mp4':
      return KnowledgeMP4
    case 'zip':
      return KnowledgeZip
    default:
      return KnowledgeWord // 默认使用原来的word图标
  }
}

export const DocLinkItem: FC<DocLinkItem> = (props) => {
  const { value } = props

  return (
    <>
      {value.map((file, index) => {
        const { highlighted_file_name, id, forest_id, highlights } = file
        const { highlighted_description } = highlights![0]
        const url = getFileURL(forest_id, id)

        // 从高亮的文件名中提取纯文本文件名（去除HTML标签）
        const plainFileName =
          highlighted_file_name?.replace(/<[^>]*>/g, '') || ''
        const fileIcon = getFileIconByType(plainFileName)

        return (
          <div key={url}>
            <Link
              to={url}
              target='_blank'
              className='text-[unset] block hover:bg-[#EFF0F6] rounded-[10px] p-2.5 transition-colors'
            >
              <Title
                image={fileIcon}
                name={highlighted_file_name!}
                desc={highlighted_description}
              />
            </Link>
            {/* {index < value.length - 1 && <Divider className='my-3' />} */}
          </div>
        )
      })}
    </>
  )
}
