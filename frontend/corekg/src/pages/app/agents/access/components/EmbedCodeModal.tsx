import { FC, useMemo, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import {
  Modal,
  Input,
  Button,
  message,
  Spin,
  Tabs,
  Space,
  Typography,
} from 'antd'
import { CloseIcon } from 'tdesign-icons-react'
import { getExternalStatus } from '@/api/request'
import { copyText } from '@/utils'
import useAccessStore from '@/stores/access'

const { TextArea } = Input
const { TabPane } = Tabs
const { Text } = Typography

interface EmbedCodeModalProps {
  visible: boolean
  onClose: () => void
}

const EmbedCodeModal: FC<EmbedCodeModalProps> = ({ visible, onClose }) => {
  const { id } = useParams<{ id: string }>()
  const { botId: storeBotId, spaceId: storeSpaceId } = useAccessStore()
  
  const [loading, setLoading] = useState(false)
  const [statusData, setStatusData] = useState<any>(null)
  const [embedType, setEmbedType] = useState<'fullscreen' | 'widget'>(
    'fullscreen',
  )

  const finalBotId = storeBotId || id || ''
  const finalSpaceId = storeSpaceId || finalBotId

  useEffect(() => {
    if (visible && finalBotId) {
      const fetchStatus = async () => {
        setLoading(true)
        try {
          const res = await getExternalStatus({
            space_id: String(finalSpaceId),
            bot_id: String(finalBotId),
          })
          setStatusData(res)
        } catch (error) {
          console.error('获取Agent状态失败:', error)
          setStatusData(null)
        } finally {
          setLoading(false)
        }
      }

      fetchStatus()
    } else if (!visible) {
      setStatusData(null)
    }
  }, [visible, finalBotId, finalSpaceId])

  const handleClose = () => {
    onClose()
  }

  const baseUrl = useMemo(() => {
    const { protocol, hostname, port } = window.location
    return `${protocol}//${hostname}${port ? `:${port}` : ''}`
  }, [])

  const shortLinkCode = useMemo(() => {
    return statusData?.data?.short_link_code
  }, [statusData])

  // 全屏iframe代码
  const fullscreenCode = useMemo(() => {
    if (!shortLinkCode) {
      return '// 正在获取 Agent Code，请稍候...'
    }

    const iframeSrc = `${baseUrl}/iframe/detail/role/${shortLinkCode}`

    return `<iframe 
  src="${iframeSrc}"
  style="width: 100%; height: 100%; border: none;"
  frameborder="0"
  allow="microphone; camera"
  sandbox="allow-same-origin allow-scripts allow-forms allow-popups">
</iframe>`
  }, [shortLinkCode, baseUrl])

  // Widget代码
  const widgetCode = useMemo(() => {
    if (!shortLinkCode) {
      return '// 正在获取 Agent Code，请稍候...'
    }

    return `<script src="${baseUrl}/widget-loader.js"></script>

<script>
  document.addEventListener('DOMContentLoaded', function() {
    const widget = MyAIWidget({
      agentId: '${shortLinkCode}',
      agentType: 'role',
      position: 'bottom-right',
      width: 380,
      height: 600,
    });
  });
</script>`
  }, [shortLinkCode, baseUrl])

  const currentCode = embedType === 'fullscreen' ? fullscreenCode : widgetCode

  const handleCopy = () => {
    if (!shortLinkCode) {
      message.warning('Agent Code 尚未加载，无法复制代码')
      return
    }

    copyText(currentCode)
    message.success('已复制到剪贴板')
  }

  return (
    <Modal
      open={visible}
      onCancel={handleClose}
      width={720}
      footer={null}
      closeIcon={<CloseIcon />}
      styles={{
        body: { paddingTop: 16 },
      }}
    >
      <Spin spinning={loading} tip='正在加载信息...'>
        <div className='min-h-[400px]'>
          <Tabs
            activeKey={embedType}
            onChange={(key) => setEmbedType(key as 'fullscreen' | 'widget')}
            className='embed-code-tabs'
          >
            {/* 全屏嵌入标签页 */}
            <TabPane tab='全屏嵌入' key='fullscreen'>
              <Space direction='vertical' className='w-full' size={16}>
                <div className='bg-gray-50 rounded-lg p-3'>
                  <Text type='secondary' className='text-sm'>
                    <strong>全屏嵌入模式：</strong>
                    将对话界面嵌入到页面中，提供沉浸式的交互体验。
                    需要为iframe预留足够的空间。
                  </Text>
                </div>

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
                  />
                </div>

                <div className='bg-blue-50 border border-blue-200 rounded-lg p-3'>
                  <Text className='text-sm text-blue-700'>
                    <strong>提示：</strong>
                    建议将iframe放置在一个具有明确宽高的容器中， 以确保最佳的显示效果。
                  </Text>
                </div>
              </Space>
            </TabPane>
            
            {/* 小组件标签页 */}
            <TabPane tab='小组件' key='widget'>
              <Space direction='vertical' className='w-full' size={16}>
                <div className='bg-gray-50 rounded-lg p-3'>
                  <Text type='secondary' className='text-sm'>
                    <strong>悬浮按钮模式：</strong>
                    在页面角落显示一个悬浮按钮，用户点击后展开对话窗口。
                    适合不影响页面布局的场景。
                  </Text>
                </div>

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
                  />
                </div>
              </Space>
            </TabPane>
          </Tabs>

          <div className='flex items-center justify-end mt-6 pt-4'>
            <Space>
              <Button onClick={handleClose}>取消</Button>
              <Button
                type='primary'
                onClick={handleCopy}
                // 如果没有 shortLinkCode，禁止点击复制
                disabled={loading || !shortLinkCode}
                loading={loading}
              >
                复制代码
              </Button>
            </Space>
          </div>
        </div>
      </Spin>
    </Modal>
  )
}

export default EmbedCodeModal