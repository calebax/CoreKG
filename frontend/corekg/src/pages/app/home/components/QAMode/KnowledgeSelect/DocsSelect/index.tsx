import { FC, useState } from 'react'
import { Button, Checkbox, Empty, List, Skeleton } from 'antd'
import { cn } from '@/utils'
import DeleteIcon from '@/assets/icons/home-delete-1.svg?react'
import DeleteIcon2 from '@/assets/icons/home-delete-2.svg?react'
import { scroll } from '@/styles/scroll'
import { Knowledge } from '..'
import { useKnowledge } from '../../../KnowledgeContext'
import { Forest } from '../../../KnowledgeContext'

export type DocsSelect = {
  value: Knowledge
  setValue: (val: Knowledge) => void
  search?: string
}
export const DocsSelect: FC<DocsSelect> = (props) => {
  const { value, setValue, search } = props
  const [showSelectedValue, setShowSelectedValue] = useState(false)
  const { forestList } = useKnowledge()

  if (!forestList) return <Skeleton active />
  if (forestList.length === 0)
    return (
      <Empty
        description='暂无知识库'
        className='mb-2 text-xs'
        image={Empty.PRESENTED_IMAGE_SIMPLE}
      />
    )
  /* 按搜索结果筛选 */
  const filterForestListBySearch = (list: Forest[]) => {
    if (!search) return list
    return list.filter((item) => item.name.includes(search))
  }

  const filteredForestList = filterForestListBySearch(forestList)
  const filteredIds = filteredForestList.map((i) => i.id)
  const selectedIds = value.map((i) => i.id)
  const numSelectedInFiltered = selectedIds.filter((id) =>
    filteredIds.includes(id),
  ).length

  // 修复全选框状态逻辑
  const isAllChecked =
    filteredForestList.length > 0 &&
    numSelectedInFiltered === filteredForestList.length
  const isIndeterminate =
    numSelectedInFiltered > 0 &&
    numSelectedInFiltered < filteredForestList.length

  const handleToggleSelectAll = () => {
    if (isAllChecked) {
      setValue(value.filter((v) => !filteredIds.includes(v.id)))
      return
    }
    const toAdd = filteredForestList.filter(
      (f) => !value.some((v) => v.id === f.id),
    )
    setValue([...value, ...toAdd])
  }

  return (
    <div className={cn('overflow-hidden flex flex-col')}>
      <div className='flex items-center border-b justify-between border-[#EEF0F5] px-2 pb-1'>
        <div className='flex items-center'>
          {!showSelectedValue && (
            <Checkbox
              checked={isAllChecked}
              indeterminate={isIndeterminate}
              onChange={handleToggleSelectAll}
              className='mr-2'
            />
          )}
          <div className='flex items-center'>
            <Button
              type={!showSelectedValue ? 'link' : 'text'}
              className={cn(
                'px-0 h-10 justify-start hover:bg-transparent',
                !showSelectedValue
                  ? 'font-medium text-[#1e1f28]'
                  : 'text-[#616373] font-normal leading-[24px]',
              )}
              onClick={() => setShowSelectedValue(false)}
            >
              知识库列表
              {!showSelectedValue && filteredForestList.length > 0
                ? `（${filteredForestList.length}）`
                : ''}
            </Button>
          </div>
        </div>
        <div className='flex items-center'>
          <Button
            type={showSelectedValue ? 'link' : 'text'}
            className={cn(
              'px-0 h-10 justify-start hover:bg-transparent',
              showSelectedValue
                ? 'font-medium text-[#1e1f28]'
                : 'text-[#616373] font-normal leading-[24px]',
            )}
            onClick={() => setShowSelectedValue(true)}
          >
            已选知识库{value.length > 0 ? `（${value.length}）` : ''}
          </Button>
        </div>
        <Button
          aria-label='删除已选择的知识库'
          size='small'
          type='text'
          className='h-6 w-6 p-0 flex items-center justify-center rounded group'
          onClick={() => {
            setValue([])
            setShowSelectedValue(false)
          }}
        >
          <span className='relative inline-flex'>
            <DeleteIcon className='w-6 h-6 group-hover:hidden' />
            <DeleteIcon2 className='w-6 h-6 hidden group-hover:inline' />
          </span>
        </Button>
      </div>
      <div
        className={cn('overflow-auto break-all max-h-[28vh] pt-1 px-1', scroll)}
      >
        {showSelectedValue ? (
          <div className='flex flex-col'>
            <List
              dataSource={value}
              renderItem={(item, index) => (
                <List.Item key={item.id} className='!py-0'>
                  <div
                    className={cn(
                      'h-8 flex items-center w-full text-sm text-[#1e1f28] rounded px-2 cursor-default',
                      index === 0 ? 'bg-[#f8f9fd]' : 'hover:bg-[#F8F9FD]',
                    )}
                  >
                    {item.name}
                  </div>
                </List.Item>
              )}
            ></List>
          </div>
        ) : (
          <Checkbox.Group
            value={value}
            onChange={setValue}
            className='flex flex-col'
          >
            {filteredForestList.map((item) => (
              <label key={item.id} className='!py-0'>
                <div className='h-8 flex items-center w-full hover:bg-[#F8F9FD] rounded'>
                  <Checkbox
                    value={item}
                    className='flex-1 text-sm text-[#1e1f28]'
                  >
                    {item.name}
                  </Checkbox>
                </div>
              </label>
            ))}
          </Checkbox.Group>
        )}
      </div>
    </div>
  )
}
