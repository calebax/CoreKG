import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Button, message } from 'antd'
import { getFileInfo, getDocContent } from '@/api/knowledge'
import Summary from '@/components/Summary'
import { copyToClipboard } from '@/utils/copy'
import CopyIcon from './images/copy.svg?react'
import { useFileDetailViewProject } from './FileDetailView'

export default function DocumentAnalysisTab() {
  const { id, fileId } = useParams<{ id: string; fileId: string }>()
  const { activeChunkId, segments, setActiveChunkId } =
    useFileDetailViewProject()!
  const [parseStatus, setParseStatus] = useState('pending')
  const [docAnalysisData, setDocAnalysisData] = useState<string>('')

  // 获取当前激活的位置标记内容
  const activeLocationContent = useEffect(() => {
    if (!activeChunkId || !segments) return undefined
    const segment = segments.find((s: any) => s.id === activeChunkId)
    if (!segment || !segment.yg_location) return undefined
    // 提取 <!--yg_pos...yg_pos--> 中的内容
    const match = segment.yg_location.match(/<!--yg_pos(.*?)yg_pos-->/)
    return match ? match[1].trim() : undefined
  }, [activeChunkId, segments])

  const handleLocationClick = (locContent: string) => {
    const segment = segments?.find((s: any) =>
      s.yg_location?.includes(locContent),
    )
    if (segment) {
      setActiveChunkId(segment.id)
    }
  }

  if (parseStatus === 'unsupported') {
    throw new Error('该文件暂不支持解析')
  }
  // 获取文件信息
  const getAnalysisStatus = async () => {
    try {
      const res = await getFileInfo({
        file_id: Number(fileId),
      })
      setParseStatus(res.parse_status.replace(/```markdown\s*|\s*```/g, ''))
    } catch (e) {
      console.error(e)
    }
  }

  // 获取文档分析
  const fetchDocData = async () => {
    try {
      const res = await getDocContent({
        file_id: Number(fileId),
      })
      if (res.status === 'success' || parseStatus === 'success') {
        setDocAnalysisData(res.content.replace(/```markdown\s*|\s*```/g, ''))
      }
    } catch (error) {
      message.warning(error as string)
    }
  }

  useEffect(() => {
    if (id && fileId) {
      getAnalysisStatus()
      fetchDocData()
    }
  }, [id, fileId])

  // 轮询
  useEffect(() => {
    const interval = setInterval(() => {
      getAnalysisStatus()
      if (parseStatus === 'success') {
        clearInterval(interval)
      }
    }, 10000)

    return () => clearInterval(interval)
  }, [parseStatus, id, fileId])

  const handleCopy = async () => {
    try {
      // 移除形如 <!--yg_pos开头 yg_pos-->结尾 的注释
      const cleanedData = docAnalysisData.replace(
        /<!--yg_pos[\s\S]*?yg_pos-->/g,
        '',
      )
      await copyToClipboard(cleanedData)
      message.success('复制成功')
    } catch {
      message.error('复制失败')
    }
  }

  // 重新计算 activeLocationContent
  const currentActiveLocation = (() => {
    if (!activeChunkId || !segments) return undefined
    const segment = segments.find((s: any) => s.id === activeChunkId)
    if (!segment || !segment.yg_location) return undefined
    const match = segment.yg_location.match(/<!--yg_pos(.*?)yg_pos-->/)
    return match ? match[1].trim() : undefined
  })()

  return (
    <div className='h-full flex flex-col min-w-0'>
      <div className='flex items-center justify-end pb-[10px]'>
        <Button
          onClick={handleCopy}
          className='text-[#0c99ff] border-[#0c99ff] flex gap-[4px] '
        >
          <CopyIcon />
          复制
        </Button>
      </div>
      {/* 内容区域 */}
      <div className='flex-1 min-h-0'>
        <Summary
          isForestPage={false}
          markdownData={docAnalysisData}
          flag={true}
          activeLocation={currentActiveLocation}
          onLocationClick={handleLocationClick}
        />
      </div>
    </div>
  )
}
