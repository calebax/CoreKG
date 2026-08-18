import React, {
  useEffect,
  useState,
  useRef,
  useCallback,
  useImperativeHandle,
  forwardRef,
  useMemo,
} from 'react'
import { Button, Spin, Input } from 'antd'
import { LeftOutlined, RightOutlined } from '@ant-design/icons'
import { Document, Page, pdfjs } from 'react-pdf'
import 'react-pdf/dist/esm/Page/AnnotationLayer.css'
import 'react-pdf/dist/esm/Page/TextLayer.css'
import { cn } from '@/utils'
import DownloadIcon from '@/assets/icons/docs/download.svg?react'
import MagnifyIcon from '@/assets/icons/docs/magnify.svg?react'
import ReduceIcon from '@/assets/icons/docs/reduce.svg?react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { DownLoadBtn } from './DownloadBtn'

// 使用 public 目录下的 legacy worker（4.8.69，内置 polyfill），兼容旧版浏览器
pdfjs.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.js'

// 手动跳转后的防抖时长（覆盖滚动动画）
const MANUAL_CHANGE_DURATION = 600

// A4 在 200 DPI 下的像素基准，用于分段高亮坐标计算
const LOC_BASE = { WIDTH: 1654, HEIGHT: 2338 }

// 分段高亮块：在 PDF 页面上显示 chunk 区域，支持选中/悬浮态
const SegmentHighlightBlock = ({
  segment,
  isActive,
  pageIndex,
  onChunkClick,
}: {
  segment: {
    id: string
    location?: number[]
    chunk_number?: number
    sequence?: number
  }
  isActive: boolean
  pageIndex: number
  onChunkClick?: (chunkId: string) => void
}) => {
  const loc = segment.location
  if (!loc || loc.length < 5) return null
  const [, x1, y1, x2, y2] = loc
  const isEven =
    (segment.chunk_number || segment.sequence || pageIndex + 1) % 2 === 0
  return (
    <div
      className={cn(
        'absolute transition-all duration-200 cursor-pointer group border',
        isActive &&
          'bg-[#CC5DE84D] border-[#CC5DE8] z-10 shadow-[0_0_8px_rgba(204,93,232,0.3)]',
        !isActive && 'hover:bg-[#CC5DE84D] hover:border-[#CC5DE8] z-10',
        !isActive && isEven && 'bg-[#E1E8FF4D] border-[#E1E8FF]',
        !isActive && !isEven && 'bg-[#97DDFF80] border-[#97DDFF]',
      )}
      style={{
        left: `${(x1 / LOC_BASE.WIDTH) * 100}%`,
        top: `${(y1 / LOC_BASE.HEIGHT) * 100}%`,
        width: `${(x2 / LOC_BASE.WIDTH) * 100}%`,
        height: `${(y2 / LOC_BASE.HEIGHT) * 100}%`,
      }}
      onClick={(e) => {
        e.stopPropagation()
        onChunkClick?.(segment.id)
      }}
    >
      <div className='hidden group-hover:block absolute -top-5 left-0 text-white text-[10px] px-1 rounded whitespace-nowrap pointer-events-none bg-[#CC5DE8]'>
        分段 {segment.sequence ?? segment.chunk_number}
      </div>
    </div>
  )
}

// 渲染 Page + 加载遮罩，不添加任何 layout 包装，overlay 用 absolute 覆盖父级
const PageWithLoadingOverlay = ({
  pageNumber,
  scale,
  onRendered,
}: {
  pageNumber: number
  scale: number
  onRendered?: (pageNum: number) => void
}) => {
  const [rendered, setRendered] = useState(false)
  return (
    <>
      <Page
        pageNumber={pageNumber}
        scale={scale}
        loading={null}
        onRenderSuccess={() => {
          setRendered(true)
          onRendered?.(pageNumber)
        }}
      />
      {!rendered && (
        <div
          className='absolute inset-0 flex flex-col items-center justify-center z-10 '
          aria-hidden
        >
          <Spin size='default' className='mb-2' />
          <span className='text-gray-500 text-sm'>
            第 {pageNumber} 页加载中...
          </span>
        </div>
      )}
    </>
  )
}

interface PDFViewerProps {
  defaultPage?: number
  file: string
  name: string
  id: string
  locationKey?: string // 用于跟踪URL查询参数变化，确保点击分段时能触发滚动
  activeChunkId?: string | null
  segments?: any[]
  onChunkClick?: (chunkId: string) => void
  activeTab?: string // 当前激活的tab，用于控制chunk色块显示
}

// 暴露给父组件的方法接口
export interface PDFViewerRef {
  scrollToPage: (page: number) => void
}

const PDFViewer = forwardRef<PDFViewerRef, PDFViewerProps>((props, ref) => {
  const {
    file,
    defaultPage = 1,
    id,
    name,
    locationKey,
    activeChunkId,
    segments = [],
    onChunkClick,
    activeTab,
  } = props
  // 使用 ref 跟踪上一次的 defaultPage 和 locationKey，确保即使值相同也能检测到变化
  const prevDefaultPageRef = useRef<number | undefined>(undefined)
  const prevLocationKeyRef = useRef<string | undefined>(undefined)
  const [numPages, setNumPages] = useState<number | null>(null)
  const [pageNumber, setPageNumber] = useState<number>(1)
  const [pageInput, setPageInput] = useState<string>('1')
  const [scale, setScale] = useState<number>(0.9)
  const [scaleInput, setScaleInput] = useState<string>('90')
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)
  const isManuallyChangingPage = useRef<boolean>(false)
  const manualPageChangeTimer = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  )
  // 统一封装：标记为手动跳转并延迟重置，避免 IntersectionObserver 在滚动过程中误更新页码
  const scheduleManualChangeReset = useCallback(
    (duration = MANUAL_CHANGE_DURATION) => {
      isManuallyChangingPage.current = true
      if (manualPageChangeTimer.current)
        clearTimeout(manualPageChangeTimer.current)
      manualPageChangeTimer.current = setTimeout(() => {
        isManuallyChangingPage.current = false
        manualPageChangeTimer.current = null
      }, duration)
    },
    [],
  )
  // 使用 ref 记录当前页码，避免 IntersectionObserver 和其他 effect 频繁重建
  const currentPageRef = useRef(pageNumber)
  useEffect(() => {
    currentPageRef.current = pageNumber
  }, [pageNumber])

  // 已渲染页面缓存：离开视口后保留 DOM，再次滚动回来时无需重新加载
  const [renderedPagesCache, setRenderedPagesCache] = useState<Set<number>>(
    () => new Set(),
  )
  const { version, mode } = useDeployConfig()

  const handlePageRendered = useCallback(
    (pageNum: number) => {
      setRenderedPagesCache((prev) => {
        const next = new Set(prev).add(pageNum)
        if (next.size <= CACHE_MAX_SIZE) return next
        return new Set(
          [...next]
            .sort((a, b) => Math.abs(a - pageNumber) - Math.abs(b - pageNumber))
            .slice(0, CACHE_MAX_SIZE),
        )
      })
    },
    [pageNumber],
  )

  // chunk色块显示：仅在 本地环境、测试环境、custom 版本且 cimc 模式下显示（生产环境暂时不显示）
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isEnhancedMode =
    isDevEnv || isTestEnv || (version === 'custom' && mode === 'cimc')

  // 引用PDF容器元素和页面元素
  const pdfContainerRef = useRef<HTMLDivElement>(null)
  const pageRefs = useRef<(HTMLDivElement | null)[]>([])

  // 大型 PDF 虚拟化：仅渲染视口附近页面，避免 Canvas 内存超限/黑屏
  const PAGE_ESTIMATE_HEIGHT = 842 // A4 标准高度 (595x842)
  const PAGE_WINDOW_SIZE = 8 // 当前页前后各 8 页，共最多 17 页同时挂载
  const CACHE_MAX_SIZE = 34 // 缓存最多 34 页（约为 2 倍窗口），超出时按距离当前页淘汰
  const LARGE_PDF_THRESHOLD = 20
  const centerPage = pageNumber // 始终以当前页为中心
  const renderStartPage =
    numPages && numPages > LARGE_PDF_THRESHOLD
      ? Math.max(1, centerPage - PAGE_WINDOW_SIZE)
      : 1
  const renderEndPage =
    numPages && numPages > LARGE_PDF_THRESHOLD
      ? Math.min(numPages, centerPage + PAGE_WINDOW_SIZE)
      : numPages || 0

  /**
   * PDF文档加载成功回调
   * @param {object} 包含numPages属性的对象
   */
  function onDocumentLoadSuccess({ numPages }: { numPages: number }) {
    setNumPages(numPages)
    setLoading(false)
    // 初始化页面引用数组
    pageRefs.current = Array(numPages).fill(null)
  }

  /**
   * PDF文档加载失败回调
   * @param {Error} error 错误对象
   */
  function onDocumentLoadError(error: Error) {
    console.error('PDF 加载失败:', error)
    setError(`PDF 加载失败: ${error.message}`)
    setLoading(false)
  }

  const zoomIn = useCallback(() => {
    scheduleManualChangeReset()
    setScale((prevScale) => {
      const newScale = Math.min(prevScale + 0.1, 2.0)
      setScaleInput(Math.round(newScale * 100).toString())
      return newScale
    })
  }, [scheduleManualChangeReset])

  const zoomOut = useCallback(() => {
    scheduleManualChangeReset()
    setScale((prevScale) => {
      const newScale = Math.max(prevScale - 0.1, 0.5)
      setScaleInput(Math.round(newScale * 100).toString())
      return newScale
    })
  }, [scheduleManualChangeReset])

  const handleScaleInputChange = (value: string) => {
    // 只允许输入数字
    if (/^\d{0,3}$/.test(value)) {
      setScaleInput(value)
    }
  }

  const handleScaleInputBlur = useCallback(() => {
    let numValue = parseInt(scaleInput) || 90
    numValue = Math.min(Math.max(numValue, 50), 200)
    scheduleManualChangeReset()
    setScale(numValue / 100)
    setScaleInput(numValue.toString())
  }, [scaleInput, scheduleManualChangeReset])

  const handleScaleInputKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleScaleInputBlur()
      ;(e.target as HTMLInputElement).blur()
    }
  }

  // 页面变更时滚动到对应位置
  const scrollToPage = useCallback(
    (pageNum: number, behavior: ScrollBehavior = 'smooth') => {
      const pageElement = pageRefs.current[pageNum - 1]
      if (pageElement && pdfContainerRef.current) {
        // 计算需要滚动的位置
        const containerTop = pdfContainerRef.current.getBoundingClientRect().top
        const pageTop = pageElement.getBoundingClientRect().top
        const scrollTop =
          pdfContainerRef.current.scrollTop + (pageTop - containerTop)

        pdfContainerRef.current.scrollTo({
          top: scrollTop,
          behavior,
        })
      }
    },
    [],
  )

  // 暴露scrollToPage方法给父组件
  useImperativeHandle(ref, () => ({
    scrollToPage,
  }))

  // scale 变化时清空缓存，避免缩放后尺寸不一致，并保持当前页在视口中
  useEffect(() => {
    setRenderedPagesCache(new Set())
    const timer = setTimeout(() => {
      if (currentPageRef.current) {
        scrollToPage(currentPageRef.current, 'auto')
      }
    }, 100)
    return () => clearTimeout(timer)
  }, [scale, scrollToPage])

  const handlePageChange = useCallback(
    (page: number) => {
      scheduleManualChangeReset()
      setPageNumber(page)
      setPageInput(page.toString())
      scrollToPage(page)
    },
    [scrollToPage, scheduleManualChangeReset],
  )

  const handlePageInputChange = (value: string) => {
    // 只允许输入数字
    if (/^\d{0,4}$/.test(value)) {
      setPageInput(value)
    }
  }

  const handlePageInputBlur = () => {
    let page = parseInt(pageInput) || 1
    // 限制范围在1到总页数之间
    page = Math.min(Math.max(page, 1), numPages || 1)
    handlePageChange(page)
  }

  const handlePageInputKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handlePageInputBlur()
      ;(e.target as HTMLInputElement).blur()
    }
  }

  const handlePrevPage = () => {
    if (pageNumber > 1) {
      handlePageChange(pageNumber - 1)
    }
  }

  const handleNextPage = () => {
    if (pageNumber < (numPages || 1)) {
      handlePageChange(pageNumber + 1)
    }
  }

  // Ctrl + 方向键缩放
  const handleWheelZoom = useCallback(
    (event: KeyboardEvent) => {
      if (!event.ctrlKey) return
      if (event.key === 'ArrowUp') {
        event.preventDefault()
        zoomIn()
      } else if (event.key === 'ArrowDown') {
        event.preventDefault()
        zoomOut()
      }
    },
    [zoomIn, zoomOut],
  )

  const handleLinkNavigation = useCallback(
    (destination: any) => {
      handlePageChange(destination?.pageNumber ?? 1)
      return false
    },
    [handlePageChange],
  )

  // IntersectionObserver：完全接管页码更新，无论谁在滚动，只要页面进入视口就更新页码
  useEffect(() => {
    if (!numPages || loading) return

    const observer = new IntersectionObserver(
      (entries) => {
        // 如果正在手动跳转（点击了页码或 chunk），暂时不更新，避免跳动
        if (isManuallyChangingPage.current) return

        // 收集当前在视口内的所有页面
        let maxRatio = 0
        let bestPage = currentPageRef.current

        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const pageIndex = parseInt(
              (entry.target as HTMLElement).dataset.pageIndex || '0',
              10,
            )
            // 选出在视口中占比最大的页面作为当前页
            if (pageIndex > 0 && entry.intersectionRatio > maxRatio) {
              maxRatio = entry.intersectionRatio
              bestPage = pageIndex
            }
          }
        })

        if (bestPage > 0 && bestPage !== currentPageRef.current) {
          setPageNumber(bestPage)
          setPageInput(bestPage.toString())
        }
      },
      {
        root: null,
        // 扩大检测范围，提前加载（上下各提前一个屏幕的高度）
        rootMargin: '100% 0px 100% 0px',
        // 增加多个阈值，更精确判断哪个页面占比最大
        threshold: [0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
      },
    )

    // 延迟绑定观察，确保 DOM 已经渲染
    const timer = requestAnimationFrame(() => {
      pageRefs.current.forEach((el) => {
        if (el) observer.observe(el)
      })
    })

    return () => {
      cancelAnimationFrame(timer)
      observer.disconnect()
    }
  }, [numPages, loading])

  // 监听 defaultPage 和 locationKey 变化并跳转
  useEffect(() => {
    // 重置逻辑：当 defaultPage 为 undefined 时清空 ref
    if (defaultPage === undefined) {
      prevDefaultPageRef.current = undefined
      prevLocationKeyRef.current = undefined
      return
    }

    // 验证 defaultPage 有效性
    if (
      typeof defaultPage !== 'number' ||
      !numPages ||
      defaultPage < 1 ||
      defaultPage > numPages ||
      loading
    ) {
      return
    }

    // 检查是否有变化
    const defaultPageChanged = prevDefaultPageRef.current !== defaultPage
    const locationKeyChanged = prevLocationKeyRef.current !== locationKey

    if (!defaultPageChanged && !locationKeyChanged) {
      return
    }

    // 更新 ref 并触发滚动（延迟 300ms 等待 DOM 就绪）
    prevDefaultPageRef.current = defaultPage
    prevLocationKeyRef.current = locationKey
    isManuallyChangingPage.current = true

    const timer = setTimeout(() => {
      setPageNumber(defaultPage)
      setPageInput(defaultPage.toString())
      scrollToPage(defaultPage)
      scheduleManualChangeReset()
    }, 300)

    return () => clearTimeout(timer)
  }, [
    defaultPage,
    locationKey,
    numPages,
    loading,
    scrollToPage,
    scheduleManualChangeReset,
  ])

  useEffect(() => {
    window.addEventListener('keydown', handleWheelZoom)
    return () => {
      window.removeEventListener('keydown', handleWheelZoom)
    }
  }, [handleWheelZoom])

  useEffect(() => {
    return () => {
      if (manualPageChangeTimer.current) {
        clearTimeout(manualPageChangeTimer.current)
        manualPageChangeTimer.current = null
      }
    }
  }, [])

  const documentOptions = useMemo(
    () => ({
      cMapUrl: '/cmaps/',
      cMapPacked: true,
      useSystemFonts: true,
      disableFontFace: false,
      disableRange: false,
      rangeChunkSize: 128 * 1024,
    }),
    [],
  )

  const LoadingComponent = () => (
    <div className='absolute inset-0 z-10 flex items-center justify-center bg-white'>
      <Spin size='large' />
    </div>
  )

  if (!file) {
    return (
      <div className='h-full flex items-center justify-center'>
        <div className='text-center'>
          <p className='text-gray-500 mb-2 text-base'>暂无预览文件</p>
          <p className='text-sm text-gray-400'>请检查文件路径或重新加载</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className='h-full flex items-center justify-center'>
        <div className='text-center'>
          <p className='text-red-500 mb-2 text-base'>PDF 加载失败</p>
          <p className='text-sm text-gray-400'>{error}</p>
          <p className='text-xs text-gray-300 mt-2'>
            请检查文件是否存在或网络连接
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className='relative min-h-[400px] h-full flex flex-col overflow-hidden w-full'>
      {loading && <LoadingComponent />}

      <div
        className={`${loading ? 'invisible' : 'visible'} relative flex-1 overflow-hidden`}
      >
        <div
          className='h-full overflow-auto w-full custom-preview-scroll'
          ref={pdfContainerRef}
        >
          <div
            className='flex flex-col items-center pb-20'
            style={{ width: 'max-content', minWidth: '100%' }}
          >
            <Document
              file={file}
              onLoadSuccess={onDocumentLoadSuccess}
              onLoadError={onDocumentLoadError}
              onItemClick={handleLinkNavigation}
              options={documentOptions}
            >
              {Array.from(new Array(numPages || 0), (_, index) => {
                const currentPageNumber = index + 1
                const isInRenderWindow =
                  currentPageNumber >= renderStartPage &&
                  currentPageNumber <= renderEndPage
                const shouldRenderPage =
                  isInRenderWindow || renderedPagesCache.has(currentPageNumber)
                const pageSegments = segments.filter(
                  (s) => s.location && s.location[0] === currentPageNumber,
                )

                return (
                  <div
                    key={`page_${currentPageNumber}`}
                    ref={(el) => {
                      pageRefs.current[index] = el
                    }}
                    data-page-index={currentPageNumber}
                    className='relative mb-4 rounded-md shadow-[0_10px_15px_0_#0000000D] flex items-center justify-center'
                    style={{
                      // 必须设置确定的最小高度和宽度，否则占位符太小会导致 IntersectionObserver 一次性检测到大量页面
                      minHeight: PAGE_ESTIMATE_HEIGHT * scale,
                      minWidth: 595 * scale, // A4 基础宽度
                    }}
                  >
                    {shouldRenderPage ? (
                      <div
                        className={cn(
                          'relative w-full flex items-center justify-center',
                          !isInRenderWindow && 'invisible',
                        )}
                      >
                        <PageWithLoadingOverlay
                          pageNumber={currentPageNumber}
                          scale={scale}
                          onRendered={handlePageRendered}
                        />
                      </div>
                    ) : (
                      <div
                        className='flex items-center justify-center bg-gray-50 text-gray-400 text-sm w-full h-full'
                        style={{
                          height: PAGE_ESTIMATE_HEIGHT * scale,
                          width: 595 * scale,
                        }}
                      >
                        第 {currentPageNumber} 页 · 滚动至此处自动加载
                      </div>
                    )}
                    {/* 分段高亮层：仅在分段规则 tab 下且当前页在渲染窗口内显示 */}
                    {isInRenderWindow &&
                      isEnhancedMode &&
                      activeTab === 'segmentRule' &&
                      pageSegments.map((segment) => (
                        <SegmentHighlightBlock
                          key={`highlight_${segment.id}`}
                          segment={segment}
                          isActive={segment.id === activeChunkId}
                          pageIndex={index}
                          onChunkClick={onChunkClick}
                        />
                      ))}
                  </div>
                )
              })}
            </Document>
          </div>
        </div>

        {numPages && numPages > 0 && (
          <div className='absolute bottom-8 left-1/2 -translate-x-1/2 z-50 rounded-xl shadow-lg p-3 flex items-center bg-white/95 backdrop-blur-sm'>
            {/* 左侧：缩放控制 */}
            <div className='flex items-center space-x-1'>
              <Button
                type='text'
                icon={<ReduceIcon />}
                onClick={zoomOut}
                className='flex-col items-center pt-1 !w-8 !h-8 hover:bg-gray-100'
                size='small'
              />
              <div className='relative'>
                <Input
                  value={scaleInput}
                  onChange={(e) => handleScaleInputChange(e.target.value)}
                  onBlur={handleScaleInputBlur}
                  onKeyDown={handleScaleInputKeyDown}
                  className='!w-[65px] text-center !h-8 !bg-[#F2F2F2] rounded !border-none'
                  suffix='%'
                />
              </div>
              <Button
                type='text'
                icon={<MagnifyIcon />}
                onClick={zoomIn}
                className='flex-col items-center pt-1 !w-8 !h-8 hover:bg-gray-100'
                size='small'
              />
            </div>

            {/* 分隔线 */}
            <div className='h-6 w-[1px] bg-[#EDEDED] mx-2'></div>

            {/* 中间：分页控制 */}
            <div className='flex items-center space-x-2'>
              <Button
                type='text'
                icon={<LeftOutlined />}
                onClick={handlePrevPage}
                disabled={pageNumber <= 1}
                className='!h-7 hover:bg-gray-100'
                size='small'
              />
              <div className='flex items-center space-x-1 gap-1'>
                <Input
                  value={pageInput}
                  onChange={(e) => handlePageInputChange(e.target.value)}
                  onBlur={handlePageInputBlur}
                  onKeyDown={handlePageInputKeyDown}
                  className='!w-[45px] rounded text-center !h-7 !border-[#EDEDED]'
                />
                <span className='text-[#0C1F17]'>/</span>
                <span className='text-[#0C1F17] min-w-[30px] text-center'>
                  {numPages || 0}
                </span>
              </div>
              <Button
                type='text'
                icon={<RightOutlined />}
                onClick={handleNextPage}
                disabled={pageNumber >= (numPages || 1)}
                className='!h-7 hover:bg-gray-100'
                size='small'
              />
            </div>

            {/* 分隔线 */}
            <div className='h-6 w-[1px] bg-[#EDEDED] mx-2'></div>

            {/* 右侧：下载按钮 */}
            <DownLoadBtn id={id} name={name} icon={<DownloadIcon />} />
          </div>
        )}
      </div>
    </div>
  )
})

PDFViewer.displayName = 'PDFViewer'

export default PDFViewer
