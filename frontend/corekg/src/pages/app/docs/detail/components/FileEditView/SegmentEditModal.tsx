import React, { useState, useEffect, useMemo } from 'react'
import { Modal, Input, Button, Image } from 'antd'
import { CloseOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import { useDeployConfig } from '@/utils/useDeployConfig'

interface SegmentEditModalProps {
  visible: boolean
  segmentId: string
  segmentTitle: string
  initialContent: string
  imageUrl?: string
  segmentType?: 'chunk' | 'table' | 'image'
  onCancel: () => void
  onConfirm: (segmentId: string, newContent: string) => Promise<void>
}

const { TextArea } = Input

type TableData = {
  type: 'markdown' | 'html'
  before: string
  after: string
  headers: string[]
  rows: string[][]
}

const splitTableRow = (line: string) =>
  line
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((cell) => cell.trim())

const buildMarkdownTable = (headers: string[], rows: string[][]) => {
  const headerLine = `| ${headers.join(' | ')} |`
  const dividerLine = `| ${headers.map(() => '---').join(' | ')} |`
  const rowLines = rows.map(
    (row) => `| ${headers.map((_, i) => row[i] ?? '').join(' | ')} |`,
  )
  return [headerLine, dividerLine, ...rowLines].join('\n')
}

const parseMarkdownTable = (markdown: string): TableData | null => {
  const lines = markdown.split('\n')
  const isTableLine = (line: string) => /\|/.test(line)
  const isDividerLine = (line: string) =>
    /^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?\s*$/.test(line)

  for (let i = 0; i < lines.length - 1; i += 1) {
    const headerLine = lines[i]
    const dividerLine = lines[i + 1]
    if (!isTableLine(headerLine) || !isDividerLine(dividerLine)) continue

    const headers = splitTableRow(headerLine)
    const rows: string[][] = []
    let endIndex = i + 2
    for (; endIndex < lines.length; endIndex += 1) {
      const line = lines[endIndex]
      if (!isTableLine(line) || line.trim() === '') break
      rows.push(splitTableRow(line))
    }

    const before = lines.slice(0, i).join('\n').trimEnd()
    const after = lines.slice(endIndex).join('\n').trimStart()
    return { type: 'markdown', before, after, headers, rows }
  }

  return null
}

const parseHtmlTable = (content: string): TableData | null => {
  const tableMatch = content.match(/<table[\s\S]*?>[\s\S]*?<\/table>/i)
  if (!tableMatch) return null

  const tableHtml = tableMatch[0]
  const before = content.slice(0, tableMatch.index ?? 0).trimEnd()
  const after = content
    .slice((tableMatch.index ?? 0) + tableHtml.length)
    .trimStart()

  const parser = new DOMParser()
  const doc = parser.parseFromString(tableHtml, 'text/html')
  const table = doc.querySelector('table')
  if (!table) return null

  const rows = Array.from(table.querySelectorAll('tr'))
  if (!rows.length) return null

  const getCells = (row: HTMLTableRowElement) =>
    Array.from(row.querySelectorAll('th, td')).map(
      (cell) => cell.textContent?.trim() || '',
    )

  const headers = getCells(rows[0])
  const bodyRows = rows.slice(1).map(getCells)

  return { type: 'html', before, after, headers, rows: bodyRows }
}

const buildHtmlTable = (headers: string[], rows: string[][]) => {
  const headerRow = `<tr>${headers
    .map((header) => `<td>${header}</td>`)
    .join('')}</tr>`
  const bodyRows = rows
    .map(
      (row) =>
        `<tr>${headers
          .map((_, index) => `<td>${row[index] ?? ''}</td>`)
          .join('')}</tr>`,
    )
    .join('')
  return `<table>${headerRow}${bodyRows}</table>`
}

const parseTableContent = (content: string): TableData | null => {
  const markdownTable = parseMarkdownTable(content)
  if (markdownTable) return markdownTable
  return parseHtmlTable(content)
}

const SegmentEditModal: React.FC<SegmentEditModalProps> = ({
  visible,
  segmentId,
  segmentTitle,
  initialContent,
  imageUrl,
  segmentType = 'chunk',
  onCancel,
  onConfirm,
}) => {
  const { t } = useTranslation('pages')
  const { version, mode } = useDeployConfig()
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(false)
  const [hasError, setHasError] = useState(false)

  // 是否启用增强模式：仅在 本地环境、测试环境、custom 版本且 cimc 模式下显示图片和表格编辑
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isEnhancedMode =
    isDevEnv || isTestEnv || (version === 'custom' && mode === 'cimc')

  // 字符计数限制
  const MAX_CHARS = 1024
  const currentChars = content.length
  const isEmpty = !content.trim()
  const isTableSegment = isEnhancedMode && segmentType === 'table'
  const isImageSegment = isEnhancedMode && segmentType === 'image'
  const tableData = useMemo(
    () => (isTableSegment ? parseTableContent(content) : null),
    [content, isTableSegment],
  )

  useEffect(() => {
    if (visible) {
      setContent(initialContent)
      setHasError(false)
    }
  }, [visible, initialContent])

  // 当内容变化时检查是否为空
  useEffect(() => {
    setHasError(isEmpty && content !== initialContent)
  }, [content, isEmpty, initialContent])

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value
    setContent(newContent)
  }

  const buildTableContent = (headers: string[], rows: string[][]) =>
    tableData?.type === 'markdown'
      ? buildMarkdownTable(headers, rows)
      : buildHtmlTable(headers, rows)

  const updateTableContent = (headers: string[], rows: string[][]) => {
    if (!tableData) return
    const tableContent = buildTableContent(headers, rows)
    setContent(tableContent)
  }

  const handleTableHeaderChange = (index: number, value: string) => {
    if (!tableData) return
    const nextHeaders = [...tableData.headers]
    nextHeaders[index] = value
    updateTableContent(nextHeaders, tableData.rows)
  }

  const handleTableCellChange = (
    rowIndex: number,
    colIndex: number,
    value: string,
  ) => {
    if (!tableData) return
    const nextRows = tableData.rows.map((row, idx) =>
      idx === rowIndex
        ? row.map((cell, cIdx) => (cIdx === colIndex ? value : cell))
        : row,
    )
    updateTableContent(tableData.headers, nextRows)
  }

  const handleConfirm = async () => {
    if (isEmpty) {
      setHasError(true)
      return
    }

    if (content.length > MAX_CHARS) {
      return
    }

    setLoading(true)
    try {
      await onConfirm(segmentId, content)
      onCancel() // 关闭弹窗
    } catch (error) {
      console.error('保存失败:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    setContent('')
    setHasError(false)
    onCancel()
  }

  return (
    <Modal
      open={visible}
      onCancel={handleCancel}
      centered
      closable={false}
      width={800}
      className='segment-edit-modal'
      footer={null}
    >
      <div className='relative py-1'>
        {/* 标题栏 */}
        <div className='flex items-center justify-between mb-4'>
          <h2 className="font-['Inter'] text-[#1e1f28] text-[16px] leading-[24px] font-medium">
            {t('app.docs.segmentEdit.editSegment', { segmentTitle })}
          </h2>
          <button
            onClick={handleCancel}
            className='cursor-pointer mt-[-20px] mr-[-4px]'
          >
            <CloseOutlined className='text-[#616373] text-base' />
          </button>
        </div>

        {/* 说明文字 */}
        <div className="mb-4 font-['PingFang_SC'] text-[#616373] text-sm leading-[22px]">
          <span className='rounded bg-[#EFF0F6] px-2 py-1 font-normal'>
            {t('app.docs.segmentEdit.editWarning')}
          </span>
        </div>

        {/* 内容编辑区域 */}
        <div className='mb-[6px]'>
          <div className={isImageSegment ? 'flex gap-1' : ''}>
            {isImageSegment && (
              <div className='w-[40%] flex items-center justify-center border border-[#d7d9e5] rounded-md p-2 bg-transparent'>
                <Image
                  src={imageUrl}
                  alt='segment'
                  className='max-h-[300px] object-contain'
                  style={{ width: '100%' }}
                />
              </div>
            )}
            {/* 输入框容器 */}
            <div
              className={`relative border rounded-md overflow-hidden ${isImageSegment ? 'w-[60%]' : 'w-full'}`}
              style={{
                borderColor: hasError ? '#F12623' : '#d7d9e5',
              }}
            >
              {/* 内容区域 - 可滚动 */}
              <div
                className={`${scrollStyles.scroll}`}
                style={{
                  height: '300px', // 内容区高度，微调以匹配图片区域
                  overflowY: 'auto',
                  overflowX: 'auto',
                  padding: '10px 0px 10px 4px',
                }}
              >
                <style>
                  {`
                    .segment-edit-modal .ant-input::placeholder {
                      color: #616373 !important;
                    }
                  `}
                </style>
                {tableData ? (
                  <div className='pr-3'>
                    <div className='overflow-x-auto overflow-y-visible max-w-full'>
                      <table
                        className='w-max border-collapse text-sm'
                        style={{ whiteSpace: 'nowrap' }}
                      >
                        <thead>
                          <tr>
                            {tableData.headers.map((header, index) => (
                              <th
                                key={`header-${index}`}
                                className='border border-[#d7d9e5] bg-[#F8F9FD] p-2 text-left whitespace-nowrap'
                              >
                                <Input
                                  value={header}
                                  onChange={(e) =>
                                    handleTableHeaderChange(
                                      index,
                                      e.target.value,
                                    )
                                  }
                                  bordered={false}
                                  className='p-0 min-w-[160px]'
                                />
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody>
                          {tableData.rows.map((row, rowIndex) => (
                            <tr key={`row-${rowIndex}`}>
                              {tableData.headers.map((_, colIndex) => (
                                <td
                                  key={`cell-${rowIndex}-${colIndex}`}
                                  className='border border-[#d7d9e5] p-2 whitespace-nowrap'
                                >
                                  <Input
                                    value={row[colIndex] ?? ''}
                                    onChange={(e) =>
                                      handleTableCellChange(
                                        rowIndex,
                                        colIndex,
                                        e.target.value,
                                      )
                                    }
                                    bordered={false}
                                    className='p-0 min-w-[160px]'
                                  />
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                ) : (
                  <TextArea
                    value={content}
                    onChange={handleContentChange}
                    placeholder={t('app.docs.segmentEdit.enterText')}
                    autoSize={false}
                    maxLength={MAX_CHARS}
                    className='resize-none'
                    bordered={false}
                    style={{
                      fontSize: '16px',
                      lineHeight: '24px',
                      fontFamily: 'Inter ',
                      padding: '0 4px 0 0',
                      minHeight: '100%',
                      fontWeight: 'normal',
                      color: '#1E1F28',
                    }}
                  />
                )}
              </div>

              {/* 字符计数 - 固定在底部，不跟随滚动 */}
              <div
                className='flex justify-end px-3 bg-white pb-[5px]'
                style={{
                  borderColor: '#e6e6e6',
                }}
              >
                <span className="font-['PingFang_SC'] text-[#616373] text-[12px] leading-[20px] font-normal">
                  {t('app.docs.segmentEdit.charCount', {
                    current: currentChars,
                    max: MAX_CHARS,
                  })}
                </span>
              </div>
            </div>
          </div>

          {/* 错误提示 */}
          {hasError && isEmpty && (
            <div className="mt-1 font-['PingFang_SC'] text-[#FF3B33] text-base font-normal leading-[24px]">
              {t('app.docs.segmentEdit.cannotBeEmpty')}
            </div>
          )}
        </div>

        {/* 按钮区域 */}
        <div className='flex justify-end gap-6 !font-medium'>
          <Button
            onClick={handleCancel}
            disabled={loading}
            style={{
              fontFamily: 'Inter ',
              fontSize: '16px',
              lineHeight: '24px',
              width: '64px',
              height: '34px',
              padding: '5px 16px',
              backgroundColor: '#EFF0F6',
              color: '#616373',
              fontWeight: 'medium',
              border: 'none',
            }}
          >
            {t('app.docs.segmentEdit.cancel')}
          </Button>
          <Button
            type='primary'
            onClick={handleConfirm}
            disabled={loading || isEmpty || currentChars > MAX_CHARS}
            style={{
              fontFamily: 'Inter ',
              fontSize: '16px',
              lineHeight: '24px',
              width: '64px',
              height: '34px',
              padding: '5px 16px',
              fontWeight: 'medium',
              backgroundColor:
                isEmpty || currentChars > MAX_CHARS ? '#EFF0F6' : '#0C99FF',
              color:
                isEmpty || currentChars > MAX_CHARS ? '#FFFFFF' : '#FFFFFF',
            }}
          >
            {loading
              ? t('app.docs.segmentEdit.saving')
              : t('app.docs.segmentEdit.done')}
          </Button>
        </div>
      </div>
    </Modal>
  )
}

export default SegmentEditModal
