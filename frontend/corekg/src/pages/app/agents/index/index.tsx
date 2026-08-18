import { FC, PropsWithChildren } from 'react'
import { Button, Input, Select, Skeleton } from 'antd'
import { useNavigate } from 'react-router-dom'
import { BasicAgentInfo } from 'Agent'
import dayjs from 'dayjs'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import { scroll } from '@/styles/scroll'
import { withAgentStyle } from '../AgentStyleProvider'
import { AddAgentBtn } from './AddAgentBtn'
import { AgentProvider, useAgentList } from './AgentContext'
import { AgentItem } from './AgentItem'
import AddIcon from './images/add.svg?react'
import bgImg from './images/bg.png'
import ChartIcon from './images/chart.svg?react'
import EmptyIcon from './images/empty.svg?react'
import SearchIcon from './images/search.svg?react'
import { DatabaseOutlined } from '@ant-design/icons'
import styles from './styles.module.scss'

const Agents: FC = () => {
  return (
    <AgentProvider>
      <InnerAgent />
    </AgentProvider>
  )
}

const InnerAgent: FC = withAgentStyle((props) => {
  const agentList = useAgentList()
  const navigate = useNavigate()
  const [sortKey, setSortKey] = useState<'CreatedAt' | 'UpdatedAt'>('CreatedAt')
  const [filterKey, setFilterKey] = useState<string>('all')
  const [inputValue, setInputValue] = useState<string>('')
  const [searchValue, setSearchValue] = useState<string>('')
  const dropdownRef = useRef<HTMLDivElement>(null)
  const filteredAgents = useMemo(() => {
    if (!agentList.data) return undefined
    return agentList.data
      .filter((item) => {
        return match({ filterKey, type: item.type })
          .with({ filterKey: 'all' }, () => true)
          .with(
            { filterKey: 'role', type: P.union('knowledge', 'role_play') },
            () => true,
          )
          .with({ filterKey: 'prompt', type: 'prompt' }, () => true)
          .with({ filterKey: 'workflow', type: 'workflow' }, () => true)
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
  }, [agentList.data, filterKey, searchValue, sortKey])
  return (
    <div className={cn('w-full h-full flex flex-col', props.extraClassName)}>
      {/* 顶部导航部分 - 面包屑 */}
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px]'>
        <div className='flex items-center gap-2 text-sm'>
          <ChartIcon />
          <span className='text-[#2D2D2D] cursor-default font-medium'>
            智能体
          </span>
        </div>
      </div>

      {/* 顶部欢迎区域 */}
      <div className={cn('rounded-2xl ', 'mx-12 mt-[10px] relative')}>
        <img src={bgImg} className='w-full' />
        <div className='text-[#2A4C95] z-10 absolute ml-12 top-1/2 -translate-y-1/2'>
          <h1 className='text-[40px] font-semibold '>欢迎使用智能体</h1>
          <p className='text-base font-medium mt-2.5'>
            可配置的任务型助手，理解上下文并持续学习
          </p>
        </div>
      </div>
      <div
        className={cn('flex-1 overflow-auto px-25 pb-12 pt-8 bg-white', scroll)}
      >
        {agentList.loading ? (
          <Skeleton active paragraph={{ rows: 10 }} />
        ) : (
          <AgentItems value={filteredAgents}>
            {/* 卡片列表上方按钮 为对齐置于此处 */}
            <div
              className='flex items-center whitespace-nowrap'
              style={{ gridColumn: '1/-1' }}
            >
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
                      { value: 'UpdatedAt', label: '按最近更新' },
                    ]}
                  />
                </div>
                <div className='flex gap-[6px] items-center'>
                  <div className='font-[500] text-[14px] text-[#919497]'>
                    智能体类型
                  </div>
                  <Select
                    defaultValue={filterKey}
                    style={{ width: 164 }}
                    classNames={{
                      popup: {
                        root: styles.filterSelect,
                      },
                    }}
                    onChange={setFilterKey}
                    popupMatchSelectWidth={false}
                    options={[
                      { value: 'all', label: '全部' },
                      { value: 'role', label: '指令型-简单应用' },
                      { value: 'prompt', label: '指令型-高级编排' },
                      { value: 'workflow', label: '工作流' },
                    ]}
                  />
                </div>
              </div>
              <div className='ml-auto flex justify-end'>
                <div
                  className='relative flex items-center gap-[12px]'
                  ref={dropdownRef}
                >
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
                  {/* 资源库 */}
                  <Button
                    icon={<DatabaseOutlined />}
                    className={cn(
                      'flex items-center gap-1 h-[30px] px-3',
                      'text-sm font-medium rounded-[6px]',
                      'border border-[#0C99FF] text-[#0C99FF]',
                      'hover:border-[#0C99FF] hover:text-[#40a9ff]',
                      'active:text-[#096dd9]',
                      'shadow-none bg-transparent'
                    )}
                    onClick={() => navigate('/agents/resources')}
                  >
                    资源库
                  </Button>
                  <AddAgentBtn
                    className={cn(
                      'text-sm font-medium rounded-[6px] border border-[#0C99FF] hover:border-[#0C99FF] text-[#0C99FF] active:text-[#0C99FF] shadow-none',
                    )}
                  />
                </div>
              </div>
            </div>
          </AgentItems>
        )}
      </div>
    </div>
  )
})
export default Agents

const AgentItems: FC<PropsWithChildren & { value?: BasicAgentInfo[] }> = (
  props,
) => {
  const { value } = props

  return (
    <div className='w-full flex justify-center'>
      <div
        className='w-full grid gap-x-10 gap-y-8 justify-center'
        style={{
          gridTemplateColumns: 'repeat(auto-fill, 300px)',
        }}
      >
        {props.children}
        {!value || value.length === 0 ? (
          <div
            className='h-60 flex flex-col items-center justify-center text-[#919497]'
            style={{ gridColumn: '1/-1' }}
          >
            <EmptyIcon />
            暂无智能体数据，请点击"新建智能体"添加
          </div>
        ) : null}
        {value?.map((item) => {
          return <AgentItem key={item.id} {...item} />
        })}
      </div>
    </div>
  )
}
