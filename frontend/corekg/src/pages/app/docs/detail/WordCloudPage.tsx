import React, { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Breadcrumb } from 'antd'
import { HomeOutlined } from '@ant-design/icons'
import { getKnowledgeBaseDetail } from '@/api/knowledge'
import KnowledgeRefresh from '@/assets/icons/knowledge-refresh.svg?react'
import KnowledgeGraphComponent from './components/KnowledgeGraph'
import WordCloudComponent from './components/WordCloud'

interface WordCloudPageProps {
  knowledgeBaseName?: string
}

const WordCloudPage: React.FC<WordCloudPageProps> = ({
  knowledgeBaseName: propKnowledgeBaseName,
}) => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()

  // 从URL参数获取tab状态，如果没有则默认为wordcloud
  const [activeTab, setActiveTab] = useState<'wordcloud' | 'graph'>(() => {
    // 根据url地址中是否包含knowledge-graph或者wordcloud来判断当前是哪个tab
    const url = window.location.href
    const tabFromUrl = url.includes('knowledge-graph')
      ? 'graph'
      : url.includes('wordcloud')
        ? 'wordcloud'
        : 'wordcloud'
    return tabFromUrl
  })

  const [knowledgeBaseName, setKnowledgeBaseName] = useState(
    propKnowledgeBaseName || '',
  )
  // 从URL参数获取选中的节点ID
  const [selectedNodeId, setSelectedNodeId] = useState<string | undefined>(
    () => {
      return searchParams.get('nodeId') || undefined
    },
  )
  const wordCloudRef = useRef<any>(null)
  const knowledgeGraphRef = useRef<any>(null)
  const wordCloudContainerRef = useRef<HTMLDivElement>(null) // 词云外部容器的ref

  useEffect(() => {
    // 如果没有传入知识库名称，则从 API 获取
    if (!propKnowledgeBaseName && id) {
      const fetchKnowledgeBaseName = async () => {
        try {
          const response = await getKnowledgeBaseDetail({ id: Number(id) })
          if (response?.data?.name) {
            setKnowledgeBaseName(response.data.name)
          }
        } catch (error) {
          console.error('获取知识库信息失败:', error)
        }
      }
      fetchKnowledgeBaseName()
    }
  }, [id, propKnowledgeBaseName])

  const handleResetView = () => {
    // 重置视角，只重置视图不重新获取数据
    if (activeTab === 'wordcloud' && wordCloudRef.current?.resetView) {
      wordCloudRef.current.resetView()
    } else if (activeTab === 'graph' && knowledgeGraphRef.current?.resetView) {
      knowledgeGraphRef.current.resetView()
    }
  }

  const handleTabChange = (tab: 'wordcloud' | 'graph') => {
    setActiveTab(tab)
    if (tab === 'wordcloud') {
      // 切换到词云时清除选中的节点
      setSelectedNodeId(undefined)
      setSearchParams({ tab })
    } else {
      // 切换到知识图谱时保持当前选中的节点
      const params: any = { tab }
      if (selectedNodeId) {
        params.nodeId = selectedNodeId
      }
      setSearchParams(params)
    }
  }

  // 页面初始化或者切换词云和知识图谱的时候先调用一下重置视图方法
  useEffect(() => {
    setTimeout(() => {
      handleResetView()
    }, 100)
  }, [activeTab])

  // 处理词云点击事件，切换到知识图谱并设置选中节点
  const handleWordCloudClick = (word: string) => {
    console.log(`词云点击: ${word}，切换到知识图谱并获取相关节点数据`)
    setSelectedNodeId(word) // 设置选中的节点ID
    setActiveTab('graph')
    setSearchParams({ tab: 'graph', nodeId: word }) // 同时更新URL参数
  }

  // 返回到文件/文件夹列表页
  const handleBackToFileList = () => {
    navigate(`/docs/detail/${id}`)
  }

  // 构建面包屑导航项
  const getBreadcrumbItems = () => {
    const items = [
      {
        title: (
          <span
            className='text-[#86909C] font-medium text-base cursor-pointer hover:text-[#0C99FF]'
            onClick={() => navigate('/docs')}
          >
            知识库
          </span>
        ),
      },
    ]

    // 知识库名称作为上一页，也就是文件/文件夹列表页
    if (knowledgeBaseName) {
      items.push({
        title: (
          <span
            className='text-[#86909C] font-medium text-base cursor-pointer hover:text-[#0C99FF]'
            onClick={handleBackToFileList}
          >
            {knowledgeBaseName}
          </span>
        ),
      })
    }

    // 动态判断当前页面是词云还是知识图谱
    if (activeTab === 'wordcloud') {
      items.push({
        title: (
          <span className='text-[#000000E5] font-medium text-base'>词云</span>
        ),
      })
    } else {
      items.push({
        title: (
          <span className='text-[#000000E5] font-medium text-base'>
            知识图谱
          </span>
        ),
      })
    }

    return items
  }

  return (
    <div className='overflow-hidden h-full p-4 bg-white'>
      <div className='w-full h-full rounded-lg overflow-hidden flex flex-col gap-5'>
        {/* 顶部导航区域 */}
        <div className='flex-shrink-0 flex items-center justify-between p-1'>
          {/* 面包屑导航 */}
          <Breadcrumb
            separator='>'
            className='text-base'
            items={getBreadcrumbItems()}
          />
          {/* Tab 切换按钮 - 参考knownow项目样式，位置在右上角 */}
          <div className='flex overflow-hidden rounded-md border border-black/10'>
            <button
              className={`cursor-pointer w-15 h-9 px-4 font-medium text-sm transition-colors duration-300 !border-none ${activeTab === 'wordcloud' ? 'bg-[#165DFF] font-medium text-white' : 'text-[#165DFF] bg-black/[0.02] hover:text-[#1677ff]'}`}
              onClick={() => handleTabChange('wordcloud')}
            >
              词云
            </button>
            <button
              className={`cursor-pointer w-26 h-9 px-4 text-sm font-medium transition-colors duration-300 !border-none ${activeTab === 'graph' ? 'bg-[#165DFF] font-medium text-white' : 'text-[#165DFF] bg-black/[0.02] hover:text-[#1677ff]'}`}
              onClick={() => handleTabChange('graph')}
            >
              知识图谱
            </button>
          </div>
        </div>

        {/* 主要内容区域 */}
        <div className='flex-1 overflow-hidden relative bg-[#F5F7FA] rounded-lg'>
          {/* 一键回归视角按钮 */}
          <div className='absolute top-4 right-4 z-10'>
            <button
              onClick={handleResetView}
              className='bg-white hover:bg-gray-100 text-gray-800 font-semibold py-1 px-3 border border-gray-200 rounded shadow text-sm flex items-center'
            >
              <HomeOutlined style={{ marginRight: 4 }} />
              重置视图
            </button>
          </div>

          {/* 内容展示区域 */}
          <div className='w-full h-full p-4'>
            {activeTab === 'wordcloud' ? (
              <div
                className='w-full h-full overflow-hidden select-none cursor-grabbing'
                ref={wordCloudContainerRef}
              >
                <WordCloudComponent
                  ref={wordCloudRef}
                  knowledgeBaseId={id}
                  onWordClick={handleWordCloudClick}
                  containerRef={wordCloudContainerRef}
                />
              </div>
            ) : (
              <div className='w-full h-full overflow-hidden border-none'>
                <KnowledgeGraphComponent
                  ref={knowledgeGraphRef}
                  knowledgeBaseId={id}
                  selectedNodeId={selectedNodeId}
                />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default WordCloudPage
