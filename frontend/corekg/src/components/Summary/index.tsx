import React, { useMemo, useEffect, useRef } from 'react'
import { Skeleton } from 'antd'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import rehypeRaw from 'rehype-raw'
import remarkGfm from 'remark-gfm'
import { cn } from '@/utils'

/**
 * 预处理 Markdown 文本，手动渲染 KaTeX 公式并转换位置标记
 */
const preprocessMarkdown = (
  markdown: string,
  activeLocation?: string | null,
): string => {
  if (!markdown) return ''

  const renderBlockFormula = (formula: string): string => {
    try {
      return katex.renderToString(formula.trim(), {
        displayMode: true,
        throwOnError: true,
      })
    } catch {
      return ''
    }
  }

  const renderInlineFormula = (formula: string): string => {
    const cleaned = formula.replace(/\\\$/g, '$')
    try {
      return katex.renderToString(cleaned.trim(), {
        displayMode: false,
        throwOnError: true,
      })
    } catch {
      return ''
    }
  }

  // 1. 处理 LaTeX 风格公式 \[ ... \] 和 \( ... \)
  let processed = markdown.replace(/\\\[([\s\S]+?)\\\]/g, (_, formula) =>
    renderBlockFormula(formula),
  )
  processed = processed.replace(/\\\(([\s\S]+?)\\\)/g, (_, formula) =>
    renderInlineFormula(formula),
  )

  // 2. 处理 $ 风格公式 $$ ... $$ 和 $ ... $
  processed = processed.replace(/\$\$([\s\S]+?)\$\$/g, (_, formula) =>
    renderBlockFormula(formula),
  )
  processed = processed.replace(
    /\$((?:[^\$]|\\[\$])+?)\$/g,
    (_, formula) => renderInlineFormula(formula),
  )

  // 3. 处理 yg_pos 标记，将其转换为可交互的 HTML
  // 注意：这里为了不破坏 Markdown 结构，将其转换为带 data 属性的 span
  processed = processed.replace(/<!--yg_pos(.*?)yg_pos-->/g, (match, loc) => {
    const isActive = activeLocation && activeLocation.includes(loc.trim())
    return `<span class="yg-chunk-marker ${isActive ? 'active-highlight' : ''}" data-location="${loc.trim()}"></span>`
  })

  return processed
}

const SafeMarkdownRenderer: React.FC<{
  markdownData: string
  activeLocation?: string | null
  onLocationClick?: (loc: string) => void
}> = ({ markdownData, activeLocation, onLocationClick }) => {
  const containerRef = useRef<HTMLDivElement>(null)

  const processedMarkdown = useMemo(
    () => preprocessMarkdown(markdownData, activeLocation),
    [markdownData, activeLocation],
  )

  // 处理点击事件
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const handleClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      // 查找最近的带 data-location 的元素（由于 marker 可能被包裹）
      const marker = target.closest('.yg-chunk-marker') as HTMLElement
      if (marker) {
        const loc = marker.getAttribute('data-location')
        if (loc) {
          onLocationClick?.(loc)
        }
      }
    }

    container.addEventListener('click', handleClick)
    return () => container.removeEventListener('click', handleClick)
  }, [onLocationClick])

  // 处理自动滚动到高亮位置
  useEffect(() => {
    if (!activeLocation) return
    const container = containerRef.current
    if (!container) return

    const activeEl = container.querySelector('.active-highlight')
    if (activeEl) {
      activeEl.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }, [activeLocation])

  return (
    <div ref={containerRef} className='h-full w-full'>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw, rehypeHighlight]}
      >
        {processedMarkdown}
      </ReactMarkdown>
    </div>
  )
}

const Summary: React.FC<{
  isForestPage?: boolean
  markdownData: string
  flag?: boolean
  intelligentAnalysisTab?: boolean
  activeLocation?: string | null
  onLocationClick?: (loc: string) => void
}> = ({
  isForestPage = false,
  markdownData,
  flag = false,
  intelligentAnalysisTab = false,
  activeLocation,
  onLocationClick,
}) => {
  const getStyleClass = () => {
    if (intelligentAnalysisTab) return 'intelligent-analysis-tab'
    if (flag) return 'custom-markdown'
    return ''
  }

  const markdownContainerStyle = {
    backgroundColor: '#ffffff',
    color: '#24292e',
  }

  return (
    <main
      className={`${isForestPage ? 'markdown-body-forest' : 'markdown-body-file'} ${getStyleClass()} w-full h-full min-w-0`}
    >
      <style>{`
        .yg-chunk-marker {
          display: inline-block;
          width: 4px;
          height: 1.2em;
          vertical-align: middle;
          margin: 0 2px;
          border-radius: 2px;
          background-color: transparent;
          transition: all 0.3s;
          cursor: pointer;
        }
        .yg-chunk-marker:hover {
          background-color: rgba(12, 153, 255, 0.3);
          width: 8px;
        }
        .yg-chunk-marker.active-highlight {
          background-color: #0c99ff;
          width: 8px;
          box-shadow: 0 0 8px #0c99ff;
        }
        /* 针对高亮块的样式 */
        .active-highlight + * {
          background-color: rgba(12, 153, 255, 0.1);
          border-radius: 4px;
        }
      `}</style>
      {markdownData === '' ? (
        <Skeleton active paragraph={{ rows: 12 }} />
      ) : (
        <div
          className='markdown-body flex max-h-[70vh] flex-col overflow-auto pb-12 md:max-h-[75vh] !bg-transparent min-w-0 w-full'
          style={markdownContainerStyle}
        >
          <SafeMarkdownRenderer
            markdownData={markdownData}
            activeLocation={activeLocation}
            onLocationClick={onLocationClick}
          />
        </div>
      )}
    </main>
  )
}

export default Summary
