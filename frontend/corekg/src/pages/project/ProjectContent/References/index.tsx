import { useMemo } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import dayjs from 'dayjs'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import type { AIDialog, Chunk } from '@/components/dialog/AIDialog'
import { useSessionInfo } from '..'
import { getFileIcon } from './utils'

type Reference = AIDialog['reference']

export default memo(function References() {
  const { dialog, dialogIndex } = useSessionInfo()
  const references = useMemo(() => {
    if (dialogIndex === -1) return []
    return (dialog[dialogIndex] as AIDialog)?.reference || []
  }, [dialog, dialogIndex])
  const parentRef = useRef<HTMLDivElement>(null)
  const rowVirtualizer = useVirtualizer({
    count: references.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 200,
  })
  const items = rowVirtualizer.getVirtualItems()
  function getDisposeChunkContent(chunk: Chunk) {
    if (chunk.type === 'table') {
      const match = chunk.content.match(/<table>(.*)<\/table>/is)
      if (match?.[1]) return `<table>${match[1]}</table>`
      return chunk.content
    }
    return chunk.content.startsWith('#')
      ? chunk.content.substring(1)
      : chunk.content
  }

  const handleOpenFile = (
    { forest_id, file_id }: Reference[number],
    chunk: Chunk,
  ) => {
    const searchParams = new URLSearchParams()

    searchParams.append(
      'location',
      encodeURIComponent(JSON.stringify(chunk.location)),
    )
    window.open(
      `${location.origin}/docs/detail/${forest_id}/file/${file_id}?${searchParams.toString()}`,
      '_blank',
    )
  }

  const renderReferenceCard = (reference: Reference[number], chunk: Chunk) => {
    return (
      <div
        key={chunk.chunk_id}
        className='rounded-[6px] cursor-pointer hover:bg-[#f7f7f7] p-[8px]'
        onClick={() => handleOpenFile(reference, chunk)}
      >
        {!['image', 'video'].includes(chunk.type) && (
          <div
            className='overflow-auto   line-clamp-3 text-[#3C4149]'
            style={{
              maxHeight: chunk.type !== 'table' ? '200px' : 'auto',
            }}
          >
            <MarkdownPreview
              content={getDisposeChunkContent(chunk)}
              className='!bg-[transparent]'
              style={{
                overflow: 'hidden',
                padding: '0 !important',
                height: 'auto',
                maxHeight: 'none',
                backgroundColor: 'transparent',
                display: '-webkit-box',
                WebkitLineClamp: 3,
                WebkitBoxOrient: 'vertical',
                textOverflow: 'ellipsis',
              }}
            />
            {/* {chunk.content} */}
          </div>
        )}
        {['image', 'video'].includes(chunk.type) && chunk.image_url && (
          <div className='w-[50%] h-auto mx-[auto] mt-[6px]'>
            <img src={chunk.image_url} />
          </div>
        )}
        <div className='flex'></div>
      </div>
    )
  }

  return (
    <div
      ref={parentRef}
      className='max-h-full flex flex-col overflow-auto relative px-[20px] py-[10px] gap-[6px]'
    >
      <div
        className='w-full relative'
        style={{
          height: rowVirtualizer.getTotalSize(),
        }}
      >
        <div
          className='absolute top-0 left-0 w-full'
          style={{ transform: `translateY(${items[0]?.start ?? 0}px)` }}
        >
          {items.map((virtualItem) => {
            const { key, index } = virtualItem
            const reference = references[index]
            if (!reference?.chunk_list?.length) return null
            return (
              <div
                key={key}
                data-index={index}
                ref={rowVirtualizer.measureElement}
                className='flex flex-col'
              >
                <div className='flex items-center '>
                  <div className='h-[30px] items-center flex overflow-hidden'>
                    {getFileIcon(reference.file_name)}
                    <div className='text-[#3C4149] text-base font-medium whitespace-nowrap flex-1 overflow-hidden text-ellipsis'>
                      {reference.file_name}
                    </div>
                  </div>
                </div>
                {reference.chunk_list.map((chunk) =>
                  renderReferenceCard(reference, chunk),
                )}

                <div className='flex gap-[8px] h-[30px] justify-between items-center'>
                  <div className='text-[#919497] text-[14px] leading-[1]  overflow-hidden whitespace-nowrap text-ellipsis max-w-[11em]'>
                    {reference?.user_name ? reference.user_name : ''}
                  </div>
                  {reference?.created_at &&
                    !reference.created_at.includes('0001') && (
                      <div className='text-[#919497] text-[14px] leading-[1]'>
                        {dayjs(reference.created_at).format('YYYY/M/DD HH:mm')}
                      </div>
                    )}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
})
