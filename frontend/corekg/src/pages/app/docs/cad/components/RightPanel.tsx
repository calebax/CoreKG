import { useState, useEffect, FC } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button } from 'antd'
import ErrorBoundary from '@/router/ErrorBoundary'
import DocumentAnalysisTab from './DocumentAnalysisTab'
import IntelligentAnalysisTab from './IntelligentAnalysisTab'
import MindMapTab from './MindMapTab'
import WisdomQATab from './WisdomQATab'

interface TabItem {
  key: string
  label: string
  component: React.ComponentType
}

export default function RightPanel() {
  const location = useLocation()
  const navigate = useNavigate()

  // 从URL查询参数中获取当前选中的tab，如果没有则默认为'document-analysis'
  const getActiveTabFromUrl = (): string => {
    try {
      const searchParams = new URLSearchParams(location.search)
      const tabFromUrl = searchParams.get('tab')
      return tabFromUrl &&
        [
          'document-analysis',
          'intelligent-analysis',
          'mind-map',
          'wisdom-qa',
        ].includes(tabFromUrl)
        ? tabFromUrl
        : 'wisdom-qa'
    } catch (error) {
      console.error('解析tab参数失败', error)
      return 'wisdom-qa'
    }
  }

  const [activeTab, setActiveTab] = useState<string>(getActiveTabFromUrl())

  // 监听URL变化，同步activeTab状态
  useEffect(() => {
    const newActiveTab = getActiveTabFromUrl()
    setActiveTab(newActiveTab)
  }, [location.search])

  const tabItems: TabItem[] = [
    {
      key: 'wisdom-qa',
      label: '智能问答',
      component: WisdomQATab,
    },
    {
      key: 'intelligent-analysis',
      label: '智能摘要',
      component: IntelligentAnalysisTab,
    },
    {
      key: 'document-analysis',
      label: '文档解析',
      component: DocumentAnalysisTab,
    },

    // {
    //   key: 'mind-map',
    //   label: '思维导图',
    //   component: MindMapTab,
    // },
  ]

  const handleTabChange = (key: string) => {
    // 更新URL查询参数
    const searchParams = new URLSearchParams(location.search)
    searchParams.set('tab', key)

    // 使用replace而不是push，避免产生过多的历史记录
    navigate(
      {
        pathname: location.pathname,
        search: searchParams.toString(),
      },
      { replace: true },
    )

    // 本地状态会通过useEffect自动更新
  }

  const renderActiveComponent = () => {
    const activeItem = tabItems.find((item) => item.key === activeTab)
    if (!activeItem) return null
    const Component = activeItem.component
    return <Component />
  }

  return (
    <ErrorBoundary fallback={<CannotParse tabItems={tabItems} />}>
      <div className='h-full flex flex-col bg-[#F3F8FF] px-4 py-6 gap-6 min-w-0'>
        {/* Tab按钮组 */}
        <div className='flex gap-0 flex-shrink-0 overflow-x-auto custom-tab-scroll'>
          {tabItems.map((item, index) => {
            const isActive = activeTab === item.key
            const isFirst = index === 0
            const isLast = index === tabItems.length - 1
            return (
              <Button
                key={item.key}
                onClick={() => handleTabChange(item.key)}
                className={`h-8 px-4 text-sm font-medium !rounded-none transition-all duration-200 !border-none flex-shrink-0
                ${isActive ? '!bg-[#4080FF] !text-white' : ' !text-[#606266] hover:!text-[#0C99FF]'}
                ${isFirst ? '!rounded-l-md' : ''}
                ${isLast ? '!rounded-r-md' : ''}
              `}
                style={{
                  boxShadow: 'none',
                }}
              >
                {item.label}
              </Button>
            )
          })}
        </div>

        {/* 内容区域 */}
        <div className='flex-1 min-h-0 overflow-hidden'>
          {renderActiveComponent()}
        </div>
      </div>
    </ErrorBoundary>
  )
}

const CannotParse: FC<{ error?: Error; tabItems: TabItem[] }> = (props) => {
  const { error, tabItems } = props
  return (
    <div className='h-full flex flex-col bg-[#F3F8FF] px-4 py-6 gap-6 min-w-0'>
      {/* Tab按钮组 */}
      <div className='flex gap-0 flex-shrink-0 overflow-x-auto custom-tab-scroll'>
        {tabItems.map((item, index) => {
          const isFirst = index === 0
          const isLast = index === tabItems.length - 1
          return (
            <Button
              disabled
              key={item.key}
              className={`h-8 px-4 text-sm font-medium !rounded-none transition-all duration-200 !border-none flex-shrink-0
                ${isFirst ? '!rounded-l-md' : ''}
                ${isLast ? '!rounded-r-md' : ''}
              `}
              style={{
                boxShadow: 'none',
              }}
            >
              {item.label}
            </Button>
          )
        })}
      </div>
      {error?.message}
    </div>
  )
}
