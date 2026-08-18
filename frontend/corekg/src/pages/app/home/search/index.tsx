import { FC, ReactNode, useState, useEffect, useMemo } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Skeleton, Tabs, Input } from 'antd'
import { ArrowLeftOutlined, SearchOutlined } from '@ant-design/icons'
import { useRequest, useDebounceFn } from 'ahooks'
import { useTranslation } from 'react-i18next'
import emptyStateIcon from '@/assets/icons/EmptyState.svg'
import backIcon from '@/assets/icons/back.svg'
import searchIcon1 from '@/assets/icons/search-1.svg'
import searchIcon2 from '@/assets/icons/search-2.svg'
import { getCompByType } from '../components/SearchMode/CommonTypeComps'
import {
  SearchType_FnMap,
  SearchType_ResultKeyMap,
  SearchType_TitleMap,
  SearchTypeOrder,
} from '../components/SearchMode/searchType'
import { SearchType } from '../components/SearchMode/searchType'
import AllContent from './all'
import DocContent from './doc'
import { ImageContent } from './image'
import './search-tabs.css'
import styles from './styles.module.scss'
import { VideoContent } from './video'

/**
 * 使用此对象进行搜索
 * text img 二选一
 * type默认 all
 */
export type SearchValue = {
  text?: string
  img?: string
  type?: SearchType
}
const Search: FC = () => {
  const [searchParams] = useSearchParams()
  const search = useMemo(() => {
    try {
      const encodedSearchString = searchParams.get('search')!
      const searchString = decodeURIComponent(encodedSearchString)
      const searchObj = JSON.parse(searchString) as SearchValue
      return searchObj
    } catch {
      return null
    }
  }, [searchParams])
  const navigate = useNavigate()
  if (!search) {
    navigate('/')
    return
  }
  return <SearchResult search={search} />
}
export default Search

const SearchResult: FC<{ search: SearchValue }> = (props) => {
  const { search } = props
  const { text, img } = search
  const [currentSearchText, setCurrentSearchText] = useState(text ?? '')
  const [activeSearchText, setActiveSearchText] = useState(text ?? '')
  const navigate = useNavigate()

  const [type, setType] = useState<SearchType>(() => {
    switch (search.type) {
      case 'all':
      case 'doc':
      case 'image':
      case 'video':
      case 'agent':
      case 'forest':
        return search.type
      default:
        return 'all'
    }
  })

  // 防抖搜索
  const { run: debouncedSearch } = useDebounceFn(
    (searchText: string) => {
      setActiveSearchText(searchText)
    },
    { wait: 500 },
  )

  // 当清除按钮点击时，立即更新activeSearchText
  const handleClear = () => {
    setCurrentSearchText('')
    setActiveSearchText('')
  }

  // 监听搜索框输入变化
  useEffect(() => {
    debouncedSearch(currentSearchText)
  }, [currentSearchText, debouncedSearch])

  const result = useRequest(
    async () => {
      // 当没有搜索文本且没有图片时，不执行搜索
      if (!activeSearchText && !img) {
        return { type, data: null }
      }

      const fn = SearchType_FnMap[type]
      const res = await fn({
        image_url: img,
        text: activeSearchText,
        is_semantics: true,
      })
      const data = (() => {
        if (type === 'all') return res
        return res?.[SearchType_ResultKeyMap[type]] || null
      })()
      return { type, data }
    },
    { refreshDeps: [type, img, activeSearchText] },
  )
  const tabItems: { key: SearchType; label: string }[] = [
    { key: 'all', label: '全部' },
    ...SearchTypeOrder.map((type) => {
      const label = SearchType_TitleMap[type]
      return { key: type, label }
    }),
  ]

  const { t: tC } = useTranslation('common')

  return (
    <div className='h-full flex flex-col bg-white'>
      {/* 顶部搜索栏 */}
      <div className='flex items-center pt-4 px-29.5 bg-white'>
        <button
          onClick={() => navigate('/')}
          className='flex items-center justify-center w-8 h-8 mr-3 hover:bg-white rounded-full hover:cursor-pointer'
        >
          <img
            src={backIcon}
            alt='back'
            className='w-6 h-6 hover:bg-[#E6E8F0] rounded'
          />
        </button>
        <div className='flex-1 max-w-lg'>
          <div
            className={`relative ${currentSearchText ? 'p-[0.9px] rounded-full bg-gradient-to-r from-[#0082FF] to-[#D84CFF]' : ''}`}
          >
            <Input
              value={currentSearchText}
              onChange={(e) => setCurrentSearchText(e.target.value)}
              placeholder={tC('button.search')}
              prefix={
                <img
                  src={currentSearchText ? searchIcon2 : searchIcon1}
                  alt='search'
                  className='w-[18px] h-[18px] relative top-[1px] mr-1'
                />
              }
              suffix={
                currentSearchText && (
                  <span
                    onClick={handleClear}
                    className='text-[#616373] font-normal text-[15px] cursor-pointer'
                  >
                    {tC('button.clear')}
                  </span>
                )
              }
              className={`rounded-full py-[10.5px] px-4 h-[42px] ${
                currentSearchText ? 'border-0' : ''
              }`}
              size='large'
            />
          </div>
        </div>
      </div>

      <div className='flex-1 flex flex-col min-h-0 bg-white'>
        <div className='bg-white px-24 py-4 flex-shrink-0 '>
          <Tabs
            activeKey={type}
            onChange={(val) => {
              if (result.loading) return
              setType(val as SearchType)
            }}
            items={tabItems}
            className='search-tabs-custom px-8'
          />
        </div>
        <div className='flex-1 px-24 min-h-0 rounded-[10px] bg-transparent overflow-hidden'>
          {result.loading || result.error ? (
            <div className='py-4'>
              <Skeleton active />
            </div>
          ) : !activeSearchText && !img ? (
            <div className='flex flex-col items-center justify-center h-full text-center text-gray-500'>
              <img
                src={emptyStateIcon}
                alt={tC('empty.emptyState')}
                className='w-40 h-40 mb-3'
              />
              <p className='text-xl text-[#616373] font-normal'>
                {tC('empty.noFind')}
              </p>
            </div>
          ) : result.data ? (
            <ResultContent
              type={result.data.type}
              setType={setType}
              value={result.data.data}
            />
          ) : (
            <div className='flex flex-col items-center justify-center h-full text-center text-gray-500'>
              <img
                src={emptyStateIcon}
                alt={tC('empty.emptyState')}
                className='w-40 h-40 mb-3'
              />
              <p className='text-xl text-[#616373] font-normal'>
                {tC('empty.noFind')}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

type ResultContent = {
  type: SearchType
  setType: (type: SearchType) => void
  value: any
}
const ResultContent: FC<ResultContent> = (props) => {
  const { type, value, setType } = props
  const withWrapper = (children: ReactNode) => (
    <div
      className={`py-4 h-full overflow-y-auto bg-[#FCFCFE] overflow-x-hidden px-8 flex flex-col gap-1 rounded-[10px] ${styles.scroll}`}
    >
      {children}
    </div>
  )
  if (type === 'all')
    return withWrapper(<AllContent value={value} setType={setType} />)
  if (type === 'doc') return withWrapper(<DocContent value={value} />)
  if (type === 'image') return withWrapper(<ImageContent value={value} />)
  if (type === 'video') return withWrapper(<VideoContent value={value} />)
  const Comp = getCompByType(type)
  return withWrapper(<Comp value={value} />)
}
