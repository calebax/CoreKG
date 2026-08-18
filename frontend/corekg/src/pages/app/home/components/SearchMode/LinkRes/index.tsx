import { FC, PropsWithChildren } from 'react'
import { Button, Divider, Skeleton } from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { globalSearch } from '@/api/knowledge'
import ArrowRight from '@/assets/icons/arrow-right-search.svg?react'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import { getCompByType } from '../CommonTypeComps'
import {
  SearchType_ResultKeyMap,
  SearchTypeOrder,
  SearchType,
  FileInSearchResult,
  LegalSearchType,
  SearchType_TitleMap,
} from '../searchType'
import { DocLinkItem } from './doc'

export type LinkResProps = {
  img?: string
  text?: string
  /** 点击查看更多 会直接进行搜索 */
  startSearch: (type: SearchType) => void
}
/** 联想结果 */
export const LinkRes: FC<LinkResProps> = (props) => {
  const { img, text, startSearch } = props
  const result = useRequest(
    async () => {
      if (!text) return
      const res = await globalSearch({ text })
      return res
    },
    {
      refreshDeps: [text],

      debounceWait: 300,
    },
  )

  if (result.error || img || !text) {
    return null
  }
  if (result.loading) {
    return (
      <div className='p-4 w-[50vw]'>
        <Skeleton active />
      </div>
    )
  }
  if (!result.data) return null
  const items = SearchTypeOrder.map((type) => {
    const data = result.data[SearchType_ResultKeyMap[type]]
    if (!Array.isArray(data) || data.length === 0) return null
    // 只对智能体、知识库、文档限制显示数量为3个，图片和视频保持原状
    const shouldLimit = ['agent', 'forest', 'doc'].includes(type)
    return {
      type,
      data: shouldLimit
        ? (data as FileInSearchResult[]).slice(0, 3)
        : (data as FileInSearchResult[]),
    }
  })

  // 如果没有有效的搜索结果，直接返回null，不显示边框
  const validItems = items.filter(Boolean)
  if (validItems.length === 0) return null

  return (
    <div
      className={`w-[50vw] overflow-auto flex flex-col bg-white rounded-lg mt-2 border border-[#D2C9FF] border-[0.5px] ${scrollStyles.scroll} shadow-[0_0_0_2px_#F2F0FF]`}
      style={
        window.innerWidth > 1440 ? { maxHeight: '45vh' } : { maxHeight: '40vh' }
      }
    >
      {items.map((item, index) => {
        if (!item) return null
        const { type, data } = item
        const isLast = index === validItems.length - 1

        if (type === 'doc') {
          return (
            <div key={type}>
              <LinkItemWrapper
                type={'doc'}
                onMoreItem={() => startSearch('doc')}
                isLast={isLast}
              >
                <DocLinkItem value={data} />
              </LinkItemWrapper>
              {/* {!isLast && <Divider className='my-0 border-gray-100' />} */}
            </div>
          )
        }
        const Comp = getCompByType(type)
        return (
          <div key={type}>
            <LinkItemWrapper
              type={type}
              onMoreItem={() => startSearch(type)}
              isLast={isLast}
            >
              <Comp value={data} />
            </LinkItemWrapper>
            {/* {!isLast && <Divider className='my-0 border-gray-100' />} */}
          </div>
        )
      })}
    </div>
  )
}

type LinkItemWrapper = {
  /** 判断value的类型 */
  type: LegalSearchType
  onMoreItem: () => void
}
type LinkItemWrapperProps = {
  /** 判断value的类型 */
  type: LegalSearchType
  onMoreItem: () => void
  isLast?: boolean
}
const LinkItemWrapper: FC<PropsWithChildren<LinkItemWrapperProps>> = (
  props,
) => {
  const { t: tC } = useTranslation('common')
  const { type, children, onMoreItem, isLast = false } = props
  return (
    <div className='flex flex-col p-4 pb-0 gap-2.5'>
      <span className='text-[#616373] text-base font-normal'>
        {SearchType_TitleMap[type]}
      </span>
      <div className='space-y-3'>{children}</div>
      <Button
        type='text'
        size='small'
        className='text-[#3D7FFF] !font-normal self-start px-0 text-base hover:shadow-none hover:bg-transparent flex items-center gap-1 mb-2.5'
        onClick={onMoreItem}
      >
        {tC('button.viewMore')}
        <ArrowRight className='w-4 h-4' />
      </Button>
      {!isLast && (
        <div className='border-b border-[#EAEAEA] height-[0.5px]'></div>
      )}
    </div>
  )
}
