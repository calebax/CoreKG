import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { message } from 'antd'
import { getIntelContent, retryParse } from '@/api/knowledge'
import Summary from '@/components/Summary'
import EmptyIcon from './images/empty.svg?react'

export default function IntelligentAnalysisTab() {
  const { fileId } = useParams<{ id: string; fileId: string }>()
  const [status, setStatus] = useState<'pending' | 'success' | 'fail'>('pending')
  const [data, setData] = useState<string>('')

  // 获取智能分析数据
  const fetchData = useCallback(async () => {
    if (!fileId) return
    try {
      const res = await getIntelContent({ file_id: Number(fileId) })
      setStatus(res.status as 'pending' | 'success' | 'fail')
      setData(
        res.status === 'success'
          ? res.analysis.replace(/```markdown\s*|\s*```/g, '')
          : '',
      )
    } catch (error) {
      console.log(error as string)
      setStatus('fail')
      setData('')
    }
  }, [fileId])

  // 重试解析
  const handleRetry = async () => {
    if (!fileId) return
    try {
      await retryParse(Number(fileId))
      message.success('重试成功')
      fetchData()
    } catch {
      console.log('重试失败，请稍后再试')
    }
  }

  // 初始加载
  useEffect(() => {
    fetchData()
  }, [fetchData])

  // 轮询：pending状态时每10秒轮询一次
  useEffect(() => {
    if (status !== 'pending') return
    const interval = setInterval(fetchData, 10000)
    return () => clearInterval(interval)
  }, [status, fetchData])

  const isSuccess = status === 'success' && data

  return (
    <div className='h-full flex flex-col'>
      <div className='flex-1 overflow-hidden'>
        {isSuccess ? (
          <Summary
            isForestPage={false}
            markdownData={data}
            intelligentAnalysisTab={true}
          />
        ) : (
          <div className='flex flex-col w-full h-full justify-center items-center text-[#616373] gap-[12px]'>
            <EmptyIcon />
            {status === 'fail' ? (
              <div className='text-[14px] leading-[22px]'>
                <span>摘要生成失败，</span>
                <span
                  onClick={handleRetry}
                  className='text-[#0c99ff] cursor-pointer hover:underline'
                >
                  点击重试
                </span>
              </div>
            ) : (
              <p className='text-[14px] leading-[22px]'>摘要生成中，请稍后～</p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
