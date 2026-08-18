import { FC } from 'react'
import { Table, TableProps } from 'antd'
import { useRequest } from 'ahooks'
import { getForestTableHeader } from '@/api/knowledge'
import TableWithPagination from '@/components/common/TableWithPagination'

/** 表头信息 */
export const Headers: FC<{
  forest_id: number
  forest_db_name: string
  forest_table_name: string
  search: string
}> = (props) => {
  const { forest_table_name, forest_db_name, forest_id, search } = props
  type HeaderItem = {
    column_comment: string
    column_name: string
    column_type: string
  }
  const { data, loading } = useRequest(
    async () => {
      const { column_list } = await getForestTableHeader({
        forest_id,
        forest_db_name,
        forest_table_name,
      })

      return column_list as HeaderItem[]
    },
    {
      refreshDeps: [forest_id, forest_table_name, search],
    },
  )
  const columns: TableProps<HeaderItem>['columns'] = useMemo(() => {
    const col: TableProps<HeaderItem>['columns'] = [
      {
        title: '列名',
        dataIndex: 'column_name',
      },
      {
        title: '类型',
        dataIndex: 'column_type',
      },
      {
        title: '描述',
        dataIndex: 'column_comment',
      },
    ]
    return col
  }, [])

  return (
    <TableWithPagination
      loading={loading}
      columns={columns}
      dataSource={data || []}
      rowKey='db_name'
      total={0}
      pageSize={100}
      scroll={{ x: 1000 }}
      tableHeight={{
        default: 'h-full',
        sm: 'sm:h-full',
        lg: 'lg:h-full',
        '2xl': '2xl:h-full',
      }}
      current={1}
      onPageChange={() => {}}
    />
  )
}
