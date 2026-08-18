import { FC } from 'react'
import { Empty, List, Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import InfoWarningIcon from '@/assets/icons/info-warning.svg?react'
import { Knowledge } from '../..'
import { KnowledgeIcon } from '../KnowledgeIcon'

/** 展示当前选中的知识 */
export const KnowledgeList: FC<{
  items?: Knowledge[]
  title: string
  /** 覆盖默认的提示文案（如表格模式：最多可展示200个已选表格） */
  tooltipText?: string
}> = (props) => {
  const { items, title, tooltipText } = props
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  if (!items || items.length === 0) {
    return <Empty className='p-4' description={tC('empty.noData')} />
  }
  return (
    <div className='bg-white rounded-[10px] shadow-[0px_2px_12px_0px_rgba(0,0,0,0.1)] p-[10px] w-[266px] flex flex-col gap-2'>
      <div className='flex flex-col gap-1'>
        <div className='flex items-center justify-between h-10 px-2 border-b border-[#eeeeee]'>
          <div className='flex items-center gap-1'>
            <span className='text-sm font-medium text-[#1e1f28]'>{title}</span>
            <Tooltip
              title={tooltipText ?? t('project.maximumDisplay', { num: 200 })}
              placement='top'
            >
              <InfoWarningIcon className='w-4 h-4 cursor-pointer mt-[1.9px]' />
            </Tooltip>
          </div>
        </div>
        <div className={cn('max-h-[28vh] overflow-auto', scroll)}>
          <List
            split={false}
            dataSource={items}
            renderItem={(item, index) => (
              <List.Item key={item.id} className='!py-0 !px-0'>
                <div
                  className={cn(
                    'h-8 w-full text-sm text-[#1e1f28] px-2 cursor-default flex items-center',
                    index === 0
                      ? 'bg-[#f8f9fd]'
                      : 'hover:bg-[#f8f9fd] hover:cursor-pointer',
                  )}
                  style={{
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    minWidth: 0,
                  }}
                >
                  <KnowledgeIcon type={item.type} />
                  <span className='truncate w-full'>{item.name}</span>
                </div>
              </List.Item>
            )}
          />
        </div>
      </div>
    </div>
  )
}
