import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { message } from 'antd'
import {
  getFileInfo,
  getIntelContent,
  saveIntelParse,
  exportIntelParse,
} from '@/api/knowledge'
import Summary from '@/components/Summary'

export default function IntelligentAnalysisTab() {
  const { id, fileId } = useParams<{ id: string; fileId: string }>()
  const [analysisStatus, setAnalysisStatus] = useState('pending')
  const [intelAnalysisData, setIntelAnalysisData] = useState<string>('')

  // 获取文件信息
  const getAnalysisStatus = async () => {
    try {
      const res = await getFileInfo({
        file_id: Number(fileId),
      })
      setAnalysisStatus(res.analysis_status)
    } catch (e) {
      console.error(e)
    }
  }

  // 获取智能分析
  const fetchIntelData = async () => {
    try {
      const res = await getIntelContent({
        file_id: Number(fileId),
      })
      if (res.status === 'success' || analysisStatus === 'success') {
        setIntelAnalysisData(res.analysis.replace(/```markdown\s*|\s*```/g, ''))
      }
    } catch (error) {
      message.warning(error as string)
    }
  }

  useEffect(() => {
    if (id && fileId) {
      getAnalysisStatus()
      fetchIntelData()
    }
  }, [id, fileId])

  // 轮询
  useEffect(() => {
    const interval = setInterval(() => {
      getAnalysisStatus()
      if (analysisStatus === 'success') {
        clearInterval(interval)
      }
    }, 10000)

    return () => clearInterval(interval)
  }, [analysisStatus, id, fileId])

  return (
    <>
      <div className='h-full flex flex-col'>
        {/* 内容区域 */}
        <div className='flex-1 overflow-hidden'>
          <Summary
            isForestPage={false}
            markdownData={intelAnalysisData}
            intelligentAnalysisTab={true}
          />
        </div>
      </div>
    </>
  )
}
