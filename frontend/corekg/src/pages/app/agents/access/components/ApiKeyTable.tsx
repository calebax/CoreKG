import { FC, useEffect, useState } from 'react'
import { Table, Switch, Button, message, Popconfirm } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useRequest } from 'ahooks'
import { CopyIcon } from 'tdesign-icons-react'
import { copyText } from '@/utils'
import {
  getAgentApiKeyList,
  deleteAgentApiKey,
  setAgentApiKeyStatus,
} from '@/api/apiKey'

interface ApiKeyItem {
  ID: number
  api_key: string
  status: string
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: string | null
  Uin: number
  company_id: number
  expired_at: string
  name: string
  purpose: string
  resource_id: number
  resource_type: string
}

interface ApiKeyTableProps {
  agentId: number
  refreshKey: number
  onRefresh: () => void
}

const ApiKeyTable: FC<ApiKeyTableProps> = ({
  agentId,
  refreshKey,
  onRefresh,
}) => {
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  })

  const {
    data: apiKeyData,
    loading,
    run: fetchApiKeys,
  } = useRequest(
    (page = 1, size = 20) =>
      getAgentApiKeyList({
        limit: size,
        offset: (page - 1) * size,
        orderBy: ['created_at desc'],
        filters: [{ field: 'agent_id', value: [agentId.toString()] }],
      }),
    {
      manual: true,
      onSuccess: (result) => {
        setPagination((prev) => ({
          ...prev,
          total: result.total || result.Data?.length || 0,
        }))
      },
    },
  )

  const { loading: deleting, run: deleteApiKey } = useRequest(
    (apiKeyId: number) =>
      deleteAgentApiKey({
        agent_id: agentId,
        apikey_id: apiKeyId,
      }),
    {
      manual: true,
      onSuccess: () => {
        message.success('删除成功')
        onRefresh()
      },
      onError: (error) => {
        message.error(error.message || '删除失败')
      },
    },
  )

  const { loading: statusUpdating, run: updateStatus } = useRequest(
    ({ apiKeyId, status }: { apiKeyId: number; status: string }) =>
      setAgentApiKeyStatus({
        agent_id: agentId,
        apikey_id: apiKeyId,
        status,
      }),
    {
      manual: true,
      onSuccess: () => {
        message.success('状态更新成功')
        onRefresh()
      },
      onError: (error) => {
        message.error(error.message || '状态更新失败')
      },
    },
  )

  useEffect(() => {
    if (agentId) {
      fetchApiKeys(pagination.current, pagination.pageSize)
    }
  }, [agentId, refreshKey, pagination.current, pagination.pageSize])

  const handleStatusChange = (record: ApiKeyItem, checked: boolean) => {
    const newStatus = checked ? 'normal' : 'disabled'
    updateStatus({
      apiKeyId: record.ID,
      status: newStatus,
    })
  }

  const handleDelete = (record: ApiKeyItem) => {
    deleteApiKey(record.ID)
  }

  // 脱敏处理API Key
  const maskApiKey = (key: string) => {
    if (!key || key.length < 10) return key
    const start = key.slice(0, 5)
    const end = key.slice(-4)
    const middle = '*'.repeat(Math.min(key.length - 9, 20))
    return `${start}${middle}${end}`
  }

  const handleCopyApiKey = (key: string) => {
    copyText(key)
    message.success('API Key已复制到剪贴板')
  }

  const columns: ColumnsType<ApiKeyItem> = [
    {
      title: 'API key',
      dataIndex: 'api_key',
      key: 'api_key',
      width: '40%',
      render: (key: string) => (
        <div className='flex items-center gap-2'>
          <span className='font-mono text-sm'>{maskApiKey(key)}</span>
          <Button
            type='text'
            size='small'
            className='text-gray-400 hover:text-gray-600'
            onClick={() => handleCopyApiKey(key)}
          >
            <CopyIcon size='14px' />
          </Button>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: '15%',
      render: (status: string, record: ApiKeyItem) => (
        <Switch
          checked={status === 'normal'}
          onChange={(checked) => handleStatusChange(record, checked)}
          loading={statusUpdating}
          size='small'
        />
      ),
    },
    {
      title: '创建日期',
      dataIndex: 'CreatedAt',
      key: 'CreatedAt',
      width: '25%',
      render: (createdAt: string) => formatDateTime(createdAt),
    },
    {
      title: '操作',
      key: 'actions',
      width: '20%',
      render: (_, record: ApiKeyItem) => (
        <div className='flex gap-2'>
          <Popconfirm
            title='确认删除'
            description='删除后无法恢复，确定要删除这个API Key吗？'
            onConfirm={() => handleDelete(record)}
            okText='确定'
            cancelText='取消'
            styles={{
              root: { maxWidth: '280px' },
              body: { padding: '16px 20px', borderRadius: '8px' },
            }}
          >
            <Button
              type='link'
              size='small'
              loading={deleting}
              className='text-red-500 hover:text-red-600 p-0'
            >
              删除
            </Button>
          </Popconfirm>
        </div>
      ),
    },
  ]

  const handleTableChange = (paginationConfig: any) => {
    setPagination((prev) => ({
      ...prev,
      current: paginationConfig.current,
      pageSize: paginationConfig.pageSize,
    }))
  }

  return (
    <div>
      <Table
        columns={columns}
        dataSource={apiKeyData?.Data || []}
        loading={loading}
        rowKey='ID'
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) =>
            `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          pageSizeOptions: ['10', '20', '50', '100'],
        }}
        onChange={handleTableChange}
        size='middle'
      />
    </div>
  )
}

const formatDateTime = (dateString: string): string => {
  if (!dateString) return '-'

  try {
    const date = new Date(dateString)
    return date
      .toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
      })
      .replace(/\//g, '-')
  } catch (error) {
    return dateString
  }
}

export default ApiKeyTable
