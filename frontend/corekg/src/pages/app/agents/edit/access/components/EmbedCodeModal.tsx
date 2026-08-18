import { FC, useMemo, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import {
  Input,
  Button,
  message,
  Spin,
  Tabs,
  Space,
  Typography,
  Drawer,
} from 'antd'
import { CloseIcon } from 'tdesign-icons-react'
import { match } from 'ts-pattern'
import { getExternalStatus } from '@/api'
import { copyText } from '@/utils'
import styles from '../../styles.module.scss'

const { TextArea } = Input
const { Text } = Typography

interface EmbedCodeModalProps {
  visible: boolean
  onClose: () => void
}

const EmbedCodeModal: FC<EmbedCodeModalProps> = ({ visible, onClose }) => {
  const { id } = useParams<{ id: string }>()
  const [externalStatus, setExternalStatus] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [embedType, setEmbedType] = useState<'fullscreen' | 'widget'>(
    'fullscreen',
  )

  useEffect(() => {
    if (visible && id && !externalStatus) {
      const fetchExternalStatus = async () => {
        setLoading(true)
        try {
          const status = await getExternalStatus({
            space_id: id,
            bot_id: id,
          })
          setExternalStatus(status)
        } catch (error) {
          console.error('获取Agent信息失败:', error)
          message.error('获取Agent信息失败')
        } finally {
          setLoading(false)
        }
      }

      fetchExternalStatus()
    }
  }, [visible, id, externalStatus])

  const handleClose = () => {
    onClose()
    setExternalStatus(null)
  }

  const baseUrl = useMemo(() => {
    const { protocol, hostname, port } = window.location
    return `${protocol}//${hostname}${port ? `:${port}` : ''}`
  }, [])

  // 获取Agent路径
  const getAgentPath = (isWidget: boolean = false): string => {
    const prefix = isWidget ? '/iframe/widget' : '/iframe/detail'
    return `${prefix}/role/${encodeURIComponent(
      externalStatus?.short_link_code || '',
    )}`
  }

  // 全屏iframe代码
  const fullscreenCode = useMemo(() => {
    if (!externalStatus?.short_link_code) {
      return '// 等待 Agent 信息加载...'
    }

    const iframeSrc = `${baseUrl}${getAgentPath(false)}`

    return `<iframe 
  src="${iframeSrc}"
  style="width: 100%; height: 100%; border: none;"
  frameborder="0"
  allow="microphone; camera"
  sandbox="allow-same-origin allow-scripts allow-forms allow-popups">
</iframe>`
  }, [externalStatus, baseUrl])

  // Widget代码
  const widgetCode = useMemo(() => {
    if (!externalStatus?.short_link_code) {
      return '// 等待 Agent 信息加载...'
    }

    return `<!-- 步骤1: 在页面中引入加载器脚本 -->
<script src="${baseUrl}/widget-loader.js"></script>

<!-- 步骤2: 初始化小组件 -->
<script>
  document.addEventListener('DOMContentLoaded', function() {
    const widget = MyAIWidget({
      agentId: '${externalStatus.short_link_code}',
      agentType: 'role',
      position: 'bottom-right', // 可选: bottom-left, top-right, top-left
      width: 380,
      height: 600,
    });
  });
</script>`
  }, [externalStatus, baseUrl])

  // 获取当前选中的代码
  const currentCode = embedType === 'fullscreen' ? fullscreenCode : widgetCode

  const handleCopy = () => {
    if (!externalStatus) {
      message.error('Agent 信息尚未加载完成')
      return
    }

    if (!currentCode || currentCode.includes('等待')) {
      message.error('代码尚未生成')
      return
    }

    copyText(currentCode)
    message.success('已复制到剪贴板')
  }

  return (
    <Drawer
      open={visible}
      onClose={handleClose}
      width={926}
      title={
        <Tabs
          activeKey={embedType}
          onChange={(key) => setEmbedType(key as 'fullscreen' | 'widget')}
          className='embed-code-tabs'
          items={[
            { label: '全屏嵌入', key: 'fullscreen' },
            { label: '小组件', key: 'widget' },
          ]}
        ></Tabs>
      }
      closeIcon={<CloseIcon />}
      className={styles.drawer}
      styles={{
        body: { paddingTop: 16 },
      }}
    >
      <Spin spinning={loading} tip='正在加载Agent信息...'>
        <div className='min-h-[400px]'>
          {match(embedType)
            .with('fullscreen', () => (
              <Space direction='vertical' className='w-full' size={16}>
                {/* 模式说明 */}
                <div className='bg-gray-50 rounded-lg p-3'>
                  <Text type='secondary' className='text-sm'>
                    <strong>全屏嵌入模式：</strong>
                    将对话界面嵌入到页面中，提供沉浸式的交互体验。
                    需要为iframe预留足够的空间。
                  </Text>
                </div>

                {/* 代码区域 */}
                <div>
                  <TextArea
                    value={fullscreenCode}
                    rows={8}
                    readOnly
                    className='font-mono text-xs text-black'
                    style={{
                      resize: 'none',
                      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
                    }}
                    placeholder={
                      loading ? '正在加载 Agent 信息...' : '等待 Agent 信息'
                    }
                  />
                </div>

                {/* 提示信息 */}
                <div className='bg-blue-50 border border-blue-200 rounded-lg p-3'>
                  <Text className='text-sm text-blue-700'>
                    <strong>提示：</strong>
                    建议将iframe放置在一个具有明确宽高的容器中，
                    以确保最佳的显示效果。
                  </Text>
                </div>
              </Space>
            ))
            .with('widget', () => (
              <Space direction='vertical' className='w-full' size={16}>
                {/* 模式说明 */}
                <div className='bg-gray-50 rounded-lg p-3'>
                  <Text type='secondary' className='text-sm'>
                    <strong>悬浮按钮模式：</strong>
                    在页面角落显示一个悬浮按钮，用户点击后展开对话窗口。
                    适合不影响页面布局的场景。
                  </Text>
                </div>

                {/* 代码区域 */}
                <div>
                  <TextArea
                    value={widgetCode}
                    rows={12}
                    readOnly
                    className='font-mono text-xs text-black'
                    style={{
                      resize: 'none',
                      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
                    }}
                    placeholder={
                      loading ? '正在加载 Agent 信息...' : '等待 Agent 信息'
                    }
                  />
                </div>
              </Space>
            ))
            .exhaustive()}
          {/* 底部操作栏 */}
          <div className='flex items-center justify-end mt-6 pt-4'>
            <Space>
              <Button onClick={handleClose}>取消</Button>
              <Button
                type='primary'
                onClick={handleCopy}
                disabled={!externalStatus || loading}
                loading={loading}
              >
                复制代码
              </Button>
            </Space>
          </div>
        </div>
      </Spin>
    </Drawer>
  )
}

export default EmbedCodeModal
