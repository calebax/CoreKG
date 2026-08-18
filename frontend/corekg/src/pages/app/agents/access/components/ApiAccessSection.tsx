import { FC, useState, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { message } from 'antd'
import { useRequest } from 'ahooks'
import { CopyIcon } from 'tdesign-icons-react'
import { copyText } from '@/utils'
import { getExternalStatus } from '@/api/request'
import useAccessStore from '@/stores/access'
import { useDeployConfig } from '@/utils/useDeployConfig'
import ApiKeyManagementModal from './ApiKeyManagementModal'

const ApiAccessSection: FC = () => {
  const { id } = useParams<{ id: string }>()
  const { botId: storeBotId, spaceId: storeSpaceId } = useAccessStore()
  const [apiKeyModalVisible, setApiKeyModalVisible] = useState(false)
  const { version } = useDeployConfig()
  const agentId = storeBotId || id ? parseInt(storeBotId || id || '', 10) : 0

  const { protocol, hostname, port } = window.location

  const { data: externalStatus } = useRequest(async () => {
    return await getExternalStatus({
      space_id: storeSpaceId || String(agentId),
      bot_id: storeBotId || String(agentId),
    })
  })

  const baseUrl = useMemo(() => {
    return `${protocol}//${hostname}${port ? `:${port}` : ''}/v3/chat.ChatAgent`
  }, [protocol, hostname, port])

  const apiDocsUrl =
    version === 'saas'
      ? 'https://docs.corekg.com/docs/corekg/'
      : `${protocol}//${hostname}${port ? `:${port}` : ''}/usage_help/`

  const handleCopy = (text: string) => {
    copyText(text)
    message.success('已复制到剪贴板')
  }

  return (
    <div>
      <div className='flex items-center gap-2 my-6'>
        <div className='h-4 w-1 rounded-full bg-[#165DFF]' />
        <h3 className='text-base font-medium text-gray-900'>API访问</h3>
      </div>

      <div className='space-y-4 flex flex-col gap-2'>
        {/* Base URL */}
        <div className='flex items-center justify-between pr-16'>
          <div className='flex items-center'>
            <span className='text-gray-600'>Base URL: {baseUrl}</span>

            <button
              onClick={() => handleCopy(baseUrl)}
              className='p-1 hover:bg-gray-100 rounded border-none bg-transparent cursor-pointer'
            >
              <CopyIcon size='14px' className='text-gray-500' />
            </button>
          </div>
          {/* <Button
            size='small'
            className='ml-4'
            onClick={() => setApiKeyModalVisible(true)}
          >
            API Key
          </Button> */}
        </div>
        {externalStatus?.data?.short_link_code ? (
          <span className='text-gray-600'>
            名称：{externalStatus?.data?.short_link_code}
          </span>
        ) : null}
        {/* API文档 */}
        <div>
          <div className='text-gray-900 mb-2'>API文档</div>
          <div className='flex items-center gap-2'>
            <a
              href={apiDocsUrl}
              target='_blank'
              rel='noopener noreferrer'
              className='text-[#0C99FF] hover:text-blue-600'
            >
              {apiDocsUrl}
            </a>
          </div>
        </div>
      </div>

      {/* API Key管理弹窗 */}
      <ApiKeyManagementModal
        visible={apiKeyModalVisible}
        onClose={() => setApiKeyModalVisible(false)}
        agentId={agentId}
      />
    </div>
  )
}

export default ApiAccessSection
