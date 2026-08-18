import React, { useState, useEffect } from 'react'
import { Button, message, Modal, Tooltip } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { getApiKeyList, createApiKey, deleteApiKey } from '@/api/apiKey'
import Alert from '@/components/common/Alert'
import TableWithPagination from '@/components/common/TableWithPagination'
import apiKeyWarning2 from '../../../assets/icons/apiKey-warning2.svg'
import CreateApiKeyModal from './components/CreateApiKeyModal'
import apiKeyQuantity from '/images/apiKey-quantity.png'

interface ApiKeyItem {
  ID: number
  name: string
  api_key: string
  CreatedAt: string
  expired_at: string
}

// 本地存储键名
const STORAGE_KEY = 'ai-yygu-table-storage'

// 排序项接口
interface SortItem {
  field: string
  order: 'ascend' | 'descend'
}

// 表格参数接口
interface TableParams {
  currentPage: number
  pageSize: number
  sorts: SortItem[]
}

export default function ApiKeyManagement() {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')
  // 状态管理
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [isModalVisible, setIsModalVisible] = useState<boolean>(false)
  const [apiKeys, setApiKeys] = useState<ApiKeyItem[]>([])
  const [isAlertClosed, setIsAlertClosed] = useState<boolean>(false)

  // 从本地存储获取初始表格参数
  const getInitialTableParams = (): TableParams => {
    try {
      const localStorageData = localStorage.getItem(STORAGE_KEY)
      const data = localStorageData ? JSON.parse(localStorageData) : {}

      if (data.apiKeyTable) {
        return data.apiKeyTable
      }

      return {
        currentPage: 1,
        pageSize: 10,
        sorts: [],
      }
    } catch (error) {
      console.error('读取本地存储失败', error)
      return {
        currentPage: 1,
        pageSize: 10,
        sorts: [],
      }
    }
  }

  const initialParams = getInitialTableParams()

  const [currentPage, setCurrentPage] = useState<number>(
    initialParams.currentPage,
  )
  const [pageSize, setPageSize] = useState<number>(initialParams.pageSize)
  const [total, setTotal] = useState<number>(400) // 假设总数为40
  const [loading, setLoading] = useState<boolean>(false)
  const [sortItems, setSortItems] = useState<SortItem[]>(
    initialParams.sorts || [],
  )
  const pageSizeOptions = [10, 20, 50, 100]

  // 字段名映射表
  const fieldMapping: { [key: string]: string } = {
    name: 'name',
    api_key: 'api_key',
    CreatedAt: 'created_at',
    expired_at: 'expired_at',
  }

  // 保存表格参数到本地存储
  const saveTableParamsToStorage = (params: TableParams) => {
    try {
      const localStorageData = localStorage.getItem(STORAGE_KEY)
      const data = localStorageData ? JSON.parse(localStorageData) : {}

      // 更新apiKeyTable参数，保持其他数据不变
      const updatedData = {
        ...data,
        apiKeyTable: params,
      }

      localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedData))
    } catch (error) {
      console.error('保存表格参数到本地存储失败', error)
    }
  }

  // 构建排序参数数组
  const buildOrderByParams = (sorts: SortItem[]): string[] => {
    if (!sorts || sorts.length === 0) return []

    return sorts.map((item) => {
      const mappedField = fieldMapping[item.field] || item.field
      if (item.order === 'ascend') {
        return `${mappedField}`
      } else {
        return `${mappedField} desc`
      }
    })
  }

  // 获取API Key列表数据
  const fetchApiKeyList = async (
    page: number,
    size: number,
    customSortItems?: SortItem[],
  ) => {
    // 传递的参数应该是每页从第几条开始，和每页显示多少条
    // 参数名是Limit和Offset
    // 例如：第一页，从第0条开始，每页显示10条
    // 参数就是：Limit=10&Offset=0
    const start = (page - 1) * size
    // console.log(start, size)
    setLoading(true)

    // 使用自定义排序项或当前状态中的排序项
    const sortsToUse =
      customSortItems !== undefined ? customSortItems : sortItems

    // 构建排序参数
    // console.log('fetchApiKeyList中的sortItems:', sortsToUse)
    const orderBy = buildOrderByParams(sortsToUse)
    // console.log('构建的orderBy:', orderBy)
    // console.log({ limit: size, offset: start, orderBy: orderBy })

    // TODO: 这里应该调用获取API Key列表的接口
    const res = await getApiKeyList({
      limit: size,
      offset: start,
      orderBy: orderBy,
    })
    // console.log('res:', res)
    res.Data &&
      res.Data.forEach((item: ApiKeyItem) => {
        item.CreatedAt = dayjs(item.CreatedAt).format('YYYY-MM-DD')
        item.expired_at = dayjs(item.expired_at).format('YYYY-MM-DD')
      })
    setLoading(false)
    setApiKeys(res.Data || [])
    setTotal(res.total || 0)

    // 保存当前表格参数到本地存储
    saveTableParamsToStorage({
      currentPage: page,
      pageSize: size,
      sorts: sortsToUse,
    })
  }

  // 初始加载数据
  useEffect(() => {
    fetchApiKeyList(currentPage, pageSize)
  }, [])

  // 表格列定义
  const columns: ColumnsType<ApiKeyItem> = [
    {
      title: (
        <div className='font-[400] text-[#0A1A3A] text-base mr-1'>
          {t('profile.name')}
        </div>
      ),
      dataIndex: 'name',
      key: 'name',
      sorter: true,
      sortOrder: sortItems.find((item) => item.field === 'name')?.order,
      width: '22%',
    },
    {
      title: (
        <div className='font-[400] text-[#0A1A3A] text-base mr-1'>
          {t('profile.apiKey')}
        </div>
      ),
      dataIndex: 'api_key',
      key: 'api_key',
      sorter: true,
      sortOrder: sortItems.find((item) => item.field === 'api_key')?.order,
      width: '30%',
    },
    {
      title: (
        <div className='font-[400] text-[#0A1A3A] text-base mr-1'>
          {t('profile.createdTime')}
        </div>
      ),
      dataIndex: 'CreatedAt',
      key: 'CreatedAt',
      sorter: true,
      sortOrder: sortItems.find((item) => item.field === 'CreatedAt')?.order,
      width: '18%',
    },
    {
      title: (
        <div className='font-[400] text-[#0A1A3A] text-base mr-1'>
          {t('profile.expirationTime')}
        </div>
      ),
      dataIndex: 'expired_at',
      key: 'expired_at',
      sorter: true,
      sortOrder: sortItems.find((item) => item.field === 'expired_at')?.order,
      width: '18%',
    },
    {
      title: (
        <div className='font-[400] text-[#0A1A3A] text-base'>
          {t('profile.operation')}
        </div>
      ),
      key: 'action',
      width: '18%',
      render: (_, record) => (
        <span
          className='text-[#0A1A3A] hover:text-[#F53F3F] !text-base cursor-pointer'
          onClick={() => handleDelete(record.ID)}
        >
          {tC('button.delete')}
        </span>
      ),
    },
  ]

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (selectedKeys: React.Key[]) => {
      setSelectedRowKeys(selectedKeys)
    },
  }

  // 处理函数
  const handleCreateApiKey = () => {
    setIsModalVisible(true)
  }

  const handleModalCancel = () => {
    setIsModalVisible(false)
  }

  const handleCreateApiKeySubmit = async (values: {
    name: string
    expireDate: string
  }) => {
    // 将日期格式转换为后端需要的格式："2006-01-02T15:04:05Z07:00"
    const formattedExpireDate = dayjs(values.expireDate).format(
      'YYYY-MM-DDTHH:mm:ssZ',
    )

    try {
      // 调用创建API Key的接口，使用转换后的日期格式
      await createApiKey({ name: values.name, expired_at: formattedExpireDate })
      message.success(tM('createSuccess', { target: t('profile.apiKey') }))
      setIsModalVisible(false)

      // 刷新数据列表
      fetchApiKeyList(currentPage, pageSize)
    } catch (error) {
      console.error('创建API Key失败:', error)
      message.error(tM('createFailure', { center: t('profile.apiKey') }))
    }
  }

  const handleDelete = (id: number) => {
    Modal.confirm({
      title: null,
      icon: null,
      centered: true,
      className: 'delete-api-key-modal !w-[30%]',
      content: (
        <div className='relative'>
          <div className='flex justify-between'>
            <div className='text-[22px] font-[500] mb-2 text-[#000000E5]'>
              {tC('button.confirmDelete')}
            </div>
            <img src={apiKeyWarning2} alt='warning' className='w-6 h-6' />
          </div>
          <div className='h-[0.5px] w-[calc(100%+48px)] bg-[#C9CDD4] mt-4 -mx-6'></div>
          <div className='mt-6 text-base text-[#616373] mb-22 font-medium fontFamily-pingFangSC'>
            {t('profile.deleteApiKeyCautionAuthFail')}
          </div>
        </div>
      ),
      okText: tC('button.confirmDelete'),
      okButtonProps: {
        className:
          'bg-[#165DFF] hover:bg-[#165DFF] !w-26 !h-[42px] !rounded-lg !text-base !px-4 !py-1',
        danger: false,
      },
      cancelButtonProps: {
        className:
          '!bg-[#F4F9FF] text-[#616373] !w-22 !h-[42px] !rounded !text-base !border-none !px-4 !py-1 !font-medium',
      },
      cancelText: tC('button.cancel'),
      onOk: async () => {
        // TODO: 这里应该调用删除API Key的接口
        await deleteApiKey({ id: Number(id) })
        setApiKeys(apiKeys.filter((item) => item.ID !== id))
        message.success(tM('deleteSuccess'))

        // 刷新数据列表
        fetchApiKeyList(currentPage, pageSize)
      },
      maskClosable: true,
    })
  }

  const handlePageChange = (page: number, size?: number) => {
    const newPageSize = size || pageSize
    setCurrentPage(page)

    if (size && size !== pageSize) {
      setPageSize(size)
    }

    // 调用获取API Key列表的接口
    fetchApiKeyList(page, newPageSize)
  }

  const handleTableChange = (sorter: any, filters?: any, extra?: any) => {
    // 处理表格排序
    let newSortItems: SortItem[] = []

    if (sorter) {
      // 处理单列排序
      if (
        sorter.field ||
        sorter.columnKey ||
        (sorter.column && sorter.column.dataIndex)
      ) {
        const field =
          sorter.field ||
          sorter.columnKey ||
          (sorter.column && sorter.column.dataIndex)

        if (sorter.order) {
          newSortItems = [
            {
              field,
              order: sorter.order as 'ascend' | 'descend',
            },
          ]
        }
      }
    }

    // 更新状态
    setSortItems(newSortItems)

    // 构建排序参数
    const orderBy = buildOrderByParams(newSortItems)

    // 使用新的排序参数获取数据
    fetchApiKeyList(currentPage, pageSize, newSortItems)
  }

  // 背景图
  const bg = {
    url: '../../../assets/icons/myInfomation-bg.svg',
  }

  // 处理Alert关闭事件
  const handleAlertClose = () => {
    setIsAlertClosed(true)
  }

  // 在组件卸载时清除存储信息
  useEffect(() => {
    return () => {
      // 组件卸载时，清除与当前API Key管理相关的所有本地存储信息
      try {
        const localStorageData = localStorage.getItem(STORAGE_KEY)
        if (localStorageData) {
          const data = JSON.parse(localStorageData)
          const storageKey = 'apiKeyTable'
          const updatedData = { ...data }

          // 删除当前页面的存储数据
          if (updatedData[storageKey]) {
            delete updatedData[storageKey]
          }

          localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedData))
        }
      } catch (error) {
        console.error('清除本地存储失败', error)
      }
    }
  }, [])

  return (
    <div
      className='w-[calc(100vw-200px)] h-[calc(100vh-64px)] p-4 sm:p-4 overflow-hidden'
      style={{
        backgroundImage: `url(${bg.url})`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}
    >
      {/* 提示信息 */}
      <Alert
        message={t('profile.apiKeyCredentialSecureRotate')}
        className='mb-6'
        onClose={handleAlertClose}
      />

      {/* 内容区域 */}
      <div className='bg-white h-auto rounded-lg px-6 py-8 flex flex-col'>
        {/* 操作按钮 */}
        <div className='flex justify-between mb-8'>
          <Button
            onClick={handleCreateApiKey}
            className='w-28 !h-10 !bg-[#165DFF] !text-white flex items-center justify-center !font-500 !text-base !rounded'
          >
            {tC('button.create', { target: t('profile.apiKey') })}
          </Button>

          <Tooltip title={t('profile.accountBalance')}>
            <Button className='flex items-center text-gray-700 font-normal border-gray-300 !rounded !text-base'>
              <span>
                <img
                  src={apiKeyQuantity}
                  alt='apiKeyQuantity'
                  className='mr-1 w-5 h-5'
                />
              </span>
              {t('profile.accountBalance')}
            </Button>
          </Tooltip>
        </div>

        {/* 表格与分页 */}
        <div className='flex-1'>
          <TableWithPagination<ApiKeyItem>
            columns={columns}
            dataSource={apiKeys}
            rowKey='id'
            current={currentPage}
            pageSize={pageSize}
            total={total}
            pageSizeOptions={pageSizeOptions}
            onPageChange={handlePageChange}
            onTableChange={handleTableChange}
            loading={loading}
            className='bg-white api-key-table'
            {...(isAlertClosed
              ? {
                tableHeight: {
                  default: 'h-[calc(100vh-416px)]',
                  sm: 'sm:h-[calc(100vh-284px)]',
                  lg: 'lg:h-[calc(100vh-284px)]',
                  '2xl': '2xl:h-[calc(100vh-286px)]',
                },
              }
              : {})}
          />
        </div>
      </div>

      {/* 创建API Key的弹窗 */}
      <CreateApiKeyModal
        visible={isModalVisible}
        onCancel={handleModalCancel}
        onOk={handleCreateApiKeySubmit}
      />
    </div>
  )
}
