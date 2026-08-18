import { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Table, TableProps } from 'antd'
import { usePagination, useRequest } from 'ahooks'
import { createSession } from '@/api/agent'
import { listForestTable } from '@/api/knowledge'
import TableWithPagination from '@/components/common/TableWithPagination'
import TableIcon from './table.png'

/** mysql表信息 */
export const Tables: FC<{
  forest_id: number
  forest_db_name: string
  search: string
}> = (props) => {
  const { forest_id, forest_db_name, search } = props
  const navigate = useNavigate()
  type TableItem = {
    data_rows: number
    data_size: number
    forest_table_name: string
  }
  const { data, pagination, loading } = usePagination(
    async (page) => {
      const { current, pageSize } = page
      const filters: CommonArgs['filters'] = []
      if (search) {
        filters.push({ field: 'forest_table_name', value: [search] })
      }
      const { table_list, total } = await listForestTable({
        forest_id,
        forest_db_name,
        limit: pageSize,
        offset: (current - 1) * pageSize,
        filters,
      })

      return { total, list: table_list as TableItem[] }
    },
    {
      refreshDeps: [forest_db_name, search],
    },
  )
  const columns: TableProps<TableItem>['columns'] = useMemo(() => {
    const col: TableProps<TableItem>['columns'] = [
      {
        title: '名称',
        render: (_, record) => {
          return (
            <div className={'flex gap-2'}>
              <img src={TableIcon} />
              {record.forest_table_name}
            </div>
          )
        },
        // sorter: (v1, v2) =>
        //   v1.forest_table_name.localeCompare(v2.forest_table_name),
      },
      {
        title: '大小',
        render: (_, record) => `${record.data_size}MB`,
        // sorter: (v1, v2) => v1.data_size - v2.data_size,
      },
      {
        title: '行数',
        dataIndex: 'data_rows',
        // sorter: true,
      },
      {
        title: '操作',
        render: (_, record) => {
          const { forest_table_name } = record
          return (
            <span className='flex gap-4'>
              {/* <QABtn name={db_name} forest_id={forest_id} /> */}
              <Button
                className='!border-0 !bg-[transparent] font-[500] hover:text-[#CC5DE8]'
                onClick={(e) => {
                  e.stopPropagation()
                  navigate(forest_table_name)
                }}
                size='small'
              >
                查看
              </Button>
            </span>
          )
        },
      },
    ]
    return col
  }, [forest_id])

  const onPageChange = (page: number, size?: number) => {
    pagination.changeCurrent(page)
    if (size) pagination.changePageSize(size)
  }

  return (
    <TableWithPagination
      loading={loading}
      columns={columns}
      dataSource={data?.list || []}
      rowKey='db_name'
      total={data?.total || 0}
      current={pagination.current}
      pageSize={pagination.pageSize}
      onPageChange={onPageChange}
      scroll={{ x: 1000 }}
      tableHeight={{
        default: 'h-full',
        sm: 'sm:h-full',
        lg: 'lg:h-full',
        '2xl': '2xl:h-full',
      }}
    />
  )
}

const QABtn: FC<{ name: string; forest_id: number }> = (props) => {
  const { name, forest_id } = props
  const { refresh, loading } = useRequest(
    async () => {
      const { ID: session_id } = (await createSession({
        base_type: 'mysql',
        names: [name],
        resource_type: 'db_table_list',
        resource_id: forest_id,
        model_id: 1,
      })) as any
      const searchParams = new URLSearchParams()
      searchParams.append('session_id', session_id)
      window.open(`${import.meta.env.BASE_URL}QA?${searchParams.toString()}`)
    },
    { manual: true },
  )

  return (
    <Button type='link' size='small' loading={loading} onClick={refresh}>
      问答
    </Button>
  )
}
