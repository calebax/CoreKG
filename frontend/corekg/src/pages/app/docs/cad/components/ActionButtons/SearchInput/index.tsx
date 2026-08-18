import { createContext, FC } from 'react'
import { Input, InputProps, Popover } from 'antd'
import { DownOutlined, UpOutlined } from '@ant-design/icons'
import { useBoolean, useClickAway } from 'ahooks'
import KnowledgeSearch from '@/assets/icons/knowledge-search.svg?react'
import { ComplexSearch, ComplexSearchValue } from './ComplexSearch'
import { LinkRes } from './LinkRes'
import { SearchResult, SearchResultInstance } from './SearchResult'

/** 联想和结果页面的不同数据类型 */
export type ResultType = 'all' | 'doc' | 'image' | 'video'

/** 搜索结果 单个文件 */
export type FileInSearchResult = {
  id: number
  forest_id: number
  highlighted_file_name: string
  /** 同一文件的多个搜索结果 */
  highlights: {
    highlighted_description: string
    image_url?: string
    location: number[]
  }[]
}

type SearchInput = {
  forest_id: number
  onSearch?: (data: { value: string; image_url?: string }) => void
  onChange?: (val: string) => void
}
/**
 * 全局搜索组件\
 * 交互逻辑:
 * - 使用input普通搜索
 * - 通过input右侧的按钮 可以展开和收起复杂搜索(ComplexSearch)
 * - 普通搜索时 需要进行联想(LinkRes)
 * - 点击联想区域外部后 不展示联想
 * - 搜索结果弹窗展示 并关闭input下拉面板
 */
export const SearchInput: FC<SearchInput> = (props) => {
  const { forest_id } = props

  const width = '264px'

  const [search, setSearch] = useState('')
  const [complexSearchValue, setComplexSearchValue] =
    useState<ComplexSearchValue>({})
  const { isComplexSearch, toggleComplexSearch, popoverOpen, setOpen } =
    useSearchProp(search)
  const popoverContentRef = useRef<HTMLDivElement>(null)
  useClickAway(() => {
    setOpen(false)
    // 重置箭头状态
    if (isComplexSearch) {
      toggleComplexSearch()
    }
  }, popoverContentRef)

  const { inputValue, setInputValue, compositionHandlers } = useComposition(
    (val) => {
      setSearch(val)
      props.onChange?.(val)
      if (!isComplexSearch) {
        setOpen(true)
      }
    },
  )

  const resultRef = useRef<SearchResultInstance | null>(null)
  const onCommonSearch = (type: ResultType = 'all') => {
    setOpen(false)
    props.onSearch?.({ value: search })
    resultRef.current?.startSearch(search, type)
  }
  const onComplexSearch = (value: ComplexSearchValue) => {
    setOpen(false)
    props.onSearch?.({
      value: value.text ?? '',
      image_url: value.img,
    })
    resultRef.current?.startSearch(value)
  }
  return (
    <SearchInputContext.Provider value={{ forest_id }}>
      <Popover
        content={
          <div style={{ width }} ref={popoverContentRef}>
            <LinkRes
              value={search}
              startSearch={onCommonSearch}
              hidden={isComplexSearch}
            ></LinkRes>
            <ComplexSearch
              value={complexSearchValue}
              onChange={setComplexSearchValue}
              onSearch={onComplexSearch}
              hidden={!isComplexSearch}
            ></ComplexSearch>
          </div>
        }
        open={popoverOpen}
        placement='bottom'
        arrow={false}
      >
        <div
          // 点击input不会收起联想
          onClick={(e) => {
            e.stopPropagation()
          }}
        >
          <Input
            value={inputValue}
            style={{ width }}
            size='large'
            placeholder='搜索'
            prefix={
              <KnowledgeSearch
                className='cursor-pointer'
                onClick={() => onCommonSearch()}
              />
            }
            suffix={
              isComplexSearch ? (
                <UpOutlined onClick={toggleComplexSearch} />
              ) : (
                <DownOutlined onClick={toggleComplexSearch} />
              )
            }
            onPressEnter={() => onCommonSearch()}
            {...compositionHandlers}
            // 聚焦input时展示联想
            onFocus={() => setOpen(true)}
          />
        </div>
      </Popover>
      <SearchResult
        ref={resultRef}
        onClose={() => {
          setSearch('')
          setInputValue('')
          setComplexSearchValue({})
          props.onSearch?.({ value: '' })
        }}
      />
    </SearchInputContext.Provider>
  )
}

/**
 * Popover相关的逻辑
 * @param search 当前搜索框内的值
 */
const useSearchProp = (search: string | undefined) => {
  /** 是否复杂检索 */
  const [isComplexSearch, { toggle }] = useBoolean(false)
  /** 手动控制popover是否展开 */
  const [open, setOpen] = useState(false)
  const popoverOpen = useMemo<boolean>(() => {
    if (!open) return false
    if (isComplexSearch) return true
    return Boolean(search)
  }, [isComplexSearch, open, search])
  const toggleComplexSearch = useCallback(() => {
    toggle()
    setOpen(true)
  }, [toggle])
  return {
    isComplexSearch,
    toggleComplexSearch,
    popoverOpen,
    setOpen,
  }
}

/**
 * 处理输入法合成问题
 */
function useComposition(onComposedChange: (val: string) => void) {
  const [inputValue, setInputValue] = useState('')
  const isComposing = useRef(false)
  const onCompositionStart: InputProps['onCompositionStart'] = () => {
    isComposing.current = true
  }
  const onCompositionEnd: InputProps['onCompositionEnd'] = (e) => {
    isComposing.current = false
    onComposedChange((e.target as HTMLInputElement).value)
  }
  const onChange: InputProps['onChange'] = (e) => {
    setInputValue(e.target.value)
    if (isComposing.current) return
    onComposedChange(e.target.value)
  }
  const compositionHandlers = { onCompositionStart, onCompositionEnd, onChange }
  return { inputValue, setInputValue, compositionHandlers }
}

type SearchInputContextValue = {
  forest_id?: number
}
const SearchInputContext = createContext<SearchInputContextValue>({})
// eslint-disable-next-line react-refresh/only-export-components
export const useSearchInputContext = () => {
  return useContext(SearchInputContext)
}
