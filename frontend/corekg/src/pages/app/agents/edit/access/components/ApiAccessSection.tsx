import { FC, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Button, message } from 'antd'
import { useRequest } from 'ahooks'
import { CopyIcon } from 'tdesign-icons-react'
import { getAgentInfo } from '@/api'
import { copyText } from '@/utils'
import { useDeployConfig } from '@/utils/useDeployConfig'
import ApiKeyManagementModal from './ApiKeyManagementModal'

const ApiAccessSection: FC = () => {
  const { id } = useParams<{ id: string }>()
  const [apiKeyModalVisible, setApiKeyModalVisible] = useState(false)
  const { version } = useDeployConfig()
  const agentId = id ? parseInt(id, 10) : 0
  // const baseUrl = `https://api.example.com/v3/chat.Agent`

  const { protocol, hostname, port } = window.location

  const { data: agentName } = useRequest(async () => {
    const { agent_info, name } = await getAgentInfo(agentId)
    return (name ?? agent_info.name) as string
  })

  const baseUrl = useMemo(() => {
    return `${protocol}//${hostname}${port ? `:${port}` : ''}/v3/chat.Agent`
  }, [])

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
        <h3 className='text-base font-medium text-gray-900'>API访问</h3>
      </div>

      <div className='space-y-4 flex flex-col gap-2'>
        {/* Base URL */}
        <div className='flex items-center justify-between '>
          <div className='flex items-center bg-[#F5F5F5] text-[#6E757F] px-1 rounded'>
            <span className='text-gray-600'>Base URL: {baseUrl}</span>

            <button
              onClick={() => handleCopy(baseUrl)}
              className='ml-2 p-1 hover:bg-gray-100 rounded border-none bg-transparent cursor-pointer'
            >
              <CopyIcon size='14px' className='text-gray-500' />
            </button>
          </div>
          <Button
            size='small'
            className='bg-[#0C99FF33] border-[#0C99FF33] ml-auto'
            onClick={() => setApiKeyModalVisible(true)}
          >
            API Key
          </Button>
        </div>
        {agentName ? (
          <span className='text-base font-medium'>
            名称：{agentName}
            <button
              onClick={() => handleCopy(agentName)}
              className='p-1 hover:bg-gray-100 rounded border-none bg-transparent cursor-pointer'
            >
              <CopyIcon size='14px' className='text-gray-500' />
            </button>
          </span>
        ) : null}
        {/* API文档 */}
        <div>
          <div className='text-base font-medium'>API文档</div>
          <div className='mt-4 mr-auto inline-flex items-center gap-2 bg-[#F5F5F5] text-[#6E757F] px-1 rounded'>
            <a
              href={apiDocsUrl}
              target='_blank'
              rel='noopener noreferrer'
              className='text-inherit'
            >
              {apiDocsUrl}
            </a>
            <button
              onClick={() => handleCopy(baseUrl)}
              className='p-1 hover:bg-gray-100 rounded border-none bg-transparent cursor-pointer'
            >
              <CopyIcon size='14px' className='text-gray-500' />
            </button>
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
