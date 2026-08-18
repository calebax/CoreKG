/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useState, useEffect, useRef } from 'react'
import { Spin } from '@coze-arch/coze-design'
import jsPreviewExcel from '@js-preview/excel'
import '@js-preview/excel/lib/index.css'

interface XLSXViewerProps {
  file: string
}

// Excel容器样式 - 防止滚动传递到浏览器导航
const containerStyles: React.CSSProperties = {
  height: '100%',
  width: '100%',
  minHeight: '300px',
  overscrollBehaviorX: 'contain', // 防止横向滚动传递到父级
  overscrollBehavior: 'contain', // 防止滚动传递
  WebkitOverflowScrolling: 'touch', // 保持iOS的原生滚动体验
}

const XLSXViewer: React.FC<XLSXViewerProps> = ({ file }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let excelPreviewer: any = null

    const initExcelPreview = async () => {
      if (!containerRef.current || !file) {
        setError('容器或文件URL无效')
        setLoading(false)
        return
      }

      try {
        // 初始化预览器
        excelPreviewer = jsPreviewExcel.init(containerRef.current)

        // 预览文件
        await excelPreviewer.preview(file)
        console.log('Excel预览加载完成')
        setLoading(false)
      } catch (err) {
        console.error('Excel预览加载失败:', err)
        setError('无法加载Excel文件')
        setLoading(false)
      }
    }

    initExcelPreview()

    // 清理函数
    return () => {
      if (excelPreviewer && typeof excelPreviewer.destroy === 'function') {
        excelPreviewer.destroy()
      }
    }
  }, [file])

  return (
    <div className='h-full w-full flex flex-col'>
      {loading && (
        <div className='absolute inset-0 flex items-center justify-center bg-white bg-opacity-70 z-10'>
          <Spin tip='加载表格中...' />
        </div>
      )}

      {error && (
        <div className='h-full flex items-center justify-center'>
          <div className='text-center text-red-500'>
            <p>加载表格失败: {error}</p>
          </div>
        </div>
      )}

      <div
        ref={containerRef}
        className='flex-1 overflow-auto excel-preview-container'
        style={containerStyles}
      />
    </div>
  )
}

export default XLSXViewer
