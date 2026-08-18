import { FC, Fragment, useMemo } from 'react'
import { Link } from 'react-router-dom'
// import { Button } from 'antd'
// import { useBoolean } from 'ahooks'
// import { ChevronDownIcon, ChevronUpIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import KnowledgeCsvBig from '@/assets/icons/knowledge-csv-big.svg'
import KnowledgeDocBig from '@/assets/icons/knowledge-doc-big.svg'
import KnowledgeDocxBig from '@/assets/icons/knowledge-docx-big.svg'
import KnowledgeJpegBig from '@/assets/icons/knowledge-jpeg-big.svg'
import KnowledgeJpgBig from '@/assets/icons/knowledge-jpg-big.svg'
import KnowledgeMdBig from '@/assets/icons/knowledge-md-big.svg'
import KnowledgeMP4Big from '@/assets/icons/knowledge-mp4-big.svg'
import KnowledgePdfBig from '@/assets/icons/knowledge-pdf-big.svg'
import KnowledgePngBig from '@/assets/icons/knowledge-png-big.svg'
import KnowledgePptBig from '@/assets/icons/knowledge-ppt-big.svg'
import KnowledgePptxBig from '@/assets/icons/knowledge-pptx-big.svg'
import KnowledgeTxtBig from '@/assets/icons/knowledge-txt-big.svg'
import word from '@/assets/icons/knowledge-word.svg'
import KnowledgeXlsBig from '@/assets/icons/knowledge-xls-big.svg'
import KnowledgeXlsxBig from '@/assets/icons/knowledge-xlsx-big.svg'
import KnowledgeXslBig from '@/assets/icons/knowledge-xsl-big.svg'
import KnowledgeZipBig from '@/assets/icons/knowledge-zip-big.svg'
import { getFileURL } from '@/utils/Forest'
import { FileInSearchResult } from '../../components/SearchMode/searchType'
import EmptyState from '../components/EmptyState'

type Content = {
  value?: FileInSearchResult[]
}
const DocContent: FC<Content> = (props) => {
  const { value } = props
  if (!value || value.length === 0)
    return <EmptyState message='暂未查询到相关内容～' />
  return (
    <div className='flex flex-col'>
      {value.map((file, index) => {
        return (
          <div key={file.id}>
            <DocItem value={file} />
            {index < value.length - 1 && (
              <div className='border-b border-[#EAEAEA] height-[0.5px] mb-5'></div>
            )}
          </div>
        )
      })}
    </div>
  )
}

// 获取文件大图标
const getFileBigIcon = (fileName: string): string => {
  // 从文件名中提取扩展名
  const fileExtension = fileName.split('.').pop()?.toLowerCase()

  if (!fileExtension) {
    return word // 默认返回 word 图标
  }

  switch (fileExtension) {
    case 'csv':
      return KnowledgeCsvBig
    case 'doc':
      return KnowledgeDocBig
    case 'docx':
      return KnowledgeDocxBig
    case 'xls':
      return KnowledgeXlsBig
    case 'xlsx':
      return KnowledgeXlsxBig
    case 'xsl':
      return KnowledgeXslBig
    case 'ppt':
      return KnowledgePptBig
    case 'pptx':
      return KnowledgePptxBig
    case 'pdf':
      return KnowledgePdfBig
    case 'png':
      return KnowledgePngBig
    case 'jpg':
      return KnowledgeJpgBig
    case 'jpeg':
      return KnowledgeJpegBig
    case 'txt':
      return KnowledgeTxtBig
    case 'md':
      return KnowledgeMdBig
    case 'mp4':
      return KnowledgeMP4Big
    case 'zip':
      return KnowledgeZipBig
    default:
      return word // 默认返回 word 图标
  }
}

export default DocContent

const DocItem: FC<{ value: FileInSearchResult }> = (props) => {
  const { forest_id, highlights, id, highlighted_file_name } = props.value
  const url = getFileURL(forest_id, id)
  // 展开收起功能相关代码（暂时注释）
  // const [hidden, { toggle }] = useBoolean(true)
  // const showItems = useMemo(() => {
  //   return hidden ? highlights!.slice(0, 3) : highlights!
  // }, [hidden, highlights])
  return (
    <div className='flex gap-3 py-2.5 px-[5px]'>
      {/* 左侧文档图标 */}
      <div className='flex-shrink-0'>
        <Link to={url} target='_blank' className={cn('text-[unset]')}>
          <img
            src={getFileBigIcon(highlighted_file_name!)}
            alt={highlighted_file_name!}
            className='w-12 h-12'
          />
        </Link>
      </div>

      {/* 右侧内容区域 */}
      <div className='flex-1 flex flex-col'>
        {/* 上半部分：文档标题 */}
        <div>
          <Link to={url} target='_blank' className={cn('text-[unset]')}>
            <h3
              className='text-lg font-medium text-[#1E1F28] hover:text-[#0C99FF] line-clamp-2 leading-[26px] py-2'
              dangerouslySetInnerHTML={{
                __html: highlighted_file_name!,
              }}
            />
          </Link>
        </div>

        {/* 下半部分：文档内容 */}
        <div className=''>
          {/* {showItems!.map((item, index) => { */}
          {highlights!.slice(0, 1).map((item, index) => {
            const { highlighted_description, location } = item
            const urlWithFileLocation = getFileURL(forest_id, id, location)
            return (
              <Fragment key={highlighted_description}>
                <Link
                  to={urlWithFileLocation}
                  target='_blank'
                  className={cn('text-[unset]')}
                >
                  <span
                    dangerouslySetInnerHTML={{
                      __html: highlighted_description,
                    }}
                    className={cn(
                      'text-base text-[#616373] line-clamp-5 overflow-hidden mb-2 text-ellipsis block leading-[24px] py-2',
                    )}
                  ></span>
                </Link>
              </Fragment>
            )
          })}
        </div>

        {/* 展开收起按钮（暂时注释） */}
        {/* {highlights!.length > 3 && (
          <div
            className='flex items-center gap-1 cursor-pointer self-start'
            onClick={toggle}
          >
            <span className='text-[#3d7fff] text-base font-normal leading-6'>
              {hidden ? '展开' : '收起'}
            </span>
            <div className='flex items-center justify-center w-[15px] h-[15px]'>
              {hidden ? (
                <ChevronDownIcon className='text-[#3d7fff] text-[15px]' />
              ) : (
                <ChevronUpIcon className='text-[#3d7fff] text-[15px]' />
              )}
            </div>
          </div>
        )} */}
      </div>
    </div>
  )
}
