import { FC, ReactNode, useEffect, useState } from 'react'
import { Empty, Input, Popover, Tree, Typography } from 'antd'
import { useControllableValue } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { ChevronDownIcon } from 'tdesign-icons-react'
import { match } from 'ts-pattern'
import { cn } from '@/utils'
import SearchIcon from '@/assets/icons/search.svg?react'
import { scroll } from '@/styles/scroll'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { ProjectInfo } from '../..'
import Deepseek from './images/deepseek.svg?react'
import Doubao from './images/doubao.svg?react'
import LastModel from './images/last-model.svg?react'
import ModelIcon from './images/model.svg?react'
import QWen from './images/qwen.svg?react'
import './index.css'

export type ModelSelect = Style & {
  value?: number
  allowSelect?: boolean
  showArrow?: boolean
  models: ProjectInfo['models']
  onChange?: (val: number) => void
}
export const ModelSelect: FC<ModelSelect> = (props) => {
  const { t: tC } = useTranslation('common')
  const { version } = useDeployConfig()
  const { allowSelect, models, className, style, showArrow = true } = props
  const [value, onChange] = useControllableValue<number | string>(props)
  const [search, setSearch] = useState('')
  const [open, setOpen] = useState(false)

  useEffect(() => {
    if (!value && models?.length) {
      onChange?.(models[0].id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allowSelect, models, value])

  const getModelBtn = (
    props: {
      text?: string
      active?: boolean
    } = {},
  ) => {
    const { text = tC('model.selectModel'), active } = props
    return (
      <div
        className={cn(
          ' cursor-pointer ',
          'text-[13px] text-[#6e757f]',
          'py-1 px-3 flex items-center gap-1',
          {
            'text-[#CC5DE8]': active,
          },
          className,
        )}
        style={style}
      >
        <ModelIcon />
        {text}
        {showArrow && <ChevronDownIcon />}
      </div>
    )
  }
  const selectedModel = models?.find((item) => item.id === value)?.name

  if (!allowSelect) {
    return getModelBtn({ text: selectedModel, active: true })
  }

  const popoverContent = (() => {
    if (!models || models.length === 0) {
      return <Empty description={tC('empty.noData')} className='p-4' />
    }
    return (
      <div
        className={cn(
          'bg-white rounded-[10px] shadow-[0px_2px_12px_0px_rgba(0,0,0,0.1)]',
          'flex flex-col gap-2.5 p-[10px]',
          'h-80 min-w-80 max-w-[40vw]',
        )}
      >
        <Input
          prefix={<SearchIcon className='w-3.5 h-3.5' />}
          allowClear
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={tC('button.search')}
          className={cn('h-8 rounded-[19px] px-2 text-sm text-[#1e1f28]')}
          style={{
            borderColor: search ? '#a895fc' : '#d7d9e5',
          }}
        />
        <div className='flex-1 overflow-hidden'>
          {match({ version })
            .with({ version: 'saas' }, () => {
              const map: Record<string, ProjectInfo['models']> = {}
              // 按 model_group 分组
              models.forEach((m) => {
                if (!map[m.model_group]) {
                  map[m.model_group] = []
                }
                map[m.model_group].push(m)
              })
              // 分组图标映射
              const groupIcons: Record<string, ReactNode> = {
                DeepSeek: <Deepseek className='w-[18px] h-[18px]' />,
                通义千问: <QWen className='w-[18px] h-[18px]' />,
                豆包: <Doubao className='w-[18px] h-[18px]' />,
              }
              const nodes = Object.entries(map)
                .map(([group, items]) => {
                  const filteredItems = search
                    ? items.filter((m) => m.name.toLowerCase().includes(search))
                    : items
                  if (filteredItems.length === 0) return null
                  // 获取分组对应的本地 SVG 图标
                  const GroupIcon = groupIcons[group]
                  return {
                    // 将图标和分组名称组合作为 title
                    title: (
                      <span className='inline-flex items-center gap-1.5'>
                        {GroupIcon}
                        <span className='font-medium'>{group}</span>
                      </span>
                    ),
                    key: group,
                    checkable: false,
                    children: filteredItems.map((m) => ({
                      ...m,
                      title: m.name,
                      key: m.id,
                      checkable: true,
                    })),
                  }
                })
                .filter(Boolean)
              const lastUsedModel = models
                .filter((item) => item.is_last_used)
                .sort((v1, v2) => {
                  const date1 = dayjs(v1.last_used_at)
                  const date2 = dayjs(v2.last_used_at)
                  return date1.isAfter(date2) ? -1 : 1
                })
                .slice(0, 3)
              return (
                <div className='max-h-full overflow-hidden flex flex-col gap-1'>
                  <span
                    className={cn('flex items-center gap-1 font-medium', {
                      hidden: !lastUsedModel.length,
                    })}
                  >
                    <LastModel />
                    最近使用
                  </span>
                  <div
                    className={cn('flex gap-7', {
                      hidden: !lastUsedModel.length,
                    })}
                  >
                    {lastUsedModel.map((item, i) => {
                      return (
                        <Typography.Paragraph
                          key={item.id}
                          className={cn(
                            'cursor-pointer bg-[#0C99FF1A] text-[#0C99FF] rounded-full',
                            'px-1.5 py-0.5 m-0 max-w-25',
                            { 'ml-2': i !== 0 },
                          )}
                          onClick={() => {
                            onChange?.(item.id)
                            setOpen(false)
                          }}
                          ellipsis={{ rows: 1, tooltip: item.name }}
                        >
                          {item.name}
                        </Typography.Paragraph>
                      )
                    })}
                  </div>
                  <Tree
                    className='flex-1 overflow-auto model-select-tree'
                    defaultExpandAll
                    treeData={nodes as any[]}
                    checkable
                    selectable={false}
                    checkedKeys={[value]}
                    onCheck={(_, info) => {
                      onChange?.(info.node.id)
                      setOpen(false)
                    }}
                  />
                </div>
              )
            })
            .otherwise(() => (
              <div
                className={cn(
                  'overflow-auto h-full max-h-full flex flex-col gap-1 ',
                  scroll,
                )}
              >
                {models
                  .filter((item) => item.name.includes(search))
                  .map((item) => (
                    <div
                      key={item.id}
                      className={cn(
                        'h-7.5 flex flex-col justify-center px-2 py-1 cursor-pointer hover:bg-[#F7F7F7] rounded',
                        item.id === value ? 'bg-[#F7F7F7] rounded' : '',
                      )}
                      onClick={() => {
                        onChange?.(item.id)
                        setOpen(false)
                      }}
                    >
                      <div className='flex flex-col justify-center mb-[4px] w-full h-[48px]'>
                        <p
                          className={cn(
                            'block leading-6 text-sm text-[#1e1f28]',
                          )}
                        >
                          {item.name}
                        </p>
                      </div>
                      <div className='flex flex-col justify-center w-full font-normal'>
                        <p className='block leading-6 text-xs text-[#616373]'>
                          {item.desc}
                        </p>
                      </div>
                    </div>
                  ))}
              </div>
            ))}
        </div>
      </div>
    )
  })()

  return (
    <Popover
      trigger={['click']}
      open={open}
      onOpenChange={setOpen}
      placement='topRight'
      arrow={false}
      content={popoverContent}
    >
      {selectedModel
        ? getModelBtn({ text: selectedModel, active: true })
        : getModelBtn()}
    </Popover>
  )
}
