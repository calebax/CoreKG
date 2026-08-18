import { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, DatePicker, Input, Pagination, Table, TableProps } from 'antd'
import { RangePickerProps } from 'antd/es/date-picker'
import { SearchOutlined } from '@ant-design/icons'
import { usePagination } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { getEmployee } from '@/api/perm'
import BackIcon from '@/assets/icons/backIcon.svg'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { CreateBtn } from './components/CreateBtn'
import { InviteBtn } from './components/InviteBtn'
import { Operator } from './components/Operator'
import styles from './styles.module.scss'

const UserManagement: FC = () => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const navigate = useNavigate()
  const [inputValue, setInputValue] = useState('')
  const [search, setSearch] = useState('')
  const [date, setDate] = useState<RangePickerProps['value']>()
  const { data, loading, pagination, run } = usePagination(
    async (args) => {
      const { pageSize, current } = args
      const body: any = {
        limit: pageSize,
        offset: (current - 1) * pageSize,
        filters: [],
      }
      if (search) {
        body.filters.push({
          exactMatch: true,
          field: 'auto',
          value: [search],
        })
      }
      if (date) {
        const [start, end] = date
        body.beginTime = start!.format('YYYY-MM-DDTHH:mm:ssZ')
        body.endTime = end!.format('YYYY-MM-DDTHH:mm:ssZ')
      }
      const { total, Data: list } = (await getEmployee(body)) as unknown as {
        total: number
        Data: any[]
      }
      return { total, list }
    },
    { defaultPageSize: 20, refreshDeps: [search, date] },
  )
  const refresh = () => {
    run({ current: 1, pageSize: pagination.pageSize })
  }
  const columns: TableProps['columns'] = [
    {
      title: t('settings.name', { target: t('settings.member') }),
      dataIndex: 'user_name',
    },
    {
      title: t('settings.role', { target: t('settings.member') }),
      render: (record) => {
        return record.sys_role === 'sys_admin'
          ? t('settings.admin')
          : t('settings.ordinaryMember')
      },
    },
    {
      title: t('settings.addTime'),
      render: (record) => {
        return dayjs(record.created_at).format('YYYY-MM-DD')
      },
    },
    {
      title: t('settings.operationManagement'),
      width: 100,
      render: (record) => {
        const { id, uin } = record
        return (
          <Operator uin={uin} id={id} reload={refresh} employeeInfo={record} />
        )
      },
    },
  ]
  return (
    <div className='h-full flex flex-col pt-4'>
      <span className='flex items-center mb-4'>
        <div
          className='w-6 h-6 cursor-pointer flex items-center justify-center ml-6 mr-2'
          onClick={() => navigate('/')}
        >
          <img src={BackIcon} alt={t('settings.goBack')} className='w-6 h-6' />
        </div>
        {version !== 'saas' ? <CreateBtn refresh={refresh} /> : <InviteBtn />}
        <Input
          className='ml-auto mr-6 w-60'
          placeholder={tC('button.search')}
          allowClear
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onPressEnter={() => setSearch(inputValue)}
          prefix={<SearchOutlined onClick={() => setSearch(inputValue)} />}
        ></Input>
        <DatePicker.RangePicker
          value={date}
          onChange={setDate}
          className='w-64'
        />
        <Button
          onClick={() => {
            setSearch('')
            setInputValue('')
            setDate(null)
          }}
          className='mx-6'
        >
          {tC('button.reset')}
        </Button>
      </span>
      <Table
        rowKey={'uin'}
        className={cn('flex-1 overflow-auto')}
        dataSource={data?.list}
        columns={columns}
        loading={loading}
        pagination={false}
      ></Table>
      <Pagination
        {...pagination}
        showSizeChanger
        className={cn('self-end mt-14 mr-6 py-6', styles.pagination)}
      />
    </div>
  )
}

export default UserManagement
