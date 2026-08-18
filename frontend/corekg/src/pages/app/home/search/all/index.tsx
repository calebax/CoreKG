import { FC, Fragment, PropsWithChildren, useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Button } from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
// 移除不再使用的导入
// import { ChevronDownIcon, ChevronUpIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import ArrowRight from '@/assets/icons/arrow-right-search.svg?react'
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
import { getCompByType } from '../../components/SearchMode/CommonTypeComps'
import { Title } from '../../components/SearchMode/Title'
import {
  LegalSearchType,
  SearchType_ResultKeyMap,
  SearchType_TitleMap,
  SearchTypeOrder,
} from '../../components/SearchMode/searchType'
import { FileInSearchResult } from '../../components/SearchMode/searchType'
import { ImageContent } from '../image'
import { VideoContent } from '../video'

type AllContent = {
  setType: (val: LegalSearchType) => void
  value: any
}
// 自定义Hook：检测屏幕尺寸以确定每行显示数量
const useItemsPerRow = () => {
  const [itemsPerRow, setItemsPerRow] = useState(2) // 默认小屏显示2个

  useEffect(() => {
    const updateItemsPerRow = () => {
      const width = window.innerWidth
      if (width >= 1536) {
        // 2xl断点
        setItemsPerRow(6)
      } else if (width >= 768) {
        // md断点
        setItemsPerRow(4)
      } else {
        setItemsPerRow(2) // 小屏幕
      }
    }

    updateItemsPerRow()
    window.addEventListener('resize', updateItemsPerRow)
    return () => window.removeEventListener('resize', updateItemsPerRow)
  }, [])

  return itemsPerRow
}

export const AllContent: FC<AllContent> = (props) => {
  const { setType } = props
  const { t } = useTranslation('common')
  const itemsPerRow = useItemsPerRow()

  const items: { value: any; title: string; type: LegalSearchType }[] =
    SearchTypeOrder.map((type) => {
      const title = SearchType_TitleMap[type]
      const value = props.value?.[SearchType_ResultKeyMap[type]] || []
      return { value, title, type }
    })

  // 过滤出有数据的项目
  const validItems = items.filter(
    (item) => Array.isArray(item.value) && item.value.length,
  )

  return (
    <div className='flex flex-col'>
      {items.map((item, index) => {
        const { value, title, type } = item
        if (!Array.isArray(value) || !value.length) return null

        // 计算当前项目在有效项目中的索引
        const validIndex = validItems.findIndex(
          (validItem) => validItem.type === type,
        )
        const moreBtn = (
          <Button
            type='text'
            size='small'
            className='text-[#3D7FFF] !font-normal self-start px-0 text-base hover:shadow-none hover:bg-transparent flex items-center gap-1 mb-2.5'
            onClick={() => setType(type)}
          >
            {t('button.viewMore')}
            <ArrowRight className='w-4 h-4' />
          </Button>
        )
        if (type === 'doc') {
          return (
            <div key={type}>
              <div className='flex flex-col gap-2.5 py-0'>
                <span className='text-[#616373] text-base font-normal'>
                  {title}
                </span>
                <DocContentInAny value={value.slice(0, 3)}>
                  {moreBtn}
                </DocContentInAny>
              </div>
              {validIndex < validItems.length - 1 && (
                <div className='border-b border-[#EAEAEA] height-[0.5px] mb-4'></div>
              )}
            </div>
          )
        }
        if (type === 'image') {
          return (
            <div key={type}>
              <div className='flex flex-col gap-2.5 py-0'>
                <span className='text-[#616373] text-base font-normal'>
                  {title}
                </span>
                <div>
                  <ImageContent value={value} maxItems={itemsPerRow} />
                </div>
                {moreBtn}
              </div>
              {validIndex < validItems.length - 1 && (
                <div className='border-b border-[#EAEAEA] height-[0.5px] mb-4'></div>
              )}
            </div>
          )
        }
        if (type === 'video') {
          return (
            <div key={type}>
              <div className='flex flex-col gap-2.5 py-0'>
                <span className='text-[#616373] text-base font-normal'>
                  {title}
                </span>
                <div>
                  <VideoContent value={value} maxItems={itemsPerRow} />
                </div>
                {moreBtn}
              </div>
              {validIndex < validItems.length - 1 && (
                <div className='border-b border-[#EAEAEA] height-[0.5px] mb-4'></div>
              )}
            </div>
          )
        }
        const Comp = getCompByType(type)
        return (
          <div key={type}>
            <div className='flex flex-col gap-2.5 py-0'>
              <span className='text-[#616373] text-base font-normal'>
                {title}
              </span>
              <Comp value={value.slice(0, 3)}></Comp>
              {moreBtn}
            </div>
            {validIndex < validItems.length - 1 && (
              <div className='border-b border-[#EAEAEA] height-[0.5px] mb-4'></div>
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

export default AllContent

const DocItem: FC<{ value: FileInSearchResult }> = (props) => {
  const { forest_id, highlights, id, highlighted_file_name } = props.value
  const url = getFileURL(forest_id, id)
  // 移除展开收起状态管理
  // const [hidden, { toggle }] = useBoolean(true)
  // 始终只展示第一段内容
  const showItems = highlights!.slice(0, 1)
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
          {showItems!.map((item, index) => {
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

        {/* 展开/收起按钮 */}
        {/* <div
          className={cn(
            'flex items-center gap-1 cursor-pointer self-start mt-2',
            {
              hidden: highlights!.length <= 1,
            },
          )}
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
        </div> */}
      </div>
    </div>
  )
}

type Content = PropsWithChildren & {
  value: FileInSearchResult[]
}
const DocContentInAny: FC<Content> = (props) => {
  const { value, children } = props
  return (
    <div className='flex flex-col gap-2.5'>
      {value.map((file) => {
        return (
          <div key={file.id}>
            <DocItem value={file} />
          </div>
        )
      })}
      {children}
    </div>
  )
}
