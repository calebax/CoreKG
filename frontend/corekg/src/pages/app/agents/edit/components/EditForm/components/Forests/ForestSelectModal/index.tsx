import { FC } from 'react'
import { Checkbox, Input, List, Modal } from 'antd'
import { Agent } from 'Agent'
import { cn } from '@/utils'
import { scroll } from '@/styles/scroll'
import { SVG } from '../SVG'

type ForestList = NonNullable<Agent['forests']>
export type ForestSelectModal = {
  onClose: () => void
  forestList: ForestList
  deafultValue: ForestList
  onChange: (val: ForestList) => void
}
export const ForestSelectModal: FC<ForestSelectModal> = (props) => {
  const { onClose, forestList, deafultValue, onChange } = props
  const [value, setValue] = useState(deafultValue)
  const [search, setSearch] = useState('')
  const selectedForestList = useMemo(() => {
    return value.filter((item) => item.name.includes(search))
  }, [search, value])
  const unselectedForestList = useMemo(() => {
    return forestList
      .filter((item) => item.name.includes(search))
      .filter((item) => value.every((x) => x.id !== item.id))
  }, [forestList, search, value])
  return (
    <Modal
      title={
        <span className='flex items-center gap-5 text-base font-medium'>
          选择知识库
          <Input.Search
            className='w-54'
            placeholder='搜索'
            onSearch={setSearch}
            onClear={() => setSearch('')}
          />
        </span>
      }
      open={true}
      onCancel={onClose}
      onOk={() => {
        onChange(value)
        onClose()
      }}
      maskClosable={false}
      keyboard={false}
    >
      <div className={cn('flex flex-col max-h-[50vh] overflow-auto', scroll)}>
        <List
          className={cn({ hidden: selectedForestList.length === 0 })}
          dataSource={selectedForestList}
          renderItem={(forest) => {
            return (
              <List.Item
                key={forest.id}
                className='h-9 px-2.5 py-2 bg-[#0000001A]'
              >
                <Checkbox
                  checked
                  onChange={() => {
                    setValue(value.filter((item) => item.id !== forest.id))
                  }}
                >
                  {forest.name}
                </Checkbox>
              </List.Item>
            )
          }}
        />
        <List
          className={cn({ hidden: unselectedForestList.length === 0 })}
          dataSource={unselectedForestList}
          renderItem={(forest) => {
            return (
              <List.Item key={forest.id} className='h-9 px-2.5 py-2'>
                <Checkbox
                  onChange={() => {
                    setValue(value.concat(forest))
                  }}
                >
                  <span className='flex gap-2'>
                    <SVG type={forest.forest_type} />
                    {forest.name}
                  </span>
                </Checkbox>
              </List.Item>
            )
          }}
        />
      </div>
    </Modal>
  )
}
