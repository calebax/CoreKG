import { FC, PropsWithChildren } from 'react'
import { Modal, Skeleton, Tabs } from 'antd'
import { useRequest } from 'ahooks'
import Moveable from 'react-moveable'
import { cn } from '@/utils'
import {
  forestSearch,
  forestSearchDoc,
  forestSearchImage,
  forestSearchVideo,
} from '@/api/knowledge'
import { ResultType, useSearchInputContext } from '..'
import { ComplexSearchValue } from '../ComplexSearch'
import { ResultContent } from './ResultContent'
import styles from './styles.module.scss'

export type SearchResult = {
  onClose: () => void
}
export type SearchValue = string | ComplexSearchValue
export type SearchResultInstance = {
  startSearch: (value: SearchValue, type?: ResultType) => void
}
export const SearchResult = forwardRef<SearchResultInstance, SearchResult>(
  (props, ref) => {
    const { onClose } = props
    const [open, setModalOpen] = useState(false)
    const [searchValue, setSearchValue] = useState<SearchValue>()
    const [type, setType] = useState<ResultType>('all')
    const { forest_id } = useSearchInputContext()

    const result = useRequest(
      async () => {
        if (!open || !forest_id || !searchValue) return null
        const data = await getData(forest_id, searchValue, type)
        return [type, data] as const
      },
      {
        refreshDeps: [open, forest_id, searchValue, type],
      },
    )
    // useRequest通过useEffect完成自动更新
    // 会导致不能第一时间进入loading
    // 且type与data的数据结构不一致
    // _type始终与数据的类型保持一致
    const [_type, data] = result.data ?? ['all', null]

    const startSearch: SearchResultInstance['startSearch'] = useCallback(
      (value, type = 'all') => {
        setModalOpen(true)
        setType(type)
        setSearchValue(value)
      },
      [setSearchValue, setType],
    )
    useImperativeHandle(ref, () => ({ startSearch }))

    const tabItems: { key: ResultType; label: string }[] = [
      { key: 'all', label: '全部' },
      { key: 'doc', label: '文档' },
      { key: 'image', label: '图片' },
      { key: 'video', label: '视频' },
    ]
    // 内容区域不允许拖拽 弹窗的其他位置可以
    const [draggable, setDraggable] = useState(true)
    return (
      <Modal
        destroyOnHidden
        mask={result.loading}
        closable={!result.loading}
        maskClosable={false}
        keyboard={false}
        className={styles.resultModal}
        width={'60%'}
        footer={null}
        open={open}
        onCancel={() => {
          onClose()
          setModalOpen(false)
        }}
        title={'搜索结果'}
        modalRender={(children) => (
          <DraggableModalContent draggable={!result.loading && draggable}>
            {children}
          </DraggableModalContent>
        )}
      >
        <div
          className='cursor-auto flex flex-col'
          onClick={() => setDraggable(false)}
          onMouseMove={() => setDraggable(false)}
          onMouseOver={() => setDraggable(false)}
          onMouseOut={() => setDraggable(true)}
        >
          <Tabs
            activeKey={type}
            onChange={(val) => {
              if (result.loading) return
              setType(val as ResultType)
            }}
            items={tabItems}
          />
          {result.loading ? (
            <Skeleton active />
          ) : (
            <ResultContent type={_type} setType={setType} value={data} />
          )}
        </div>
      </Modal>
    )
  },
)

type DraggableModalContentProps = PropsWithChildren & {
  draggable: boolean
}
const DraggableModalContent: FC<DraggableModalContentProps> = (props) => {
  const { children, draggable } = props
  const targetRef = useRef<HTMLDivElement>(null)
  return (
    <>
      <div ref={targetRef} className={cn({ 'cursor-move': draggable })}>
        {children}
      </div>
      {draggable && (
        <Moveable
          target={targetRef}
          draggable={true}
          throttleDrag={1}
          onDrag={(e) => {
            e.target.style.transform = e.transform
          }}
        />
      )}
    </>
  )
}

const getData = async (
  forest_id: number,
  searchValue: SearchValue,
  type: ResultType,
) => {
  if (!forest_id || !searchValue) return
  const text =
    typeof searchValue === 'string' ? searchValue : (searchValue.text ?? '')

  const image_url =
    typeof searchValue === 'string' ? undefined : searchValue.img
  const body = { forest_id, is_semantics: true, text, image_url }
  if (type === 'all') {
    return forestSearch(body)
  }
  if (type === 'doc') {
    const { doc_search_result } = await forestSearchDoc(body)
    return doc_search_result
  }
  if (type === 'image') {
    const { image_search_result } = await forestSearchImage(body)
    return image_search_result
  }
  if (type === 'video') {
    const { video_search_result } = await forestSearchVideo(body)
    return video_search_result
  }
  return null
}
