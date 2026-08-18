import { useState, useEffect, useRef, FC } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button, message } from 'antd'
import { cn } from '@/utils'
import { getFileInfo, retryParse } from '@/api/knowledge'
import ErrorBoundary from '@/router/ErrorBoundary'
import DocumentAnalysisTab from './DocumentAnalysisTab'
import { useFileDetailViewProject } from './FileDetailView'
import IntelligentAnalysisTab from './IntelligentAnalysisTab'
import Qa from './Qa'
import SegmentList from './SegmentList'
import EmptyIcon from './images/empty.svg?react'

interface TabItem {
  key: string
  label: string
  component: React.ComponentType
}

export default function RightPanel(props: { activeKey: string }) {
  const location = useLocation()
  const navigate = useNavigate()
  const { fileId: file_id } = useFileDetailViewProject<{ fileId: number }>()!
  const [statusLoading, setStatusLoading] = useState<boolean>(false)
  // 这里分别保存不同维度的状态，便于按 Tab 精细控制展示逻辑
  const [knowledgeStatus, setKnowledgeStatus] = useState<string | null>(null)
  const [parseStatus, setParseStatus] = useState<string | null>(null)
  const [descStatus, setDescStatus] = useState<string | null>(null)
  const timerRef = useRef<NodeJS.Timeout | null>(null)

  const normalizeStatus = (value?: string | null) => {
    if (!value) return null
    // 后端有可能带 ```markdown 包装，这里统一清理掉
    return value.replace(/```markdown\s*|\s*```/g, '')
  }

  const isFinishedStatus = (value?: string | null) => {
    const normalized = normalizeStatus(value)
    if (!normalized) return false
    return ['success', 'fail', 'unsupported', 'error'].includes(normalized)
  }

  const loadStatus = async () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
    }
    setStatusLoading(true)
    const fileInfo = await getFileInfo({
      file_id,
    })
    // 这里分别记录三种状态，后面会根据不同 Tab 使用不同的状态字段
    setKnowledgeStatus(normalizeStatus(fileInfo.knowledge_status))
    setParseStatus(normalizeStatus(fileInfo.parse_status))
    setDescStatus(normalizeStatus(fileInfo.desc_status))

    // 只要还有任一维度状态未到终态，就继续轮询，避免状态不同步
    if (
      !isFinishedStatus(fileInfo.knowledge_status) ||
      !isFinishedStatus(fileInfo.parse_status) ||
      !isFinishedStatus(fileInfo.desc_status)
    ) {
      timerRef.current = setTimeout(() => {
        loadStatus()
      }, 10000)
    }

    setStatusLoading(false)
  }

  useEffect(() => {
    loadStatus()
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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

  const handleReload = async () => {
    try {
      await retryParse(file_id)
      await loadStatus()
      message.success('重试成功')
    } catch (err) {
      message.error('重试失败，请稍后再试')
    }
  }

  // 监听URL变化，同步activeTab状态
  useEffect(() => {
    const newActiveTab = getActiveTabFromUrl()
    setActiveTab(newActiveTab)
  }, [location.search])

  const tabItems: TabItem[] = [
    {
      key: 'wisdom-qa',
      label: '智能问答',
      component: Qa,
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
    {
      key: 'segmentRule',
      label: '分段规则',
      component: SegmentList as React.ComponentType,
    },
  ]

  const renderEmpty = (panelStatus: 'loading' | 'error') => {
    return (
      <div className='flex flex-col w-full h-full justify-center items-center text-[#616373] gap-[6px]'>
        <EmptyIcon />
        {panelStatus === 'error' ? (
          <div>
            <span>解析失败,</span>
            {/* 解析失败时的重试文案需要展示为可点击，因此这里增加鼠标样式为pointer */}
            <span
              onClick={handleReload}
              className='text-[#0c99ff] cursor-pointer'
            >
              点击重试
            </span>
          </div>
        ) : (
          '资源处理中，请稍后~'
        )}
      </div>
    )
  }

  const getPanelStatusByActiveKey = () => {
    let rawStatus: string | null = null
    // 不同 Tab 使用不同的后端状态字段
    if (props.activeKey === 'intelligent-analysis') {
      rawStatus = descStatus
    } else if (props.activeKey === 'document-analysis') {
      rawStatus = parseStatus
    } else {
      // 智能问答、分段规则等沿用原有的 knowledge_status 逻辑
      rawStatus = knowledgeStatus
    }

    const normalized = normalizeStatus(rawStatus)
    if (!normalized) return null
    if (['fail', 'error', 'unsupported'].includes(normalized)) {
      return 'error' as const
    }
    if (normalized === 'success') {
      return 'loaded' as const
    }
    return 'loading' as const
  }

  const renderActiveComponent = () => {
    const activeItem = tabItems.find((item) => item.key === props.activeKey)
    if (!activeItem) return null
    const Component = activeItem.component

    const panelStatus = getPanelStatusByActiveKey()
    if (!panelStatus) return null
    if (panelStatus !== 'loaded') {
      return renderEmpty(panelStatus)
    }

    return <Component />
  }

  return (
    <ErrorBoundary fallback={<CannotParse tabItems={tabItems} />}>
      <div
        className={cn('h-full flex-1 flex flex-col px-4 py-6 gap-6 min-w-0', {
          'p-[0]': activeTab === 'wisdom-qa',
        })}
      >
        {/* Tab按钮组 */}
        {/* <div className='flex gap-0 flex-shrink-0 overflow-x-auto custom-tab-scroll'>
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
        </div> */}

        {/* 内容区域 */}
        <div
          className={cn('flex-1 min-h-0 overflow-hidden', {
            hidden: props.activeKey !== 'wisdom-qa',
          })}
        >
          <Qa />
        </div>
        {props.activeKey !== 'wisdom-qa' ? (
          <div className='flex-1 min-h-0 overflow-hidden'>
            {renderActiveComponent()}
          </div>
        ) : null}
      </div>
    </ErrorBoundary>
  )
}

const CannotParse: FC<{ error?: Error; tabItems: TabItem[] }> = (props) => {
  const { error, tabItems } = props
  return (
    <div className='h-full flex-1 flex flex-col  px-4 py-6 gap-6 min-w-0'>
      {/* Tab按钮组 */}
      {/* <div className='flex gap-0 flex-shrink-0 overflow-x-auto custom-tab-scroll'>
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
      </div> */}
      {error?.message}
    </div>
  )
}
