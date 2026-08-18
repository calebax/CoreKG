import { FC, useState, useEffect, useMemo, useCallback } from 'react'
import { Button, Switch, Table, TableProps, Tooltip, message } from 'antd'
import { usePagination, useRequest } from 'ahooks'
import { useNavigate } from 'react-router-dom'
import { createSession } from '@/api/agent'
import { listForestDB, updateResourceEnable } from '@/api/knowledge'
import type { CommonArgs } from '@/api/knowledge'
import TableWithPagination from '@/components/common/TableWithPagination'
import HelpIcon from '../../../images/help-tip.svg?react'
import type { CommonArgs } from '@/api/knowledge'

/** mysql数据库信息 */
export const DBs: FC<{ forest_id: number; search: string }> = (props) => {
  const { forest_id, search } = props
  type DBItem = {
    data_rows: number
    data_size: number
    db_name: string
    forest_db_id: number
    enable?: boolean
  }
  const navigate = useNavigate()

  const [current, setCurrent] = useState<number>(1)
  const [pageSize, setPageSize] = useState<number>(10)
  const [loading, setLoading] = useState<boolean>(false)
  const [total, setTotal] = useState<number>(0)
  const [dbList, setDbList] = useState<DBItem[]>([])
  // 启用状态切换loading状态（记录正在处理的数据库ID）
  const [enablingDbId, setEnablingDbId] = useState<number | null>(null)

  const fetchDbList = async () => {
    setLoading(true)
    const filters: CommonArgs['filters'] = []
    if (search) {
      filters.push({ field: 'db_name', value: [search] })
    }
    filters.push()
    const { db_list, total } = await listForestDB({
      forest_id,
      limit: pageSize,
      offset: (current - 1) * pageSize,
      filters,
    })
    setTotal(
      total === 0 && Array.isArray(db_list) && db_list?.length
        ? db_list.length
        : total,
    )
    setDbList(db_list)
    console.log(db_list)

    setLoading(false)
  }

  useEffect(() => {
    fetchDbList()
  }, [forest_id, search, current, pageSize])

  const handleEnableChange = useCallback(
    async (enable: boolean, file: DBItem) => {
      // 如果正在处理该数据库，直接返回
      if (enablingDbId === file.forest_db_id) {
        return
      }

      // 设置loading状态，禁用开关
      setEnablingDbId(file.forest_db_id)

      try {
        await updateResourceEnable({
          enable: enable ? 1 : -1,
          forest_id: Number(forest_id),
          resource_ids: [String(file.forest_db_id)],
          resource_type: 'mysql_db',
        })

        // 接口成功后再更新UI状态
        setDbList((db_list) => {
          return db_list.map((item) =>
            item.forest_db_id === file.forest_db_id ? { ...item, enable } : item,
          )
        })
      } catch (error) {
        console.log(error)
      } finally {
        // 清除loading状态
        setEnablingDbId(null)
      }
    },
    [forest_id, enablingDbId],
  )

  const columns: TableProps<DBItem>['columns'] = useMemo(() => {
    const col: TableProps<DBItem>['columns'] = [
      {
        title: '名称',
        render: (_, record) => {
          return <div>{record.db_name}</div>
          return <NavLink to={`${record.db_name}`}>{}</NavLink>
        },
        width: '40%',
        // sorter: (v1, v2 ) => v1.db_name.localeCompare(v2.db_name),
      },
      {
        title: '大小',
        render: (_, record) => `${record.data_size}MB`,
        // sorter: (v1, v2) => v1.data_size - v2.data_size,
        width: '20%',
      },
      {
        title: '行数',
        dataIndex: 'data_rows',
        width: '20%',

        // sorter: true,
      },
      {
        title: (
          <div className='flex items-center gap-[4px] font-medium text-[#919497] text-sm leading-[22px]'>
            启用状态
            <Tooltip title='启用后，系统将在问答时检索并引用该资源。'>
              <HelpIcon />
            </Tooltip>
          </div>
        ),
        dataIndex: 'enable',
        key: 'enable',
        width: '15%',
        render: (enable: boolean, record: any) => {
          // 如果该数据库正在处理启用状态，禁用开关
          const isProcessing = enablingDbId === record.forest_db_id
          return (
            <div onClick={(e) => e.stopPropagation()}>
              <Switch
                value={enable}
                size='small'
                onChange={(v) => handleEnableChange(v, record)}
                disabled={isProcessing}
              />
            </div>
          )
        },
      },
      {
        title: '操作',
        render: (_, record) => {
          const { db_name } = record
          return (
            <span className='flex gap-4'>
              {/* <QABtn name={db_name} forest_id={forest_id} /> */}
              <Button
                className='!border-0 !shadow-none !bg-[transparent] font-[500] hover:text-[#0C99FF]'
                onClick={(e) => {
                  e.stopPropagation()
                  navigate(db_name)
                }}
                size='small'
              >
                查看
              </Button>
            </span>
          )
        },
        width: '20%',
      },
    ]
    return col
  }, [forest_id, enablingDbId, handleEnableChange])

  const onPageChange = (page: number, size?: number) => {
    setCurrent(page)
    if (size) setPageSize(size)
  }

  return (
    <div className='flex-1 overflow-hidden'>
      <TableWithPagination
        loading={loading}
        columns={columns}
        dataSource={dbList || []}
        rowKey='db_name'
        total={total || 0}
        current={current}
        pageSize={pageSize}
        onPageChange={onPageChange}
        scroll={{ x: 1000 }}
        tableHeight={{
          default: 'h-full',
          sm: 'sm:h-full',
          lg: 'lg:h-full',
          '2xl': '2xl:h-full',
        }}
      />
    </div>
  )
}

const QABtn: FC<{ name: string; forest_id: number }> = (props) => {
  const { name, forest_id } = props
  const { run, loading } = useRequest(
    async () => {
      const { ID: session_id } = (await createSession({
        base_type: 'mysql',
        names: [name],
        resource_type: 'db_list',
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
    <Button type='link' size='small' loading={loading} onClick={run}>
      问答
    </Button>
  )
}
