import { FC } from 'react'
import { Empty, Tabs } from 'antd'
import { CloseOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import styles from './styles.module.scss'

export const NodeDetail: FC<
  Style & {
    item: {
      name: string
      tags?:
        | null
        | {
            name: string
            properties_values?: null | { name: string; value: string }[]
          }[]
    }
    clearItem: () => void
  }
> = (props) => {
  const { item, clearItem, className, style } = props
  const { name, tags } = item
  const [activeKey, setActiveKey] = useState(tags?.[0].name)
  return (
    <div
      className={cn(
        'bg-white rounded',
        'h-80 w-80 overflow-hidden',
        'p-2.5 flex flex-col',
        styles.nodeDetail,
        className,
      )}
      style={style}
    >
      <div className='flex items-baseline justify-between'>
        <span className='px-2.5 py-0.5 bg-[#FFF1C3] text-[#FF6600]'>
          {name}
        </span>
        <CloseOutlined onClick={clearItem} />
      </div>

      {!tags || tags.length === 0 ? (
        <Empty description='没有实体类型' />
      ) : (
        <div className='flex-1 flex flex-col overflow-hidden'>
          <Tabs
            className={cn('overflow-hidden overflow-x-auto', styles.tabs)}
            activeKey={activeKey}
            onChange={setActiveKey}
            items={tags.map((t) => {
              return {
                key: t.name,
                label: t.name,
              }
            })}
          />
          <Properties
            className='flex-1'
            key={activeKey}
            value={
              tags.find((item) => item.name === activeKey)!.properties_values
            }
          />
        </div>
      )}
    </div>
  )
}

const Properties: FC<
  Style & { value?: null | { name: string; value: string }[] }
> = (props) => {
  const { value, className, style } = props
  if (!value || value.length === 0) {
    return <Empty description='本实体在此类型下没有属性' />
  }
  return (
    <div
      className={cn(' overflow-auto flex flex-col gap-3', className)}
      style={style}
    >
      {value.map((item) => {
        const { name, value } = item
        return (
          <span
            key={name}
            className='flex items-baseline whitespace-nowrap text-[#616373]'
          >
            {name}：
            <span className='text-[#1E1F28]  break-all whitespace-pre-wrap'>
              {value}
            </span>
          </span>
        )
      })}
    </div>
  )
}
