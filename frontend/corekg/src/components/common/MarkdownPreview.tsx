import React, { useState, useEffect, useMemo, useRef } from 'react'
import { Spin, Alert } from 'antd'
import 'github-markdown-css'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import rehypeRaw from 'rehype-raw'
import remarkGfm from 'remark-gfm'
import '@/styles/markdown.css'
import { processReferenceMarkdown } from '@/utils/referenceProcessor'
import { useDeployConfig } from '@/utils/useDeployConfig'
import EchartsBlock from './EchartsBlock'
import ReferencePreviewModal from './ReferencePreviewModal'

interface MarkdownPreviewProps {
  content?: string
  file?: string
  className?: string
  style?: React.CSSProperties
  /** echarts编辑回调 */
  onEchartsEdit?: (id: string, config: any) => void
  /** 图表配置映射表 */
  chartConfigMap?: Map<string, any>
  /** 引用跳转 */
  references?: any
  /** 是否禁用引用标注功能 */
  disableReference?: boolean
  /** 定位信息，用于滚动到指定位置 */
  location?: number[]
  /** locationKey 用于跟踪URL变化 */
  locationKey?: string
}

const getReferenceChunkUniqueKey = (chunk: any) => {
  if (chunk?.chunk_id) return `chunk_id:${chunk.chunk_id}`
  if (chunk?.sequence !== undefined) return `sequence:${chunk.sequence}`
  return JSON.stringify(chunk)
}

const buildMergedReferenceByMarker = (
  references: any[] = [],
  docId: number,
  chunkId?: string,
  chunkIndex?: string,
) => {
  const sequence = Number(chunkIndex)
  const fileMatchedReferences = references.filter((ref) => ref.file_id === docId)
  if (!fileMatchedReferences.length) return null

  const baseReference = fileMatchedReferences[0]
  const mergedChunkMap = new Map<string, any>()

  fileMatchedReferences.forEach((ref) => {
    ref.chunk_list?.forEach((chunk: any) => {
      mergedChunkMap.set(getReferenceChunkUniqueKey(chunk), chunk)
    })
  })

  const mergedChunkList = Array.from(mergedChunkMap.values())
  const sortedChunkList = mergedChunkList.sort((a: any, b: any) => {
    const aSequence =
      typeof a.sequence === 'number' ? a.sequence : Number.MAX_SAFE_INTEGER
    const bSequence =
      typeof b.sequence === 'number' ? b.sequence : Number.MAX_SAFE_INTEGER
    return aSequence - bSequence
  })

  const matchedChunk =
    sortedChunkList.find(
      (chunk: any) =>
        (chunkId && chunk.chunk_id === chunkId) ||
        (!Number.isNaN(sequence) && chunk.sequence === sequence),
    ) ?? sortedChunkList[0]

  return {
    ...baseReference,
    chunk_list: sortedChunkList,
    matchedChunk,
  }
}

/**
 * 预处理 Markdown 文本，手动渲染 KaTeX 公式并处理Reference格式
 * @param markdown - 原始的 Markdown 字符串
 * @param disableReference - 是否禁用引用标注功能
 * @returns {string} - 处理过的、包含 HTML 的 Markdown 字符串
 */
const preprocessMarkdownWithKatex = (
  markdown: string,
  disableReference: boolean = false,
): string => {
  if (!markdown) return ''

  // 1. 处理Reference格式
  let processedReference: string
  if (disableReference) {
    // 禁用引用标注时，直接移除所有Reference格式的字符串
    processedReference = markdown.replace(/\{Reference[^}]*\}/g, '')
  } else {
    // 启用引用标注时，转换为数字标注
    processedReference = processReferenceMarkdown(markdown)
  }

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

  // 2. 处理 LaTeX 风格块级公式 (\[ ... \]) - 接口常返回此格式
  let processed = processedReference.replace(/\\\[([\s\S]+?)\\\]/g, (_, formula) =>
    renderBlockFormula(formula),
  )

  // 3. 处理 $ 风格块级公式 ($$ ... $$)
  processed = processed.replace(/\$\$([\s\S]+?)\$\$/g, (_, formula) =>
    renderBlockFormula(formula),
  )

  // 4. 处理 LaTeX 风格行内公式 (\( ... \)) - 接口常返回此格式
  processed = processed.replace(/\\\(([\s\S]+?)\\\)/g, (_, formula) =>
    renderInlineFormula(formula),
  )

  // 5. 处理 $ 风格行内公式 ($ ... $)
  processed = processed.replace(
    /\$((?:[^\$]|\\[\$])+?)\$/g,
    (_, formula) => renderInlineFormula(formula),
  )

  return processed
}

export default function MarkdownPreview({
  content,
  file,
  className = '',
  style = {},
  onEchartsEdit,
  chartConfigMap,
  references,
  disableReference = false,
  location,
  locationKey,
}: MarkdownPreviewProps) {
  const { version } = useDeployConfig()
  const markdownWrapRef = useRef<HTMLDivElement>(null)
  const [fileContent, setFileContent] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(!!file)
  const [error, setError] = useState<string | null>(null)
  const prevLocationKeyRef = useRef<string | undefined>(undefined)
  const [visible, setVisible] = useState<boolean>(false)
  const [modalModel, setModalModel] = useState<any>(null)
  const [clickPosition, setClickPosition] = useState<
    { x: number; y: number } | undefined
  >(undefined)

  useEffect(() => {
    if (content) {
      setFileContent(content)
      setLoading(false)
      setError(null)
      return
    }

    if (file) {
      setLoading(true)
      setError(null)

      fetch(file)
        .then((response) => {
          if (!response.ok) {
            throw new Error(
              `Failed to fetch file: ${response.status} ${response.statusText}`,
            )
          }
          return response.text()
        })
        .then((text) => {
          setFileContent(text)
          setLoading(false)
        })
        .catch((err) => {
          console.error('Error loading file:', err)
          setError(`加载文件失败: ${err.message}`)
          setLoading(false)
        })
    }
  }, [content, file])

  const processedContent = useMemo(() => {
    const displayContent = content || fileContent
    const contentWithKatex = preprocessMarkdownWithKatex(
      displayContent,
      disableReference,
    )
    const contentWithoutIllegalHtmlTag = contentWithKatex.replace(
      /<\/?([^>]+)>/g,
      (match, el: string) => {
        const tagName = el.trim().split(/\s+/)[0]?.replace(/^\/|\/$/g, '')
        if (tagName && /^[a-z][a-z0-9-]*$/i.test(tagName)) {
          return match
        }
        return match.replace('<', '&lt;').replace('>', '&gt;')
      },
    )
    return contentWithoutIllegalHtmlTag
  }, [content, fileContent, disableReference])

  // 监听 location 变化并滚动到指定位置
  useEffect(() => {
    if (!location?.length || !markdownWrapRef.current || loading) {
      return
    }

    // 仅在 locationKey 变化时触发（避免重复滚动）
    const locationKeyChanged = prevLocationKeyRef.current !== locationKey
    if (!locationKeyChanged && prevLocationKeyRef.current !== undefined) {
      return
    }
    prevLocationKeyRef.current = locationKey

    const timer = setTimeout(() => {
      const container = markdownWrapRef.current
      if (!container) return

      const rawPosition = location[0]
      if (typeof rawPosition !== 'number' || Number.isNaN(rawPosition)) {
        return
      }

      const maxScrollTop = container.scrollHeight - container.clientHeight
      const clampedTop = Math.max(0, Math.min(maxScrollTop, rawPosition))

      container.scrollTo({ top: clampedTop, behavior: 'smooth' })
    }, 300)

    return () => clearTimeout(timer)
  }, [location, locationKey, loading])

  // 处理流式输出中的Reference格式（仅在未禁用时）
  useEffect(() => {
    if (disableReference) return

    const markdownContainer =
      markdownWrapRef.current?.querySelector('.markdown-body')
    if (!markdownContainer) return

    const html = markdownContainer.innerHTML
    // 检查是否包含Reference格式且未处理过
    if (
      html.includes('{Reference') &&
      !markdownContainer.querySelector('.reference-annotation')
    ) {
      const processedHtml = processReferenceMarkdown(html)
      if (processedHtml !== html) {
        markdownContainer.innerHTML = processedHtml
      }
    }
  }, [content, fileContent, disableReference])

  const handleMove = (event: React.MouseEvent<HTMLDivElement>) => {
    if (disableReference) return

    const target = event.target as HTMLElement
    if (target.classList.contains('reference-annotation')) {
      // 在hover
      const tipDom = target.querySelector(
        '.reference-annotation-tip',
      ) as HTMLSpanElement
      if (!tipDom) return
      const rect = target.getBoundingClientRect()
      const wrapRect = markdownWrapRef.current!.getBoundingClientRect()
      const leftWidth = rect.x - wrapRect.x + rect.width
      const tipOffset = rect.y - wrapRect.y - rect.height
      const tipText = '点击预览引文'
      tipDom.innerText = tipText
      const width = tipText.length * 14 + 20
      // 假设宽度比左边宽度大，往右偏移
      const minRight = Math.min(leftWidth - width, 0)
      const minTop = tipOffset > 0 ? -34 : 22

      Object.assign(tipDom.style, {
        width: width + 'px',
        right: minRight + 'px',
        top: minTop + 'px',
      } as React.CSSProperties)
    }
  }

  // 处理点击外部关闭弹窗
  useEffect(() => {
    if (!visible || !clickPosition) return

    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement
      if (!target.closest('.reference-modal-container')) {
        closeModal()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [visible, clickPosition])

  const closeModal = () => {
    setVisible(false)
    setModalModel(null)
    setClickPosition(undefined)
  }

  const handleAnnotationClick = (event: React.MouseEvent<HTMLDivElement>) => {
    if (disableReference) return

    const target = event.target as HTMLElement
    if (!target.classList.contains('reference-annotation')) return

    event.preventDefault()
    event.stopPropagation()

    const docId = target.getAttribute('data-doc-id')!
    const reference = buildMergedReferenceByMarker(
      references,
      +docId,
      target.dataset.chunkId,
      target.dataset.chunkIndex,
    )
    if (!reference) return

    // 获取点击位置
    const rect = target.getBoundingClientRect()
    setClickPosition({
      x: rect.left + rect.width / 2,
      y: rect.bottom,
    })
    setModalModel({
      forest_id: reference.forest_id,
      docId: +docId,
      chunkId: reference.matchedChunk?.chunk_id || target.dataset.chunkId,
      chunkIndex:
        String(reference.matchedChunk?.sequence ?? '') ||
        target.dataset.chunkIndex,
    })
    setVisible(true)
  }

  // 使链接在新标签页打开
  const components = {
    // 自定义链接渲染
    a: ({ node, ...props }: any) => (
      <a {...props} target='_blank' rel='noopener noreferrer' />
    ),
    // 自定义图片渲染
    img: ({ node, src, ...props }: any) => {
      let finalSrc = src
      // 修复私有化版本图片路径缺失前缀的问题
      // 这里的逻辑是：如果是 custom 版本，且 src 不为空，包含 corekg-bucket，则提取 corekg-bucket 及其后面的部分，然后添加当前域名+/的前缀
      if (
        version === 'custom' &&
        finalSrc &&
        finalSrc.includes('corekg-bucket')
      ) {
        // 提取 corekg-bucket 及其后面的部分
        const bucketIndex = finalSrc.indexOf('corekg-bucket')
        const bucketPath = finalSrc.substring(bucketIndex)
        // 拼接完整的 URL
        finalSrc = `${window.location.origin}/${bucketPath}`
      }
      return <img src={finalSrc} {...props} style={{ maxWidth: '100%' }} />
    },
    // 自定义代码块渲染，支持echarts
    code: ({ node, inline, className, children, ...props }: any) => {
      // ECharts 代码块现在由 processEchartsBlocks 处理，这里只处理普通代码块
      return (
        <code className={className} {...props}>
          {children}
        </code>
      )
    },
  }

  const containerStyle: React.CSSProperties = {
    height: '100%',
    width: '100%',
    overflow: 'auto',
    ...style,
  }

  if (loading) {
    return (
      <div
        className='flex items-center justify-center h-full w-full p-8'
        style={containerStyle}
      >
        <Spin tip='加载文件中...' />
      </div>
    )
  }

  if (error) {
    return (
      <div className='p-4' style={containerStyle}>
        <Alert message='错误' description={error} type='error' showIcon />
      </div>
    )
  }

  const displayContent = content || fileContent

  // 直接在 markdown 内容中处理 ECharts 和 SQL 代码块
  const processSpecialBlocks = (markdownContent: string): React.ReactNode[] => {
    const parts: React.ReactNode[] = []
    const specialBlocksRegex = /```(echarts|sql)\n([\s\S]*?)```/g
    let lastIndex = 0
    let match
    let blockIndex = 0

    while ((match = specialBlocksRegex.exec(markdownContent)) !== null) {
      // 添加代码块之前的普通内容
      if (match.index > lastIndex) {
        const beforeContent = markdownContent.substring(lastIndex, match.index)
        // 处理 Reference 格式
        let processedBeforeContent: string
        if (disableReference) {
          // 禁用引用标注时，直接移除所有Reference格式的字符串
          processedBeforeContent = beforeContent.replace(
            /\{Reference[^}]*\}/g,
            '',
          )
        } else {
          // 启用引用标注时，转换为数字标注
          processedBeforeContent = processReferenceMarkdown(beforeContent)
        }
        parts.push(
          <ReactMarkdown
            key={`before-${blockIndex}`}
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeHighlight, rehypeRaw]}
            components={components}
          >
            {processedBeforeContent}
          </ReactMarkdown>,
        )
      }

      // 处理特殊代码块
      const blockType = match[1] // echarts 或 sql
      const blockContent = match[2].trim()
      const blockId = `${blockType}-processed-${blockIndex}`

      if (blockType === 'echarts') {
        try {
          // 验证 JSON 格式
          JSON.parse(blockContent)
          const updatedConfig = chartConfigMap?.get(blockId)
          const finalConfig = updatedConfig
            ? JSON.stringify(updatedConfig, null, 2)
            : blockContent

          parts.push(
            <EchartsBlock
              key={blockId}
              config={finalConfig}
              id={blockId}
              editable={true}
              onEdit={onEchartsEdit}
            />,
          )
        } catch (error) {
          console.error('ECharts JSON 解析错误:', error)
          // 解析错误时直接跳过，不显示任何内容
        }
      } else if (blockType === 'sql') {
        // 处理 SQL 代码块
        parts.push(
          <div
            key={blockId}
            className='border border-gray-300 rounded-lg p-4 bg-gray-50 my-4'
          >
            <div className='flex items-center mb-2'>
              <span className='text-sm font-medium text-gray-700 bg-gray-200 px-2 py-1 rounded'>
                SQL
              </span>
            </div>
            <pre className='text-sm bg-gray-900 text-green-400 p-3 rounded overflow-auto font-mono'>
              <code>{blockContent}</code>
            </pre>
          </div>,
        )
      }

      lastIndex = match.index + match[0].length
      blockIndex++
    }

    // 添加最后剩余的内容
    if (lastIndex < markdownContent.length) {
      const afterContent = markdownContent.substring(lastIndex)
      // 处理 Reference 格式
      let processedAfterContent: string
      if (disableReference) {
        // 禁用引用标注时，直接移除所有Reference格式的字符串
        processedAfterContent = afterContent.replace(/\{Reference[^}]*\}/g, '')
      } else {
        // 启用引用标注时，转换为数字标注
        processedAfterContent = processReferenceMarkdown(afterContent)
      }
      parts.push(
        <ReactMarkdown
          key={`after-${blockIndex}`}
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeHighlight, rehypeRaw]}
          components={components}
        >
          {processedAfterContent}
        </ReactMarkdown>,
      )
    }

    return parts
  }

  if (!displayContent) {
    return (
      <div className='p-4' style={containerStyle}>
        <Alert
          message='提示'
          description='没有内容可显示'
          type='info'
          showIcon
        />
      </div>
    )
  }

  // 处理特殊代码块 (ECharts 和 SQL)
  const processedParts = processSpecialBlocks(displayContent)
  const hasSpecialBlocks =
    processedParts.length > 1 ||
    (processedParts.length === 1 &&
      processedParts.some(
        (part) =>
          React.isValidElement(part) &&
          (part.key?.toString().includes('echarts') ||
            part.key?.toString().includes('sql')),
      ))

  return (
    <>
      <div
        className={`markdown-body ${className}`}
        style={{
          padding: '12px 16px',
          overflowY: 'auto',
          overflowX: 'hidden',
          maxWidth: '100%',
          height: '100%',
          boxSizing: 'border-box',
          borderRadius: '4px',
          backgroundColor: '#F8F9FD',
          color: '#0C1F17',
          fontSize: '16px',
          fontWeight: '400',
          ...style,
        }}
        ref={markdownWrapRef}
        onClick={handleAnnotationClick}
        onMouseMove={handleMove}
      >
        {hasSpecialBlocks ? (
          processedParts
        ) : (
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeHighlight, rehypeRaw]}
            components={components}
          >
            {processedContent}
          </ReactMarkdown>
        )}
      </div>
      {clickPosition && (
        <ReferencePreviewModal
          visible={visible}
          onClose={closeModal}
          modalModel={modalModel}
          references={references}
          position={clickPosition}
        />
      )}
    </>
  )
}
