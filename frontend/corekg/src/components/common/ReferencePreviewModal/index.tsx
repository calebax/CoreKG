import { Image } from 'antd'
import dayjs from 'dayjs'
import CloseIcon from '@/assets/icons/close.svg?react'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import MarkdownPreview from '../MarkdownPreview'
import styles from './index.module.scss'
import { getFileIcon } from './utils'

const getReferenceChunkUniqueKey = (chunk: any) => {
  if (chunk?.chunk_id) return `chunk_id:${chunk.chunk_id}`
  if (chunk?.sequence !== undefined) return `sequence:${chunk.sequence}`
  return JSON.stringify(chunk)
}

export default function ReferencePreviewModal(props: {
  visible: boolean
  onClose: () => void
  references: any
  modalModel: {
    forest_id: string
    docId: number
    chunkId?: string
    chunkIndex?: string
  }
  position: {
    x: number
    y: number
  }
}) {
  const file = useMemo(() => {
    if (!props.visible) return null
    const sequence = Number(props.modalModel.chunkIndex)
    const matchedReferences = props.references.filter(
      (item: any) =>
        item.forest_id === props.modalModel.forest_id &&
        item.file_id === props.modalModel.docId,
    )
    if (!matchedReferences.length) return null

    const reference = matchedReferences[0]
    const mergedChunkMap = new Map<string, any>()
    matchedReferences.forEach((item: any) => {
      item.chunk_list?.forEach((chunk: any) => {
        mergedChunkMap.set(getReferenceChunkUniqueKey(chunk), chunk)
      })
    })
    const mergedChunkList = Array.from(mergedChunkMap.values()).sort(
      (a: any, b: any) => {
        const aSequence =
          typeof a.sequence === 'number' ? a.sequence : Number.MAX_SAFE_INTEGER
        const bSequence =
          typeof b.sequence === 'number' ? b.sequence : Number.MAX_SAFE_INTEGER
        return aSequence - bSequence
      },
    )
    const chunk =
      mergedChunkList.find(
        (c: any) =>
          (props.modalModel.chunkId &&
            c.chunk_id === props.modalModel.chunkId) ||
          (!Number.isNaN(sequence) && c.sequence === sequence),
      ) ?? mergedChunkList[0]
    if (!chunk) return null
    return {
      title: reference.file_name,
      forest_id: reference.forest_id,
      file_id: reference.file_id,
      chunk,
      chunk_list: mergedChunkList,
      user_name: reference?.user_name || '',
      created_at:
        reference?.created_at && !reference.created_at.includes('0001')
          ? dayjs(reference?.created_at).format('YYYY/M/DD HH:mm')
          : '',
    }
  }, [props.modalModel, props.references, props.visible])

  if (!file) return null

  function getDisposeChunkContent(chunk: any) {
    if (chunk.type === 'table') {
      const match = chunk.content.match(/<table>(.*)<\/table>/is)
      if (match?.[1]) return `<table>${match[1]}</table>`
      return chunk.content
    }
    return chunk.content.startsWith('#')
      ? chunk.content.substring(1)
      : chunk.content
  }

  const handleOpenFile = () => {
    const searchParams = new URLSearchParams()
    searchParams.append(
      'location',
      encodeURIComponent(JSON.stringify(file.chunk.location)),
    )
    window.open(
      `${location.origin}/docs/detail/${file.forest_id}/file/${file.file_id}?${searchParams.toString()}`,
      '_blank',
    )
  }

  const renderReferenceCard = (file: any) => {
    return (
      <div key={file.chunk.chunk_id} className='rounded-[6px] p-[8px]'>
        {!['image', 'video'].includes(file.chunk.type) && (
          <div
            className='overflow-auto line-clamp-3 text-[#3C4149]'
            style={{
              maxHeight: file.chunk.type !== 'table' ? '200px' : 'auto',
              overflowX: file.chunk.type === 'table' ? 'auto' : 'hidden',
            }}
          >
            <MarkdownPreview
              content={getDisposeChunkContent(file.chunk)}
              className='!bg-[transparent]'
              style={{
                overflow: file.chunk.type === 'table' ? 'visible' : 'hidden',
                padding: '0 !important',
                height: 'auto',
                maxHeight: 'none',
                backgroundColor: 'transparent',
                display: file.chunk.type === 'table' ? 'block' : '-webkit-box',
                WebkitLineClamp: file.chunk.type === 'table' ? 'unset' : 3,
                WebkitBoxOrient:
                  file.chunk.type === 'table' ? 'unset' : 'vertical',
                textOverflow:
                  file.chunk.type === 'table' ? 'unset' : 'ellipsis',
                minWidth: file.chunk.type === 'table' ? 'max-content' : 'auto',
              }}
            />
            {/* {chunk.content} */}
          </div>
        )}
        {['image', 'video'].includes(file.chunk.type) &&
          file.chunk.image_url && (
            <div className='w-[50%] h-auto mx-[auto] mt-[6px]'>
              <Image
                src={file.chunk.image_url}
                alt={file.title}
                preview={{
                  mask: '点击查看大图',
                }}
                className='cursor-pointer'
              />
            </div>
          )}
      </div>
    )
  }

  // 计算弹窗位置和高度
  const getModalStyle = () => {
    const { x, y } = props.position
    const { innerWidth: viewportWidth, innerHeight: viewportHeight } = window
    const isMobile = viewportWidth < 768
    const modalWidth = isMobile ? Math.min(viewportWidth - 32, 400) : 480
    const padding = 16
    const offset = 20

    const left = Math.max(
      padding,
      Math.min(x - modalWidth / 2, viewportWidth - modalWidth - padding),
    )
    const top = y + offset
    const maxHeight = Math.max(200, viewportHeight - top - padding)

    return {
      position: 'fixed' as const,
      left: `${left}px`,
      top: `${top}px`,
      margin: 0,
      boxShadow: '0px 2px 12px 0px rgba(0,0,0,0.1)',
      zIndex: 1000,
      pointerEvents: 'auto' as const,
      width: `${modalWidth}px`,
      maxHeight: `${maxHeight}px`,
    }
  }

  return (
    <div
      className={`${styles.referenceModal} reference-modal-container flex flex-col bg-white rounded-lg border border-gray-200 shadow-lg`}
      style={getModalStyle()}
      onClick={(e) => e.stopPropagation()}
    >
      <div className='h-[50px] flex gap-[10px] items-center text-[16px] border-b border-b-[#EFF1F4] px-4 relative flex-shrink-0'>
        {getFileIcon(file.title)}
        <div className='font-[500] max-w-[330px] text-[#0C1F17] overflow-hidden whitespace-nowrap text-ellipsis'>
          {file.title}
        </div>
        <button
          onClick={props.onClose}
          className='absolute right-2 top-3 w-6 h-6 flex items-center justify-center'
          aria-label='关闭'
        >
          <CloseIcon className='text-[#616373] text-base cursor-pointer' />
        </button>
      </div>
      <div
        className={`py-[20px] px-4 flex-1 overflow-auto min-h-0 ${scrollStyles.scroll}`}
      >
        {renderReferenceCard(file)}
      </div>
      <div className='h-[60px] border-t border-t-[#EFF1F4] flex items-center justify-between px-4 flex-shrink-0'>
        <div className='flex gap-[8px] items-center'>
          {file?.user_name && (
            <div className='text-[#919497] text-[14px] leading-[1]  overflow-hidden whitespace-nowrap text-ellipsis max-w-[11em]'>
              {file.user_name}
            </div>
          )}
          {file?.created_at && (
            <div className='text-[#919497] text-[12px] leading-[1]'>
              {file.created_at}
            </div>
          )}
        </div>
        <div
          onClick={handleOpenFile}
          className='border-0 h-[30px] cursor-pointer flex items-center bg-[#FBE9FF] hover:bg-[#F3D9FF] text-[#CC5DE8] hover:text-[#B832E6] text-[14px] px-[10px] rounded-[6px] transition-colors duration-200'
        >
          查看原文
        </div>
      </div>
    </div>
  )
}
