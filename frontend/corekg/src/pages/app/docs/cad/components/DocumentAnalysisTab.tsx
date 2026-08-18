import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { message } from 'antd'
import { getFileInfo, getDocContent } from '@/api/knowledge'
import Summary from '@/components/Summary'

export default function DocumentAnalysisTab() {
  const { id, fileId } = useParams<{ id: string; fileId: string }>()
  const [parseStatus, setParseStatus] = useState('pending')
  const [docAnalysisData, setDocAnalysisData] = useState<string>('')
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

  return (
    <div className='h-full flex flex-col min-w-0'>
      {/* 内容区域 */}
      <div className='flex-1 min-h-0'>
        <Summary
          isForestPage={false}
          markdownData={docAnalysisData}
          flag={true}
        />
      </div>
    </div>
  )
}
