import { FC } from 'react'
import { Breadcrumb, Input, Tag, Tooltip, Typography } from 'antd'
import { ColumnProps } from 'antd/es/table'
import { Table } from 'antd/lib'
import { usePagination } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { getPaymentRecord } from '@/api/pay'
import Sep from '../AccountBindings/images/separator.svg?react'
import accountBindingsStyles from '../AccountBindings/index.module.scss'

const OrderManagement: FC = () => {
  const nav = useNavigate()
  const { t } = useTranslation('pages')
  const [search, setSearch] = useState<string>()
  const { loading, pagination, data, run } = usePagination(
    async (args) => {
      const { pageSize, current } = args
      const filters: CommonArgs['filters'] = [
        {
          field: 'status',
          value: ['success'],
        },
      ]
      if (search) {
        filters.push({ field: 'order_sn', value: [search] })
      }
      const res = await getPaymentRecord({
        limit: pageSize,
        offset: pageSize * (current - 1),
        orderBy: ['created_at desc'],
        filters: filters.length === 0 ? undefined : filters,
      })
      const { data: list, total } = res
      return { list, total }
    },
    {
      refreshDeps: [search],
    },
  )
  const columns = useMemo(() => {
    const _col: ColumnProps<NonNullable<typeof data>['list'][number]>[] = [
      {
        title: '订单编号',
        dataIndex: 'order_sn',
      },
      {
        title: '订单金额(元)',
        render: (_, record) => `¥${(record.amount / 100).toFixed(2)}`,
      },
      {
        title: '订单状态',
        render: (_, record) => {
          // 只查成功的
          return (
            <div className='py-0.5 px-2.5 inline rounded-full bg-[#D7F5E5] text-[#13A374]'>
              已完成
            </div>
          )
        },
      },
      {
        title: '创建时间',
        render: (_, record) => {
          const { created_at } = record
          return dayjs(created_at).format('YYYY-MM-DD HH:mm:ss')
        },
      },
      {
        title: '支付时间',
        render: (_, record) => {
          const { paid_at } = record
          return dayjs(paid_at).format('YYYY-MM-DD HH:mm:ss')
        },
      },
    ]
    return _col
  }, [])

  return (
    <div className='w-full h-full flex flex-col'>
      <div className='h-[50px] bg-[#fff] border-b border-b-[#EFF1F4] px-[16px] flex items-center'>
        <Breadcrumb
          className={accountBindingsStyles.accountBindingsHeader}
          separator={<Sep />}
          items={[
            {
              title: (
                <span
                  className='cursor-pointer'
                  onClick={() => {
                    nav(-1)
                  }}
                >
                  {t('app.sidebar.project')}
                </span>
              ),
            },
            {
              title: <span className='cursor-pointer'>订单管理</span>,
            },
          ]}
        />
      </div>
      <div className='my-4 px-4 flex items-center'>
        <Input.Search
          className='ml-auto w-54'
          placeholder='搜索订单号'
          allowClear
          onSearch={(v) => {
            setSearch(v)
          }}
          onClear={() => setSearch('')}
        />
      </div>
      <Table
        className='px-4 flex-1 overflow-auto'
        loading={loading}
        columns={columns}
        dataSource={data?.list}
        pagination={pagination}
      />
    </div>
  )
}

export default OrderManagement
