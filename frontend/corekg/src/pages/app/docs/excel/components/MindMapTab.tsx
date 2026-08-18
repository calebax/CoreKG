import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { message, Skeleton, Button, Tooltip } from 'antd'
import { getFileInfo, getMindMapContent } from '@/api/knowledge'
import KnowledgeRefresh from '@/assets/icons/knowledge-refresh.svg?react'
import ForestGraph from '@/components/ForestGraph'

export default function MindMapTab() {
  const { id, fileId } = useParams<{ id: string; fileId: string }>()
  const [graphStatus, setGraphStatus] = useState('success') // 模拟成功状态
  const [mindData, setMindData] = useState<any>({})

  // 获取文件信息
  const getAnalysisStatus = async () => {
    try {
      const res = await getFileInfo({
        file_id: Number(fileId),
      })
      setGraphStatus(res.graph_status)
    } catch (e) {
      console.error(e)
    }
  }

  // 获取思维导图
  const fetchMindData = async () => {
    try {
      const res = await getMindMapContent({
        file_id: Number(fileId),
      })
      if (res.status === 'success' || graphStatus === 'success') {
        setMindData(JSON.parse(res.mind_map))
      }
    } catch (error) {
      message.warning(error as string)
    }
  }

  useEffect(() => {
    if (id && fileId) {
      getAnalysisStatus()
      fetchMindData()
    }
  }, [id, fileId])

  // 轮询
  useEffect(() => {
    const interval = setInterval(() => {
      getAnalysisStatus()
      if (graphStatus === 'success') {
        clearInterval(interval)
      }
    }, 10000)

    return () => clearInterval(interval)
  }, [graphStatus, id, fileId])

  return (
    <div className='h-full flex flex-col min-w-0'>
      {/* 内容区域 */}
      <div className='flex-1 overflow-hidden relative'>
        {Object.keys(mindData).length === 0 ? (
          <div className='py-2 pl-2 pt-4'>
            <Skeleton active paragraph={{ rows: 12 }} />
          </div>
        ) : (
          <>
            {mindData.id && <ForestGraph renderData={mindData} />}

            {/* 重置视图按钮 - 右上角位置 */}
            <div className='absolute top-0 right-0 z-10 flex gap-2'>
              <Tooltip title='重置视图' placement='left'>
                <Button
                  type='default'
                  icon={<KnowledgeRefresh />}
                  size='small'
                  className='!bg-transparent !border-none !shadow-none'
                  onClick={() => {
                    // 触发ForestGraph重新渲染来重置视图
                    setMindData({})
                    setTimeout(() => {
                      setMindData(mindData)
                    }, 50)
                  }}
                />
              </Tooltip>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
