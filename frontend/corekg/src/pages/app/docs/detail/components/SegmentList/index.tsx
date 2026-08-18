import { useRef, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Button, Image, Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import ChangeRuleIcon from '@/assets/icons/docs/change-rules.svg'
import deleteIcon from '@/assets/icons/docs/delete.svg'
import editIcon from '@/assets/icons/docs/edit.svg'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useFileDetailViewProject } from '../FileDetailView'
import styles from './index.module.scss'

interface FileSegment {
  id: string
  content: string
  table?: string
  type?: 'chunk' | 'table' | 'image'
  chunk_number: number
  charCount: number
  location?: number[] // 用于定位文件位置，如PDF页码
  imageUrl?: string
}

interface ProjectValue {
  segments: FileSegment[]
  activeChunkId: string | null
  setActiveChunkId: (id: string | null) => void
  handleSegmentEdit: (id: string) => void
  handleSegmentDeleteConfirm: (id: string) => void
  segmentTotal: number
  handleConfiguration: () => void
}

export default function SegmentList() {
  const { t } = useTranslation('pages')
  const [searchParams, setSearchParams] = useSearchParams()
  const { version, mode } = useDeployConfig()
  const listRef = useRef<HTMLDivElement>(null)
  const {
    segments,
    segmentTotal,
    activeChunkId,
    setActiveChunkId,
    handleSegmentEdit,
    handleSegmentDeleteConfirm,
    handleConfiguration,
  } = useFileDetailViewProject<ProjectValue>()!

  // 处理自动滚动到激活项
  useEffect(() => {
    if (!activeChunkId || !listRef.current) return

    const scrollToActiveChunk = (behavior: ScrollBehavior = 'smooth') => {
      const container = listRef.current
      if (!container) return

      const activeEl = container.querySelector<HTMLElement>(
        `[data-chunk-id="${activeChunkId}"]`,
      )
      if (!activeEl) return

      const containerRect = container.getBoundingClientRect()
      const activeRect = activeEl.getBoundingClientRect()
      const scrollTop =
        container.scrollTop +
        activeRect.top -
        containerRect.top -
        container.clientHeight / 2 +
        activeRect.height / 2

      container.scrollTo({
        top: Math.max(0, scrollTop),
        behavior,
      })
    }

    const frameId = requestAnimationFrame(() => scrollToActiveChunk())
    // 图片/Markdown 渲染完成后高度会变化，补一次定位，避免目标分段被挤出可视区
    const timer = setTimeout(() => scrollToActiveChunk('auto'), 300)

    return () => {
      cancelAnimationFrame(frameId)
      clearTimeout(timer)
    }
  }, [activeChunkId, segments.length])

  // 是否显示编辑和删除功能：仅在 本地环境、测试环境、custom 版本且 cimc 模式下显示（生产环境暂时不显示）
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isEnhancedMode =
    isDevEnv || isTestEnv || (version === 'custom' && mode === 'cimc')
  const showEditAndDelete = isEnhancedMode

  // 处理分段点击：更新URL的location参数，添加时间戳确保每次点击都能触发滚动
  const handleSegmentClick = (segment: FileSegment) => {
    setActiveChunkId(segment.id)
    if (segment.location?.length) {
      const newParams = new URLSearchParams(searchParams)
      newParams.set('location', JSON.stringify(segment.location))
      newParams.set('_t', Date.now().toString())
      setSearchParams(newParams, { replace: true })
    }
  }

  // 渲染分段内容：根据是否有图片和文本来决定布局
  const renderSegmentContent = (segment: FileSegment) => {
    const segmentType = segment.type || 'chunk'
    const hasImage = !!segment.imageUrl
    const hasContent = !!segment.content
    const isTableSegment = segmentType === 'table'
    const isImageSegment = segmentType === 'image'

    if (isTableSegment) {
      if (!isEnhancedMode) {
        return (
          <div className='max-h-[300px] overflow-y-auto custom-preview-scroll'>
            <MarkdownPreview
              content={segment.content}
              className='text-[#1e1f28] text-base leading-[22px]'
            />
          </div>
        )
      }

      return (
        <div className='max-h-[300px] overflow-y-auto custom-preview-scroll'>
          <MarkdownPreview
            content={segment.table || ''}
            className='text-[#1e1f28] text-base leading-[22px] flex items-center justify-center w-full'
            style={{ backgroundColor: 'transparent', padding: 0 }}
          />
        </div>
      )
    }

    if (isImageSegment) {
      if (!hasContent) {
        return (
          <div className='flex justify-center w-full py-2'>
            <Image
              src={segment.imageUrl!}
              alt='segment image'
              className='object-contain max-h-[300px]'
            />
          </div>
        )
      }

      return (
        <div className='flex gap-3'>
          <div className='w-[30%] flex-shrink-0 flex items-center justify-center rounded p-1'>
            <Image
              src={segment.imageUrl!}
              alt='segment image'
              className='object-contain max-h-[200px]'
              style={{ width: '100%' }}
            />
          </div>
          <div className='w-[70%] max-h-[300px] overflow-y-auto custom-preview-scroll'>
            <MarkdownPreview
              content={segment.content}
              className='text-[#1e1f28] text-base leading-[22px]'
              style={
                isEnhancedMode
                  ? { backgroundColor: 'transparent', padding: 0 }
                  : {}
              }
            />
          </div>
        </div>
      )
    }

    // 只有文本：全宽展示
    return (
      <div className='max-h-[300px] overflow-y-auto custom-preview-scroll'>
        <MarkdownPreview
          content={segment.content}
          className='text-[#1e1f28] text-base leading-[22px]'
          style={
            isEnhancedMode ? { backgroundColor: 'transparent', padding: 0 } : {}
          }
        />
      </div>
    )
  }

  return (
    <div className='h-full overflow-auto' ref={listRef}>
      <div className='flex items-center gap-[10px] mb-[10px]'>
        <div className='text-[14px] font-[500] text-[#3C4149]'>
          分段数：{segmentTotal}
        </div>
        <Button
          type='text'
          onClick={handleConfiguration}
          className='flex items-center gap-1 px-2.5 py-2 bg-[#F5F5F5] rounded-[6px] hover:bg-[#F5F5F5] text-[#0C1F17] text-sm font-medium'
        >
          <span>
            <img src={ChangeRuleIcon} alt='change rule' className='w-4 h-4' />
          </span>
          {t('app.docs.fileDetail.editRules')}
        </Button>
      </div>
      <div className='space-y-3 h-full custom-preview-scroll pr-1'>
        {segments.map((segment, index) => {
          const isActive = segment.id === activeChunkId
          // 使用 chunk_number 判断奇偶，如果没有则用索引
          const isEven = (segment.chunk_number || index + 1) % 2 === 0

          return (
            <div
              key={segment.id}
              className='mb-[16px]'
              data-chunk-id={segment.id}
            >
              {/* 分段标题行 */}
              <div className='flex items-center mb-[6px] justify-between'>
                <div className='flex items-center gap-1'>
                  <span className='text-[#ABAFB2] text-[12px] font-normal'>
                    {t('app.docs.fileDetail.segmentTitle', {
                      number: segment.chunk_number,
                      count: segment.charCount,
                    })}
                  </span>
                </div>
              </div>

              {/* 文本/图片内容框 */}
              <div
                className={cn(
                  'border relative rounded-md px-2.5 py-[5px] w-full cursor-pointer transition-all duration-200',
                  isEnhancedMode
                    ? {
                        // 选中或悬浮状态：#CC5DE84D 背景 + #CC5DE8 边框
                        'bg-[#CC5DE84D] border-[#CC5DE8] shadow-[0_0_8px_rgba(204,93,232,0.2)] z-10':
                          isActive,
                        'hover:bg-[#CC5DE84D] hover:border-[#CC5DE8] z-10':
                          !isActive,
                        // 默认偶数状态：#E1E8FF4D 背景 + #E1E8FF 边框
                        'bg-[#E1E8FF4D] border-[#E1E8FF]': !isActive && isEven,
                        // 默认奇数状态：#97DDFF80 背景 + #97DDFF 边框
                        'bg-[#97DDFF80] border-[#97DDFF]': !isActive && !isEven,
                      }
                    : {
                        'border-gray-200 bg-white': !isActive,
                        'border-[#165DFF] bg-[#165DFF]/5': isActive,
                      },
                  styles.textWrap,
                )}
                onClick={(e) => {
                  // 如果点击的是图片预览组件(包含mask)，不触发定位
                  if (
                    (e.target as HTMLElement).closest('.ant-image-mask') ||
                    (e.target as HTMLElement).closest('.ant-image-preview')
                  ) {
                    return
                  }
                  // 如果点击的是编辑或删除按钮，不触发定位
                  if (
                    (e.target as HTMLElement).closest('button') ||
                    (e.target as HTMLElement).closest('.ant-btn')
                  ) {
                    return
                  }
                  handleSegmentClick(segment)
                }}
              >
                {renderSegmentContent(segment)}
                {showEditAndDelete && (
                  <div
                    className={cn(
                      'flex items-center gap-2.5 absolute right-[4px] top-[4px] bg-[#fff] rounded-sm p-0.5',
                      styles.textWrapOperator,
                    )}
                  >
                    <Tooltip title={t('app.docs.fileDetail.edit')}>
                      <Button
                        type='text'
                        size='small'
                        icon={
                          <img src={editIcon} alt='edit' className='w-4 h-4' />
                        }
                        onClick={(e) => {
                          e.stopPropagation()
                          handleSegmentEdit(segment.id)
                        }}
                        className='hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] w-4 h-4 p-0 rounded transition-all duration-200'
                      />
                    </Tooltip>
                    <Tooltip title={t('app.docs.fileDetail.delete')}>
                      <Button
                        type='text'
                        size='small'
                        icon={
                          <img
                            src={deleteIcon}
                            alt='delete'
                            className='w-4 h-4'
                          />
                        }
                        onClick={(e) => {
                          e.stopPropagation()
                          handleSegmentDeleteConfirm(segment.id)
                        }}
                        className='hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] w-4 h-4 p-0 rounded transition-all duration-200'
                      />
                    </Tooltip>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
