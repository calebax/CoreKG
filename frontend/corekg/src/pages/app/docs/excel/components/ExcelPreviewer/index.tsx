import { FC, lazy, useEffect } from 'react'
import { Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { getPreviewFileURL } from '@/api/knowledge'
import { withSuspense } from '@/utils/withSuspense'

const loadSheetViewer = () => import('@/components/SheetViewer')
const SheetViewer = withSuspense(lazy(loadSheetViewer))
export type ExcelPreviewer = {
  id: number
  onError?: () => void
}
// 触控板导航手势阈值
const NAVIGATION_GESTURE_THRESHOLD = 30

// 辅助函数：查找可横向滚动的父元素
const findScrollableParent = (
  element: HTMLElement | null,
): HTMLElement | null => {
  while (element) {
    const { overflowX, overflow } = window.getComputedStyle(element)
    const canScrollX =
      overflowX === 'auto' ||
      overflowX === 'scroll' ||
      overflow === 'auto' ||
      overflow === 'scroll'

    if (canScrollX && element.scrollWidth > element.clientWidth) {
      return element
    }
    element = element.parentElement
  }
  return null
}

export const ExcelPreviewer: FC<ExcelPreviewer> = (props) => {
  const { id, onError } = props
  const { data: url } = useRequest(
    async () => {
      try {
        const { url } = await getPreviewFileURL({ file_id: id })
        return url as string
      } catch {
        onError?.()
      }
    },
    { refreshDeps: [id] },
  )

  // 预加载SheetViewer组件
  useEffect(() => {
    loadSheetViewer()
  }, [])

  // 阻止导航手势但保留滚动功能
  useEffect(() => {
    // 防止浏览器手势导航的CSS
    const style = document.createElement('style')
    style.id = 'disable-navigation-styles'
    style.textContent = `
      html, body, .excel-preview-container {
        overscroll-behavior-x: contain !important;
      }
    `
    document.head.appendChild(style)

    // 保存原始样式
    const originalStyles = {
      body: document.body.style.overscrollBehaviorX,
      html: document.documentElement.style.overscrollBehaviorX,
    }

    // 应用防导航样式
    document.body.style.overscrollBehaviorX = 'contain'
    document.documentElement.style.overscrollBehaviorX = 'contain'

    // 处理滚轮事件 - 智能阻止导航手势
    const handleWheel = (e: WheelEvent) => {
      const isHorizontalSwipe = Math.abs(e.deltaX) > Math.abs(e.deltaY)
      const isNavigationGesture =
        Math.abs(e.deltaX) > NAVIGATION_GESTURE_THRESHOLD

      if (isHorizontalSwipe && isNavigationGesture) {
        e.preventDefault()
        e.stopPropagation()

        // 如果内容可横向滚动，手动应用滚动
        const scrollable = findScrollableParent(e.target as HTMLElement)
        if (scrollable) {
          scrollable.scrollLeft += e.deltaX
        }
        return false
      }
      return true
    }

    // 阻止键盘导航快捷键
    const handleKeydown = (e: KeyboardEvent) => {
      const isNavigationShortcut =
        (e.altKey && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) ||
        ((e.metaKey || e.ctrlKey) && (e.key === '[' || e.key === ']'))

      if (isNavigationShortcut) {
        e.preventDefault()
        return false
      }
    }

    // 添加事件监听器
    const wheelOptions = { passive: false, capture: true }
    document.addEventListener('wheel', handleWheel, wheelOptions)
    window.addEventListener('wheel', handleWheel, wheelOptions)
    document.addEventListener('keydown', handleKeydown, { capture: true })

    // 清理函数
    return () => {
      // 移除样式
      document.getElementById('disable-navigation-styles')?.remove()

      // 恢复原始样式
      document.body.style.overscrollBehaviorX = originalStyles.body
      document.documentElement.style.overscrollBehaviorX = originalStyles.html

      // 移除事件监听器
      document.removeEventListener('wheel', handleWheel)
      window.removeEventListener('wheel', handleWheel)
      document.removeEventListener('keydown', handleKeydown)
    }
  }, [])

  if (!url) return <Skeleton active />
  return <SheetViewer file={url} />
}
