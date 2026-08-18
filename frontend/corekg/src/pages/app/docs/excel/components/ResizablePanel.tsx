import { useState, useCallback, useRef, ReactNode, useEffect } from 'react'
import KnowledgeDrag from '@/assets/icons/knowledge-drag.svg?react'

interface ResizablePanelProps {
  leftPanel: ReactNode
  rightPanel: ReactNode
  initialLeftWidth?: number
  minLeftWidth?: number
  minRightWidth?: number
}

export default function ResizablePanel({
  leftPanel,
  rightPanel,
  initialLeftWidth = 50,
  minLeftWidth = 30,
  minRightWidth = 25,
}: ResizablePanelProps) {
  const [leftWidth, setLeftWidth] = useState<number>(initialLeftWidth)
  const [isDragging, setIsDragging] = useState<boolean>(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !containerRef.current) return

      const containerRect = containerRef.current.getBoundingClientRect()
      const newLeftWidth =
        ((e.clientX - containerRect.left) / containerRect.width) * 100

      // 限制最小和最大宽度
      const clampedWidth = Math.max(
        minLeftWidth,
        Math.min(100 - minRightWidth, newLeftWidth),
      )
      setLeftWidth(clampedWidth)
    },
    [isDragging, minLeftWidth, minRightWidth],
  )

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
  }, [])

  // 监听全局鼠标事件
  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
    } else {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, handleMouseMove, handleMouseUp])

  return (
    <div ref={containerRef} className='flex h-full w-full relative'>
      {/* 左侧面板 */}
      <div
        className='h-full overflow-hidden'
        style={{ width: `${leftWidth}%` }}
      >
        {leftPanel}
      </div>

      {/* 拖动分割控制器 */}
      <div className='relative flex items-center justify-center w-3 transition-colors duration-200 group'>
        <div
          className='w-40 h-6 flex items-center justify-center rounded-sm cursor-ew-resize'
          onMouseDown={handleMouseDown}
        >
          <KnowledgeDrag
            className='text-gray-400 group-hover:text-[#0C99FF] transition-colors duration-200'
            style={{ fontSize: '10px' }}
          />
        </div>
      </div>

      {/* 右侧面板 */}
      <div
        className='h-full flex flex-col'
        style={{ width: `${100 - leftWidth - 0.5}%`, minWidth: '0' }}
      >
        {rightPanel}
      </div>
    </div>
  )
}
