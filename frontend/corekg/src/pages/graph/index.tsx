import { FC, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button, Input, Select, Skeleton } from 'antd'
import { GraphBaseInfo } from 'Graph'
import { useBoolean, useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import { listForestGraph } from '@/api/graph'
import AddIcon from '@/pages/app/docs/images/add.svg?react'
import { scroll } from '@/styles/scroll'
import { EmployeeProvider } from './components/EmployeeProvider'
import { Graphs } from './components/Graphs'
import bgImg from './images/bg.png'
import GraphIcon from './images/graph.svg'
import SearchIcon from './images/search.svg?react'
import styles from './styles.module.scss'

/** 知识图谱 */
const Graph: FC = () => {
  const { data: _data, loading, run } = useGraphList()
  const [sortKey, setSortKey] = useState<'CreatedAt' | 'updateAt'>('CreatedAt')
  const [statusKey, setStatusKey] = useState<string>('all')
  const [inputValue, setInputValue] = useState<string>('')
  const [searchValue, setSearchValue] = useState<string>('')

  const data = useMemo(() => {
    if (!_data) return undefined
    return _data
      .filter((item) => {
        // 如果状态是草稿且不是管理员，则过滤掉
        if (item.status === 'draft' && !item.is_admin) {
          return false
        }
        const { status, task_count = 0, success_task_count = 0 } = item
        // 计算构建进度
        const buildProgress =
          task_count === 0
            ? 0
            : Math.floor((success_task_count / task_count) * 100)
        // 如果 status 是 updatable 且进度是 100%，则视为 success 状态
        const actualStatus =
          (status as string) === 'updatable' && buildProgress === 100
            ? ('success' as const)
            : status
        return match({ statusKey, status: actualStatus })
          .with({ statusKey: 'all' }, () => true)
          .with({ statusKey: 'draft', status: 'draft' }, () => true)
          .with({ statusKey: 'success', status: 'success' }, () => true)
          .with(
            { statusKey: 'running', status: P.union('pending', 'running') },
            () => true,
          )
          .otherwise(() => false)
      })
      .filter((item) => {
        return JSON.stringify(item).includes(searchValue)
      })
      .sort((v1, v2) => {
        const date1 = dayjs(v1[sortKey])
        const date2 = dayjs(v2[sortKey])
        return date1.isAfter(date2) ? -1 : 1
      })
  }, [_data, searchValue, sortKey, statusKey])

  return (
    <EmployeeProvider>
      <div className='w-full h-full flex flex-col'>
        {/* 顶部导航部分 - 面包屑 */}
        <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px]'>
          <div className='flex items-center gap-2 text-sm'>
            <img src={GraphIcon} className='w-4 h-4' alt='' />
            <span className='text-[#2D2D2D] cursor-default font-medium'>
              知识图谱
            </span>
          </div>
        </div>

        <div className={cn('rounded-2xl ', 'mx-12 mt-[10px] relative')}>
          <img src={bgImg} className='w-full' />
          <div className='text-[#2A4C95] z-10 absolute ml-12 top-1/2 -translate-y-1/2'>
            <h1 className='text-[40px] font-semibold '>欢迎使用知识图谱</h1>
            <p className='text-base font-medium mt-2.5'>
              全局管理您的知识实体与关系，构建可视化网络
            </p>
          </div>
        </div>

        {/* 知识图谱卡片列表区域 - 可滚动 */}
        <div className={cn('flex-1 overflow-auto bg-[#ffffff]', scroll)}>
          {loading ? (
            <Skeleton
              active
              paragraph={{ rows: 10 }}
              className='px-12 pb-12 pt-2 '
            />
          ) : (
            <div className='w-full flex justify-center'>
              <div
                className='w-full px-25 pb-12 pt-8 grid gap-x-10 gap-y-8 justify-center'
                style={{
                  gridTemplateColumns: 'repeat(auto-fill, 300px)',
                }}
              >
                <Graphs data={data ?? []} reload={run} className=''>
                  <div
                    className='flex items-center whitespace-nowrap'
                    style={{ gridColumn: '1/-1' }}
                  >
                    {/* 卡片列表上方按钮 为对齐置于此处 */}
                    <div className='mr-40 flex gap-[16px] items-center'>
                      <div className='flex gap-[6px] items-center'>
                        <div className='font-[500] text-[14px] text-[#919497]'>
                          排序方式
                        </div>
                        <Select
                          defaultValue={sortKey}
                          style={{ width: 114 }}
                          popupMatchSelectWidth={false}
                          classNames={{
                            popup: {
                              root: styles.filterSelect,
                            },
                          }}
                          onChange={setSortKey}
                          options={[
                            { value: 'CreatedAt', label: '按最新创建' },
                            { value: 'updateAt', label: '按最近更新' },
                          ]}
                        />
                      </div>
                      <div className='flex gap-[6px] items-center'>
                        <div className='font-[500] text-[14px] text-[#919497]'>
                          图谱状态
                        </div>
                        <Select
                          defaultValue={statusKey}
                          style={{ width: 164 }}
                          classNames={{
                            popup: {
                              root: styles.filterSelect,
                            },
                          }}
                          onChange={setStatusKey}
                          popupMatchSelectWidth={false}
                          options={[
                            { value: 'all', label: '全部' },
                            { value: 'draft', label: '草稿' },
                            { value: 'success', label: '构建成功' },
                            { value: 'running', label: '构建中' },
                          ]}
                        />
                      </div>
                    </div>
                    <div className='ml-auto flex justify-end'>
                      <div className='relative flex items-center gap-[12px]'>
                        <Input
                          value={inputValue}
                          placeholder={'搜索'}
                          prefix={<SearchIcon />}
                          onChange={(e) => setInputValue(e.target.value)}
                          onBlur={() =>
                            !inputValue?.trim?.() && setSearchValue(inputValue)
                          }
                          onPressEnter={() => setSearchValue(inputValue)}
                          className={`w-[70px] h-[30px] border-[#0C99FF] shadow-none  ${styles.searchInputWrap} ${inputValue?.trim?.() ? styles.searchInputWrapSearching : ''}`}
                        />
                        <Link to={`/graph/edit`}>
                          <Button
                            className={cn(
                              styles.createBtn,
                              'h-[30px] text-sm font-medium rounded-[6px] border-[#0C99FF] hover:border-[#0C99FF] text-[#0C99FF] active:text-[#0C99FF] shadow-none',
                            )}
                          >
                            <AddIcon /> 新建知识图谱
                          </Button>
                        </Link>
                      </div>
                    </div>
                  </div>
                </Graphs>
              </div>
            </div>
          )}
        </div>
      </div>
    </EmployeeProvider>
  )
}
export default Graph

const useGraphList = () => {
  return useRequest(async () => {
    const res = await listForestGraph({})
    const data: any[] = res.Data ?? []
    const _data: GraphBaseInfo[] = data.map((item) => {
      const {
        ID,
        name,
        description,
        node_count,
        edge_count,
        updated_at: UpdatedAt,
        created_at: CreatedAt,
      } = item
      return {
        ...item,
        UpdatedAt,
        CreatedAt,
        id: ID,
        name,
        desc: description,
        totalNodes: node_count,
        totalRelationships: edge_count,
        updateAt: dayjs(UpdatedAt).format(`YYYY-MM-DD`),
      }
    })

    return _data
  })
}
