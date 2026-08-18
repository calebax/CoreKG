import { FC, useState, useEffect } from 'react'
import { ConfigProvider, theme } from 'antd'
import {
  MessageOutlined,
  CloseOutlined,
  MinusOutlined,
} from '@ant-design/icons'
import Logo from '../../../components/Layout/MenuWrapper/images/logo.svg?react'
import DetailPromptWidget from '../detail/prompt'
import DetailQuestionWidget from '../detail/question'
import DetailRoleWidget from '../detail/role'
import './widget.css'

interface WidgetConfig {
  position?: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left'
  theme?: 'light' | 'dark'
  primaryColor?: string
  width?: number
  height?: number
  title?: string
  minimizable?: boolean
  type?: string
}

const EmbedWidget: FC<{ workflow?: boolean }> = (props) => {
  const { workflow } = props
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [config, setConfig] = useState<WidgetConfig>({
    position: 'bottom-right',
    theme: 'light',
    width: 380,
    height: 600,
    minimizable: true,
  })

  // 通知父页面widget状态变化
  const notifyParentStateChange = (open: boolean) => {
    window.parent.postMessage(
      { type: 'WIDGET_STATE_CHANGE', isOpen: open },
      '*',
    )
  }

  // 监听来自父页面的消息
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data.type === 'WIDGET_CONFIG') {
        setConfig((prev) => ({ ...prev, ...event.data.config }))
      }
      if (event.data.type === 'WIDGET_OPEN') {
        setIsOpen(true)
        notifyParentStateChange(true)
      }
      if (event.data.type === 'WIDGET_CLOSE') {
        setIsOpen(false)
        notifyParentStateChange(false)
      }
    }

    window.addEventListener('message', handleMessage)

    // 通知父页面widget已加载
    window.parent.postMessage({ type: 'WIDGET_READY' }, '*')

    return () => window.removeEventListener('message', handleMessage)
  }, [])

  // 监听isOpen状态变化
  useEffect(() => {
    notifyParentStateChange(isOpen)
  }, [isOpen])

  const renderContent = () => {
    if (workflow) return <DetailPromptWidget workflow />
    switch (config.type) {
      case 'role':
        return <DetailRoleWidget />
      case 'prompt':
        return <DetailPromptWidget />
      case 'question':
        return <DetailQuestionWidget />
      default:
        return <DetailRoleWidget />
    }
  }

  const positionClasses = {
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4',
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
  }

  return (
    <ConfigProvider
      theme={{
        algorithm:
          config.theme === 'dark'
            ? theme.darkAlgorithm
            : theme.defaultAlgorithm,
        token: {
          colorPrimary: config.primaryColor || '#1890ff',
        },
      }}
    >
      <div className='widget-container'>
        {/* 悬浮按钮 */}
        {!isOpen && (
          // <button
          //   className={`widget-trigger ${positionClasses[config.position!]}`}
          //   onClick={() => setIsOpen(true)}
          //   aria-label='打开对话'
          // >
          //   <MessageOutlined style={{ fontSize: 24 }} />
          // </button>
          <Logo
            className={`widget-trigger ${positionClasses[config.position!]}`}
            onClick={() => setIsOpen(true)}
          />
        )}

        {/* 主面板 */}
        {isOpen && (
          <div
            className={`widget-panel ${positionClasses[config.position!]} ${
              isMinimized ? 'minimized' : ''
            }`}
            style={{
              width: isMinimized ? 300 : config.width,
              height: isMinimized ? 60 : config.height,
            }}
          >
            {/* 标题栏 */}
            <div className='widget-header'>
              <div className='widget-controls'>
                {config.minimizable && (
                  <button
                    onClick={() => setIsMinimized(!isMinimized)}
                    className='widget-control-btn'
                  >
                    <MinusOutlined />
                  </button>
                )}
                <button
                  onClick={() => setIsOpen(false)}
                  className='widget-control-btn'
                >
                  <CloseOutlined />
                </button>
              </div>
            </div>

            {/* 内容区 */}
            {!isMinimized && (
              <div className='widget-content'>{renderContent()}</div>
            )}
          </div>
        )}
      </div>
    </ConfigProvider>
  )
}

export default EmbedWidget
