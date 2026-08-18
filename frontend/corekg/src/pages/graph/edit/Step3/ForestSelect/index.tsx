import { FC } from 'react'
import { Button, Input, Select } from 'antd'
import dayjs from 'dayjs'
import { SearchIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { Forest } from '../../KnowledgeContext'
import { KnowledgeItem } from './KnowledgeItem'
import styles from './styles.module.scss'

export const ForestSelect: FC<
  Style &
    ValueController<number> & {
      disabled?: boolean
      data: Forest[]
    }
> = (props) => {
  const { data, disabled, value, onChange, className, style } = props
  const [sortKey, setSortKey] = useState<'CreatedAt' | 'UpdatedAt'>('CreatedAt')
  const [search, setSearch] = useState<string>('')
  const filteredData = useMemo(() => {
    return data
      .filter((item) => item.name.includes(search))
      .sort((v1, v2) => {
        const date1 = dayjs(v1[sortKey])
        const date2 = dayjs(v2[sortKey])
        return date1.isAfter(date2) ? -1 : 1
      })
  }, [data, search, sortKey])
  return (
    <div
      className={cn(
        className,
        'flex flex-col p-4 gap-4',
        'bg-[#FCFCFE] border-2 border-[#D7D9E5] rounded-xl',
      )}
      style={style}
    >
      <div className='flex items-center whitespace-nowrap'>
        <div className='mr-40 flex gap-[16px] items-center  whitespace-nowrap'>
          <div className='flex gap-[6px] items-center'>
            <div className='font-[500] text-[14px] text-[#919497]'>
              排序方式
            </div>
            <Select
              defaultValue={sortKey}
              style={{ width: 114 }}
              popupMatchSelectWidth={false}
              onChange={setSortKey}
              classNames={{
                popup: {
                  root: styles.filterSelect,
                },
              }}
              options={[
                { value: 'CreatedAt', label: '按最新创建' },
                { value: 'updatedAt', label: '按最近更新' },
              ]}
            />
          </div>
        </div>
        <div className='ml-auto flex justify-end'>
          <div className='relative flex items-center gap-[12px]'>
            <Input
              value={search}
              placeholder={'搜索'}
              prefix={<SearchIcon className='text-[#0C99FF]' />}
              onChange={(e) => setSearch(e.target.value)}
              onBlur={() => !search?.trim?.() && setSearch(search)}
              onPressEnter={() => setSearch(search)}
              className={`w-[70px] h-[30px] border-[#0C99FF] shadow-none  ${styles.searchInputWrap} ${search?.trim?.() ? styles.searchInputWrapSearching : ''}`}
            />
          </div>
        </div>
      </div>
      <div className='flex-1 overflow-auto flex flex-wrap p-4 gap-4'>
        {filteredData.map((item) => {
          const { graph_status, id, name } = item
          return (
            <KnowledgeItem
              key={id}
              disabled={disabled || graph_status !== 'uncreated'}
              status={id === value ? 'checked' : 'unchecked'}
              title={name}
              name={name}
              onCheck={() => {
                onChange?.(id)
              }}
              onUncheck={() => {
                onChange?.(undefined)
              }}
            ></KnowledgeItem>
          )
        })}
      </div>
    </div>
  )
}
